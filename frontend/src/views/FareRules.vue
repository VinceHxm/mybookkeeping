<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>{{ t('fare.title') }}</h1>
      <v-btn icon="mdi-plus" variant="text" color="primary" @click="openCreate" />
    </div>
    <p class="page-hint">{{ t('fare.hint') }}</p>

    <div class="card-grid">
    <EntityCard
      v-for="r in list"
      :key="r.id"
      :title="r.name"
      :icon="r.icon || 'mdi-ticket-percent'"
      :dim="!r.enabled"
      @edit="openEdit(r)"
      @delete="askDelete(r)"
    >
      <template #badges>
        <v-chip v-if="!r.enabled" size="x-small" variant="tonal">{{ t('fare.disabled') }}</v-chip>
        <v-chip v-if="r.city" size="x-small" variant="tonal" color="primary">{{ r.city }}</v-chip>
        <v-chip v-if="r.freeOnHoliday" size="x-small" variant="tonal" color="success">{{ t('fare.holidayFree') }}</v-chip>
      </template>
      <template #meta>
        <div>{{ r.baseFen ? t('fare.defaultBase', { amount: `¥${fenToYuan(r.baseFen)}` }) : t('fare.baseFromTemplate') }}</div>
        <div>
          <span v-if="r.cardRate && r.cardRate !== 1">{{ t('fare.cardRate', { zhe: rateToZhe(r.cardRate), pct: ratePct(r.cardRate) }) }} · </span>
          {{ cycleLabel(r) }}
          <span v-if="r.tiers?.length"> · {{ t('fare.amountTiers', { n: r.tiers.length }) }}</span>
          <span v-if="r.countTiers?.length"> · {{ t('fare.countTiers', { n: r.countTiers.length }) }}</span>
          <span v-if="r.timeWindows?.length"> · {{ t('fare.timeWindows', { n: r.timeWindows.length }) }}</span>
          <span v-if="r.freePeriods?.length"> · {{ t('fare.freeDays', { n: r.freePeriods.length }) }}</span>
        </div>
      </template>
    </EntityCard>
    </div>
    <p v-if="!list.length" class="empty">{{ t('fare.empty') }}</p>

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
import { fenToYuan } from '../utils/money'
import { cycleLabel, ratePct, rateToZhe, type FareRule } from '../utils/fare'

const { t } = useI18n()
const router = useRouter()
const list = ref<FareRule[]>([])
const confirmOpen = ref(false)
const pendingDelete = ref<FareRule | null>(null)
const deleting = ref(false)

async function reload() {
  const { data } = await http.get('/fare-rules')
  list.value = data || []
}

onMounted(reload)

function openCreate() {
  router.push('/fare-rules/new')
}

function openEdit(r: FareRule) {
  router.push(`/fare-rules/${r.id}/edit`)
}

function askDelete(r: FareRule) {
  pendingDelete.value = r
  confirmOpen.value = true
}

async function doDelete() {
  if (!pendingDelete.value?.id) return
  deleting.value = true
  try {
    await http.delete(`/fare-rules/${pendingDelete.value.id}`)
    confirmOpen.value = false
    pendingDelete.value = null
    await reload()
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; }
.toolbar h1 { font-size: 1.15rem; margin: 0; font-family: var(--font-display); }
.page-hint { color: var(--muted); font-size: 0.85rem; margin: 0 0 12px; line-height: 1.45; }
.empty { color: var(--muted); text-align: center; margin-top: 32px; }
</style>
