package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybookkeeping/internal/model"
)

type FareRuleService struct {
	db       *gorm.DB
	holidays *HolidayService
}

func NewFareRuleService(db *gorm.DB, holidays *HolidayService) *FareRuleService {
	return &FareRuleService{db: db, holidays: holidays}
}

type FareRuleInput struct {
	Name           string                  `json:"name"`
	City           string                  `json:"city"`
	Note           string                  `json:"note"`
	ValidFrom      string                  `json:"validFrom"`
	ValidTo        string                  `json:"validTo"`
	BaseFen        int64                   `json:"baseFen"`
	CardRate       float64                 `json:"cardRate"`
	AmountOffFen   int64                   `json:"amountOffFen"`
	StackMode      string                  `json:"stackMode"`
	FreeOnHoliday  *bool                   `json:"freeOnHoliday"`
	FreePeriods    []model.FareFreePeriod  `json:"freePeriods"`
	Tiers          []model.FareTier        `json:"tiers"`
	TimeWindows    []model.FareTimeWindow  `json:"timeWindows"`
	CountTiers     []model.FareCountTier   `json:"countTiers"`
	CycleType      string                  `json:"cycleType"`
	CycleStartDay  int                     `json:"cycleStartDay"`
	CycleSeedFen   *int64                  `json:"cycleSeedFen"` // nil=不改；0=清除补录
	CycleCountSeed *int                    `json:"cycleCountSeed"`
	Enabled        *bool                   `json:"enabled"`
	Icon           string                  `json:"icon"`
	Sort           int                     `json:"sort"`
}

type FarePreviewResult struct {
	AmountFen      int64  `json:"amountFen"`
	Reason         string `json:"reason"`
	MonthSpentFen  int64  `json:"monthSpentFen"` // 本周期生效累计 = 流水 + 补录
	TxSpentFen     int64  `json:"txSpentFen"`
	SeedFen        int64  `json:"seedFen"`
	RideCount      int    `json:"rideCount"` // 本周期已乘次数（流水+补录）
	TxRideCount    int    `json:"txRideCount"`
	SeedRideCount  int    `json:"seedRideCount"`
	ThisRideNo     int    `json:"thisRideNo"` // 本趟序号 = rideCount+1
	CycleStart     string `json:"cycleStart,omitempty"`
	CycleEnd       string `json:"cycleEnd,omitempty"`
	HolidayName    string `json:"holidayName,omitempty"`
	Applied        struct {
		Free       bool    `json:"free"`
		CardRate   float64 `json:"cardRate"`
		TierRate   float64 `json:"tierRate"`
		TimeRate   float64 `json:"timeRate"`
		CountRate  float64 `json:"countRate"`
		AmountOff  int64   `json:"amountOffFen"`
		StackMode  string  `json:"stackMode"`
	} `json:"applied"`
}

type CalcContext struct {
	SpentFen       int64
	RideCount      int // 本周期已乘次数（不含本趟）
	BaseFen        int64
	IsHolidayRest  bool
	HolidayName    string
}

func encodeFareJSON[T any](v T) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func hydrateFareRule(r *model.FareRule) {
	if r == nil {
		return
	}
	r.FreePeriods = nil
	r.Tiers = nil
	r.TimeWindows = nil
	r.CountTiers = nil
	if strings.TrimSpace(r.FreePeriodsJSON) != "" {
		_ = json.Unmarshal([]byte(r.FreePeriodsJSON), &r.FreePeriods)
	}
	if strings.TrimSpace(r.TiersJSON) != "" {
		_ = json.Unmarshal([]byte(r.TiersJSON), &r.Tiers)
	}
	if strings.TrimSpace(r.TimeWindowsJSON) != "" {
		_ = json.Unmarshal([]byte(r.TimeWindowsJSON), &r.TimeWindows)
	}
	if strings.TrimSpace(r.CountTiersJSON) != "" {
		_ = json.Unmarshal([]byte(r.CountTiersJSON), &r.CountTiers)
	}
	if r.FreePeriods == nil {
		r.FreePeriods = []model.FareFreePeriod{}
	}
	if r.Tiers == nil {
		r.Tiers = []model.FareTier{}
	}
	if r.TimeWindows == nil {
		r.TimeWindows = []model.FareTimeWindow{}
	}
	if r.CountTiers == nil {
		r.CountTiers = []model.FareCountTier{}
	}
	normalizeTiersInPlace(r.Tiers)
	normalizeTimeWindowsInPlace(r.TimeWindows)
	normalizeCountTiersInPlace(r.CountTiers)
	sort.Slice(r.Tiers, func(i, j int) bool {
		return r.Tiers[i].MinFen < r.Tiers[j].MinFen
	})
	sort.Slice(r.CountTiers, func(i, j int) bool {
		return r.CountTiers[i].MinCount < r.CountTiers[j].MinCount
	})
	if r.StackMode == "" {
		r.StackMode = "prefer"
	}
}

