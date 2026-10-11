import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '../api/http'
import { t } from '../i18n'

export interface Attachment {
  id: number
  url: string
  contentType: string
  size: number
}

/** 团购券状态：'' 非券 | pending 待核销 | used 已核销 | refunded 已退款 */
export type VoucherStatus = '' | 'pending' | 'used' | 'refunded'

export function voucherStatusLabel(status: VoucherStatus): string {
  return status ? t(`voucher.status.${status}`) : ''
}

export const voucherStatusColor: Record<Exclude<VoucherStatus, ''>, string> = {
  pending: 'warning',
  used: 'success',
  refunded: 'info',
}

export interface VoucherRefundPayload {
  amount?: number
  accountId?: number
  categoryId?: number | null
  happenedAt?: string
  remark?: string
}

export interface Transaction {
  id: number
  /** revalue：理财账户市值校准，amount 为有符号差额 */
  type: 'expense' | 'income' | 'transfer' | 'revalue'
  amount: number
  accountId: number
  toAccountId?: number | null
  categoryId?: number | null
  remark: string
  geoMode?: string
  geoLng?: number | null
  geoLat?: number | null
  geoName?: string
  geoEndLng?: number | null
  geoEndLat?: number | null
  geoEndName?: string
  fareRuleId?: number | null
  installmentPeriods?: number
  interestFen?: number
  /** 分期首期推迟的账单周期数：0=随刷卡出账，1=下期出账 */
  installmentDeferCycles?: number
  amountPending?: boolean
  foreignCurrency?: string
  foreignAmount?: number
  voucherStatus?: VoucherStatus
  voucherExpireOn?: string
  refundOfId?: number | null
  happenedAt: string
  account?: { id: number; name: string }
  toAccount?: { id: number; name: string }
  category?: { id: number; name: string }
  attachments?: Attachment[]
  tags?: { id: number; name: string; color: string }[]
}

export interface TxPayload {
  type: string
  amount: number
  accountId: number
  toAccountId?: number | null
  categoryId?: number | null
  remark?: string
  geoMode?: string
  geoLng?: number | null
  geoLat?: number | null
  geoName?: string
  geoEndLng?: number | null
  geoEndLat?: number | null
  geoEndName?: string
  fareRuleId?: number | null
  installmentPeriods?: number
  interestFen?: number
  /** 分期首期推迟的账单周期数：0=随刷卡出账，1=下期出账 */
  installmentDeferCycles?: number
  amountPending?: boolean
  foreignCurrency?: string
  foreignAmount?: number
  voucherStatus?: VoucherStatus
  voucherExpireOn?: string
  happenedAt?: string
  attachmentIds?: number[]
  tagIds?: number[]
}

export const useTransactionStore = defineStore('transaction', () => {
  const list = ref<Transaction[]>([])
  const total = ref(0)
  const recent = ref<Transaction[]>([])

  async function load(params: Record<string, unknown> = {}) {
    const { data } = await http.get('/transactions', { params })
    // 兼容 {items,total} 与旧数组
    if (Array.isArray(data)) {
      list.value = data
      total.value = data.length
      return data as Transaction[]
    }
    list.value = data.items || []
    total.value = data.total || 0
    return list.value
  }

  async function loadRecent(limit = 5) {
    const { data } = await http.get('/transactions', { params: { limit } })
    recent.value = Array.isArray(data) ? data : (data.items || [])
    return recent.value
  }

  async function get(id: number) {
    const { data } = await http.get(`/transactions/${id}`)
    return data as Transaction
  }

  async function create(payload: TxPayload) {
    const { data } = await http.post('/transactions', payload)
    return data as Transaction
  }

  async function update(id: number, payload: TxPayload) {
    const { data } = await http.put(`/transactions/${id}`, payload)
    return data as Transaction
  }

  async function remove(id: number) {
    await http.delete(`/transactions/${id}`)
  }

  function patchLocal(item: Transaction) {
    const i = list.value.findIndex((x) => x.id === item.id)
    if (i >= 0) list.value[i] = item
  }

  async function setVoucherStatus(id: number, status: 'pending' | 'used') {
    const { data } = await http.post(`/transactions/${id}/voucher-status`, { status })
    patchLocal(data as Transaction)
    return data as Transaction
  }

  /** 团购券退款：后端生成退款收入并把券标为已退款，返回退款流水 */
  async function refundVoucher(id: number, payload: VoucherRefundPayload = {}) {
    const { data } = await http.post(`/transactions/${id}/refund`, payload)
    const i = list.value.findIndex((x) => x.id === id)
    if (i >= 0) list.value[i] = { ...list.value[i], voucherStatus: 'refunded' }
    return data as Transaction
  }

  return { list, total, recent, load, loadRecent, get, create, update, remove, setVoucherStatus, refundVoucher }
})
