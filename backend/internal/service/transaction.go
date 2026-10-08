package service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mybookkeeping/internal/model"
)

type TransactionInput struct {
	Type               string
	Amount             int64
	AccountID          uint64
	ToAccountID        *uint64
	CategoryID         *uint64
	Remark             string
	GeoMode            string
	GeoLng             *float64
	GeoLat             *float64
	GeoName            string
	GeoEndLng          *float64
	GeoEndLat          *float64
	GeoEndName         string
	FareRuleID         *uint64
	InstallmentPeriods int
	InterestFen        int64
	InstallmentDeferCycles int
	AmountPending      bool
	ForeignCurrency    string
	ForeignAmount      int64
	VoucherStatus      string // "" | pending | used；refunded 只能经退款接口产生
	VoucherExpireOn    string
	HappenedAt         time.Time
	AttachmentIDs      []uint64
	TagIDs             []uint64
}

type TransactionQuery struct {
	AccountID  *uint64
	CategoryID *uint64
	TagID      *uint64
	Pending    bool
	// Voucher 团购券筛选：pending / used / refunded / all（所有券）
	Voucher    string
	Type       string
	Keyword    string
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

type TransactionListResult struct {
	Items []model.Transaction `json:"items"`
	Total int64               `json:"total"`
}

type TransactionService struct {
	db *gorm.DB
}

func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{db: db}
}

func (s *TransactionService) buildQuery(userID uint64, q TransactionQuery) *gorm.DB {
	db := s.db.Model(&model.Transaction{}).Where("transactions.user_id = ?", userID)
	if q.AccountID != nil {
		db = db.Where("account_id = ? OR to_account_id = ?", *q.AccountID, *q.AccountID)
	}
	if q.CategoryID != nil {
		// 含子分类
		var ids []uint64
		s.db.Model(&model.Category{}).Where("user_id = ? AND (id = ? OR parent_id = ?)", userID, *q.CategoryID, *q.CategoryID).
			Pluck("id", &ids)
		if len(ids) == 0 {
			ids = []uint64{*q.CategoryID}
		}
		db = db.Where("category_id IN ?", ids)
	}
	if q.TagID != nil {
		db = db.Joins("JOIN transaction_tags ON transaction_tags.transaction_id = transactions.id AND transaction_tags.tag_id = ?", *q.TagID)
	}
	if q.Pending {
		db = db.Where("amount_pending = ?", true)
	}
	switch q.Voucher {
	case VoucherPending, VoucherUsed, VoucherRefunded:
		db = db.Where("voucher_status = ?", q.Voucher)
	case "all":
		db = db.Where("voucher_status <> ''")
	}
	if q.Type != "" {
		db = db.Where("type = ?", q.Type)
	}
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("remark LIKE ?", like)
	}
	if q.From != nil {
		db = db.Where("happened_at >= ?", *q.From)
	}
	if q.To != nil {
		db = db.Where("happened_at < ?", *q.To)
	}
	return db
}

// briefAccount 流水关联账户只取名称等，避免把卡号、户名带进流水响应
func briefAccount(db *gorm.DB) *gorm.DB {
	return db.Select("id", "name", "type")
}

// preloadTxRefs 流水只需账户 / 分类名称，只取必要列
func preloadTxRefs(db *gorm.DB) *gorm.DB {
	return db.Preload("Account", briefAccount).Preload("ToAccount", briefAccount).
		Preload("Category", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "kind", "icon", "parent_id")
		}).
		Preload("Attachments").Preload("Tags")
}

func (s *TransactionService) List(userID uint64, q TransactionQuery) (*TransactionListResult, error) {
	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	base := s.buildQuery(userID, q)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.Transaction
	err := preloadTxRefs(s.buildQuery(userID, q)).
		Order("happened_at desc, id desc").
		Limit(limit).Offset(q.Offset).
		Find(&list).Error
	return &TransactionListResult{Items: list, Total: total}, err
}

// RecentPlace 近期流水上的地点（选点页「最近地点」）
type RecentPlace struct {
	GeoMode    string   `json:"geoMode"`
	GeoLng     *float64 `json:"geoLng"`
	GeoLat     *float64 `json:"geoLat"`
	GeoName    string   `json:"geoName"`
	GeoEndLng  *float64 `json:"geoEndLng"`
	GeoEndLat  *float64 `json:"geoEndLat"`
	GeoEndName string   `json:"geoEndName"`
}

