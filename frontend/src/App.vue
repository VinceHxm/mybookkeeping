<template>
  <v-app
    :class="{
      'app-shell--desktop': isDesktop && showNav,
      'app-shell--mobile-nav': showNav && !isDesktop,
    }"
    :theme="themeName"
  >
    <!-- 桌面端：左侧导航 -->
    <nav v-if="showNav && isDesktop" class="side-nav" aria-label="主导航">
      <div class="side-brand">
        <div class="side-logo">记</div>
        <div>
          <div class="side-title">我的账本</div>
          <div class="side-sub">快记优先</div>
        </div>
      </div>
      <v-btn
        v-for="item in navItems"
        :key="item.value"
        :to="item.to"
        :variant="tab === item.value ? 'flat' : 'text'"
        :color="tab === item.value ? 'primary' : undefined"
        class="side-link"
        block
        rounded="lg"
        size="large"
      >
        <v-icon start>{{ item.icon }}</v-icon>
        {{ item.label }}
      </v-btn>
      <div class="side-spacer" />
      <v-btn
        variant="tonal"
        color="primary"
        class="side-theme"
        block
        rounded="lg"
        @click="cycleTheme"
      >
        <v-icon start>{{ themeIcon }}</v-icon>
        {{ themeLabel }}
      </v-btn>
    </nav>

    <router-view />

    <!-- 移动端：底部导航 -->
    <v-bottom-navigation
      v-if="showNav && !isDesktop"
      v-model="tab"
      color="primary"
      grow
      elevation="0"
      class="nav"
    >
      <v-btn v-for="item in navItems" :key="item.value" :value="item.value" :to="item.to">
        <v-icon>{{ item.icon }}</v-icon>
        <span>{{ item.label }}</span>
      </v-btn>
    </v-bottom-navigation>
  </v-app>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useTheme } from 'vuetify'
import { useAuthStore } from './stores/auth'
import {
  getStoredTheme,
  useAppTheme,
  type ThemePreference,
} from './composables/useAppTheme'

const route = useRoute()
const auth = useAuthStore()
const vuetifyTheme = useTheme()
const { setPreference, syncFromProfile } = useAppTheme()

const tab = ref('home')
const width = ref(typeof window !== 'undefined' ? window.innerWidth : 1024)
const themePref = ref<ThemePreference>(getStoredTheme())
const isDesktop = computed(() => width.value >= 960)
const showNav = computed(() => {
  if (route.meta.hideNav || route.meta.public) return false
  const name = String(route.name || '')
  return !['login', 'register', 'forgot', 'transaction-edit', 'transaction-edit-id', 'text-recognize'].includes(name)
})
const themeName = computed(() => vuetifyTheme.global.name.value)
const themeIcon = computed(() => {
  if (themePref.value === 'dark') return 'mdi-weather-night'
  if (themePref.value === 'light') return 'mdi-white-balance-sunny'
  return 'mdi-theme-light-dark'
})
const themeLabel = computed(() => {
  if (themePref.value === 'dark') return '深色'
  if (themePref.value === 'light') return '浅色'
  return '跟随系统'
})

const navItems = [
  { value: 'home', to: '/', label: '首页', icon: 'mdi-home-outline' },
  { value: 'list', to: '/transactions', label: '明细', icon: 'mdi-format-list-bulleted' },
  { value: 'stats', to: '/stats', label: '统计', icon: 'mdi-chart-pie-outline' },
  { value: 'mine', to: '/mine', label: '我的', icon: 'mdi-account-outline' },
]

function onResize() {
  width.value = window.innerWidth
}

function cycleTheme() {
  const order: ThemePreference[] = ['light', 'dark', 'system']
  const next = order[(order.indexOf(themePref.value) + 1) % order.length]
  themePref.value = next
  setPreference(next, {
    expenseColor: auth.profile?.expenseColor,
    incomeColor: auth.profile?.incomeColor,
  })
  if (auth.token) {
    auth.updateSettings({ theme: next }).catch(() => {})
  }
}

onMounted(async () => {
  window.addEventListener('resize', onResize)
  themePref.value = getStoredTheme()
  setPreference(themePref.value)

  if (auth.token) {
    try {
      const profile = await auth.fetchMe()
      syncFromProfile(profile)
      themePref.value = (profile.theme as ThemePreference) || getStoredTheme()
    } catch {
      // 未登录或 token 失效时忽略
    }
  }
})

onUnmounted(() => window.removeEventListener('resize', onResize))

watch(
  () => auth.profile,
  (p) => {
    if (!p) return
    syncFromProfile(p)
    if (p.theme === 'light' || p.theme === 'dark' || p.theme === 'system') {
      themePref.value = p.theme
    }
  },
)

watch(
  () => route.path,
  (p) => {
    if (p.startsWith('/transactions') && route.name !== 'transaction-edit' && route.name !== 'transaction-edit-id') {
      tab.value = 'list'
    } else if (p.startsWith('/stats')) tab.value = 'stats'
    else if (
      p.startsWith('/mine') ||
      p.startsWith('/accounts') ||
      p.startsWith('/categories') ||
      p.startsWith('/tags') ||
      p.startsWith('/templates') ||
      p.startsWith('/fare-rules') ||
      p.startsWith('/schedules') ||
      p.startsWith('/settings') ||
      p.startsWith('/admin')
    ) {
      tab.value = 'mine'
    } else tab.value = 'home'
  },
  { immediate: true },
)
</script>

<style scoped>
.nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 10;
  /* 覆盖刘海机左右安全区由 body padding 处理；底部由 --sab 撑高 */
}

.side-nav {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: var(--nav-rail-width);
  z-index: 20;
  padding: 22px 14px;
  padding-top: max(22px, var(--sat));
  display: flex;
  flex-direction: column;
  gap: 6px;
  background: var(--nav-bg);
  border-right: 1px solid var(--surface-border);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  box-shadow: var(--shadow-soft);
}
.side-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 10px 22px;
  margin-bottom: 4px;
}
.side-logo {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  background: linear-gradient(145deg, #1b7f5a, #145a41);
  color: #fff;
  font-family: var(--font-display);
  font-weight: 700;
  font-size: 1.3rem;
  display: grid;
  place-items: center;
  box-shadow: 0 8px 18px rgba(27, 127, 90, 0.28);
}
.side-title {
  font-family: var(--font-display);
  font-weight: 700;
  line-height: 1.2;
  letter-spacing: -0.02em;
}
.side-sub { font-size: 0.75rem; color: var(--muted); margin-top: 2px; }
.side-link { justify-content: flex-start !important; letter-spacing: 0.02em; }
.side-spacer { flex: 1; }
.side-theme { justify-content: flex-start !important; margin-top: 8px; }
</style>
