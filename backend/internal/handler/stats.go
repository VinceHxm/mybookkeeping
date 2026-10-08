package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

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
