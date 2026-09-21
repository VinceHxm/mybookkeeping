<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>周期记账</h1>
      <v-btn icon="mdi-plus" variant="text" color="primary" @click="openCreate" />
    </div>
    <p class="hint">后台每分钟检查到期项并自动生成流水</p>

    <EntityCard
      v-for="s in list"
      :key="s.id"
      :title="s.name"
      :icon="txTypeIcon(s.type)"
      :dim="!s.enabled"
      @edit="openEdit(s)"
      @delete="askDelete(s)"
    >
      <template #badges>
        <v-chip
          size="x-small"
          class="no-card-nav"
          :color="s.enabled ? 'primary' : undefined"
          variant="tonal"
          @click.stop="toggle(s, !s.enabled)"
          @pointerdown.stop
        >{{ s.enabled ? '启用中 · 点按停用' : '已停用 · 点按启用' }}</v-chip>
      </template>
      <template #meta>
        <div>{{ freqLabel(s.frequency, s.intervalN) }} · ¥{{ fenToYuan(s.amount) }}</div>
        <div>下次执行 {{ formatNext(s.nextRunAt) }}</div>
      </template>
    </EntityCard>
    <div v-if="!list.length" class="empty-tip">暂无周期记账</div>

    <v-dialog v-if="!isMobile" v-model="dialog" max-width="500" scrollable>
      <v-card>
        <v-card-title>{{ editing ? '编辑周期' : '新建周期' }}</v-card-title>
        <v-card-text>
          <ScheduleEditForm
            v-if="dialog"
            ref="dialogFormRef"
            :schedule-id="editing?.id ?? null"
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
import dayjs from 'dayjs'
import http from '../api/http'
import EntityCard from '../components/EntityCard.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import ScheduleEditForm from '../components/ScheduleEditForm.vue'
import { fenToYuan } from '../utils/money'
import { txTypeIcon } from '../utils/icons'
import { useBreakpoint } from '../composables/useBreakpoint'

const router = useRouter()
const { isMobile } = useBreakpoint()
const list = ref<any[]>([])
const dialog = ref(false)
const editing = ref<any>(null)
const dialogFormRef = ref<InstanceType<typeof ScheduleEditForm> | null>(null)
const confirmOpen = ref(false)
const pendingDelete = ref<any>(null)
const deleting = ref(false)

function freqLabel(f: string, n: number) {
  const m: Record<string, string> = { daily: '每天', weekly: '每周', monthly: '每月', yearly: '每年', every_n_days: `每 ${n} 天` }
  return m[f] || f
}
function formatNext(s: string) {
  return dayjs(s).format('YYYY-MM-DD HH:mm')
}

async function reload() {
  const { data } = await http.get('/schedules')
  list.value = data
}

onMounted(reload)

function openCreate() {
  if (isMobile.value) {
    router.push('/schedules/new')
    return
  }
  editing.value = null
  dialog.value = true
}

function openEdit(s: any) {
  if (isMobile.value) {
    router.push(`/schedules/${s.id}/edit`)
    return
  }
  editing.value = s
  dialog.value = true
}

async function onDialogSave() {
  await dialogFormRef.value?.save()
}

async function onDialogSaved() {
  dialog.value = false
  await reload()
}

async function toggle(s: any, enabled: boolean) {
  await http.put(`/schedules/${s.id}`, { ...s, enabled })
  await reload()
}

function askDelete(s: any) {
  pendingDelete.value = s
  confirmOpen.value = true
}

async function doDelete() {
  if (!pendingDelete.value?.id) return
  deleting.value = true
  try {
    await http.delete(`/schedules/${pendingDelete.value.id}`)
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
.hint { color: var(--muted); font-size: 0.85rem; margin: 0 0 12px; }
</style>
