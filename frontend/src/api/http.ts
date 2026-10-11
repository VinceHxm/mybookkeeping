import axios from 'axios'
import { useAuthStore } from '../stores/auth'
import { currentLocale, t } from '../i18n'

const http = axios.create({ baseURL: '/api', timeout: 30000 })

http.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  config.headers['Accept-Language'] = currentLocale()
  return config
})

http.interceptors.response.use(
  (res) => {
    const body = res.data
    if (body && body.ok === false) {
      return Promise.reject(new Error(body.error || t('common.requestFailed')))
    }
    return body?.data !== undefined ? { ...res, data: body.data } : res
  },
  (err) => {
    if (err.response?.status === 401) {
      const auth = useAuthStore()
      auth.clear()
      if (location.pathname !== '/login') location.href = '/login'
    }
    if (err.response?.status === 403 && err.response?.data?.code === 'user_disabled') {
      const auth = useAuthStore()
      auth.clear()
      if (location.pathname !== '/login') location.href = '/login'
    }
    const msg = err.response?.data?.error || err.message
    return Promise.reject(new Error(msg))
  },
)

export default http
