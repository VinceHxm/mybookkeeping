import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import http from '../api/http'
import { setLocale } from '../i18n'
import { useCategoryStore } from './category'

const TOKEN_KEY = 'mbk_token'

export interface UserProfile {
  id: number
  username: string
  email?: string
  emailVerified?: boolean
  role?: 'user' | 'admin'
  /** 默认还款转出账户（非信用） */
  defaultAccountId?: number | null
  /** 默认消费/收支账户 */
  defaultExpenseAccountId?: number | null
  weekStart?: number
  expenseColor?: string
  incomeColor?: string
  theme?: string
  /** 界面语言；老账号为空时按浏览器 / 本地缓存 */
  locale?: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const username = ref('')
  const profile = ref<UserProfile | null>(null)

  watch(profile, (p) => {
    if (p?.locale) setLocale(p.locale)
  })

  function isAdmin() {
    return profile.value?.role === 'admin'
  }

  function setToken(t: string) {
    token.value = t
    localStorage.setItem(TOKEN_KEY, t)
  }

  function clear() {
    token.value = null
    username.value = ''
    profile.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  async function login(u: string, p: string) {
    const { data } = await http.post('/auth/login', { username: u, password: p })
    setToken(data.token)
    username.value = data.user.username
    profile.value = data.user
    return data
  }

  async function register(u: string, p: string, email = '') {
    const { data } = await http.post('/auth/register', { username: u, password: p, email })
    if (data.token) {
      setToken(data.token)
      username.value = data.user.username
      profile.value = data.user
    }
    return data
  }

  async function logout() {
    try {
      await http.post('/auth/logout')
    } finally {
      clear()
    }
  }

  async function fetchMe() {
    const { data } = await http.get('/me')
    username.value = data.username
    profile.value = data
    return data as UserProfile
  }

  async function changePassword(oldPassword: string, newPassword: string) {
    await http.put('/me/password', { oldPassword, newPassword })
  }

  async function updateSettings(payload: Record<string, unknown>) {
    const { data } = await http.put('/me/settings', payload)
    profile.value = data
    username.value = data.username
    if ('locale' in payload) useCategoryStore().reloadIfLoaded().catch(() => {})
    return data as UserProfile
  }

  return {
    token, username, profile, isAdmin,
    login, register, logout, fetchMe, changePassword, updateSettings,
    clear, setToken,
  }
})
