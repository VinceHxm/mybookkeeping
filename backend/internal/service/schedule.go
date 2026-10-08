package service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybookkeeping/internal/model"
)

type ScheduleService struct {
	db   *gorm.DB
	txSvc *TransactionService
}

func NewScheduleService(db *gorm.DB, txSvc *TransactionService) *ScheduleService {
	return &ScheduleService{db: db, txSvc: txSvc}
}

type ScheduleInput struct {
	Name        string
	Type        string
	Amount      int64
	AccountID   uint64
	ToAccountID *uint64
	CategoryID  *uint64
	Remark      string
	TagIDs      []uint64
	Frequency   string
	IntervalN   int
	AnchorDay   int
	Enabled     *bool
	NextRunAt   time.Time
	EndAt       *time.Time

	AmountPending   bool
	ForeignCurrency string
	ForeignAmount   int64
}

func hydrateSchedule(sch *model.Schedule) {
	sch.TagIDs = decodeIDs(sch.TagIDsJSON)
}

func (s *ScheduleService) List(userID uint64) ([]model.Schedule, error) {
	var list []model.Schedule
	err := s.db.Where("user_id = ?", userID).Order("next_run_at asc, id asc").Find(&list).Error
	for i := range list {
		hydrateSchedule(&list[i])
	}
	return list, err
}

func validateSchedule(in ScheduleInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("名称不能为空")
	}
	if in.Amount <= 0 {
		return errors.New("金额必须大于 0")
	}
	if in.AccountID == 0 {
		return errors.New("请选择账户")
	}
	switch in.Frequency {
	case "daily", "weekly", "monthly", "yearly", "every_n_days":
	default:
		return errors.New("频率无效")
	}
	if in.Frequency == "every_n_days" && in.IntervalN < 1 {
		return errors.New("间隔天数至少为 1")
	}
	if in.IntervalN > 3660 {
		return errors.New("间隔过大")
	}
	if in.AnchorDay < 0 || in.AnchorDay > 31 {
		return errors.New("执行日无效")
	}
	return nil
}

func anchorDayOf(t time.Time) int {
	return t.In(time.Local).Day()
}

func (s *ScheduleService) Create(userID uint64, in ScheduleInput) (*model.Schedule, error) {
	if err := validateSchedule(in); err != nil {
		return nil, err
	}
	if err := assertRefsOwned(s.db, userID, refIDs{Account: &in.AccountID, ToAccount: in.ToAccountID, Category: in.CategoryID}); err != nil {
		return nil, err
	}
	if in.IntervalN < 1 {
		in.IntervalN = 1
	}
	if in.Type == "" {
		in.Type = "expense"
	}
	cur, famt, err := normalizeForeign(in.Type, in.ForeignCurrency, in.ForeignAmount)
	if err != nil {
		return nil, err
	}
	pending := in.AmountPending && in.Type != "transfer"
	if in.NextRunAt.IsZero() {
		in.NextRunAt = time.Now()
	}
	if in.AnchorDay == 0 {
		in.AnchorDay = anchorDayOf(in.NextRunAt)
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	sch := &model.Schedule{
		UserID: userID, Name: strings.TrimSpace(in.Name), Type: in.Type,
		Amount: in.Amount, AccountID: in.AccountID, ToAccountID: in.ToAccountID,
		CategoryID: in.CategoryID, Remark: in.Remark, TagIDsJSON: encodeIDs(in.TagIDs),
		Frequency: in.Frequency, IntervalN: in.IntervalN, AnchorDay: in.AnchorDay, Enabled: enabled,
		NextRunAt: in.NextRunAt, EndAt: in.EndAt,
		AmountPending: pending, ForeignCurrency: cur, ForeignAmount: famt,
	}
	if err := s.db.Create(sch).Error; err != nil {
		return nil, err
	}
	hydrateSchedule(sch)
	return sch, nil
}

func (s *ScheduleService) Update(userID, id uint64, in ScheduleInput) (*model.Schedule, error) {
	var sch model.Schedule
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&sch).Error; err != nil {
		return nil, err
	}
	if in.Name != "" {
		sch.Name = strings.TrimSpace(in.Name)
	}
	if in.Type != "" {
		sch.Type = in.Type
	}
	if in.Amount > 0 {
		sch.Amount = in.Amount
	}
	if in.AccountID > 0 {
		sch.AccountID = in.AccountID
	}
	sch.ToAccountID = in.ToAccountID
	sch.CategoryID = in.CategoryID
	sch.Remark = in.Remark
	sch.TagIDsJSON = encodeIDs(in.TagIDs)
	if in.Frequency != "" {
		sch.Frequency = in.Frequency
	}
	if in.IntervalN >= 1 {
		sch.IntervalN = in.IntervalN
	}
	if in.AnchorDay < 0 || in.AnchorDay > 31 {
		return nil, errors.New("执行日无效")
	}
	if in.Enabled != nil {
		sch.Enabled = *in.Enabled
	}
	if !in.NextRunAt.IsZero() {
		if in.AnchorDay == 0 && !sch.NextRunAt.Equal(in.NextRunAt) {
			sch.AnchorDay = anchorDayOf(in.NextRunAt)
		}
		sch.NextRunAt = in.NextRunAt
	}
	if in.AnchorDay > 0 {
		sch.AnchorDay = in.AnchorDay
	}
	sch.EndAt = in.EndAt
	cur, famt, err := normalizeForeign(sch.Type, in.ForeignCurrency, in.ForeignAmount)
	if err != nil {
		return nil, err
	}
	sch.ForeignCurrency, sch.ForeignAmount = cur, famt
	sch.AmountPending = in.AmountPending && sch.Type != "transfer"
	if err := assertRefsOwned(s.db, userID, refIDs{Account: &sch.AccountID, ToAccount: sch.ToAccountID, Category: sch.CategoryID}); err != nil {
		return nil, err
	}
	if err := s.db.Save(&sch).Error; err != nil {
		return nil, err
	}
	hydrateSchedule(&sch)
	return &sch, nil
}

