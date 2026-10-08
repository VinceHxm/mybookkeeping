package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

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
