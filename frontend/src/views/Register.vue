<template>
  <div class="page narrow">
    <div class="head">
      <h1 class="title">{{ t('auth.register') }}</h1>
      <LocaleSwitcher />
    </div>
    <v-text-field v-model="username" :label="t('auth.username')" variant="outlined" />
    <v-text-field v-model="email" :label="t('auth.emailOptional')" variant="outlined" />
    <v-text-field v-model="password" :label="t('auth.passwordMin', { n: 8 })" type="password" variant="outlined" />
    <v-text-field v-model="password2" :label="t('auth.confirmPassword')" type="password" variant="outlined" @keyup.enter="onSubmit" />
    <v-alert v-if="error" type="error" density="compact" class="mb-3">{{ error }}</v-alert>
    <v-btn color="primary" block size="large" :loading="loading" @click="onSubmit">{{ t('auth.registerAndLogin') }}</v-btn>
    <div class="mt-4"><router-link to="/login">{{ t('auth.haveAccount') }}</router-link></div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import LocaleSwitcher from '../components/LocaleSwitcher.vue'

const { t } = useI18n()
const auth = useAuthStore()
const router = useRouter()
const username = ref('')
const email = ref('')
const password = ref('')
const password2 = ref('')
const loading = ref(false)
const error = ref('')

async function onSubmit() {
  error.value = ''
  if (password.value !== password2.value) {
    error.value = t('auth.passwordMismatch')
    return
  }
  loading.value = true
  try {
    await auth.register(username.value.trim(), password.value, email.value.trim())
    await router.replace('/')
  } catch (e: any) {
    error.value = e.message || t('auth.registerFailed')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.narrow { max-width: 420px; margin: 0 auto; padding-top: 48px; }
.head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.title { margin: 0; }
a { color: var(--primary); }
</style>
