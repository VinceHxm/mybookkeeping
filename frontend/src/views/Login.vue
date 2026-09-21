<template>
  <div class="login-shell">
    <div class="login-hero d-none d-md-flex">
      <div class="hero-inner">
        <div class="logo">记</div>
        <h1>我的账本</h1>
        <p>个人精简记账 · 快记优先</p>
      </div>
    </div>
    <div class="login-panel">
      <div class="panel-inner">
        <div class="brand d-md-none">
          <div class="logo">记</div>
          <h1>我的账本</h1>
        </div>
        <h2 class="panel-title">登录</h2>
        <v-card class="form surface-card" elevation="0" rounded="lg">
          <v-card-text>
            <v-text-field v-model="username" label="用户名" variant="outlined" density="comfortable" />
            <v-text-field v-model="password" label="密码" type="password" variant="outlined" density="comfortable" @keyup.enter="onLogin" />
            <v-alert v-if="error" type="error" density="compact" class="mb-3">{{ error }}</v-alert>
            <v-btn color="primary" block size="large" :loading="loading" class="hero-cta" @click="onLogin">登录</v-btn>
            <div class="links">
              <router-link to="/register">注册账号</router-link>
              <router-link to="/forgot-password">忘记密码</router-link>
            </div>
          </v-card-text>
        </v-card>
        <div class="login-theme">
          <v-btn
            v-for="opt in themeOptions"
            :key="opt.value"
            size="small"
            :variant="themePref === opt.value ? 'tonal' : 'text'"
            color="primary"
            @click="setTheme(opt.value)"
          >
            <v-icon start size="16">{{ opt.icon }}</v-icon>
            {{ opt.title }}
          </v-btn>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { getStoredTheme, useAppTheme, type ThemePreference } from '../composables/useAppTheme'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const { setPreference, syncFromProfile } = useAppTheme()
const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const themePref = ref<ThemePreference>(getStoredTheme())

const themeOptions: { title: string; value: ThemePreference; icon: string }[] = [
  { title: '浅色', value: 'light', icon: 'mdi-white-balance-sunny' },
  { title: '深色', value: 'dark', icon: 'mdi-weather-night' },
  { title: '系统', value: 'system', icon: 'mdi-theme-light-dark' },
]

function setTheme(pref: ThemePreference) {
  themePref.value = pref
  setPreference(pref)
}

onMounted(() => {
  themePref.value = getStoredTheme()
  setPreference(themePref.value)
})

async function onLogin() {
  error.value = ''
  loading.value = true
  try {
    const data = await auth.login(username.value.trim(), password.value)
    syncFromProfile(data.user)
    await router.replace((route.query.redirect as string) || '/')
  } catch (e: any) {
    error.value = e.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-shell { min-height: 100%; display: grid; grid-template-columns: 1fr; }
@media (min-width: 960px) { .login-shell { grid-template-columns: 1.15fr 1fr; } }
.login-hero {
  align-items: center; justify-content: center; padding: 48px; color: #fff;
  background:
    radial-gradient(ellipse at 30% 20%, rgba(255, 255, 255, 0.12), transparent 55%),
    linear-gradient(160deg, #145a41 0%, #1b7f5a 45%, #2a6b52 100%);
}
.hero-inner { max-width: 420px; }
.logo {
  width: 72px; height: 72px; margin-bottom: 16px; border-radius: 20px;
  background: rgba(255,255,255,0.16); color: #fff; font-size: 36px; font-weight: 700;
  font-family: var(--font-display);
  display: grid; place-items: center;
}
.login-hero h1 {
  margin: 0;
  font-size: 2.5rem;
  font-family: var(--font-display);
  letter-spacing: -0.03em;
}
.login-hero p { opacity: 0.88; margin-top: 10px; }
.login-panel { display: flex; align-items: center; justify-content: center; padding: 32px 20px 48px; }
.panel-inner { width: 100%; max-width: 440px; }
.panel-title {
  margin: 0 0 16px;
  font-size: 1.7rem;
  font-family: var(--font-display);
  letter-spacing: -0.02em;
}
.brand { text-align: center; margin-bottom: 24px; }
.brand h1 { font-family: var(--font-display); margin: 0; }
.brand .logo { margin: 0 auto 12px; background: linear-gradient(145deg, #1b7f5a, #145a41); }
.form { border: 1px solid var(--surface-border); }
.links { display: flex; justify-content: space-between; margin-top: 16px; font-size: 0.9rem; }
.login-theme {
  display: flex;
  justify-content: center;
  gap: 4px;
  margin-top: 20px;
  flex-wrap: wrap;
}
</style>
