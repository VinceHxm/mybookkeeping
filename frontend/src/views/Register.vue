<template>
  <div class="page narrow">
    <h1 class="title">注册</h1>
    <v-text-field v-model="username" label="用户名" variant="outlined" />
    <v-text-field v-model="email" label="邮箱（可选）" variant="outlined" />
    <v-text-field v-model="password" label="密码（至少 8 位）" type="password" variant="outlined" />
    <v-text-field v-model="password2" label="确认密码" type="password" variant="outlined" @keyup.enter="onSubmit" />
    <v-alert v-if="error" type="error" density="compact" class="mb-3">{{ error }}</v-alert>
    <v-btn color="primary" block size="large" :loading="loading" @click="onSubmit">注册并登录</v-btn>
    <div class="mt-4"><router-link to="/login">已有账号？去登录</router-link></div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

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
    error.value = '两次密码不一致'
    return
  }
  loading.value = true
  try {
    await auth.register(username.value.trim(), password.value, email.value.trim())
    await router.replace('/')
  } catch (e: any) {
    error.value = e.message || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.narrow { max-width: 420px; margin: 0 auto; padding-top: 48px; }
.title { margin-bottom: 16px; }
a { color: var(--primary); }
</style>
