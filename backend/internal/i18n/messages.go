package i18n

// 键为代码里的中文原文，改原文时须同步改这里，否则会回退成中文。
// 新增模块的报错按模块追加一个 register 块即可。

func init() {
	// 通用 / handler 层
	register(map[string]entry{
		"参数错误": {"Invalid parameters", "パラメータが正しくありません", "잘못된 요청입니다"},
		"服务器内部错误，请稍后重试":              {"Internal server error. Please try again later", "サーバー内部エラーです。しばらくしてから再試行してください", "서버 내부 오류입니다. 잠시 후 다시 시도하세요"},
		"时间格式无效":                     {"Invalid time format", "日時の形式が正しくありません", "시간 형식이 올바르지 않습니다"},
		"流水不存在":                      {"Record not found", "記録が見つかりません", "내역을 찾을 수 없습니다"},
		"账户不存在":                      {"Account not found", "口座が見つかりません", "계좌를 찾을 수 없습니다"},
		"请填写当前市值":                    {"Please enter the current market value", "現在の評価額を入力してください", "현재 평가액을 입력하세요"},
		"未配置 DeepSeek API Key":       {"DeepSeek API key is not configured", "DeepSeek API キーが設定されていません", "DeepSeek API 키가 설정되지 않았습니다"},
		"请上传图片（字段名 file，上限约 12MB）":   {"Please upload an image (field \"file\", up to about 12MB)", "画像をアップロードしてください（フィールド名 file、上限約 12MB）", "이미지를 업로드하세요 (필드명 file, 최대 약 12MB)"},
		"图片过大（上限约 12MB）":             {"Image too large (limit about 12MB)", "画像が大きすぎます（上限約 12MB）", "이미지가 너무 큽니다 (최대 약 12MB)"},
		"附件服务未就绪（请启动 MinIO）":         {"Attachment service is not ready (start MinIO)", "添付ファイルサービスが利用できません（MinIO を起動してください）", "첨부파일 서비스가 준비되지 않았습니다 (MinIO를 시작하세요)"},
		"请上传文件（上限 10MB）":             {"Please upload a file (up to 10MB)", "ファイルをアップロードしてください（上限 10MB）", "파일을 업로드하세요 (최대 10MB)"},
		"请指定要修改的字段（role / disabled）": {"Specify a field to change (role / disabled)", "変更する項目を指定してください（role / disabled）", "변경할 항목을 지정하세요 (role / disabled)"},
		"节假日服务未启用":                   {"Holiday service is not enabled", "祝日サービスが有効になっていません", "공휴일 서비스가 활성화되지 않았습니다"},
		"不支持的语言":                     {"Unsupported language", "サポートされていない言語です", "지원하지 않는 언어입니다"},
	})

	// 找回密码结果提示
	register(map[string]entry{
		"未开放自助找回密码，请联系管理员在「用户管理」中为你重置": {
			"Self-service password recovery is disabled. Ask an administrator to reset it in User Management",
			"セルフサービスのパスワード再設定は無効です。管理者に「ユーザー管理」でリセットを依頼してください",
			"셀프 비밀번호 찾기가 비활성화되어 있습니다. 관리자에게 '사용자 관리'에서 재설정을 요청하세요",
		},
		"若用户存在，已生成重置令牌": {
			"If the user exists, a reset token has been generated",
			"ユーザーが存在する場合、リセットトークンを生成しました",
			"사용자가 존재하면 재설정 토큰이 생성되었습니다",
		},
		"已生成重置令牌（个人部署默认直接返回，请尽快使用）": {
			"Reset token generated (returned directly on personal deployments; use it soon)",
			"リセットトークンを生成しました（個人環境では直接返されます。早めに使用してください）",
			"재설정 토큰이 생성되었습니다(개인 배포에서는 바로 반환됩니다. 빨리 사용하세요)",
		},
	})

	// 登录态 / 中间件
	register(map[string]entry{
		"未登录":     {"Not signed in", "ログインしていません", "로그인하지 않았습니다"},
		"登录已过期":   {"Session expired", "ログインの有効期限が切れました", "로그인이 만료되었습니다"},
		"登录已失效":   {"Session is no longer valid", "ログインが無効になりました", "로그인이 더 이상 유효하지 않습니다"},
		"会话校验失败":  {"Failed to verify session", "セッションの確認に失敗しました", "세션 확인에 실패했습니다"},
		"需要管理员权限": {"Administrator permission required", "管理者権限が必要です", "관리자 권한이 필요합니다"},
	})

	// 账号 / 认证
	register(map[string]entry{
		"请输入用户名和密码":             {"Please enter username and password", "ユーザー名とパスワードを入力してください", "아이디와 비밀번호를 입력하세요"},
		"请输入用户名":                {"Please enter a username", "ユーザー名を入力してください", "아이디를 입력하세요"},
		"请输入登录密码":               {"Please enter your password", "ログインパスワードを入力してください", "로그인 비밀번호를 입력하세요"},
		"请提供令牌和新密码":             {"Please provide the token and a new password", "トークンと新しいパスワードを入力してください", "토큰과 새 비밀번호를 입력하세요"},
		"当前未开放注册":               {"Sign-up is currently closed", "現在、新規登録は受け付けていません", "현재 회원가입이 닫혀 있습니다"},
		"用户名或密码错误":              {"Incorrect username or password", "ユーザー名またはパスワードが正しくありません", "아이디 또는 비밀번호가 올바르지 않습니다"},
		"密码错误":                  {"Incorrect password", "パスワードが正しくありません", "비밀번호가 올바르지 않습니다"},
		"原密码错误":                 {"Current password is incorrect", "現在のパスワードが正しくありません", "현재 비밀번호가 올바르지 않습니다"},
		"用户名已存在":                {"Username already exists", "このユーザー名は既に使われています", "이미 존재하는 아이디입니다"},
		"用户名不能为空":               {"Username is required", "ユーザー名を入力してください", "아이디를 입력하세요"},
		"用户不存在":                 {"User not found", "ユーザーが見つかりません", "사용자를 찾을 수 없습니다"},
		"密码至少 %d 位":             {"Password must be at least %d characters", "パスワードは %d 文字以上にしてください", "비밀번호는 %d자 이상이어야 합니다"},
		"重置令牌无效或已过期":            {"Reset token is invalid or expired", "リセットトークンが無効か期限切れです", "재설정 토큰이 유효하지 않거나 만료되었습니다"},
		"账户已停用，请联系管理员":          {"Account disabled. Please contact the administrator", "アカウントは停止されています。管理者に連絡してください", "계정이 정지되었습니다. 관리자에게 문의하세요"},
		"尝试次数过多，请 %d 分钟后再试":     {"Too many attempts. Try again in %d minute(s)", "試行回数が多すぎます。%d 分後に再試行してください", "시도 횟수가 너무 많습니다. %d분 후 다시 시도하세요"},
		"不能对自己执行此操作":            {"You cannot perform this action on yourself", "自分自身には実行できません", "자기 자신에게는 실행할 수 없습니다"},
		"不能停用或降级最后一个管理员":        {"Cannot disable or demote the last administrator", "最後の管理者を停止・降格することはできません", "마지막 관리자는 정지하거나 강등할 수 없습니다"},
		"角色无效，仅支持 user / admin": {"Invalid role; only user / admin are supported", "無効な権限です（user / admin のみ）", "잘못된 권한입니다 (user / admin만 지원)"},
		"请在「设置」中修改自己的密码":        {"Change your own password in Settings", "自分のパスワードは「設定」で変更してください", "본인 비밀번호는 「설정」에서 변경하세요"},
		"默认还款账户不存在":             {"Default repayment account not found", "既定の返済口座が見つかりません", "기본 상환 계좌를 찾을 수 없습니다"},
		"默认还款账户已归档":             {"Default repayment account is archived", "既定の返済口座はアーカイブ済みです", "기본 상환 계좌가 보관 처리되었습니다"},
		"默认还款账户不能是信用账户":         {"Default repayment account cannot be a credit account", "既定の返済口座にクレジット口座は指定できません", "기본 상환 계좌는 신용 계좌일 수 없습니다"},
		"默认消费账户不存在":             {"Default spending account not found", "既定の支払い口座が見つかりません", "기본 결제 계좌를 찾을 수 없습니다"},
		"默认消费账户已归档":             {"Default spending account is archived", "既定の支払い口座はアーカイブ済みです", "기본 결제 계좌가 보관 처리되었습니다"},
	})
}
