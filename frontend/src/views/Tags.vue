<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>标签管理</h1>
      <v-btn icon="mdi-plus" variant="text" color="primary" @click="openCreate" />
    </div>

    <EntityCard
      v-for="t in tags.list"
      :key="t.id"
      :title="t.name"
      :dot="t.color"
      @edit="openEdit(t)"
      @delete="askDelete(t)"
    >
      <template #meta>标签色用于流水筛选与展示</template>
    </EntityCard>
    <div v-if="!tags.list.length" class="empty-tip">暂无标签</div>

    <v-dialog v-model="dialog" max-width="400">
      <v-card>
        <v-card-title>{{ editing ? '编辑标签' : '新建标签' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.name" label="名称" variant="outlined" hide-details class="mb-3" />
          <div class="color-row mb-3">
            <v-text-field v-model="form.color" label="颜色" type="color" variant="outlined" hide-details />
            <div class="swatches">
              <button
                v-for="c in swatches"
                :key="c"
                type="button"
                class="swatch"
                :class="{ active: form.color === c }"
                :style="{ background: c }"
                @click="form.color = c"
              />
            </div>
          </div>
          <v-text-field v-model.number="form.sort" label="排序" type="number" variant="outlined" hide-details class="mb-6" />

          <div v-if="editing" class="danger-zone">
            <div class="danger-title">危险操作</div>
            <p class="danger-tip">删除标签不可恢复，已关联流水会失去该标签。</p>
            <v-btn color="error" variant="outlined" block @click="askDelete(editing)">删除此标签</v-btn>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">取消</v-btn>
          <v-btn color="primary" @click="onSave">保存</v-btn>
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
import { onMounted, reactive, ref } from 'vue'
import EntityCard from '../components/EntityCard.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import { useTagStore, type Tag } from '../stores/tag'

const tags = useTagStore()
const dialog = ref(false)
const editing = ref<Tag | null>(null)
const confirmOpen = ref(false)
const pendingDelete = ref<Tag | null>(null)
const deleting = ref(false)
const form = reactive({ name: '', color: '#1b7f5a', sort: 0 })
const swatches = ['#1b7f5a', '#3d9b74', '#c45c3e', '#d4a017', '#2e7d6f', '#5c6bc0', '#8e24aa', '#546e7a']

onMounted(() => tags.load())

function openCreate() {
  editing.value = null
  form.name = ''
  form.color = '#1b7f5a'
  form.sort = 0
  dialog.value = true
}
function openEdit(t: Tag) {
  editing.value = t
  form.name = t.name
  form.color = t.color
  form.sort = t.sort
  dialog.value = true
}
async function onSave() {
  if (editing.value) await tags.update(editing.value.id, { ...form })
  else await tags.create({ ...form })
  dialog.value = false
}

function askDelete(t: Tag) {
  pendingDelete.value = t
  confirmOpen.value = true
}

async function doDelete() {
  if (!pendingDelete.value) return
  deleting.value = true
  try {
    await tags.remove(pendingDelete.value.id)
    confirmOpen.value = false
    if (editing.value?.id === pendingDelete.value.id) dialog.value = false
    pendingDelete.value = null
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; }
.toolbar h1 { font-size: 1.15rem; margin: 0; }
.color-row { display: grid; gap: 10px; }
.swatches { display: flex; flex-wrap: wrap; gap: 8px; }
.swatch {
  width: 28px; height: 28px; border-radius: 50%; border: 2px solid transparent;
  cursor: pointer; padding: 0;
}
.swatch.active { border-color: rgb(var(--v-theme-on-surface)); box-shadow: 0 0 0 2px var(--primary-soft); }
.danger-zone {
  margin-top: 8px;
  padding: 16px;
  border-radius: 14px;
  border: 1px solid rgba(198, 40, 40, 0.35);
  background: rgba(198, 40, 40, 0.08);
}
.danger-title { font-weight: 700; color: #c62828; margin-bottom: 6px; }
.danger-tip { margin: 0 0 12px; font-size: 0.8rem; color: var(--muted); line-height: 1.4; }
</style>
