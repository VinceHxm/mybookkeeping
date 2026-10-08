<template>
  <div class="page">
    <h1 class="page-title">我的</h1>
    <v-card class="mb-4 user-card surface-card" rounded="lg" elevation="0">
      <v-card-text>
        <div class="user-row">
          <v-avatar color="primary" size="52" class="mr-3">
            <span class="avatar-letter">{{ (auth.username || '用').slice(0, 1) }}</span>
          </v-avatar>
          <div>
            <div class="user">{{ auth.username || '用户' }}</div>
            <div class="muted">
              <span v-if="auth.isAdmin()" class="role-admin">管理员</span>
              <span v-else>普通用户</span>
              · {{ auth.profile?.email || '个人精简账本' }}
            </div>
          </div>
        </div>
      </v-card-text>
    </v-card>

    <div class="theme-bar surface-card mb-3 d-md-none">
      <span class="theme-bar-label">外观</span>
      <div class="theme-bar-actions">
        <v-btn
          v-for="opt in themeOptions"
          :key="opt.value"
          size="small"
          :variant="themePref === opt.value ? 'flat' : 'text'"
          :color="themePref === opt.value ? 'primary' : undefined"
          @click="quickTheme(opt.value)"
        >
          <v-icon start size="16">{{ opt.icon }}</v-icon>
          {{ opt.title }}
        </v-btn>
      </div>
    </div>

    <v-list rounded="lg" class="menu mb-3 surface-card" elevation="0">
      <v-list-item
        v-for="m in menu"
        :key="m.to"
        :to="m.to"
        :prepend-icon="m.icon"
        :title="m.title"
        :subtitle="m.desc"
        append-icon="mdi-chevron-right"
      />
    </v-list>

    <v-btn class="mt-2 logout-btn" block color="error" variant="tonal" size="large" @click="onLogout">退出登录</v-btn>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { getStoredTheme, useAppTheme, type ThemePreference } from '../composables/useAppTheme'

const auth = useAuthStore()
const router = useRouter()
const { setPreference } = useAppTheme()
const themePref = ref<ThemePreference>(getStoredTheme())

const menu = computed(() => [
  { to: '/accounts', icon: 'mdi-wallet-outline', title: '账户管理', desc: '余额、信用卡账单、默认账户' },
  { to: '/categories', icon: 'mdi-shape-outline', title: '分类管理', desc: '收支一级 / 二级分类' },
  { to: '/tags', icon: 'mdi-tag-outline', title: '标签管理', desc: '跨分类的标记与筛选' },
  { to: '/templates', icon: 'mdi-flash-outline', title: '记账模板', desc: '首页一键入账' },
  { to: '/fare-rules', icon: 'mdi-ticket-percent-outline', title: '计费规则', desc: '公交地铁折扣与阶梯' },
  { to: '/schedules', icon: 'mdi-calendar-clock-outline', title: '周期记账', desc: '房租、订阅等自动入账' },
  { to: '/settings', icon: 'mdi-cog-outline', title: '设置 / 改密', desc: '外观、偏好、密码' },
  ...(auth.isAdmin()
    ? [{ to: '/admin/users', icon: 'mdi-shield-account-outline', title: '用户管理', desc: '角色、停用、重置密码' }]
    : []),
])

const themeOptions: { title: string; value: ThemePreference; icon: string }[] = [
  { title: '浅', value: 'light', icon: 'mdi-white-balance-sunny' },
  { title: '深', value: 'dark', icon: 'mdi-weather-night' },
  { title: '系统', value: 'system', icon: 'mdi-theme-light-dark' },
]

function quickTheme(pref: ThemePreference) {
  themePref.value = pref
  setPreference(pref, {
    expenseColor: auth.profile?.expenseColor,
    incomeColor: auth.profile?.incomeColor,
  })
  if (auth.token) {
    auth.updateSettings({ theme: pref }).catch(() => {})
  }
}

onMounted(() => {
  themePref.value = getStoredTheme()
  auth.fetchMe().then((p) => {
    if (p.theme === 'light' || p.theme === 'dark' || p.theme === 'system') {
      themePref.value = p.theme
    }
  }).catch(() => {})
})

async function onLogout() {
  await auth.logout()
  await router.replace('/login')
}
</script>

<style scoped>
.user-card { border: 1px solid var(--surface-border); }
.user-row { display: flex; align-items: center; }
.avatar-letter { font-family: var(--font-display); font-weight: 700; color: #fff; }
.user { font-size: 1.25rem; font-weight: 700; font-family: var(--font-display); }
.muted { color: var(--muted); margin-top: 4px; font-size: 0.9rem; }
.menu {
  background: var(--surface) !important;
  border: 1px solid var(--surface-border);
  overflow: hidden;
}
.theme-bar {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 14px;
}
.theme-bar-label { color: var(--muted); font-size: 0.85rem; font-weight: 600; }
.theme-bar-actions { display: flex; flex-wrap: wrap; gap: 4px; }
.role-admin { color: var(--primary); font-weight: 600; }

/* 手机保持原来的单行菜单，说明文字只在宽屏磁贴里出现 */
.menu :deep(.v-list-item-subtitle) { display: none; }

@media (min-width: 768px) {
  .menu {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(270px, 1fr));
    gap: 12px;
    padding: 0 !important;
    background: transparent !important;
    border: 0;
    box-shadow: none !important;
    backdrop-filter: none;
    overflow: visible;
  }
  .menu :deep(.v-list-item) {
    min-height: 76px;
    padding: 14px 16px !important;
    border-radius: var(--radius-md) !important;
    background: var(--surface);
    border: 1px solid var(--surface-border);
    box-shadow: var(--shadow-soft);
  }
  .menu :deep(.v-list-item__prepend > .v-icon) {
    color: var(--primary);
    opacity: 1;
  }
  .menu :deep(.v-list-item-title),
  .menu :deep(.v-list-item-subtitle) {
    display: block;
    -webkit-line-clamp: unset;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .menu :deep(.v-list-item-title) { font-weight: 600; }
  .menu :deep(.v-list-item-subtitle) { margin-top: 2px; }
  .logout-btn { width: auto !important; min-width: 200px; }
}
</style>
