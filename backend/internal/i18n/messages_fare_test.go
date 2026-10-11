package i18n

import "testing"

func TestFareReasonParts(t *testing.T) {
	cases := []struct{ loc, zh, want string }{
		{En, "原价¥2.00", "Base fare ¥2.00"},
		{En, "票卡9折", "Card fare 90%"},
		{Ja, "票卡9折", "カード運賃 9掛け"},
		{En, "全价7折", "Full fare 70%"},
		{En, "累计超¥80.00全价7折", "Over ¥80.00 in total: full fare 70%"},
		{En, "累计超¥300.00恢复票卡9折", "Over ¥300.00 in total: back to card fare 90%"},
		{En, "累计超¥150.00票卡后再5折", "Over ¥150.00 in total: card fare, then 50%"},
		{Ko, "第40次免费", "40번째 탑승 무료"},
		{En, "第3次全价8折", "Ride #3: full fare 80%"},
		{En, "第3次票卡后再8折", "Ride #3: card fare, then 80%"},
		{En, "立减¥0.50", "Flat discount ¥0.50"},
		{En, "再叠时段7折", "Plus time window 70%"},
		{Ja, "%d 年暂无节假日数据（可能尚未公布）", "%d 年の祝日データはまだありません（未発表の可能性があります）"},
		{En, "2027 年暂无节假日数据（可能尚未公布）", "No holiday data for 2027 yet (it may not be published)"},
	}
	for i := 0; i < 20; i++ {
		for _, c := range cases {
			if got := T(c.loc, c.zh); got != c.want {
				t.Fatalf("T(%s, %q) = %q, want %q", c.loc, c.zh, got, c.want)
			}
		}
	}
}
