<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>{{ t('schedules.title') }}</h1>
      <v-btn icon="mdi-plus" variant="text" color="primary" @click="openCreate" />
    </div>
    <p class="hint">{{ t('schedules.hint') }}</p>

    <div class="card-grid">
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
        >{{ s.enabled ? t('schedules.enabledChip') : t('schedules.disabledChip') }}</v-chip>
      </template>
      <template #meta>
        <div>
          {{ freqLabel(s.frequency, s.intervalN) }} · {{ s.amountPending ? t('schedules.estimated', { amount: `¥${fenToYuan(s.amount)}` }) : `¥${fenToYuan(s.amount)}` }}
          <template v-if="s.foreignCurrency"> · {{ formatForeign(s.foreignCurrency, s.foreignAmount) }}</template>
          <template v-if="s.amountPending"> · {{ t('schedules.amountPending') }}</template>
        </div>
        <div>{{ t('schedules.nextRun', { time: formatNext(s.nextRunAt) }) }}</div>
      </template>
    </EntityCard>
    </div>
    <div v-if="!list.length" class="empty-tip">{{ t('schedules.empty') }}</div>

    <v-dialog v-if="!isMobile" v-model="dialog" max-width="500" scrollable>
      <v-card>
        <v-card-title>{{ editing ? t('schedules.titleEdit') : t('schedules.titleNew') }}</v-card-title>
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
import dayjs from 'dayjs'
import http from '../api/http'
import EntityCard from '../components/EntityCard.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import ScheduleEditForm from '../components/ScheduleEditForm.vue'
import { fenToYuan, formatForeign } from '../utils/money'
import { txTypeIcon } from '../utils/icons'
import { useBreakpoint } from '../composables/useBreakpoint'

const { t } = useI18n()
const router = useRouter()
const { isMobile } = useBreakpoint()
const list = ref<any[]>([])
const dialog = ref(false)
const editing = ref<any>(null)
const dialogFormRef = ref<InstanceType<typeof ScheduleEditForm> | null>(null)
const confirmOpen = ref(false)
const pendingDelete = ref<any>(null)
const deleting = ref(false)

const FREQS = ['daily', 'weekly', 'monthly', 'yearly']

function freqLabel(f: string, n: number) {
  if (f === 'every_n_days') return t('schedules.freq.everyNDays', { n })
  if (!FREQS.includes(f)) return f
  if (n > 1) return t(`schedules.freq.everyN.${f}`, { n })
  return t(`schedules.freq.every.${f}`)
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
