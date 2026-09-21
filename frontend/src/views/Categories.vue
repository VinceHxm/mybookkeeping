<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>分类管理</h1>
      <v-btn icon="mdi-plus" variant="text" color="primary" @click="openCreate(null)" />
    </div>

    <v-tabs v-model="tab" color="primary" class="mb-3" density="comfortable">
      <v-tab value="expense">支出</v-tab>
      <v-tab value="income">收入</v-tab>
    </v-tabs>

    <template v-for="c in roots" :key="c.id">
      <EntityCard
        :title="c.name"
        :icon="c.icon || 'mdi-shape'"
        @edit="openEdit(c)"
        @delete="askDelete(c)"
      >
        <template #meta>
          一级分类
          <span v-if="childrenOf(c.id).length"> · {{ childrenOf(c.id).length }} 个子分类</span>
        </template>
      </EntityCard>
      <EntityCard
        v-for="ch in childrenOf(c.id)"
        :key="ch.id"
        class="child-card"
        :title="ch.name"
        :icon="ch.icon || c.icon || 'mdi-subdirectory-arrow-right'"
        @edit="openEdit(ch)"
        @delete="askDelete(ch)"
      >
        <template #meta>属于「{{ c.name }}」</template>
      </EntityCard>
    </template>
    <div v-if="!roots.length" class="empty-tip">暂无{{ tab === 'expense' ? '支出' : '收入' }}分类</div>

    <v-dialog v-if="!isMobile" v-model="dialog" max-width="440" scrollable>
      <v-card>
        <v-card-title>{{ editing ? '编辑分类' : '新建分类' }}</v-card-title>
        <v-card-text>
          <CategoryEditForm
            v-if="dialog"
            ref="dialogFormRef"
            :category-id="editing?.id ?? null"
            :initial="editing"
            :default-kind="(tab as 'expense' | 'income')"
            :default-parent-id="createParentId"
            @saved="onDialogSaved"
            @delete="editing && askDelete(editing)"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">取消</v-btn>
          <v-btn color="primary" :loading="dialogFormRef?.saving" @click="onDialogSave">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <DeleteConfirmDialog
      v-model="confirmOpen"
      :name="pendingDelete?.name || ''"
      :loading="deleting"
      @confirm="doDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import EntityCard from '../components/EntityCard.vue'
import CategoryEditForm from '../components/CategoryEditForm.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import { useCategoryStore, type Category } from '../stores/category'
import { useBreakpoint } from '../composables/useBreakpoint'

const categories = useCategoryStore()
const router = useRouter()
const { isMobile } = useBreakpoint()

const tab = ref('expense')
const dialog = ref(false)
const editing = ref<Category | null>(null)
const createParentId = ref<number | null>(null)
const dialogFormRef = ref<InstanceType<typeof CategoryEditForm> | null>(null)
const confirmOpen = ref(false)
const pendingDelete = ref<Category | null>(null)
const deleting = ref(false)

const roots = computed(() =>
  categories.list.filter((c) => c.kind === tab.value && !c.parentId),
)

function childrenOf(id: number) {
  return categories.list.filter((c) => c.parentId === id)
}

onMounted(() => categories.load())

function openCreate(parentId: number | null) {
  if (isMobile.value) {
    const q: Record<string, string> = { kind: tab.value }
    if (parentId) q.parentId = String(parentId)
    router.push({ path: '/categories/new', query: q })
    return
  }
  editing.value = null
  createParentId.value = parentId
  dialog.value = true
}

function openEdit(c: Category) {
  if (isMobile.value) {
    router.push(`/categories/${c.id}/edit`)
    return
  }
  editing.value = c
  createParentId.value = null
  dialog.value = true
}

async function onDialogSave() {
  await dialogFormRef.value?.save()
}

async function onDialogSaved() {
  dialog.value = false
  await categories.load()
}

function askDelete(c: Category) {
  pendingDelete.value = c
  confirmOpen.value = true
}

async function doDelete() {
  if (!pendingDelete.value) return
  deleting.value = true
  try {
    await categories.remove(pendingDelete.value.id)
    confirmOpen.value = false
    if (editing.value?.id === pendingDelete.value.id) dialog.value = false
    pendingDelete.value = null
  } catch (e: any) {
    alert(e.message || '删除失败')
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; }
.toolbar h1 { font-size: 1.15rem; margin: 0; }
.child-card { margin-left: 18px; }
</style>
