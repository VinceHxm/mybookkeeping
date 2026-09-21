import type { FareRule } from './fare'

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

export const FARE_PRESETS: FarePreset[] = [
  {
    id: 'card-only',
    title: '仅刷卡折',
    hint: '例：每次按票卡 9 折计',
    rule: {
      name: '刷卡折扣',
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
    title: '周末免费',
    hint: '例：每周六、日免费',
    modules: { freeDays: true },
    rule: {
      name: '周末免费',
      cardRate: 1,
      freePeriods: [{ startDate: '', endDate: '', recur: 'weekly', weekdays: [0, 6] }],
      tiers: [],
      timeWindows: [],
      countTiers: [],
    },
  },
  {
    id: 'month-tier',
    title: '月累计阶梯',
    hint: '例：超 80 元 7 折、超 150 元 5 折、超 300 回票卡',
    modules: { amountTier: true },
    rule: {
      name: '月累计优惠',
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
    title: '早间折扣',
    hint: '例：工作日 7 点前全价 7 折',
    modules: { timeWin: true },
    rule: {
      name: '早间优惠',
      cardRate: 1,
      timeWindows: [
        {
          name: '早间',
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
    title: '第 N 次免费',
    hint: '例：本月第 40 次乘车免费',
    modules: { countTier: true },
    rule: {
      name: '乘次奖励',
      cardRate: 1,
      countTiers: [{ minCount: 39, maxCount: 40, free: true, rate: 1, base: 'card' }],
      tiers: [],
      timeWindows: [],
      freePeriods: [],
    },
  },
]
