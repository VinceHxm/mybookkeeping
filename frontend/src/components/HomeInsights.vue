<template>
  <section class="insights surface-card">
    <div class="ins-head">
      <h2 class="ins-title">本月支出构成</h2>
      <router-link class="ins-more" to="/stats">统计 <v-icon size="14">mdi-chevron-right</v-icon></router-link>
    </div>

    <div v-if="!cats.length" class="ins-empty">{{ data ? '本月还没有支出' : '加载中…' }}</div>
    <ul v-else class="cat-list">
      <li v-for="(c, i) in cats" :key="c.id">
        <router-link class="cat-row" :to="catRoute(c)">
          <span class="cat-dot" :style="{ background: colorOf(c, i) }" />
          <span class="cat-name">{{ c.name }}</span>
          <span class="cat-pct">{{ pct(c.amount) }}</span>
          <span class="cat-amt">{{ yuan(c.amount) }}</span>
          <span class="cat-bar"><i :style="{ width: barWidth(c.amount), background: colorOf(c, i) }" /></span>
        </router-link>
      </li>
    </ul>

    <div class="ins-head ins-head--sub">
      <h2 class="ins-title">近 7 天支出</h2>
      <span class="ins-sum">日均 {{ yuan(avg7) }}</span>
    </div>
    <div class="week">
      <router-link
        v-for="d in week"
        :key="d.key"
        class="day"
        :class="{ 'day--today': d.today }"
        :to="{ path: '/transactions', query: { from: d.key, to: d.key, type: 'expense' } }"
        :title="`${d.key} 支出 ${yuan(d.expense)}`"
      >
        <span class="day-val">{{ d.expense ? compactYuan(d.expense) : '' }}</span>
        <span class="day-track"><i :style="{ height: dayHeight(d.expense) }" /></span>
        <span class="day-label">{{ d.label }}</span>
      </router-link>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import dayjs from 'dayjs'
import { CHART_PALETTE } from '../composables/useChartTheme'
import { compactYuan, yuan } from '../utils/stats'

interface CatItem { id: number; name: string; color: string; amount: number }
interface DayPoint { key: string; expense: number }

/** 与 /home/overview?insights=1 返回的 insights 字段一致 */
export interface HomeInsightsData {
  categories: { total: number; items: CatItem[] } | null
  week: DayPoint[]
}

const props = defineProps<{ data: HomeInsightsData | null }>()

const WEEKDAY = ['日', '一', '二', '三', '四', '五', '六']

const total = computed(() => props.data?.categories?.total || 0)
const points = computed(() => props.data?.week || [])

const monthRange = () => ({
  from: dayjs().startOf('month').format('YYYY-MM-DD'),
  to: dayjs().endOf('month').format('YYYY-MM-DD'),
})

const cats = computed(() => (props.data?.categories?.items || []).filter((c) => c.amount > 0))
const maxCat = computed(() => Math.max(1, ...cats.value.map((c) => c.amount)))

const week = computed(() => {
  const byKey = new Map(points.value.map((p) => [p.key, p.expense]))
  const today = dayjs().format('YYYY-MM-DD')
  return Array.from({ length: 7 }, (_, i) => {
    const d = dayjs().subtract(6 - i, 'day')
    const key = d.format('YYYY-MM-DD')
    return {
      key,
      expense: byKey.get(key) || 0,
      today: key === today,
      label: key === today ? '今天' : `周${WEEKDAY[d.day()]}`,
    }
  })
})
const maxDay = computed(() => Math.max(1, ...week.value.map((d) => d.expense)))
const avg7 = computed(() => Math.round(week.value.reduce((s, d) => s + d.expense, 0) / 7))

function colorOf(c: CatItem, i: number) {
  return c.color || CHART_PALETTE[i % CHART_PALETTE.length]
}
function pct(amount: number) {
  return total.value ? `${((amount / total.value) * 100).toFixed(0)}%` : ''
}
function barWidth(amount: number) {
  return `${Math.max(3, (amount / maxCat.value) * 100)}%`
}
function dayHeight(amount: number) {
  return amount ? `${Math.max(6, (amount / maxDay.value) * 100)}%` : '0%'
}
function catRoute(c: CatItem) {
  const query: Record<string, string> = { ...monthRange(), type: 'expense' }
  if (c.id) query.categoryId = String(c.id)
  return { path: '/transactions', query }
}
</script>

<style scoped>
.insights {
  padding: 16px 18px 18px;
  border-radius: var(--radius-lg);
}
.ins-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 10px;
}
.ins-head--sub { margin-top: 18px; }
.ins-title { margin: 0; font-size: 0.92rem; font-weight: 700; }
.ins-more, .ins-sum {
  font-size: 0.78rem;
  color: var(--muted);
  text-decoration: none;
  white-space: nowrap;
}
.ins-more:hover { color: rgb(var(--v-theme-primary)); }
.ins-empty { padding: 12px 0; font-size: 0.82rem; color: var(--muted); }

.cat-list { list-style: none; margin: 0; padding: 0; display: grid; gap: 4px; }
.cat-row {
  display: grid;
  grid-template-columns: 10px minmax(0, 1fr) auto auto;
  grid-template-areas: 'dot name pct amt' 'bar bar bar bar';
  align-items: center;
  column-gap: 8px;
  row-gap: 5px;
  padding: 6px 6px;
  border-radius: 10px;
  color: inherit;
  text-decoration: none;
}
.cat-row:hover { background: var(--primary-soft); }
.cat-dot { grid-area: dot; width: 10px; height: 10px; border-radius: 50%; }
.cat-name { grid-area: name; font-size: 0.85rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cat-pct { grid-area: pct; font-size: 0.75rem; color: var(--muted); font-variant-numeric: tabular-nums; }
.cat-amt { grid-area: amt; font-size: 0.85rem; font-weight: 600; font-variant-numeric: tabular-nums; }
.cat-bar {
  grid-area: bar;
  height: 4px;
  border-radius: 2px;
  background: color-mix(in srgb, var(--muted) 16%, transparent);
  overflow: hidden;
}
.cat-bar i { display: block; height: 100%; border-radius: 2px; }

.week {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 6px;
  height: 128px;
}
.day {
  display: grid;
  grid-template-rows: auto 1fr auto;
  justify-items: center;
  gap: 4px;
  min-width: 0;
  padding: 2px 0;
  border-radius: 8px;
  color: inherit;
  text-decoration: none;
}
.day:hover { background: var(--primary-soft); }
.day-val { font-size: 0.68rem; color: var(--muted); font-variant-numeric: tabular-nums; min-height: 1em; }
.day-track {
  display: flex;
  align-items: flex-end;
  width: 60%;
  max-width: 26px;
  height: 100%;
}
.day-track i {
  display: block;
  width: 100%;
  border-radius: 5px 5px 2px 2px;
  background: var(--expense);
  opacity: 0.75;
}
.day--today .day-track i { opacity: 1; }
.day-label { font-size: 0.7rem; color: var(--muted); white-space: nowrap; }
.day--today .day-label { color: rgb(var(--v-theme-primary)); font-weight: 700; }
</style>
