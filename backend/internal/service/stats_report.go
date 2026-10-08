package service

import (
	"sort"
	"time"

	"gorm.io/gorm"

	"mybookkeeping/internal/model"
)

const (
	GranularityDay   = "day"
	GranularityMonth = "month"
	GranularityYear  = "year"
)

// statTx 统计只需要的流水列
type statTx struct {
	ID          uint64
	Type        string
	Amount      int64
	AccountID   uint64
	ToAccountID *uint64
	CategoryID  *uint64
	HappenedAt  time.Time
}

type PeriodTotals struct {
	From         string `json:"from"`
	To           string `json:"to"`
	Expense      int64  `json:"expense"`
	Income       int64  `json:"income"`
	Net          int64  `json:"net"`
	ExpenseCount int    `json:"expenseCount"`
	IncomeCount  int    `json:"incomeCount"`
}

type StatsOverview struct {
	PeriodTotals
	// Days 区间内已经过去的天数（日均用；未来日期不计）
	Days       int          `json:"days"`
	MaxExpense int64        `json:"maxExpense"`
	Prev       PeriodTotals `json:"prev"`
}

type BreakdownItem struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	AccountType string `json:"accountType,omitempty"`
	Amount      int64  `json:"amount"`
	Count       int    `json:"count"`
	HasChildren bool   `json:"hasChildren"`
}

type Breakdown struct {
	Kind  string          `json:"kind"`
	By    string          `json:"by"`
	Total int64           `json:"total"`
	Items []BreakdownItem `json:"items"`
}

type BreakdownQuery struct {
	StatsQuery
	Kind     string // expense | income
	By       string // category | account | tag
	ParentID *uint64
}

type TrendPoint struct {
	Key          string `json:"key"`
	Expense      int64  `json:"expense"`
	Income       int64  `json:"income"`
	Net          int64  `json:"net"`
	ExpenseCount int    `json:"expenseCount"`
	IncomeCount  int    `json:"incomeCount"`
}

type Trend struct {
	Granularity string       `json:"granularity"`
	Points      []TrendPoint `json:"points"`
}

type AssetPoint struct {
	Key         string `json:"key"`
	Assets      int64  `json:"assets"`
	Liabilities int64  `json:"liabilities"`
	Net         int64  `json:"net"`
}

