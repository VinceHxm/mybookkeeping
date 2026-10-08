package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
)

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