func (s *ScheduleService) Delete(userID, id uint64) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Schedule{}).Error
}

// addMonthsAnchored 按执行日推进 months 个月；执行日超过当月天数时取月末，下月再回到执行日
func addMonthsAnchored(from time.Time, months, anchorDay int) time.Time {
	lf := from.In(time.Local)
	if anchorDay < 1 {
		anchorDay = lf.Day()
	}
	first := time.Date(lf.Year(), lf.Month()+time.Month(months), 1, 0, 0, 0, 0, time.Local)
	last := first.AddDate(0, 1, -1).Day()
	day := anchorDay
	if day > last {
		day = last
	}
	return time.Date(first.Year(), first.Month(), day, lf.Hour(), lf.Minute(), lf.Second(), lf.Nanosecond(), time.Local)
}

func nextRun(from time.Time, freq string, n, anchorDay int) time.Time {
	if n < 1 {
		n = 1
	}
	switch freq {
	case "daily":
		return from.AddDate(0, 0, n)
	case "weekly":
		return from.AddDate(0, 0, 7*n)
	case "monthly":
		return addMonthsAnchored(from, n, anchorDay)
	case "yearly":
		return addMonthsAnchored(from, 12*n, anchorDay)
	case "every_n_days":
		return from.AddDate(0, 0, n)
	default:
		return from.AddDate(0, 0, 1)
	}
}

// RunDue 执行到期的周期记账（由后台定时调用）
func (s *ScheduleService) RunDue(now time.Time) (int, error) {
	var list []model.Schedule
	if err := s.db.Where("enabled = ? AND next_run_at <= ?", true, now).Find(&list).Error; err != nil {
		return 0, err
	}
	done := 0
	for _, sch := range list {
		hydrateSchedule(&sch)
		if sch.EndAt != nil && now.After(*sch.EndAt) {
			_ = s.db.Model(&sch).Update("enabled", false).Error
			continue
		}
		remark := strings.TrimSpace(sch.Remark)
		if remark == "" {
			remark = strings.TrimSpace(sch.Name)
		}
		in := TransactionInput{
			Type: sch.Type, Amount: sch.Amount, AccountID: sch.AccountID,
			ToAccountID: sch.ToAccountID, CategoryID: sch.CategoryID,
			Remark: remark, HappenedAt: sch.NextRunAt, TagIDs: sch.TagIDs,
			AmountPending: sch.AmountPending,
			ForeignCurrency: sch.ForeignCurrency, ForeignAmount: sch.ForeignAmount,
		}
		if _, err := s.txSvc.Create(sch.UserID, in); err != nil {
			continue
		}
		nxt := nextRun(sch.NextRunAt, sch.Frequency, sch.IntervalN, sch.AnchorDay)
		// 追赶到当前，避免长时间停机后一次补太多
		for nxt.Before(now) && (sch.EndAt == nil || !nxt.After(*sch.EndAt)) {
			nxt = nextRun(nxt, sch.Frequency, sch.IntervalN, sch.AnchorDay)
		}
		updates := map[string]any{"next_run_at": nxt}
		if sch.EndAt != nil && nxt.After(*sch.EndAt) {
			updates["enabled"] = false
		}
		_ = s.db.Model(&model.Schedule{}).Where("id = ?", sch.ID).Updates(updates).Error
		done++
	}
	return done, nil
}
