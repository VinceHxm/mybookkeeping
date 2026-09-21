package service

import (
	"time"

	"gorm.io/gorm"
)

type CategoryStat struct {
	CategoryID   uint64 `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Amount       int64  `json:"amount"`
}

type MonthTrend struct {
	Month   string `json:"month"`
	Expense int64  `json:"expense"`
	Income  int64  `json:"income"`
}

type Summary struct {
	From         string         `json:"from"`
	To           string         `json:"to"`
	MonthExpense int64          `json:"monthExpense"` // 兼容旧字段名：区间支出
	MonthIncome  int64          `json:"monthIncome"`
	Expense      int64          `json:"expense"`
	Income       int64          `json:"income"`
	ByCategory   []CategoryStat `json:"byCategory"`
	Trends       []MonthTrend   `json:"trends"`
}

type StatsQuery struct {
	From      time.Time
	To        time.Time
	AccountID *uint64
}

type StatsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
}

func (s *StatsService) Summary(userID uint64, q StatsQuery) (*Summary, error) {
	loc := time.Local
	now := time.Now().In(loc)
	from := q.From
	to := q.To
	if from.IsZero() {
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}
	if to.IsZero() {
		to = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0)
	}
	if !to.After(from) {
		to = from.AddDate(0, 1, 0)
	}

	out := &Summary{
		From: from.Format("2006-01-02"),
		To:   to.Format("2006-01-02"),
	}

	txFilter := func(db *gorm.DB) *gorm.DB {
		db = db.Where("user_id = ? AND happened_at >= ? AND happened_at < ?", userID, from, to)
		if q.AccountID != nil {
			db = db.Where("account_id = ? OR to_account_id = ?", *q.AccountID, *q.AccountID)
		}
		return db
	}

	type agg struct {
		Type  string
		Total int64
	}
	var aggs []agg
	if err := txFilter(s.db.Table("transactions")).
		Select("type, COALESCE(SUM(amount),0) as total").
		Where("type IN ?", []string{"expense", "income"}).
		Group("type").Scan(&aggs).Error; err != nil {
		return nil, err
	}
	for _, a := range aggs {
		if a.Type == "expense" {
			out.Expense = a.Total
			out.MonthExpense = a.Total
		} else if a.Type == "income" {
			out.Income = a.Total
			out.MonthIncome = a.Total
		}
	}

	type catRow struct {
		CategoryID   uint64
		CategoryName string
		Amount       int64
	}
	var cats []catRow
	qdb := s.db.Table("transactions AS t").
		Select("t.category_id as category_id, COALESCE(c.name,'未分类') as category_name, COALESCE(SUM(t.amount),0) as amount").
		Joins("LEFT JOIN categories c ON c.id = t.category_id").
		Where("t.user_id = ? AND t.type = ? AND t.happened_at >= ? AND t.happened_at < ?", userID, "expense", from, to)
	if q.AccountID != nil {
		qdb = qdb.Where("t.account_id = ? OR t.to_account_id = ?", *q.AccountID, *q.AccountID)
	}
	if err := qdb.Group("t.category_id, c.name").Order("amount desc").Scan(&cats).Error; err != nil {
		return nil, err
	}
	for _, c := range cats {
		out.ByCategory = append(out.ByCategory, CategoryStat{
			CategoryID: c.CategoryID, CategoryName: c.CategoryName, Amount: c.Amount,
		})
	}

	// 趋势：以 to 所在月为终点，向前最多 6 个自然月（覆盖区间所跨月）
	monthEnd := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, loc)
	monthStart := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, loc)
	months := 0
	for m := monthStart; !m.After(monthEnd); m = m.AddDate(0, 1, 0) {
		months++
	}
	if months < 1 {
		months = 1
	}
	if months > 12 {
		// 过长区间只展示最近 12 个月趋势
		monthStart = monthEnd.AddDate(0, -11, 0)
		months = 12
	}
	trendFrom := monthStart
	trendTo := monthEnd.AddDate(0, 1, 0)

	type monthRow struct {
		Ym     string
		Type   string
		Amount int64
	}
	var rows []monthRow
	tdb := s.db.Table("transactions").
		Select("DATE_FORMAT(happened_at, '%Y-%m') as ym, type, COALESCE(SUM(amount),0) as amount").
		Where("user_id = ? AND happened_at >= ? AND happened_at < ? AND type IN ?", userID, trendFrom, trendTo, []string{"expense", "income"})
	if q.AccountID != nil {
		tdb = tdb.Where("account_id = ? OR to_account_id = ?", *q.AccountID, *q.AccountID)
	}
	if err := tdb.Group("ym, type").Scan(&rows).Error; err != nil {
		return nil, err
	}
	trendMap := map[string]*MonthTrend{}
	for i := 0; i < months; i++ {
		m := trendFrom.AddDate(0, i, 0)
		key := m.Format("2006-01")
		trendMap[key] = &MonthTrend{Month: key}
	}
	for _, r := range rows {
		t, ok := trendMap[r.Ym]
		if !ok {
			continue
		}
		if r.Type == "expense" {
			t.Expense = r.Amount
		} else {
			t.Income = r.Amount
		}
	}
	for i := 0; i < months; i++ {
		m := trendFrom.AddDate(0, i, 0)
		key := m.Format("2006-01")
		out.Trends = append(out.Trends, *trendMap[key])
	}
	return out, nil
}