// RecentPlaces 只取地理字段，最近 limit 条带地点的流水（不去重，由前端按标签去重）
func (s *TransactionService) RecentPlaces(userID uint64, limit int) ([]RecentPlace, error) {
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	out := make([]RecentPlace, 0, limit)
	err := s.db.Model(&model.Transaction{}).
		Select("geo_mode", "geo_lng", "geo_lat", "geo_name", "geo_end_lng", "geo_end_lat", "geo_end_name").
		Where("user_id = ? AND geo_lng IS NOT NULL AND geo_lat IS NOT NULL AND geo_name <> ''", userID).
		Order("happened_at desc, id desc").Limit(limit).
		Scan(&out).Error
	return out, err
}

func (s *TransactionService) Get(userID, id uint64) (*model.Transaction, error) {
	var t model.Transaction
	err := preloadTxRefs(s.db).
		Where("id = ? AND user_id = ?", id, userID).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TransactionService) bindTags(tx *gorm.DB, userID, txID uint64, tagIDs []uint64) error {
	if err := tx.Where("transaction_id = ?", txID).Delete(&model.TransactionTag{}).Error; err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	var valid []model.Tag
	if err := tx.Where("user_id = ? AND id IN ?", userID, tagIDs).Find(&valid).Error; err != nil {
		return err
	}
	for _, tag := range valid {
		if err := tx.Create(&model.TransactionTag{TransactionID: txID, TagID: tag.ID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *TransactionService) Create(userID uint64, in TransactionInput) (*model.Transaction, error) {
	if in.Type == TxTypeRevalue {
		return nil, errors.New("市值校准请在理财账户里操作")
	}
	if err := validateInput(in); err != nil {
		return nil, err
	}
	cur, famt, err := normalizeForeign(in.Type, in.ForeignCurrency, in.ForeignAmount)
	if err != nil {
		return nil, err
	}
	in.ForeignCurrency, in.ForeignAmount = cur, famt
	in.VoucherStatus, in.VoucherExpireOn = normalizeVoucher(in.Type, in.VoucherStatus, in.VoucherExpireOn)
	var out *model.Transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := assertRefsOwned(tx, userID, txRefs(in)); err != nil {
			return err
		}
		t := &model.Transaction{
			UserID:             userID,
			Type:               in.Type,
			Amount:             in.Amount,
			AccountID:          in.AccountID,
			ToAccountID:        in.ToAccountID,
			CategoryID:         in.CategoryID,
			Remark:             in.Remark,
			GeoMode:            normalizeGeoMode(in),
			GeoLng:             in.GeoLng,
			GeoLat:             in.GeoLat,
			GeoName:            in.GeoName,
			GeoEndLng:          in.GeoEndLng,
			GeoEndLat:          in.GeoEndLat,
			GeoEndName:         in.GeoEndName,
			FareRuleID:         in.FareRuleID,
			InstallmentPeriods: normalizeInstallmentPeriods(in.InstallmentPeriods),
			InterestFen:        maxInt64(0, in.InterestFen),
			InstallmentDeferCycles: normalizeDeferCycles(in.InstallmentPeriods, in.InstallmentDeferCycles),
			AmountPending:      in.AmountPending,
			ForeignCurrency:    in.ForeignCurrency,
			ForeignAmount:      in.ForeignAmount,
			VoucherStatus:      in.VoucherStatus,
			VoucherExpireOn:    in.VoucherExpireOn,
			HappenedAt:         in.HappenedAt,
		}
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		if err := applyBalance(tx, userID, in, true); err != nil {
			return err
		}
		if len(in.AttachmentIDs) > 0 {
			if err := tx.Model(&model.Attachment{}).
				Where("id IN ? AND user_id = ? AND transaction_id IS NULL", in.AttachmentIDs, userID).
				Update("transaction_id", t.ID).Error; err != nil {
				return err
			}
		}
		if err := s.bindTags(tx, userID, t.ID, in.TagIDs); err != nil {
			return err
		}
		out = t
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Get(userID, out.ID)
}

func (s *TransactionService) Update(userID, id uint64, in TransactionInput) (*model.Transaction, error) {
	if err := validateInput(in); err != nil {
		return nil, err
	}
	cur, famt, err := normalizeForeign(in.Type, in.ForeignCurrency, in.ForeignAmount)
	if err != nil {
		return nil, err
	}
	in.ForeignCurrency, in.ForeignAmount = cur, famt
	in.VoucherStatus, in.VoucherExpireOn = normalizeVoucher(in.Type, in.VoucherStatus, in.VoucherExpireOn)
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var old model.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", id, userID).First(&old).Error; err != nil {
			return err
		}
		if err := assertRefsOwned(tx, userID, txRefs(in)); err != nil {
			return err
		}
		if (old.Type == TxTypeRevalue) != (in.Type == TxTypeRevalue) {
			return errors.New("市值校准流水不能与其他类型互改")
		}
		if in.Type == TxTypeRevalue {
			if err := assertInvestAccount(tx, userID, in.AccountID); err != nil {
				return err
			}
		}
		oldIn := TransactionInput{
			Type:        old.Type,
			Amount:      old.Amount,
			AccountID:   old.AccountID,
			ToAccountID: old.ToAccountID,
		}
		if err := applyBalance(tx, userID, oldIn, false); err != nil {
			return err
		}
		old.Type = in.Type
		old.Amount = in.Amount
		old.AccountID = in.AccountID
		old.ToAccountID = in.ToAccountID
		old.CategoryID = in.CategoryID
		old.Remark = in.Remark
		old.GeoMode = normalizeGeoMode(in)
		old.GeoLng = in.GeoLng
		old.GeoLat = in.GeoLat
		old.GeoName = in.GeoName
		old.GeoEndLng = in.GeoEndLng
		old.GeoEndLat = in.GeoEndLat
		old.GeoEndName = in.GeoEndName
		old.FareRuleID = in.FareRuleID
		old.InstallmentPeriods = normalizeInstallmentPeriods(in.InstallmentPeriods)
		old.InterestFen = maxInt64(0, in.InterestFen)
		old.InstallmentDeferCycles = normalizeDeferCycles(in.InstallmentPeriods, in.InstallmentDeferCycles)
		old.AmountPending = in.AmountPending
		old.ForeignCurrency = in.ForeignCurrency
		old.ForeignAmount = in.ForeignAmount
		// 已退款的券状态由退款流水维护，普通编辑不改
		if old.VoucherStatus == VoucherRefunded && in.Type == "expense" {
			old.VoucherExpireOn = in.VoucherExpireOn
		} else {
			old.VoucherStatus = in.VoucherStatus
			old.VoucherExpireOn = in.VoucherExpireOn
		}
		old.HappenedAt = in.HappenedAt
		if err := tx.Save(&old).Error; err != nil {
			return err
		}
		if err := applyBalance(tx, userID, in, true); err != nil {
			return err
		}
		tx.Model(&model.Attachment{}).Where("transaction_id = ? AND user_id = ?", id, userID).
			Update("transaction_id", nil)
		if len(in.AttachmentIDs) > 0 {
			if err := tx.Model(&model.Attachment{}).
				Where("id IN ? AND user_id = ?", in.AttachmentIDs, userID).
				Update("transaction_id", id).Error; err != nil {
				return err
			}
		}
		return s.bindTags(tx, userID, id, in.TagIDs)
	})
	if err != nil {
		return nil, err
	}
	return s.Get(userID, id)
}

