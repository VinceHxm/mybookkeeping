package database

import (
	"path/filepath"
	"testing"
	"time"

	"mybookkeeping/internal/model"
)

// 同一时刻用 UTC 与本地时区分别写入/查询，范围比较结果必须一致
func TestSQLiteTimeParamsNormalized(t *testing.T) {
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	u := &model.User{Username: "tz", PasswordHash: "x"}
	if err := db.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 10, 1, 0, 30, 0, 0, time.FixedZone("CST", 8*3600))
	tx := &model.Transaction{UserID: u.ID, Type: "expense", Amount: 100, HappenedAt: at.UTC()}
	if err := db.Create(tx).Error; err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := db.Model(&model.Transaction{}).
		Where("happened_at >= ? AND happened_at < ?", at.Add(-time.Minute), at.Add(time.Minute)).
		Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("range query matched %d rows, want 1", n)
	}
	var got model.Transaction
	if err := db.First(&got, tx.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !got.HappenedAt.Equal(at) {
		t.Fatalf("round trip: got %v want %v", got.HappenedAt, at)
	}
}
