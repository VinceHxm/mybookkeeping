package service

import (
	"gorm.io/gorm"

	"mybookkeeping/internal/i18n"
	"mybookkeeping/internal/model"
)

// EnsureDefaults 维护默认分类，locale 为用户当前界面语言：
//   - 空库：按预设写入完整二级树（四语名入库，Name 取当前语言）；
//   - 老数据：名字仍等于某个预设名（任一语言）的分类回填 PresetKey 与四语名；
//   - 未改名的默认分类：Name 切到当前语言；
//   - 按 PresetKey 补齐缺失的一级 / 二级（改过名的仍按 PresetKey 识别，不会重复补）。
func (s *CategoryService) EnsureDefaults(userID uint64, locale string) error {
	if i18n.Normalize(locale) == "" {
		locale = i18n.ZhCN
	}
	var count int64
	s.db.Model(&model.Category{}).Where("user_id = ?", userID).Count(&count)
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, k := range []struct {
			kind    string
			presets []presetNode
		}{{"expense", defaultExpensePresets}, {"income", defaultIncomePresets}} {
			if count == 0 {
				if err := createPresetTree(tx, userID, k.kind, k.presets, locale); err != nil {
					return err
				}
				continue
			}
			if err := syncPresetKind(tx, userID, k.kind, k.presets, locale); err != nil {
				return err
			}
		}
		return nil
	})
}

func newPresetCategory(userID uint64, kind string, n presetNode, sort int, parentID *uint64, locale string) model.Category {
	c := model.Category{UserID: userID, Kind: kind, Icon: n.Icon, Sort: sort, ParentID: parentID}
	fillPreset(&c, n)
	c.Name = localizedName(&c, locale)
	return c
}

func fillPreset(c *model.Category, n presetNode) {
	c.PresetKey = n.Name
	c.NameZh = n.Name
	c.NameEn = i18n.Preset(i18n.En, n.Name)
	c.NameJa = i18n.Preset(i18n.Ja, n.Name)
	c.NameKo = i18n.Preset(i18n.Ko, n.Name)
}

func localizedName(c *model.Category, locale string) string {
	switch locale {
	case i18n.En:
		return c.NameEn
	case i18n.Ja:
		return c.NameJa
	case i18n.Ko:
		return c.NameKo
	}
	return c.NameZh
}

// matchPreset 按任一语言的预设名匹配（兼容老数据与按外语生成的数据）
func matchPreset(nodes []presetNode, name string) (presetNode, bool) {
	for _, n := range nodes {
		if name == n.Name {
			return n, true
		}
		for _, l := range []string{i18n.En, i18n.Ja, i18n.Ko} {
			if name == i18n.Preset(l, n.Name) {
				return n, true
			}
		}
	}
	return presetNode{}, false
}

func findPreset(nodes []presetNode, key string) (presetNode, bool) {
	for _, n := range nodes {
		if n.Name == key {
			return n, true
		}
	}
	return presetNode{}, false
}

func createPresetTree(tx *gorm.DB, userID uint64, kind string, nodes []presetNode, locale string) error {
	for i, n := range nodes {
		parent := newPresetCategory(userID, kind, n, i+1, nil, locale)
		if err := tx.Create(&parent).Error; err != nil {
			return err
		}
		for j, ch := range n.Children {
			pid := parent.ID
			child := newPresetCategory(userID, kind, ch, j+1, &pid, locale)
			if err := tx.Create(&child).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func syncPresetKind(tx *gorm.DB, userID uint64, kind string, presets []presetNode, locale string) error {
	var existing []model.Category
	if err := tx.Where("user_id = ? AND kind = ?", userID, kind).Order("sort asc, id asc").Find(&existing).Error; err != nil {
		return err
	}
	byID := map[uint64]*model.Category{}
	for i := range existing {
		byID[existing[i].ID] = &existing[i]
	}
	dirty := map[uint64]bool{}

	// 回填：先一级再二级，二级只在父级已识别时按父预设的子项匹配
	for i := range existing {
		c := &existing[i]
		if c.ParentID != nil || c.PresetKey != "" {
			continue
		}
		if n, ok := matchPreset(presets, c.Name); ok {
			fillPreset(c, n)
			dirty[c.ID] = true
		}
	}
	for i := range existing {
		c := &existing[i]
		if c.ParentID == nil || c.PresetKey != "" {
			continue
		}
		p := byID[*c.ParentID]
		if p == nil || p.PresetKey == "" {
			continue
		}
		pn, ok := findPreset(presets, p.PresetKey)
		if !ok {
			continue
		}
		if n, ok := matchPreset(pn.Children, c.Name); ok {
			fillPreset(c, n)
			dirty[c.ID] = true
		}
	}

	for i := range existing {
		c := &existing[i]
		if c.NameZh == "" {
			continue
		}
		if want := localizedName(c, locale); want != "" && c.Name != want {
			c.Name = want
			dirty[c.ID] = true
		}
	}
	for id := range dirty {
		c := byID[id]
		if err := tx.Model(c).Select("Name", "PresetKey", "NameZh", "NameEn", "NameJa", "NameKo").Updates(c).Error; err != nil {
			return err
		}
	}

	return supplementPresets(tx, userID, kind, presets, existing, locale)
}

// supplementPresets 按 PresetKey（自建分类按名字）补齐缺失的预设一级 / 二级
func supplementPresets(tx *gorm.DB, userID uint64, kind string, presets []presetNode, existing []model.Category, locale string) error {
	key := func(c model.Category) string {
		if c.PresetKey != "" {
			return c.PresetKey
		}
		return c.Name
	}
	rootByKey := map[string]model.Category{}
	childKeysUnder := map[uint64]map[string]struct{}{}
	for _, c := range existing {
		if c.ParentID == nil {
			rootByKey[key(c)] = c
			continue
		}
		m := childKeysUnder[*c.ParentID]
		if m == nil {
			m = map[string]struct{}{}
			childKeysUnder[*c.ParentID] = m
		}
		m[key(c)] = struct{}{}
	}

	nextSort := len(rootByKey) + 1
	for _, p := range presets {
		parent, ok := rootByKey[p.Name]
		if !ok {
			// 子名已作为一级存在时跳过整棵树，避免「工资」一级与「职业收入/工资」重复
			conflict := false
			for _, ch := range p.Children {
				if _, exists := rootByKey[ch.Name]; exists {
					conflict = true
					break
				}
			}
			if conflict {
				continue
			}
			parent = newPresetCategory(userID, kind, p, nextSort, nil, locale)
			if err := tx.Create(&parent).Error; err != nil {
				return err
			}
			nextSort++
			rootByKey[p.Name] = parent
		}
		have := childKeysUnder[parent.ID]
		if have == nil {
			have = map[string]struct{}{}
			childKeysUnder[parent.ID] = have
		}
		for j, ch := range p.Children {
			if _, exists := have[ch.Name]; exists {
				continue
			}
			pid := parent.ID
			child := newPresetCategory(userID, kind, ch, j+1, &pid, locale)
			if err := tx.Create(&child).Error; err != nil {
				return err
			}
			have[ch.Name] = struct{}{}
		}
	}
	return nil
}
