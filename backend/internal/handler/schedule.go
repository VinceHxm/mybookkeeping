package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

type scheduleBody struct {
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
	AnchorDay   int      `json:"anchorDay"`
	Enabled     *bool    `json:"enabled"`
	NextRunAt   string   `json:"nextRunAt"`
	EndAt       string   `json:"endAt"`

	AmountPending   bool   `json:"amountPending"`
	ForeignCurrency string `json:"foreignCurrency"`
	ForeignAmount   int64  `json:"foreignAmount"`
}

func (b scheduleBody) toInput() service.ScheduleInput {
	in := service.ScheduleInput{
		Name: b.Name, Type: b.Type, Amount: b.Amount, AccountID: b.AccountID,
		ToAccountID: b.ToAccountID, CategoryID: b.CategoryID, Remark: b.Remark,
		TagIDs: b.TagIDs, Frequency: b.Frequency, IntervalN: b.IntervalN, AnchorDay: b.AnchorDay, Enabled: b.Enabled,
		AmountPending: b.AmountPending, ForeignCurrency: b.ForeignCurrency, ForeignAmount: b.ForeignAmount,
		EndAt: parseTimePtr(b.EndAt),
	}
	if t := parseTimePtr(b.NextRunAt); t != nil {
		in.NextRunAt = *t
	}
	return in
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
	var body scheduleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	sch, err := a.schedules.Create(middleware.GetUserID(c), body.toInput())
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, sch)
}

func (a *API) UpdateSchedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body scheduleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	sch, err := a.schedules.Update(middleware.GetUserID(c), id, body.toInput())
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
