export type FareFreePeriod = {
  startDate: string
  endDate: string
  recur: 'none' | 'yearly' | 'weekly'
  weekdays?: number[]
}

export type FareTier = {
  minFen: number
  maxFen: number // 0 = 无上限
  rate: number
  base: 'full' | 'card'
  /** @deprecated 旧字段 */
  thresholdFen?: number
}

export type FareTimeWindow = {
  name?: string
  startHm: string
  endHm: string
  weekdays?: number[]
  rate: number
  free?: boolean
  base?: 'full' | 'card'
}

export type FareCountTier = {
  minCount: number
  maxCount: number
  rate: number
  base?: 'full' | 'card'
  free?: boolean
  amountOffFen?: number
}

export type FareRule = {
  id?: number
  name: string
  city: string
  note?: string
  validFrom?: string
  validTo?: string
  baseFen: number
  cardRate: number
  amountOffFen?: number
  stackMode?: 'prefer' | 'lowest' | 'stack'
  freeOnHoliday?: boolean
  freePeriods: FareFreePeriod[]
  tiers: FareTier[]
  timeWindows?: FareTimeWindow[]
  countTiers?: FareCountTier[]
  cycleType: 'calendar_month' | 'from_day'
  cycleStartDay: number
  /** 本周期手工补录累计（分），仅 cycleSeedKey 对应当前周期时生效 */
  cycleSeedFen?: number
  cycleSeedKey?: string
  cycleCountSeed?: number
  enabled: boolean
  icon?: string
  sort?: number
}

export function zheToRate(zhe: string | number) {
  const n = Number(zhe)
  if (!Number.isFinite(n) || n <= 0) return 1
  return n / 10
}

export function rateToZhe(rate: number) {
  if (!rate || rate <= 0) return '10'
  return String(Math.round(rate * 10 * 100) / 100)
}

export function normalizeCity(s: string) {
  return (s || '').trim().replace(/市$/, '')
}

export const CITY_KEY = 'mbk_amap_city'

export function getCachedCity() {
  try {
    return normalizeCity(localStorage.getItem(CITY_KEY) || '')
  } catch {
    return ''
  }
}

export function setCachedCity(city: string) {
  const n = normalizeCity(city)
  if (!n) return
  try {
    localStorage.setItem(CITY_KEY, n)
  } catch { /* ignore */ }
}

/** 规则地市空或不匹配策略：无用户城市 → 允许自动；有则需匹配 */
export function cityAllowsAuto(ruleCity: string, userCity: string) {
  const rc = normalizeCity(ruleCity)
  if (!rc) return true
  const uc = normalizeCity(userCity)
  if (!uc) return true
  return uc === rc || uc.includes(rc) || rc.includes(uc)
}
