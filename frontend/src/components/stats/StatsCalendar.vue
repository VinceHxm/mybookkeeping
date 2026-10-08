<template>
  <div class="cal">
    <div class="cal-controls">
      <div class="month-nav">
        <v-btn icon="mdi-chevron-left" variant="text" size="small" aria-label="上个月" @click="shift(-1)" />
        <span class="month-title">{{ month.format('YYYY年M月') }}</span>
        <v-btn icon="mdi-chevron-right" variant="text" size="small" aria-label="下个月" :disabled="isCurrentMonth" @click="shift(1)" />
      </div>
      <v-btn-toggle v-model="kind" mandatory density="compact" color="primary" variant="outlined" divided rounded="lg">
        <v-btn value="expense" size="small">支出</v-btn>
        <v-btn value="income" size="small">收入</v-btn>
      </v-btn-toggle>
    </div>

    <div class="cal-grid-wrap" :class="{ loading }">
      <div class="cal-board surface-card" :style="{ '--heat': heatColor }">
        <div class="cal-week">
          <span v-for="w in weekLabels" :key="w">{{ w }}</span>
        </div>
        <div class="cal-days">
          <span v-for="n in lead" :key="`b${n}`" class="cell cell--blank" />
          <button
            v-for="c in cells"
            :key="c.date"
            type="button"
            class="cell"
            :class="[`lv${c.level}`, { today: c.today, future: c.future }]"
            :disabled="c.future"
            :title="`${c.date} ${kindLabel} ${yuan(c.amount)}（${c.count} 笔）`"
            @click="openDay(c.date)"
          >
            <span class="cell-day">{{ c.day }}</span>
            <span v-if="c.amount" class="cell-amt">
              <span class="amt-short">{{ compactYuan(c.amount) }}</span>
              <span class="amt-full">{{ yuan(c.amount) }}</span>
            </span>
          </button>
        </div>
        <div class="cal-legend">
          <span>少</span>
          <i v-for="lv in 4" :key="lv" :class="`lv${lv}`" />
          <span>多</span>
        </div>
      </div>

      <div class="cal-side surface-card">
        <div class="side-item">
          <span class="side-label">本月{{ kindLabel }}</span>
          <span class="side-val" :class="kind === 'income' ? 'amount-income' : 'amount-expense'">{{ yuan(sum) }}</span>
        </div>
        <div class="side-item">
          <span class="side-label">有{{ kindLabel }}的天数</span>
          <span class="side-val">{{ activeDays }} / {{ passedDays }} 天</span>
        </div>
        <div class="side-item">
          <span class="side-label">日均{{ kindLabel }}</span>
          <span class="side-val">{{ yuan(Math.round(sum / Math.max(1, passedDays))) }}</span>
        </div>
        <div class="side-item">
          <span class="side-label">最高的一天</span>
          <button v-if="peak" type="button" class="side-val side-link" @click="openDay(peak.date)">
            {{ dayjs(peak.date).format('M月D日') }} · {{ yuan(peak.amount) }}
          </button>
          <span v-else class="side-val">—</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import dayjs, { type Dayjs } from 'dayjs'
import http from '../../api/http'
import { useAuthStore } from '../../stores/auth'
import { compactYuan, statsParams, yuan, type StatsFilter } from '../../utils/stats'

interface Point {
  key: string
  expense: number
  income: number
  expenseCount: number
  incomeCount: number
}

const props = defineProps<{ filter: StatsFilter }>()
const router = useRouter()
const auth = useAuthStore()

const month = ref<Dayjs>(initialMonth())
const kind = ref<'expense' | 'income'>('expense')
const points = ref<Point[]>([])
const loading = ref(false)

const kindLabel = computed(() => (kind.value === 'income' ? '收入' : '支出'))
const heatColor = computed(() => (kind.value === 'income' ? 'var(--income)' : 'var(--expense)'))
const isCurrentMonth = computed(() => !month.value.isBefore(dayjs(), 'month'))
const weekStart = computed(() => auth.profile?.weekStart ?? 1)
const weekLabels = computed(() => {
  const names = ['日', '一', '二', '三', '四', '五', '六']
  return names.map((_, i) => names[(i + weekStart.value) % 7])
})
const lead = computed(() => (month.value.startOf('month').day() - weekStart.value + 7) % 7)

/** 统计区间的结束月（不晚于本月） */
function initialMonth() {
  const end = dayjs(props.filter.to)
  return (end.isAfter(dayjs()) ? dayjs() : end).startOf('month')
}

const byDay = computed(() => {
  const m = new Map<string, { amount: number; count: number }>()
  for (const p of points.value) {
    m.set(p.key, kind.value === 'income'
      ? { amount: p.income, count: p.incomeCount }
      : { amount: p.expense, count: p.expenseCount })
  }
  return m
})
const maxAmount = computed(() => Math.max(0, ...[...byDay.value.values()].map((v) => v.amount)))

