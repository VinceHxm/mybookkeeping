<template>
  <div class="page">
    <h1 class="page-title">{{ t('mine.title') }}</h1>
    <v-card class="mb-4 user-card surface-card" rounded="lg" elevation="0">
      <v-card-text>
        <div class="user-row">
          <v-avatar color="primary" size="52" class="mr-3">
            <span class="avatar-letter">{{ (auth.username || t('mine.defaultUser')).slice(0, 1) }}</span>
          </v-avatar>
          <div>
            <div class="user">{{ auth.username || t('mine.defaultUser') }}</div>
            <div class="muted">
              <span v-if="auth.isAdmin()" class="role-admin">{{ t('mine.admin') }}</span>
              <span v-else>{{ t('mine.normalUser') }}</span>
              · {{ auth.profile?.email || t('mine.defaultSub') }}
            </div>
          </div>
        </div>
      </v-card-text>
    </v-card>

    <div class="theme-bar surface-card mb-3 d-md-none">
      <span class="theme-bar-label">{{ t('mine.appearance') }}</span>
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

    <div class="locale-bar surface-card mb-3">
      <span class="theme-bar-label">
        <v-icon size="18" class="mr-1">mdi-translate</v-icon>{{ t('locale.label') }}
      </span>
      <LocaleSwitcher />
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

    <v-btn class="mt-2 logout-btn" block color="error" variant="tonal" size="large" @click="onLogout">{{ t('mine.logout') }}</v-btn>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { getStoredTheme, useAppTheme, type ThemePreference } from '../composables/useAppTheme'
import LocaleSwitcher from '../components/LocaleSwitcher.vue'

const { t } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const { setPreference } = useAppTheme()
const themePref = ref<ThemePreference>(getStoredTheme())

const menuItem = (to: string, icon: string, key: string) => ({
  to, icon, title: t(`mine.menu.${key}`), desc: t(`mine.menu.${key}Desc`),
})

const menu = computed(() => [
  menuItem('/accounts', 'mdi-wallet-outline', 'accounts'),
  menuItem('/categories', 'mdi-shape-outline', 'categories'),
  menuItem('/tags', 'mdi-tag-outline', 'tags'),
  menuItem('/templates', 'mdi-flash-outline', 'templates'),
  menuItem('/fare-rules', 'mdi-ticket-percent-outline', 'fareRules'),
  menuItem('/schedules', 'mdi-calendar-clock-outline', 'schedules'),
  menuItem('/settings', 'mdi-cog-outline', 'settings'),
  ...(auth.isAdmin() ? [menuItem('/admin/users', 'mdi-shield-account-outline', 'users')] : []),
])

const themeOptions = computed<{ title: string; value: ThemePreference; icon: string }[]>(() => [
  { title: t('theme.lightShort'), value: 'light', icon: 'mdi-white-balance-sunny' },
  { title: t('theme.darkShort'), value: 'dark', icon: 'mdi-weather-night' },
  { title: t('theme.systemShort'), value: 'system', icon: 'mdi-theme-light-dark' },
])

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
.locale-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 8px 8px 14px;
}
.locale-bar .theme-bar-label { display: inline-flex; align-items: center; }
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
