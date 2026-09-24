package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mybookkeeping/internal/config"
	"mybookkeeping/internal/model"
)

func OpenMySQL(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("mysql connect: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

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
