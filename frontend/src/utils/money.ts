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
export function amountClass(t: { type: string; amount: number }): string {
  if (t.type === 'income' || (t.type === 'revalue' && t.amount >= 0)) return 'amount-income'
  return 'amount-expense'
}

export function typeLabel(t: string): string {
  return ({ expense: '支出', income: '收入', transfer: '转账', revalue: '市值校准' } as Record<string, string>)[t] || t
}

/** 等额分期的首期本金（分）：余数并入首期 */
export function firstShareFen(total: number, periods: number): number {
  const n = Math.max(1, Math.min(60, Math.floor(periods) || 1))
  const t = Math.max(0, Math.floor(total))
  return Math.floor(t / n) + (t % n)
}

/** 常用外币；输入框也允许手填其它 ISO 代码 */
export const FOREIGN_CURRENCIES = [
  { code: 'USD', name: '美元' },
  { code: 'EUR', name: '欧元' },
  { code: 'JPY', name: '日元' },
  { code: 'HKD', name: '港币' },
  { code: 'GBP', name: '英镑' },
  { code: 'SGD', name: '新加坡元' },
  { code: 'AUD', name: '澳元' },
  { code: 'CAD', name: '加元' },
  { code: 'KRW', name: '韩元' },
  { code: 'TWD', name: '新台币' },
  { code: 'THB', name: '泰铢' },
  { code: 'MYR', name: '马来西亚林吉特' },
  { code: 'CHF', name: '瑞士法郎' },
]

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
export function txTitle(t: TxLike): string {
  return t.remark?.trim() || t.category?.name || typeLabel(t.type)
}

/** 标题已用备注时，副标题补上分类，避免分类信息丢失 */
export function txSubCategory(t: TxLike): string {
  return t.remark?.trim() && t.category?.name ? t.category.name : ''
}
