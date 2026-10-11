package i18n

// 账户 / 信用卡账单 / 理财市值校准 / 归属校验
func init() {
	register(map[string]entry{
		"账户名称不能为空":        {"Account name is required", "口座名を入力してください", "계좌 이름을 입력하세요"},
		"分期期数无效":          {"Invalid number of installment periods", "分割払いの回数が正しくありません", "할부 회차가 올바르지 않습니다"},
		"分期本金无效":          {"Invalid installment principal", "分割払いの元金が正しくありません", "할부 원금이 올바르지 않습니다"},
		"分期利息无效":          {"Invalid installment interest", "分割払いの利息が正しくありません", "할부 이자가 올바르지 않습니다"},
		"请填写首期出账日":        {"Please enter the first statement date", "初回の請求日を入力してください", "첫 회차 청구일을 입력하세요"},
		"首期出账日格式无效":       {"Invalid first statement date format", "初回の請求日の形式が正しくありません", "첫 회차 청구일 형식이 올바르지 않습니다"},
		"账户下仍有流水，请先归档或迁移": {"This account still has records. Archive it or move the records first", "この口座にはまだ記録があります。先にアーカイブするか記録を移してください", "이 계좌에 아직 내역이 있습니다. 먼저 보관하거나 내역을 옮기세요"},
		"仅信用卡账户支持账单展示":    {"Statements are only available for credit card accounts", "明細はクレジットカード口座でのみ表示できます", "명세서는 신용카드 계좌에서만 볼 수 있습니다"},
		"请先设置账单日（1-28）":   {"Please set a statement date (1-28) first", "先に締め日（1-28）を設定してください", "먼저 명세서 마감일(1-28)을 설정하세요"},
		"转入账户不存在":         {"Destination account not found", "振替先の口座が見つかりません", "입금 계좌를 찾을 수 없습니다"},
		"分类不存在":           {"Category not found", "カテゴリが見つかりません", "분류를 찾을 수 없습니다"},
		"计费规则不存在":         {"Fare rule not found", "料金ルールが見つかりません", "요금 규칙을 찾을 수 없습니다"},
		"市值不能为负":          {"Market value cannot be negative", "評価額はマイナスにできません", "평가액은 음수일 수 없습니다"},
		"只有理财账户可以校准市值":    {"Only investment accounts can be revalued", "評価額調整は投資口座でのみ行えます", "투자 계좌만 평가액을 조정할 수 있습니다"},
		"请选择账户":           {"Please select an account", "口座を選択してください", "계좌를 선택하세요"},
		"校准差额不能为 0":       {"Revaluation difference cannot be 0", "調整差額を 0 にはできません", "조정 차액은 0일 수 없습니다"},
		"市值校准不能设置转入账户":    {"A revaluation cannot have a destination account", "評価額調整には振替先の口座を設定できません", "평가액 조정에는 입금 계좌를 설정할 수 없습니다"},
		"市值校准不能设置分类":      {"A revaluation cannot have a category", "評価額調整にはカテゴリを設定できません", "평가액 조정에는 분류를 설정할 수 없습니다"},
		"市值校准只能记在理财账户":    {"Revaluations can only be recorded in investment accounts", "評価額調整は投資口座にのみ記録できます", "평가액 조정은 투자 계좌에만 기록할 수 있습니다"},
	})

	// 账单 / 流水页的兜底展示文案
	register(map[string]entry{
		"存量分期": {"Existing installment", "既存の分割払い", "기존 할부"},
		"分期消费": {"Installment purchase", "分割払いの購入", "할부 결제"},
		"状态按「先出账先还」分摊还款与退款估算：还款优先抵扣最早出账的消费。": {"Statuses are estimated by allocating repayments and refunds to the earliest billed spending first.", "状態は「先に請求されたものから返済」として返済と返金を按分した推定です。", "상태는 '먼저 청구된 것부터 상환' 방식으로 상환과 환불을 배분한 추정치입니다."},
		"未设置账单日，无法区分已出账 / 未出账。":              {"No statement date set, so billed and unbilled amounts can't be told apart.", "締め日が未設定のため、請求済み / 未請求を区別できません。", "명세서 마감일이 설정되지 않아 청구 / 미청구를 구분할 수 없습니다."},
	})
}
