<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>设置</h1>
      <span style="width:40px" />
    </div>

    <h2 class="sec">外观</h2>
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

    <h2 class="sec">偏好</h2>
    <v-select
      v-model="form.defaultAccountId"
      :items="[{ id: null, name: '不指定' }, ...accounts.list]"
      item-title="name"
      item-value="id"
      label="默认账户（还款时优先作转出储蓄卡）"
      variant="outlined"
      class="mb-2"
    />
    <v-select
      v-model="form.weekStart"
      :items="[{ title: '周一', value: 1 }, { title: '周日', value: 0 }]"
      label="一周起始"
      variant="outlined"
      class="mb-2"
    />
    <div class="colors mb-3">
      <v-text-field v-model="form.expenseColor" label="支出颜色" type="color" variant="outlined" />
      <v-text-field v-model="form.incomeColor" label="收入颜色" type="color" variant="outlined" />
    </div>
    <v-text-field v-model="form.email" label="邮箱" variant="outlined" class="mb-2" />
    <v-btn color="primary" block class="mb-6" :loading="saving" @click="saveSettings">保存偏好</v-btn>

    <h2 class="sec">修改密码</h2>
    <v-text-field v-model="oldPassword" label="原密码" type="password" variant="outlined" />
    <v-text-field v-model="newPassword" label="新密码" type="password" variant="outlined" />
    <v-alert v-if="msg" :type="msgType" density="compact" class="mb-2">{{ msg }}</v-alert>
    <v-btn color="secondary" block :loading="pwdLoading" @click="changePwd">更新密码</v-btn>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useAccountStore } from '../stores/account'
import { useAppTheme, type ThemePreference } from '../composables/useAppTheme'

const auth = useAuthStore()
const accounts = useAccountStore()
const { setPreference } = useAppTheme()
const saving = ref(false)
const pwdLoading = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const msg = ref('')
const msgType = ref<'success' | 'error'>('success')

const themeOptions: { title: string; value: ThemePreference; icon: string }[] = [
  { title: '浅色', value: 'light', icon: 'mdi-white-balance-sunny' },
  { title: '深色', value: 'dark', icon: 'mdi-weather-night' },
  { title: '跟随系统', value: 'system', icon: 'mdi-theme-light-dark' },
]

const form = reactive({
  defaultAccountId: null as number | null,
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
  await Promise.all([accounts.load(), auth.fetchMe()])
  const p = auth.profile
  if (p) {
    form.defaultAccountId = p.defaultAccountId ?? null
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
      defaultAccountId: form.defaultAccountId,
      clearDefaultAccount: form.defaultAccountId == null,
      weekStart: form.weekStart,
      theme: form.theme,
      expenseColor: form.expenseColor,
      incomeColor: form.incomeColor,
    })
    applyLocalTheme()
    msg.value = '已保存'
    msgType.value = 'success'
  } catch (e: any) {
    msg.value = e.message || '保存失败'
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
    msg.value = '密码已更新'
    msgType.value = 'success'
    oldPassword.value = ''
    newPassword.value = ''
  } catch (e: any) {
    msg.value = e.message || '失败'
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
</style>
