import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '../api/http'

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
    return data as Account[]
  }

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

  async function creditRepayReminders() {
    const { data } = await http.get('/accounts/credit-repay-reminders')
    return (data || []) as CreditRepayReminder[]
  }

  /** 临时拉取明文敏感字段；调用方用完应丢弃，勿写入 list */
  async function fetchSensitive(id: number) {
    const { data } = await http.get(`/accounts/${id}/sensitive`)
    return data as { cardNo: string; holderName: string }
  }

  return { list, load, create, update, remove, creditStatement, creditRepayReminders, fetchSensitive }
})