func (s *TransactionService) Delete(userID, id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var old model.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", id, userID).First(&old).Error; err != nil {
			return err
		}
		oldIn := TransactionInput{
			Type:        old.Type,
			Amount:      old.Amount,
			AccountID:   old.AccountID,
			ToAccountID: old.ToAccountID,
		}
		if err := applyBalance(tx, userID, oldIn, false); err != nil {
			return err
		}
		tx.Where("transaction_id = ?", id).Delete(&model.TransactionTag{})
		tx.Model(&model.Attachment{}).Where("transaction_id = ?", id).Update("transaction_id", nil)
		// 删退款流水：原券恢复待核销；删原券：退款流水解除关联
		if old.RefundOfID != nil {
			if err := tx.Model(&model.Transaction{}).
				Where("id = ? AND user_id = ? AND voucher_status = ?", *old.RefundOfID, userID, VoucherRefunded).
				Update("voucher_status", VoucherPending).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&model.Transaction{}).
			Where("refund_of_id = ? AND user_id = ?", id, userID).
			Update("refund_of_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&old).Error
	})
}

func validateInput(in TransactionInput) error {
	if in.Type == TxTypeRevalue {
		return validateRevalueInput(in)
	}
	if in.Amount < 0 {
		return errors.New("金额无效")
	}
	if in.Amount == 0 && (in.FareRuleID == nil || *in.FareRuleID == 0) {
		return errors.New("金额必须大于 0")
	}
	if in.AccountID == 0 {
		return errors.New("请选择账户")
	}
	if in.InstallmentPeriods < 0 || in.InstallmentPeriods > 60 {
		return errors.New("分期期数无效")
	}
	if in.InterestFen < 0 {
		return errors.New("利息无效")
	}
	if in.InstallmentDeferCycles < 0 || in.InstallmentDeferCycles > maxDeferCycles {
		return errors.New("分期首期出账设置无效")
	}
	switch in.Type {
	case "expense", "income":
		if in.CategoryID == nil || *in.CategoryID == 0 {
			return errors.New("请选择分类")
		}
	case "transfer":
		if in.ToAccountID == nil || *in.ToAccountID == 0 {
			return errors.New("请选择转入账户")
		}
		if *in.ToAccountID == in.AccountID {
			return errors.New("转出与转入账户不能相同")
		}
	default:
		return errors.New("类型无效")
	}
	if in.HappenedAt.IsZero() {
		in.HappenedAt = time.Now()
	}
	return nil
}

