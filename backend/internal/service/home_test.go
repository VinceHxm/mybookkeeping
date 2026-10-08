package service

import (
	"testing"
	"time"

	"mybookkeeping/internal/model"
)

func TestHomeQueries(t *testing.T) {
	f := newTenantFixture(t)
	svc := NewTransactionService(f.db)
	mk := func(tx *model.Transaction) {
		t.Helper()
		tx.UserID = f.a
		tx.AccountID = f.aAcc
		if tx.Type == "" {
			tx.Type = "expense"
		}
		if tx.HappenedAt.IsZero() {
			tx.HappenedAt = time.Now()
		}
		if err := f.db.Create(tx).Error; err != nil {
			t.Fatal(err)
		}
	}

	if n, first, err := svc.PendingSummary(f.a); err != nil || n != 0 || first != nil {
		t.Fatalf("empty pending: n=%d first=%+v err=%v", n, first, err)
	}
	mk(&model.Transaction{Amount: 100, CategoryID: &f.aCat, AmountPending: true, Remark: "电费"})
	n, first, err := svc.PendingSummary(f.a)
	if err != nil || n != 1 || first == nil || first.Remark != "电费" || first.Category == nil || first.Category.Name != "A 餐饮" {
		t.Fatalf("one pending: n=%d first=%+v err=%v", n, first, err)
	}
	mk(&model.Transaction{Amount: 200, AmountPending: true})
	if n, first, _ := svc.PendingSummary(f.a); n != 2 || first != nil {
		t.Fatalf("two pending: n=%d first=%+v", n, first)
	}

	mk(&model.Transaction{Amount: 1, VoucherStatus: VoucherPending, VoucherExpireOn: "2026-01-02"})
	mk(&model.Transaction{Amount: 1, VoucherStatus: VoucherPending, VoucherExpireOn: "2026-02-01"})
	mk(&model.Transaction{Amount: 1, VoucherStatus: VoucherPending})
	mk(&model.Transaction{Amount: 1, VoucherStatus: VoucherUsed, VoucherExpireOn: "2026-01-01"})
	if n, err := svc.CountExpiringVouchers(f.a, "2026-01-05"); err != nil || n != 1 {
		t.Fatalf("expiring vouchers n=%d err=%v", n, err)
	}
	if n, _ := svc.CountExpiringVouchers(f.b, "2099-01-01"); n != 0 {
		t.Fatalf("B 不应数到 A 的券: %d", n)
	}

	lng, lat := 120.1, 30.2
	mk(&model.Transaction{Amount: 1, GeoMode: "point", GeoLng: &lng, GeoLat: &lat, GeoName: "西湖", HappenedAt: time.Now().Add(time.Hour)})
	mk(&model.Transaction{Amount: 1, GeoLng: &lng, GeoLat: &lat})
	places, err := svc.RecentPlaces(f.a, 0)
	if err != nil || len(places) != 1 || places[0].GeoName != "西湖" || *places[0].GeoLng != lng {
		t.Fatalf("recent places=%+v err=%v", places, err)
	}
	if places, _ := svc.RecentPlaces(f.b, 10); len(places) != 0 {
		t.Fatalf("B 不应看到 A 的地点: %+v", places)
	}

	recent, err := svc.Recent(f.a, 3)
	if err != nil || len(recent) != 3 || recent[0].GeoName != "西湖" {
		t.Fatalf("recent=%d err=%v", len(recent), err)
	}
}

func TestTransactionPreloadHidesAccountSecrets(t *testing.T) {
	f := newTenantFixture(t)
	svc := NewTransactionService(f.db)
	tx, err := svc.Get(f.b, f.bTx)
	if err != nil {
		t.Fatal(err)
	}
	if tx.Account == nil || tx.Account.Name == "" {
		t.Fatalf("account brief missing: %+v", tx.Account)
	}
	if tx.Account.CardNo != "" || tx.Account.HolderName != "" {
		t.Errorf("流水接口不应带出卡号/持卡人: %+v", tx.Account)
	}
	res, err := svc.List(f.b, TransactionQuery{})
	if err != nil || len(res.Items) == 0 {
		t.Fatalf("list err=%v", err)
	}
	for _, it := range res.Items {
		if it.Account != nil && it.Account.CardNo != "" {
			t.Errorf("列表不应带出卡号: %+v", it.Account)
		}
	}
}
