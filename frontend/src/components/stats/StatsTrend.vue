<template>
  <div class="tr">
    <div class="tr-controls">
      <v-btn-toggle v-model="granularity" mandatory density="compact" color="primary" variant="outlined" divided rounded="lg">
        <v-btn value="day" size="small">{{ t('stats.gran.day') }}</v-btn>
        <v-btn value="month" size="small">{{ t('stats.gran.month') }}</v-btn>
        <v-btn value="year" size="small">{{ t('stats.gran.year') }}</v-btn>
      </v-btn-toggle>
      <span v-if="actual !== granularity && points.length" class="tr-note">{{ t('stats.trend.autoSwitched', { unit: granularityUnit(actual) }) }}</span>
    </div>

    <div class="tr-grid" :class="{ loading }">
      <div class="tr-chart surface-card">
        <v-chart class="chart" :option="option" autoresize />
      </div>
      <div class="tr-side surface-card">
        <div class="side-item">
          <span class="side-label">{{ t(`stats.trend.avgExpense.${actual}`) }}</span>
          <span class="side-val amount-expense">{{ yuan(avg.expense) }}</span>
        </div>
        <div class="side-item">
          <span class="side-label">{{ t(`stats.trend.avgIncome.${actual}`) }}</span>
          <span class="side-val amount-income">{{ yuan(avg.income) }}</span>
        </div>
        <div class="side-item">
          <span class="side-label">{{ t(`stats.trend.peak.${actual}`) }}</span>
          <span class="side-val">
            <template v-if="peak">{{ periodLabel(peak.key, actual) }} · {{ yuan(peak.expense) }}</template>
            <template v-else>—</template>
          </span>
        </div>
        <div class="side-item">
          <span class="side-label">{{ t(`stats.trend.deficit.${actual}`) }}</span>
          <span class="side-val">{{ deficitCount }} / {{ elapsed.length }}</span>
        </div>
      </div>
    </div>

    <section class="tr-table surface-card">
      <div class="tbl-head tbl-row">
        <span>{{ t('stats.trend.period') }}</span>
        <span class="num">{{ t('txType.income') }}</span>
        <span class="num">{{ t('txType.expense') }}</span>
        <span class="num">{{ t('stats.net') }}</span>
        <span class="num col-count">{{ t('stats.trend.count') }}</span>
      </div>
      <div v-if="!rows.length" class="tbl-empty">{{ t('stats.trend.empty') }}</div>
      <button v-for="p in rows" :key="p.key" type="button" class="tbl-row tbl-body" @click="openPeriod(p.key)">
        <span>{{ periodLabel(p.key, actual) }}</span>
        <span class="num amount-income">{{ yuan(p.income) }}</span>
        <span class="num amount-expense">{{ yuan(p.expense) }}</span>
        <span class="num" :class="p.net >= 0 ? 'net-pos' : 'net-neg'">{{ yuan(p.net) }}</span>
        <span class="num col-count">{{ p.expenseCount + p.incomeCount }}</span>
      </button>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, LineChart } from 'echarts/charts'
import { DataZoomComponent, GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import http from '../../api/http'
import {
  compactYuan, granularityUnit, periodLabel, periodRange, statsParams, yuan,
  type Granularity, type StatsFilter,
} from '../../utils/stats'
import { useChartTheme } from '../../composables/useChartTheme'
import { useBreakpoint } from '../../composables/useBreakpoint'

use([CanvasRenderer, BarChart, LineChart, DataZoomComponent, GridComponent, LegendComponent, TooltipComponent])

interface Point {
  key: string
  expense: number
  income: number
  net: number
  expenseCount: number
  incomeCount: number
}

const props = defineProps<{ filter: StatsFilter }>()
const router = useRouter()
const { t } = useI18n()
const chart = useChartTheme()
const { isDesktop } = useBreakpoint()

const granularity = ref<Granularity>(defaultGranularity())
const actual = ref<Granularity>(granularity.value)
const points = ref<Point[]>([])
const loading = ref(false)

function defaultGranularity(): Granularity {
  const days = dayjs(props.filter.to).diff(dayjs(props.filter.from), 'day') + 1
  if (days <= 62) return 'day'
  if (days <= 3 * 366) return 'month'
  return 'year'
}

/** 未来的周期不参与平均 */
const elapsed = computed(() => {
  const now = dayjs()
  return points.value.filter((p) => !periodStart(p.key).isAfter(now))
})
function periodStart(key: string) {
  return dayjs(periodRange(key, actual.value).from)
}
const avg = computed(() => {
  const n = Math.max(1, elapsed.value.length)
  const sum = elapsed.value.reduce((s, p) => ({ expense: s.expense + p.expense, income: s.income + p.income }), { expense: 0, income: 0 })
  return { expense: Math.round(sum.expense / n), income: Math.round(sum.income / n) }
})
const peak = computed(() => {
  const top = [...points.value].sort((a, b) => b.expense - a.expense)[0]
  return top && top.expense > 0 ? top : null
})
const deficitCount = computed(() => elapsed.value.filter((p) => p.net < 0).length)
const rows = computed(() => points.value.filter((p) => p.expenseCount || p.incomeCount).reverse())

const option = computed(() => {
  const c = chart.value
  const pts = points.value
  const visible = isDesktop.value ? 40 : 14
  const zoom = pts.length > visible
  const names = { income: t('txType.income'), expense: t('txType.expense'), net: t('stats.net') }
  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      valueFormatter: (v: number) => yuan(v),
    },
    legend: { data: [names.income, names.expense, names.net], top: 0, textStyle: { color: c.text }, itemWidth: 14, itemHeight: 8 },
    grid: { left: 8, right: 8, top: 34, bottom: zoom ? 44 : 8, containLabel: true },
    xAxis: {
      type: 'category',
      data: pts.map((p) => periodLabel(p.key, actual.value, true)),
      axisLabel: { color: c.text, hideOverlap: true },
      axisLine: { lineStyle: { color: c.axis } },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: c.text, formatter: (v: number) => compactYuan(v) },
      splitLine: { lineStyle: { color: c.axis, type: 'dashed' } },
    },
    dataZoom: zoom
      ? [{
          type: 'slider',
          height: 18,
          bottom: 8,
          startValue: Math.max(0, pts.length - visible),
          endValue: pts.length - 1,
          brushSelect: false,
          textStyle: { color: c.text },
          borderColor: c.axis,
        }]
      : [],
    series: [
      { name: names.income, type: 'bar', data: pts.map((p) => p.income), itemStyle: { color: c.income, borderRadius: [4, 4, 0, 0] }, barMaxWidth: 18 },
      { name: names.expense, type: 'bar', data: pts.map((p) => p.expense), itemStyle: { color: c.expense, borderRadius: [4, 4, 0, 0] }, barMaxWidth: 18 },
      {
        name: names.net,
        type: 'line',
        data: pts.map((p) => p.net),
        smooth: true,
        symbol: 'circle',
        symbolSize: 5,
        itemStyle: { color: c.dark ? '#e8c46a' : '#c99a2e' },
        lineStyle: { width: 2 },
      },
    ],
  }
})

