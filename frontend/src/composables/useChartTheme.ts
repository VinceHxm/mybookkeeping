import { computed } from 'vue'
import { useTheme } from 'vuetify'
import { useAuthStore } from '../stores/auth'

export const CHART_PALETTE = [
  '#1b7f5a', '#e8a838', '#3f7cc4', '#c45c3e', '#8e5bb5', '#2aa5a0',
  '#d4699a', '#7a8b3a', '#5c6bc0', '#b0764a', '#4f9d69', '#78909c',
]

export const OTHER_COLOR = '#9aa8a0'

export function useChartTheme() {
  const theme = useTheme()
  const auth = useAuthStore()
  return computed(() => {
    const dark = theme.global.current.value.dark
    // 收支色来自用户设置（写在根元素 CSS 变量上），读 profile 以便改色后重算
    void auth.profile?.expenseColor
    void auth.profile?.incomeColor
    const css = getComputedStyle(document.documentElement)
    const v = (name: string, fallback: string) => css.getPropertyValue(name).trim() || fallback
    return {
      dark,
      text: dark ? '#c5d4cb' : '#5a6f64',
      axis: dark ? '#3a4a42' : '#d5e0d9',
      surface: dark ? '#1c2620' : '#ffffff',
      expense: v('--expense', '#c45c3e'),
      income: v('--income', '#1b7f5a'),
      primary: v('--primary', '#1b7f5a'),
      palette: CHART_PALETTE,
    }
  })
}