const cells = computed(() => {
  const start = month.value.startOf('month')
  const today = dayjs().format('YYYY-MM-DD')
  return Array.from({ length: start.daysInMonth() }, (_, i) => {
    const d = start.add(i, 'day')
    const date = d.format('YYYY-MM-DD')
    const v = byDay.value.get(date) || { amount: 0, count: 0 }
    const level = v.amount <= 0 || !maxAmount.value ? 0 : Math.max(1, Math.ceil((v.amount / maxAmount.value) * 4))
    return { date, day: i + 1, amount: v.amount, count: v.count, level, today: date === today, future: date > today }
  })
})

const passedDays = computed(() => cells.value.filter((c) => !c.future).length)
const sum = computed(() => cells.value.reduce((s, c) => s + c.amount, 0))
const activeDays = computed(() => cells.value.filter((c) => c.amount > 0).length)
const peak = computed(() => {
  const top = [...cells.value].sort((a, b) => b.amount - a.amount)[0]
  return top && top.amount > 0 ? top : null
})

function shift(n: number) {
  month.value = month.value.add(n, 'month')
}

function openDay(date: string) {
  const query: Record<string, string> = { from: date, to: date, type: kind.value }
  if (props.filter.accountId) query.accountId = String(props.filter.accountId)
  router.push({ path: '/transactions', query })
}

async function load() {
  loading.value = true
  try {
    const m = month.value
    const range: StatsFilter = {
      from: m.startOf('month').format('YYYY-MM-DD'),
      to: m.endOf('month').format('YYYY-MM-DD'),
      accountId: props.filter.accountId,
    }
    const { data } = await http.get('/stats/trend', { params: statsParams(range, { granularity: 'day' }) })
    points.value = data.points || []
  } finally {
    loading.value = false
  }
}

watch(() => [props.filter.from, props.filter.to], () => {
  const m = initialMonth()
  if (!m.isSame(month.value, 'month')) month.value = m
})
watch([month, () => props.filter.accountId], load, { immediate: true })
</script>

<style scoped>
.cal-controls {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}
.month-nav { display: flex; align-items: center; gap: 2px; margin-left: -8px; }
.month-title { font-weight: 700; font-family: var(--font-display); min-width: 92px; text-align: center; }

.cal-grid-wrap { display: grid; gap: 12px; transition: opacity 0.2s ease; }
.cal-grid-wrap.loading { opacity: 0.6; }
.cal-board { padding: 10px 8px 8px; min-width: 0; }
.cal-week,
.cal-days {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 4px;
}
.cal-week { margin-bottom: 4px; }
.cal-week span { text-align: center; font-size: 0.72rem; color: var(--muted); }

.cell {
  position: relative;
  min-height: 48px;
  border: 0;
  border-radius: 8px;
  padding: 4px 3px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
  background: rgba(var(--v-theme-on-surface), 0.035);
  color: inherit;
  font: inherit;
  cursor: pointer;
  min-width: 0;
}
.cell--blank { background: transparent; cursor: default; }
.cell:not(.cell--blank):not(:disabled):hover { outline: 2px solid var(--primary); outline-offset: -2px; }
.cell.today { box-shadow: inset 0 0 0 2px var(--primary); }
.cell.future { opacity: 0.4; cursor: default; }
.cell-day { font-size: 0.78rem; font-weight: 600; line-height: 1.1; }
.cell-amt {
  font-size: 0.64rem;
  font-variant-numeric: tabular-nums;
  line-height: 1.1;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.amt-full { display: none; }

.lv1 { background: color-mix(in srgb, var(--heat) 16%, transparent); }
.lv2 { background: color-mix(in srgb, var(--heat) 32%, transparent); }
.lv3 { background: color-mix(in srgb, var(--heat) 54%, transparent); }
.lv4 { background: color-mix(in srgb, var(--heat) 80%, transparent); color: #fff; }

.cal-legend {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
  margin-top: 8px;
  font-size: 0.7rem;
  color: var(--muted);
}
.cal-legend i { width: 12px; height: 12px; border-radius: 3px; display: inline-block; }

.cal-side {
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
.side-link {
  border: 0;
  padding: 0;
  background: none;
  font: inherit;
  font-weight: 600;
  text-align: left;
  color: var(--primary);
  cursor: pointer;
}

@media (max-width: 380px) {
  .cal-days, .cal-week { gap: 3px; }
  .cell { min-height: 42px; border-radius: 6px; }
  .cell-amt { font-size: 0.58rem; }
}

/* 宽屏：大格子显示完整金额，摘要在右侧 */
@media (min-width: 960px) {
  .cal-grid-wrap { grid-template-columns: minmax(0, 1fr) 240px; align-items: start; }
  .cal-board { padding: 14px; }
  .cal-week, .cal-days { gap: 6px; }
  .cell { min-height: 76px; align-items: flex-start; padding: 8px 10px; border-radius: 10px; }
  .cell-day { font-size: 0.9rem; }
  .cell-amt { font-size: 0.78rem; align-self: flex-end; }
  .amt-short { display: none; }
  .amt-full { display: inline; }
  .cal-side { grid-template-columns: 1fr; }
  .side-item { padding: 14px 16px; }
  .side-val { font-size: 0.98rem; }
}
</style>
