import type { LocaleBundle } from '../types'

const zh = {
  cardOnly: { title: '仅刷卡折', hint: '例：每次按票卡 9 折计', name: '刷卡折扣' },
  weekendFree: { title: '周末免费', hint: '例：每周六、日免费', name: '周末免费' },
  monthTier: { title: '月累计阶梯', hint: '例：超 80 元 7 折、超 150 元 5 折、超 300 回票卡', name: '月累计优惠' },
  morning: { title: '早间折扣', hint: '例：工作日 7 点前全价 7 折', name: '早间优惠', window: '早间' },
  nthFree: { title: '第 N 次免费', hint: '例：本月第 40 次乘车免费', name: '乘次奖励' },
}

const en: typeof zh = {
  cardOnly: { title: 'Card discount only', hint: 'e.g. every ride at 90% with the card', name: 'Card discount' },
  weekendFree: { title: 'Free on weekends', hint: 'e.g. free every Saturday and Sunday', name: 'Free on weekends' },
  monthTier: { title: 'Monthly tiers', hint: 'e.g. 70% over ¥80, 50% over ¥150, back to card fare over ¥300', name: 'Monthly tier discount' },
  morning: { title: 'Morning discount', hint: 'e.g. 70% of full fare before 7:00 on weekdays', name: 'Morning discount', window: 'Morning' },
  nthFree: { title: 'Nth ride free', hint: 'e.g. 40th ride of the month is free', name: 'Ride count reward' },
}

const ja: typeof zh = {
  cardOnly: { title: 'カード割引のみ', hint: '例：毎回カード運賃の 9掛け', name: 'カード割引' },
  weekendFree: { title: '週末無料', hint: '例：毎週土日は無料', name: '週末無料' },
  monthTier: { title: '月間累計の段階', hint: '例：80 元超で 7掛け、150 元超で 5掛け、300 元超でカード運賃に戻る', name: '月間累計割引' },
  morning: { title: '早朝割引', hint: '例：平日 7 時前は全額の 7掛け', name: '早朝割引', window: '早朝' },
  nthFree: { title: 'N 回目無料', hint: '例：今月 40 回目の乗車は無料', name: '乗車回数特典' },
}

const ko: typeof zh = {
  cardOnly: { title: '카드 할인만', hint: '예: 매번 카드 요금 90%', name: '카드 할인' },
  weekendFree: { title: '주말 무료', hint: '예: 매주 토·일 무료', name: '주말 무료' },
  monthTier: { title: '월 누적 구간', hint: '예: 80위안 초과 70%, 150위안 초과 50%, 300위안 초과 시 카드 요금으로 복귀', name: '월 누적 할인' },
  morning: { title: '아침 할인', hint: '예: 평일 7시 이전 정가의 70%', name: '아침 할인', window: '아침' },
  nthFree: { title: 'N번째 무료', hint: '예: 이번 달 40번째 탑승 무료', name: '탑승 횟수 혜택' },
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
