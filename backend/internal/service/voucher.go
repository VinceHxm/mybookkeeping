package service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mybookkeeping/internal/i18n"
	"mybookkeeping/internal/model"
)

const (
	VoucherPending  = "pending"
	VoucherUsed     = "used"
	VoucherRefunded = "refunded"

	refundCategoryName = "退款"
	refundRemarkPrefix = "团购券退款"
)

// normalizeVoucher 仅支出可标记为券；refunded 不能经普通保存写入
func normalizeVoucher(txType, status, expireOn string) (string, string) {
	if txType != "expense" {
		return "", ""
	}
	status = strings.TrimSpace(status)
	if status != VoucherPending && status != VoucherUsed {
		return "", ""
	}
	expireOn = strings.TrimSpace(expireOn)
	if _, err := time.Parse("2006-01-02", expireOn); err != nil {
		expireOn = ""
	}
	return status, expireOn
}

// SetVoucherStatus 一键核销 / 撤销核销（pending ↔ used）
func (s *TransactionService) SetVoucherStatus(userID, id uint64, status string) (*model.Transaction, error) {
	if status != VoucherPending && status != VoucherUsed {
		return nil, errors.New("券状态无效")
	}
	var t model.Transaction
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&t).Error; err != nil {
		return nil, errors.New("流水不存在")
	}
	if t.Type != "expense" {
		return nil, errors.New("仅支出可标记团购券")
	}
	if t.VoucherStatus == VoucherRefunded {
		return nil, errors.New("该券已退款，删除退款流水后才能改状态")
	}
	if err := s.db.Model(&t).Update("voucher_status", status).Error; err != nil {
		return nil, err
	}
	return s.Get(userID, id)
}

type VoucherRefundInput struct {
	Amount     int64     // 0 表示按原券金额全额退
	AccountID  uint64    // 0 表示退回原支付账户
	CategoryID *uint64   // nil 表示用收入分类「退款」（不存在则自动创建）
	HappenedAt time.Time // 退款入账时间（信用卡建议填银行记账日）
	Remark     string
}

// RefundVoucher 为待核销 / 已核销的团购券生成一笔退款收入，并把券标记为已退款
func (s *TransactionService) RefundVoucher(userID, id uint64, in VoucherRefundInput) (*model.Transaction, error) {
	var refundID uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var orig model.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", id, userID).First(&orig).Error; err != nil {
			return errors.New("流水不存在")
		}
		if orig.Type != "expense" || (orig.VoucherStatus != VoucherPending && orig.VoucherStatus != VoucherUsed) {
			return errors.New("只有待核销 / 已核销的团购券可以退款")
		}
		amount := in.Amount
		if amount <= 0 {
			amount = orig.Amount
		}
		if amount > orig.Amount {
			return errors.New("退款金额不能超过原支付金额")
		}
		accountID := in.AccountID
		if accountID == 0 {
			accountID = orig.AccountID
		}
		locale := userLocale(tx, userID)
		catID := in.CategoryID
		if catID == nil || *catID == 0 {
			id, err := ensureRefundCategory(tx, userID, locale)
			if err != nil {
				return err
			}
			catID = &id
		}
		happened := in.HappenedAt
		if happened.IsZero() {
			happened = time.Now()
		}
		remark := strings.TrimSpace(in.Remark)
		if remark == "" {
			remark = refundRemark(locale, orig.Remark)
		}
		refIn := TransactionInput{Type: "income", Amount: amount, AccountID: accountID, CategoryID: catID}
		if err := assertRefsOwned(tx, userID, txRefs(refIn)); err != nil {
			return err
		}
		origID := orig.ID
		r := &model.Transaction{
			UserID:             userID,
			Type:               "income",
			Amount:             amount,
			AccountID:          accountID,
			CategoryID:         catID,
			Remark:             remark,
			InstallmentPeriods: 1,
			RefundOfID:         &origID,
			HappenedAt:         happened,
		}
		if err := tx.Create(r).Error; err != nil {
			return err
		}
		if err := applyBalance(tx, userID, refIn, true); err != nil {
			return err
		}
		if err := tx.Model(&orig).Update("voucher_status", VoucherRefunded).Error; err != nil {
			return err
		}
		refundID = r.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Get(userID, refundID)
}

// userLocale 用户界面语言；未设置（老账号）按中文
func userLocale(tx *gorm.DB, userID uint64) string {
	var u model.User
	if err := tx.Select("locale").Where("id = ?", userID).First(&u).Error; err != nil {
		return i18n.ZhCN
	}
	if l := i18n.Normalize(u.Locale); l != "" {
		return l
	}
	return i18n.ZhCN
}

// ensureRefundCategory 找或建收入分类「退款」：按 PresetKey 识别，兼容老数据按任一语言的名字识别；
// 新建时与默认分类一样写入四语名，Name 取用户当前语言。
func ensureRefundCategory(tx *gorm.DB, userID uint64, locale string) (uint64, error) {
	names := []string{refundCategoryName}
	for _, l := range []string{i18n.En, i18n.Ja, i18n.Ko} {
		names = append(names, i18n.Preset(l, refundCategoryName))
	}
	var c model.Category
	err := tx.Where("user_id = ? AND kind = ? AND (preset_key = ? OR name IN ?)", userID, "income", refundCategoryName, names).
		Order("preset_key = '', parent_id IS NOT NULL, id").First(&c).Error
	if err == nil {
		return c.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	c = model.Category{UserID: userID, Kind: "income", Icon: "mdi-cash-refund"}
	fillPreset(&c, presetNode{Name: refundCategoryName})
	c.Name = localizedName(&c, locale)
	if err := tx.Create(&c).Error; err != nil {
		return 0, err
	}
	return c.ID, nil
}

// refundRemark 退款流水的默认备注，按用户语言生成
func refundRemark(locale, origRemark string) string {
	remark := i18n.T(locale, refundRemarkPrefix)
	if r := strings.TrimSpace(origRemark); r != "" {
		sep := "："
		if locale == i18n.En || locale == i18n.Ko {
			sep = ": "
		}
		remark += sep + r
	}
	return remark
}
