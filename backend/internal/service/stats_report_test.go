package service

import (
	"testing"
	"time"

	"mybookkeeping/internal/model"
)

func TestStatsReports(t *testing.T) {
	f := newTenantFixture(t)
	txs := NewTransactionService(f.db)
	stats := NewStatsService(f.db)
	loc := time.Local

	child := &model.Category{UserID: f.a, Name: "A 午餐", Kind: "expense", ParentID: &f.aCat}
	salary := &model.Category{UserID: f.a, Name: "A 工资", Kind: "income"}
	credit := &model.Account{UserID: f.a, Name: "A 信用卡", Type: "credit"}
	tag := &model.Tag{UserID: f.a, Name: "A 出差"}
	for _, v := range []any{child, salary, credit, tag} {
		if err := f.db.Create(v).Error; err != nil {
			t.Fatal(err)
		}
	}

	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 12, 0, 0, 0, loc) }
	mk := func(in TransactionInput) {
		t.Helper()
		if _, err := txs.Create(f.a, in); err != nil {
			t.Fatal(err)
		}
	}
	mk(TransactionInput{Type: "income", Amount: 10000, AccountID: f.aAcc, CategoryID: &salary.ID, HappenedAt: day(1, 5)})
	mk(TransactionInput{Type: "expense", Amount: 1000, AccountID: f.aAcc, CategoryID: &f.aCat, HappenedAt: day(1, 10)})
	mk(TransactionInput{Type: "income", Amount: 20000, AccountID: f.aAcc, CategoryID: &salary.ID, HappenedAt: day(2, 5)})
	mk(TransactionInput{Type: "expense", Amount: 300, AccountID: f.aAcc, CategoryID: &f.aCat, HappenedAt: day(2, 6)})
	mk(TransactionInput{Type: "expense", Amount: 500, AccountID: f.aAcc, CategoryID: &child.ID, HappenedAt: day(2, 6), TagIDs: []uint64{tag.ID}})
	// 历史数据可能没有分类：直接落库并同步余额
	if err := f.db.Create(&model.Transaction{UserID: f.a, Type: "expense", Amount: 700, AccountID: credit.ID, HappenedAt: day(2, 20)}).Error; err != nil {
		t.Fatal(err)
	}
	f.db.Model(credit).Update("balance", -700)
	mk(TransactionInput{Type: "transfer", Amount: 400, AccountID: f.aAcc, ToAccountID: &credit.ID, HappenedAt: day(3, 1)})

	feb := StatsQuery{From: time.Date(2026, 2, 1, 0, 0, 0, 0, loc), To: time.Date(2026, 3, 1, 0, 0, 0, 0, loc)}

	ov, err := stats.Overview(f.a, feb)
	if err != nil {
		t.Fatal(err)
	}
	if ov.Expense != 1500 || ov.Income != 20000 || ov.Net != 18500 || ov.ExpenseCount != 3 || ov.MaxExpense != 700 {
		t.Fatalf("overview=%+v", ov)
	}
	if ov.Prev.From != "2026-01-01" || ov.Prev.Expense != 1000 || ov.Prev.Income != 10000 {
		t.Fatalf("prev=%+v", ov.Prev)
	}
	if ov.Days != 28 {
		t.Fatalf("days=%d", ov.Days)
	}

	bd, err := stats.Breakdown(f.a, BreakdownQuery{StatsQuery: feb, Kind: "expense", By: "category"})
	if err != nil {
		t.Fatal(err)
	}
	if bd.Total != 1500 || len(bd.Items) != 2 || bd.Items[0].ID != f.aCat || bd.Items[0].Amount != 800 ||
		!bd.Items[0].HasChildren || bd.Items[0].Count != 2 || bd.Items[1].ID != 0 || bd.Items[1].Amount != 700 {
		t.Fatalf("category breakdown=%+v", bd)
	}
	bd, _ = stats.Breakdown(f.a, BreakdownQuery{StatsQuery: feb, Kind: "expense", By: "category", ParentID: &f.aCat})
	if bd.Total != 800 || len(bd.Items) != 2 || bd.Items[0].ID != child.ID || bd.Items[1].Name != "A 餐饮（未细分）" {
		t.Fatalf("child breakdown=%+v", bd)
	}
	bd, _ = stats.Breakdown(f.a, BreakdownQuery{StatsQuery: feb, Kind: "expense", By: "account"})
	if len(bd.Items) != 2 || bd.Items[0].ID != f.aAcc || bd.Items[0].Amount != 800 || bd.Items[1].AccountType != "credit" {
		t.Fatalf("account breakdown=%+v", bd)
	}
	bd, _ = stats.Breakdown(f.a, BreakdownQuery{StatsQuery: feb, Kind: "expense", By: "tag"})
	if len(bd.Items) != 2 || bd.Items[0].ID != 0 || bd.Items[0].Amount != 1000 || bd.Items[1].ID != tag.ID || bd.Items[1].Amount != 500 {
		t.Fatalf("tag breakdown=%+v", bd)
	}
	if bd, _ := stats.Breakdown(f.b, BreakdownQuery{StatsQuery: feb, By: "tag"}); bd.Total != 0 || len(bd.Items) != 0 {
		t.Fatalf("B 不应统计到 A 的数据: %+v", bd)
	}

	q1 := StatsQuery{From: time.Date(2026, 1, 1, 0, 0, 0, 0, loc), To: time.Date(2026, 4, 1, 0, 0, 0, 0, loc)}
	tr, err := stats.Trend(f.a, q1, GranularityMonth)
	if err != nil {
		t.Fatal(err)
	}
	if tr.Granularity != GranularityMonth || len(tr.Points) != 3 || tr.Points[1].Key != "2026-02" || tr.Points[1].Net != 18500 {
		t.Fatalf("trend=%+v", tr)
	}
	tr, _ = stats.Trend(f.a, feb, GranularityDay)
	if len(tr.Points) != 28 || tr.Points[5].Key != "2026-02-06" || tr.Points[5].Expense != 800 || tr.Points[5].ExpenseCount != 2 {
		t.Fatalf("daily trend=%+v", tr.Points[5])
	}

	as, err := stats.AssetTrend(f.a, q1, GranularityMonth)
	if err != nil {
		t.Fatal(err)
	}
	// 现金：+10000 -1000 +20000 -300 -500 -400 = 27800；信用卡：-700 +400 = -300
	if as.Assets != 27800 || as.Liabilities != 300 || as.Net != 27500 || len(as.Points) != 3 {
		t.Fatalf("assets=%+v", as)
	}
	want := []AssetPoint{
		{Key: "2026-01", Assets: 9000, Liabilities: 0, Net: 9000},
		{Key: "2026-02", Assets: 28200, Liabilities: 700, Net: 27500},
		{Key: "2026-03", Assets: 27800, Liabilities: 300, Net: 27500},
	}
	for i, p := range as.Points {
		if p != want[i] {
			t.Fatalf("asset point %d=%+v want %+v", i, p, want[i])
		}
	}

	lng, lat, lng2 := 120.15501, 30.27401, 120.15503
	mk(TransactionInput{Type: "expense", Amount: 50, AccountID: f.aAcc, CategoryID: &f.aCat, HappenedAt: day(2, 7), GeoLng: &lng, GeoLat: &lat, GeoName: "西湖店"})
	mk(TransactionInput{Type: "expense", Amount: 70, AccountID: f.aAcc, CategoryID: &f.aCat, HappenedAt: day(2, 8), GeoLng: &lng2, GeoLat: &lat, GeoName: "西湖店"})
	pl, err := stats.Places(f.a, feb, "expense")
	if err != nil {
		t.Fatal(err)
	}
	if len(pl.Places) != 1 || pl.Places[0].Amount != 120 || pl.Places[0].Count != 2 ||
		pl.LocatedCount != 2 || pl.MissingCount != 3 || pl.MissingAmount != 1500 {
		t.Fatalf("places=%+v", pl)
	}
	if pl, _ := stats.Places(f.b, feb, "expense"); len(pl.Places) != 0 {
		t.Fatalf("B 不应看到 A 的地点: %+v", pl)
	}

	top, err := stats.TopTransactions(f.a, feb, "expense", 2)
	if err != nil || len(top) != 2 || top[0].Amount != 700 || top[1].Amount != 500 || top[1].Category == nil {
		t.Fatalf("top=%+v err=%v", top, err)
	}
}
