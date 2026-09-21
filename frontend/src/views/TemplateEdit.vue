<template>
  <div class="page edit-page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" @click="goBack" />
      <h1>{{ templateId ? '编辑模板' : '新建模板' }}</h1>
      <v-btn color="primary" variant="text" :loading="formRef?.saving" @click="onSave">保存</v-btn>
    </div>

    <div v-if="loading" class="loading-tip">加载中…</div>
    <TemplateEditForm
      v-else
      ref="formRef"
      :template-id="templateId"
      :initial="initial"
      @saved="onSaved"
      @delete="askDelete"
    />

    <DeleteConfirmDialog
      v-model="confirmOpen"
      :name="initial?.name || '此模板'"
      :loading="deleting"
      @confirm="doDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import http from '../api/http'
import TemplateEditForm, { type Template } from '../components/TemplateEditForm.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'

const route = useRoute()
const router = useRouter()

const formRef = ref<InstanceType<typeof TemplateEditForm> | null>(null)
const loading = ref(true)
const initial = ref<Template | null>(null)
const confirmOpen = ref(false)
const deleting = ref(false)

const templateId = computed(() => {
  const raw = route.params.id
  if (!raw || raw === 'new') return null
  const n = Number(raw)
  return Number.isFinite(n) && n > 0 ? n : null
})

onMounted(async () => {
  loading.value = true
  try {
    if (templateId.value) {
      const { data } = await http.get('/templates')
      initial.value = (data || []).find((t: Template) => t.id === templateId.value) || null
      if (!initial.value) {
        router.replace('/templates')
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
  else router.push('/templates')
}

async function onSave() {
  await formRef.value?.save()
}

function onSaved() {
  router.replace('/templates')
}

function askDelete() {
  confirmOpen.value = true
}

async function doDelete() {
  if (!templateId.value) return
  deleting.value = true
  try {
    await http.delete(`/templates/${templateId.value}`)
    confirmOpen.value = false
    router.replace('/templates')
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
