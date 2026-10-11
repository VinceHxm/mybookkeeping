<template>
  <v-dialog :model-value="state.open" max-width="380" persistent @update:model-value="(v) => !v && cancel()">
    <v-card class="pwd-card">
      <v-card-title class="title">{{ state.title || t('pwdConfirm.verifyTitle') }}</v-card-title>
      <v-card-text>
        <p v-if="state.message" class="msg">{{ state.message }}</p>
        <form @submit.prevent="submit">
          <input type="text" name="username" autocomplete="username" :value="auth.profile?.username || ''" hidden />
          <v-text-field
            ref="inputRef"
            v-model="password"
            :label="t('pwdConfirm.password')"
            :type="showPwd ? 'text' : 'password'"
            autocomplete="current-password"
            variant="outlined"
            density="comfortable"
            :error-messages="error"
            :append-inner-icon="showPwd ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
            autofocus
            @click:append-inner="showPwd = !showPwd"
            @update:model-value="error = ''"
          />
        </form>
      </v-card-text>
      <v-card-actions class="actions">
        <v-btn variant="text" :disabled="loading" @click="cancel">{{ t('common.cancel') }}</v-btn>
        <v-spacer />
        <v-btn color="primary" variant="flat" :loading="loading" :disabled="!password" @click="submit">{{ t('pwdConfirm.confirm') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { passwordConfirmState as state } from '../composables/usePasswordConfirm'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const auth = useAuthStore()
const password = ref('')
const error = ref('')
const loading = ref(false)
const showPwd = ref(false)

watch(
  () => state.open,
  (open) => {
    if (open) {
      password.value = ''
      error.value = ''
      showPwd.value = false
    }
  },
)

function close(result: unknown) {
  const resolve = state.resolve
  state.open = false
  state.run = null
  state.resolve = null
  password.value = ''
  resolve?.(result)
}

function cancel() {
  if (loading.value) return
  close(null)
}

async function submit() {
  if (!password.value || !state.run || loading.value) return
  loading.value = true
  try {
    const result = await state.run(password.value)
    close(result)
  } catch (e: any) {
    error.value = e?.message || t('pwdConfirm.failed')
    password.value = ''
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.pwd-card { background: rgb(var(--v-theme-surface)) !important; }
.title { font-size: 1.1rem !important; font-weight: 700; }
.msg { margin: 0 0 12px; font-size: 0.85rem; line-height: 1.5; color: var(--muted); }
.actions { padding: 0 16px 16px !important; }
</style>
