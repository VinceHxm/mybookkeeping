package handler

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/model"
	"mybookkeeping/internal/service"
)

// homeConcurrency 首页各块并行取数的上限：一次打开首页最多同时占用这么多数据库连接
const homeConcurrency = 3

const homeTemplateCount = 9

const (
	homeRecentDefault = 5
	homeRecentMax     = 20
	homeTopCategories = 5
	homeWeekDays      = 7
)

type homeOverview struct {
	Templates        []model.Template              `json:"templates"`
	Recent           []model.Transaction           `json:"recent"`
	Summary          *service.Summary              `json:"summary"`
	RepayReminders   []service.CreditRepayReminder `json:"repayReminders"`
	PendingCount     int64                         `json:"pendingCount"`
	PendingFirst     *model.Transaction            `json:"pendingFirst"`
	ExpiringVouchers int64                         `json:"expiringVouchers"`
	Insights         *homeInsights                 `json:"insights,omitempty"`
}

// homeInsights 宽屏首页的概览卡：本月支出分类 Top N + 近 7 天每日收支
type homeInsights struct {
	Categories *service.Breakdown   `json:"categories"`
	Week       []service.TrendPoint `json:"week"`
}

// HomeOverview 首页一次取齐：模板、最近流水、本月汇总、还款提醒、待确认、快过期团购券。
// 可选参数：recent=最近流水条数（默认 5，最多 20）；insights=1 时附带概览卡数据（宽屏用）。
// 某块失败只记日志并留空，不影响其余块。
func (a *API) HomeOverview(c *gin.Context) {
	uid := middleware.GetUserID(c)
	now := time.Now()
	out := homeOverview{
		Templates:      []model.Template{},
		Recent:         []model.Transaction{},
		RepayReminders: []service.CreditRepayReminder{},
	}
	recentN := homeRecentDefault
	if n, err := strconv.Atoi(c.Query("recent")); err == nil && n > 0 {
		recentN = min(n, homeRecentMax)
	}

	tasks := map[string]func() error{
		"templates": func() error {
			list, err := a.templates.List(uid)
			if len(list) > homeTemplateCount {
				list = list[:homeTemplateCount]
			}
			if list != nil {
				out.Templates = list
			}
			return err
		},
		"recent": func() error {
			list, err := a.transactions.Recent(uid, recentN)
			if err == nil {
				if a.attachments != nil {
					a.attachments.RefreshTxListAttachments(list)
				}
				out.Recent = list
			}
			return err
		},
		"summary": func() error {
			sum, err := a.stats.Summary(uid, service.StatsQuery{})
			out.Summary = sum
			return err
		},
		"repayReminders": func() error {
			list, err := a.accounts.CreditRepayReminders(uid, now)
			if err == nil {
				out.RepayReminders = list
			}
			return err
		},
		"pending": func() error {
			n, first, err := a.transactions.PendingSummary(uid)
			out.PendingCount, out.PendingFirst = n, first
			return err
		},
		"vouchers": func() error {
			n, err := a.transactions.CountExpiringVouchers(uid, now.AddDate(0, 0, 3).Format("2006-01-02"))
			out.ExpiringVouchers = n
			return err
		},
	}
	if c.Query("insights") == "1" {
		ins := &homeInsights{Week: []service.TrendPoint{}}
		out.Insights = ins
		tasks["insightCategories"] = func() error {
			b, err := a.stats.Breakdown(uid, service.BreakdownQuery{Kind: "expense", By: "category"})
			if err == nil {
				if len(b.Items) > homeTopCategories {
					b.Items = b.Items[:homeTopCategories]
				}
				ins.Categories = b
			}
			return err
		}
		tasks["insightWeek"] = func() error {
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
			t, err := a.stats.Trend(uid, service.StatsQuery{
				From: today.AddDate(0, 0, 1-homeWeekDays),
				To:   today.AddDate(0, 0, 1),
			}, service.GranularityDay)
			if err == nil {
				ins.Week = t.Points
			}
			return err
		}
	}

	sem := make(chan struct{}, homeConcurrency)
	var wg sync.WaitGroup
	for name, fn := range tasks {
		wg.Add(1)
		go func(name string, fn func() error) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := fn(); err != nil {
				log.Printf("home overview %s (user %d): %v", name, uid, err)
			}
		}(name, fn)
	}
	wg.Wait()
	OK(c, out)
}
