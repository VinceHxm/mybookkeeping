package i18n

func init() {
	// 流水 / 团购券
	register(map[string]entry{
		"市值校准请在理财账户里操作":   {"Revaluations must be made in an investment account", "評価額調整は投資口座で行ってください", "평가액 조정은 투자 계좌에서 하세요"},
		"市值校准流水不能与其他类型互改": {"Revaluation records cannot be changed to or from other types", "評価額調整の記録は他の種類と相互に変更できません", "평가액 조정 내역은 다른 유형과 서로 바꿀 수 없습니다"},
		"金额无效":        {"Invalid amount", "金額が正しくありません", "금액이 올바르지 않습니다"},
		"利息无效":        {"Invalid interest", "利息が正しくありません", "이자가 올바르지 않습니다"},
		"分期首期出账设置无效":  {"Invalid first-installment billing setting", "分割払いの初回請求の設定が正しくありません", "할부 첫 회 청구 설정이 올바르지 않습니다"},
		"请选择分类":       {"Please select a category", "カテゴリを選択してください", "분류를 선택하세요"},
		"请选择转入账户":     {"Please select the destination account", "入金口座を選択してください", "입금 계좌를 선택하세요"},
		"转出与转入账户不能相同": {"Source and destination accounts cannot be the same", "出金口座と入金口座を同じにすることはできません", "출금 계좌와 입금 계좌는 같을 수 없습니다"},
		"类型无效":        {"Invalid type", "種類が正しくありません", "유형이 올바르지 않습니다"},
		"外币币种无效":      {"Invalid foreign currency", "外貨の通貨が正しくありません", "외화 통화가 올바르지 않습니다"},
		"外币币种不能是人民币":  {"Foreign currency cannot be CNY", "外貨に人民元は指定できません", "외화 통화는 위안화일 수 없습니다"},
		"请填写外币金额":     {"Please enter the foreign amount", "外貨金額を入力してください", "외화 금액을 입력하세요"},
		"券状态无效":       {"Invalid voucher status", "クーポンの状態が正しくありません", "쿠폰 상태가 올바르지 않습니다"},
		"仅支出可标记团购券":   {"Only expenses can be marked as vouchers", "クーポンとして登録できるのは支出のみです", "지출만 쿠폰으로 표시할 수 있습니다"},
		"该券已退款，删除退款流水后才能改状态":  {"This voucher has been refunded; delete the refund record before changing its status", "このクーポンは返金済みです。状態を変更するには返金の記録を削除してください", "이 쿠폰은 환불되었습니다. 상태를 바꾸려면 환불 내역을 먼저 삭제하세요"},
		"只有待核销 / 已核销的团购券可以退款": {"Only unused / used vouchers can be refunded", "返金できるのは未使用 / 使用済みのクーポンのみです", "미사용 / 사용 완료 쿠폰만 환불할 수 있습니다"},
		"退款金额不能超过原支付金额":       {"Refund amount cannot exceed the amount paid", "返金額は支払額を超えられません", "환불 금액은 결제 금액을 초과할 수 없습니다"},
		// 退款流水的默认备注前缀（非报错，写入流水时按用户语言取）
		"团购券退款": {"Voucher refund", "クーポン返金", "쿠폰 환불"},
	})

	// 附件
	register(map[string]entry{
		"文件过大（上限 10MB）":                             {"File too large (limit 10MB)", "ファイルが大きすぎます（上限 10MB）", "파일이 너무 큽니다 (최대 10MB)"},
		"仅支持 JPG / PNG / GIF / WEBP / HEIC 图片或 PDF": {"Only JPG / PNG / GIF / WEBP / HEIC images or PDF are supported", "対応形式は JPG / PNG / GIF / WEBP / HEIC 画像または PDF のみです", "JPG / PNG / GIF / WEBP / HEIC 이미지 또는 PDF만 지원합니다"},
		"附件不存在或链接已过期":                               {"Attachment not found or link expired", "添付ファイルが見つからないか、リンクの有効期限が切れています", "첨부파일이 없거나 링크가 만료되었습니다"},
	})

	// AI 识别
	register(map[string]entry{
		"未配置 DEEPSEEK_API_KEY，AI 识别不可用": {"DEEPSEEK_API_KEY is not configured; AI recognition is unavailable", "DEEPSEEK_API_KEY が設定されていないため、AI 認識は利用できません", "DEEPSEEK_API_KEY가 설정되지 않아 AI 인식을 사용할 수 없습니다"},
		"请输入要识别的文字":                     {"Please enter the text to recognize", "認識するテキストを入力してください", "인식할 텍스트를 입력하세요"},
		"请上传图片":                         {"Please upload an image", "画像をアップロードしてください", "이미지를 업로드하세요"},
		"仅支持 JPEG / PNG / GIF / WebP":   {"Only JPEG / PNG / GIF / WebP are supported", "対応形式は JPEG / PNG / GIF / WebP のみです", "JPEG / PNG / GIF / WebP만 지원합니다"},
		"模型返回无法解析: %s":                  {"Could not parse the model response: %s", "モデルの応答を解析できません: %s", "모델 응답을 해석할 수 없습니다: %s"},
		"LLM 响应异常: %s":                  {"Unexpected LLM response: %s", "LLM の応答が異常です: %s", "LLM 응답이 비정상입니다: %s"},
		"LLM 无返回内容":                     {"LLM returned no content", "LLM から応答がありませんでした", "LLM이 내용을 반환하지 않았습니다"},
	})
}
