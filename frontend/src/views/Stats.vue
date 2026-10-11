<template>
  <div class="page stats-page">
    <div class="stats-head">
      <h1 class="page-title">{{ t('stats.title') }}</h1>
      <span class="range-text">{{ rangeText }}</span>
    </div>

    <div class="filter-bar">
      <div class="presets hide-scrollbar">
        <v-chip
          v-for="p in presets"
          :key="p.value"
          size="small"
          class="preset"
          :color="preset === p.value ? 'primary' : undefined"
          :variant="preset === p.value ? 'flat' : 'tonal'"
          @click="choose(p.value)"
        >{{ p.label }}</v-chip>
      </div>
      <v-select
        v-model="accountId"
        :items="[{ id: 0, name: t('stats.allAccounts') }, ...accounts.list]"
        item-title="name"
        item-value="id"
        density="compact"
        variant="outlined"
        hide-details
        prepend-inner-icon="mdi-wallet-outline"
        class="acct"
      />
    </div>
    <div v-if="preset === 'custom'" class="custom-range">
      <NativeDateField v-model="from" :label="t('stats.from')" density="compact" variant="outlined" hide-details class="f" />
      <NativeDateField v-model="to" :label="t('stats.to')" density="compact" variant="outlined" hide-details class="f" />
    </div>

    <StatsOverviewCards :filter="filter" />

    <v-tabs v-model="tab" color="primary" class="stats-tabs" :grow="!isDesktop" density="comfortable">
      <v-tab
        v-for="tb in tabs"
        :key="tb.value"
        :value="tb.value"
        :prepend-icon="width >= 600 ? tb.icon : undefined"
        class="stats-tab"
      >{{ tb.label }}</v-tab>
    </v-tabs>
    <div class="tab-body">
      <KeepAlive>
        <component :is="tabComponent" :key="tab" :filter="filter" />
      </KeepAlive>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAccountStore } from '../stores/account'
import { useBreakpoint } from '../composables/useBreakpoint'
import type { StatsFilter } from '../utils/stats'
import NativeDateField from '../components/NativeDateField.vue'
import StatsOverviewCards from '../components/stats/StatsOverviewCards.vue'
import StatsBreakdown from '../components/stats/StatsBreakdown.vue'
import StatsTrend from '../components/stats/StatsTrend.vue'
import StatsCalendar from '../components/stats/StatsCalendar.vue'
import StatsAssets from '../components/stats/StatsAssets.vue'
import StatsMap from '../components/stats/StatsMap.vue'

type Preset = 'month' | 'lastMonth' | 'last3' | 'year' | 'lastYear' | 'custom'
type Tab = 'breakdown' | 'trend' | 'calendar' | 'map' | 'assets'

const STATE_KEY = 'mbk_stats_state'

const PRESETS: Preset[] = ['month', 'lastMonth', 'last3', 'year', 'lastYear', 'custom']
const TAB_ICONS: Record<Tab, string> = {
  breakdown: 'mdi-chart-donut',
  trend: 'mdi-chart-bar',
  calendar: 'mdi-calendar-month-outline',
  map: 'mdi-map-marker-radius-outline',
  assets: 'mdi-finance',
}
const { t } = useI18n()
const presets = computed(() => PRESETS.map((value) => ({ value, label: t(`stats.presets.${value}`) })))
const tabs = computed(() =>
  (Object.keys(TAB_ICONS) as Tab[]).map((value) => ({ value, label: t(`stats.tabs.${value}`), icon: TAB_ICONS[value] })),
)
const tabComponents = {
  breakdown: StatsBreakdown,
  trend: StatsTrend,
  calendar: StatsCalendar,
  map: StatsMap,
  assets: StatsAssets,
}

const accounts = useAccountStore()
const route = useRoute()
const { isDesktop, width } = useBreakpoint()

