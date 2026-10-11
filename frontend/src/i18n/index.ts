import { createI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import 'dayjs/locale/ja'
import 'dayjs/locale/ko'
import zhCN from './locales/zh-CN'
import en from './locales/en'
import ja from './locales/ja'
import ko from './locales/ko'
import type { AppLocale, LocaleBundle } from './types'

export type { AppLocale } from './types'

export const DEFAULT_LOCALE: AppLocale = 'zh-CN'

/** 语言名用各自母语显示，不随界面语言翻译 */
export const SUPPORTED_LOCALES: { value: AppLocale; label: string }[] = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en', label: 'English' },
  { value: 'ja', label: '日本語' },
  { value: 'ko', label: '한국어' },
]

const STORAGE_KEY = 'mbk_locale'

const DAYJS_LOCALE: Record<AppLocale, string> = { 'zh-CN': 'zh-cn', en: 'en', ja: 'ja', ko: 'ko' }

/** 把 zh / zh-Hans / en-US / ja-JP 等归一到支持的语言；繁体与不支持的返回 null */
export function normalizeLocale(raw?: string | null): AppLocale | null {
  if (!raw) return null
  const s = raw.trim().toLowerCase().replace('_', '-')
  if (s === 'zh' || s === 'zh-cn' || s === 'zh-sg' || s.startsWith('zh-hans')) return 'zh-CN'
  if (s.startsWith('zh')) return null
  if (s.startsWith('en')) return 'en'
  if (s.startsWith('ja')) return 'ja'
  if (s.startsWith('ko')) return 'ko'
  return null
}

export function detectLocale(): AppLocale {
  const stored = normalizeLocale(localStorage.getItem(STORAGE_KEY))
  if (stored) return stored
  for (const l of navigator.languages || [navigator.language]) {
    const n = normalizeLocale(l)
    if (n) return n
  }
  return DEFAULT_LOCALE
}

const initial = detectLocale()

const modules = import.meta.glob<{ default: LocaleBundle<Record<string, unknown>> }>('./modules/*.ts', { eager: true })

function buildMessages(l: AppLocale, base: Record<string, unknown>) {
  const out: Record<string, unknown> = { ...base }
  for (const [path, mod] of Object.entries(modules)) {
    const ns = path.slice(path.lastIndexOf('/') + 1, -3)
    out[ns] = mod.default[l]
  }
  return out
}

export const i18n = createI18n<[Record<string, unknown>], AppLocale, false>({
  legacy: false,
  locale: initial,
  fallbackLocale: DEFAULT_LOCALE,
  messages: {
    'zh-CN': buildMessages('zh-CN', zhCN),
    en: buildMessages('en', en),
    ja: buildMessages('ja', ja),
    ko: buildMessages('ko', ko),
  },
})

function applySideEffects(l: AppLocale) {
  dayjs.locale(DAYJS_LOCALE[l])
  document.documentElement.lang = l
  document.title = i18n.global.t('app.name')
}

applySideEffects(initial)

export function currentLocale(): AppLocale {
  return i18n.global.locale.value as AppLocale
}

export function setLocale(raw: string | null | undefined): AppLocale {
  const l = normalizeLocale(raw) || currentLocale()
  i18n.global.locale.value = l
  localStorage.setItem(STORAGE_KEY, l)
  applySideEffects(l)
  return l
}

/** 供组件外（工具函数、store、axios 拦截器）使用 */
export const t = i18n.global.t
