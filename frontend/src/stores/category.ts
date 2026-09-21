import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import http from '../api/http'

export interface Category {
  id: number
  name: string
  kind: 'expense' | 'income'
  icon: string
  sort: number
  parentId?: number | null
  children?: Category[]
}

export const useCategoryStore = defineStore('category', () => {
  const list = ref<Category[]>([])

  const expense = computed(() => list.value.filter((c) => c.kind === 'expense'))
  const income = computed(() => list.value.filter((c) => c.kind === 'income'))

  /** 记账用：叶子优先；无子类的一级也可选 */
  const expenseOptions = computed(() => flattenOptions(expense.value))
  const incomeOptions = computed(() => flattenOptions(income.value))

  function flattenOptions(roots: Category[]) {
    const out: { id: number; name: string; icon: string }[] = []
    for (const r of roots) {
      const children = list.value.filter((c) => c.parentId === r.id)
      if (children.length) {
        for (const ch of children) {
          out.push({ id: ch.id, name: `${r.name} / ${ch.name}`, icon: ch.icon || r.icon })
        }
      } else if (!r.parentId) {
        out.push({ id: r.id, name: r.name, icon: r.icon })
      }
    }
    // 也包含孤立子类
    for (const c of list.value) {
      if (c.parentId && !out.find((o) => o.id === c.id) && roots.some((r) => r.kind === c.kind)) {
        const p = list.value.find((x) => x.id === c.parentId)
        out.push({ id: c.id, name: p ? `${p.name} / ${c.name}` : c.name, icon: c.icon })
      }
    }
    return out
  }

  async function load(kind?: string) {
    const { data } = await http.get('/categories', { params: { flat: 1, ...(kind ? { kind } : {}) } })
    list.value = data
    return data as Category[]
  }

  async function create(payload: Partial<Category>) {
    const { data } = await http.post('/categories', payload)
    await load()
    return data
  }

  async function update(id: number, payload: Partial<Category>) {
    const { data } = await http.put(`/categories/${id}`, payload)
    await load()
    return data
  }

  async function remove(id: number) {
    await http.delete(`/categories/${id}`)
    await load()
  }

  return { list, expense, income, expenseOptions, incomeOptions, load, create, update, remove }
})
