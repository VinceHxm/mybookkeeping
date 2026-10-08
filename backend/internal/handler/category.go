package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

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
