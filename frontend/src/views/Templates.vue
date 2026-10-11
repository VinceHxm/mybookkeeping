<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>{{ t('templates.title') }}</h1>
      <v-btn icon="mdi-plus" variant="text" color="primary" @click="openCreate" />
    </div>
    <p class="page-hint">{{ t('templates.hint') }}</p>

    <div class="card-grid">
    <EntityCard
      v-for="tpl in list"
      :key="tpl.id"
      :title="tpl.name"
      :icon="tpl.icon || txTypeIcon(tpl.type)"
      @edit="openEdit(tpl)"
      @delete="askDelete(tpl)"
    >
      <template #badges>
        <v-chip size="x-small" variant="tonal" color="primary">{{ typeLabel(tpl.type) }}</v-chip>
      </template>
      <template #meta>
        <div v-if="tpl.amount">{{ t('templates.metaAmount', { amount: `¥${fenToYuan(tpl.amount)}` }) }}</div>
        <div v-if="tpl.fareRule">{{ t('templates.metaRule', { name: tpl.fareRule.name }) }}</div>
        <div v-if="geoLabel(tpl)">{{ t('templates.metaPlace', { name: geoLabel(tpl) }) }}</div>
        <div v-if="tpl.remark">{{ t('templates.metaRemark', { text: tpl.remark }) }}</div>
      </template>
    </EntityCard>
    </div>
    <div v-if="!list.length" class="empty-tip">{{ t('templates.empty') }}</div>

    <!-- 宽屏：弹窗编辑 -->
    <v-dialog v-if="!isMobile" v-model="dialog" max-width="520" scrollable>
      <v-card>
        <v-card-title>{{ editing ? t('templates.titleEdit') : t('templates.titleNew') }}</v-card-title>
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
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import http from '../api/http'
import EntityCard from '../components/EntityCard.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import TemplateEditForm, { type Template } from '../components/TemplateEditForm.vue'
import { fenToYuan, typeLabel } from '../utils/money'
import { txTypeIcon } from '../utils/icons'
import { useBreakpoint } from '../composables/useBreakpoint'

const { t } = useI18n()
const router = useRouter()
const { isMobile } = useBreakpoint()

const list = ref<Template[]>([])
const dialog = ref(false)
const editing = ref<Template | null>(null)
const dialogFormRef = ref<InstanceType<typeof TemplateEditForm> | null>(null)
const confirmOpen = ref(false)
const pendingDelete = ref<Template | null>(null)
const deleting = ref(false)

function geoLabel(tpl: Template) {
  if (tpl.geoMode === 'route' && tpl.geoName && tpl.geoEndName) return `${tpl.geoName} → ${tpl.geoEndName}`
  return tpl.geoName || (tpl.geoLng != null ? t('templates.pickedLocation') : '')
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

function openEdit(tpl: Template) {
  if (isMobile.value) {
    router.push(`/templates/${tpl.id}/edit`)
    return
  }
  editing.value = tpl
  dialog.value = true
}

async function onDialogSave() {
  await dialogFormRef.value?.save()
}

async function onDialogSaved() {
  dialog.value = false
  await reload()
}

function askDelete(tpl: Template) {
  pendingDelete.value = tpl
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
