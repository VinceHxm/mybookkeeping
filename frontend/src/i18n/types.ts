export type AppLocale = 'zh-CN' | 'en' | 'ja' | 'ko'

/**
 * 模块文案：src/i18n/modules/<ns>.ts 默认导出四语对象，文件名即命名空间（t('<ns>.xxx')）。
 * 写法：const zh = {...}; const en: typeof zh = {...}; export default { 'zh-CN': zh, en, ja, ko } satisfies LocaleBundle<typeof zh>
 */
export type LocaleBundle<T> = Record<AppLocale, T>
