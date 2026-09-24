import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '../api/http'

export interface AdminUser {
  id: number
  username: string
  email?: string
  role: 'user' | 'admin'
  disabled: boolean
  createdAt?: string
  updatedAt?: string
}

export const useAdminUsersStore = defineStore('adminUsers', () => {
  const list = ref<AdminUser[]>([])

  async function load() {
    const { data } = await http.get('/admin/users')
    list.value = (data || []) as AdminUser[]
    return list.value
  }

  async function update(id: number, payload: { role?: string; disabled?: boolean }) {
    const { data } = await http.put(`/admin/users/${id}`, payload)
    await load()
    return data as AdminUser
  }

  async function remove(id: number) {
    await http.delete(`/admin/users/${id}`)
    await load()
  }

  async function resetPassword(id: number, newPassword: string) {
    await http.put(`/admin/users/${id}/password`, { newPassword })
  }

  return { list, load, update, remove, resetPassword }
})
