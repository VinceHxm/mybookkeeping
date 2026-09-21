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
  const sign = type === 'income' ? '+' : type === 'expense' ? '-' : ''
  return `${sign}¥${fenToYuan(Math.abs(fen))}`
}

export function typeLabel(t: string): string {
  return ({ expense: '支出', income: '收入', transfer: '转账' } as Record<string, string>)[t] || t
}
