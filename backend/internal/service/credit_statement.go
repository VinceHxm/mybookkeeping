package service

import (
	"errors"
	"time"

	"mybookkeeping/internal/model"
)

// CreditStatement 信用卡账单展示（只读，不落库、不改余额）
type CreditStatement struct {
	AccountID          uint64    `json:"accountId"`
	AccountName        string    `json:"accountName"`
	BillingDay         int       `json:"billingDay"`
	StatementDate      string    `json:"statementDate"`     // 本期出账日 YYYY-MM-DD
	CycleStart         string    `json:"cycleStart"`        // 账单周期起
	CycleEnd           string    `json:"cycleEnd"`          // 账单周期止（出账日）
	NextStatementDate  string    `json:"nextStatementDate"` // 下期出账日
	SpentInCycle       int64     `json:"spentInCycle"`      // 周期内刷卡总额（流水本金）
	InstallmentDue     int64     `json:"installmentDue"`    // 本期入账的分期本金份额
	NonInstallmentDue  int64     `json:"nonInstallmentDue"` // 本期一次性入账本金
	InterestDue        int64     `json:"interestDue"`       // 本期计入应还的利息/手续费
	DueAmount          int64     `json:"dueAmount"`         // 本期应还 = 本金份额 + 利息（含手填已出账兜底）
	RepaidAmount       int64     `json:"repaidAmount"`      // 出账后至下期出账前的还款转账合计
	RepaidInterest     int64     `json:"repaidInterest"`    // 还款流水上标注的利息（展示拆分）
	PeriodRemaining    int64     `json:"periodRemaining"`   // 本期待还 = max(0, 本期应还−本期已还)
	RemainingAfterPay  int64     `json:"remainingAfterPay"` // 还后待还 = 整体待还 − 已出账（即未出账）
	TotalOutstanding   int64     `json:"totalOutstanding"`  // 整体待还 = 已用额度（方便提前还款）
	FutureInstallment  int64     `json:"futureInstallment"` // 后续分期待还本金
	OutstandingBalance int64     `json:"outstandingBalance"` // 同 totalOutstanding（兼容旧字段）
	BilledOutstanding  int64     `json:"billedOutstanding"`  // 账户手填的已出账
	UnbilledOutstanding int64    `json:"unbilledOutstanding"` // 已用 − 已出账（同 remainingAfterPay）
	AvailableCredit    *int64    `json:"availableCredit"`
	DisplayOnly        bool      `json:"displayOnly"`
	Note               string    `json:"note"`
	ComputedAt         time.Time `json:"computedAt"`
}

