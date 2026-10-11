<template>
  <div class="bd">
    <div class="bd-controls">
      <v-btn-toggle v-model="kind" mandatory density="compact" color="primary" variant="outlined" divided rounded="lg">
        <v-btn value="expense" size="small">{{ t('txType.expense') }}</v-btn>
        <v-btn value="income" size="small">{{ t('txType.income') }}</v-btn>
      </v-btn-toggle>
      <v-btn-toggle v-model="by" mandatory density="compact" color="primary" variant="outlined" divided rounded="lg">
        <v-btn value="category" size="small">{{ t('stats.bd.category') }}</v-btn>
        <v-btn value="account" size="small">{{ t('stats.bd.account') }}</v-btn>
        <v-btn value="tag" size="small">{{ t('stats.bd.tag') }}</v-btn>
      </v-btn-toggle>
    </div>

    <div v-if="parent" class="crumb">
      <v-btn variant="text" size="small" color="primary" prepend-icon="mdi-chevron-left" @click="parent = null">{{ t('stats.bd.allCategories') }}</v-btn>
      <span class="crumb-sep">/</span>
      <span class="crumb-cur">{{ parent.name }}</span>
    </div>

    <div v-if="!items.length" class="empty-tip surface-card">
      {{ loading ? t('common.loading') : t(`stats.bd.empty.${kind}`) }}
    </div>
    <div v-else class="bd-grid" :class="{ loading }">
      <div class="bd-chart surface-card">
        <v-chart class="pie" :option="pieOption" autoresize @click="onPieClick" />
      </div>
      <div class="bd-list surface-card">
        <div class="bd-list-head">
          <span>{{ t('stats.bd.itemCount', { n: items.length }) }}</span>
          <i18n-t keypath="stats.bd.total" tag="span">
            <template #amount><b>{{ yuan(total) }}</b></template>
          </i18n-t>
        </div>
        <component
          :is="canOpen(it) ? 'button' : 'div'"
          v-for="(it, i) in items"
          :key="it.id"
          type="button"
          class="bd-row"
          :class="{ clickable: canOpen(it) }"
          @click="open(it)"
        >
          <span class="bd-dot" :style="{ background: colorAt(i) }">
            <v-icon size="16" color="white">{{ iconOf(it) }}</v-icon>
          </span>
          <div class="bd-main">
            <div class="bd-line">
              <span class="bd-name">{{ it.name || t('stats.bd.deleted') }}</span>
              <span class="bd-amt">{{ yuan(it.amount) }}</span>
            </div>
            <div class="bd-bar"><i :style="{ width: barWidth(it), background: colorAt(i) }" /></div>
            <div class="bd-sub">
              <span>{{ t('stats.count', { n: it.count }) }}<template v-if="drillable(it)"> · {{ t('stats.bd.withSub') }}</template></span>
              <span>{{ total ? percent(it.amount / total) : '—' }}</span>
            </div>
          </div>
          <v-icon v-if="canOpen(it)" size="18" class="bd-chev">
            {{ drillable(it) ? 'mdi-chevron-down' : 'mdi-chevron-right' }}
          </v-icon>
        </component>
      </div>
    </div>
    <p v-if="by === 'tag' && items.length" class="bd-hint">{{ t('stats.bd.tagHint') }}</p>

    <StatsTopList :filter="filter" :kind="kind" class="mt-3" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import http from '../../api/http'
import { accountTypeIcon } from '../../utils/icons'
import { percent, statsParams, yuan, type StatsFilter } from '../../utils/stats'
import { OTHER_COLOR, useChartTheme } from '../../composables/useChartTheme'
import { useBreakpoint } from '../../composables/useBreakpoint'
import StatsTopList from './StatsTopList.vue'

use([CanvasRenderer, PieChart, TitleComponent, TooltipComponent])

interface Item {
  id: number
  name: string
  icon: string
  color: string
  accountType?: string
  amount: number
  count: number
  hasChildren: boolean
}

const props = defineProps<{ filter: StatsFilter }>()
const router = useRouter()
const { t, locale } = useI18n()
const chart = useChartTheme()
const { isDesktop } = useBreakpoint()

const kind = ref<'expense' | 'income'>('expense')
const by = ref<'category' | 'account' | 'tag'>('category')
const parent = ref<{ id: number; name: string } | null>(null)
const items = ref<Item[]>([])
const total = ref(0)
const loading = ref(false)

const maxAmount = computed(() => Math.max(1, ...items.value.map((i) => i.amount)))
/** 饼图最多 8 块，其余并成「其他」，列表仍完整展示 */
const PIE_MAX = 8

function colorAt(i: number) {
  return chart.value.palette[i % chart.value.palette.length]
}

function iconOf(it: Item) {
  if (by.value === 'account') return accountTypeIcon(it.accountType || '')
  if (by.value === 'tag') return it.id ? 'mdi-tag-outline' : 'mdi-tag-off-outline'
  return it.icon || (it.id ? 'mdi-shape-outline' : 'mdi-help-circle-outline')
}

function barWidth(it: Item) {
  return `${Math.max(2, (it.amount / maxAmount.value) * 100)}%`
}

function drillable(it: Item) {
  return by.value === 'category' && !parent.value && it.hasChildren
}

function canOpen(it: Item) {
  return it.id > 0
}