const saved = readState()
const preset = ref<Preset>(saved.preset || 'month')
const from = ref(saved.from || '')
const to = ref(saved.to || '')
const accountId = ref<number>(saved.accountId || 0)
const queryTab = route.query.tab as Tab | undefined
const tab = ref<Tab>(queryTab && queryTab in tabComponents ? queryTab : saved.tab || 'breakdown')
if (preset.value !== 'custom' || !from.value || !to.value) applyPreset(preset.value === 'custom' ? 'month' : preset.value)

const tabComponent = computed(() => tabComponents[tab.value] || StatsBreakdown)

/** 自定义区间填反了也照常统计 */
const filter = computed<StatsFilter>(() => {
  const [a, b] = from.value <= to.value ? [from.value, to.value] : [to.value, from.value]
  return { from: a, to: b, accountId: accountId.value }
})

const rangeText = computed(() => {
  const f = dayjs(filter.value.from)
  const e = dayjs(filter.value.to)
  const fmt = t(f.year() === e.year() && f.year() === dayjs().year() ? 'stats.fmt.monthDay' : 'stats.fmt.fullDate')
  return `${f.format(fmt)} – ${e.format(fmt)}`
})

function readState(): Partial<{ preset: Preset; from: string; to: string; accountId: number; tab: Tab }> {
  try {
    return JSON.parse(sessionStorage.getItem(STATE_KEY) || '{}')
  } catch {
    return {}
  }
}

function applyPreset(p: Preset) {
  const now = dayjs()
  const set = (a: dayjs.Dayjs, b: dayjs.Dayjs) => {
    from.value = a.format('YYYY-MM-DD')
    to.value = b.format('YYYY-MM-DD')
  }
  switch (p) {
    case 'month': set(now.startOf('month'), now.endOf('month')); break
    case 'lastMonth': {
      const m = now.subtract(1, 'month')
      set(m.startOf('month'), m.endOf('month'))
      break
    }
    case 'last3': set(now.subtract(2, 'month').startOf('month'), now.endOf('month')); break
    case 'year': set(now.startOf('year'), now.endOf('year')); break
    case 'lastYear': {
      const y = now.subtract(1, 'year')
      set(y.startOf('year'), y.endOf('year'))
      break
    }
  }
  preset.value = p
}

function choose(p: Preset) {
  if (p === 'custom') preset.value = 'custom'
  else applyPreset(p)
}

watch([preset, from, to, accountId, tab], () => {
  sessionStorage.setItem(STATE_KEY, JSON.stringify({
    preset: preset.value, from: from.value, to: to.value, accountId: accountId.value, tab: tab.value,
  }))
})

onMounted(() => accounts.ensure())
</script>

<style scoped>
.stats-head { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; }
.range-text { font-size: 0.8rem; color: var(--muted); white-space: nowrap; }

.filter-bar { display: flex; flex-direction: column; gap: 8px; margin-bottom: 10px; }
.presets {
  display: flex;
  gap: 6px;
  overflow-x: auto;
  margin: 0 calc(-1 * var(--content-pad-x));
  padding: 2px var(--content-pad-x);
}
.preset { flex-shrink: 0; }
.acct { min-width: 0; }
.custom-range { display: flex; gap: 8px; margin-bottom: 10px; }
.f { flex: 1; min-width: 0; }

.stats-tabs {
  margin: 18px 0 12px;
  border-bottom: 1px solid var(--surface-border);
}
.tab-body { min-height: 320px; }
.stats-tab { min-width: 0 !important; padding: 0 10px !important; letter-spacing: 0; }

/* 平板以上：快捷区间与账户筛选同一行 */
@media (min-width: 600px) {
  .filter-bar { flex-direction: row; align-items: center; }
  .presets { flex: 1; margin: 0; padding: 2px 0; flex-wrap: wrap; overflow: visible; }
  .acct { flex: 0 0 220px; }
  .custom-range { max-width: 440px; margin-left: auto; }
}

@media (min-width: 960px) {
  .stats-tabs { margin-top: 24px; }
}
</style>
