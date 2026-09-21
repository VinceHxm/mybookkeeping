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
      <v-list-item to="/accounts" prepend-icon="mdi-wallet-outline" title="账户管理" append-icon="mdi-chevron-right" />
      <v-list-item to="/categories" prepend-icon="mdi-shape-outline" title="分类管理" append-icon="mdi-chevron-right" />
      <v-list-item to="/tags" prepend-icon="mdi-tag-outline" title="标签管理" append-icon="mdi-chevron-right" />
      <v-list-item to="/templates" prepend-icon="mdi-flash-outline" title="记账模板" append-icon="mdi-chevron-right" />
      <v-list-item to="/fare-rules" prepend-icon="mdi-ticket-percent-outline" title="计费规则" append-icon="mdi-chevron-right" />
      <v-list-item to="/schedules" prepend-icon="mdi-calendar-clock-outline" title="周期记账" append-icon="mdi-chevron-right" />
      <v-list-item to="/settings" prepend-icon="mdi-cog-outline" title="设置 / 改密" append-icon="mdi-chevron-right" />
      <v-list-item
        v-if="auth.isAdmin()"
        to="/admin/users"
        prepend-icon="mdi-shield-account-outline"
        title="用户管理"
        append-icon="mdi-chevron-right"
      />
    </v-list>

    <v-btn class="mt-2" block color="error" variant="tonal" size="large" @click="onLogout">退出登录</v-btn>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { getStoredTheme, useAppTheme, type ThemePreference } from '../composables/useAppTheme'

const auth = useAuthStore()
const router = useRouter()
const { setPreference } = useAppTheme()
const themePref = ref<ThemePreference>(getStoredTheme())

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
</style>
