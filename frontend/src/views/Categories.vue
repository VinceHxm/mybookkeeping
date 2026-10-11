<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>{{ t('categories.title') }}</h1>
      <v-btn icon="mdi-plus" variant="text" color="primary" @click="openCreate(null)" />
    </div>

    <v-tabs v-model="tab" color="primary" class="mb-3" density="comfortable">
      <v-tab value="expense">{{ t('txType.expense') }}</v-tab>
      <v-tab value="income">{{ t('txType.income') }}</v-tab>
    </v-tabs>

    <div class="cat-columns">
    <section v-for="c in roots" :key="c.id" class="cat-group">
      <EntityCard
        :title="c.name"
        :icon="c.icon || 'mdi-shape'"
        @edit="openEdit(c)"
        @delete="askDelete(c)"
      >
        <template #meta>
          {{ t('categories.root') }}
          <span v-if="childrenOf(c.id).length"> · {{ t('categories.childCount', { n: childrenOf(c.id).length }) }}</span>
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
        <template #meta>{{ t('categories.belongsTo', { name: c.name }) }}</template>
      </EntityCard>
    </section>
    </div>
    <div v-if="!roots.length" class="empty-tip">
      {{ tab === 'expense' ? t('categories.emptyExpense') : t('categories.emptyIncome') }}
    </div>

    <v-dialog v-if="!isMobile" v-model="dialog" max-width="440" scrollable>
      <v-card>
        <v-card-title>{{ editing ? t('categories.edit') : t('categories.create') }}</v-card-title>
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
          <v-btn variant="text" @click="dialog = false">{{ t('common.cancel') }}</v-btn>
          <v-btn color="primary" :loading="dialogFormRef?.saving" @click="onDialogSave">{{ t('common.save') }}</v-btn>
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
import { useI18n } from 'vue-i18n'
import EntityCard from '../components/EntityCard.vue'
import CategoryEditForm from '../components/CategoryEditForm.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import { useCategoryStore, type Category } from '../stores/category'
import { useBreakpoint } from '../composables/useBreakpoint'

const { t } = useI18n()
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
    alert(e.message || t('ui.deleteFailed'))
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; }
.toolbar h1 { font-size: 1.15rem; margin: 0; }
.child-card { margin-left: 18px; }

/* 宽屏：每个一级分类连同子类自成一组；各组子类数不同、高度不一，
   用多栏（瀑布流）自上而下填充，避免网格同行等高留下的空洞 */
@media (min-width: 768px) {
  .cat-columns {
    columns: 320px;
    column-gap: 14px;
  }
  .cat-group {
    break-inside: avoid;
    padding: 10px 10px 0;
    margin-bottom: 14px;
    border-radius: var(--radius-lg);
    border: 1px dashed var(--surface-border);
    background: color-mix(in srgb, var(--surface) 40%, transparent);
  }
  .child-card { margin-left: 14px; }
}
</style>
