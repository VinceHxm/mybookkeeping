package service

import (
	"errors"

	"gorm.io/gorm"

	"mybookkeeping/internal/model"
)

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) List(userID uint64, kind string) ([]model.Category, error) {
	q := s.db.Where("user_id = ?", userID)
	if kind != "" {
		q = q.Where("kind = ?", kind)
	}
	var list []model.Category
	err := q.Order("sort asc, id asc").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return nestCategories(list), nil
}

func (s *CategoryService) ListFlat(userID uint64, kind string) ([]model.Category, error) {
	q := s.db.Where("user_id = ?", userID)
	if kind != "" {
		q = q.Where("kind = ?", kind)
	}
	var list []model.Category
	err := q.Order("sort asc, id asc").Find(&list).Error
	return list, err
}

func nestCategories(flat []model.Category) []model.Category {
	byID := map[uint64]*model.Category{}
	for i := range flat {
		c := flat[i]
		c.Children = nil
		byID[c.ID] = &flat[i]
	}
	var roots []model.Category
	for i := range flat {
		c := &flat[i]
		if c.ParentID != nil && byID[*c.ParentID] != nil {
			p := byID[*c.ParentID]
			p.Children = append(p.Children, *c)
		} else {
			roots = append(roots, *c)
		}
	}
	// 重新挂 children（上面 append 的是拷贝，再扫一遍）
	for i := range roots {
		id := roots[i].ID
		roots[i].Children = nil
		for j := range flat {
			if flat[j].ParentID != nil && *flat[j].ParentID == id {
				child := flat[j]
				child.Children = nil
				roots[i].Children = append(roots[i].Children, child)
			}
		}
	}
	return roots
}

type CategoryInput struct {
	Name     string
	Kind     string
	Icon     string
	Sort     int
	ParentID *uint64
}

func (s *CategoryService) Create(userID uint64, in CategoryInput) (*model.Category, error) {
	if in.Name == "" {
		return nil, errors.New("分类名称不能为空")
	}
	if in.Kind != "expense" && in.Kind != "income" {
		return nil, errors.New("kind 必须是 expense 或 income")
	}
	if in.ParentID != nil && *in.ParentID > 0 {
		var parent model.Category
		if err := s.db.Where("id = ? AND user_id = ?", *in.ParentID, userID).First(&parent).Error; err != nil {
			return nil, errors.New("父分类不存在")
		}
		if parent.ParentID != nil {
			return nil, errors.New("仅支持两级分类")
		}
		if parent.Kind != in.Kind {
			return nil, errors.New("子分类类型须与父分类一致")
		}
	} else {
		in.ParentID = nil
	}
	c := &model.Category{
		UserID: userID, Name: in.Name, Kind: in.Kind,
		Icon: in.Icon, Sort: in.Sort, ParentID: in.ParentID,
	}
	if err := s.db.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) Update(userID, id uint64, in CategoryInput) (*model.Category, error) {
	var c model.Category
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&c).Error; err != nil {
		return nil, err
	}
	if in.Name != "" {
		c.Name = in.Name
	}
	c.Icon = in.Icon
	c.Sort = in.Sort
	if in.ParentID != nil {
		if *in.ParentID == 0 {
			c.ParentID = nil
		} else if *in.ParentID == id {
			return nil, errors.New("不能将自己设为父分类")
		} else {
			var parent model.Category
			if err := s.db.Where("id = ? AND user_id = ?", *in.ParentID, userID).First(&parent).Error; err != nil {
				return nil, errors.New("父分类不存在")
			}
			if parent.ParentID != nil {
				return nil, errors.New("仅支持两级分类")
			}
			// 若当前分类已有子类，不可再变为二级
			var childCount int64
			s.db.Model(&model.Category{}).Where("parent_id = ?", id).Count(&childCount)
			if childCount > 0 {
				return nil, errors.New("已有子分类，不能降为二级")
			}
			c.ParentID = in.ParentID
		}
	}
	if err := s.db.Model(&c).Select("Name", "Icon", "Sort", "ParentID").Updates(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *CategoryService) Delete(userID, id uint64) error {
	var childCount int64
	s.db.Model(&model.Category{}).Where("user_id = ? AND parent_id = ?", userID, id).Count(&childCount)
	if childCount > 0 {
		return errors.New("请先删除子分类")
	}
	var count int64
	s.db.Model(&model.Transaction{}).Where("user_id = ? AND category_id = ?", userID, id).Count(&count)
	if count > 0 {
		return errors.New("分类下仍有流水，无法删除")
	}
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Category{}).Error
}

