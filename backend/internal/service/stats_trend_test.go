package service

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mybookkeeping/internal/model"
)

// 月度趋势在内存分月汇总，不依赖 MySQL 的 DATE_FORMAT，SQLite 下同样可用
func TestSummaryTrendsByMonth(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:TestSummaryTrendsByMonth?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Category{}, &model.Transaction{}); err != nil {
		t.Fatal(err)
	}
	u := &model.User{Username: "trend", PasswordHash: "x"}
	db.Create(u)
	loc := time.Local
	add := func(typ string, amount int64, at time.Time) {
		if err := db.Create(&model.Transaction{UserID: u.ID, Type: typ, Amount: amount, AccountID: 1, HappenedAt: at}).Error; err != nil {
			t.Fatal(err)
		}
	}
	add("expense", 100, time.Date(2026, 8, 31, 23, 59, 0, 0, loc))
	add("expense", 50, time.Date(2026, 9, 1, 0, 1, 0, 0, loc))
	add("income", 300, time.Date(2026, 9, 15, 12, 0, 0, 0, loc))
	add("expense", 20, time.Date(2026, 9, 30, 23, 0, 0, 0, loc))
	add("transfer", 999, time.Date(2026, 9, 10, 12, 0, 0, 0, loc))

	out, err := NewStatsService(db).Summary(u.ID, StatsQuery{
		From: time.Date(2026, 8, 1, 0, 0, 0, 0, loc),
		To:   time.Date(2026, 10, 1, 0, 0, 0, 0, loc),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]MonthTrend{}
	for _, m := range out.Trends {
		got[m.Month] = m
	}
	for _, w := range []MonthTrend{{Month: "2026-08", Expense: 100}, {Month: "2026-09", Expense: 70, Income: 300}} {
		if got[w.Month] != w {
			t.Fatalf("trend %s = %+v, want %+v (all: %+v)", w.Month, got[w.Month], w, out.Trends)
		}
	}
}
