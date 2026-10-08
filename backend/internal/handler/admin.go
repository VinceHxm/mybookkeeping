package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

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

func (a *API) AdminResetPassword(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	if err := a.auth.AdminResetPassword(c.Request.Context(), middleware.GetUserID(c), id, req.NewPassword); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"ok": true})
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
