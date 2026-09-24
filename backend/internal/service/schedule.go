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
	Enabled     *bool
	NextRunAt   time.Time
	EndAt       *time.Time
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
	if in.IntervalN < 1 {
		in.IntervalN = 1
	}
	return nil
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
	if in.NextRunAt.IsZero() {
		in.NextRunAt = time.Now()
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	sch := &model.Schedule{
		UserID: userID, Name: strings.TrimSpace(in.Name), Type: in.Type,
		Amount: in.Amount, AccountID: in.AccountID, ToAccountID: in.ToAccountID,
		CategoryID: in.CategoryID, Remark: in.Remark, TagIDsJSON: encodeIDs(in.TagIDs),
		Frequency: in.Frequency, IntervalN: in.IntervalN, Enabled: enabled,
		NextRunAt: in.NextRunAt, EndAt: in.EndAt,
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
	if in.Enabled != nil {
		sch.Enabled = *in.Enabled
	}
	if !in.NextRunAt.IsZero() {
		sch.NextRunAt = in.NextRunAt
	}
	sch.EndAt = in.EndAt
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

func nextRun(from time.Time, freq string, n int) time.Time {
	if n < 1 {
		n = 1
	}
	switch freq {
	case "daily":
		return from.AddDate(0, 0, n)
	case "weekly":
		return from.AddDate(0, 0, 7*n)
	case "monthly":
		return from.AddDate(0, n, 0)
	case "yearly":
		return from.AddDate(n, 0, 0)
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
		in := TransactionInput{
			Type: sch.Type, Amount: sch.Amount, AccountID: sch.AccountID,
			ToAccountID: sch.ToAccountID, CategoryID: sch.CategoryID,
			Remark: sch.Remark, HappenedAt: sch.NextRunAt, TagIDs: sch.TagIDs,
		}
		if _, err := s.txSvc.Create(sch.UserID, in); err != nil {
			continue
		}
		nxt := nextRun(sch.NextRunAt, sch.Frequency, sch.IntervalN)
		// 追赶到当前，避免长时间停机后一次补太多
		for nxt.Before(now) && (sch.EndAt == nil || !nxt.After(*sch.EndAt)) {
			nxt = nextRun(nxt, sch.Frequency, sch.IntervalN)
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
