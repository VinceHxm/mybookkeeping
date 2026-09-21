package service

import (
	"strings"

	"mybookkeeping/internal/model"
)

// MaskCardNo 卡号脱敏：仅保留后 4 位，其余用 ·。
func MaskCardNo(raw string) string {
	s := strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
	if s == "" {
		return ""
	}
	if len(s) <= 4 {
		return s
	}
	hidden := len(s) - 4
	if hidden > 12 {
		hidden = 12
	}
	return strings.Repeat("·", hidden) + s[len(s)-4:]
}

// MaskHolderName 户名脱敏：保留首字，其余最多 3 个 *。
func MaskHolderName(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	runes := []rune(s)
	if len(runes) == 1 {
		return s
	}
	n := len(runes) - 1
	if n > 3 {
		n = 3
	}
	return string(runes[0]) + strings.Repeat("*", n)
}

// LooksMaskedSecret 判断是否为脱敏展示串（不可写回库）。
func LooksMaskedSecret(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	return strings.Contains(s, "·") ||
		strings.Contains(s, "•") ||
		strings.Contains(s, "*") ||
		strings.Contains(s, "＊")
}

// SanitizeSecretWrite 创建/更新时：脱敏串视为「未改动」返回 apply=false；空串可清空；明文可写。
func SanitizeSecretWrite(raw string) (value string, apply bool) {
	v := strings.TrimSpace(raw)
	if LooksMaskedSecret(v) {
		return "", false
	}
	return v, true
}

// MaskAccountInPlace 列表/普通详情：卡号、户名仅下发脱敏值；明文不离开服务端。
func MaskAccountInPlace(a *model.Account) {
	if a == nil {
		return
	}
	if a.CardNo != "" {
		a.CardNo = MaskCardNo(a.CardNo)
	}
	if a.HolderName != "" {
		a.HolderName = MaskHolderName(a.HolderName)
	}
}

func MaskAccountsInPlace(list []model.Account) {
	for i := range list {
		MaskAccountInPlace(&list[i])
	}
}

// FormatCardGroups 明文卡号按 4 位分组（仅敏感接口返回时用）。
func FormatCardGroups(raw string) string {
	s := strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
	if s == "" {
		return ""
	}
	var b strings.Builder
	for i, r := range s {
		if i > 0 && i%4 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}
