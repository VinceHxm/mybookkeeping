import http from '../api/http'
import { t } from '../i18n'

declare global {
  interface Window {
    _AMapSecurityConfig?: { securityJsCode: string }
    AMap?: any
  }
}

let pending: Promise<any> | null = null

function loadScript(src: string) {
  return new Promise<void>((resolve, reject) => {
    if (document.querySelector(`script[src="${src}"]`)) {
      resolve()
      return
    }
    const s = document.createElement('script')
    s.src = src
    s.async = true
    s.onload = () => resolve()
    s.onerror = () => reject(new Error(t('amap.scriptFailed')))
    document.head.appendChild(s)
  })
}

/** 加载高德 JSAPI 2.0（全局只加载一次），返回 window.AMap */
export function loadAmap(): Promise<any> {
  if (window.AMap) return Promise.resolve(window.AMap)
  if (!pending) {
    pending = (async () => {
      const { data } = await http.get('/amap/config')
      if (!data?.key) throw new Error(t('amap.noKey'))
      window._AMapSecurityConfig = { securityJsCode: data.securityJsCode }
      await loadScript(`https://webapi.amap.com/maps?v=2.0&key=${data.key}`)
      if (!window.AMap) throw new Error(t('amap.scriptFailed'))
      return window.AMap
    })().catch((e) => {
      pending = null
      throw e
    })
  }
  return pending
}
