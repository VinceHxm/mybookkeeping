package service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mybookkeeping/internal/model"
)

const (
	AccountTypeInvest = "invest"
	// TxTypeRevalue 市值校准：Amount 为有符号差额（正=浮盈，负=浮亏），只改理财账户余额，不计入收支统计
	TxTypeRevalue = "revalue"
)

type RevalueInput struct {
	MarketValue int64 // 当前市值（分）
	HappenedAt  time.Time
	Remark      string
}

// Revalue 按当前市值校准理财账户：与账户余额的差额记一笔 revalue 流水，并刷新对账时间。
// 市值与余额一致时只刷新对账时间，返回 nil 流水。
func (s *TransactionService) Revalue(userID, accountID uint64, in RevalueInput) (*model.Transaction, error) {
	if in.MarketValue < 0 {
		return nil, errors.New("市值不能为负")
	}
	happened := in.HappenedAt
	if happened.IsZero() {
		happened = time.Now()
	}
	var created uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var acc model.Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", accountID, userID).First(&acc).Error; err != nil {
			return errors.New("账户不存在")
		}
		if acc.Type != AccountTypeInvest {
			return errors.New("只有理财账户可以校准市值")
		}
		if err := tx.Model(&acc).Update("last_reconciled_at", happened).Error; err != nil {
			return err
		}
		delta := in.MarketValue - acc.Balance
		if delta == 0 {
			return nil
		}
		remark := strings.TrimSpace(in.Remark)
		if remark == "" {
			remark = "市值校准"
		}
		t := &model.Transaction{
			UserID:             userID,
			Type:               TxTypeRevalue,
			Amount:             delta,
			AccountID:          accountID,
			Remark:             remark,
			InstallmentPeriods: 1,
			HappenedAt:         happened,
		}
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		if err := adjust(tx, userID, accountID, delta); err != nil {
			return err
		}
		created = t.ID
		return nil
	})
	if err != nil || created == 0 {
		return nil, err
	}
	return s.Get(userID, created)
}

func validateRevalueInput(in TransactionInput) error {
	if in.AccountID == 0 {
		return errors.New("请选择账户")
	}
	if in.Amount == 0 {
		return errors.New("校准差额不能为 0")
	}
	if in.ToAccountID != nil && *in.ToAccountID != 0 {
		return errors.New("市值校准不能设置转入账户")
	}
	if in.CategoryID != nil && *in.CategoryID != 0 {
		return errors.New("市值校准不能设置分类")
	}
	return nil
}

func assertInvestAccount(tx *gorm.DB, userID, accountID uint64) error {
	var acc model.Account
	if err := tx.Select("id", "type").Where("id = ? AND user_id = ?", accountID, userID).First(&acc).Error; err != nil {
		return errors.New("账户不存在")
	}
	if acc.Type != AccountTypeInvest {
		return errors.New("市值校准只能记在理财账户")
	}
	return nil
}