func (s *AccountService) CreditStatement(userID, accountID uint64, at time.Time) (*CreditStatement, error) {
	acc, err := s.Get(userID, accountID)
	if err != nil {
		return nil, err
	}
	if acc.Type != "credit" {
		return nil, errors.New("仅信用卡账户支持账单展示")
	}
	if acc.BillingDay < 1 || acc.BillingDay > 28 {
		return nil, errors.New("请先设置账单日（1-28）")
	}
	if at.IsZero() {
		at = time.Now()
	}
	loc := at.Location()

	stmt := lastClosedStatementDate(acc.BillingDay, at)
	cycleStart, cycleEnd := cycleRangeForStatement(stmt, acc.BillingDay)
	nextStmt := nextStatementDate(stmt, acc.BillingDay)
	repayFrom := time.Date(stmt.Year(), stmt.Month(), stmt.Day(), 0, 0, 0, 0, loc)
	repayTo := time.Date(nextStmt.Year(), nextStmt.Month(), nextStmt.Day(), 0, 0, 0, 0, loc)
	cycleTo := cycleEnd.Add(24 * time.Hour)

	var expenses []model.Transaction
	if err := s.db.Where("user_id = ? AND type = ? AND account_id = ?", userID, "expense", accountID).
		Find(&expenses).Error; err != nil {
		return nil, err
	}

	var spentInCycle, installmentDue, nonInstallmentDue, interestDue, futureInstallment int64
	for _, e := range expenses {
		if !e.HappenedAt.Before(cycleStart) && e.HappenedAt.Before(cycleTo) {
			spentInCycle += e.Amount
		}

		periods := e.InstallmentPeriods
		if periods < 1 {
			periods = 1
		}
		firstDue := firstDueStatementForPurchase(e.HappenedAt, acc.BillingDay)
		shares := splitAmountEvenly(e.Amount, periods)
		for i := 0; i < periods; i++ {
			dueOn := addBillingMonths(firstDue, i, acc.BillingDay)
			if sameYMD(dueOn, stmt) {
				if periods == 1 {
					nonInstallmentDue += shares[i]
				} else {
					installmentDue += shares[i]
				}
				if i == 0 && e.InterestFen > 0 {
					interestDue += e.InterestFen
				}
			} else if dueOn.After(stmt) {
				futureInstallment += shares[i]
			}
		}
	}

	var repayments []model.Transaction
	if err := s.db.Where(
		"user_id = ? AND type = ? AND to_account_id = ? AND happened_at >= ? AND happened_at < ?",
		userID, "transfer", accountID, repayFrom, repayTo,
	).Find(&repayments).Error; err != nil {
		return nil, err
	}
	var repaid, repaidInterest int64
	for _, r := range repayments {
		repaid += r.Amount
		repaidInterest += r.InterestFen
	}

	dueFromTx := installmentDue + nonInstallmentDue + interestDue

	outstanding := CreditUsedFen(acc)
	billed := acc.CreditBilledFen
	if billed < 0 {
		billed = 0
	}
	if billed > outstanding {
		billed = outstanding
	}
	unbilled := outstanding - billed

	// 本期应还：以流水拆分为准；手填「已出账」仅在尚未被本期还款覆盖时兜底抬高。
	// 避免：账单已还清后仍把「剩余已用/下期未出账」算进本期应还。
	due := resolveStatementDue(dueFromTx, billed, repaid)
	periodRemaining := due - repaid
	if periodRemaining < 0 {
		periodRemaining = 0
	}
	// 本期已结清后：若应还额仍等于当前已用、且还款大于该应还，
	// 多半是还清了更大的上期账单，剩余已用实为下期未出账，不应再展示为本期应还。
	if periodRemaining == 0 && due > 0 && due == outstanding && repaid > due {
		due = 0
	}
	// 还后待还：本期结清后，剩余已用全部视为未出账/下期；未结清时 = 已用 − 已出账
	afterPay := unbilled
	displayUnbilled := unbilled
	if periodRemaining == 0 {
		afterPay = outstanding
		displayUnbilled = outstanding
	}

	var available *int64
	if acc.CreditLimit > 0 {
		v := acc.CreditLimit - outstanding
		available = &v
	}

	note := "账单出账仅供展示，不自动记账；还款以转账流水为准。"
	if billed > dueFromTx && repaid < billed {
		note += "本期应还已采用手填「已出账」。"
	}
	if periodRemaining == 0 && outstanding > 0 {
		note += "本期已还清，剩余已用计入还后待还（下期/未出账），不计入本期应还。"
	} else {
		note += "「还后待还」= 整体待还 − 已出账；「整体待还」= 已用额度，可用于提前还款。"
	}

	return &CreditStatement{
		AccountID:           acc.ID,
		AccountName:         acc.Name,
		BillingDay:          acc.BillingDay,
		StatementDate:       fmtDate(stmt),
		CycleStart:          fmtDate(cycleStart),
		CycleEnd:            fmtDate(cycleEnd),
		NextStatementDate:   fmtDate(nextStmt),
		SpentInCycle:        spentInCycle,
		InstallmentDue:      installmentDue,
		NonInstallmentDue:   nonInstallmentDue,
		InterestDue:         interestDue,
		DueAmount:           due,
		RepaidAmount:        repaid,
		RepaidInterest:      repaidInterest,
		PeriodRemaining:     periodRemaining,
		RemainingAfterPay:   afterPay,
		TotalOutstanding:    outstanding,
		FutureInstallment:   futureInstallment,
		OutstandingBalance:  outstanding,
		BilledOutstanding:   billed,
		UnbilledOutstanding: displayUnbilled,
		AvailableCredit:     available,
		DisplayOnly:         true,
		Note:                note,
		ComputedAt:          at,
	}, nil
}

// resolveStatementDue 手填已出账与流水应还取较大值；若本期还款已覆盖手填已出账，则不再用手填抬高（避免下期未出账混入本期应还）
func resolveStatementDue(txDue, billedFen, repaid int64) int64 {
	if billedFen > txDue && repaid < billedFen {
		return billedFen
	}
	return txDue
}

func fmtDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func sameYMD(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func lastClosedStatementDate(billingDay int, now time.Time) time.Time {
	loc := now.Location()
	y, m, d := now.Date()
	if d >= billingDay {
		return clampDay(y, m, billingDay, loc)
	}
	prev := time.Date(y, m, 1, 0, 0, 0, 0, loc).AddDate(0, -1, 0)
	return clampDay(prev.Year(), prev.Month(), billingDay, loc)
}

func nextStatementDate(from time.Time, billingDay int) time.Time {
	n := from.AddDate(0, 1, 0)
	return clampDay(n.Year(), n.Month(), billingDay, from.Location())
}

func cycleRangeForStatement(stmt time.Time, billingDay int) (start, end time.Time) {
	loc := stmt.Location()
	prev := stmt.AddDate(0, -1, 0)
	prevStmt := clampDay(prev.Year(), prev.Month(), billingDay, loc)
	start = time.Date(prevStmt.Year(), prevStmt.Month(), prevStmt.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
	end = time.Date(stmt.Year(), stmt.Month(), stmt.Day(), 0, 0, 0, 0, loc)
	return start, end
}

func firstDueStatementForPurchase(happenedAt time.Time, billingDay int) time.Time {
	loc := happenedAt.Location()
	y, m, d := happenedAt.Date()
	thisStmt := clampDay(y, m, billingDay, loc)
	if d <= billingDay {
		return thisStmt
	}
	return nextStatementDate(thisStmt, billingDay)
}

func addBillingMonths(stmt time.Time, n, billingDay int) time.Time {
	t := stmt.AddDate(0, n, 0)
	return clampDay(t.Year(), t.Month(), billingDay, stmt.Location())
}

func clampDay(year int, month time.Month, day int, loc *time.Location) time.Time {
	if day < 1 {
		day = 1
	}
	if day > 28 {
		day = 28
	}
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func splitAmountEvenly(total int64, periods int) []int64 {
	if periods < 1 {
		periods = 1
	}
	out := make([]int64, periods)
	base := total / int64(periods)
	rem := total % int64(periods)
	for i := 0; i < periods; i++ {
		out[i] = base
	}
	out[0] += rem
	return out
}

// CreditRepayReminder 临近还款日提醒（有欠款才返回）
type CreditRepayReminder struct {
	AccountID         uint64  `json:"accountId"`
	AccountName       string  `json:"accountName"`
	PaymentDueDay     int     `json:"paymentDueDay"`
	DueDate           string  `json:"dueDate"`
	DaysUntilDue      int     `json:"daysUntilDue"` // 0=今天到期，1=明天到期
	AmountFen         int64   `json:"amountFen"`    // 建议还款额（本期待还）
	FromAccountID     *uint64 `json:"fromAccountId"`
	FromAccountName   string  `json:"fromAccountName,omitempty"`
}

// CreditRepayReminders 返回「还款日前一天或当天」且仍有欠款的信用卡
func (s *AccountService) CreditRepayReminders(userID uint64, at time.Time) ([]CreditRepayReminder, error) {
	if at.IsZero() {
		at = time.Now()
	}
	list, err := s.List(userID, false)
	if err != nil {
		return nil, err
	}

	var defaultFrom *model.Account
	var user model.User
	if err := s.db.Select("id", "default_account_id").Where("id = ?", userID).First(&user).Error; err == nil && user.DefaultAccountID != nil {
		for i := range list {
			if list[i].ID == *user.DefaultAccountID && list[i].Type != "credit" && !list[i].Archived {
				defaultFrom = &list[i]
				break
			}
		}
	}
	if defaultFrom == nil {
		for i := range list {
			if list[i].Type != "credit" && !list[i].Archived {
				defaultFrom = &list[i]
				break
			}
		}
	}

	today := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location())
	out := make([]CreditRepayReminder, 0)
	for _, acc := range list {
		if acc.Type != "credit" || acc.Archived || acc.PaymentDueDay < 1 || acc.PaymentDueDay > 28 {
			continue
		}
		if acc.BillingDay < 1 || acc.BillingDay > 28 {
			continue
		}
		due := relevantPaymentDueDate(acc.BillingDay, acc.PaymentDueDay, today)
		days := int(due.Sub(today).Hours() / 24)
		// 仅还款日前一天或当天
		if days < 0 || days > 1 {
			continue
		}
		stmt, err := s.CreditStatement(userID, acc.ID, at)
		if err != nil {
			continue
		}
		amount := stmt.PeriodRemaining
		if amount <= 0 {
			continue
		}
		item := CreditRepayReminder{
			AccountID:     acc.ID,
			AccountName:   acc.Name,
			PaymentDueDay: acc.PaymentDueDay,
			DueDate:       fmtDate(due),
			DaysUntilDue:  days,
			AmountFen:     amount,
		}
		if defaultFrom != nil {
			id := defaultFrom.ID
			item.FromAccountID = &id
			item.FromAccountName = defaultFrom.Name
		}
		out = append(out, item)
	}
	return out, nil
}

// paymentDueForStatement 出账日对应的还款日
// 还款日 > 账单日 → 同月；否则 → 出账月的下月
func paymentDueForStatement(stmt time.Time, billingDay, paymentDueDay int) time.Time {
	loc := stmt.Location()
	if paymentDueDay > billingDay {
		return clampDay(stmt.Year(), stmt.Month(), paymentDueDay, loc)
	}
	n := stmt.AddDate(0, 1, 0)
	return clampDay(n.Year(), n.Month(), paymentDueDay, loc)
}

// relevantPaymentDueDate 相对今天应关注的还款日（已过期则看下期）
func relevantPaymentDueDate(billingDay, paymentDueDay int, today time.Time) time.Time {
	stmt := lastClosedStatementDate(billingDay, today)
	due := paymentDueForStatement(stmt, billingDay, paymentDueDay)
	if !due.Before(today) {
		return due
	}
	nextStmt := nextStatementDate(stmt, billingDay)
	return paymentDueForStatement(nextStmt, billingDay, paymentDueDay)
}