// presetNode 默认分类树节点（对齐原版 ezbookkeeping 中文预设，图标用 MDI）
type presetNode struct {
	Name     string
	Icon     string
	Children []presetNode
}

// 支出：参考 src/consts/category.ts + zh_Hans.json，并保留本项目已有一级名（餐饮/交通/购物等）便于老用户补二级
var defaultExpensePresets = []presetNode{
	{Name: "餐饮", Icon: "mdi-food", Children: []presetNode{
		{Name: "食品", Icon: "mdi-food"},
		{Name: "饮料", Icon: "mdi-cup"},
		{Name: "水果零食", Icon: "mdi-food-apple"},
		{Name: "外卖", Icon: "mdi-noodles"},
	}},
	{Name: "服饰外貌", Icon: "mdi-tshirt-crew", Children: []presetNode{
		{Name: "衣服", Icon: "mdi-tshirt-crew"},
		{Name: "饰品", Icon: "mdi-watch"},
		{Name: "化妆品", Icon: "mdi-spray"},
		{Name: "美容美发", Icon: "mdi-scissors-cutting"},
	}},
	{Name: "居住", Icon: "mdi-home", Children: []presetNode{
		{Name: "家居用品", Icon: "mdi-sofa"},
		{Name: "电子产品", Icon: "mdi-laptop"},
		{Name: "维修保养", Icon: "mdi-tools"},
		{Name: "家政服务", Icon: "mdi-broom"},
		{Name: "水电煤气", Icon: "mdi-lightning-bolt"},
		{Name: "租金房贷", Icon: "mdi-home-city"},
	}},
	{Name: "交通", Icon: "mdi-bus", Children: []presetNode{
		{Name: "公共交通", Icon: "mdi-subway-variant"},
		{Name: "打车租车", Icon: "mdi-taxi"},
		{Name: "私家车", Icon: "mdi-car"},
		{Name: "火车票", Icon: "mdi-train"},
		{Name: "飞机票", Icon: "mdi-airplane"},
	}},
	{Name: "通讯", Icon: "mdi-cellphone", Children: []presetNode{
		{Name: "话费", Icon: "mdi-phone"},
		{Name: "网费", Icon: "mdi-wifi"},
		{Name: "快递费", Icon: "mdi-truck-delivery"},
	}},
	{Name: "购物", Icon: "mdi-cart", Children: []presetNode{
		{Name: "日用百货", Icon: "mdi-store"},
		{Name: "数码电器", Icon: "mdi-cellphone"},
		{Name: "其他购物", Icon: "mdi-shopping"},
	}},
	{Name: "娱乐", Icon: "mdi-movie", Children: []presetNode{
		{Name: "运动健身", Icon: "mdi-dumbbell"},
		{Name: "聚会", Icon: "mdi-account-group"},
		{Name: "电影演出", Icon: "mdi-movie"},
		{Name: "游戏玩具", Icon: "mdi-gamepad-variant"},
		{Name: "会员订阅", Icon: "mdi-star-circle"},
		{Name: "宠物", Icon: "mdi-dog"},
		{Name: "旅游", Icon: "mdi-beach"},
	}},
	{Name: "教育学习", Icon: "mdi-school", Children: []presetNode{
		{Name: "书报杂志", Icon: "mdi-book-open-page-variant"},
		{Name: "培训课程", Icon: "mdi-bookshelf"},
		{Name: "考试认证", Icon: "mdi-school"},
	}},
	{Name: "礼物捐赠", Icon: "mdi-gift", Children: []presetNode{
		{Name: "礼物", Icon: "mdi-gift"},
		{Name: "捐赠", Icon: "mdi-heart"},
	}},
	{Name: "医疗健康", Icon: "mdi-hospital", Children: []presetNode{
		{Name: "检查治疗", Icon: "mdi-hospital"},
		{Name: "药品", Icon: "mdi-pill"},
		{Name: "医疗器械", Icon: "mdi-needle"},
	}},
	{Name: "金融保险", Icon: "mdi-bank", Children: []presetNode{
		{Name: "税费", Icon: "mdi-receipt"},
		{Name: "手续费", Icon: "mdi-cash-minus"},
		{Name: "保险", Icon: "mdi-shield-check"},
		{Name: "利息支出", Icon: "mdi-percent"},
		{Name: "罚款赔偿", Icon: "mdi-alert-circle"},
	}},
	{Name: "其他支出", Icon: "mdi-dots-horizontal", Children: []presetNode{
		{Name: "其他", Icon: "mdi-dots-horizontal"},
	}},
}

