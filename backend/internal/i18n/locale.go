// Package i18n 后端多语言：service 层错误仍用中文原文，响应出口按 Accept-Language 查表翻译。
package i18n

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ZhCN = "zh-CN"
	En   = "en"
	Ja   = "ja"
	Ko   = "ko"

	Default = ZhCN
)

var Supported = []string{ZhCN, En, Ja, Ko}

// Normalize 把 zh / zh-Hans / en-US / ja-JP 等归一到支持的语言；繁体与不支持的返回空串
func Normalize(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, "_", "-")
	switch {
	case s == "zh" || s == "zh-cn" || s == "zh-sg" || strings.HasPrefix(s, "zh-hans"):
		return ZhCN
	case strings.HasPrefix(s, "zh"):
		return ""
	case strings.HasPrefix(s, "en"):
		return En
	case strings.HasPrefix(s, "ja"):
		return Ja
	case strings.HasPrefix(s, "ko"):
		return Ko
	}
	return ""
}

// ParseAcceptLanguage 按出现顺序取第一个支持的语言（前端只发单个值，q 权重不做排序）
func ParseAcceptLanguage(header string) string {
	for _, part := range strings.Split(header, ",") {
		tag := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if l := Normalize(tag); l != "" {
			return l
		}
	}
	return Default
}

// FromRequest 当前请求的语言
func FromRequest(c *gin.Context) string {
	return ParseAcceptLanguage(c.GetHeader("Accept-Language"))
}
