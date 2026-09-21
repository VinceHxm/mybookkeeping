<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>记账模板</h1>
      <v-btn icon="mdi-plus" variant="text" color="primary" @click="openCreate" />
    </div>
    <p class="page-hint">记一笔可一键套用。绑定计费规则时，模板金额会作为原价计算建议价。</p>

    <EntityCard
      v-for="t in list"
      :key="t.id"
      :title="t.name"
      :icon="t.icon || txTypeIcon(t.type)"
      @edit="openEdit(t)"
      @delete="askDelete(t)"
    >
      <template #badges>
        <v-chip size="x-small" variant="tonal" color="primary">{{ typeLabel(t.type) }}</v-chip>
      </template>
      <template #meta>
        <div v-if="t.amount">金额 ¥{{ fenToYuan(t.amount) }}</div>
        <div v-if="t.fareRule">规则 {{ t.fareRule.name }}</div>
        <div v-if="geoLabel(t)">地点 {{ geoLabel(t) }}</div>
        <div v-if="t.remark">备注 {{ t.remark }}</div>
      </template>
    </EntityCard>
    <div v-if="!list.length" class="empty-tip">暂无模板，点右上角添加</div>

    <!-- 宽屏：弹窗编辑 -->
    <v-dialog v-if="!isMobile" v-model="dialog" max-width="520" scrollable>
      <v-card>
        <v-card-title>{{ editing ? '编辑模板' : '新建模板' }}</v-card-title>
        <v-card-text>
          <TemplateEditForm
            v-if="dialog"
            ref="dialogFormRef"
            :template-id="editing?.id ?? null"
            :initial="editing"
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
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import http from '../api/http'
import EntityCard from '../components/EntityCard.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import TemplateEditForm, { type Template } from '../components/TemplateEditForm.vue'
import { fenToYuan, typeLabel } from '../utils/money'
import { txTypeIcon } from '../utils/icons'
import { useBreakpoint } from '../composables/useBreakpoint'

const router = useRouter()
const { isMobile } = useBreakpoint()

const list = ref<Template[]>([])
const dialog = ref(false)
const editing = ref<Template | null>(null)
const dialogFormRef = ref<InstanceType<typeof TemplateEditForm> | null>(null)
const confirmOpen = ref(false)
const pendingDelete = ref<Template | null>(null)
const deleting = ref(false)

function geoLabel(t: Template) {
  if (t.geoMode === 'route' && t.geoName && t.geoEndName) return `${t.geoName} → ${t.geoEndName}`
  return t.geoName || (t.geoLng != null ? '已选位置' : '')
}

async function reload() {
  const { data } = await http.get('/templates')
  list.value = data || []
}

onMounted(reload)

function openCreate() {
  if (isMobile.value) {
    router.push('/templates/new')
    return
  }
  editing.value = null
  dialog.value = true
}

function openEdit(t: Template) {
  if (isMobile.value) {
    router.push(`/templates/${t.id}/edit`)
    return
  }
  editing.value = t
  dialog.value = true
}

async function onDialogSave() {
  await dialogFormRef.value?.save()
}

async function onDialogSaved() {
  dialog.value = false
  await reload()
}

function askDelete(t: Template) {
  pendingDelete.value = t
  confirmOpen.value = true
}

async function doDelete() {
  if (!pendingDelete.value?.id) return
  deleting.value = true
  try {
    await http.delete(`/templates/${pendingDelete.value.id}`)
    confirmOpen.value = false
    if (editing.value?.id === pendingDelete.value.id) dialog.value = false
    pendingDelete.value = null
    await reload()
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; }
.toolbar h1 { font-size: 1.15rem; margin: 0; }
.page-hint { color: var(--muted); font-size: 0.85rem; margin: 0 0 12px; line-height: 1.45; }
</style>
