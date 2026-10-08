package service

import (
	"gorm.io/gorm"

	"mybookkeeping/internal/model"
)

// Recent 最近 n 条流水（不做总数统计）
func (s *TransactionService) Recent(userID uint64, n int) ([]model.Transaction, error) {
	if n <= 0 || n > 50 {
		n = 5
	}
	var list []model.Transaction
	err := preloadTxRefs(s.db.Where("user_id = ?", userID)).
		Order("happened_at desc, id desc").Limit(n).Find(&list).Error
	return list, err
}

// PendingSummary 金额待确认笔数；恰好 1 笔时附带该笔（首页直达编辑）
func (s *TransactionService) PendingSummary(userID uint64) (int64, *model.Transaction, error) {
	var n int64
	if err := s.db.Model(&model.Transaction{}).
		Where("user_id = ? AND amount_pending = ?", userID, true).Count(&n).Error; err != nil {
		return 0, nil, err
	}
	if n != 1 {
		return n, nil, nil
	}
	var t model.Transaction
	if err := s.db.Preload("Category", func(db *gorm.DB) *gorm.DB { return db.Select("id", "name") }).
		Select("id", "type", "remark", "category_id").
		Where("user_id = ? AND amount_pending = ?", userID, true).First(&t).Error; err != nil {
		return n, nil, err
	}
	return n, &t, nil
}

// CountExpiringVouchers 待核销且过期日不晚于 until（YYYY-MM-DD）的团购券数
func (s *TransactionService) CountExpiringVouchers(userID uint64, until string) (int64, error) {
	var n int64
	err := s.db.Model(&model.Transaction{}).
		Where("user_id = ? AND voucher_status = ? AND voucher_expire_on <> '' AND voucher_expire_on <= ?",
			userID, VoucherPending, until).
		Count(&n).Error
	return n, err
}
