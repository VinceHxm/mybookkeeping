<template>
  <div class="ov" :class="{ loading }">
    <div class="ov-card surface-card">
      <div class="ov-label"><v-icon size="16" class="ov-ic ov-ic--expense">mdi-arrow-up-bold-circle</v-icon>{{ t('txType.expense') }}</div>
      <div class="ov-val amount-expense">{{ yuan(d.expense) }}</div>
      <div class="ov-delta" :class="deltaClass(d.expense, d.prev.expense, false)" :title="prevTitle(d.prev.expense)">
        {{ deltaText(d.expense, d.prev.expense) }}
      </div>
    </div>
    <div class="ov-card surface-card">
      <div class="ov-label"><v-icon size="16" class="ov-ic ov-ic--income">mdi-arrow-down-bold-circle</v-icon>{{ t('txType.income') }}</div>
      <div class="ov-val amount-income">{{ yuan(d.income) }}</div>
      <div class="ov-delta" :class="deltaClass(d.income, d.prev.income, true)" :title="prevTitle(d.prev.income)">
        {{ deltaText(d.income, d.prev.income) }}
      </div>
    </div>
    <div class="ov-card ov-card--net surface-card">
      <div class="ov-net-main">
        <div class="ov-label"><v-icon size="16" class="ov-ic">mdi-scale-balance</v-icon>{{ t('stats.net') }}</div>
        <div class="ov-val" :class="d.net >= 0 ? 'amount-income' : 'amount-expense'">{{ yuan(d.net) }}</div>
        <div class="ov-delta" :class="deltaClass(d.net, d.prev.net, true)" :title="prevTitle(d.prev.net)">
          {{ deltaText(d.net, d.prev.net) }}
        </div>
      </div>
      <div class="ov-rate">
        <v-progress-circular
          :model-value="Math.max(0, Math.min(100, savingRate * 100))"
          :color="savingRate >= 0 ? 'primary' : 'error'"
          :size="56"
          :width="6"
        >
          <span class="ov-rate-num">{{ d.income ? percent(savingRate, 0) : '—' }}</span>
        </v-progress-circular>
        <span class="ov-rate-label">{{ t('stats.ov.savingRate') }}</span>
      </div>
    </div>
    <div class="ov-mini surface-card">
      <div class="mini">
        <span class="mini-label">{{ t('stats.ov.dailyExpense') }}</span>
        <span class="mini-val">{{ yuan(Math.round(d.expense / Math.max(1, d.days))) }}</span>
      </div>
      <div class="mini">
        <span class="mini-label">{{ t('stats.ov.recordCount') }}</span>
        <span class="mini-val">{{ d.expenseCount + d.incomeCount }}<small> {{ t('stats.ov.countUnit') }}</small></span>
      </div>
      <div class="mini">
        <span class="mini-label">{{ t('stats.ov.maxSingle') }}</span>
        <span class="mini-val">{{ yuan(d.maxExpense) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import http from '../../api/http'
import { changeRatio, percent, statsParams, yuan, type StatsFilter } from '../../utils/stats'

interface Totals {
  from: string
  to: string
  expense: number
  income: number
  net: number
  expenseCount: number
  incomeCount: number
}
interface Overview extends Totals {
  days: number
  maxExpense: number
  prev: Totals
}

const props = defineProps<{ filter: StatsFilter }>()
const { t } = useI18n()

const empty: Totals = { from: '', to: '', expense: 0, income: 0, net: 0, expenseCount: 0, incomeCount: 0 }
const data = ref<Overview | null>(null)
const loading = ref(false)
const d = computed<Overview>(() => data.value || { ...empty, days: 1, maxExpense: 0, prev: { ...empty } })
const savingRate = computed(() => (d.value.income ? d.value.net / d.value.income : 0))

function deltaText(cur: number, prev: number) {
  const r = changeRatio(cur, prev)
  if (r === null) return t(cur ? 'stats.ov.noPrev' : 'stats.ov.flat')
  if (Math.abs(r) < 0.0005) return t('stats.ov.flat')
  return t('stats.ov.vsPrev', { change: `${r > 0 ? '↑' : '↓'}${percent(Math.abs(r))}` })
}

/** goodWhenUp：收入、结余涨是好事；支出涨是坏事 */
function deltaClass(cur: number, prev: number, goodWhenUp: boolean) {
  const r = changeRatio(cur, prev)
  if (r === null || Math.abs(r) < 0.0005) return 'is-flat'
  return (r > 0) === goodWhenUp ? 'is-good' : 'is-bad'
}

function prevTitle(v: number) {
  const p = d.value.prev
  if (!p.from) return ''
  return t('stats.ov.prevTip', { from: p.from, to: dayjs(p.to).subtract(1, 'day').format('YYYY-MM-DD'), amount: yuan(v) })
}

async function load() {
  loading.value = true
  try {
    const { data: res } = await http.get('/stats/overview', { params: statsParams(props.filter) })
    data.value = res
  } finally {
    loading.value = false
  }
}

watch(() => [props.filter.from, props.filter.to, props.filter.accountId], load, { immediate: true })
</script>

<style scoped>
.ov {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  transition: opacity 0.2s ease;
}
.ov.loading { opacity: 0.6; }
.ov-card { padding: 12px 14px; min-width: 0; }
.ov-label {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--muted);
  font-size: 0.82rem;
}
.ov-ic { color: var(--muted); }
.ov-ic--expense { color: var(--expense); }
.ov-ic--income { color: var(--income); }
.ov-val {
  font-family: var(--font-display);
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  font-size: clamp(1.05rem, 4.6vw, 1.45rem);
  margin-top: 4px;
  overflow-wrap: anywhere;
  line-height: 1.2;
}
.ov-delta { margin-top: 4px; font-size: 0.74rem; color: var(--muted); }
.ov-delta.is-good { color: var(--income); }
.ov-delta.is-bad { color: var(--expense); }

.ov-card--net {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  gap: 12px;
}
.ov-net-main { flex: 1; min-width: 0; }
.ov-rate { display: flex; flex-direction: column; align-items: center; gap: 2px; flex-shrink: 0; }
.ov-rate-num { font-size: 0.78rem; font-weight: 700; font-variant-numeric: tabular-nums; }
.ov-rate-label { font-size: 0.7rem; color: var(--muted); }

.ov-mini {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  padding: 10px 4px;
}
.mini {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  min-width: 0;
  padding: 0 6px;
}
.mini + .mini { border-left: 1px solid var(--surface-border); }
.mini-label { font-size: 0.72rem; color: var(--muted); }
.mini-val {
  font-weight: 600;
  font-size: 0.9rem;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
  text-align: center;
}
.mini-val small { font-weight: 400; color: var(--muted); }

@media (min-width: 600px) {
  .ov { grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; }
  .ov-card--net { grid-column: auto; }
  .ov-card { padding: 16px; }
}

/* 宽屏：四列，补充指标竖排成第四张卡 */
@media (min-width: 960px) {
  .ov { grid-template-columns: repeat(3, minmax(0, 1fr)) minmax(0, 0.9fr); }
  .ov-mini {
    grid-column: auto;
    grid-template-columns: 1fr;
    padding: 8px 16px;
    align-content: center;
  }
  .mini {
    flex-direction: row;
    justify-content: space-between;
    padding: 5px 0;
  }
  .mini + .mini { border-left: 0; border-top: 1px dashed var(--surface-border); }
  .mini-label { font-size: 0.8rem; }
}
</style>
