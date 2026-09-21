package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/config"
	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

type API struct {
	cfg          *config.Config
	auth         *service.AuthService
	accounts     *service.AccountService
	categories   *service.CategoryService
	tags         *service.TagService
	templates    *service.TemplateService
	fareRules    *service.FareRuleService
	holidays     *service.HolidayService
	schedules    *service.ScheduleService
	transactions *service.TransactionService
	attachments  *service.AttachmentService
	stats        *service.StatsService
	llm          *service.LLMService
	minioReady   bool
}

func NewAPI(
	cfg *config.Config,
	auth *service.AuthService,
	accounts *service.AccountService,
	categories *service.CategoryService,
	tags *service.TagService,
	templates *service.TemplateService,
	fareRules *service.FareRuleService,
	holidays *service.HolidayService,
	schedules *service.ScheduleService,
	transactions *service.TransactionService,
	attachments *service.AttachmentService,
	stats *service.StatsService,
	llm *service.LLMService,
	minioReady bool,
) *API {
	return &API{
		cfg: cfg, auth: auth, accounts: accounts, categories: categories,
		tags: tags, templates: templates, fareRules: fareRules, holidays: holidays,
		schedules: schedules, transactions: transactions, attachments: attachments,
		stats: stats, llm: llm, minioReady: minioReady,
	}
}

func (a *API) Health(c *gin.Context) {
	OK(c, gin.H{
		"status":     "up",
		"minio":      a.minioReady,
		"ai":         a.llm != nil && a.llm.Enabled(),
		"register":   a.cfg.AllowRegister,
	})
}

func (a *API) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		BadRequest(c, "请输入用户名和密码")
		return
	}
	token, user, err := a.auth.Login(c.Request.Context(), req.Username, req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		Fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	if errors.Is(err, service.ErrUserDisabled) {
		Fail(c, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	_ = a.categories.EnsureDefaults(user.ID)
	_ = a.tags.EnsureDefaults(user.ID)
	OK(c, gin.H{"token": token, "user": user})
}

func (a *API) Register(c *gin.Context) {
	if !a.cfg.AllowRegister {
		Fail(c, http.StatusForbidden, "当前未开放注册")
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	user, err := a.auth.Register(req.Username, req.Password, req.Email)
	if errors.Is(err, service.ErrUserExists) || errors.Is(err, service.ErrWeakPassword) {
		BadRequest(c, err.Error())
		return
	}
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	_ = a.categories.EnsureDefaults(user.ID)
	_ = a.tags.EnsureDefaults(user.ID)
	// 注册后直接登录
	token, _, err := a.auth.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		OK(c, gin.H{"user": user})
		return
	}
	OK(c, gin.H{"token": token, "user": user})
}

func (a *API) ForgotPassword(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" {
		BadRequest(c, "请输入用户名")
		return
	}
	res, err := a.auth.ForgotPassword(c.Request.Context(), req.Username)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, res)
}

func (a *API) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" {
		BadRequest(c, "请提供令牌和新密码")
		return
	}
	err := a.auth.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if errors.Is(err, service.ErrBadResetToken) || errors.Is(err, service.ErrWeakPassword) {
		BadRequest(c, err.Error())
		return
	}
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, gin.H{"ok": true})
}

func (a *API) Logout(c *gin.Context) {
	token, _ := c.Get("sessionToken")
	if t, ok := token.(string); ok && t != "" {
		_ = a.auth.Logout(c.Request.Context(), t)
	}
	OK(c, gin.H{"loggedOut": true})
}

func (a *API) Me(c *gin.Context) {
	uid := middleware.GetUserID(c)
	user, err := a.auth.GetUser(uid)
	if err != nil {
		NotFound(c, "用户不存在")
		return
	}
	if user.Disabled {
		if tok, ok := c.Get("sessionToken"); ok {
			if s, _ := tok.(string); s != "" {
				_ = a.auth.Logout(c.Request.Context(), s)
			}
		}
		Fail(c, http.StatusForbidden, service.ErrUserDisabled.Error())
		return
	}
	_ = a.categories.EnsureDefaults(user.ID)
	_ = a.tags.EnsureDefaults(user.ID)
	OK(c, user)
}

