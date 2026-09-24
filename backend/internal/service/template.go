package service

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"mybookkeeping/internal/model"
)

type TemplateService struct {
	db       *gorm.DB
	fareSvc  *FareRuleService
}

func NewTemplateService(db *gorm.DB, fareSvc *FareRuleService) *TemplateService {
	return &TemplateService{db: db, fareSvc: fareSvc}
}

type TemplateInput struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Amount      int64    `json:"amount"`
	AccountID   *uint64  `json:"accountId"`
	ToAccountID *uint64  `json:"toAccountId"`
	CategoryID  *uint64  `json:"categoryId"`
	Remark      string   `json:"remark"`
	GeoMode     string   `json:"geoMode"`
	GeoLng      *float64 `json:"geoLng"`
	GeoLat      *float64 `json:"geoLat"`
	GeoName     string   `json:"geoName"`
	GeoEndLng   *float64 `json:"geoEndLng"`
	GeoEndLat   *float64 `json:"geoEndLat"`
	GeoEndName  string   `json:"geoEndName"`
	FareRuleID  *uint64  `json:"fareRuleId"`
	TagIDs      []uint64 `json:"tagIds"`
	Icon        string   `json:"icon"`
	Sort        int      `json:"sort"`
}

func encodeIDs(ids []uint64) string {
	if len(ids) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(ids)
	return string(b)
}

func decodeIDs(s string) []uint64 {
	if s == "" {
		return nil
	}
	var ids []uint64
	_ = json.Unmarshal([]byte(s), &ids)
	return ids
}

func hydrateTemplate(t *model.Template) {
	t.TagIDs = decodeIDs(t.TagIDsJSON)
	if t.FareRule != nil {
		hydrateFareRule(t.FareRule)
	}
}

func normalizeTemplateGeo(in *TemplateInput) (mode string, endLng, endLat *float64, endName string) {
	if in.GeoLng == nil || in.GeoLat == nil {
		return "", nil, nil, ""
	}
	if in.GeoMode == "route" && in.GeoEndLng != nil && in.GeoEndLat != nil {
		return "route", in.GeoEndLng, in.GeoEndLat, strings.TrimSpace(in.GeoEndName)
	}
	return "point", nil, nil, ""
}

func (s *TemplateService) List(userID uint64) ([]model.Template, error) {
	var list []model.Template
	err := s.db.Preload("FareRule").Where("user_id = ?", userID).Order("sort asc, id asc").Find(&list).Error
	for i := range list {
		hydrateTemplate(&list[i])
	}
	return list, err
}

func (s *TemplateService) Create(userID uint64, in TemplateInput) (*model.Template, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.New("模板名称不能为空")
	}
	if in.Type == "" {
		in.Type = "expense"
	}
	if err := assertRefsOwned(s.db, userID, refIDs{Account: in.AccountID, ToAccount: in.ToAccountID, Category: in.CategoryID, FareRule: in.FareRuleID}); err != nil {
		return nil, err
	}
	mode, endLng, endLat, endName := normalizeTemplateGeo(&in)
	var fareID *uint64
	if in.FareRuleID != nil && *in.FareRuleID > 0 {
		fareID = in.FareRuleID
	}
	t := &model.Template{
		UserID: userID, Name: strings.TrimSpace(in.Name), Type: in.Type,
		Amount: in.Amount, AccountID: in.AccountID, ToAccountID: in.ToAccountID,
		CategoryID: in.CategoryID, Remark: in.Remark, Sort: in.Sort,
		Icon: strings.TrimSpace(in.Icon),
		GeoMode: mode, GeoLng: in.GeoLng, GeoLat: in.GeoLat, GeoName: strings.TrimSpace(in.GeoName),
		GeoEndLng: endLng, GeoEndLat: endLat, GeoEndName: endName,
		FareRuleID: fareID,
		TagIDsJSON: encodeIDs(in.TagIDs),
	}
	if mode == "" {
		t.GeoLng, t.GeoLat, t.GeoName = nil, nil, ""
	}
	if err := s.db.Create(t).Error; err != nil {
		return nil, err
	}
	return s.Get(userID, t.ID)
}

func (s *TemplateService) Update(userID, id uint64, in TemplateInput) (*model.Template, error) {
	var t model.Template
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&t).Error; err != nil {
		return nil, err
	}
	if err := assertRefsOwned(s.db, userID, refIDs{Account: in.AccountID, ToAccount: in.ToAccountID, Category: in.CategoryID, FareRule: in.FareRuleID}); err != nil {
		return nil, err
	}
	if in.Name != "" {
		t.Name = strings.TrimSpace(in.Name)
	}
	if in.Type != "" {
		t.Type = in.Type
	}
	t.Amount = in.Amount
	t.AccountID = in.AccountID
	t.ToAccountID = in.ToAccountID
	t.CategoryID = in.CategoryID
	t.Remark = in.Remark
	t.Sort = in.Sort
	t.Icon = strings.TrimSpace(in.Icon)
	if t.Icon == "" {
		t.Icon = "mdi-flash"
	}
	t.TagIDsJSON = encodeIDs(in.TagIDs)
	if in.FareRuleID != nil && *in.FareRuleID > 0 {
		t.FareRuleID = in.FareRuleID
	} else {
		t.FareRuleID = nil
	}
	t.FareRuleJSON = ""

	mode, endLng, endLat, endName := normalizeTemplateGeo(&in)
	t.GeoMode = mode
	if mode == "" {
		t.GeoLng, t.GeoLat, t.GeoName = nil, nil, ""
		t.GeoEndLng, t.GeoEndLat, t.GeoEndName = nil, nil, ""
	} else {
		t.GeoLng = in.GeoLng
		t.GeoLat = in.GeoLat
		t.GeoName = strings.TrimSpace(in.GeoName)
		t.GeoEndLng = endLng
		t.GeoEndLat = endLat
		t.GeoEndName = endName
	}

	if err := s.db.Model(&t).Select(
		"Name", "Type", "Amount", "AccountID", "ToAccountID", "CategoryID", "Remark", "Sort", "Icon",
		"TagIDsJSON", "FareRuleID", "FareRuleJSON",
		"GeoMode", "GeoLng", "GeoLat", "GeoName", "GeoEndLng", "GeoEndLat", "GeoEndName",
	).Updates(&t).Error; err != nil {
		return nil, err
	}
	return s.Get(userID, t.ID)
}

func (s *TemplateService) Delete(userID, id uint64) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Template{}).Error
}

func (s *TemplateService) Get(userID, id uint64) (*model.Template, error) {
	var t model.Template
	if err := s.db.Preload("FareRule").Where("id = ? AND user_id = ?", id, userID).First(&t).Error; err != nil {
		return nil, err
	}
	hydrateTemplate(&t)
	return &t, nil
}

func (s *TemplateService) PreviewFare(userID, id uint64, at time.Time, spentOverride *int64) (*FarePreviewResult, error) {
	t, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	if t.FareRuleID == nil || *t.FareRuleID == 0 {
		return &FarePreviewResult{
			AmountFen: t.Amount,
			Reason:    "模板未绑定计费规则，使用固定金额",
		}, nil
	}
	// 模板金额优先作原价；规则内 BaseFen 仅作兜底
	var base *int64
	if t.Amount > 0 {
		base = &t.Amount
	}
	return s.fareSvc.Preview(userID, *t.FareRuleID, at, spentOverride, base, nil)
}