func normalizeTiersInPlace(tiers []model.FareTier) {
	for i := range tiers {
		t := &tiers[i]
		if t.MinFen == 0 && t.ThresholdFen > 0 {
			t.MinFen = t.ThresholdFen
		}
		t.ThresholdFen = 0
		base := strings.ToLower(strings.TrimSpace(t.Base))
		if base != "full" && base != "card" {
			base = "card"
		}
		t.Base = base
		if t.Rate <= 0 {
			t.Rate = 1
		}
	}
}

func normalizeTimeWindowsInPlace(wins []model.FareTimeWindow) {
	for i := range wins {
		w := &wins[i]
		base := strings.ToLower(strings.TrimSpace(w.Base))
		if base != "full" && base != "card" {
			base = "card"
		}
		w.Base = base
		if w.Rate <= 0 {
			w.Rate = 1
		}
		w.StartHM = strings.TrimSpace(w.StartHM)
		w.EndHM = strings.TrimSpace(w.EndHM)
	}
}

func normalizeCountTiersInPlace(tiers []model.FareCountTier) {
	for i := range tiers {
		t := &tiers[i]
		base := strings.ToLower(strings.TrimSpace(t.Base))
		if base != "full" && base != "card" {
			base = "card"
		}
		t.Base = base
		if t.Rate <= 0 {
			t.Rate = 1
		}
		if t.MinCount < 0 {
			t.MinCount = 0
		}
		if t.MaxCount < 0 {
			t.MaxCount = 0
		}
		if t.AmountOffFen < 0 {
			t.AmountOffFen = 0
		}
	}
}

func normalizeFareInput(in *FareRuleInput) {
	if in.CardRate <= 0 {
		in.CardRate = 1
	}
	if in.BaseFen < 0 {
		in.BaseFen = 0
	}
	if in.AmountOffFen < 0 {
		in.AmountOffFen = 0
	}
	sm := strings.ToLower(strings.TrimSpace(in.StackMode))
	if sm != "lowest" && sm != "stack" {
		sm = "prefer"
	}
	in.StackMode = sm
	ct := strings.ToLower(strings.TrimSpace(in.CycleType))
	if ct != "from_day" {
		ct = "calendar_month"
	}
	in.CycleType = ct
	if in.CycleStartDay < 1 {
		in.CycleStartDay = 1
	}
	if in.CycleStartDay > 28 {
		in.CycleStartDay = 28
	}
	if in.FreePeriods == nil {
		in.FreePeriods = []model.FareFreePeriod{}
	}
	for i := range in.FreePeriods {
		rec := strings.ToLower(strings.TrimSpace(in.FreePeriods[i].Recur))
		if rec != "yearly" && rec != "weekly" {
			rec = "none"
		}
		in.FreePeriods[i].Recur = rec
	}
	if in.Tiers == nil {
		in.Tiers = []model.FareTier{}
	}
	if in.TimeWindows == nil {
		in.TimeWindows = []model.FareTimeWindow{}
	}
	if in.CountTiers == nil {
		in.CountTiers = []model.FareCountTier{}
	}
	normalizeTiersInPlace(in.Tiers)
	normalizeTimeWindowsInPlace(in.TimeWindows)
	normalizeCountTiersInPlace(in.CountTiers)
	sort.Slice(in.Tiers, func(i, j int) bool {
		return in.Tiers[i].MinFen < in.Tiers[j].MinFen
	})
	sort.Slice(in.CountTiers, func(i, j int) bool {
		return in.CountTiers[i].MinCount < in.CountTiers[j].MinCount
	})
	in.City = strings.TrimSpace(strings.TrimSuffix(in.City, "市"))
	in.Note = strings.TrimSpace(in.Note)
	in.ValidFrom = strings.TrimSpace(in.ValidFrom)
	in.ValidTo = strings.TrimSpace(in.ValidTo)
}