func (a *API) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	err := a.auth.ChangePassword(middleware.GetUserID(c), req.OldPassword, req.NewPassword)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"ok": true})
}

func (a *API) UpdateSettings(c *gin.Context) {
	var req struct {
		Email            *string `json:"email"`
		DefaultAccountID *uint64 `json:"defaultAccountId"`
		ClearDefaultAcc  bool    `json:"clearDefaultAccount"`
		WeekStart        *int    `json:"weekStart"`
		ExpenseColor     *string `json:"expenseColor"`
		IncomeColor      *string `json:"incomeColor"`
		Theme            *string `json:"theme"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	in := service.UserSettingsInput{
		Email: req.Email, WeekStart: req.WeekStart,
		ExpenseColor: req.ExpenseColor, IncomeColor: req.IncomeColor, Theme: req.Theme,
	}
	if req.ClearDefaultAcc {
		var nilID *uint64
		in.DefaultAccountID = &nilID
	} else if req.DefaultAccountID != nil {
		id := req.DefaultAccountID
		in.DefaultAccountID = &id
	}
	user, err := a.auth.UpdateSettings(middleware.GetUserID(c), in)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, user)
}

func (a *API) ListAccounts(c *gin.Context) {
	include := c.Query("archived") == "1"
	list, err := a.accounts.List(middleware.GetUserID(c), include)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (a *API) CreateAccount(c *gin.Context) {
	var req struct {
		Name            string   `json:"name"`
		Type            string   `json:"type"`
		Balance         int64    `json:"balance"`
		Sort            int      `json:"sort"`
		CreditLimit     *int64   `json:"creditLimit"`
		CreditBilledFen *int64   `json:"creditBilledFen"`
		BillingDay      *int     `json:"billingDay"`
		PaymentDueDay   *int     `json:"paymentDueDay"`
		Institution     *string  `json:"institution"`
		CardNo          *string  `json:"cardNo"`
		HolderName      *string  `json:"holderName"`
		StorageNote     *string  `json:"storageNote"`
		Remark          *string  `json:"remark"`
		AttachmentIDs   []uint64 `json:"attachmentIds"`
		// UsedCredit：信用账户前端可传「已用额度（正数）」；若有则覆盖 balance
		UsedCredit *int64 `json:"usedCredit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	bal := req.Balance
	if req.Type == "credit" && req.UsedCredit != nil {
		bal = *req.UsedCredit
	}
	acc, err := a.accounts.Create(middleware.GetUserID(c), service.AccountInput{
		Name: req.Name, Type: req.Type, Balance: bal, Sort: req.Sort,
		CreditLimit: req.CreditLimit, CreditBilledFen: req.CreditBilledFen,
		BillingDay: req.BillingDay, PaymentDueDay: req.PaymentDueDay,
		Institution: req.Institution, CardNo: req.CardNo, HolderName: req.HolderName,
		StorageNote: req.StorageNote, Remark: req.Remark, AttachmentIDs: req.AttachmentIDs,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, acc)
}

func (a *API) UpdateAccount(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name             string   `json:"name"`
		Type             string   `json:"type"`
		Sort             int      `json:"sort"`
		Archived         *bool    `json:"archived"`
		Balance          *int64   `json:"balance"`
		UsedCredit       *int64   `json:"usedCredit"` // 信用：已用额度（正）
		CreditLimit      *int64   `json:"creditLimit"`
		CreditBilledFen  *int64   `json:"creditBilledFen"`
		BillingDay       *int     `json:"billingDay"`
		PaymentDueDay    *int     `json:"paymentDueDay"`
		Institution      *string  `json:"institution"`
		CardNo           *string  `json:"cardNo"`
		HolderName       *string  `json:"holderName"`
		StorageNote      *string  `json:"storageNote"`
		Remark           *string  `json:"remark"`
		AttachmentIDs    []uint64 `json:"attachmentIds"`
		LastReconciledAt *string  `json:"lastReconciledAt"`
		ClearReconciled  bool     `json:"clearReconciled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	in := service.AccountInput{
		Name: req.Name, Type: req.Type, Sort: req.Sort, Archived: req.Archived,
		CreditLimit: req.CreditLimit, CreditBilledFen: req.CreditBilledFen,
		BillingDay: req.BillingDay, PaymentDueDay: req.PaymentDueDay,
		Institution: req.Institution, CardNo: req.CardNo, HolderName: req.HolderName,
		StorageNote: req.StorageNote, Remark: req.Remark, AttachmentIDs: req.AttachmentIDs,
		ClearReconciled: req.ClearReconciled,
	}
	if req.UsedCredit != nil {
		in.Balance = *req.UsedCredit
		in.SetBalance = true
	} else if req.Balance != nil {
		in.Balance = *req.Balance
		in.SetBalance = true
	}
	if req.LastReconciledAt != nil && *req.LastReconciledAt != "" {
		if t := parseTimePtr(*req.LastReconciledAt); t != nil {
			in.LastReconciledAt = t
		}
	}
	acc, err := a.accounts.Update(middleware.GetUserID(c), id, in)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, acc)
}

// AccountSensitive 显式拉取卡号/户名明文（前端不常驻保留）
func (a *API) AccountSensitive(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	cardNo, holder, err := a.accounts.GetSensitive(middleware.GetUserID(c), id)
	if err != nil {
		BadRequest(c, "账户不存在")
		return
	}
	OK(c, gin.H{
		"cardNo":     service.FormatCardGroups(cardNo),
		"holderName": holder,
	})
}

func (a *API) DeleteAccount(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := a.accounts.Delete(middleware.GetUserID(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"deleted": true})
}

// CreditStatement 信用卡账单展示（只读）
func (a *API) CreditStatement(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	stmt, err := a.accounts.CreditStatement(middleware.GetUserID(c), id, time.Now())
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, stmt)
}

// CreditRepayReminders 临近还款日且有欠款的信用卡提醒
func (a *API) CreditRepayReminders(c *gin.Context) {
	list, err := a.accounts.CreditRepayReminders(middleware.GetUserID(c), time.Now())
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (a *API) ListCategories(c *gin.Context) {
	flat := c.Query("flat") == "1"
	var (
		list any
		err  error
	)
	if flat {
		list, err = a.categories.ListFlat(middleware.GetUserID(c), c.Query("kind"))
	} else {
		list, err = a.categories.List(middleware.GetUserID(c), c.Query("kind"))
	}
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (a *API) CreateCategory(c *gin.Context) {
	var req struct {
		Name     string  `json:"name"`
		Kind     string  `json:"kind"`
		Icon     string  `json:"icon"`
		Sort     int     `json:"sort"`
		ParentID *uint64 `json:"parentId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	cat, err := a.categories.Create(middleware.GetUserID(c), service.CategoryInput{
		Name: req.Name, Kind: req.Kind, Icon: req.Icon, Sort: req.Sort, ParentID: req.ParentID,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, cat)
}

func (a *API) UpdateCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name     string  `json:"name"`
		Icon     string  `json:"icon"`
		Sort     int     `json:"sort"`
		ParentID *uint64 `json:"parentId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	cat, err := a.categories.Update(middleware.GetUserID(c), id, service.CategoryInput{
		Name: req.Name, Icon: req.Icon, Sort: req.Sort, ParentID: req.ParentID,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, cat)
}

func (a *API) DeleteCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := a.categories.Delete(middleware.GetUserID(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (a *API) ListTags(c *gin.Context) {
	list, err := a.tags.List(middleware.GetUserID(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (a *API) CreateTag(c *gin.Context) {
	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
		Sort  int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	t, err := a.tags.Create(middleware.GetUserID(c), req.Name, req.Color, req.Sort)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, t)
}

func (a *API) UpdateTag(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
		Sort  int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	t, err := a.tags.Update(middleware.GetUserID(c), id, req.Name, req.Color, req.Sort)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, t)
}

func (a *API) DeleteTag(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := a.tags.Delete(middleware.GetUserID(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (a *API) ListTemplates(c *gin.Context) {
	list, err := a.templates.List(middleware.GetUserID(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (a *API) CreateTemplate(c *gin.Context) {
	var req service.TemplateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	t, err := a.templates.Create(middleware.GetUserID(c), req)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, t)
}

func (a *API) UpdateTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.TemplateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	t, err := a.templates.Update(middleware.GetUserID(c), id, req)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, t)
}

func (a *API) DeleteTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := a.templates.Delete(middleware.GetUserID(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (a *API) PreviewTemplateFare(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Date          string `json:"date"`
		MonthSpentFen *int64 `json:"monthSpentFen"`
	}
	_ = c.ShouldBindJSON(&req)
	at := time.Now()
	if req.Date != "" {
		if t, err := time.ParseInLocation("2006-01-02", req.Date, time.Local); err == nil {
			at = t
		}
	}
	res, err := a.templates.PreviewFare(middleware.GetUserID(c), id, at, req.MonthSpentFen)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, res)
}

func (a *API) ListFareRules(c *gin.Context) {
	list, err := a.fareRules.List(middleware.GetUserID(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (a *API) CreateFareRule(c *gin.Context) {
	var req service.FareRuleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	r, err := a.fareRules.Create(middleware.GetUserID(c), req)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, r)
}

func (a *API) UpdateFareRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.FareRuleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	r, err := a.fareRules.Update(middleware.GetUserID(c), id, req)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, r)
}

func (a *API) DeleteFareRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := a.fareRules.Delete(middleware.GetUserID(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (a *API) PreviewFareRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Date          string `json:"date"`
		MonthSpentFen *int64 `json:"monthSpentFen"`
		BaseFen       *int64 `json:"baseFen"`
		RideCount     *int   `json:"rideCount"`
	}
	_ = c.ShouldBindJSON(&req)
	at := time.Now()
	if req.Date != "" {
		if t, err := time.ParseInLocation("2006-01-02", req.Date, time.Local); err == nil {
			at = t
		}
	}
	res, err := a.fareRules.Preview(middleware.GetUserID(c), id, at, req.MonthSpentFen, req.BaseFen, req.RideCount)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, res)
}

func (a *API) GetHolidays(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	if year == 0 {
		year = time.Now().Year()
	}
	if a.holidays == nil {
		BadRequest(c, "节假日服务未启用")
		return
	}
	y, err := a.holidays.Year(c.Request.Context(), year)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, y)
}

func (a *API) RefreshHolidays(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	if year == 0 {
		year = time.Now().Year()
	}
	if a.holidays == nil {
		BadRequest(c, "节假日服务未启用")
		return
	}
	y, err := a.holidays.RefreshYear(c.Request.Context(), year, "manual")
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, y)
}

func (a *API) AdminListUsers(c *gin.Context) {
	list, err := a.auth.ListUsers()
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (a *API) AdminUpdateUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Role     *string `json:"role"`
		Disabled *bool   `json:"disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if req.Role == nil && req.Disabled == nil {
		BadRequest(c, "请指定要修改的字段（role / disabled）")
		return
	}
	user, err := a.auth.AdminUpdateUser(c.Request.Context(), middleware.GetUserID(c), id, service.AdminUserUpdateInput{
		Role: req.Role, Disabled: req.Disabled,
	})
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, user)
}

