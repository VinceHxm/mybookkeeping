import { registerSW } from 'virtual:pwa-register'
import type { Router } from 'vue-router'

/** 主屏 App 从后台切回不会触发导航，浏览器不会自己查 sw.js，需要主动轮询 */
const CHECK_INTERVAL_MS = 30 * 60 * 1000

/** 表单 / 识别页填到一半时刷新会丢内容，等离开页面或切到后台再更新 */
const FORM_ROUTE = /^\/transactions\/(new|recognize|\d+)$|\/(new|edit)$/

export function setupPwaUpdate(router: Router) {
  if (!('serviceWorker' in navigator)) return

  let pending = false
  let applying = false

  const updateSW = registerSW({
    immediate: true,
    onNeedRefresh() {
      pending = true
      tryApply()
    },
    onRegisteredSW(swUrl, reg) {
      if (!reg) return
      const check = async () => {
        if (reg.installing || !navigator.onLine) return
        try {
          const resp = await fetch(swUrl, { cache: 'no-store', headers: { 'cache-control': 'no-cache' } })
          if (resp.status === 200) await reg.update()
        } catch { /* 离线等下次 */ }
      }
      setInterval(check, CHECK_INTERVAL_MS)
      document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible') void check()
        else tryApply()
      })
    },
  })

  function tryApply() {
    if (!pending || applying) return
    const onForm = FORM_ROUTE.test(router.currentRoute.value.path)
    if (onForm && document.visibilityState === 'visible') return
    applying = true
    void updateSW(true)
  }

  router.afterEach(() => tryApply())
}
