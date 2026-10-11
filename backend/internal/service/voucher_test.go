package service

import (
	"testing"
	"time"

	"mybookkeeping/internal/model"
)

func TestVoucherUseAndRefundFlow(t *testing.T) {
	f := newTenantFixture(t)
	svc := NewTransactionService(f.db)

	v, err := svc.Create(f.a, TransactionInput{
		Type: "expense", Amount: 5000, AccountID: f.aAcc, CategoryID: &f.aCat, Remark: "火锅双人券",
		VoucherStatus: VoucherPending, VoucherExpireOn: "2026-12-31", HappenedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if v.VoucherStatus != VoucherPending || v.VoucherExpireOn != "2026-12-31" {
		t.Fatalf("voucher=%q expire=%q", v.VoucherStatus, v.VoucherExpireOn)
	}
	if res, _ := svc.List(f.a, TransactionQuery{Voucher: VoucherPending}); res.Total != 1 {
		t.Fatalf("pending filter total=%d", res.Total)
	}

	if _, err := svc.SetVoucherStatus(f.a, v.ID, VoucherUsed); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetVoucherStatus(f.a, v.ID, VoucherPending); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.RefundVoucher(f.a, v.ID, VoucherRefundInput{Amount: 6000}); err == nil {
		t.Error("退款金额超过原金额应失败")
	}
	r, err := svc.RefundVoucher(f.a, v.ID, VoucherRefundInput{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Type != "income" || r.Amount != 5000 || r.AccountID != f.aAcc || r.RefundOfID == nil || *r.RefundOfID != v.ID {
		t.Fatalf("refund=%+v", r)
	}
	if r.Category == nil || r.Category.Name != refundCategoryName || r.Category.Kind != "income" {
		t.Fatalf("refund category=%+v", r.Category)
	}
	var acc model.Account
	f.db.First(&acc, f.aAcc)
	if acc.Balance != 0 {
		t.Fatalf("balance after refund=%d", acc.Balance)
	}
	got, _ := svc.Get(f.a, v.ID)
	if got.VoucherStatus != VoucherRefunded {
		t.Fatalf("status after refund=%q", got.VoucherStatus)
	}
	if _, err := svc.RefundVoucher(f.a, v.ID, VoucherRefundInput{}); err == nil {
		t.Error("已退款的券不能重复退款")
	}
	if _, err := svc.SetVoucherStatus(f.a, v.ID, VoucherUsed); err == nil {
		t.Error("已退款的券不能直接改状态")
	}

	// 普通编辑不应把已退款改掉
	if _, err := svc.Update(f.a, v.ID, TransactionInput{
		Type: "expense", Amount: 5000, AccountID: f.aAcc, CategoryID: &f.aCat,
		VoucherStatus: VoucherPending, HappenedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	if got, _ := svc.Get(f.a, v.ID); got.VoucherStatus != VoucherRefunded {
		t.Fatalf("edit changed refunded to %q", got.VoucherStatus)
	}

	// 删除退款流水：券恢复待核销，余额回滚
	if err := svc.Delete(f.a, r.ID); err != nil {
		t.Fatal(err)
	}
	if got, _ := svc.Get(f.a, v.ID); got.VoucherStatus != VoucherPending {
		t.Fatalf("status after deleting refund=%q", got.VoucherStatus)
	}
	f.db.First(&acc, f.aAcc)
	if acc.Balance != -5000 {
		t.Fatalf("balance after deleting refund=%d", acc.Balance)
	}

	// 再退一次复用已有「退款」分类
	if _, err := svc.RefundVoucher(f.a, v.ID, VoucherRefundInput{Amount: 3000}); err != nil {
		t.Fatal(err)
	}
	if n := f.count(t, &model.Category{}, "user_id = ? AND name = ?", f.a, refundCategoryName); n != 1 {
		t.Fatalf("refund categories=%d", n)
	}
}

func TestRefundCategoryFollowsUserLocale(t *testing.T) {
	f := newTenantFixture(t)
	svc := NewTransactionService(f.db)
	if err := f.db.Model(&model.User{}).Where("id = ?", f.a).Update("locale", "en").Error; err != nil {
		t.Fatal(err)
	}
	newVoucher := func() uint64 {
		v, err := svc.Create(f.a, TransactionInput{
			Type: "expense", Amount: 2000, AccountID: f.aAcc, CategoryID: &f.aCat, Remark: "Hotpot",
			VoucherStatus: VoucherPending, HappenedAt: time.Now(),
		})
		if err != nil {
			t.Fatal(err)
		}
		return v.ID
	}

	r, err := svc.RefundVoucher(f.a, newVoucher(), VoucherRefundInput{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Category == nil || r.Category.Name != "Refund" {
		t.Fatalf("refund category=%+v", r.Category)
	}
	if r.Remark != "Voucher refund: Hotpot" {
		t.Fatalf("refund remark=%q", r.Remark)
	}
	var c model.Category
	f.db.First(&c, *r.CategoryID)
	if c.PresetKey != refundCategoryName || c.NameZh != "退款" || c.NameEn != "Refund" || c.NameJa != "返金" || c.NameKo != "환불" {
		t.Fatalf("preset fields=%+v", c)
	}

	// 改名后仍按 PresetKey 复用，不再新建
	f.db.Model(&c).Update("name", "Refunds")
	if _, err := svc.RefundVoucher(f.a, newVoucher(), VoucherRefundInput{}); err != nil {
		t.Fatal(err)
	}
	if n := f.count(t, &model.Category{}, "user_id = ? AND kind = ?", f.a, "income"); n != 1 {
		t.Fatalf("income categories=%d", n)
	}
}

func TestRefundCategoryReusesLegacyName(t *testing.T) {
	f := newTenantFixture(t)
	svc := NewTransactionService(f.db)
	legacy := model.Category{UserID: f.a, Name: refundCategoryName, Kind: "income", Icon: "mdi-cash-refund"}
	if err := f.db.Create(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	f.db.Model(&model.User{}).Where("id = ?", f.a).Update("locale", "ja")
	v, err := svc.Create(f.a, TransactionInput{
		Type: "expense", Amount: 1000, AccountID: f.aAcc, CategoryID: &f.aCat,
		VoucherStatus: VoucherPending, HappenedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	r, err := svc.RefundVoucher(f.a, v.ID, VoucherRefundInput{})
	if err != nil {
		t.Fatal(err)
	}
	if r.CategoryID == nil || *r.CategoryID != legacy.ID {
		t.Fatalf("expected legacy category %d, got %+v", legacy.ID, r.CategoryID)
	}
}

func TestNormalizeVoucher(t *testing.T) {
	if s, e := normalizeVoucher("income", VoucherPending, "2026-01-01"); s != "" || e != "" {
		t.Error("收入不能标记为券")
	}
	if s, _ := normalizeVoucher("expense", VoucherRefunded, ""); s != "" {
		t.Error("普通保存不能写入已退款")
	}
	if _, e := normalizeVoucher("expense", VoucherPending, "bad"); e != "" {
		t.Error("非法过期日应清空")
	}
}
