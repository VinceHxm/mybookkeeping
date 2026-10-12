<template>
  <transition name="hint">
    <div v-if="visible" class="install-hint" role="dialog" :aria-label="t('installHint.title')">
      <img class="hint-icon" src="/apple-touch-icon-180x180.png" alt="" />
      <div class="hint-text">
        <div class="hint-title">{{ t('installHint.title') }}</div>
        <i18n-t keypath="installHint.body" tag="div" class="hint-body">
          <template #share>
            <v-icon size="16" class="share-ico">mdi-export-variant</v-icon>
          </template>
        </i18n-t>
      </div>
      <v-btn size="small" variant="tonal" color="primary" class="hint-ok" @click="dismiss">
        {{ t('installHint.ok') }}
      </v-btn>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

const STORAGE_KEY = 'mbk_install_hint_dismissed'

const { t } = useI18n()
const visible = ref(false)
let timer: ReturnType<typeof setTimeout> | null = null

function shouldShow() {
  if (localStorage.getItem(STORAGE_KEY)) return false
  const ua = navigator.userAgent
  // iPadOS 13+ 默认伪装成 Mac Safari，需靠触控点数区分
  const isIOS = /iPhone|iPad|iPod/.test(ua) || (/Macintosh/.test(ua) && navigator.maxTouchPoints > 1)
  if (!isIOS) return false
  const standalone = (navigator as Navigator & { standalone?: boolean }).standalone === true
    || window.matchMedia('(display-mode: standalone)').matches
  if (standalone) return false
  // App 内置浏览器没有「添加到主屏幕」
  return !/MicroMessenger|QQ\/|Weibo|DingTalk|AlipayClient/i.test(ua)
}

function dismiss() {
  visible.value = false
  localStorage.setItem(STORAGE_KEY, '1')
}

onMounted(() => {
  if (!shouldShow()) return
  timer = setTimeout(() => { visible.value = true }, 1500)
})

onUnmounted(() => {
  if (timer) clearTimeout(timer)
})
</script>

<style scoped>
.install-hint {
  position: fixed;
  left: 50%;
  bottom: calc(var(--nav-total-h, 0px) + 12px);
  transform: translateX(-50%);
  z-index: 30;
  width: min(calc(100% - 24px), 420px);
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 12px 12px 14px;
  border-radius: 18px;
  background: var(--surface-solid);
  border: 1px solid color-mix(in srgb, var(--primary) 24%, transparent);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.22);
}

.hint-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  flex: 0 0 auto;
}

.hint-text {
  flex: 1 1 auto;
  min-width: 0;
}

.hint-title {
  font-size: 0.86rem;
  font-weight: 700;
  line-height: 1.3;
}

.hint-body {
  margin-top: 2px;
  font-size: 0.76rem;
  color: var(--muted);
  line-height: 1.45;
}

.share-ico {
  color: var(--primary);
  vertical-align: -3px;
}

.hint-ok {
  flex: 0 0 auto;
}

.hint-enter-active,
.hint-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.hint-enter-from,
.hint-leave-to {
  opacity: 0;
  transform: translate(-50%, 12px);
}
</style>