func (a *API) AdminDeleteUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	err := a.auth.AdminDeleteUser(c.Request.Context(), middleware.GetUserID(c), id)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (a *API) ListSchedules(c *gin.Context) {
	list, err := a.schedules.List(middleware.GetUserID(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (a *API) CreateSchedule(c *gin.Context) {
	var req struct {
		Name        string   `json:"name"`
		Type        string   `json:"type"`
		Amount      int64    `json:"amount"`
		AccountID   uint64   `json:"accountId"`
		ToAccountID *uint64  `json:"toAccountId"`
		CategoryID  *uint64  `json:"categoryId"`
		Remark      string   `json:"remark"`
		TagIDs      []uint64 `json:"tagIds"`
		Frequency   string   `json:"frequency"`
		IntervalN   int      `json:"intervalN"`
		Enabled     *bool    `json:"enabled"`
		NextRunAt   string   `json:"nextRunAt"`
		EndAt       string   `json:"endAt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	in := service.ScheduleInput{
		Name: req.Name, Type: req.Type, Amount: req.Amount, AccountID: req.AccountID,
		ToAccountID: req.ToAccountID, CategoryID: req.CategoryID, Remark: req.Remark,
		TagIDs: req.TagIDs, Frequency: req.Frequency, IntervalN: req.IntervalN, Enabled: req.Enabled,
	}
	if t := parseTimePtr(req.NextRunAt); t != nil {
		in.NextRunAt = *t
	}
	if t := parseTimePtr(req.EndAt); t != nil {
		in.EndAt = t
	}
	sch, err := a.schedules.Create(middleware.GetUserID(c), in)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, sch)
}

func (a *API) UpdateSchedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name        string   `json:"name"`
		Type        string   `json:"type"`
		Amount      int64    `json:"amount"`
		AccountID   uint64   `json:"accountId"`
		ToAccountID *uint64  `json:"toAccountId"`
		CategoryID  *uint64  `json:"categoryId"`
		Remark      string   `json:"remark"`
		TagIDs      []uint64 `json:"tagIds"`
		Frequency   string   `json:"frequency"`
		IntervalN   int      `json:"intervalN"`
		Enabled     *bool    `json:"enabled"`
		NextRunAt   string   `json:"nextRunAt"`
		EndAt       string   `json:"endAt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	in := service.ScheduleInput{
		Name: req.Name, Type: req.Type, Amount: req.Amount, AccountID: req.AccountID,
		ToAccountID: req.ToAccountID, CategoryID: req.CategoryID, Remark: req.Remark,
		TagIDs: req.TagIDs, Frequency: req.Frequency, IntervalN: req.IntervalN, Enabled: req.Enabled,
	}
	if t := parseTimePtr(req.NextRunAt); t != nil {
		in.NextRunAt = *t
	}
	if req.EndAt != "" {
		in.EndAt = parseTimePtr(req.EndAt)
	}
	sch, err := a.schedules.Update(middleware.GetUserID(c), id, in)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, sch)
}

func (a *API) DeleteSchedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := a.schedules.Delete(middleware.GetUserID(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"deleted": true})
}

func parseTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	layouts := []string{time.RFC3339, "2006-01-02", "2006-01-02 15:04:05", "2006-01-02T15:04"}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
			return &t
		}
	}
	return nil
}

