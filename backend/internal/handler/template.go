package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

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
