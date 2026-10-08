package database

import (
	"context"
	"fmt"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mybookkeeping/internal/config"
	"mybookkeeping/internal/model"
)

// 远程库建连偶发丢包时，系统默认要等 ~21s 才报错；超时尽快失败，由连接池换一条重试
const (
	mysqlDialTimeout = 5 * time.Second
	mysqlIOTimeout   = 30 * time.Second
	mysqlMaxConns    = 20
)

// withTimeouts DSN 未显式配置时补上建连 / 读写超时
func withTimeouts(dsn string) string {
	c, err := mysqldriver.ParseDSN(dsn)
	if err != nil {
		return dsn
	}
	if c.Timeout == 0 {
		c.Timeout = mysqlDialTimeout
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = mysqlIOTimeout
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = mysqlIOTimeout
	}
	return c.FormatDSN()
}

func OpenMySQL(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(withTimeouts(cfg.MySQLDSN)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("mysql connect: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// 空闲上限与最大连接一致：并发请求复用已建好的连接，不必临时向远程库新建
	sqlDB.SetMaxOpenConns(mysqlMaxConns)
	sqlDB.SetMaxIdleConns(mysqlMaxConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	// 闲置过久的连接可能已被中间网络设备静默断开，提前回收
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := db.AutoMigrate(
		&model.User{},
		&model.Account{},
		&model.CreditInstallmentPlan{},
		&model.Category{},
		&model.Tag{},
		&model.Transaction{},
		&model.TransactionTag{},
		&model.Attachment{},
		&model.FareRule{},
		&model.Template{},
		&model.Schedule{},
		&model.HolidayYearRecord{},
		&model.AppMeta{},
	); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}
	// 兜底：确保图标字段存在（旧进程未重启时也能在下次启动补齐）
	if !db.Migrator().HasColumn(&model.Template{}, "icon") {
		if err := db.Migrator().AddColumn(&model.Template{}, "Icon"); err != nil {
			return nil, fmt.Errorf("add templates.icon: %w", err)
		}
	}
	if !db.Migrator().HasColumn(&model.FareRule{}, "icon") {
		if err := db.Migrator().AddColumn(&model.FareRule{}, "Icon"); err != nil {
			return nil, fmt.Errorf("add fare_rules.icon: %w", err)
		}
	}
	return db, nil
}

func OpenRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return rdb, nil
}
