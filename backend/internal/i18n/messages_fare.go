package i18n

// 计费规则 / 节假日 / 模板 / 周期记账（service/fare_rule.go、holiday.go、template.go、schedule.go）
// 「X 折」= 按原价的 X/10 计；英韩译作百分比（%s0%），日文用「X掛け」。

func init() {
	// 报错
	register(map[string]entry{
		"规则名称不能为空":  {"Rule name is required", "ルール名を入力してください", "규칙 이름을 입력하세요"},
		"模板名称不能为空":  {"Template name is required", "テンプレート名を入力してください", "템플릿 이름을 입력하세요"},
		"名称不能为空":    {"Name is required", "名前を入力してください", "이름을 입력하세요"},
		"金额必须大于 0":  {"Amount must be greater than 0", "金額は 0 より大きくしてください", "금액은 0보다 커야 합니다"},
		"请选择账户":     {"Please select an account", "口座を選択してください", "계좌를 선택하세요"},
		"频率无效":      {"Invalid frequency", "繰り返し頻度が正しくありません", "반복 주기가 올바르지 않습니다"},
		"间隔天数至少为 1": {"Interval must be at least 1 day", "間隔は 1 日以上にしてください", "간격은 1일 이상이어야 합니다"},
		"间隔过大":      {"Interval is too large", "間隔が大きすぎます", "간격이 너무 큽니다"},
		"执行日无效":     {"Invalid run day", "実行日が正しくありません", "실행일이 올바르지 않습니다"},
		"年份无效":      {"Invalid year", "年が正しくありません", "연도가 올바르지 않습니다"},
		"数据库未就绪":    {"Database is not ready", "データベースの準備ができていません", "데이터베이스가 준비되지 않았습니다"},
		"%d 年暂无节假日数据（可能尚未公布）": {"No holiday data for %d yet (it may not be published)", "%d 年の祝日データはまだありません（未発表の可能性があります）", "%d년 공휴일 데이터가 아직 없습니다 (아직 발표되지 않았을 수 있습니다)"},
		"节假日接口 HTTP %d":       {"Holiday API returned HTTP %d", "祝日 API が HTTP %d を返しました", "공휴일 API가 HTTP %d를 반환했습니다"},
		"节假日接口返回 code=%d":     {"Holiday API returned code=%d", "祝日 API が code=%d を返しました", "공휴일 API가 code=%d를 반환했습니다"},
	})

	// 计费预览说明（handler 按「 · 」拆段后逐段翻译）
	register(map[string]entry{
		"未启用计费规则":  {"Fare rule is not enabled", "料金ルールが無効です", "요금 규칙이 비활성화되어 있습니다"},
		"规则不在生效期内": {"Rule is outside its valid period", "ルールの有効期間外です", "규칙 유효 기간이 아닙니다"},
		"免费时段":     {"Free day", "無料日", "무료일"},
		"时段免费":     {"Free time window", "無料の時間帯", "무료 시간대"},
		"法定节假日":    {"Public holiday", "祝日", "공휴일"},
		"法定节假日免费":  {"Free on public holiday", "祝日は無料", "공휴일 무료"},
		"就低":       {"Lowest price", "最安値を適用", "최저가 적용"},
		"次数档更优":    {"Ride count tier is cheaper", "乗車回数の段階の方が安い", "탑승 횟수 구간이 더 저렴"},
		"模板未绑定计费规则，使用固定金额": {"No fare rule linked to this template; using its fixed amount", "テンプレートに料金ルールが設定されていないため、固定金額を使います", "템플릿에 요금 규칙이 연결되지 않아 고정 금액을 사용합니다"},
		"原价¥%s":         {"Base fare ¥%s", "通常運賃 ¥%s", "기본 요금 ¥%s"},
		"票卡%s折":         {"Card fare %s0%", "カード運賃 %s掛け", "카드 요금 %s0%"},
		"全价%s折":         {"Full fare %s0%", "全額 %s掛け", "정가 %s0%"},
		"累计超¥%s全价%s折":   {"Over ¥%s in total: full fare %s0%", "累計 ¥%s 超：全額 %s掛け", "누적 ¥%s 초과: 정가 %s0%"},
		"累计超¥%s恢复票卡%s折": {"Over ¥%s in total: back to card fare %s0%", "累計 ¥%s 超：カード運賃 %s掛けに戻る", "누적 ¥%s 초과: 카드 요금 %s0%로 복귀"},
		"累计超¥%s票卡后再%s折": {"Over ¥%s in total: card fare, then %s0%", "累計 ¥%s 超：カード運賃からさらに %s掛け", "누적 ¥%s 초과: 카드 요금에서 추가 %s0%"},
		"第%d次免费":        {"Ride #%d is free", "%d 回目は無料", "%d번째 탑승 무료"},
		"第%d次全价%s折":     {"Ride #%d: full fare %s0%", "%d 回目：全額 %s掛け", "%d번째 탑승: 정가 %s0%"},
		"第%d次票卡后再%s折":   {"Ride #%d: card fare, then %s0%", "%d 回目：カード運賃からさらに %s掛け", "%d번째 탑승: 카드 요금에서 추가 %s0%"},
		"立减¥%s":         {"Flat discount ¥%s", "値引き ¥%s", "즉시 할인 ¥%s"},
		"再叠时段%s折":       {"Plus time window %s0%", "さらに時間帯 %s掛け", "시간대 %s0% 추가 적용"},
	})
}