func (s *FareRuleService) List(userID uint64) ([]model.FareRule, error) {
	var list []model.FareRule
	err := s.db.Where("user_id = ?", userID).Order("sort asc, id asc").Find(&list).Error
	for i := range list {
		hydrateFareRule(&list[i])
	}
	return list, err
}

func (s *FareRuleService) Get(userID, id uint64) (*model.FareRule, error) {
	var r model.FareRule
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&r).Error; err != nil {
		return nil, err
	}
	hydrateFareRule(&r)
	return &r, nil
}

func (s *FareRuleService) Create(userID uint64, in FareRuleInput) (*model.FareRule, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.New("规则名称不能为空")
	}
	normalizeFareInput(&in)
	en := true
	if in.Enabled != nil {
		en = *in.Enabled
	}
	freeHol := false
	if in.FreeOnHoliday != nil {
		freeHol = *in.FreeOnHoliday
	}
	r := &model.FareRule{
		UserID:          userID,
		Name:            strings.TrimSpace(in.Name),
		City:            in.City,
		Note:            in.Note,
		ValidFrom:       in.ValidFrom,
		ValidTo:         in.ValidTo,
		BaseFen:         in.BaseFen,
		CardRate:        in.CardRate,
		AmountOffFen:    in.AmountOffFen,
		StackMode:       in.StackMode,
		FreeOnHoliday:   freeHol,
		FreePeriodsJSON: encodeFareJSON(in.FreePeriods),
		TiersJSON:       encodeFareJSON(in.Tiers),
		TimeWindowsJSON: encodeFareJSON(in.TimeWindows),
		CountTiersJSON:  encodeFareJSON(in.CountTiers),
		CycleType:       in.CycleType,
		CycleStartDay:   in.CycleStartDay,
		Enabled:         en,
		Icon:            strings.TrimSpace(in.Icon),
		Sort:            in.Sort,
	}
	applyCycleSeed(r, in.CycleSeedFen, in.CycleCountSeed, time.Now())
	if err := s.db.Create(r).Error; err != nil {
		return nil, err
	}
	hydrateFareRule(r)
	return r, nil
}

func (s *FareRuleService) Update(userID, id uint64, in FareRuleInput) (*model.FareRule, error) {
	var r model.FareRule
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&r).Error; err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.New("规则名称不能为空")
	}
	normalizeFareInput(&in)
	r.Name = strings.TrimSpace(in.Name)
	r.City = in.City
	r.Note = in.Note
	r.ValidFrom = in.ValidFrom
	r.ValidTo = in.ValidTo
	r.BaseFen = in.BaseFen
	r.CardRate = in.CardRate
	r.AmountOffFen = in.AmountOffFen
	r.StackMode = in.StackMode
	r.FreePeriodsJSON = encodeFareJSON(in.FreePeriods)
	r.TiersJSON = encodeFareJSON(in.Tiers)
	r.TimeWindowsJSON = encodeFareJSON(in.TimeWindows)
	r.CountTiersJSON = encodeFareJSON(in.CountTiers)
	r.CycleType = in.CycleType
	r.CycleStartDay = in.CycleStartDay
	r.Sort = in.Sort
	r.Icon = strings.TrimSpace(in.Icon)
	if r.Icon == "" {
		r.Icon = "mdi-ticket-percent"
	}
	if in.Enabled != nil {
		r.Enabled = *in.Enabled
	}
	if in.FreeOnHoliday != nil {
		r.FreeOnHoliday = *in.FreeOnHoliday
	}
	applyCycleSeed(&r, in.CycleSeedFen, in.CycleCountSeed, time.Now())
	if err := s.db.Model(&r).Select(
		"Name", "City", "Note", "ValidFrom", "ValidTo",
		"BaseFen", "CardRate", "AmountOffFen", "StackMode", "FreeOnHoliday",
		"FreePeriodsJSON", "TiersJSON", "TimeWindowsJSON", "CountTiersJSON",
		"CycleType", "CycleStartDay", "CycleSeedFen", "CycleSeedKey", "CycleCountSeed",
		"Enabled", "Icon", "Sort",
	).Updates(&r).Error; err != nil {
		return nil, err
	}
	hydrateFareRule(&r)
	return &r, nil
}