func (a *API) ListTransactions(c *gin.Context) {
	q := service.TransactionQuery{Limit: 50}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q.Limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q.Offset = n
		}
	}
	if v := c.Query("accountId"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			q.AccountID = &n
		}
	}
	if v := c.Query("categoryId"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			q.CategoryID = &n
		}
	}
	if v := c.Query("tagId"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			q.TagID = &n
		}
	}
	q.Type = c.Query("type")
	q.Keyword = c.Query("keyword")
	q.From = parseTimePtr(c.Query("from"))
	q.To = parseTimePtr(c.Query("to"))
	res, err := a.transactions.List(middleware.GetUserID(c), q)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	if a.attachments != nil {
		a.attachments.RefreshTxListAttachments(c.Request.Context(), res.Items)
	}
	OK(c, res)
}

func (a *API) GetTransaction(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	t, err := a.transactions.Get(middleware.GetUserID(c), id)
	if err != nil {
		NotFound(c, "流水不存在")
		return
	}
	if a.attachments != nil {
		a.attachments.RefreshTxAttachments(c.Request.Context(), t)
	}
	OK(c, t)
}

type txBody struct {
	Type               string   `json:"type"`
	Amount             int64    `json:"amount"`
	AccountID          uint64   `json:"accountId"`
	ToAccountID        *uint64  `json:"toAccountId"`
	CategoryID         *uint64  `json:"categoryId"`
	Remark             string   `json:"remark"`
	GeoMode            string   `json:"geoMode"`
	GeoLng             *float64 `json:"geoLng"`
	GeoLat             *float64 `json:"geoLat"`
	GeoName            string   `json:"geoName"`
	GeoEndLng          *float64 `json:"geoEndLng"`
	GeoEndLat          *float64 `json:"geoEndLat"`
	GeoEndName         string   `json:"geoEndName"`
	FareRuleID         *uint64  `json:"fareRuleId"`
	InstallmentPeriods int      `json:"installmentPeriods"`
	InterestFen        int64    `json:"interestFen"`
	HappenedAt         string   `json:"happenedAt"`
	AttachmentIDs      []uint64 `json:"attachmentIds"`
	TagIDs             []uint64 `json:"tagIds"`
}

