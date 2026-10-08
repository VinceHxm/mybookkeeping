import dayjs from 'dayjs'

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
  const s = Math.abs(n).toLocaleString('zh-CN', { minimumFractionDigits: digits, maximumFractionDigits: digits })
  return `${n < 0 ? '-' : ''}¥${s}`
}

/** 日历格、坐标轴用的短金额：128 / 1.2万 / 3.4亿（单位元，不带 ¥） */
export function compactYuan(fen: number): string {
  const y = Math.abs(fen || 0) / 100
  const sign = fen < 0 ? '-' : ''
  const trim = (s: string) => s.replace(/\.0$/, '')
  if (y >= 1e8) return sign + trim((y / 1e8).toFixed(1)) + '亿'
  if (y >= 1e4) return sign + trim((y / 1e4).toFixed(1)) + '万'
  if (y >= 100) return sign + Math.round(y)
  return sign + trim(y.toFixed(1))
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
  if (g === 'day') return short ? dayjs(key).format('D') : dayjs(key).format('M月D日')
  if (g === 'year') return `${key}年`
  return short ? `${Number(key.slice(5))}月` : dayjs(`${key}-01`).format('YYYY年M月')
}

/** 某个周期对应的起止日期（含当天），用于跳转明细 */
export function periodRange(key: string, g: Granularity): { from: string; to: string } {
  const unit = g === 'day' ? 'day' : g === 'year' ? 'year' : 'month'
  const start = g === 'year' ? dayjs(`${key}-01-01`) : g === 'month' ? dayjs(`${key}-01`) : dayjs(key)
  return { from: start.format('YYYY-MM-DD'), to: start.endOf(unit).format('YYYY-MM-DD') }
}

export const granularityUnit: Record<Granularity, string> = { day: '日', month: '月', year: '年' }
