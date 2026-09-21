<template>
  <div class="page">
    <h1 class="page-title">统计</h1>
    <div class="filters">
      <NativeDateField v-model="from" label="开始" density="compact" variant="outlined" hide-details class="f" />
      <NativeDateField v-model="to" label="结束" density="compact" variant="outlined" hide-details class="f" />
    </div>
    <div class="filters mb-3">
      <v-select
        v-model="accountId"
        :items="[{ id: 0, name: '全部账户' }, ...accounts.list]"
        item-title="name"
        item-value="id"
        density="compact"
        variant="outlined"
        hide-details
        class="f"
      />
      <v-btn color="primary" @click="load">查询</v-btn>
    </div>
    <div class="quick mb-3">
      <v-chip size="small" class="ma-1" variant="tonal" color="primary" @click="setMonth(0)">本月</v-chip>
      <v-chip size="small" class="ma-1" variant="tonal" @click="setMonth(-1)">上月</v-chip>
      <v-chip size="small" class="ma-1" variant="tonal" @click="setYear">今年</v-chip>
    </div>

    <div class="cards">
      <div class="card surface-card">
        <div class="label">区间支出</div>
        <div class="amount-expense val">¥{{ fenToYuan(summary?.expense || summary?.monthExpense || 0) }}</div>
      </div>
      <div class="card surface-card">
        <div class="label">区间收入</div>
        <div class="amount-income val">¥{{ fenToYuan(summary?.income || summary?.monthIncome || 0) }}</div>
      </div>
    </div>

    <h2 class="section-label">分类占比</h2>
    <div v-if="!pieOption" class="empty-tip">暂无支出</div>
    <div v-else class="chart-wrap surface-card">
      <v-chart class="chart" :option="pieOption" autoresize />
    </div>

    <h2 class="section-label">趋势</h2>
    <div class="chart-wrap surface-card">
      <v-chart class="chart" :option="lineOption" autoresize />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart, LineChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { useTheme } from 'vuetify'
import http from '../api/http'
import { fenToYuan } from '../utils/money'
import { useAccountStore } from '../stores/account'
import NativeDateField from '../components/NativeDateField.vue'

use([CanvasRenderer, PieChart, LineChart, TooltipComponent, LegendComponent, GridComponent])

interface Summary {
  monthExpense: number
  monthIncome: number
  expense?: number
  income?: number
  byCategory: { categoryId: number; categoryName: string; amount: number }[]
  trends: { month: string; expense: number; income: number }[]
}

const accounts = useAccountStore()
const theme = useTheme()
const summary = ref<Summary | null>(null)
const from = ref(dayjs().startOf('month').format('YYYY-MM-DD'))
const to = ref(dayjs().endOf('month').format('YYYY-MM-DD'))
const accountId = ref(0)

const isDark = computed(() => theme.global.current.value.dark)
const chartText = computed(() => (isDark.value ? '#c5d4cb' : '#5a6f64'))
const chartAxis = computed(() => (isDark.value ? '#3a4a42' : '#d5e0d9'))

function setMonth(offset: number) {
  const m = dayjs().add(offset, 'month')
  from.value = m.startOf('month').format('YYYY-MM-DD')
  to.value = m.endOf('month').format('YYYY-MM-DD')
  load()
}
function setYear() {
  from.value = dayjs().startOf('year').format('YYYY-MM-DD')
  to.value = dayjs().endOf('year').format('YYYY-MM-DD')
  load()
}

async function load() {
  const params: Record<string, unknown> = {
    from: from.value,
    to: dayjs(to.value).add(1, 'day').format('YYYY-MM-DD'),
  }
  if (accountId.value) params.accountId = accountId.value
  const { data } = await http.get('/stats/summary', { params })
  summary.value = data
}

onMounted(async () => {
  await accounts.load()
  await load()
})

watch(isDark, () => {
  // 触发 computed 图表重绘
})

const pieOption = computed(() => {
  const cats = summary.value?.byCategory || []
  if (!cats.length) return null
  return {
    tooltip: { trigger: 'item', formatter: (p: any) => `${p.name}: ¥${fenToYuan(p.value)}` },
    textStyle: { color: chartText.value },
    series: [{
      type: 'pie',
      radius: ['38%', '68%'],
      itemStyle: { borderRadius: 6, borderColor: isDark.value ? '#1c2620' : '#fff', borderWidth: 2 },
      label: { color: chartText.value },
      data: cats.map((c) => ({ name: c.categoryName, value: c.amount })),
    }],
  }
})

const lineOption = computed(() => {
  const trends = summary.value?.trends || []
  const expense = getComputedStyle(document.documentElement).getPropertyValue('--expense').trim() || '#c45c3e'
  const income = getComputedStyle(document.documentElement).getPropertyValue('--income').trim() || '#1b7f5a'
  return {
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis' },
    legend: { data: ['支出', '收入'], textStyle: { color: chartText.value } },
    grid: { left: 44, right: 16, top: 40, bottom: 28 },
    textStyle: { color: chartText.value },
    xAxis: {
      type: 'category',
      data: trends.map((t) => t.month),
      axisLabel: { color: chartText.value },
      axisLine: { lineStyle: { color: chartAxis.value } },
    },
    yAxis: {
      type: 'value',
      axisLabel: { formatter: (v: number) => (v / 100).toFixed(0), color: chartText.value },
      splitLine: { lineStyle: { color: chartAxis.value, type: 'dashed' } },
    },
    series: [
      { name: '支出', type: 'line', smooth: true, data: trends.map((t) => t.expense), itemStyle: { color: expense }, areaStyle: { color: expense + '22' } },
      { name: '收入', type: 'line', smooth: true, data: trends.map((t) => t.income), itemStyle: { color: income }, areaStyle: { color: income + '22' } },
    ],
  }
})
</script>

<style scoped>
.filters { display: flex; gap: 8px; align-items: center; margin-bottom: 8px; }
.f { flex: 1; }
.cards { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.card { padding: 16px; }
.label { color: var(--muted); font-size: 0.85rem; }
.val {
  font-size: 1.35rem;
  margin-top: 6px;
  font-family: var(--font-display);
  font-variant-numeric: tabular-nums;
}
.chart-wrap { padding: 8px; }
.chart { height: 280px; width: 100%; }
@media (max-width: 600px) {
  .filters { flex-wrap: wrap; }
}
</style>
