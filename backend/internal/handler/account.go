package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

type installmentPlanBody struct {
	Name                 string `json:"name"`
	PrincipalFen         int64  `json:"principalFen"`
	Periods              int    `json:"periods"`
	InterestPerPeriodFen int64  `json:"interestPerPeriodFen"`
	FirstDueOn           string `json:"firstDueOn"`
	Sort                 int    `json:"sort"`
}

// toPlanInputs 未传 installmentPlans 时返回 nil，表示不改动已有分期
func toPlanInputs(src *[]installmentPlanBody) *[]service.CreditInstallmentPlanInput {
	if src == nil {
		return nil
	}
	plans := make([]service.CreditInstallmentPlanInput, 0, len(*src))
	for _, p := range *src {
		plans = append(plans, service.CreditInstallmentPlanInput{
			Name: p.Name, PrincipalFen: p.PrincipalFen, Periods: p.Periods,
			InterestPerPeriodFen: p.InterestPerPeriodFen, FirstDueOn: p.FirstDueOn, Sort: p.Sort,
		})
	}
	return &plans
}

func (a *API) ListAccounts(c *gin.Context) {
	include := c.Query("archived") == "1"
	list, err := a.accounts.List(middleware.GetUserID(c), include)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	if a.attachments != nil {
		a.attachments.RefreshAccountListAttachments(list)
	}
	OK(c, list)
}

func (a *API) CreateAccount(c *gin.Context) {
	var req struct {
		Name             string                 `json:"name"`
		Type             string                 `json:"type"`
		Balance          int64                  `json:"balance"`
		Sort             int                    `json:"sort"`
		CreditLimit      *int64                 `json:"creditLimit"`
		CreditBilledFen  *int64                 `json:"creditBilledFen"`
		BillingDay       *int                   `json:"billingDay"`
		PaymentDueDay    *int                   `json:"paymentDueDay"`
		Institution      *string                `json:"institution"`
		CardNo           *string                `json:"cardNo"`
		HolderName       *string                `json:"holderName"`
		StorageNote      *string                `json:"storageNote"`
		Remark           *string                `json:"remark"`
		AttachmentIDs    []uint64               `json:"attachmentIds"`
		UsedCredit       *int64                 `json:"usedCredit"`
		InstallmentPlans *[]installmentPlanBody `json:"installmentPlans"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	bal := req.Balance
	if req.Type == "credit" && req.UsedCredit != nil {
		bal = *req.UsedCredit
	}
	in := service.AccountInput{
		Name: req.Name, Type: req.Type, Balance: bal, Sort: req.Sort,
		CreditLimit: req.CreditLimit, CreditBilledFen: req.CreditBilledFen,
		BillingDay: req.BillingDay, PaymentDueDay: req.PaymentDueDay,
		Institution: req.Institution, CardNo: req.CardNo, HolderName: req.HolderName,
		StorageNote: req.StorageNote, Remark: req.Remark, AttachmentIDs: req.AttachmentIDs,
		InstallmentPlans: toPlanInputs(req.InstallmentPlans),
	}
	acc, err := a.accounts.Create(middleware.GetUserID(c), in)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	if a.attachments != nil {
		a.attachments.RefreshAccountAttachments(acc)
	}
	OK(c, acc)
}

func (a *API) UpdateAccount(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name             string                 `json:"name"`
		Type             string                 `json:"type"`
		Sort             int                    `json:"sort"`
		Archived         *bool                  `json:"archived"`
		Balance          *int64                 `json:"balance"`
		UsedCredit       *int64                 `json:"usedCredit"` // 信用：已用额度（正）
		CreditLimit      *int64                 `json:"creditLimit"`
		CreditBilledFen  *int64                 `json:"creditBilledFen"`
		BillingDay       *int                   `json:"billingDay"`
		PaymentDueDay    *int                   `json:"paymentDueDay"`
		Institution      *string                `json:"institution"`
		CardNo           *string                `json:"cardNo"`
		HolderName       *string                `json:"holderName"`
		StorageNote      *string                `json:"storageNote"`
		Remark           *string                `json:"remark"`
		AttachmentIDs    []uint64               `json:"attachmentIds"`
		LastReconciledAt *string                `json:"lastReconciledAt"`
		ClearReconciled  bool                   `json:"clearReconciled"`
		InstallmentPlans *[]installmentPlanBody `json:"installmentPlans"`
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
		ClearReconciled:  req.ClearReconciled,
		InstallmentPlans: toPlanInputs(req.InstallmentPlans),
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
	if a.attachments != nil {
		a.attachments.RefreshAccountAttachments(acc)
	}
	OK(c, acc)
}

// AccountSensitive 显式拉取卡号/户名明文（前端不常驻保留）
func (a *API) AccountSensitive(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		BadRequest(c, "请输入登录密码")
		return
	}
	uid := middleware.GetUserID(c)
	if err := a.auth.VerifyPassword(c.Request.Context(), uid, req.Password, c.ClientIP()); err != nil {
		var tooMany *service.TooManyAttemptsError
		if errors.As(err, &tooMany) {
			c.Header("Retry-After", strconv.Itoa(int(tooMany.RetryAfter.Seconds())))
			Fail(c, http.StatusTooManyRequests, err.Error())
			return
		}
		BadRequest(c, err.Error())
		return
	}
	cardNo, holder, err := a.accounts.GetSensitive(uid, id)
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

// AccountLedger 账户流水明细（按月分组，信用卡带出账/还款状态）
func (a *API) AccountLedger(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ledger, err := a.accounts.Ledger(middleware.GetUserID(c), id, time.Now())
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, ledger)
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
