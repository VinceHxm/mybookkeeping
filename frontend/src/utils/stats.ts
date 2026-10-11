import dayjs from 'dayjs'
import { currentLocale, t } from '../i18n'

/** 统计页筛选：from / to 均为含当天的 YYYY-MM-DD */
export interface StatsFilter {
  from: string
  to: string
  accountId: number
}

export type Granularity = 'day' | 'month' | 'year'

export function statsParams(f: StatsFilter, extra: Record<string, unknown> = {}) {
  const p: Record<string, unknown> = {
    from: f.from,
    to: dayjs(f.to).add(1, 'day').format('YYYY-MM-DD'),
    ...extra,
  }
  if (f.accountId) p.accountId = f.accountId
  return p
}

/** ¥1,234.56 */
export function yuan(fen: number, digits = 2): string {
  const n = (fen || 0) / 100
  const s = Math.abs(n).toLocaleString(currentLocale(), { minimumFractionDigits: digits, maximumFractionDigits: digits })
  return `${n < 0 ? '-' : ''}¥${s}`
}

/** 日历格、坐标轴用的短金额（单位元，不带 ¥）：中日 1.2万 / 韩 1.2만 / 英 1.2K，进位由 Intl 按语言决定 */
export function compactYuan(fen: number): string {
  const y = Math.abs(fen || 0) / 100
  const sign = fen < 0 ? '-' : ''
  const v = y >= 100 ? Math.round(y) : y
  return sign + new Intl.NumberFormat(currentLocale(), { notation: 'compact', maximumFractionDigits: 1 }).format(v)
}

/** 相对上期的变化比例；上期为 0 时无意义返回 null */
export function changeRatio(cur: number, prev: number): number | null {
  if (!prev) return null
  return (cur - prev) / Math.abs(prev)
}

export function percent(r: number, digits = 1): string {
  return `${(r * 100).toFixed(digits)}%`
}

export function periodLabel(key: string, g: Granularity, short = false): string {
  if (g === 'day') return short ? dayjs(key).format('D') : dayjs(key).format(t('stats.fmt.monthDay'))
  if (g === 'year') return dayjs(`${key}-01-01`).format(t('stats.fmt.year'))
  return dayjs(`${key}-01`).format(t(short ? 'stats.fmt.monthShort' : 'stats.fmt.month'))
}

/** 某个周期对应的起止日期（含当天），用于跳转明细 */
export function periodRange(key: string, g: Granularity): { from: string; to: string } {
  const unit = g === 'day' ? 'day' : g === 'year' ? 'year' : 'month'
  const start = g === 'year' ? dayjs(`${key}-01-01`) : g === 'month' ? dayjs(`${key}-01`) : dayjs(key)
  return { from: start.format('YYYY-MM-DD'), to: start.endOf(unit).format('YYYY-MM-DD') }
}

export function granularityUnit(g: Granularity): string {
  return t(`stats.unit.${g}`)
}
