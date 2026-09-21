package service

import (
	"testing"
	"time"

	"mybookkeeping/internal/model"
)

func TestCalcFare_FreePeriod(t *testing.T) {
	rule := &model.FareRule{
		Enabled:  true,
		BaseFen:  200,
		CardRate: 0.9,
		FreePeriods: []model.FareFreePeriod{
			{StartDate: "2026-05-01", EndDate: "2026-05-07", Recur: "none"},
		},
	}
	at := time.Date(2026, 5, 3, 10, 0, 0, 0, time.Local)
	r := CalcFare(rule, at, 0, rule.BaseFen)
	if r.AmountFen != 0 || !r.Applied.Free {
		t.Fatalf("expected free, got %+v", r)
	}
}

func TestCalcFare_Yearly(t *testing.T) {
	rule := &model.FareRule{
		Enabled:  true,
		BaseFen:  200,
		CardRate: 1,
		FreePeriods: []model.FareFreePeriod{
			{StartDate: "2020-05-01", EndDate: "2020-05-07", Recur: "yearly"},
		},
	}
	at := time.Date(2026, 5, 3, 0, 0, 0, 0, time.Local)
	r := CalcFare(rule, at, 0, rule.BaseFen)
	if !r.Applied.Free {
		t.Fatalf("expected yearly free, got %+v", r)
	}
}

func TestCalcFare_Weekly(t *testing.T) {
	rule := &model.FareRule{
		Enabled:  true,
		BaseFen:  200,
		CardRate: 1,
		FreePeriods: []model.FareFreePeriod{
			{Recur: "weekly", Weekdays: []int{6}},
		},
	}
	sat := time.Date(2026, 9, 19, 0, 0, 0, 0, time.Local)
	if !CalcFare(rule, sat, 0, rule.BaseFen).Applied.Free {
		t.Fatalf("Saturday should be free")
	}
}

func TestCalcFare_CardAndTier(t *testing.T) {
	// 旧语义兼容：票卡后再折
	rule := &model.FareRule{
		Enabled:  true,
		BaseFen:  1000,
		CardRate: 0.9,
		Tiers: []model.FareTier{
			{MinFen: 10000, MaxFen: 0, Rate: 0.8, Base: "card"},
		},
	}
	r := CalcFare(rule, time.Now(), 15000, rule.BaseFen)
	if r.AmountFen != 720 {
		t.Fatalf("expected 720 fen, got %d", r.AmountFen)
	}
}

func TestCalcFare_FuzhouStyle(t *testing.T) {
	// 0~80 票卡9折；超80~150 全价7折；超150~300 全价5折；超300 恢复票卡
	rule := &model.FareRule{
		Enabled:  true,
		BaseFen:  200, // 2元
		CardRate: 0.9,
		Tiers: []model.FareTier{
			{MinFen: 8000, MaxFen: 15000, Rate: 0.7, Base: "full"},
			{MinFen: 15000, MaxFen: 30000, Rate: 0.5, Base: "full"},
		},
	}
	at := time.Now()
	base := rule.BaseFen
	// spent=80 → 仍票卡
	r0 := CalcFare(rule, at, 8000, base)
	if r0.AmountFen != 180 {
		t.Fatalf("at 80 expect card 1.8yuan=%d, got %d %s", 180, r0.AmountFen, r0.Reason)
	}
	// spent=80.01 → 全价7折
	r1 := CalcFare(rule, at, 8001, base)
	if r1.AmountFen != 140 {
		t.Fatalf("expect full 7折 140, got %d %s", r1.AmountFen, r1.Reason)
	}
	// spent=150 → 仍全价7折（含）
	r2 := CalcFare(rule, at, 15000, base)
	if r2.AmountFen != 140 {
		t.Fatalf("at 150 expect 140, got %d", r2.AmountFen)
	}
	// spent=150.01 → 全价5折
	r3 := CalcFare(rule, at, 15001, base)
	if r3.AmountFen != 100 {
		t.Fatalf("expect 100, got %d", r3.AmountFen)
	}
	// spent=300 → 全价5折
	r4 := CalcFare(rule, at, 30000, base)
	if r4.AmountFen != 100 {
		t.Fatalf("at 300 expect 100, got %d", r4.AmountFen)
	}
	// spent>300 → 票卡
	r5 := CalcFare(rule, at, 30001, base)
	if r5.AmountFen != 180 {
		t.Fatalf("over 300 expect card 180, got %d %s", r5.AmountFen, r5.Reason)
	}
}

func TestCalcFare_BaseOverride(t *testing.T) {
	// 规则内原价 2 元，外部传入 5 元作原价
	rule := &model.FareRule{
		Enabled:  true,
		BaseFen:  200,
		CardRate: 0.9,
	}
	r := CalcFare(rule, time.Now(), 0, 500)
	if r.AmountFen != 450 {
		t.Fatalf("expect 450 fen from override base 5*0.9, got %d %s", r.AmountFen, r.Reason)
	}
}

