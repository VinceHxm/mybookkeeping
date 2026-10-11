package i18n

import "testing"

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"zh": ZhCN, "zh-CN": ZhCN, "zh_Hans": ZhCN, "zh-Hans-CN": ZhCN,
		"zh-TW": "", "zh-Hant": "",
		"en": En, "en-US": En, "ja-JP": Ja, "ko-KR": Ko,
		"fr": "", "": "",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	cases := map[string]string{
		"":                        ZhCN,
		"ja":                      Ja,
		"fr-FR,fr;q=0.9,en;q=0.8": En,
		"zh-TW,ko;q=0.9":          Ko,
		"de":                      ZhCN,
	}
	for in, want := range cases {
		if got := ParseAcceptLanguage(in); got != want {
			t.Errorf("ParseAcceptLanguage(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestT(t *testing.T) {
	if got := T(En, "参数错误"); got != "Invalid parameters" {
		t.Errorf("exact: %q", got)
	}
	if got := T(ZhCN, "参数错误"); got != "参数错误" {
		t.Errorf("zh passthrough: %q", got)
	}
	if got := T(Ja, "密码至少 8 位"); got != "パスワードは 8 文字以上にしてください" {
		t.Errorf("pattern: %q", got)
	}
	if got := T(Ko, "尝试次数过多，请 15 分钟后再试"); got != "시도 횟수가 너무 많습니다. 15분 후 다시 시도하세요" {
		t.Errorf("pattern: %q", got)
	}
	if got := T(En, "还没翻译的报错"); got != "还没翻译的报错" {
		t.Errorf("fallback: %q", got)
	}
}

func TestTablesComplete(t *testing.T) {
	check := func(kind, zh string, e entry) {
		if e.en == "" || e.ja == "" || e.ko == "" {
			t.Errorf("%s %q 缺少译文: %+v", kind, zh, e)
		}
	}
	for zh, e := range exact {
		check("message", zh, e)
	}
	for _, p := range patterns {
		check("pattern", p.re.String(), p.text)
	}
	for zh, e := range presetNames {
		check("preset", zh, e)
	}
}

func TestNoConflictingMessages(t *testing.T) {
	for _, zh := range conflicts {
		t.Errorf("%q 在多个文件登记且译文不一致", zh)
	}
}

func TestPreset(t *testing.T) {
	if got := Preset(Ja, "餐饮"); got != "食費" {
		t.Errorf("ja: %q", got)
	}
	if got := Preset(ZhCN, "餐饮"); got != "餐饮" {
		t.Errorf("zh: %q", got)
	}
	if got := Preset(En, "没有的分类"); got != "没有的分类" {
		t.Errorf("fallback: %q", got)
	}
}
