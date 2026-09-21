<template>
  <div class="page edit-page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" @click="goBack" />
      <h1>{{ accountId ? '编辑账户' : '新建账户' }}</h1>
      <v-btn color="primary" variant="text" :loading="formRef?.saving" @click="onSave">保存</v-btn>
    </div>

    <div v-if="loading" class="loading-tip">加载中…</div>
    <AccountEditForm
      v-else
      ref="formRef"
      :account-id="accountId"
      :initial="initial"
      @saved="onSaved"
      @reconciled="onSaved"
      @delete="askDelete"
    />

    <DeleteConfirmDialog
      v-model="confirmOpen"
      :name="initial?.name || '此账户'"
      :loading="deleting"
      @confirm="doDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AccountEditForm from '../components/AccountEditForm.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import { useAccountStore, type Account } from '../stores/account'

const route = useRoute()
const router = useRouter()
const accounts = useAccountStore()

const formRef = ref<InstanceType<typeof AccountEditForm> | null>(null)
const loading = ref(true)
const initial = ref<Account | null>(null)
const confirmOpen = ref(false)
const deleting = ref(false)

const accountId = computed(() => {
  const raw = route.params.id
  if (!raw || raw === 'new') return null
  const n = Number(raw)
  return Number.isFinite(n) && n > 0 ? n : null
})

onMounted(async () => {
  loading.value = true
  try {
    if (accountId.value) {
      await accounts.load(true)
      initial.value = accounts.list.find((a) => a.id === accountId.value) || null
      if (!initial.value) {
        router.replace('/accounts')
        return
      }
    } else {
      initial.value = null
    }
  } finally {
    loading.value = false
  }
})

function goBack() {
  if (window.history.length > 1) router.back()
  else router.push('/accounts')
}

async function onSave() {
  await formRef.value?.save()
}

function onSaved() {
  router.replace('/accounts')
}

function askDelete() {
  confirmOpen.value = true
}

async function doDelete() {
  if (!accountId.value) return
  deleting.value = true
  try {
    await accounts.remove(accountId.value)
    confirmOpen.value = false
    router.replace('/accounts')
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  position: sticky;
  top: 0;
  z-index: 2;
  background: var(--bg, rgb(var(--v-theme-background)));
  padding: 4px 0;
}
.toolbar h1 {
  font-size: 1.15rem;
  margin: 0;
  flex: 1;
  text-align: center;
}
.edit-page {
  padding-bottom: calc(24px + env(safe-area-inset-bottom, 0px));
}
.loading-tip {
  padding: 24px;
  text-align: center;
  color: var(--muted);
}
</style>
