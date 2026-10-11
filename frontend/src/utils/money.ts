import { t, currentLocale } from '../i18n'

/** 金额：前端展示元，后端存分 */
export function fenToYuan(fen: number): string {
  return (fen / 100).toFixed(2)
}

export function yuanToFen(yuan: string | number): number {
  const n = typeof yuan === 'number' ? yuan : parseFloat(yuan || '0')
  if (Number.isNaN(n)) return 0
  return Math.round(n * 100)
}

export function formatMoney(fen: number, type?: string): string {
  let sign = type === 'income' ? '+' : type === 'expense' ? '-' : ''
  if (type === 'revalue') sign = fen >= 0 ? '+' : '-'
  return `${sign}¥${fenToYuan(Math.abs(fen))}`
}

/** 金额配色：收入、浮盈为绿，其余为红 */
export function amountClass(tx: { type: string; amount: number }): string {
  if (tx.type === 'income' || (tx.type === 'revalue' && tx.amount >= 0)) return 'amount-income'
  return 'amount-expense'
}

const TX_TYPES = ['expense', 'income', 'transfer', 'revalue'] as const

export function typeLabel(type: string): string {
  return (TX_TYPES as readonly string[]).includes(type) ? t(`txType.${type}`) : type
}

/** 等额分期的首期本金（分）：余数并入首期 */
export function firstShareFen(total: number, periods: number): number {
  const n = Math.max(1, Math.min(60, Math.floor(periods) || 1))
  const sum = Math.max(0, Math.floor(total))
  return Math.floor(sum / n) + (sum % n)
}

/** 常用外币；输入框也允许手填其它 ISO 代码 */
export const FOREIGN_CURRENCIES = [
  'USD', 'EUR', 'JPY', 'HKD', 'GBP', 'SGD', 'AUD', 'CAD', 'KRW', 'TWD', 'THB', 'MYR', 'CHF',
] as const

/** 币种名按当前界面语言显示；需在渲染时调用以便切换语言后刷新 */
export function currencyName(code: string): string {
  try {
    return new Intl.DisplayNames([currentLocale()], { type: 'currency' }).of(code) || code
  } catch {
    return code
  }
}

/** 外币金额同样按 ×100 存储 */
export function formatForeign(currency?: string | null, amount?: number | null): string {
  if (!currency || !amount) return ''
  return `${currency} ${fenToYuan(amount)}`
}

/** 支付汇率：1 外币 = ? 人民币，保留 4 位 */
export function payRate(amountFen: number, foreignAmount?: number | null): string {
  if (!foreignAmount || foreignAmount <= 0 || !amountFen) return ''
  return (amountFen / foreignAmount).toFixed(4)
}

type TxLike = { type: string; remark?: string | null; category?: { name?: string } | null }

/** 流水标题：有备注用备注（花在哪），否则退回分类 / 类型 */
export function txTitle(tx: TxLike): string {
  return tx.remark?.trim() || tx.category?.name || typeLabel(tx.type)
}

/** 标题已用备注时，副标题补上分类，避免分类信息丢失 */
export function txSubCategory(tx: TxLike): string {
  return tx.remark?.trim() && tx.category?.name ? tx.category.name : ''
}