// applyCycleSeed：金额/次数补录；nil 不改；<=0 清除；>0 写入并绑定当前周期起点
func applyCycleSeed(r *model.FareRule, seedFen *int64, seedCount *int, at time.Time) {
	if r == nil {
		return
	}
	changed := false
	if seedFen != nil {
		changed = true
		if *seedFen <= 0 {
			r.CycleSeedFen = 0
		} else {
			r.CycleSeedFen = *seedFen
		}
	}
	if seedCount != nil {
		changed = true
		if *seedCount <= 0 {
			r.CycleCountSeed = 0
		} else {
			r.CycleCountSeed = *seedCount
		}
	}
	if !changed {
		return
	}
	if r.CycleSeedFen <= 0 && r.CycleCountSeed <= 0 {
		r.CycleSeedKey = ""
		return
	}
	start, _ := CycleWindow(r, at)
	r.CycleSeedKey = start.Format("2006-01-02")
}

func activeCycleSeed(rule *model.FareRule, cycleStart time.Time) (seedFen int64, seedCount int) {
	if rule == nil || rule.CycleSeedKey == "" {
		return 0, 0
	}
	if rule.CycleSeedKey != cycleStart.Format("2006-01-02") {
		return 0, 0
	}
	return rule.CycleSeedFen, rule.CycleCountSeed
}

func (s *FareRuleService) Delete(userID, id uint64) error {
	_ = s.db.Model(&model.Template{}).
		Where("user_id = ? AND fare_rule_id = ?", userID, id).
		Update("fare_rule_id", nil).Error
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.FareRule{}).Error
}

// CycleWindow 返回 [start, end)
func CycleWindow(rule *model.FareRule, at time.Time) (time.Time, time.Time) {
	loc := at.Location()
	if rule != nil && rule.CycleType == "from_day" {
		sd := rule.CycleStartDay
		if sd < 1 {
			sd = 1
		}
		if sd > 28 {
			sd = 28
		}
		var start time.Time
		if at.Day() >= sd {
			start = time.Date(at.Year(), at.Month(), sd, 0, 0, 0, 0, loc)
		} else {
			prev := at.AddDate(0, -1, 0)
			start = time.Date(prev.Year(), prev.Month(), sd, 0, 0, 0, 0, loc)
		}
		return start, start.AddDate(0, 1, 0)
	}
	start := time.Date(at.Year(), at.Month(), 1, 0, 0, 0, 0, loc)
	return start, start.AddDate(0, 1, 0)
}

func (s *FareRuleService) SpentInCycle(userID, ruleID uint64, at time.Time) (int64, time.Time, time.Time, error) {
	total, _, _, _, _, _, start, end, err := s.SpentBreakdown(userID, ruleID, at)
	return total, start, end, err
}

// SpentBreakdown 返回 生效合计、流水合计、补录金额、乘次合计、流水次数、补录次数、周期窗
func (s *FareRuleService) SpentBreakdown(userID, ruleID uint64, at time.Time) (
	total, txSum, seedFen int64, rideTotal, txRides, seedRides int, start, end time.Time, err error,
) {
	rule, err := s.Get(userID, ruleID)
	if err != nil {
		return
	}
	start, end = CycleWindow(rule, at)
	err = s.db.Model(&model.Transaction{}).
		Where("user_id = ? AND type = ? AND fare_rule_id = ? AND happened_at >= ? AND happened_at < ?",
			userID, "expense", ruleID, start, end).
		Select("COALESCE(SUM(amount), 0)").Scan(&txSum).Error
	if err != nil {
		return
	}
	var cnt int64
	err = s.db.Model(&model.Transaction{}).
		Where("user_id = ? AND type = ? AND fare_rule_id = ? AND happened_at >= ? AND happened_at < ?",
			userID, "expense", ruleID, start, end).
		Count(&cnt).Error
	if err != nil {
		return
	}
	txRides = int(cnt)
	seedFen, seedRides = activeCycleSeed(rule, start)
	total = txSum + seedFen
	rideTotal = txRides + seedRides
	return
}