// normalizeForeign 外币字段：币种为 3–8 位字母（转大写），金额 >0；仅收支可用，转账或未填币种时清空
func normalizeForeign(txType, currency string, amount int64) (string, int64, error) {
	cur := strings.ToUpper(strings.TrimSpace(currency))
	if cur == "" || txType == "transfer" {
		return "", 0, nil
	}
	if len(cur) < 3 || len(cur) > 8 {
		return "", 0, errors.New("外币币种无效")
	}
	for _, r := range cur {
		if r < 'A' || r > 'Z' {
			return "", 0, errors.New("外币币种无效")
		}
	}
	if cur == "CNY" {
		return "", 0, errors.New("外币币种不能是人民币")
	}
	if amount <= 0 {
		return "", 0, errors.New("请填写外币金额")
	}
	return cur, amount, nil
}

func normalizeInstallmentPeriods(n int) int {
	if n < 1 {
		return 1
	}
	if n > 60 {
		return 60
	}
	return n
}

const maxDeferCycles = 3

// normalizeDeferCycles 仅分期（期数 > 1）可推迟首期
func normalizeDeferCycles(periods, n int) int {
	if periods <= 1 || n < 0 {
		return 0
	}
	if n > maxDeferCycles {
		return maxDeferCycles
	}
	return n
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func normalizeGeoMode(in TransactionInput) string {
	if in.GeoMode == "route" && in.GeoEndLng != nil && in.GeoEndLat != nil {
		return "route"
	}
	if in.GeoLng == nil && in.GeoLat == nil {
		return ""
	}
	return "point"
}

func applyBalance(tx *gorm.DB, userID uint64, in TransactionInput, apply bool) error {
	sign := int64(1)
	if !apply {
		sign = -1
	}
	switch in.Type {
	case "expense":
		return adjust(tx, userID, in.AccountID, -in.Amount*sign)
	case "income", TxTypeRevalue:
		return adjust(tx, userID, in.AccountID, in.Amount*sign)
	case "transfer":
		if err := adjust(tx, userID, in.AccountID, -in.Amount*sign); err != nil {
			return err
		}
		return adjust(tx, userID, *in.ToAccountID, in.Amount*sign)
	}
	return nil
}

func adjust(tx *gorm.DB, userID, accountID uint64, delta int64) error {
	res := tx.Model(&model.Account{}).
		Where("id = ? AND user_id = ?", accountID, userID).
		Update("balance", gorm.Expr("balance + ?", delta))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("账户不存在")
	}
	// 信用账户：已用额度下降时，已出账不得超过已用
	var a model.Account
	if err := tx.Select("id", "type", "balance", "credit_billed_fen").
		Where("id = ? AND user_id = ?", accountID, userID).First(&a).Error; err != nil {
		return err
	}
	if a.Type == "credit" {
		used := CreditUsedFen(&a)
		if a.CreditBilledFen > used {
			if err := tx.Model(&a).Update("credit_billed_fen", used).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
