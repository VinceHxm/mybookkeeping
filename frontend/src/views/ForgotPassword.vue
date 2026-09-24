<template>
  <div class="page narrow">
    <h1 class="title">重置密码</h1>
    <p class="hint">默认不开放自助找回，请联系管理员在「用户管理」中为你重置密码。</p>

    <v-stepper v-model="step" alt-labels flat>
      <v-stepper-header>
        <v-stepper-item :value="1" title="获取令牌" />
        <v-divider />
        <v-stepper-item :value="2" title="设置新密码" />
      </v-stepper-header>
    </v-stepper>

    <div v-if="step === 1" class="mt-4">
      <v-text-field v-model="username" label="用户名" variant="outlined" />
      <v-btn color="primary" block :loading="loading" @click="onForgot">生成重置令牌</v-btn>
      <v-alert v-if="token" type="success" class="mt-3" density="compact">
        令牌：{{ token }}（已自动填入下一步）
      </v-alert>
      <v-alert v-else-if="info" type="info" class="mt-3" density="compact">{{ info }}</v-alert>
    </div>

    <div v-else class="mt-4">
      <v-text-field v-model="token" label="重置令牌" variant="outlined" />
      <v-text-field v-model="password" label="新密码" type="password" variant="outlined" />
      <v-btn color="primary" block :loading="loading" @click="onReset">确认重置</v-btn>
    </div>

    <v-alert v-if="error" type="error" class="mt-3" density="compact">{{ error }}</v-alert>
    <v-alert v-if="ok" type="success" class="mt-3" density="compact">密码已重置，请登录</v-alert>
    <div class="mt-4"><router-link to="/login">返回登录</router-link></div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import http from '../api/http'

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
    else info.value = data.message || '请联系管理员重置密码'
  } catch (e: any) {
    error.value = e.message || '失败'
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
    error.value = e.message || '失败'
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