// Preview 预览应付；baseOverride>0 时覆盖规则内原价。
func (s *FareRuleService) Preview(userID, id uint64, at time.Time, spentOverride, baseOverride *int64, rideOverride *int) (*FarePreviewResult, error) {
	rule, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	var spent, txSum, seed int64
	var rideTotal, txRides, seedRides int
	var start, end time.Time
	if spentOverride != nil || rideOverride != nil {
		start, end = CycleWindow(rule, at)
		if spentOverride != nil {
			spent = *spentOverride
			txSum = spent
			seed = 0
		} else {
			spent, txSum, seed, rideTotal, txRides, seedRides, start, end, err = s.SpentBreakdown(userID, id, at)
			if err != nil {
				return nil, err
			}
		}
		if rideOverride != nil {
			rideTotal = *rideOverride
			txRides = rideTotal
			seedRides = 0
		} else if spentOverride != nil {
			// 只覆盖金额时仍取真实次数
			_, _, _, rideTotal, txRides, seedRides, _, _, err = s.SpentBreakdown(userID, id, at)
			if err != nil {
				return nil, err
			}
		}
	} else {
		spent, txSum, seed, rideTotal, txRides, seedRides, start, end, err = s.SpentBreakdown(userID, id, at)
		if err != nil {
			return nil, err
		}
	}
	baseFen := rule.BaseFen
	if baseOverride != nil && *baseOverride > 0 {
		baseFen = *baseOverride
	}
	ctx := CalcContext{
		SpentFen:  spent,
		RideCount: rideTotal,
		BaseFen:   baseFen,
	}
	if rule.FreeOnHoliday && s.holidays != nil {
		rest, name, hErr := s.holidays.IsRestDay(context.Background(), at)
		if hErr == nil {
			ctx.IsHolidayRest = rest
			ctx.HolidayName = name
		}
	}
	res := CalcFareFull(rule, at, ctx)
	res.TxSpentFen = txSum
	res.SeedFen = seed
	res.RideCount = rideTotal
	res.TxRideCount = txRides
	res.SeedRideCount = seedRides
	res.ThisRideNo = rideTotal + 1
	res.CycleStart = start.Format("2006-01-02")
	res.CycleEnd = end.Format("2006-01-02")
	res.HolidayName = ctx.HolidayName
	return res, nil
}

// MigrateEmbeddedTemplateFareRules 将模板内嵌 FareRuleJSON 迁到全局表
func (s *FareRuleService) MigrateEmbeddedTemplateFareRules() error {
	var templates []model.Template
	if err := s.db.Where("fare_rule_json IS NOT NULL AND fare_rule_json <> '' AND fare_rule_id IS NULL").
		Find(&templates).Error; err != nil {
		return err
	}
	for _, t := range templates {
		var legacy struct {
			Enabled     bool                   `json:"enabled"`
			BaseFen     int64                  `json:"baseFen"`
			CardRate    float64                `json:"cardRate"`
			FreePeriods []model.FareFreePeriod `json:"freePeriods"`
			Tiers       []model.FareTier       `json:"tiers"`
		}
		if err := json.Unmarshal([]byte(t.FareRuleJSON), &legacy); err != nil || !legacy.Enabled {
			_ = s.db.Model(&t).Updates(map[string]any{"fare_rule_json": ""}).Error
			continue
		}
		name := t.Name + "·计费"
		if len([]rune(name)) > 64 {
			name = string([]rune(name)[:64])
		}
		en := true
		created, err := s.Create(t.UserID, FareRuleInput{
			Name:        name,
			BaseFen:     legacy.BaseFen,
			CardRate:    legacy.CardRate,
			FreePeriods: legacy.FreePeriods,
			Tiers:       legacy.Tiers,
			CycleType:   "calendar_month",
			Enabled:     &en,
		})
		if err != nil {
			continue
		}
		_ = s.db.Model(&model.Template{}).Where("id = ?", t.ID).Updates(map[string]any{
			"fare_rule_id":   created.ID,
			"fare_rule_json": "",
		}).Error
	}
	return nil
}