var defaultIncomePresets = []presetNode{
	{Name: "职业收入", Icon: "mdi-cash-multiple", Children: []presetNode{
		{Name: "工资", Icon: "mdi-cash"},
		{Name: "奖金", Icon: "mdi-trophy"},
		{Name: "加班", Icon: "mdi-clock-outline"},
		{Name: "兼职", Icon: "mdi-handshake"},
	}},
	{Name: "金融投资", Icon: "mdi-chart-line", Children: []presetNode{
		{Name: "投资收入", Icon: "mdi-chart-line"},
		{Name: "租金收入", Icon: "mdi-home-city"},
		{Name: "利息收入", Icon: "mdi-piggy-bank"},
	}},
	{Name: "其他收入", Icon: "mdi-dots-horizontal", Children: []presetNode{
		{Name: "礼品红包", Icon: "mdi-gift"},
		{Name: "中奖", Icon: "mdi-dice-5"},
		{Name: "意外收入", Icon: "mdi-cash-plus"},
		{Name: "其他", Icon: "mdi-dots-horizontal"},
	}},
}

// EnsureDefaults 新用户写入完整二级预设；老用户按名称补齐缺失的一级/二级（不改已有结构）
func (s *CategoryService) EnsureDefaults(userID uint64) error {
	var count int64
	s.db.Model(&model.Category{}).Where("user_id = ?", userID).Count(&count)
	if count == 0 {
		return s.seedFullPresets(userID)
	}
	return s.supplementPresets(userID)
}

func (s *CategoryService) seedFullPresets(userID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := createPresetTree(tx, userID, "expense", defaultExpensePresets); err != nil {
			return err
		}
		return createPresetTree(tx, userID, "income", defaultIncomePresets)
	})
}

func createPresetTree(tx *gorm.DB, userID uint64, kind string, nodes []presetNode) error {
	for i, n := range nodes {
		parent := model.Category{
			UserID: userID, Name: n.Name, Kind: kind,
			Icon: n.Icon, Sort: i + 1,
		}
		if err := tx.Create(&parent).Error; err != nil {
			return err
		}
		for j, ch := range n.Children {
			pid := parent.ID
			child := model.Category{
				UserID: userID, Name: ch.Name, Kind: kind,
				Icon: ch.Icon, Sort: j + 1, ParentID: &pid,
			}
			if err := tx.Create(&child).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *CategoryService) supplementPresets(userID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := supplementKind(tx, userID, "expense", defaultExpensePresets); err != nil {
			return err
		}
		return supplementKind(tx, userID, "income", defaultIncomePresets)
	})
}

func supplementKind(tx *gorm.DB, userID uint64, kind string, presets []presetNode) error {
	var existing []model.Category
	if err := tx.Where("user_id = ? AND kind = ?", userID, kind).Find(&existing).Error; err != nil {
		return err
	}
	rootByName := map[string]model.Category{}
	childNamesUnder := map[uint64]map[string]struct{}{}
	rootNames := map[string]struct{}{}
	for _, c := range existing {
		if c.ParentID == nil {
			rootByName[c.Name] = c
			rootNames[c.Name] = struct{}{}
		} else {
			m := childNamesUnder[*c.ParentID]
			if m == nil {
				m = map[string]struct{}{}
				childNamesUnder[*c.ParentID] = m
			}
			m[c.Name] = struct{}{}
		}
	}

	nextSort := len(rootByName) + 1
	for _, p := range presets {
		parent, ok := rootByName[p.Name]
		if !ok {
			// 子名已作为一级存在时跳过整棵树，避免「工资」一级与「职业收入/工资」重复
			conflict := false
			for _, ch := range p.Children {
				if _, exists := rootNames[ch.Name]; exists {
					conflict = true
					break
				}
			}
			if conflict {
				continue
			}
			parent = model.Category{
				UserID: userID, Name: p.Name, Kind: kind,
				Icon: p.Icon, Sort: nextSort,
			}
			if err := tx.Create(&parent).Error; err != nil {
				return err
			}
			nextSort++
			rootByName[p.Name] = parent
			rootNames[p.Name] = struct{}{}
			childNamesUnder[parent.ID] = map[string]struct{}{}
		}
		have := childNamesUnder[parent.ID]
		if have == nil {
			have = map[string]struct{}{}
			childNamesUnder[parent.ID] = have
		}
		for j, ch := range p.Children {
			if _, exists := have[ch.Name]; exists {
				continue
			}
			pid := parent.ID
			child := model.Category{
				UserID: userID, Name: ch.Name, Kind: kind,
				Icon: ch.Icon, Sort: j + 1, ParentID: &pid,
			}
			if err := tx.Create(&child).Error; err != nil {
				return err
			}
			have[ch.Name] = struct{}{}
		}
	}
	return nil
}
