package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/i18n"
	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/model"
	"mybookkeeping/internal/service"
)

func (a *API) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		BadRequest(c, "请输入用户名和密码")
		return
	}
	token, user, err := a.auth.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP())
	var tooMany *service.TooManyAttemptsError
	if errors.As(err, &tooMany) {
		c.Header("Retry-After", strconv.Itoa(int(tooMany.RetryAfter.Seconds())))
		Fail(c, http.StatusTooManyRequests, err.Error())
		return
	}
	if errors.Is(err, service.ErrInvalidCredentials) {
		Fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	if errors.Is(err, service.ErrUserDisabled) {
		FailCode(c, http.StatusForbidden, "user_disabled", err.Error())
		return
	}
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	a.ensureDefaults(c, user)
	OK(c, gin.H{"token": token, "user": user})
}

// ensureDefaults 默认分类跟随界面语言（老账号未设语言时取本次请求语言）；
// 默认标签按注册语言生成，老账号 PresetLocale 为空即当初按中文生成。
func (a *API) ensureDefaults(c *gin.Context, u *model.User) {
	uiLocale := u.Locale
	if uiLocale == "" {
		uiLocale = i18n.FromRequest(c)
	}
	_ = a.categories.EnsureDefaults(u.ID, uiLocale)
	tagLocale := u.PresetLocale
	if tagLocale == "" {
		tagLocale = i18n.ZhCN
	}
	_ = a.tags.EnsureDefaults(u.ID, tagLocale)
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
	user, err := a.auth.Register(req.Username, req.Password, req.Email, i18n.FromRequest(c))
	if errors.Is(err, service.ErrUserExists) || errors.Is(err, service.ErrWeakPassword) {
		BadRequest(c, err.Error())
		return
	}
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	a.ensureDefaults(c, user)
	// 注册后直接登录
	token, _, err := a.auth.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP())
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
	res.Message = i18n.T(i18n.FromRequest(c), res.Message)
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
		FailCode(c, http.StatusForbidden, "user_disabled", service.ErrUserDisabled.Error())
		return
	}
	a.ensureDefaults(c, user)
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
	err := a.auth.ChangePassword(c.Request.Context(), middleware.GetUserID(c), req.OldPassword, req.NewPassword, c.GetString("sessionToken"))
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"ok": true})
}

func (a *API) UpdateSettings(c *gin.Context) {
	var req struct {
		Email                   *string `json:"email"`
		DefaultAccountID        *uint64 `json:"defaultAccountId"`
		ClearDefaultAcc         bool    `json:"clearDefaultAccount"`
		DefaultExpenseAccountID *uint64 `json:"defaultExpenseAccountId"`
		ClearDefaultExpenseAcc  bool    `json:"clearDefaultExpenseAccount"`
		WeekStart               *int    `json:"weekStart"`
		ExpenseColor            *string `json:"expenseColor"`
		IncomeColor             *string `json:"incomeColor"`
		Theme                   *string `json:"theme"`
		Locale                  *string `json:"locale"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	in := service.UserSettingsInput{
		Email: req.Email, WeekStart: req.WeekStart,
		ExpenseColor: req.ExpenseColor, IncomeColor: req.IncomeColor, Theme: req.Theme,
		Locale: req.Locale,
	}
	if req.ClearDefaultAcc {
		var nilID *uint64
		in.DefaultAccountID = &nilID
	} else if req.DefaultAccountID != nil {
		id := req.DefaultAccountID
		in.DefaultAccountID = &id
	}
	if req.ClearDefaultExpenseAcc {
		var nilID *uint64
		in.DefaultExpenseAccountID = &nilID
	} else if req.DefaultExpenseAccountID != nil {
		id := req.DefaultExpenseAccountID
		in.DefaultExpenseAccountID = &id
	}
	user, err := a.auth.UpdateSettings(middleware.GetUserID(c), in)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	if req.Locale != nil {
		_ = a.categories.EnsureDefaults(user.ID, user.Locale)
	}
	OK(c, user)
}
