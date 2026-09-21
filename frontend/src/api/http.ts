import axios from 'axios'
import { useAuthStore } from '../stores/auth'

const http = axios.create({ baseURL: '/api', timeout: 30000 })

http.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  return config
})

http.interceptors.response.use(
  (res) => {
    const body = res.data
    if (body && body.ok === false) {
      return Promise.reject(new Error(body.error || '请求失败'))
    }
    return body?.data !== undefined ? { ...res, data: body.data } : res
  },
  (err) => {
    if (err.response?.status === 401) {
      const auth = useAuthStore()
      auth.clear()
      if (location.pathname !== '/login') location.href = '/login'
    }
    if (err.response?.status === 403) {
      const msg = err.response?.data?.error || ''
      if (typeof msg === 'string' && msg.includes('停用')) {
        const auth = useAuthStore()
        auth.clear()
        if (location.pathname !== '/login') location.href = '/login'
      }
    }
    const msg = err.response?.data?.error || err.message
    return Promise.reject(new Error(msg))
  },
)

export default http