func (b txBody) toInput() (service.TransactionInput, error) {
	in := service.TransactionInput{
		Type: b.Type, Amount: b.Amount, AccountID: b.AccountID, ToAccountID: b.ToAccountID,
		CategoryID: b.CategoryID, Remark: b.Remark, GeoMode: b.GeoMode,
		GeoLng: b.GeoLng, GeoLat: b.GeoLat, GeoName: b.GeoName,
		GeoEndLng: b.GeoEndLng, GeoEndLat: b.GeoEndLat, GeoEndName: b.GeoEndName,
		FareRuleID: b.FareRuleID,
		InstallmentPeriods: b.InstallmentPeriods, InterestFen: b.InterestFen,
		AttachmentIDs: b.AttachmentIDs, TagIDs: b.TagIDs,
	}
	if b.HappenedAt == "" {
		in.HappenedAt = time.Now()
	} else if t := parseTimePtr(b.HappenedAt); t != nil {
		in.HappenedAt = *t
	} else {
		return in, errors.New("时间格式无效")
	}
	return in, nil
}

func (a *API) CreateTransaction(c *gin.Context) {
	var body txBody
	if err := c.ShouldBindJSON(&body); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	in, err := body.toInput()
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	t, err := a.transactions.Create(middleware.GetUserID(c), in)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	if a.attachments != nil {
		a.attachments.RefreshTxAttachments(c.Request.Context(), t)
	}
	OK(c, t)
}

