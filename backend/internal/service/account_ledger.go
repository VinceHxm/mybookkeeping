package service

import (
	"sort"
	"time"

	"mybookkeeping/internal/model"
)

const (
	CreditStatusPaid         = "paid"          // 已还
	CreditStatusBilledUnpaid = "billed_unpaid" // 已出账未还
	CreditStatusUnbilled     = "unbilled"      // 未出账
)

// AccountLedgerItem 账户明细中的一笔流水（信用卡附带出账/还款状态）
type AccountLedgerItem struct {
	model.Transaction
	Direction string `json:"direction"` // in=流入本账户 / out=流出本账户

	// 以下仅信用卡「流出」流水有值
	CreditStatus       string `json:"creditStatus,omitempty"`
	PaidAmount         int64  `json:"paidAmount"`
	BilledUnpaidAmount int64  `json:"billedUnpaidAmount"`
	UnbilledAmount     int64  `json:"unbilledAmount"`
	StatementDate      string `json:"statementDate,omitempty"` // 首期出账日
	BilledPeriods      int    `json:"billedPeriods,omitempty"` // 已出账期数（分期时有意义）
}

type AccountLedgerMonth struct {
	Month        string              `json:"month"` // YYYY-MM
	InAmount     int64               `json:"inAmount"`
	OutAmount    int64               `json:"outAmount"`
	Paid         int64               `json:"paid"`
	BilledUnpaid int64               `json:"billedUnpaid"`
	Unbilled     int64               `json:"unbilled"`
	Revalued     int64               `json:"revalued"` // 市值校准净额（理财账户），不含在流入/流出里
	Items        []AccountLedgerItem `json:"items"`
}

// InvestSummary 理财账户收益概览；收益率为简单收益率（累计收益 ÷ 净投入），不考虑资金时间
type InvestSummary struct {
	NetInvested int64 `json:"netInvested"` // 净投入 = 期初余额 + 转入 − 转出
	Gain        int64 `json:"gain"`        // 累计收益 = 当前市值 − 净投入（含市值校准、分红收入、手续费支出）
	Revalued    int64 `json:"revalued"`    // 其中市值校准合计
}

type AccountLedgerSummary struct {
	Paid               int64 `json:"paid"`
	BilledUnpaid       int64 `json:"billedUnpaid"`
	Unbilled           int64 `json:"unbilled"`
	OpeningDebt        int64 `json:"openingDebt"`        // 建账/校准带入、无对应流水的欠款
	OpeningDebtUnpaid  int64 `json:"openingDebtUnpaid"`  // 其中尚未被还款覆盖的部分
	LastStatementDate  string `json:"lastStatementDate,omitempty"`
	NextStatementDate  string `json:"nextStatementDate,omitempty"`
}

type AccountLedger struct {
	AccountID      uint64               `json:"accountId"`
	AccountName    string               `json:"accountName"`
	AccountType    string               `json:"accountType"`
	Balance        int64                `json:"balance"`
	CreditTracking bool                 `json:"creditTracking"` // 信用卡且已设账单日
	Summary        AccountLedgerSummary `json:"summary"`
	Invest         *InvestSummary       `json:"invest,omitempty"`
	Months         []AccountLedgerMonth `json:"months"`
	Note           string               `json:"note,omitempty"`
}

type creditShare struct {
	item   int // -1 = 期初欠款
	amount int64
	dueOn  time.Time
	billed bool
	at     time.Time
	id     uint64
}

// Ledger 账户流水明细，按自然月分组；信用卡按「先出账先还」分摊还款得到每笔状态。
func (s *AccountService) Ledger(userID, accountID uint64, at time.Time) (*AccountLedger, error) {
	acc, err := s.Get(userID, accountID)
	if err != nil {
		return nil, err
	}
	if at.IsZero() {
		at = time.Now()
	}
	loc := at.Location()

	var txs []model.Transaction
	if err := s.db.Where("user_id = ? AND (account_id = ? OR to_account_id = ?)", userID, accountID, accountID).
		Preload("Account", briefAccount).Preload("ToAccount", briefAccount).Preload("Category").Preload("Tags").
		Order("happened_at desc, id desc").
		Find(&txs).Error; err != nil {
		return nil, err
	}

	items := make([]AccountLedgerItem, len(txs))
	var effect, transferIn, transferOut, revalued int64
	for i, t := range txs {
		t.HappenedAt = t.HappenedAt.In(loc)
		dir := "out"
		switch t.Type {
		case "income":
			dir = "in"
		case "transfer":
			if t.ToAccountID != nil && *t.ToAccountID == accountID {
				dir = "in"
			}
		case TxTypeRevalue:
			if t.Amount > 0 {
				dir = "in"
			}
		}
		switch {
		case t.Type == TxTypeRevalue:
			effect += t.Amount
			revalued += t.Amount
		case dir == "in":
			effect += t.Amount
			if t.Type == "transfer" {
				transferIn += t.Amount
			}
		default:
			effect -= t.Amount
			if t.Type == "transfer" {
				transferOut += t.Amount
			}
		}
		items[i] = AccountLedgerItem{Transaction: t, Direction: dir}
	}

	out := &AccountLedger{
		AccountID:   acc.ID,
		AccountName: acc.Name,
		AccountType: acc.Type,
		Balance:     acc.Balance,
	}

	if acc.Type == "credit" {
		if acc.BillingDay >= 1 && acc.BillingDay <= 28 {
			out.CreditTracking = true
			applyCreditStatus(items, acc, acc.Balance-effect, at, &out.Summary)
			out.Note = "状态按「先出账先还」分摊还款与退款估算：还款优先抵扣最早出账的消费。"
		} else {
			out.Note = "未设置账单日，无法区分已出账 / 未出账。"
		}
	}
	if acc.Type == AccountTypeInvest {
		netInvested := acc.Balance - effect + transferIn - transferOut
		out.Invest = &InvestSummary{
			NetInvested: netInvested,
			Gain:        acc.Balance - netInvested,
			Revalued:    revalued,
		}
	}

	out.Months = groupLedgerByMonth(items, out.CreditTracking)
	return out, nil
}

