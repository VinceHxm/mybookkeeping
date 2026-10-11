/** 记账常用 MDI 图标库（项目已引入 @mdi/font）；分组名与图标说明见 i18n/modules/icons.ts */
import { i18n, t } from '../i18n'

export const ICON_GROUPS = [
  'common',
  'food',
  'transport',
  'shopping',
  'home',
  'social',
  'health',
  'finance',
  'mark',
] as const

export type IconGroup = (typeof ICON_GROUPS)[number]
export type IconItem = { value: string; group: IconGroup }

const ICONS_BY_GROUP: Record<IconGroup, string[]> = {
  common: ['mdi-shape', 'mdi-star', 'mdi-heart', 'mdi-bookmark', 'mdi-flash', 'mdi-checkbox-marked-circle', 'mdi-dots-horizontal'],
  food: ['mdi-food', 'mdi-food-apple', 'mdi-coffee', 'mdi-beer', 'mdi-noodles', 'mdi-pizza', 'mdi-cup'],
  transport: [
    'mdi-bus', 'mdi-subway-variant', 'mdi-train', 'mdi-airplane', 'mdi-taxi', 'mdi-car', 'mdi-bike',
    'mdi-gas-station', 'mdi-parking', 'mdi-map-marker', 'mdi-map-marker-path', 'mdi-ticket-percent',
  ],
  shopping: [
    'mdi-cart', 'mdi-shopping', 'mdi-store', 'mdi-gift', 'mdi-tshirt-crew', 'mdi-shoe-formal',
    'mdi-cellphone', 'mdi-laptop', 'mdi-scissors-cutting', 'mdi-dog',
  ],
  home: ['mdi-home', 'mdi-home-city', 'mdi-lightning-bolt', 'mdi-water', 'mdi-wifi', 'mdi-sofa', 'mdi-broom', 'mdi-tools'],
  social: [
    'mdi-movie', 'mdi-gamepad-variant', 'mdi-music', 'mdi-microphone', 'mdi-account-group',
    'mdi-beach', 'mdi-camera', 'mdi-book-open-page-variant',
  ],
  health: ['mdi-hospital', 'mdi-pill', 'mdi-tooth', 'mdi-school', 'mdi-bookshelf', 'mdi-baby-carriage', 'mdi-dumbbell'],
  finance: [
    'mdi-cash', 'mdi-cash-multiple', 'mdi-cash-refund', 'mdi-wallet', 'mdi-credit-card', 'mdi-bank',
    'mdi-chart-line', 'mdi-piggy-bank', 'mdi-handshake', 'mdi-trophy', 'mdi-receipt', 'mdi-calculator',
    'mdi-calendar-clock',
  ],
  mark: ['mdi-tag', 'mdi-flag', 'mdi-alert-circle', 'mdi-information', 'mdi-fire', 'mdi-leaf', 'mdi-flower', 'mdi-weather-sunny'],
}

export const APP_ICONS: IconItem[] = ICON_GROUPS.flatMap((group) =>
  ICONS_BY_GROUP[group].map((value) => ({ value, group })),
)

/** mdi-food-apple → foodApple，对应 icons.labels 下的键 */
function labelKey(value: string): string {
  return 'icons.labels.' + value.replace(/^mdi-/, '').replace(/-([a-z0-9])/g, (_, c: string) => c.toUpperCase())
}

export function iconLabel(value: string): string {
  const key = labelKey(value)
  return i18n.global.te(key) ? t(key) : value.replace(/^mdi-/, '')
}

export function iconGroupLabel(group: IconGroup): string {
  return t(`icons.groups.${group}`)
}

export function accountTypeIcon(type: string): string {
  switch (type) {
    case 'cash': return 'mdi-cash'
    case 'bank': return 'mdi-bank'
    case 'credit': return 'mdi-credit-card'
    case 'invest': return 'mdi-chart-line'
    default: return 'mdi-wallet'
  }
}

export function txTypeIcon(type: string): string {
  switch (type) {
    case 'income': return 'mdi-arrow-down-bold-circle'
    case 'transfer': return 'mdi-swap-horizontal'
    case 'revalue': return 'mdi-chart-line'
    default: return 'mdi-arrow-up-bold-circle'
  }
}
