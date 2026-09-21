import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '../api/http'

export interface Attachment {
  id: number
  url: string
  contentType: string
  size: number
}

export interface Transaction {
  id: number
  type: 'expense' | 'income' | 'transfer'
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

  async function loadRecent() {
    const { data } = await http.get('/transactions', { params: { limit: 5 } })
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

  return { list, total, recent, load, loadRecent, get, create, update, remove }
})
