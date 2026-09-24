package main

import (
	"context"
	"crypto/rand"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/config"
	"mybookkeeping/internal/database"
	"mybookkeeping/internal/handler"
	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
	"mybookkeeping/internal/storage"
)

func main() {
	cfg := config.Load()

	db, err := database.OpenMySQL(cfg)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}

	rdb, err := database.OpenRedis(cfg)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	var minioStore *storage.MinIO
	minioReady := false
	if ms, err := storage.NewMinIO(cfg); err != nil {
		log.Printf("warn: minio unavailable: %v (attachments disabled)", err)
	} else {
		minioStore = ms
		minioReady = true
	}

	authSvc := service.NewAuthService(db, rdb, cfg.SessionTTL, cfg.ResetTokenTTLHours, cfg.ResetTokenInResponse)
	if err := authSvc.EnsureAdminRoles(); err != nil {
		log.Printf("warn: ensure admin roles: %v", err)
	}
	accountSvc := service.NewAccountService(db)
	categorySvc := service.NewCategoryService(db)
	tagSvc := service.NewTagService(db)
	holidaySvc := service.NewHolidayService(db, rdb)
	fareRuleSvc := service.NewFareRuleService(db, holidaySvc)
	if err := fareRuleSvc.MigrateEmbeddedTemplateFareRules(); err != nil {
		log.Printf("warn: migrate fare rules: %v", err)
	}
	templateSvc := service.NewTemplateService(db, fareRuleSvc)
	txSvc := service.NewTransactionService(db)
	scheduleSvc := service.NewScheduleService(db, txSvc)
	statsSvc := service.NewStatsService(db)
	llmSvc := service.NewLLMService(cfg, db)
	var attachSvc *service.AttachmentService
	if minioReady {
		attachSvc = service.NewAttachmentService(db, minioStore, appSecret(cfg))
	}

	api := handler.NewAPI(cfg, authSvc, accountSvc, categorySvc, tagSvc, templateSvc, fareRuleSvc, holidaySvc, scheduleSvc, txSvc, attachSvc, statsSvc, llmSvc, minioReady)
	authMw := middleware.NewAuth(rdb)

	// 周期记账后台任务
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if n, err := scheduleSvc.RunDue(time.Now()); err != nil {
				log.Printf("schedule run: %v", err)
			} else if n > 0 {
				log.Printf("schedule created %d transaction(s)", n)
			}
		}
	}()

	// 法定节假日：仅 12 月第三周自动同步入库（每小时检查一次）
	go func() {
		run := func() {
			if err := holidaySvc.TryAutoSync(context.Background(), time.Now()); err != nil {
				log.Printf("holiday auto sync: %v", err)
			}
		}
		run()
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Fatalf("TRUSTED_PROXIES: %v", err)
	}
	r.Use(securityHeaders(), corsMiddleware(cfg.CORSOrigins))

	r.GET("/api/health", api.Health)
	r.POST("/api/auth/login", api.Login)
	r.POST("/api/auth/register", api.Register)
	r.POST("/api/auth/forgot-password", api.ForgotPassword)
	r.POST("/api/auth/reset-password", api.ResetPassword)
	// 附件凭签名短链访问（<img> 无法携带 Authorization 头）
	r.GET("/api/attachments/:id/file", api.ServeAttachment)

	auth := r.Group("/api")
	auth.Use(authMw.Required())
	{
		auth.POST("/auth/logout", api.Logout)
		auth.GET("/me", api.Me)
		auth.PUT("/me/password", api.ChangePassword)
		auth.PUT("/me/settings", api.UpdateSettings)

		auth.GET("/accounts", api.ListAccounts)
		auth.POST("/accounts", api.CreateAccount)
		auth.PUT("/accounts/:id", api.UpdateAccount)
		auth.DELETE("/accounts/:id", api.DeleteAccount)
		auth.GET("/accounts/credit-repay-reminders", api.CreditRepayReminders)
		auth.GET("/accounts/:id/credit-statement", api.CreditStatement)
		auth.GET("/accounts/:id/sensitive", api.AccountSensitive)

		auth.GET("/categories", api.ListCategories)
		auth.POST("/categories", api.CreateCategory)
		auth.PUT("/categories/:id", api.UpdateCategory)
		auth.DELETE("/categories/:id", api.DeleteCategory)

		auth.GET("/tags", api.ListTags)
		auth.POST("/tags", api.CreateTag)
		auth.PUT("/tags/:id", api.UpdateTag)
		auth.DELETE("/tags/:id", api.DeleteTag)

		auth.GET("/templates", api.ListTemplates)
		auth.POST("/templates", api.CreateTemplate)
		auth.PUT("/templates/:id", api.UpdateTemplate)
		auth.DELETE("/templates/:id", api.DeleteTemplate)
		auth.POST("/templates/:id/preview-fare", api.PreviewTemplateFare)

		auth.GET("/fare-rules", api.ListFareRules)
		auth.POST("/fare-rules", api.CreateFareRule)
		auth.PUT("/fare-rules/:id", api.UpdateFareRule)
		auth.DELETE("/fare-rules/:id", api.DeleteFareRule)
		auth.POST("/fare-rules/:id/preview", api.PreviewFareRule)
		auth.GET("/holidays/:year", api.GetHolidays)
		auth.POST("/holidays/:year/refresh", middleware.RequireAdmin(authSvc.GetUser), api.RefreshHolidays)

		admin := auth.Group("/admin")
		admin.Use(middleware.RequireAdmin(authSvc.GetUser))
		{
			admin.GET("/users", api.AdminListUsers)
			admin.PUT("/users/:id", api.AdminUpdateUser)
			admin.PUT("/users/:id/password", api.AdminResetPassword)
			admin.DELETE("/users/:id", api.AdminDeleteUser)
		}

		auth.GET("/schedules", api.ListSchedules)
		auth.POST("/schedules", api.CreateSchedule)
		auth.PUT("/schedules/:id", api.UpdateSchedule)
		auth.DELETE("/schedules/:id", api.DeleteSchedule)

		auth.GET("/transactions", api.ListTransactions)
		auth.GET("/transactions/:id", api.GetTransaction)
		auth.POST("/transactions", api.CreateTransaction)
		auth.PUT("/transactions/:id", api.UpdateTransaction)
		auth.DELETE("/transactions/:id", api.DeleteTransaction)

		auth.POST("/attachments", api.UploadAttachment)
		auth.GET("/stats/summary", api.StatsSummary)
		auth.POST("/ai/recognize-text", api.RecognizeText)
		auth.POST("/ai/recognize-text-batch", api.RecognizeTextBatch)
		auth.POST("/ai/recognize-image", api.RecognizeImage)
		auth.GET("/amap/config", api.AmapConfig)
	}

	serveFrontend(r)

	log.Printf("listening on %s (ai=%v register=%v)", cfg.ServerAddr, llmSvc.Enabled(), cfg.AllowRegister)
	if err := r.Run(cfg.ServerAddr); err != nil {
		log.Fatal(err)
	}
}

func appSecret(cfg *config.Config) []byte {
	if cfg.AppSecret != "" {
		return []byte(cfg.AppSecret)
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("generate app secret: %v", err)
	}
	log.Printf("warn: APP_SECRET 未配置，已随机生成（重启后旧的附件链接失效，刷新页面即可）")
	return b
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "same-origin")
		c.Next()
	}
}

// corsMiddleware 仅对白名单源放行；未配置时不下发任何 CORS 头（同源部署与 Vite 代理都不需要）
func corsMiddleware(origins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(origins))
	for _, o := range origins {
		allowed[strings.TrimRight(o, "/")] = true
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
		}
		c.Next()
	}
}

func serveFrontend(r *gin.Engine) {
	public := filepath.Join(".", "public")
	if _, err := os.Stat(public); err != nil {
		return
	}
	r.Static("/assets", filepath.Join(public, "assets"))
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(public, "index.html"))
	})
}
