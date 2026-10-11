<template>
  <div class="page narrow">
    <h1 class="title">{{ t('forgot.title') }}</h1>
    <p class="hint">{{ t('forgot.hint') }}</p>

    <v-stepper v-model="step" alt-labels flat>
      <v-stepper-header>
        <v-stepper-item :value="1" :title="t('forgot.stepToken')" />
        <v-divider />
        <v-stepper-item :value="2" :title="t('forgot.stepPassword')" />
      </v-stepper-header>
    </v-stepper>

    <div v-if="step === 1" class="mt-4">
      <v-text-field v-model="username" :label="t('auth.username')" variant="outlined" />
      <v-btn color="primary" block :loading="loading" @click="onForgot">{{ t('forgot.generate') }}</v-btn>
      <v-alert v-if="token" type="success" class="mt-3" density="compact">
        {{ t('forgot.tokenShown', { token }) }}
      </v-alert>
      <v-alert v-else-if="info" type="info" class="mt-3" density="compact">{{ info }}</v-alert>
    </div>

    <div v-else class="mt-4">
      <v-text-field v-model="token" :label="t('forgot.token')" variant="outlined" />
      <v-text-field v-model="password" :label="t('forgot.newPassword')" type="password" variant="outlined" />
      <v-btn color="primary" block :loading="loading" @click="onReset">{{ t('forgot.confirm') }}</v-btn>
    </div>

    <v-alert v-if="error" type="error" class="mt-3" density="compact">{{ error }}</v-alert>
    <v-alert v-if="ok" type="success" class="mt-3" density="compact">{{ t('forgot.done') }}</v-alert>
    <div class="mt-4"><router-link to="/login">{{ t('forgot.backToLogin') }}</router-link></div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import http from '../api/http'

const { t } = useI18n()
const step = ref(1)
const username = ref('')
const token = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const info = ref('')
const ok = ref(false)

async function onForgot() {
  error.value = ''
  info.value = ''
  loading.value = true
  try {
    const { data } = await http.post('/auth/forgot-password', { username: username.value.trim() })
    token.value = data.resetToken || ''
    if (token.value) step.value = 2
    else info.value = data.message || t('forgot.noToken')
  } catch (e: any) {
    error.value = e.message || t('common.failed')
  } finally {
    loading.value = false
  }
}

async function onReset() {
  error.value = ''
  ok.value = false
  loading.value = true
  try {
    await http.post('/auth/reset-password', { token: token.value, newPassword: password.value })
    ok.value = true
  } catch (e: any) {
    error.value = e.message || t('common.failed')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.narrow { max-width: 480px; margin: 0 auto; padding-top: 40px; }
.hint { color: var(--muted); font-size: 0.9rem; margin-bottom: 12px; }
a { color: var(--primary); }
</style>
