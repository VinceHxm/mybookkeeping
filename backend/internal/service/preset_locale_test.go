package service

import (
	"testing"

	"mybookkeeping/internal/i18n"
)

func TestPresetNamesTranslated(t *testing.T) {
	var names []string
	var walk func([]presetNode)
	walk = func(nodes []presetNode) {
		for _, n := range nodes {
			names = append(names, n.Name)
			walk(n.Children)
		}
	}
	walk(defaultExpensePresets)
	walk(defaultIncomePresets)
	for _, p := range defaultTagPresets {
		names = append(names, p.Name)
	}
	// 日文可能与中文同形（衣服、交通），只用英文判断是否已收录；三语是否齐全由 i18n 包测试保证
	for _, name := range names {
		if i18n.Preset(i18n.En, name) == name {
			t.Errorf("预设 %q 未收录译名", name)
		}
	}
}
