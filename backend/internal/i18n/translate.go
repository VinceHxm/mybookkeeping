package i18n

import (
	"regexp"
	"strings"
)

// entry 一条中文原文的各语言译文；原文含 %d / %s 时按位置替换
type entry struct {
	en, ja, ko string
}

type pattern struct {
	re   *regexp.Regexp
	text entry
}

var (
	exact    = map[string]entry{}
	patterns []pattern
	// conflicts 多个文件登记了同一原文但译文不同；由测试保证为空
	conflicts []string
)

var placeholder = regexp.MustCompile(`%[ds]`)

func register(table map[string]entry) {
	for zh, e := range table {
		if !placeholder.MatchString(zh) {
			if old, ok := exact[zh]; ok {
				if old != e {
					conflicts = append(conflicts, zh)
				}
				continue
			}
			exact[zh] = e
			continue
		}
		parts := placeholder.Split(zh, -1)
		for i := range parts {
			parts[i] = regexp.QuoteMeta(parts[i])
		}
		patterns = append(patterns, pattern{
			re:   regexp.MustCompile("^" + strings.Join(parts, "(.+?)") + "$"),
			text: e,
		})
	}
}

func pick(e entry, locale string) string {
	switch locale {
	case En:
		return e.en
	case Ja:
		return e.ja
	case Ko:
		return e.ko
	}
	return ""
}

// T 把中文原文翻成目标语言；中文或查不到时原样返回
func T(locale, zh string) string {
	if locale == ZhCN || locale == "" || zh == "" {
		return zh
	}
	if e, ok := exact[zh]; ok {
		if s := pick(e, locale); s != "" {
			return s
		}
		return zh
	}
	for _, p := range patterns {
		m := p.re.FindStringSubmatch(zh)
		if m == nil {
			continue
		}
		s := pick(p.text, locale)
		if s == "" {
			return zh
		}
		args := m[1:]
		i := 0
		return placeholder.ReplaceAllStringFunc(s, func(string) string {
			if i >= len(args) {
				return ""
			}
			v := args[i]
			i++
			return v
		})
	}
	return zh
}