// CalcFare 兼容旧调用：仅金额累计 + 原价
func CalcFare(rule *model.FareRule, at time.Time, cycleSpentFen, baseFen int64) *FarePreviewResult {
	return CalcFareFull(rule, at, CalcContext{SpentFen: cycleSpentFen, BaseFen: baseFen})
}

type fareCandidate struct {
	amount float64
	label  string
	free   bool
	kind   string // card|tier|time|count
	rate   float64
}

// CalcFareFull 完整计价
func CalcFareFull(rule *model.FareRule, at time.Time, ctx CalcContext) *FarePreviewResult {
	res := &FarePreviewResult{MonthSpentFen: ctx.SpentFen, RideCount: ctx.RideCount, ThisRideNo: ctx.RideCount + 1}
	res.Applied.CardRate = 1
	res.Applied.TierRate = 1
	res.Applied.TimeRate = 1
	res.Applied.CountRate = 1
	res.Applied.StackMode = "prefer"
	if rule == nil || !rule.Enabled {
		res.Reason = "未启用计费规则"
		return res
	}
	if !ruleInValidRange(rule, at) {
		res.Reason = "规则不在生效期内"
		return res
	}
	res.Applied.StackMode = rule.StackMode
	if rule.StackMode == "" {
		res.Applied.StackMode = "prefer"
	}

	if isFreeDay(rule.FreePeriods, at) {
		res.AmountFen = 0
		res.Applied.Free = true
		res.Reason = "免费时段"
		return res
	}
	if rule.FreeOnHoliday && ctx.IsHolidayRest {
		res.AmountFen = 0
		res.Applied.Free = true
		name := ctx.HolidayName
		if name == "" {
			name = "法定节假日"
		}
		res.Reason = "法定节假日免费 · " + name
		res.HolidayName = name
		return res
	}

	baseFen := ctx.BaseFen
	if baseFen < 0 {
		baseFen = 0
	}
	card := rule.CardRate
	if card <= 0 {
		card = 1
	}
	res.Applied.CardRate = card

	cardAmt := float64(baseFen) * card
	cands := []fareCandidate{{
		amount: cardAmt,
		label:  fmt.Sprintf("原价¥%.2f", float64(baseFen)/100),
		kind:   "card",
		rate:   card,
	}}
	if card != 1 {
		cands[0].label += fmt.Sprintf(" · 票卡%.0f折", card*10)
	}

	if tier, ok := matchTier(rule.Tiers, ctx.SpentFen); ok {
		rate := tier.Rate
		if rate <= 0 {
			rate = 1
		}
		res.Applied.TierRate = rate
		var amt float64
		var lab string
		if tier.Base == "full" {
			amt = float64(baseFen) * rate
			lab = fmt.Sprintf("累计超¥%.2f全价%.0f折", float64(tier.MinFen)/100, rate*10)
		} else {
			amt = float64(baseFen) * card * rate
			if rate == 1 {
				lab = fmt.Sprintf("累计超¥%.2f恢复票卡%.0f折", float64(tier.MinFen)/100, card*10)
			} else {
				lab = fmt.Sprintf("累计超¥%.2f票卡后再%.0f折", float64(tier.MinFen)/100, rate*10)
			}
		}
		cands = append(cands, fareCandidate{amount: amt, label: lab, kind: "tier", rate: rate})
	}

	thisRide := ctx.RideCount + 1
	if ct, ok := matchCountTier(rule.CountTiers, thisRide); ok {
		if ct.Free {
			res.AmountFen = 0
			res.Applied.Free = true
			res.Applied.CountRate = 0
			res.Reason = fmt.Sprintf("第%d次免费", thisRide)
			return res
		}
		rate := ct.Rate
		if rate <= 0 {
			rate = 1
		}
		res.Applied.CountRate = rate
		var amt float64
		var lab string
		if ct.Base == "full" {
			amt = float64(baseFen) * rate
			lab = fmt.Sprintf("第%d次全价%.0f折", thisRide, rate*10)
		} else {
			amt = float64(baseFen) * card * rate
			lab = fmt.Sprintf("第%d次票卡后再%.0f折", thisRide, rate*10)
		}
		if ct.AmountOffFen > 0 {
			amt -= float64(ct.AmountOffFen)
			if amt < 0 {
				amt = 0
			}
			lab += fmt.Sprintf(" · 立减¥%.2f", float64(ct.AmountOffFen)/100)
		}
		cands = append(cands, fareCandidate{amount: amt, label: lab, kind: "count", rate: rate})
	}

	if tw, ok := matchTimeWindow(rule.TimeWindows, at); ok {
		if tw.Free {
			res.AmountFen = 0
			res.Applied.Free = true
			res.Applied.TimeRate = 0
			name := tw.Name
			if name == "" {
				name = "时段免费"
			}
			res.Reason = name
			return res
		}
		rate := tw.Rate
		if rate <= 0 {
			rate = 1
		}
		res.Applied.TimeRate = rate
		var amt float64
		baseLabel := "票卡"
		if tw.Base == "full" {
			amt = float64(baseFen) * rate
			baseLabel = "全价"
		} else {
			amt = float64(baseFen) * card * rate
		}
		name := tw.Name
		if name == "" {
			name = fmt.Sprintf("%s–%s", tw.StartHM, tw.EndHM)
		}
		lab := fmt.Sprintf("%s%s%.0f折", name, baseLabel, rate*10)
		cands = append(cands, fareCandidate{amount: amt, label: lab, kind: "time", rate: rate})
	}

	mode := res.Applied.StackMode
	var amount float64
	var parts []string

	switch mode {
	case "lowest":
		best := cands[0]
		for _, c := range cands[1:] {
			if c.amount < best.amount {
				best = c
			}
		}
		amount = best.amount
		parts = []string{fmt.Sprintf("原价¥%.2f", float64(baseFen)/100), "就低 · " + best.label}
	case "stack":
		// 金额阶梯（或票卡）→ 再乘时段折 → 次数档金额若更优则改用（免费已提前返回）
		baseCand := cands[0]
		for _, c := range cands {
			if c.kind == "tier" {
				baseCand = c
				break
			}
		}
		amount = baseCand.amount
		parts = []string{fmt.Sprintf("原价¥%.2f", float64(baseFen)/100), baseCand.label}
		for _, c := range cands {
			if c.kind == "time" && c.rate > 0 && c.rate < 1 {
				amount *= c.rate
				parts = append(parts, fmt.Sprintf("再叠时段%.0f折", c.rate*10))
				break
			}
		}
		for _, c := range cands {
			if c.kind == "count" && c.amount < amount {
				amount = c.amount
				parts = append(parts, "次数档更优 · "+c.label)
				break
			}
		}
	default: // prefer：免费已提前返回；次数 > 时段 > 金额阶梯 > 票卡
		pick := cands[0]
		priority := map[string]int{"count": 3, "time": 2, "tier": 1, "card": 0}
		for _, c := range cands[1:] {
			if priority[c.kind] > priority[pick.kind] {
				pick = c
			}
		}
		amount = pick.amount
		parts = []string{fmt.Sprintf("原价¥%.2f", float64(baseFen)/100)}
		if pick.kind != "card" || card != 1 {
			parts = append(parts, pick.label)
		} else if pick.kind == "card" && strings.Contains(pick.label, "票卡") {
			parts = []string{pick.label}
		}
	}

	off := rule.AmountOffFen
	if off > 0 {
		amount -= float64(off)
		if amount < 0 {
			amount = 0
		}
		res.Applied.AmountOff = off
		parts = append(parts, fmt.Sprintf("立减¥%.2f", float64(off)/100))
	}

	res.AmountFen = int64(math.Round(amount))
	if res.AmountFen < 0 {
		res.AmountFen = 0
	}
	res.Reason = strings.Join(parts, " · ")
	return res
}

