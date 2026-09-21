package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"mybookkeeping/internal/model"
)

type TagService struct {
	db *gorm.DB
}

func NewTagService(db *gorm.DB) *TagService {
	return &TagService{db: db}
}

func (s *TagService) List(userID uint64) ([]model.Tag, error) {
	var list []model.Tag
	err := s.db.Where("user_id = ?", userID).Order("sort asc, id asc").Find(&list).Error
	return list, err
}

func (s *TagService) Create(userID uint64, name, color string, sort int) (*model.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("标签名称不能为空")
	}
	if color == "" {
		color = "#1b7f5a"
	}
	t := &model.Tag{UserID: userID, Name: name, Color: color, Sort: sort}
	if err := s.db.Create(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TagService) Update(userID, id uint64, name, color string, sort int) (*model.Tag, error) {
	var t model.Tag
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&t).Error; err != nil {
		return nil, err
	}
	if name != "" {
		t.Name = strings.TrimSpace(name)
	}
	if color != "" {
		t.Color = color
	}
	t.Sort = sort
	if err := s.db.Save(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TagService) Delete(userID, id uint64) error {
	s.db.Where("tag_id = ?", id).Delete(&model.TransactionTag{})
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Tag{}).Error
}

// 常用标签预设（原版无默认标签，按个人记账场景补充）
var defaultTagPresets = []struct {
	Name  string
	Color string
}{
	{"报销", "#c45c3e"},
	{"出差", "#5c6bc0"},
	{"工作", "#546e7a"},
	{"家庭", "#1b7f5a"},
	{"个人", "#2e7d6f"},
	{"情侣", "#8e24aa"},
	{"大额", "#d4a017"},
	{"待还", "#c62828"},
	{"礼物", "#e91e63"},
	{"可抵税", "#00897b"},
}

// EnsureDefaults 无标签时写入常用预设；已有标签则按名称补齐缺失项
func (s *TagService) EnsureDefaults(userID uint64) error {
	var existing []model.Tag
	if err := s.db.Where("user_id = ?", userID).Find(&existing).Error; err != nil {
		return err
	}
	have := map[string]struct{}{}
	for _, t := range existing {
		have[t.Name] = struct{}{}
	}
	sortBase := len(existing)
	var toCreate []model.Tag
	for i, p := range defaultTagPresets {
		if _, ok := have[p.Name]; ok {
			continue
		}
		toCreate = append(toCreate, model.Tag{
			UserID: userID, Name: p.Name, Color: p.Color, Sort: sortBase + i + 1,
		})
	}
	if len(toCreate) == 0 {
		return nil
	}
	return s.db.Create(&toCreate).Error
}
