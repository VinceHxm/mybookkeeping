package service

import (
	"testing"
	"time"
)

func TestLastClosedStatementAndCycle(t *testing.T) {
	loc := time.Local
	// 账单日 15：9/18 → 出账 9/15，周期 8/16～9/15
	now := time.Date(2026, 9, 18, 12, 0, 0, 0, loc)
	stmt := lastClosedStatementDate(15, now)
	if got := fmtDate(stmt); got != "2026-09-15" {
		t.Fatalf("stmt=%s", got)
	}
	start, end := cycleRangeForStatement(stmt, 15)
	if fmtDate(start) != "2026-08-16" || fmtDate(end) != "2026-09-15" {
		t.Fatalf("cycle %s ~ %s", fmtDate(start), fmtDate(end))
	}

	// 9/10 → 出账仍为 8/15
	now2 := time.Date(2026, 9, 10, 12, 0, 0, 0, loc)
	stmt2 := lastClosedStatementDate(15, now2)
	if fmtDate(stmt2) != "2026-08-15" {
		t.Fatalf("stmt2=%s", fmtDate(stmt2))
	}
}

func TestPaymentDueForStatement(t *testing.T) {
	loc := time.Local
	stmt := time.Date(2026, 9, 15, 0, 0, 0, 0, loc)
	// 账单 15、还款 25 → 同月 9/25
	if got := fmtDate(paymentDueForStatement(stmt, 15, 25)); got != "2026-09-25" {
		t.Fatalf("same month due=%s", got)
	}
	// 账单 15、还款 5 → 下月 10/5
	if got := fmtDate(paymentDueForStatement(stmt, 15, 5)); got != "2026-10-05" {
		t.Fatalf("next month due=%s", got)
	}
}

func TestRelevantPaymentDueNear(t *testing.T) {
	loc := time.Local
	// 9/24：账单日 15、还款日 25 → 关注 9/25，days=1
	today := time.Date(2026, 9, 24, 12, 0, 0, 0, loc)
	due := relevantPaymentDueDate(15, 25, time.Date(2026, 9, 24, 0, 0, 0, 0, loc))
	if fmtDate(due) != "2026-09-25" {
		t.Fatalf("due=%s", fmtDate(due))
	}
	days := int(due.Sub(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)).Hours() / 24)
	if days != 1 {
		t.Fatalf("days=%d", days)
	}
}

func TestResolveStatementDue(t *testing.T) {
	if got := resolveStatementDue(0, 196302); got != 196302 {
		t.Fatalf("hand billed should win when tx=0, got %d", got)
	}
	if got := resolveStatementDue(200000, 196302); got != 200000 {
		t.Fatalf("higher tx due should win, got %d", got)
	}
	if got := resolveStatementDue(100, 100); got != 100 {
		t.Fatalf("equal ok, got %d", got)
	}
}

func TestFirstDueAndInstallmentSplit(t *testing.T) {
	loc := time.Local
	buy := time.Date(2026, 9, 20, 10, 0, 0, 0, loc)
	first := firstDueStatementForPurchase(buy, 15)
	if fmtDate(first) != "2026-10-15" {
		t.Fatalf("first=%s", fmtDate(first))
	}
	shares := splitAmountEvenly(10001, 3)
	if shares[0] != 3335 || shares[1] != 3333 || shares[2] != 3333 {
		t.Fatalf("shares=%v", shares)
	}
	buy2 := time.Date(2026, 8, 10, 10, 0, 0, 0, loc)
	if fmtDate(firstDueStatementForPurchase(buy2, 15)) != "2026-08-15" {
		t.Fatal("expected Aug statement")
	}
}

