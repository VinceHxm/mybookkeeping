package service

import (
	"reflect"
	"testing"
	"time"

	"mybookkeeping/internal/model"
)

func TestCreditStatementsBatchMatchesSingle(t *testing.T) {
	f := newTenantFixture(t)
	svc := NewTransactionService(f.db)
	accSvc := NewAccountService(f.db)
	at := time.Date(2026, 10, 8, 12, 0, 0, 0, time.Local)

	mkCard := func(name string, billingDay int) uint64 {
		a := &model.Account{UserID: f.a, Name: name, Type: "credit", BillingDay: billingDay, PaymentDueDay: 9, CreditLimit: 1000000}
		if err := f.db.Create(a).Error; err != nil {
			t.Fatal(err)
		}
		return a.ID
	}
	c1, c2 := mkCard("招行", 5), mkCard("花呗", 20)
	noDay := &model.Account{UserID: f.a, Name: "未设账单日", Type: "credit"}
	f.db.Create(noDay)
	bCard := &model.Account{UserID: f.b, Name: "B 的卡", Type: "credit", BillingDay: 5}
	f.db.Create(bCard)

	add := func(in TransactionInput) {
		t.Helper()
		if _, err := svc.Create(f.a, in); err != nil {
			t.Fatal(err)
		}
	}
	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 10, 0, 0, 0, time.Local) }
	for _, c := range []uint64{c1, c2} {
		to := c
		add(TransactionInput{Type: "expense", Amount: 30000, AccountID: c, CategoryID: &f.aCat, HappenedAt: day(9, 3)})
		add(TransactionInput{Type: "expense", Amount: 60000, AccountID: c, CategoryID: &f.aCat, InstallmentPeriods: 3, HappenedAt: day(8, 1)})
		add(TransactionInput{Type: "expense", Amount: 12000, AccountID: c, CategoryID: &f.aCat, HappenedAt: day(10, 6)})
		add(TransactionInput{Type: "transfer", Amount: 20000, AccountID: f.aAcc, ToAccountID: &to, HappenedAt: day(10, 7)})
		add(TransactionInput{Type: "transfer", Amount: 5000, AccountID: f.aAcc, ToAccountID: &to, HappenedAt: day(8, 10)})
		add(TransactionInput{Type: "income", Amount: 3000, AccountID: c, CategoryID: &f.aCat, HappenedAt: day(10, 6)})
	}

	batch, err := accSvc.CreditStatements(f.a, at)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch) != 2 {
		t.Fatalf("应只返回本人已设账单日的 2 张卡，得到 %d", len(batch))
	}
	for _, got := range batch {
		want, err := accSvc.CreditStatement(f.a, got.AccountID, at)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, *want) {
			t.Errorf("卡 %d 批量与单卡结果不一致:\nbatch=%+v\nsingle=%+v", got.AccountID, got, *want)
		}
		if got.RepaidAmount == 0 || got.InstallmentDue == 0 {
			t.Errorf("卡 %d 未取到流水: %+v", got.AccountID, got)
		}
	}
}

func TestCreditStatementInstallmentItems(t *testing.T) {
	loc := time.Local
	at := time.Date(2026, 10, 8, 16, 0, 0, 0, loc)
	acc := &model.Account{
		ID: 6, Name: "招行", Type: "credit", BillingDay: 18, Balance: -140789, CreditLimit: 1000000,
		InstallmentPlans: []model.CreditInstallmentPlan{
			{ID: 3, Name: "旧分期", PrincipalFen: 30000, Periods: 3, InterestPerPeriodFen: 500, FirstDueOn: "2026-09-18"},
			{ID: 4, Name: "已还完", PrincipalFen: 10000, Periods: 1, FirstDueOn: "2026-08-18"},
		},
	}
	txs := &creditTxs{expenses: []model.Transaction{
		{ID: 49, Type: "expense", Amount: 129149, InstallmentPeriods: 12, InterestFen: 235, Remark: "手机分期", HappenedAt: time.Date(2026, 10, 8, 15, 55, 0, 0, loc)},
		{ID: 50, Type: "expense", Amount: 3000, InstallmentPeriods: 1, HappenedAt: time.Date(2026, 10, 1, 10, 0, 0, 0, loc)},
	}}

	s := computeCreditStatement(acc, txs, at)
	if len(s.Installments) != 2 {
		t.Fatalf("应列出 2 笔未出完的分期，得到 %+v", s.Installments)
	}
	tx := s.Installments[0]
	if tx.Source != "transaction" || tx.SourceID != 49 || tx.Name != "手机分期" ||
		tx.BilledPeriods != 0 || tx.RemainingFen != 129149 || tx.NextDueOn != "2026-10-18" ||
		tx.NextShareFen != 10767 || tx.LastDueOn != "2027-09-18" || tx.InterestFen != 235 {
		t.Errorf("流水分期: %+v", tx)
	}
	plan := s.Installments[1]
	if plan.Source != "plan" || plan.BilledPeriods != 1 || plan.RemainingFen != 20000 ||
		plan.NextDueOn != "2026-10-18" || plan.LastDueOn != "2026-11-18" || plan.InterestPerPeriodFen != 500 {
		t.Errorf("存量分期: %+v", plan)
	}
	if s.FutureInstallment != 129149+20000 {
		t.Errorf("futureInstallment=%d", s.FutureInstallment)
	}

	txs.expenses[0].InstallmentDeferCycles = 1
	s = computeCreditStatement(acc, txs, at)
	tx = s.Installments[len(s.Installments)-1]
	if tx.SourceID != 49 || tx.NextDueOn != "2026-11-18" || tx.LastDueOn != "2027-10-18" || tx.RemainingFen != 129149 {
		t.Errorf("推迟一期后首期应在 11/18: %+v", tx)
	}
	if s.NextPeriodInstallment != 10000 {
		t.Errorf("推迟后下期(10/18)只应含存量分期一期，得到 %d", s.NextPeriodInstallment)
	}
}

func TestNormalizeDeferCycles(t *testing.T) {
	cases := []struct{ periods, n, want int }{{1, 1, 0}, {12, -1, 0}, {12, 1, 1}, {12, 9, maxDeferCycles}}
	for _, c := range cases {
		if got := normalizeDeferCycles(c.periods, c.n); got != c.want {
			t.Errorf("normalizeDeferCycles(%d,%d)=%d want %d", c.periods, c.n, got, c.want)
		}
	}
}
