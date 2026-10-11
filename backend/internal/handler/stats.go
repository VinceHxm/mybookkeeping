package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/i18n"
	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

func parseStatsQuery(c *gin.Context) service.StatsQuery {
	q := service.StatsQuery{}
	if t := parseTimePtr(c.Query("from")); t != nil {
		q.From = *t
	}
	if t := parseTimePtr(c.Query("to")); t != nil {
		q.To = *t
	}
	if v := c.Query("accountId"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil && n > 0 {
			q.AccountID = &n
		}
	}
	return q
}

func (a *API) StatsSummary(c *gin.Context) {
	sum, err := a.stats.Summary(middleware.GetUserID(c), parseStatsQuery(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	loc := i18n.FromRequest(c)
	for i := range sum.ByCategory {
		if sum.ByCategory[i].CategoryID == 0 {
			sum.ByCategory[i].CategoryName = i18n.T(loc, sum.ByCategory[i].CategoryName)
		}
	}
	OK(c, sum)
}

func (a *API) StatsOverview(c *gin.Context) {
	out, err := a.stats.Overview(middleware.GetUserID(c), parseStatsQuery(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, out)
}

func (a *API) StatsBreakdown(c *gin.Context) {
	q := service.BreakdownQuery{
		StatsQuery: parseStatsQuery(c),
		Kind:       c.Query("kind"),
		By:         c.Query("by"),
	}
	if v := c.Query("parentId"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil && n > 0 {
			q.ParentID = &n
		}
	}
	out, err := a.stats.Breakdown(middleware.GetUserID(c), q)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	// 只翻 service 生成的占位名：ID 0 为「未分类 / 无标签」，下钻时与父类同 ID 的项为「xx（未细分）」；其余是用户数据
	loc := i18n.FromRequest(c)
	for i := range out.Items {
		it := &out.Items[i]
		if it.ID == 0 || (out.By == "category" && q.ParentID != nil && it.ID == *q.ParentID) {
			it.Name = i18n.T(loc, it.Name)
		}
	}
	OK(c, out)
}

func (a *API) StatsTrend(c *gin.Context) {
	out, err := a.stats.Trend(middleware.GetUserID(c), parseStatsQuery(c), c.Query("granularity"))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, out)
}

func (a *API) StatsAssets(c *gin.Context) {
	out, err := a.stats.AssetTrend(middleware.GetUserID(c), parseStatsQuery(c), c.Query("granularity"))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, out)
}

func (a *API) StatsPlaces(c *gin.Context) {
	out, err := a.stats.Places(middleware.GetUserID(c), parseStatsQuery(c), c.Query("kind"))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, out)
}

func (a *API) StatsTop(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	out, err := a.stats.TopTransactions(middleware.GetUserID(c), parseStatsQuery(c), c.Query("kind"), limit)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, out)
}
