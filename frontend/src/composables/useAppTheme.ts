import { useTheme } from 'vuetify'

export type ThemePreference = 'light' | 'dark' | 'system'

const THEME_KEY = 'mbk_theme'
const EXPENSE_KEY = 'mbk_expense'
const INCOME_KEY = 'mbk_income'

let mediaListener: ((e: MediaQueryListEvent) => void) | null = null
let mediaQuery: MediaQueryList | null = null

export function getStoredTheme(): ThemePreference {
  const v = localStorage.getItem(THEME_KEY)
  if (v === 'dark' || v === 'light' || v === 'system') return v
  return 'system'
}

export function resolveThemeMode(pref: ThemePreference): 'light' | 'dark' {
  if (pref === 'dark') return 'dark'
  if (pref === 'light') return 'light'
  if (typeof window === 'undefined') return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

/** 仅同步 DOM / CSS（可在 Vue 挂载前调用，避免闪白） */
export function syncDocumentTheme(
  preference: ThemePreference = getStoredTheme(),
  colors?: { expenseColor?: string; incomeColor?: string },
) {
  const mode = resolveThemeMode(preference)
  const root = document.documentElement
  root.dataset.theme = mode
  root.style.colorScheme = mode

  const expenseColor = colors?.expenseColor || localStorage.getItem(EXPENSE_KEY) || '#c45c3e'
  const incomeColor = colors?.incomeColor || localStorage.getItem(INCOME_KEY) || '#1b7f5a'
  root.style.setProperty('--expense', expenseColor)
  root.style.setProperty('--income', incomeColor)
  localStorage.setItem(EXPENSE_KEY, expenseColor)
  localStorage.setItem(INCOME_KEY, incomeColor)
  localStorage.setItem(THEME_KEY, preference)
  return mode
}

function clearSystemListener() {
  if (mediaQuery && mediaListener) {
    mediaQuery.removeEventListener('change', mediaListener)
  }
  mediaQuery = null
  mediaListener = null
}

/** 在组件 setup / 生命周期内使用 */
export function useAppTheme() {
  const theme = useTheme()

  function setPreference(
    preference: ThemePreference,
    colors?: { expenseColor?: string; incomeColor?: string },
  ) {
    const mode = syncDocumentTheme(preference, colors)
    theme.global.name.value = mode

    clearSystemListener()
    if (preference === 'system') {
      mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      mediaListener = () => {
        const next = syncDocumentTheme('system', colors)
        theme.global.name.value = next
      }
      mediaQuery.addEventListener('change', mediaListener)
    }
    return mode
  }

  function syncFromProfile(profile?: {
    theme?: string
    expenseColor?: string
    incomeColor?: string
  } | null) {
    const raw = profile?.theme || getStoredTheme()
    const pref: ThemePreference =
      raw === 'dark' || raw === 'light' || raw === 'system' ? raw : 'system'
    return setPreference(pref, {
      expenseColor: profile?.expenseColor,
      incomeColor: profile?.incomeColor,
    })
  }

  return { setPreference, syncFromProfile, getStoredTheme, resolveThemeMode }
}
