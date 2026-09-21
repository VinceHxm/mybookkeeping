<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>计费规则</h1>
      <v-btn icon="mdi-plus" variant="text" color="primary" @click="openCreate" />
    </div>
    <p class="page-hint">先定名称与基础折扣，再按需打开免费日、阶梯、时段等模块。原价可空；记账手填金额始终为准。</p>

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
        <v-chip v-if="!r.enabled" size="x-small" variant="tonal">已停用</v-chip>
        <v-chip v-if="r.city" size="x-small" variant="tonal" color="primary">{{ r.city }}</v-chip>
        <v-chip v-if="r.freeOnHoliday" size="x-small" variant="tonal" color="success">节假日免</v-chip>
      </template>
      <template #meta>
        <div>{{ r.baseFen ? `默认原价 ¥${fenToYuan(r.baseFen)}` : '原价由模板 / 记一笔提供' }}</div>
        <div>
          <span v-if="r.cardRate && r.cardRate !== 1">刷卡 {{ rateToZhe(r.cardRate) }} 折 · </span>
          {{ cycleLabel(r) }}
          <span v-if="r.tiers?.length"> · {{ r.tiers.length }} 金额档</span>
          <span v-if="r.countTiers?.length"> · {{ r.countTiers.length }} 次数档</span>
          <span v-if="r.timeWindows?.length"> · {{ r.timeWindows.length }} 时段</span>
          <span v-if="r.freePeriods?.length"> · {{ r.freePeriods.length }} 免费日</span>
        </div>
      </template>
    </EntityCard>
    <p v-if="!list.length" class="empty">暂无规则，点右上角添加</p>

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
import { fenToYuan } from '../utils/money'
import { rateToZhe, type FareRule } from '../utils/fare'

const router = useRouter()
const list = ref<FareRule[]>([])
const confirmOpen = ref(false)
const pendingDelete = ref<FareRule | null>(null)
const deleting = ref(false)

function cycleLabel(r: FareRule) {
  if (r.cycleType === 'from_day') return `每月${r.cycleStartDay}日起算`
  return '自然月'
}

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