func ruleInValidRange(rule *model.FareRule, at time.Time) bool {
	day := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location())
	if rule.ValidFrom != "" {
		if t, ok := parseYMD(rule.ValidFrom); ok && day.Before(t) {
			return false
		}
	}
	if rule.ValidTo != "" {
		if t, ok := parseYMD(rule.ValidTo); ok && day.After(t) {
			return false
		}
	}
	return true
}

func matchTier(tiers []model.FareTier, spent int64) (model.FareTier, bool) {
	var best model.FareTier
	found := false
	for _, t := range tiers {
		if spent <= t.MinFen {
			continue
		}
		if t.MaxFen > 0 && spent > t.MaxFen {
			continue
		}
		if !found || t.MinFen > best.MinFen {
			best = t
			found = true
		}
	}
	return best, found
}

func matchCountTier(tiers []model.FareCountTier, thisRide int) (model.FareCountTier, bool) {
	var best model.FareCountTier
	found := false
	for _, t := range tiers {
		if thisRide <= t.MinCount {
			continue
		}
		if t.MaxCount > 0 && thisRide > t.MaxCount {
			continue
		}
		if !found || t.MinCount > best.MinCount {
			best = t
			found = true
		}
	}
	return best, found
}

func matchTimeWindow(wins []model.FareTimeWindow, at time.Time) (model.FareTimeWindow, bool) {
	wd := int(at.Weekday())
	mins := at.Hour()*60 + at.Minute()
	for _, w := range wins {
		if len(w.Weekdays) > 0 && !weekdayMatch(w.Weekdays, wd) {
			continue
		}
		startM, ok1 := parseHM(w.StartHM)
		endM, ok2 := parseHM(w.EndHM)
		if !ok1 || !ok2 {
			continue
		}
		if inHMRange(mins, startM, endM) {
			return w, true
		}
	}
	return model.FareTimeWindow{}, false
}

