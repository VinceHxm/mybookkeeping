const TOKEN_KEY = 'mbk_token'

/**
 * 字典类数据（账户 / 分类 / 标签）的按登录态缓存：
 * 同一登录 token 下已完整加载过就直接复用；并发调用共用同一个请求。
 * 增删改后 store 自身会重新 load，缓存随之保持最新。
 */
export function createDictCache<T>(load: () => Promise<T>, current: () => T) {
  let loadedFor: string | null = null
  let inflight: Promise<T> | null = null

  function markLoaded() {
    loadedFor = localStorage.getItem(TOKEN_KEY)
  }

  function invalidate() {
    loadedFor = null
  }

  async function ensure(): Promise<T> {
    const token = localStorage.getItem(TOKEN_KEY)
    if (token && token === loadedFor) return current()
    if (!inflight) {
      inflight = load().finally(() => { inflight = null })
    }
    return inflight
  }

  return { ensure, markLoaded, invalidate }
}
