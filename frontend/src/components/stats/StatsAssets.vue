<template>
  <div class="as">
    <div class="as-kpis">
      <div class="kpi kpi--net surface-card">
        <span class="kpi-label">净资产（当前）</span>
        <span class="kpi-val" :class="data.net >= 0 ? 'amount-income' : 'amount-expense'">{{ yuan(data.net) }}</span>
        <span v-if="netChange !== null" class="kpi-delta" :class="netChange >= 0 ? 'is-good' : 'is-bad'">
          区间内 {{ netChange >= 0 ? '+' : '' }}{{ yuan(netChange) }}
        </span>
      </div>
      <div class="kpi surface-card">
        <span class="kpi-label">总资产</span>
        <span class="kpi-val">{{ yuan(data.assets) }}</span>
      </div>
      <div class="kpi surface-card">
        <span class="kpi-label">总负债</span>
        <span class="kpi-val amount-expense">{{ yuan(data.liabilities) }}</span>
      </div>
    </div>

    <div class="as-controls">
      <v-btn-toggle v-model="granularity" mandatory density="compact" color="primary" variant="outlined" divided rounded="lg">
        <v-btn value="day" size="small">按日</v-btn>
        <v-btn value="month" size="small">按月</v-btn>
        <v-btn value="year" size="small">按年</v-btn>
      </v-btn-toggle>
      <span class="as-note">
        {{ filter.accountId ? '仅所选账户' : '全部账户' }}，按当前余额倒推各期末余额
      </span>
    </div>

    <div class="as-chart surface-card" :class="{ loading }">
      <v-chart class="chart" :option="option" autoresize />
    </div>

    <div class="as-dist">
      <section v-for="g in groups" :key="g.title" class="dist surface-card">
        <header class="dist-head">
          <span class="dist-title">{{ g.title }}</span>
          <span class="dist-total" :class="g.cls">{{ yuan(g.total) }}</span>
        </header>
        <div v-if="!g.list.length" class="dist-empty">无</div>
        <router-link v-for="a in g.list" :key="a.id" :to="`/accounts/${a.id}/ledger`" class="dist-row">
          <v-icon size="18" class="dist-ic">{{ accountTypeIcon(a.type) }}</v-icon>
          <div class="dist-main">
            <div class="dist-line">
              <span class="dist-name">{{ a.name }}</span>
              <span class="dist-amt">{{ yuan(Math.abs(a.balance)) }}</span>
            </div>
            <div class="dist-bar"><i :style="{ width: share(a, g.total), background: g.color }" /></div>
          </div>
          <span class="dist-pct">{{ g.total ? percent(Math.abs(a.balance) / g.total, 0) : '' }}</span>
        </router-link>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { DataZoomComponent, GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import http from '../../api/http'
import { accountTypeIcon } from '../../utils/icons'
import { compactYuan, percent, periodLabel, statsParams, yuan, type Granularity, type StatsFilter } from '../../utils/stats'
import { useChartTheme } from '../../composables/useChartTheme'
import { useBreakpoint } from '../../composables/useBreakpoint'

use([CanvasRenderer, LineChart, DataZoomComponent, GridComponent, LegendComponent, TooltipComponent])

interface AssetAccount {
  id: number
  name: string
  type: string
  balance: number
}
interface AssetTrend {
  granularity: Granularity
  assets: number
  liabilities: number
  net: number
  points: { key: string; assets: number; liabilities: number; net: number }[]
  accounts: AssetAccount[]
}

const props = defineProps<{ filter: StatsFilter }>()
const chart = useChartTheme()
const { isDesktop } = useBreakpoint()

const granularity = ref<Granularity>(defaultGranularity())
const data = ref<AssetTrend>({ granularity: 'month', assets: 0, liabilities: 0, net: 0, points: [], accounts: [] })
const loading = ref(false)

function defaultGranularity(): Granularity {
  const days = dayjs(props.filter.to).diff(dayjs(props.filter.from), 'day') + 1
  if (days <= 62) return 'day'
  if (days <= 5 * 366) return 'month'
  return 'year'
}

/** 区间首末两个期末的净资产差 */
const netChange = computed(() => {
  const pts = data.value.points
  if (pts.length < 2) return null
  return pts[pts.length - 1].net - pts[0].net
})

const groups = computed(() => {
  const c = chart.value
  const assets = data.value.accounts.filter((a) => a.balance > 0)
  const debts = data.value.accounts.filter((a) => a.balance < 0).sort((a, b) => a.balance - b.balance)
  return [
    { title: '资产账户', list: assets, total: data.value.assets, color: c.income, cls: '' },
    { title: '负债账户', list: debts, total: data.value.liabilities, color: c.expense, cls: 'amount-expense' },
  ]
})

function share(a: AssetAccount, total: number) {
  return total ? `${Math.max(2, (Math.abs(a.balance) / total) * 100)}%` : '0%'
}

const option = computed(() => {
  const c = chart.value
  const pts = data.value.points
  const g = data.value.granularity
  const visible = isDesktop.value ? 60 : 20
  const zoom = pts.length > visible
  const netColor = c.dark ? '#e8c46a' : '#c99a2e'
  return {
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis', valueFormatter: (v: number) => yuan(v) },
    legend: { data: ['净资产', '资产', '负债'], top: 0, textStyle: { color: c.text }, itemWidth: 14, itemHeight: 8 },
    grid: { left: 8, right: 8, top: 34, bottom: zoom ? 44 : 8, containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: pts.map((p) => periodLabel(p.key, g, g !== 'year' && pts.length > 12)),
      axisLabel: { color: c.text, hideOverlap: true },
      axisLine: { lineStyle: { color: c.axis } },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      scale: true,
      axisLabel: { color: c.text, formatter: (v: number) => compactYuan(v) },
      splitLine: { lineStyle: { color: c.axis, type: 'dashed' } },
    },
    dataZoom: zoom
      ? [{ type: 'slider', height: 18, bottom: 8, startValue: pts.length - visible, endValue: pts.length - 1, brushSelect: false, textStyle: { color: c.text }, borderColor: c.axis }]
      : [],
    series: [
      {
        name: '净资产',
        type: 'line',
        data: pts.map((p) => p.net),
        smooth: true,
        showSymbol: pts.length <= 24,
        lineStyle: { width: 3, color: netColor },
        itemStyle: { color: netColor },
        areaStyle: { color: netColor + '22' },
      },
      { name: '资产', type: 'line', data: pts.map((p) => p.assets), smooth: true, showSymbol: false, itemStyle: { color: c.income }, lineStyle: { width: 1.5 } },
      { name: '负债', type: 'line', data: pts.map((p) => p.liabilities), smooth: true, showSymbol: false, itemStyle: { color: c.expense }, lineStyle: { width: 1.5 } },
    ],
  }
})

async function load() {
  loading.value = true
  try {
    const { data: res } = await http.get('/stats/assets', {
      params: statsParams(props.filter, { granularity: granularity.value }),
    })
    data.value = { ...res, points: res.points || [], accounts: res.accounts || [] }
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
.as-kpis {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 12px;
}
.kpi { display: flex; flex-direction: column; gap: 3px; padding: 12px 14px; min-width: 0; }
.kpi--net { grid-column: 1 / -1; }
.kpi-label { font-size: 0.78rem; color: var(--muted); }
.kpi-val {
  font-family: var(--font-display);
  font-weight: 700;
  font-size: clamp(1.02rem, 4.4vw, 1.35rem);
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}
.kpi--net .kpi-val { font-size: clamp(1.25rem, 6vw, 1.7rem); }
.kpi-delta { font-size: 0.74rem; color: var(--muted); }
.kpi-delta.is-good { color: var(--income); }
.kpi-delta.is-bad { color: var(--expense); }

.as-controls { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 12px; margin-bottom: 10px; }
.as-note { font-size: 0.74rem; color: var(--muted); }

.as-chart { padding: 10px 6px 4px; transition: opacity 0.2s ease; }
.as-chart.loading { opacity: 0.6; }
.chart { height: 260px; width: 100%; }

.as-dist { display: grid; gap: 12px; margin-top: 12px; }
.dist { padding: 10px 6px 6px; min-width: 0; }
.dist-head { display: flex; justify-content: space-between; align-items: baseline; padding: 0 10px 6px; }
.dist-title { font-weight: 700; font-size: 0.9rem; }
.dist-total { font-weight: 600; font-variant-numeric: tabular-nums; font-size: 0.9rem; }
.dist-empty { color: var(--muted); font-size: 0.82rem; padding: 6px 10px 10px; }
.dist-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 10px;
  color: inherit;
  text-decoration: none;
}
.dist-row:hover { background: var(--primary-soft); }
.dist-ic { color: var(--muted); flex-shrink: 0; }
.dist-main { flex: 1; min-width: 0; }
.dist-line { display: flex; justify-content: space-between; gap: 8px; }
.dist-name { font-size: 0.88rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.dist-amt { font-size: 0.86rem; font-weight: 600; font-variant-numeric: tabular-nums; flex-shrink: 0; }
.dist-bar { height: 4px; border-radius: 2px; background: rgba(var(--v-theme-on-surface), 0.06); margin-top: 5px; overflow: hidden; }
.dist-bar i { display: block; height: 100%; border-radius: 2px; }
.dist-pct { width: 36px; text-align: right; font-size: 0.74rem; color: var(--muted); flex-shrink: 0; }

@media (min-width: 600px) {
  .as-kpis { grid-template-columns: 1.4fr 1fr 1fr; }
  .kpi--net { grid-column: auto; }
}

/* 宽屏：资产 / 负债账户左右并排 */
@media (min-width: 960px) {
  .kpi { padding: 16px 18px; }
  .chart { height: 340px; }
  .as-dist { grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: start; }
}
</style>