func parseHM(s string) (int, bool) {
	s = strings.TrimSpace(s)
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, false
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

// inHMRange [start, end)；若 end<=start 视为跨午夜
func inHMRange(mins, start, end int) bool {
	if start == end {
		return true // 全天
	}
	if end > start {
		return mins >= start && mins < end
	}
	return mins >= start || mins < end
}

func parseYMD(s string) (time.Time, bool) {
	t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(s), time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func isFreeDay(periods []model.FareFreePeriod, at time.Time) bool {
	day := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location())
	for _, p := range periods {
		switch p.Recur {
		case "weekly":
			if !weekdayMatch(p.Weekdays, int(day.Weekday())) {
				continue
			}
			if p.StartDate != "" && p.EndDate != "" {
				start, ok1 := parseYMD(p.StartDate)
				end, ok2 := parseYMD(p.EndDate)
				if ok1 && ok2 {
					if day.Before(start) || day.After(end) {
						continue
					}
				}
			}
			return true
		case "yearly":
			if inYearlyRange(day, p.StartDate, p.EndDate) {
				return true
			}
		default:
			start, ok1 := parseYMD(p.StartDate)
			end, ok2 := parseYMD(p.EndDate)
			if !ok1 || !ok2 {
				continue
			}
			if !day.Before(start) && !day.After(end) {
				return true
			}
		}
	}
	return false
}

func weekdayMatch(days []int, wd int) bool {
	for _, d := range days {
		if d == wd {
			return true
		}
	}
	return false
}

func inYearlyRange(day time.Time, startS, endS string) bool {
	start, ok1 := parseYMD(startS)
	end, ok2 := parseYMD(endS)
	if !ok1 || !ok2 {
		return false
	}
	y := day.Year()
	s := time.Date(y, start.Month(), start.Day(), 0, 0, 0, 0, day.Location())
	e := time.Date(y, end.Month(), end.Day(), 0, 0, 0, 0, day.Location())
	if !e.Before(s) {
		return !day.Before(s) && !day.After(e)
	}
	return !day.Before(s) || !day.After(e)
}

// CityMatches 规则地市是否允许自动启用；ruleCity 空=不限；userCity 空=允许自动
func CityMatches(ruleCity, userCity string) bool {
	rc := strings.TrimSpace(strings.TrimSuffix(ruleCity, "市"))
	if rc == "" {
		return true
	}
	uc := strings.TrimSpace(strings.TrimSuffix(userCity, "市"))
	if uc == "" {
		return true
	}
	return strings.EqualFold(rc, uc) || strings.Contains(uc, rc) || strings.Contains(rc, uc)
}
