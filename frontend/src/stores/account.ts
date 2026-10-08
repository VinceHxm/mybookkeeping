import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '../api/http'
import { createDictCache } from './dictCache'
import type { Transaction } from './transaction'

export interface CreditInstallmentPlan {
  id?: number
  name: string
  principalFen: number
  periods: number
  interestPerPeriodFen: number
  firstDueOn: string
  sort?: number
}

export interface Account {
  id: number
  name: string
  type: string
  balance: number
  sort: number
  archived: boolean
  creditLimit?: number
  creditBilledFen?: number
  billingDay?: number
  paymentDueDay?: number
  institution?: string
  cardNo?: string
  holderName?: string
  storageNote?: string
  remark?: string
  attachments?: { id: number; url: string }[]
  installmentPlans?: CreditInstallmentPlan[]
  lastReconciledAt?: string | null
}

export interface CreditStatement {
  accountId: number
  accountName: string
  billingDay: number
  statementDate: string
  cycleStart: string
  cycleEnd: string
  nextStatementDate: string
  spentInCycle: number
  installmentDue: number
  nonInstallmentDue: number
  interestDue: number
  dueAmount: number
  repaidAmount: number
  /** 本期已还中的退款部分 */
  refundAmount?: number
  repaidInterest: number
  remainingAfterPay: number
  /** 整体待还 = 已用额度 */
  totalOutstanding: number
  periodRemaining: number
  futureInstallment: number
  outstandingBalance: number
  billedOutstanding: number
  unbilledOutstanding: number
  /** 未出账中的非分期部分 */
  nonInstallmentUnbilled?: number
  /** 预计下期分期本金入账 */
  nextPeriodInstallment?: number
  /** 预计下期约还（下期分期 + 未出账非分期） */
  nextPeriodEstimate?: number
  availableCredit?: number | null
  displayOnly: boolean
  note: string
  /** 尚有未出账期数的分期（流水 + 存量计划） */
  installments?: CreditInstallmentItem[]
}

export interface CreditInstallmentItem {
  source: 'transaction' | 'plan'
  sourceId: number
  name: string
  happenedOn?: string
  principalFen: number
  periods: number
  billedPeriods: number
  remainingFen: number
  nextDueOn: string
  nextShareFen: number
  lastDueOn: string
  /** 流水分期：手续费合计（计入首期） */
  interestFen: number
  /** 存量计划：每期利息 */
  interestPerPeriodFen: number
}

export interface CreditRepayReminder {
  accountId: number
  accountName: string
  paymentDueDay: number
  dueDate: string
  daysUntilDue: number
  amountFen: number
  fromAccountId?: number | null
  fromAccountName?: string
}

export type CreditTxStatus = 'paid' | 'billed_unpaid' | 'unbilled'

export interface AccountLedgerItem extends Transaction {
  direction: 'in' | 'out'
  creditStatus?: CreditTxStatus
  paidAmount: number
  billedUnpaidAmount: number
  unbilledAmount: number
  statementDate?: string
  billedPeriods?: number
}

export interface AccountLedgerMonth {
  month: string
  inAmount: number
  outAmount: number
  paid: number
  billedUnpaid: number
  unbilled: number
  /** 市值校准净额（理财账户），不含在流入/流出里 */
  revalued: number
  items: AccountLedgerItem[]
}

export interface InvestSummary {
  /** 净投入 = 期初余额 + 转入 − 转出 */
  netInvested: number
  /** 累计收益 = 当前市值 − 净投入 */
  gain: number
  /** 其中市值校准合计 */
  revalued: number
}

export const ACCOUNT_TYPE_LABELS: Record<string, string> = {
  cash: '现金',
  bank: '银行卡',
  credit: '信用',
  invest: '理财',
  other: '其他',
}

export interface AccountLedger {
  accountId: number
  accountName: string
  accountType: string
  balance: number
  creditTracking: boolean
  summary: {
    paid: number
    billedUnpaid: number
    unbilled: number
    openingDebt: number
    openingDebtUnpaid: number
    lastStatementDate?: string
    nextStatementDate?: string
  }
  invest?: InvestSummary
  months: AccountLedgerMonth[]
  note?: string
}

export function creditUsedFen(a: Account): number {
  return a.balance < 0 ? -a.balance : 0
}

export function creditBilledFenOf(a: Account): number {
  const used = creditUsedFen(a)
  const billed = Math.max(0, a.creditBilledFen || 0)
  return Math.min(billed, used)
}

export function creditUnbilledFen(a: Account): number {
  return creditUsedFen(a) - creditBilledFenOf(a)
}

export const useAccountStore = defineStore('account', () => {
  const list = ref<Account[]>([])

  async function load(archived = false) {
    const { data } = await http.get('/accounts', { params: archived ? { archived: 1 } : {} })
    list.value = data
    if (archived) cache.invalidate()
    else cache.markLoaded()
    return data as Account[]
  }
  const cache = createDictCache(() => load(), () => list.value)

  async function create(
    payload: Partial<Account> & {
      usedCredit?: number
      attachmentIds?: number[]
      installmentPlans?: CreditInstallmentPlan[]
    },
  ) {
    const { data } = await http.post('/accounts', payload)
    await load()
    return data
  }

  async function update(
    id: number,
    payload: Partial<Account> & {
      clearReconciled?: boolean
      usedCredit?: number
      attachmentIds?: number[]
      installmentPlans?: CreditInstallmentPlan[]
    },
  ) {
    const { data } = await http.put(`/accounts/${id}`, payload)
    await load(true)
    return data
  }

  async function remove(id: number) {
    await http.delete(`/accounts/${id}`)
    await load(true)
  }

  async function creditStatement(id: number) {
    const { data } = await http.get(`/accounts/${id}/credit-statement`)
    return data as CreditStatement
  }

  /** 全部已设账单日（未归档）信用卡的账单，一次请求 */
  async function creditStatements() {
    const { data } = await http.get('/accounts/credit-statements')
    return (data || []) as CreditStatement[]
  }

  async function creditRepayReminders() {
    const { data } = await http.get('/accounts/credit-repay-reminders')
    return (data || []) as CreditRepayReminder[]
  }

  async function ledger(id: number) {
    const { data } = await http.get(`/accounts/${id}/ledger`)
    return data as AccountLedger
  }

  /** 理财账户按当前市值校准；与余额一致时返回 null（只刷新对账时间） */
  async function revalue(id: number, payload: { marketValue: number; happenedAt?: string; remark?: string }) {
    const { data } = await http.post(`/accounts/${id}/revalue`, payload)
    return (data?.transaction ?? null) as Transaction | null
  }

  /** 临时拉取明文敏感字段（需登录密码）；调用方用完应丢弃，勿写入 list */
  async function fetchSensitive(id: number, password: string) {
    const { data } = await http.post(`/accounts/${id}/sensitive`, { password })
    return data as { cardNo: string; holderName: string }
  }

  return { list, load, ensure: cache.ensure, create, update, remove, creditStatement, creditStatements, creditRepayReminders, ledger, revalue, fetchSensitive }
})