function openPeriod(key: string) {
  const r = periodRange(key, actual.value)
  const from = r.from < props.filter.from ? props.filter.from : r.from
  const to = r.to > props.filter.to ? props.filter.to : r.to
  const query: Record<string, string> = { from, to }
  if (props.filter.accountId) query.accountId = String(props.filter.accountId)
  router.push({ path: '/transactions', query })
}

async function load() {
  loading.value = true
  try {
    const { data } = await http.get('/stats/trend', {
      params: statsParams(props.filter, { granularity: granularity.value }),
    })
    actual.value = data.granularity
    points.value = data.points || []
  } finally {
    loading.value = false
  }
}

watch(() => [props.filter.from, props.filter.to], () => {
  const g = defaultGranularity()
  if (g !== granularity.value) granularity.value = g
  else load()
})
watch([granularity, () => props.filter.accountId], load, { immediate: true })
</script>

<style scoped>
.tr-controls { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 12px; margin-bottom: 10px; }
.tr-note { font-size: 0.74rem; color: var(--muted); }

.tr-grid { display: grid; gap: 12px; transition: opacity 0.2s ease; }
.tr-grid.loading { opacity: 0.6; }
.tr-chart { padding: 10px 6px 4px; min-width: 0; }
.chart { height: 260px; width: 100%; }

.tr-side {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1px;
  overflow: hidden;
  background: var(--surface-border);
}
.side-item {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 10px 12px;
  background: var(--surface-solid);
  min-width: 0;
}
.side-label { font-size: 0.72rem; color: var(--muted); }
.side-val { font-size: 0.88rem; font-weight: 600; font-variant-numeric: tabular-nums; overflow-wrap: anywhere; }

.tr-table { margin-top: 12px; padding: 4px 6px 6px; }
.tbl-row {
  display: grid;
  grid-template-columns: minmax(64px, 1fr) repeat(3, minmax(0, 1.2fr));
  gap: 6px;
  align-items: center;
  padding: 9px 8px;
  font-size: 0.8rem;
  font-variant-numeric: tabular-nums;
}
.tbl-head { color: var(--muted); font-size: 0.74rem; border-bottom: 1px solid var(--surface-border); }
.tbl-body {
  width: 100%;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  font-size: 0.8rem;
  text-align: left;
  border-radius: 8px;
  cursor: pointer;
}
.tbl-body:hover { background: var(--primary-soft); }
.tbl-body + .tbl-body { border-top: 1px dashed var(--surface-border); }
.num { text-align: right; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.tbl-body .amount-income,
.tbl-body .amount-expense { font-weight: 500; }
.net-pos { color: var(--income); }
.net-neg { color: var(--expense); }
.col-count { display: none; }
.tbl-empty { color: var(--muted); text-align: center; padding: 18px 0; font-size: 0.85rem; }

@media (min-width: 600px) {
  .col-count { display: block; }
  .tbl-row { grid-template-columns: minmax(90px, 1fr) repeat(3, minmax(0, 1.2fr)) 56px; font-size: 0.86rem; }
}

/* 宽屏：图表 + 右侧摘要并排 */
@media (min-width: 960px) {
  .tr-grid { grid-template-columns: minmax(0, 1fr) 240px; }
  .chart { height: 340px; }
  .tr-side { grid-template-columns: 1fr; align-content: stretch; }
  .side-item { justify-content: center; padding: 12px 16px; }
  .side-val { font-size: 0.98rem; }
}
</style>