func (a *API) UpdateTransaction(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body txBody
	if err := c.ShouldBindJSON(&body); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	in, err := body.toInput()
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	t, err := a.transactions.Update(middleware.GetUserID(c), id, in)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	if a.attachments != nil {
		a.attachments.RefreshTxAttachments(c.Request.Context(), t)
	}
	OK(c, t)
}

func (a *API) DeleteTransaction(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := a.transactions.Delete(middleware.GetUserID(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (a *API) UploadAttachment(c *gin.Context) {
	if !a.minioReady || a.attachments == nil {
		Fail(c, http.StatusServiceUnavailable, "附件服务未就绪（请启动 MinIO）")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "请上传文件")
		return
	}
	f, err := file.Open()
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	defer f.Close()
	ct := file.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	att, err := a.attachments.Upload(c.Request.Context(), middleware.GetUserID(c), file.Filename, ct, f, file.Size)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, att)
}

func (a *API) StatsSummary(c *gin.Context) {
	q := service.StatsQuery{}
	if t := parseTimePtr(c.Query("from")); t != nil {
		q.From = *t
	}
	if t := parseTimePtr(c.Query("to")); t != nil {
		q.To = *t
	}
	if v := c.Query("accountId"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			q.AccountID = &n
		}
	}
	sum, err := a.stats.Summary(middleware.GetUserID(c), q)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, sum)
}

func (a *API) RecognizeText(c *gin.Context) {
	if a.llm == nil || !a.llm.Enabled() {
		Fail(c, http.StatusServiceUnavailable, "未配置 DeepSeek API Key")
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	res, err := a.llm.RecognizeText(c.Request.Context(), middleware.GetUserID(c), req.Text)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, res)
}

func (a *API) RecognizeTextBatch(c *gin.Context) {
	if a.llm == nil || !a.llm.Enabled() {
		Fail(c, http.StatusServiceUnavailable, "未配置 DeepSeek API Key")
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	list, err := a.llm.RecognizeTextBatch(c.Request.Context(), middleware.GetUserID(c), req.Text)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"items": list})
}

func (a *API) RecognizeImage(c *gin.Context) {
	if a.llm == nil || !a.llm.Enabled() {
		Fail(c, http.StatusServiceUnavailable, "未配置 DeepSeek API Key")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "请上传图片（字段名 file）")
		return
	}
	f, err := file.Open()
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 12<<20+1))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	if len(data) > 12<<20 {
		BadRequest(c, "图片过大（上限约 12MB）")
		return
	}
	ct := file.Header.Get("Content-Type")
	hint := c.PostForm("hint")
	res, err := a.llm.RecognizeImage(c.Request.Context(), middleware.GetUserID(c), data, ct, hint)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, res)
}

func (a *API) AmapConfig(c *gin.Context) {
	OK(c, gin.H{
		"key":            a.cfg.AmapKey,
		"securityJsCode": a.cfg.AmapSecurityCode,
	})
}