func applyCreditStatus(items []AccountLedgerItem, acc *model.Account, openingBalance int64, at time.Time, sum *AccountLedgerSummary) {
	billingDay := acc.BillingDay
	lastStmt := lastClosedStatementDate(billingDay, at)
	sum.LastStatementDate = fmtDate(lastStmt)
	sum.NextStatementDate = fmtDate(nextStatementDate(lastStmt, billingDay))

	var credits int64
	var shares []creditShare
	if openingBalance < 0 {
		sum.OpeningDebt = -openingBalance
		shares = append(shares, creditShare{item: -1, amount: -openingBalance, billed: true})
	} else {
		credits += openingBalance
	}

	for i := range items {
		it := &items[i]
		if it.Direction == "in" {
			credits += it.Amount
			continue
		}
		if it.Amount <= 0 {
			continue
		}
		periods := 1
		if it.Type == "expense" && it.InstallmentPeriods > 1 {
			periods = it.InstallmentPeriods
		}
		firstDue := firstDueStatementForPurchase(it.HappenedAt, billingDay)
		if periods > 1 {
			firstDue = firstDueForExpense(it.Transaction, billingDay)
		}
		it.StatementDate = fmtDate(firstDue)
		for p, amt := range splitAmountEvenly(it.Amount, periods) {
			dueOn := addBillingMonths(firstDue, p, billingDay)
			billed := !dueOn.After(lastStmt)
			if billed {
				it.BilledPeriods++
			}
			shares = append(shares, creditShare{item: i, amount: amt, dueOn: dueOn, billed: billed, at: it.HappenedAt, id: it.ID})
		}
	}

	sort.SliceStable(shares, func(a, b int) bool {
		x, y := shares[a], shares[b]
		if (x.item == -1) != (y.item == -1) {
			return x.item == -1
		}
		if !x.dueOn.Equal(y.dueOn) {
			return x.dueOn.Before(y.dueOn)
		}
		if !x.at.Equal(y.at) {
			return x.at.Before(y.at)
		}
		return x.id < y.id
	})

	for _, sh := range shares {
		paid := sh.amount
		if credits < paid {
			paid = credits
		}
		credits -= paid
		rest := sh.amount - paid
		if sh.item == -1 {
			sum.OpeningDebtUnpaid = rest
			continue
		}
		it := &items[sh.item]
		it.PaidAmount += paid
		if sh.billed {
			it.BilledUnpaidAmount += rest
		} else {
			it.UnbilledAmount += rest
		}
	}

	for i := range items {
		it := &items[i]
		if it.Direction != "out" || it.Amount <= 0 {
			continue
		}
		switch {
		case it.BilledUnpaidAmount > 0:
			it.CreditStatus = CreditStatusBilledUnpaid
		case it.UnbilledAmount > 0:
			it.CreditStatus = CreditStatusUnbilled
		default:
			it.CreditStatus = CreditStatusPaid
		}
		sum.Paid += it.PaidAmount
		sum.BilledUnpaid += it.BilledUnpaidAmount
		sum.Unbilled += it.UnbilledAmount
	}
}

func groupLedgerByMonth(items []AccountLedgerItem, credit bool) []AccountLedgerMonth {
	months := make([]AccountLedgerMonth, 0)
	idx := map[string]int{}
	for _, it := range items {
		key := it.HappenedAt.Format("2006-01")
		i, ok := idx[key]
		if !ok {
			i = len(months)
			idx[key] = i
			months = append(months, AccountLedgerMonth{Month: key, Items: []AccountLedgerItem{}})
		}
		m := &months[i]
		if it.Type == TxTypeRevalue {
			m.Revalued += it.Amount
		} else if it.Direction == "in" {
			m.InAmount += it.Amount
		} else {
			m.OutAmount += it.Amount
		}
		if credit {
			m.Paid += it.PaidAmount
			m.BilledUnpaid += it.BilledUnpaidAmount
			m.Unbilled += it.UnbilledAmount
		}
		m.Items = append(m.Items, it)
	}
	return months
}