type AssetAccount struct {
	ID      uint64 `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Balance int64  `json:"balance"`
}

type AssetTrend struct {
	Granularity string         `json:"granularity"`
	Assets      int64          `json:"assets"`
	Liabilities int64          `json:"liabilities"`
	Net         int64          `json:"net"`
	Points      []AssetPoint   `json:"points"`
	Accounts    []AssetAccount `json:"accounts"`
}

// statsRange 区间 [from, to)，缺省为本月
func statsRange(q StatsQuery) (time.Time, time.Time) {
	loc := time.Local
	now := time.Now().In(loc)
	from, to := q.From, q.To
	if from.IsZero() {
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}
	if to.IsZero() {
		to = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0)
	}
	if !to.After(from) {
		to = from.AddDate(0, 1, 0)
	}
	return from, to
}

// prevRange 上一个同等长度区间：整月区间按月平移，其余按时长平移
func prevRange(from, to time.Time) (time.Time, time.Time) {
	if isMonthStart(from) && isMonthStart(to) {
		months := (to.Year()-from.Year())*12 + int(to.Month()-from.Month())
		if months > 0 {
			return from.AddDate(0, -months, 0), from
		}
	}
	return from.Add(-to.Sub(from)), from
}

func isMonthStart(t time.Time) bool {
	return t.Day() == 1 && t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0
}

func startOfDay(t time.Time) time.Time {
	t = t.In(time.Local)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func periodKey(t time.Time, g string) string {
	t = t.In(time.Local)
	switch g {
	case GranularityDay:
		return t.Format("2006-01-02")
	case GranularityYear:
		return t.Format("2006")
	}
	return t.Format("2006-01")
}

// periodKeys [from, to) 内每个周期的 key，按时间升序
func periodKeys(from, to time.Time, g string) []string {
	f := from.In(time.Local)
	var cur time.Time
	switch g {
	case GranularityDay:
		cur = startOfDay(f)
	case GranularityYear:
		cur = time.Date(f.Year(), 1, 1, 0, 0, 0, 0, time.Local)
	default:
		cur = time.Date(f.Year(), f.Month(), 1, 0, 0, 0, 0, time.Local)
	}
	var keys []string
	for cur.Before(to) {
		keys = append(keys, periodKey(cur, g))
		switch g {
		case GranularityDay:
			cur = cur.AddDate(0, 0, 1)
		case GranularityYear:
			cur = cur.AddDate(1, 0, 0)
		default:
			cur = cur.AddDate(0, 1, 0)
		}
	}
	return keys
}

// fitGranularity 未指定或点数过多时自动选粒度，避免一次返回上千个点
func fitGranularity(from, to time.Time, want string) string {
	days := int(to.Sub(from).Hours() / 24)
	months := days / 30
	switch want {
	case GranularityDay:
		if days <= 366 {
			return GranularityDay
		}
		fallthrough
	case GranularityMonth:
		if months <= 120 {
			return GranularityMonth
		}
		return GranularityYear
	case GranularityYear:
		return GranularityYear
	}
	switch {
	case days <= 62:
		return GranularityDay
	case months <= 36:
		return GranularityMonth
	default:
		return GranularityYear
	}
}

func (s *StatsService) loadTx(userID uint64, from, to time.Time, accountID *uint64, types ...string) ([]statTx, error) {
	db := s.db.Table("transactions").
		Select("id, type, amount, account_id, to_account_id, category_id, happened_at").
		Where("user_id = ? AND happened_at >= ?", userID, from)
	if !to.IsZero() {
		db = db.Where("happened_at < ?", to)
	}
	if accountID != nil {
		db = db.Where("account_id = ? OR to_account_id = ?", *accountID, *accountID)
	}
	if len(types) > 0 {
		db = db.Where("type IN ?", types)
	}
	var rows []statTx
	err := db.Scan(&rows).Error
	return rows, err
}

func (s *StatsService) totals(userID uint64, from, to time.Time, accountID *uint64) (PeriodTotals, int64, error) {
	out := PeriodTotals{From: from.Format("2006-01-02"), To: to.Format("2006-01-02")}
	rows, err := s.loadTx(userID, from, to, accountID, "expense", "income")
	if err != nil {
		return out, 0, err
	}
	var maxExpense int64
	for _, r := range rows {
		if r.Type == "expense" {
			out.Expense += r.Amount
			out.ExpenseCount++
			if r.Amount > maxExpense {
				maxExpense = r.Amount
			}
		} else {
			out.Income += r.Amount
			out.IncomeCount++
		}
	}
	out.Net = out.Income - out.Expense
	return out, maxExpense, nil
}

// Overview 区间收支汇总 + 上一同等区间对比
func (s *StatsService) Overview(userID uint64, q StatsQuery) (*StatsOverview, error) {
	from, to := statsRange(q)
	cur, maxExpense, err := s.totals(userID, from, to, q.AccountID)
	if err != nil {
		return nil, err
	}
	pf, pt := prevRange(from, to)
	prev, _, err := s.totals(userID, pf, pt, q.AccountID)
	if err != nil {
		return nil, err
	}
	end := to
	if tomorrow := startOfDay(time.Now()).AddDate(0, 0, 1); end.After(tomorrow) {
		end = tomorrow
	}
	days := int((end.Sub(from).Hours() + 23) / 24)
	if days < 1 {
		days = 1
	}
	return &StatsOverview{PeriodTotals: cur, Days: days, MaxExpense: maxExpense, Prev: prev}, nil
}

// Breakdown 收支构成：按分类（一级，或某一级下的二级）/ 账户 / 标签
func (s *StatsService) Breakdown(userID uint64, q BreakdownQuery) (*Breakdown, error) {
	if q.Kind != "income" {
		q.Kind = "expense"
	}
	if q.By != "account" && q.By != "tag" {
		q.By = "category"
	}
	from, to := statsRange(q.StatsQuery)
	rows, err := s.loadTx(userID, from, to, q.AccountID, q.Kind)
	if err != nil {
		return nil, err
	}
	out := &Breakdown{Kind: q.Kind, By: q.By, Items: []BreakdownItem{}}
	items := map[uint64]*BreakdownItem{}
	add := func(id uint64, amount int64, mk func() BreakdownItem) {
		it, ok := items[id]
		if !ok {
			v := mk()
			v.ID = id
			it = &v
			items[id] = it
		}
		it.Amount += amount
		it.Count++
	}

	switch q.By {
	case "category":
		var cats []model.Category
		if err := s.db.Select("id", "parent_id", "name", "icon").
			Where("user_id = ?", userID).Find(&cats).Error; err != nil {
			return nil, err
		}
		byID := map[uint64]model.Category{}
		hasChild := map[uint64]bool{}
		for _, c := range cats {
			byID[c.ID] = c
			if c.ParentID != nil {
				hasChild[*c.ParentID] = true
			}
		}
		for _, r := range rows {
			var cat model.Category
			var found bool
			if r.CategoryID != nil {
				cat, found = byID[*r.CategoryID]
			}
			if q.ParentID != nil {
				if !found || (cat.ID != *q.ParentID && (cat.ParentID == nil || *cat.ParentID != *q.ParentID)) {
					continue
				}
				c := cat
				add(c.ID, r.Amount, func() BreakdownItem {
					name := c.Name
					if c.ID == *q.ParentID {
						name = c.Name + "（未细分）"
					}
					return BreakdownItem{Name: name, Icon: c.Icon}
				})
				out.Total += r.Amount
				continue
			}
			out.Total += r.Amount
			if !found {
				add(0, r.Amount, func() BreakdownItem { return BreakdownItem{Name: "未分类"} })
				continue
			}
			top := cat
			if cat.ParentID != nil {
				if p, ok := byID[*cat.ParentID]; ok {
					top = p
				}
			}
			add(top.ID, r.Amount, func() BreakdownItem {
				return BreakdownItem{Name: top.Name, Icon: top.Icon, HasChildren: hasChild[top.ID]}
			})
		}

	case "account":
		var accs []model.Account
		if err := s.db.Select("id", "name", "type").Where("user_id = ?", userID).Find(&accs).Error; err != nil {
			return nil, err
		}
		byID := map[uint64]model.Account{}
		for _, a := range accs {
			byID[a.ID] = a
		}
		for _, r := range rows {
			out.Total += r.Amount
			a := byID[r.AccountID]
			add(r.AccountID, r.Amount, func() BreakdownItem {
				return BreakdownItem{Name: a.Name, AccountType: a.Type}
			})
		}

	case "tag":
		type tagRow struct {
			TransactionID uint64
			TagID         uint64
			Name          string
			Color         string
		}
		var links []tagRow
		db := s.db.Table("transaction_tags AS tt").
			Select("tt.transaction_id, tt.tag_id, tg.name, tg.color").
			Joins("JOIN transactions t ON t.id = tt.transaction_id").
			Joins("JOIN tags tg ON tg.id = tt.tag_id").
			Where("t.user_id = ? AND tg.user_id = ? AND t.type = ? AND t.happened_at >= ? AND t.happened_at < ?",
				userID, userID, q.Kind, from, to)
		if q.AccountID != nil {
			db = db.Where("t.account_id = ? OR t.to_account_id = ?", *q.AccountID, *q.AccountID)
		}
		if err := db.Scan(&links).Error; err != nil {
			return nil, err
		}
		tagsOf := map[uint64][]tagRow{}
		for _, l := range links {
			tagsOf[l.TransactionID] = append(tagsOf[l.TransactionID], l)
		}
		for _, r := range rows {
			out.Total += r.Amount
			tags := tagsOf[r.ID]
			if len(tags) == 0 {
				add(0, r.Amount, func() BreakdownItem { return BreakdownItem{Name: "无标签"} })
				continue
			}
			for _, tg := range tags {
				t := tg
				add(t.TagID, r.Amount, func() BreakdownItem { return BreakdownItem{Name: t.Name, Color: t.Color} })
			}
		}
	}

	for _, it := range items {
		out.Items = append(out.Items, *it)
	}
	sort.Slice(out.Items, func(i, j int) bool {
		if out.Items[i].Amount != out.Items[j].Amount {
			return out.Items[i].Amount > out.Items[j].Amount
		}
		return out.Items[i].ID < out.Items[j].ID
	})
	return out, nil
}

// Trend 按日 / 月 / 年的收支与结余
func (s *StatsService) Trend(userID uint64, q StatsQuery, granularity string) (*Trend, error) {
	from, to := statsRange(q)
	g := fitGranularity(from, to, granularity)
	rows, err := s.loadTx(userID, from, to, q.AccountID, "expense", "income")
	if err != nil {
		return nil, err
	}
	keys := periodKeys(from, to, g)
	idx := make(map[string]int, len(keys))
	out := &Trend{Granularity: g, Points: make([]TrendPoint, len(keys))}
	for i, k := range keys {
		idx[k] = i
		out.Points[i].Key = k
	}
	for _, r := range rows {
		i, ok := idx[periodKey(r.HappenedAt, g)]
		if !ok {
			continue
		}
		p := &out.Points[i]
		if r.Type == "expense" {
			p.Expense += r.Amount
			p.ExpenseCount++
		} else {
			p.Income += r.Amount
			p.IncomeCount++
		}
	}
	for i := range out.Points {
		out.Points[i].Net = out.Points[i].Income - out.Points[i].Expense
	}
	return out, nil
}

// AssetTrend 资产 / 负债 / 净资产走势：由当前余额倒推每个周期末的余额
func (s *StatsService) AssetTrend(userID uint64, q StatsQuery, granularity string) (*AssetTrend, error) {
	from, to := statsRange(q)
	if tomorrow := startOfDay(time.Now()).AddDate(0, 0, 1); to.After(tomorrow) {
		to = tomorrow
	}
	if !to.After(from) {
		// 整个区间都在未来：只看当天
		from = startOfDay(time.Now())
	}
	g := fitGranularity(from, to, granularity)

	accDB := s.db.Select("id", "name", "type", "balance", "archived").Where("user_id = ?", userID)
	if q.AccountID != nil {
		accDB = accDB.Where("id = ?", *q.AccountID)
	}
	var accs []model.Account
	if err := accDB.Order("sort asc, id asc").Find(&accs).Error; err != nil {
		return nil, err
	}
	bal := make(map[uint64]int64, len(accs))
	out := &AssetTrend{Granularity: g, Points: []AssetPoint{}, Accounts: []AssetAccount{}}
	for _, a := range accs {
		bal[a.ID] = a.Balance
		if !a.Archived && a.Balance != 0 {
			out.Accounts = append(out.Accounts, AssetAccount{ID: a.ID, Name: a.Name, Type: a.Type, Balance: a.Balance})
		}
	}
	sort.SliceStable(out.Accounts, func(i, j int) bool { return out.Accounts[i].Balance > out.Accounts[j].Balance })
	out.Assets, out.Liabilities, out.Net = sumBalances(bal)

	rows, err := s.loadTx(userID, from, time.Time{}, q.AccountID, "expense", "income", "transfer", TxTypeRevalue)
	if err != nil {
		return nil, err
	}
	// deltas[key][accountID]：该周期内流水对余额的影响；to 之后的统一归到 afterKey
	const afterKey = "~"
	deltas := map[string]map[uint64]int64{}
	addDelta := func(key string, accountID uint64, d int64) {
		if _, ok := bal[accountID]; !ok {
			return
		}
		m := deltas[key]
		if m == nil {
			m = map[uint64]int64{}
			deltas[key] = m
		}
		m[accountID] += d
	}
	for _, r := range rows {
		key := afterKey
		if r.HappenedAt.Before(to) {
			key = periodKey(r.HappenedAt, g)
		}
		switch r.Type {
		case "expense":
			addDelta(key, r.AccountID, -r.Amount)
		case "income", TxTypeRevalue:
			addDelta(key, r.AccountID, r.Amount)
		case "transfer":
			addDelta(key, r.AccountID, -r.Amount)
			if r.ToAccountID != nil {
				addDelta(key, *r.ToAccountID, r.Amount)
			}
		}
	}
	undo := func(key string) {
		for id, d := range deltas[key] {
			bal[id] -= d
		}
	}
	undo(afterKey)
	keys := periodKeys(from, to, g)
	out.Points = make([]AssetPoint, len(keys))
	for i := len(keys) - 1; i >= 0; i-- {
		a, l, n := sumBalances(bal)
		out.Points[i] = AssetPoint{Key: keys[i], Assets: a, Liabilities: l, Net: n}
		undo(keys[i])
	}
	return out, nil
}

func sumBalances(bal map[uint64]int64) (assets, liabilities, net int64) {
	for _, b := range bal {
		if b >= 0 {
			assets += b
		} else {
			liabilities -= b
		}
		net += b
	}
	return
}

type PlaceStat struct {
	Name   string    `json:"name"`
	Lng    float64   `json:"lng"`
	Lat    float64   `json:"lat"`
	Amount int64     `json:"amount"`
	Count  int       `json:"count"`
	LastAt time.Time `json:"lastAt"`
}

type PlaceStats struct {
	Kind          string      `json:"kind"`
	Places        []PlaceStat `json:"places"`
	LocatedAmount int64       `json:"locatedAmount"`
	LocatedCount  int         `json:"locatedCount"`
	MissingAmount int64       `json:"missingAmount"`
	MissingCount  int         `json:"missingCount"`
	Truncated     bool        `json:"truncated"`
}

// maxPlaces 地图上最多画这么多个点，再多既看不清也拖慢页面
const maxPlaces = 500

// Places 按地点聚合收支：同名且坐标相近（约 10 米）的流水合并为一个点；行程类取起点
func (s *StatsService) Places(userID uint64, q StatsQuery, kind string) (*PlaceStats, error) {
	if kind != "income" {
		kind = "expense"
	}
	from, to := statsRange(q)
	db := s.db.Table("transactions").
		Select("amount, geo_lng, geo_lat, geo_name, happened_at").
		Where("user_id = ? AND type = ? AND happened_at >= ? AND happened_at < ?", userID, kind, from, to)
	if q.AccountID != nil {
		db = db.Where("account_id = ? OR to_account_id = ?", *q.AccountID, *q.AccountID)
	}
	type row struct {
		Amount     int64
		GeoLng     *float64
		GeoLat     *float64
		GeoName    string
		HappenedAt time.Time
	}
	var rows []row
	if err := db.Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := &PlaceStats{Kind: kind, Places: []PlaceStat{}}
	type key struct {
		name     string
		lng, lat int64
	}
	byKey := map[key]*PlaceStat{}
	for _, r := range rows {
		if r.GeoLng == nil || r.GeoLat == nil {
			out.MissingAmount += r.Amount
			out.MissingCount++
			continue
		}
		out.LocatedAmount += r.Amount
		out.LocatedCount++
		k := key{name: r.GeoName, lng: int64(*r.GeoLng * 1e4), lat: int64(*r.GeoLat * 1e4)}
		p, ok := byKey[k]
		if !ok {
			p = &PlaceStat{Name: r.GeoName, Lng: *r.GeoLng, Lat: *r.GeoLat}
			byKey[k] = p
		}
		p.Amount += r.Amount
		p.Count++
		if r.HappenedAt.After(p.LastAt) {
			p.LastAt = r.HappenedAt
		}
	}
	for _, p := range byKey {
		out.Places = append(out.Places, *p)
	}
	sort.Slice(out.Places, func(i, j int) bool {
		if out.Places[i].Amount != out.Places[j].Amount {
			return out.Places[i].Amount > out.Places[j].Amount
		}
		return out.Places[i].LastAt.After(out.Places[j].LastAt)
	})
	if len(out.Places) > maxPlaces {
		out.Places = out.Places[:maxPlaces]
		out.Truncated = true
	}
	return out, nil
}

// TopTransactions 区间内金额最大的若干笔收支
func (s *StatsService) TopTransactions(userID uint64, q StatsQuery, kind string, limit int) ([]model.Transaction, error) {
	if kind != "income" {
		kind = "expense"
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	from, to := statsRange(q)
	db := s.db.Where("user_id = ? AND type = ? AND happened_at >= ? AND happened_at < ?", userID, kind, from, to)
	if q.AccountID != nil {
		db = db.Where("account_id = ? OR to_account_id = ?", *q.AccountID, *q.AccountID)
	}
	list := []model.Transaction{}
	err := db.Preload("Account", briefAccount).
		Preload("Category", func(db *gorm.DB) *gorm.DB { return db.Select("id", "name", "kind", "icon", "parent_id") }).
		Preload("Tags").
		Order("amount desc, happened_at desc").Limit(limit).Find(&list).Error
	return list, err
}
