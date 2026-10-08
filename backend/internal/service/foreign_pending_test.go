package service

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mybookkeeping/internal/model"
)

func TestNormalizeForeign(t *testing.T) {
	cases := []struct {
		name    string
		typ     string
		cur     string
		amt     int64
		wantCur string
		wantAmt int64
		wantErr bool
	}{
		{"未填币种清空金额", "expense", "", 2000, "", 0, false},
		{"小写转大写", "expense", " usd ", 2000, "USD", 2000, false},
		{"转账忽略外币", "transfer", "USD", 2000, "", 0, false},
		{"人民币拒绝", "expense", "cny", 2000, "", 0, true},
		{"金额缺失", "income", "EUR", 0, "", 0, true},
		{"非字母", "expense", "US1", 2000, "", 0, true},
		{"过短", "expense", "US", 2000, "", 0, true},
	}
	for _, c := range cases {
		cur, amt, err := normalizeForeign(c.typ, c.cur, c.amt)
		if (err != nil) != c.wantErr {
			t.Fatalf("%s: err=%v wantErr=%v", c.name, err, c.wantErr)
		}
		if cur != c.wantCur || amt != c.wantAmt {
			t.Fatalf("%s: got %s/%d want %s/%d", c.name, cur, amt, c.wantCur, c.wantAmt)
		}
	}
}

func TestScheduleRunDueCarriesPendingAndForeign(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Account{}, &model.Category{}, &model.Tag{},
		&model.Transaction{}, &model.TransactionTag{}, &model.Attachment{}, &model.Schedule{},
	); err != nil {
		t.Fatal(err)
	}
	u := &model.User{Username: "u", PasswordHash: "x"}
	db.Create(u)
	acc := &model.Account{UserID: u.ID, Name: "信用卡", Type: "credit", CreditLimit: 100000}
	cat := &model.Category{UserID: u.ID, Name: "会员", Kind: "expense"}
	db.Create(acc)
	db.Create(cat)

	txSvc := NewTransactionService(db)
	schSvc := NewScheduleService(db, txSvc)
	start := time.Now().Add(-time.Minute)
	if _, err := schSvc.Create(u.ID, ScheduleInput{
		Name: "ChatGPT", Type: "expense", Amount: 14500, AccountID: acc.ID, CategoryID: &cat.ID,
		Frequency: "monthly", NextRunAt: start,
		AmountPending: true, ForeignCurrency: "usd", ForeignAmount: 2000,
	}); err != nil {
		t.Fatal(err)
	}
	if n, err := schSvc.RunDue(time.Now()); err != nil || n != 1 {
		t.Fatalf("RunDue n=%d err=%v", n, err)
	}

	res, err := txSvc.List(u.ID, TransactionQuery{Pending: true})
	if err != nil || res.Total != 1 {
		t.Fatalf("pending list total=%d err=%v", res.Total, err)
	}
	got := res.Items[0]
	if !got.AmountPending || got.ForeignCurrency != "USD" || got.ForeignAmount != 2000 || got.Amount != 14500 {
		t.Fatalf("unexpected tx: %+v", got)
	}

	// 按账单实际金额确认：清除待确认，余额按新金额重算
	if _, err := txSvc.Update(u.ID, got.ID, TransactionInput{
		Type: "expense", Amount: 14620, AccountID: acc.ID, CategoryID: &cat.ID, HappenedAt: got.HappenedAt,
		ForeignCurrency: "USD", ForeignAmount: 2000,
	}); err != nil {
		t.Fatal(err)
	}
	res, _ = txSvc.List(u.ID, TransactionQuery{Pending: true})
	if res.Total != 0 {
		t.Fatalf("still pending: %d", res.Total)
	}
	var a model.Account
	db.First(&a, acc.ID)
	if a.Balance != -14620 {
		t.Fatalf("balance=%d want -14620", a.Balance)
	}
}