function open(it: Item) {
  if (!canOpen(it)) return
  if (drillable(it)) {
    parent.value = { id: it.id, name: it.name }
    return
  }
  const key = by.value === 'account' ? 'accountId' : by.value === 'tag' ? 'tagId' : 'categoryId'
  const query: Record<string, string> = { type: kind.value, from: props.filter.from, to: props.filter.to, [key]: String(it.id) }
  if (props.filter.accountId && key !== 'accountId') query.accountId = String(props.filter.accountId)
  router.push({ path: '/transactions', query })
}

function onPieClick(e: any) {
  const it = items.value.find((x) => x.id === e?.data?.itemId)
  if (it) open(it)
}

const pieOption = computed(() => {
  const c = chart.value
  const head = items.value.slice(0, PIE_MAX)
  const rest = items.value.slice(PIE_MAX)
  const data = head.map((it, i) => ({ name: it.name, value: it.amount, itemId: it.id, itemStyle: { color: colorAt(i) } }))
  if (rest.length) {
    data.push({
      name: t('stats.bd.others', { n: rest.length }),
      value: rest.reduce((s, it) => s + it.amount, 0),
      itemId: -1,
      itemStyle: { color: OTHER_COLOR },
    })
  }
  return {
    backgroundColor: 'transparent',
    title: {
      text: yuan(total.value, 0),
      subtext: parent.value ? parent.value.name : t(`stats.bd.kindTotal.${kind.value}`),
      left: 'center',
      top: 'center',
      textStyle: { color: c.dark ? '#e6f0ea' : '#15261f', fontSize: isDesktop.value ? 20 : 16, fontWeight: 700 },
      subtextStyle: { color: c.text, fontSize: 12 },
      itemGap: 4,
    },
    tooltip: {
      trigger: 'item',
      formatter: (p: any) => `${p.marker}${p.name}<br/>${yuan(p.value)}${t('stats.bd.tipPct', { pct: `${p.percent}%` })}`,
    },
    series: [{
      type: 'pie',
      radius: ['58%', '84%'],
      center: ['50%', '50%'],
      avoidLabelOverlap: true,
      label: { show: false },
      itemStyle: { borderRadius: 6, borderColor: c.surface, borderWidth: 2 },
      emphasis: { scale: true, scaleSize: 6 },
      data,
    }],
  }
})

async function load() {
  loading.value = true
  try {
    const extra: Record<string, unknown> = { kind: kind.value, by: by.value }
    if (parent.value) extra.parentId = parent.value.id
    const { data } = await http.get('/stats/breakdown', { params: statsParams(props.filter, extra) })
    items.value = data.items || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

watch([kind, by, () => props.filter.from, () => props.filter.to, () => props.filter.accountId], () => {
  if (parent.value) parent.value = null
  else load()
}, { immediate: true })
watch(parent, load)
// 「未分类 / 无标签」等占位名由后端按语言返回，切换语言后重新拉取
watch(locale, load)
</script>

<style scoped>
.bd-controls { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 10px; }
.crumb { display: flex; align-items: center; gap: 2px; margin: -4px 0 8px -8px; font-size: 0.88rem; }
.crumb-sep { color: var(--muted); }
.crumb-cur { font-weight: 600; }

.bd-grid { display: grid; gap: 12px; transition: opacity 0.2s ease; }
.bd-grid.loading { opacity: 0.6; }
.bd-chart { padding: 4px; }
.pie { height: 240px; width: 100%; }

.bd-list { padding: 6px; min-width: 0; }
.bd-list-head {
  display: flex;
  justify-content: space-between;
  padding: 6px 10px 8px;
  font-size: 0.78rem;
  color: var(--muted);
}
.bd-list-head b { color: rgb(var(--v-theme-on-surface)); font-variant-numeric: tabular-nums; }
.bd-row {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 8px 10px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: inherit;
  font: inherit;
  text-align: left;
}
.bd-row.clickable { cursor: pointer; }
.bd-row.clickable:hover { background: var(--primary-soft); }
.bd-dot {
  width: 30px;
  height: 30px;
  border-radius: 9px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.bd-main { flex: 1; min-width: 0; }
.bd-line { display: flex; justify-content: space-between; gap: 8px; align-items: baseline; }
.bd-name { font-size: 0.9rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.bd-amt { font-weight: 600; font-size: 0.9rem; font-variant-numeric: tabular-nums; flex-shrink: 0; }
.bd-bar {
  height: 5px;
  border-radius: 3px;
  background: rgba(var(--v-theme-on-surface), 0.06);
  margin: 5px 0 3px;
  overflow: hidden;
}
.bd-bar i { display: block; height: 100%; border-radius: 3px; }
.bd-sub { display: flex; justify-content: space-between; font-size: 0.72rem; color: var(--muted); }
.bd-chev { color: var(--muted); flex-shrink: 0; }
.bd-hint { margin: 8px 2px 0; font-size: 0.74rem; color: var(--muted); }

/* 宽屏：图在左（滚动时吸顶），排行在右 */
@media (min-width: 960px) {
  .bd-grid { grid-template-columns: minmax(0, 5fr) minmax(0, 7fr); align-items: start; }
  .bd-chart { position: sticky; top: 16px; padding: 12px; }
  .pie { height: 320px; }
}
@media (max-width: 380px) {
  .pie { height: 210px; }
}
</style>
