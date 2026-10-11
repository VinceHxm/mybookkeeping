package i18n

// 分类 / 标签（service/category.go、service/tag.go）；管理员相关报错已在 messages.go 登记

func init() {
	register(map[string]entry{
		"分类名称不能为空":                  {"Category name is required", "カテゴリ名を入力してください", "분류 이름을 입력하세요"},
		"kind 必须是 expense 或 income": {"kind must be expense or income", "kind は expense または income を指定してください", "kind는 expense 또는 income이어야 합니다"},
		"父分类不存在":                    {"Parent category not found", "大分類が見つかりません", "대분류를 찾을 수 없습니다"},
		"仅支持两级分类":                   {"Only two levels of categories are supported", "カテゴリは 2 階層までです", "분류는 2단계까지만 지원합니다"},
		"子分类类型须与父分类一致":              {"Sub category must have the same type as its parent", "小分類の種類は大分類と同じにしてください", "소분류 유형은 대분류와 같아야 합니다"},
		"不能将自己设为父分类":                {"A category cannot be its own parent", "自分自身を大分類に設定することはできません", "자기 자신을 대분류로 지정할 수 없습니다"},
		"已有子分类，不能降为二级":              {"This category has sub categories and cannot become a sub category", "小分類があるため、小分類に変更できません", "소분류가 있어 소분류로 변경할 수 없습니다"},
		"请先删除子分类":                   {"Delete its sub categories first", "先に小分類を削除してください", "먼저 소분류를 삭제하세요"},
		"分类下仍有流水，无法删除":              {"This category still has records and cannot be deleted", "このカテゴリには記録があるため削除できません", "이 분류에 내역이 있어 삭제할 수 없습니다"},
		"标签名称不能为空":                  {"Tag name is required", "タグ名を入力してください", "태그 이름을 입력하세요"},
		"标签不存在":                     {"Tag not found", "タグが見つかりません", "태그를 찾을 수 없습니다"},
	})
}