func TestCycleWindow_FromDay(t *testing.T) {
	rule := &model.FareRule{CycleType: "from_day", CycleStartDay: 15}
	at := time.Date(2026, 9, 18, 12, 0, 0, 0, time.Local)
	start, end := CycleWindow(rule, at)
	if start.Day() != 15 || start.Month() != time.September {
		t.Fatalf("start=%v", start)
	}
	if end.Day() != 15 || end.Month() != time.October {
		t.Fatalf("end=%v", end)
	}
	at2 := time.Date(2026, 9, 10, 0, 0, 0, 0, time.Local)
	start2, _ := CycleWindow(rule, at2)
	if start2.Month() != time.August || start2.Day() != 15 {
		t.Fatalf("start2=%v", start2)
	}
}

func TestCityMatches(t *testing.T) {
	if !CityMatches("", "福州") {
		t.Fatal("empty rule city should match")
	}
	if !CityMatches("福州", "") {
		t.Fatal("empty user city should match")
	}
	if !CityMatches("福州", "福州市") {
		t.Fatal("should match with 市 suffix")
	}
	if CityMatches("上海", "福州") {
		t.Fatal("should not match different cities")
	}
}

func TestActiveCycleSeed(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local)
	rule := &model.FareRule{
		CycleSeedFen:   5000,
		CycleCountSeed: 3,
		CycleSeedKey:   "2026-09-01",
	}
	fen, cnt := activeCycleSeed(rule, start)
	if fen != 5000 || cnt != 3 {
		t.Fatalf("expect 5000/3, got %d/%d", fen, cnt)
	}
	// 换月后补录失效
	fen, cnt = activeCycleSeed(rule, start.AddDate(0, 1, 0))
	if fen != 0 || cnt != 0 {
		t.Fatalf("expect 0 after cycle change, got %d/%d", fen, cnt)
	}
	applyCycleSeed(rule, ptrInt64(0), ptrInt(0), start)
	if rule.CycleSeedFen != 0 || rule.CycleCountSeed != 0 || rule.CycleSeedKey != "" {
		t.Fatalf("cleared seed expect empty, got fen=%d count=%d key=%q", rule.CycleSeedFen, rule.CycleCountSeed, rule.CycleSeedKey)
	}
	applyCycleSeed(rule, ptrInt64(1200), ptrInt(5), start)
	if rule.CycleSeedFen != 1200 || rule.CycleCountSeed != 5 || rule.CycleSeedKey != "2026-09-01" {
		t.Fatalf("seed apply failed: fen=%d count=%d key=%q", rule.CycleSeedFen, rule.CycleCountSeed, rule.CycleSeedKey)
	}
}

func TestCalcFare_TimeWindow(t *testing.T) {
	rule := &model.FareRule{
		Enabled:  true,
		BaseFen:  1000,
		CardRate: 1,
		TimeWindows: []model.FareTimeWindow{
			{StartHM: "06:00", EndHM: "07:00", Weekdays: []int{1, 2, 3, 4, 5}, Rate: 0.7, Base: "full"},
		},
	}
	// 周一 06:30
	at := time.Date(2026, 9, 21, 6, 30, 0, 0, time.Local)
	r := CalcFare(rule, at, 0, 1000)
	if r.AmountFen != 700 {
		t.Fatalf("expect 700, got %d %s", r.AmountFen, r.Reason)
	}
	// 周一 08:00 — 不在窗口
	at2 := time.Date(2026, 9, 21, 8, 0, 0, 0, time.Local)
	r2 := CalcFare(rule, at2, 0, 1000)
	if r2.AmountFen != 1000 {
		t.Fatalf("expect full 1000, got %d", r2.AmountFen)
	}
}

func TestCalcFare_CountFree(t *testing.T) {
	rule := &model.FareRule{
		Enabled:  true,
		BaseFen:  200,
		CardRate: 0.9,
		CountTiers: []model.FareCountTier{
			{MinCount: 39, MaxCount: 40, Free: true},
		},
	}
	r := CalcFareFull(rule, time.Now(), CalcContext{BaseFen: 200, RideCount: 39})
	if !r.Applied.Free || r.AmountFen != 0 {
		t.Fatalf("40th should be free, got %+v", r)
	}
	r2 := CalcFareFull(rule, time.Now(), CalcContext{BaseFen: 200, RideCount: 38})
	if r2.Applied.Free {
		t.Fatalf("39th should not be free")
	}
}

func TestCalcFare_AmountOff(t *testing.T) {
	rule := &model.FareRule{
		Enabled:      true,
		BaseFen:      500,
		CardRate:     1,
		AmountOffFen: 200,
	}
	r := CalcFare(rule, time.Now(), 0, 500)
	if r.AmountFen != 300 {
		t.Fatalf("expect 300 after 2yuan off, got %d", r.AmountFen)
	}
}

func TestCalcFare_HolidayFree(t *testing.T) {
	rule := &model.FareRule{Enabled: true, BaseFen: 200, FreeOnHoliday: true}
	r := CalcFareFull(rule, time.Now(), CalcContext{BaseFen: 200, IsHolidayRest: true, HolidayName: "国庆节"})
	if !r.Applied.Free || r.AmountFen != 0 {
		t.Fatalf("holiday should be free")
	}
}

func ptrInt64(v int64) *int64 { return &v }
func ptrInt(v int) *int       { return &v }
