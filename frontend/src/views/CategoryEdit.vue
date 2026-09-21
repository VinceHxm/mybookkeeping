<template>
  <div class="page edit-page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" @click="goBack" />
      <h1>{{ categoryId ? '编辑分类' : '新建分类' }}</h1>
      <v-btn color="primary" variant="text" :loading="formRef?.saving" @click="onSave">保存</v-btn>
    </div>

    <div v-if="loading" class="loading-tip">加载中…</div>
    <CategoryEditForm
      v-else
      ref="formRef"
      :category-id="categoryId"
      :initial="initial"
      :default-kind="defaultKind"
      :default-parent-id="defaultParentId"
      @saved="onSaved"
      @delete="askDelete"
    />

    <DeleteConfirmDialog
      v-model="confirmOpen"
      :name="initial?.name || '此分类'"
      :loading="deleting"
      @confirm="doDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CategoryEditForm from '../components/CategoryEditForm.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import { useCategoryStore, type Category } from '../stores/category'

const route = useRoute()
const router = useRouter()
const categories = useCategoryStore()

const formRef = ref<InstanceType<typeof CategoryEditForm> | null>(null)
const loading = ref(true)
const initial = ref<Category | null>(null)
const confirmOpen = ref(false)
const deleting = ref(false)

const categoryId = computed(() => {
  const raw = route.params.id
  if (!raw || raw === 'new') return null
  const n = Number(raw)
  return Number.isFinite(n) && n > 0 ? n : null
})

const defaultKind = computed(() =>
  route.query.kind === 'income' ? 'income' as const : 'expense' as const,
)
const defaultParentId = computed(() => {
  const p = Number(route.query.parentId)
  return Number.isFinite(p) && p > 0 ? p : null
})

onMounted(async () => {
  loading.value = true
  try {
    await categories.load()
    if (categoryId.value) {
      initial.value = categories.list.find((c) => c.id === categoryId.value) || null
      if (!initial.value) {
        router.replace('/categories')
        return
      }
    } else initial.value = null
  } finally {
    loading.value = false
  }
})

function goBack() {
  if (window.history.length > 1) router.back()
  else router.push('/categories')
}

async function onSave() {
  await formRef.value?.save()
}

function onSaved() {
  router.replace('/categories')
}

function askDelete() {
  confirmOpen.value = true
}

async function doDelete() {
  if (!categoryId.value) return
  deleting.value = true
  try {
    await categories.remove(categoryId.value)
    confirmOpen.value = false
    router.replace('/categories')
  } catch (e: any) {
    alert(e.message || '删除失败')
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
.toolbar h1 { font-size: 1.15rem; margin: 0; flex: 1; text-align: center; }
.edit-page { padding-bottom: calc(24px + env(safe-area-inset-bottom, 0px)); }
.loading-tip { padding: 24px; text-align: center; color: var(--muted); }
</style>
