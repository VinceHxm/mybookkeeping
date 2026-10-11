import type { LocaleBundle } from '../types'

const zh = {
  title: '计费规则',
  hint: '先定名称与基础折扣，再按需打开免费日、阶梯、时段等模块。原价可空；记账手填金额始终为准。',
  disabled: '已停用',
  holidayFree: '节假日免',
  defaultBase: '默认原价 {amount}',
  baseFromTemplate: '原价由模板 / 记一笔提供',
  cardRate: '刷卡 {zhe} 折',
  amountTiers: '{n} 金额档',
  countTiers: '{n} 次数档',
  timeWindows: '{n} 时段',
  freeDays: '{n} 免费日',
  empty: '暂无规则，点右上角添加',
  cycleCalendar: '自然月',
  cycleFromDay: '每月{day}日起算',
}

const en: typeof zh = {
  title: 'Fare rules',
  hint: 'Set a name and the basic discount first, then turn on free days, tiers or time windows as needed. Base fare is optional; the amount you enter when recording always wins.',
  disabled: 'Disabled',
  holidayFree: 'Free on holidays',
  defaultBase: 'Default base fare {amount}',
  baseFromTemplate: 'Base fare comes from the template / record',
  cardRate: 'Card {pct}%',
  amountTiers: '{n} amount tier(s)',
  countTiers: '{n} ride count tier(s)',
  timeWindows: '{n} time window(s)',
  freeDays: '{n} free day rule(s)',
  empty: 'No rules yet. Tap + at the top right to add one',
  cycleCalendar: 'Calendar month',
  cycleFromDay: 'Monthly from day {day}',
}

const ja: typeof zh = {
  title: '料金ルール',
  hint: 'まず名前と基本の割引を決め、必要に応じて無料日・段階・時間帯などを有効にします。通常運賃は空欄でも構いません。記録時に入力した金額が常に優先されます。',
  disabled: '停止中',
  holidayFree: '祝日無料',
  defaultBase: '既定の通常運賃 {amount}',
  baseFromTemplate: '通常運賃はテンプレート / 記録時に指定',
  cardRate: 'カード {zhe}掛け',
  amountTiers: '金額の段階 {n} 件',
  countTiers: '乗車回数の段階 {n} 件',
  timeWindows: '時間帯 {n} 件',
  freeDays: '無料日 {n} 件',
  empty: 'ルールはまだありません。右上から追加してください',
  cycleCalendar: '暦月',
  cycleFromDay: '毎月 {day} 日起算',
}

const ko: typeof zh = {
  title: '요금 규칙',
  hint: '먼저 이름과 기본 할인을 정한 뒤, 필요에 따라 무료일·구간·시간대 등을 켜세요. 기본 요금은 비워 둘 수 있으며, 기록할 때 입력한 금액이 항상 우선합니다.',
  disabled: '사용 중지',
  holidayFree: '공휴일 무료',
  defaultBase: '기본 요금 {amount}',
  baseFromTemplate: '기본 요금은 템플릿 / 기록 시 입력',
  cardRate: '카드 {pct}%',
  amountTiers: '금액 구간 {n}개',
  countTiers: '탑승 횟수 구간 {n}개',
  timeWindows: '시간대 {n}개',
  freeDays: '무료일 {n}개',
  empty: '아직 규칙이 없습니다. 오른쪽 위에서 추가하세요',
  cycleCalendar: '달력 기준 월',
  cycleFromDay: '매월 {day}일부터 계산',
}

export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
