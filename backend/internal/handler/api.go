package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/config"
	"mybookkeeping/internal/service"
)

type API struct {
	cfg          *config.Config
	auth         *service.AuthService
	accounts     *service.AccountService
	categories   *service.CategoryService
	tags         *service.TagService
	templates    *service.TemplateService
	fareRules    *service.FareRuleService
	holidays     *service.HolidayService
	schedules    *service.ScheduleService
	transactions *service.TransactionService
	attachments  *service.AttachmentService
	stats        *service.StatsService
	llm          *service.LLMService
	minioReady   bool
}

func NewAPI(
	cfg *config.Config,
	auth *service.AuthService,
	accounts *service.AccountService,
	categories *service.CategoryService,
	tags *service.TagService,
	templates *service.TemplateService,
	fareRules *service.FareRuleService,
	holidays *service.HolidayService,
	schedules *service.ScheduleService,
	transactions *service.TransactionService,
	attachments *service.AttachmentService,
	stats *service.StatsService,
	llm *service.LLMService,
	minioReady bool,
) *API {
	return &API{
		cfg: cfg, auth: auth, accounts: accounts, categories: categories,
		tags: tags, templates: templates, fareRules: fareRules, holidays: holidays,
		schedules: schedules, transactions: transactions, attachments: attachments,
		stats: stats, llm: llm, minioReady: minioReady,
	}
}

func (a *API) Health(c *gin.Context) {
	OK(c, gin.H{
		"status":   "up",
		"minio":    a.minioReady,
		"ai":       a.llm != nil && a.llm.Enabled(),
		"register": a.cfg.AllowRegister,
	})
}

func parseTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	layouts := []string{time.RFC3339, "2006-01-02", "2006-01-02 15:04:05", "2006-01-02T15:04"}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
			return &t
		}
	}
	return nil
}

func (a *API) AmapConfig(c *gin.Context) {
	OK(c, gin.H{
		"key":            a.cfg.AmapKey,
		"securityJsCode": a.cfg.AmapSecurityCode,
	})
}
