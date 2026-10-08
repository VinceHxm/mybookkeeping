package service

import (
	"testing"
	"time"

	"mybookkeeping/internal/model"
)

func TestInvestRevalueFlow(t *testing.T) {
	f := newTenantFixture(t)
	svc := NewTransactionService(f.db)
	accSvc := NewAccountService(f.db)

	fund := &model.Account{UserID: f.a, Name: "支付宝基金", Type: AccountTypeInvest, Balance: 100000}
	if err := f.db.Create(fund).Error; err != nil {
		t.Fatal(err)
	}
	balanceOf := func() int64 {
		var a model.Account
		f.db.First(&a, fund.ID)
		return a.Balance
	}

	if _, err := svc.Revalue(f.a, f.aAcc, RevalueInput{MarketValue: 1}); err == nil {
		t.Error("非理财账户不能校准市值")
	}
	if _, err := svc.Create(f.a, TransactionInput{Type: TxTypeRevalue, Amount: 100, AccountID: fund.ID, HappenedAt: time.Now()}); err == nil {
		t.Error("普通记账接口不能直接创建市值校准")
	}

	// 买入 500 元：现金 → 基金
	toFund := fund.ID
	if _, err := svc.Create(f.a, TransactionInput{Type: "transfer", Amount: 50000, AccountID: f.aAcc, ToAccountID: &toFund, HappenedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	// 市值涨到 1580 元：浮盈 80 元
	up, err := svc.Revalue(f.a, fund.ID, RevalueInput{MarketValue: 158000})
	if err != nil {
		t.Fatal(err)
	}
	if up == nil || up.Type != TxTypeRevalue || up.Amount != 8000 || up.CategoryID != nil {
		t.Fatalf("revalue up=%+v", up)
	}
	if b := balanceOf(); b != 158000 {
		t.Fatalf("balance after revalue=%d", b)
	}
	// 跌到 1500 元：浮亏 80 元
	down, err := svc.Revalue(f.a, fund.ID, RevalueInput{MarketValue: 150000})
	if err != nil {
		t.Fatal(err)
	}
	if down.Amount != -8000 {
		t.Fatalf("revalue down amount=%d", down.Amount)
	}
	// 与余额一致：不生成流水，只刷新对账时间
	same, err := svc.Revalue(f.a, fund.ID, RevalueInput{MarketValue: 150000})
	if err != nil || same != nil {
		t.Fatalf("same value: tx=%+v err=%v", same, err)
	}
	var acc model.Account
	f.db.First(&acc, fund.ID)
	if acc.LastReconciledAt == nil {
		t.Error("校准后应记录对账时间")
	}

	// 改差额：余额按新差额重算
	if _, err := svc.Update(f.a, down.ID, TransactionInput{Type: TxTypeRevalue, Amount: -3000, AccountID: fund.ID, HappenedAt: down.HappenedAt}); err != nil {
		t.Fatal(err)
	}
	if b := balanceOf(); b != 155000 {
		t.Fatalf("balance after editing revalue=%d", b)
	}
	if _, err := svc.Update(f.a, down.ID, TransactionInput{Type: "expense", Amount: 3000, AccountID: fund.ID, CategoryID: &f.aCat, HappenedAt: time.Now()}); err == nil {
		t.Error("市值校准不能改成支出")
	}
	if _, err := svc.Update(f.a, down.ID, TransactionInput{Type: TxTypeRevalue, Amount: -3000, AccountID: f.aAcc, HappenedAt: time.Now()}); err == nil {
		t.Error("市值校准不能挪到非理财账户")
	}

	// 账户明细：净投入 = 期初 1000 + 转入 500；收益 = 80 − 30
	ledger, err := accSvc.Ledger(f.a, fund.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if ledger.Invest == nil || ledger.Invest.NetInvested != 150000 || ledger.Invest.Gain != 5000 || ledger.Invest.Revalued != 5000 {
		t.Fatalf("invest summary=%+v", ledger.Invest)
	}
	if m := ledger.Months[0]; m.InAmount != 50000 || m.OutAmount != 0 || m.Revalued != 5000 {
		t.Fatalf("month=%+v", m)
	}

	// 删除校准流水：余额回滚
	if err := svc.Delete(f.a, up.ID); err != nil {
		t.Fatal(err)
	}
	if b := balanceOf(); b != 147000 {
		t.Fatalf("balance after deleting revalue=%d", b)
	}
}
