package service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybookkeeping/internal/model"
)

type AccountService struct {
	db *gorm.DB
}

func NewAccountService(db *gorm.DB) *AccountService {
	return &AccountService{db: db}
}

func (s *AccountService) List(userID uint64, includeArchived bool) ([]model.Account, error) {
	q := s.db.Where("user_id = ?", userID)
	if !includeArchived {
		q = q.Where("archived = ?", false)
	}
	var list []model.Account
	err := q.Preload("Attachments").Order("sort asc, id asc").Find(&list).Error
	if err != nil {
		return nil, err
	}
	MaskAccountsInPlace(list)
	return list, nil
}

type AccountInput struct {
	Name             string
	Type             string
	Balance          int64
	SetBalance       bool // 更新时是否写入余额（含设为 0）
	Sort             int
	Archived         *bool
	CreditLimit      *int64
	CreditBilledFen  *int64 // 信用卡已出账（分）
	BillingDay       *int
	PaymentDueDay    *int
	Institution      *string
	CardNo           *string
	HolderName       *string
	StorageNote      *string
	Remark           *string
	AttachmentIDs    []uint64 // nil=不改附件；非 nil（可空切片）=覆盖绑定
	LastReconciledAt *time.Time
	ClearReconciled  bool
}

func (s *AccountService) Create(userID uint64, in AccountInput) (*model.Account, error) {
	if in.Name == "" {
		return nil, errors.New("账户名称不能为空")
	}
	if in.Type == "" {
		in.Type = "cash"
	}
	bal := in.Balance
	var billed int64
	if in.Type == "credit" {
		bal, billed = normalizeCreditBalances(in.Balance, in.CreditBilledFen)
	}
	a := &model.Account{
		UserID: userID, Name: in.Name, Type: in.Type,
		Balance: bal, Sort: in.Sort, CreditBilledFen: billed,
		Institution: strPtr(in.Institution),
		StorageNote: strPtr(in.StorageNote),
		Remark:      strPtr(in.Remark),
	}
	if in.CardNo != nil {
		if v, ok := SanitizeSecretWrite(*in.CardNo); ok {
			a.CardNo = v
		}
	}
	if in.HolderName != nil {
		if v, ok := SanitizeSecretWrite(*in.HolderName); ok {
			a.HolderName = v
		}
	}
	if in.CreditLimit != nil {
		a.CreditLimit = *in.CreditLimit
	}
	if in.BillingDay != nil {
		a.BillingDay = *in.BillingDay
	}
	if in.PaymentDueDay != nil {
		a.PaymentDueDay = *in.PaymentDueDay
	}
	if err := s.db.Create(a).Error; err != nil {
		return nil, err
	}
	if in.AttachmentIDs != nil {
		_ = s.bindAttachments(userID, a.ID, in.AttachmentIDs)
	}
	return s.Get(userID, a.ID)
}

