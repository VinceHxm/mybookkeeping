package service

import (
	"testing"
	"time"

	"mybookkeeping/internal/model"
)

func ledgerTx(id uint64, typ, dir string, amount int64, at time.Time, periods int) AccountLedgerItem {
	return AccountLedgerItem{
		Transaction: model.Transaction{ID: id, Type: typ, Amount: amount, HappenedAt: at, InstallmentPeriods: periods},
		Direction:   dir,
	}
}

func TestApplyCreditStatusFIFO(t *testing.T) {
	loc := time.Local
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, loc) // 账单日 15 → 已出账至 9/15
	acc := &model.Account{BillingDay: 15}
	items := []AccountLedgerItem{
		ledgerTx(5, "expense", "out", 3000, time.Date(2026, 9, 18, 10, 0, 0, 0, loc), 1),  // 未出账
		ledgerTx(4, "transfer", "in", 10000, time.Date(2026, 9, 16, 10, 0, 0, 0, loc), 1), // 还款 100
		ledgerTx(3, "expense", "out", 8000, time.Date(2026, 9, 1, 10, 0, 0, 0, loc), 1),   // 9/15 出账
		ledgerTx(2, "expense", "out", 5000, time.Date(2026, 8, 10, 10, 0, 0, 0, loc), 1),  // 8/15 出账
		ledgerTx(1, "expense", "out", 1200, time.Date(2026, 9, 2, 10, 0, 0, 0, loc), 3),   // 分期：9/15 出 400，后两期未出账
	}
	var sum AccountLedgerSummary
	applyCreditStatus(items, acc, 0, now, &sum)

	// 还款 10000 依次抵扣：8/15 的 5000 → 9/15 的 8000(抵 4600) 与 400(按时间先 8000)
	if items[3].CreditStatus != CreditStatusPaid || items[3].PaidAmount != 5000 {
		t.Fatalf("aug: %+v", items[3])
	}
	if items[2].CreditStatus != CreditStatusBilledUnpaid || items[2].PaidAmount != 5000 || items[2].BilledUnpaidAmount != 3000 {
		t.Fatalf("sep1: status=%s paid=%d unpaid=%d", items[2].CreditStatus, items[2].PaidAmount, items[2].BilledUnpaidAmount)
	}
	if items[4].BilledPeriods != 1 || items[4].BilledUnpaidAmount != 400 || items[4].UnbilledAmount != 800 {
		t.Fatalf("installment: %+v", items[4])
	}
	if items[0].CreditStatus != CreditStatusUnbilled || items[0].UnbilledAmount != 3000 {
		t.Fatalf("unbilled: %+v", items[0])
	}
	if items[1].CreditStatus != "" {
		t.Fatalf("repayment should have no status")
	}
	if sum.Paid != 10000 || sum.BilledUnpaid != 3400 || sum.Unbilled != 3800 {
		t.Fatalf("sum: %+v", sum)
	}
}

func TestApplyCreditStatusOpeningDebtFirst(t *testing.T) {
	loc := time.Local
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, loc)
	acc := &model.Account{BillingDay: 15}
	items := []AccountLedgerItem{
		ledgerTx(2, "transfer", "in", 6000, time.Date(2026, 9, 16, 10, 0, 0, 0, loc), 1),
		ledgerTx(1, "expense", "out", 5000, time.Date(2026, 9, 1, 10, 0, 0, 0, loc), 1),
	}
	var sum AccountLedgerSummary
	applyCreditStatus(items, acc, -4000, now, &sum)
	if sum.OpeningDebt != 4000 || sum.OpeningDebtUnpaid != 0 {
		t.Fatalf("opening: %+v", sum)
	}
	if items[1].PaidAmount != 2000 || items[1].BilledUnpaidAmount != 3000 {
		t.Fatalf("expense: %+v", items[1])
	}
}
