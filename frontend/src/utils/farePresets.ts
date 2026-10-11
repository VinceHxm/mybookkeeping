import type { FareRule } from './fare'
import { t } from '../i18n'

/** 场景起步：一键打开对应模块并填入可改示例（不绑定具体城市） */
export type FarePreset = {
  id: string
  title: string
  hint: string
  /** 打开哪些编辑模块 */
  modules?: Partial<{
    freeDays: boolean
    holiday: boolean
    timeWin: boolean
    amountTier: boolean
    countTier: boolean
  }>
  rule: Partial<FareRule>
}

const workdays = [1, 2, 3, 4, 5]

/** 文案随当前语言生成，放在 computed 里调用即可随切换语言刷新 */
export function getFarePresets(): FarePreset[] {
  return [
    {
      id: 'card-only',
      title: t('farePresets.cardOnly.title'),
      hint: t('farePresets.cardOnly.hint'),
      rule: {
        name: t('farePresets.cardOnly.name'),
        cardRate: 0.9,
        cycleType: 'calendar_month',
        freePeriods: [],
        tiers: [],
        timeWindows: [],
        countTiers: [],
        freeOnHoliday: false,
      },
    },
    {
      id: 'weekend-free',
      title: t('farePresets.weekendFree.title'),
      hint: t('farePresets.weekendFree.hint'),
      modules: { freeDays: true },
      rule: {
        name: t('farePresets.weekendFree.name'),
        cardRate: 1,
        freePeriods: [{ startDate: '', endDate: '', recur: 'weekly', weekdays: [0, 6] }],
        tiers: [],
        timeWindows: [],
        countTiers: [],
      },
    },
    {
      id: 'month-tier',
      title: t('farePresets.monthTier.title'),
      hint: t('farePresets.monthTier.hint'),
      modules: { amountTier: true },
      rule: {
        name: t('farePresets.monthTier.name'),
        cardRate: 0.9,
        cycleType: 'calendar_month',
        tiers: [
          { minFen: 8000, maxFen: 15000, rate: 0.7, base: 'full' },
          { minFen: 15000, maxFen: 30000, rate: 0.5, base: 'full' },
          { minFen: 30000, maxFen: 0, rate: 1, base: 'card' },
        ],
        timeWindows: [],
        countTiers: [],
        freePeriods: [],
      },
    },
    {
      id: 'morning',
      title: t('farePresets.morning.title'),
      hint: t('farePresets.morning.hint'),
      modules: { timeWin: true },
      rule: {
        name: t('farePresets.morning.name'),
        cardRate: 1,
        timeWindows: [
          {
            name: t('farePresets.morning.window'),
            startHm: '00:00',
            endHm: '07:00',
            weekdays: workdays,
            rate: 0.7,
            base: 'full',
          },
        ],
        tiers: [],
        countTiers: [],
        freePeriods: [],
      },
    },
    {
      id: 'nth-free',
      title: t('farePresets.nthFree.title'),
      hint: t('farePresets.nthFree.hint'),
      modules: { countTier: true },
      rule: {
        name: t('farePresets.nthFree.name'),
        cardRate: 1,
        countTiers: [{ minCount: 39, maxCount: 40, free: true, rate: 1, base: 'card' }],
        tiers: [],
        timeWindows: [],
        freePeriods: [],
      },
    },
  ]
}
