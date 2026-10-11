package i18n

// 统计：分类 / 标签构成里的占位项名称（handler 返回前按请求语言翻译）
func init() {
	register(map[string]entry{
		"未分类":     {"Uncategorized", "未分類", "미분류"},
		"无标签":     {"No tag", "タグなし", "태그 없음"},
		"%s（未细分）": {"%s (no subcategory)", "%s（サブカテゴリなし）", "%s (하위 분류 없음)"},
	})
}
