<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>{{ t('settings.title') }}</h1>
      <span style="width:40px" />
    </div>

    <div class="settings-grid">
    <section class="set-block">
    <h2 class="sec">{{ t('settings.appearance') }}</h2>
    <div class="theme-grid mb-4">
      <button
        v-for="opt in themeOptions"
        :key="opt.value"
        type="button"
        class="theme-opt"
        :class="{ active: form.theme === opt.value }"
        @click="form.theme = opt.value"
      >
        <v-icon size="22">{{ opt.icon }}</v-icon>
        <span>{{ opt.title }}</span>
      </button>
    </div>

    <h2 class="sec">{{ t('locale.label') }}</h2>
    <div class="locale-grid">
      <button
        v-for="opt in SUPPORTED_LOCALES"
        :key="opt.value"
        type="button"
        class="theme-opt"
        :class="{ active: locale === opt.value }"
        @click="changeLocale(opt.value)"
      >
        <span>{{ opt.label }}</span>
      </button>
    </div>
    <p class="pref-tip mb-4">{{ t('settings.localeTip') }}</p>

    <h2 class="sec">{{ t('settings.preferences') }}</h2>
    <v-select
      v-model="form.weekStart"
      :items="[{ title: t('settings.monday'), value: 1 }, { title: t('settings.sunday'), value: 0 }]"
      :label="t('settings.weekStart')"
      variant="outlined"
      class="mb-2"
    />
    <div class="colors mb-3">
      <v-text-field v-model="form.expenseColor" :label="t('settings.expenseColor')" type="color" variant="outlined" />
      <v-text-field v-model="form.incomeColor" :label="t('settings.incomeColor')" type="color" variant="outlined" />
    </div>
    <v-text-field v-model="form.email" :label="t('settings.email')" variant="outlined" class="mb-2" />
    <p class="pref-tip mb-3">{{ t('settings.defaultAccountTip') }}</p>
    <v-btn color="primary" block class="mb-6 set-submit" :loading="saving" @click="saveSettings">{{ t('settings.savePrefs') }}</v-btn>
    </section>

    <section class="set-block">
    <h2 class="sec">{{ t('settings.changePassword') }}</h2>
    <v-text-field v-model="oldPassword" :label="t('settings.oldPassword')" type="password" variant="outlined" />
    <v-text-field v-model="newPassword" :label="t('settings.newPassword')" type="password" variant="outlined" />
    <v-alert v-if="msg" :type="msgType" density="compact" class="mb-2">{{ msg }}</v-alert>
    <v-btn color="secondary" block class="set-submit" :loading="pwdLoading" @click="changePwd">{{ t('settings.updatePassword') }}</v-btn>
    </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useAppTheme, type ThemePreference } from '../composables/useAppTheme'
import { SUPPORTED_LOCALES, setLocale, type AppLocale } from '../i18n'

const { t, locale } = useI18n()
const auth = useAuthStore()
const { setPreference } = useAppTheme()
const saving = ref(false)
const pwdLoading = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const msg = ref('')
const msgType = ref<'success' | 'error'>('success')

const themeOptions = computed<{ title: string; value: ThemePreference; icon: string }[]>(() => [
  { title: t('theme.light'), value: 'light', icon: 'mdi-white-balance-sunny' },
  { title: t('theme.dark'), value: 'dark', icon: 'mdi-weather-night' },
  { title: t('theme.system'), value: 'system', icon: 'mdi-theme-light-dark' },
])

/** 语言与主题预览不同：选中即保存，避免刷新后又变回去 */
function changeLocale(l: AppLocale) {
  if (l === locale.value) return
  setLocale(l)
  auth.updateSettings({ locale: l }).catch(() => {})
}

const form = reactive({
  weekStart: 1,
  theme: 'light' as ThemePreference,
  expenseColor: '#c45c3e',
  incomeColor: '#1b7f5a',
  email: '',
})

function applyLocalTheme() {
  setPreference(form.theme, {
    expenseColor: form.expenseColor,
    incomeColor: form.incomeColor,
  })
}

onMounted(async () => {
  await auth.fetchMe()
  const p = auth.profile
  if (p) {
    form.weekStart = p.weekStart ?? 1
    form.theme = (p.theme as ThemePreference) || 'system'
    form.expenseColor = p.expenseColor || '#c45c3e'
    form.incomeColor = p.incomeColor || '#1b7f5a'
    form.email = p.email || ''
    applyLocalTheme()
  }
})

watch(() => [form.theme, form.expenseColor, form.incomeColor], applyLocalTheme)

async function saveSettings() {
  saving.value = true
  msg.value = ''
  try {
    await auth.updateSettings({
      email: form.email,
      weekStart: form.weekStart,
      theme: form.theme,
      expenseColor: form.expenseColor,
      incomeColor: form.incomeColor,
    })
    applyLocalTheme()
    msg.value = t('settings.saved')
    msgType.value = 'success'
  } catch (e: any) {
    msg.value = e.message || t('settings.saveFailed')
    msgType.value = 'error'
  } finally {
    saving.value = false
  }
}

async function changePwd() {
  msg.value = ''
  pwdLoading.value = true
  try {
    await auth.changePassword(oldPassword.value, newPassword.value)
    msg.value = t('settings.passwordUpdated')
    msgType.value = 'success'
    oldPassword.value = ''
    newPassword.value = ''
  } catch (e: any) {
    msg.value = e.message || t('common.failed')
    msgType.value = 'error'
  } finally {
    pwdLoading.value = false
  }
}
</script>

<style scoped>
.sec { font-size: 0.95rem; color: var(--muted); margin: 8px 0 12px; font-weight: 600; }
.colors { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.theme-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}
.locale-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  margin-bottom: 8px;
}
@media (max-width: 420px) {
  .locale-grid { grid-template-columns: repeat(2, 1fr); }
}
.theme-opt {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 14px 8px;
  border-radius: var(--radius-md);
  border: 1px solid var(--surface-border);
  background: var(--surface);
  color: var(--ink);
  cursor: pointer;
  font-size: 0.85rem;
  transition: border-color 0.15s ease, box-shadow 0.15s ease, background 0.15s ease;
}
.theme-opt.active {
  border-color: var(--primary);
  background: var(--primary-soft);
  box-shadow: 0 0 0 1px var(--primary);
  font-weight: 600;
}
.pref-tip {
  margin: 0;
  font-size: 0.82rem;
  color: var(--muted);
  line-height: 1.45;
}

/* 宽屏：偏好与改密两张卡并排，表单不再拉满整屏 */
@media (min-width: 768px) {
  .settings-grid {
    display: grid;
    grid-template-columns: minmax(0, 3fr) minmax(0, 2fr);
    gap: 20px;
    align-items: start;
    max-width: 980px;
  }
  .set-block {
    padding: 18px 20px 20px;
    border-radius: var(--radius-lg);
    background: var(--surface);
    border: 1px solid var(--surface-border);
    box-shadow: var(--shadow-soft);
  }
  .set-block .sec { margin-top: 0; }
  .set-submit { margin-bottom: 0 !important; }
}
</style>