func strPtr(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

func (s *AccountService) Get(userID, id uint64) (*model.Account, error) {
	var a model.Account
	if err := s.db.Preload("Attachments").Where("id = ? AND user_id = ?", id, userID).First(&a).Error; err != nil {
		return nil, err
	}
	MaskAccountInPlace(&a)
	return &a, nil
}

// GetSensitive 返回卡号/户名明文（仅显式请求；不缓存到列表）。
func (s *AccountService) GetSensitive(userID, id uint64) (cardNo, holderName string, err error) {
	var a model.Account
	if err = s.db.Select("id", "card_no", "holder_name").
		Where("id = ? AND user_id = ?", id, userID).First(&a).Error; err != nil {
		return "", "", err
	}
	return a.CardNo, a.HolderName, nil
}

func (s *AccountService) bindAttachments(userID, accountID uint64, ids []uint64) error {
	// 先解绑该账户旧附件
	if err := s.db.Model(&model.Attachment{}).
		Where("user_id = ? AND account_id = ?", userID, accountID).
		Update("account_id", nil).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	return s.db.Model(&model.Attachment{}).
		Where("id IN ? AND user_id = ?", ids, userID).
		Updates(map[string]any{"account_id": accountID, "transaction_id": nil}).Error
}

func (s *AccountService) Update(userID, id uint64, in AccountInput) (*model.Account, error) {
	var a model.Account
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&a).Error; err != nil {
		return nil, err
	}
	if in.Name != "" {
		a.Name = in.Name
	}
	if in.Type != "" {
		a.Type = in.Type
	}
	a.Sort = in.Sort
	if in.Archived != nil {
		a.Archived = *in.Archived
	}
	if in.CreditLimit != nil {
		a.CreditLimit = *in.CreditLimit
	}
	if in.BillingDay != nil {
		a.BillingDay = *in.BillingDay
	}
	if in.PaymentDueDay != nil {
		a.PaymentDueDay = *in.PaymentDueDay
	}
	if in.Institution != nil {
		a.Institution = strings.TrimSpace(*in.Institution)
	}
	if in.CardNo != nil {
		if v, ok := SanitizeSecretWrite(*in.CardNo); ok {
			a.CardNo = v
		}
		// 脱敏串：忽略，保留库内原值
	}
	if in.HolderName != nil {
		if v, ok := SanitizeSecretWrite(*in.HolderName); ok {
			a.HolderName = v
		}
	}
	if in.StorageNote != nil {
		a.StorageNote = strings.TrimSpace(*in.StorageNote)
	}
	if in.Remark != nil {
		a.Remark = strings.TrimSpace(*in.Remark)
	}
	if in.ClearReconciled {
		a.LastReconciledAt = nil
	} else if in.LastReconciledAt != nil {
		a.LastReconciledAt = in.LastReconciledAt
	}

	accType := a.Type
	if in.Type != "" {
		accType = in.Type
	}
	if accType == "credit" {
		bal := a.Balance
		if in.SetBalance {
			bal = in.Balance
		}
		billedPtr := in.CreditBilledFen
		if billedPtr == nil {
			v := a.CreditBilledFen
			billedPtr = &v
		}
		a.Balance, a.CreditBilledFen = normalizeCreditBalances(bal, billedPtr)
	} else {
		if in.SetBalance {
			a.Balance = in.Balance
		}
		a.CreditBilledFen = 0
	}

	if err := s.db.Save(&a).Error; err != nil {
		return nil, err
	}
	if in.AttachmentIDs != nil {
		_ = s.bindAttachments(userID, a.ID, in.AttachmentIDs)
	}
	return s.Get(userID, a.ID)
}

// normalizeCreditBalances 信用账户：余额存为「−已用额度」；已出账夹在 [0, 已用] 内。
// 入参 Balance 可为「已用额度（正）」或「库内负余额」；若提供 CreditBilledFen 则按已用夹紧。
func normalizeCreditBalances(balance int64, billedPtr *int64) (storedBalance, billed int64) {
	used := balance
	if used < 0 {
		used = -used
	}
	storedBalance = -used
	if billedPtr == nil {
		return storedBalance, 0
	}
	billed = *billedPtr
	if billed < 0 {
		billed = 0
	}
	if billed > used {
		billed = used
	}
	return storedBalance, billed
}

// CreditUsedFen 已用额度（分）
func CreditUsedFen(a *model.Account) int64 {
	if a == nil {
		return 0
	}
	if a.Balance >= 0 {
		return 0
	}
	return -a.Balance
}

// CreditUnbilledFen 未出账（分）
func CreditUnbilledFen(a *model.Account) int64 {
	used := CreditUsedFen(a)
	billed := a.CreditBilledFen
	if billed < 0 {
		billed = 0
	}
	if billed > used {
		billed = used
	}
	return used - billed
}

func (s *AccountService) Delete(userID, id uint64) error {
	var count int64
	s.db.Model(&model.Transaction{}).Where("user_id = ? AND (account_id = ? OR to_account_id = ?)", userID, id, id).Count(&count)
	if count > 0 {
		return errors.New("账户下仍有流水，请先归档或迁移")
	}
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Account{}).Error
}
