import { computed, onMounted, onUnmounted, ref } from 'vue'

/** 与 App 桌面侧栏断点一致：&lt;960 视为手机/窄屏 */
export const DESKTOP_MIN = 960

export function useBreakpoint() {
  const width = ref(typeof window !== 'undefined' ? window.innerWidth : DESKTOP_MIN)
  const isDesktop = computed(() => width.value >= DESKTOP_MIN)
  const isMobile = computed(() => width.value < DESKTOP_MIN)

  function onResize() {
    width.value = window.innerWidth
  }

  onMounted(() => window.addEventListener('resize', onResize))
  onUnmounted(() => window.removeEventListener('resize', onResize))

  return { width, isDesktop, isMobile }
}
