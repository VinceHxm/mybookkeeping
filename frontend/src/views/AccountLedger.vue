<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/accounts" />
      <h1>{{ ledger?.accountName || '账户明细' }}</h1>
      <span class="toolbar-gap" />
    </div>

    <div v-if="loading" class="empty-tip">加载中…</div>
    <div v-else-if="error" class="empty-tip err">{{ error }}</div>
    <div v-else-if="ledger" class="ledger-layout">
      <aside class="ledger-side">
      <div class="bal-card">
        <span>{{ isCredit ? '已用额度' : isInvest ? '当前市值' : '当前余额' }}</span>
        <strong>¥{{ fenToYuan(isCredit ? Math.max(0, -ledger.balance) : ledger.balance) }}</strong>
      </div>

      <template v-if="ledger.invest">
        <div class="status-tiles">
          <div class="tile static">
            <span>净投入</span>
            <strong>¥{{ fenToYuan(ledger.invest.netInvested) }}</strong>
          </div>
          <div class="tile static">
            <span>累计收益</span>
            <strong :class="ledger.invest.gain >= 0 ? 'amount-income' : 'amount-expense'">
              {{ formatMoney(ledger.invest.gain, 'revalue') }}
            </strong>
          </div>
          <div class="tile static">
            <span>收益率</span>
            <strong :class="ledger.invest.gain >= 0 ? 'amount-income' : 'amount-expense'">{{ returnRate }}</strong>
          </div>
        </div>
        <p class="ledger-tip">
          净投入 = 建账余额 + 转入 − 转出；累计收益 = 当前市值 − 净投入（含市值校准、分红、手续费）。收益率为简单收益率，不考虑持有时长。
        </p>
        <v-btn block variant="tonal" color="primary" prepend-icon="mdi-chart-line" class="mb-1" @click="openRevalue(null)">
          校准市值
        </v-btn>
      </template>

      <template v-if="ledger.creditTracking">
        <div class="status-tiles">
          <button
            v-for="s in statusTiles"
            :key="s.key"
            type="button"
            class="tile"
            :class="[s.key, { active: statusFilter === s.key }]"
            @click="toggleFilter(s.key)"
          >
            <span>{{ s.label }}</span>
            <strong>¥{{ fenToYuan(s.amount) }}</strong>
          </button>
        </div>
        <p class="ledger-tip">
          已出账至 {{ ledger.summary.lastStatementDate }}，下次出账 {{ ledger.summary.nextStatementDate }}。
          <template v-if="ledger.summary.openingDebt > 0">
            建账/校准带入欠款 ¥{{ fenToYuan(ledger.summary.openingDebt) }}（无对应流水）<template v-if="ledger.summary.openingDebtUnpaid > 0">，其中未还 ¥{{ fenToYuan(ledger.summary.openingDebtUnpaid) }}</template>。
          </template>
          {{ ledger.note }}
          <template v-if="statusFilter">点击上方同一格可取消筛选。</template>
        </p>
      </template>
      <p v-else-if="ledger.note" class="ledger-tip">{{ ledger.note }}</p>
      </aside>

      <div class="ledger-main">
      <div v-if="!months.length" class="empty-tip">暂无流水</div>
      <section v-for="m in months" :key="m.month" class="month">
        <button type="button" class="month-head" @click="toggleMonth(m.month)">
          <div class="month-title">
            <v-icon size="18">{{ isOpen(m.month) ? 'mdi-chevron-down' : 'mdi-chevron-right' }}</v-icon>
            <span>{{ monthLabel(m.month) }}</span>
            <span class="month-count">{{ m.items.length }} 笔</span>
          </div>
          <div class="month-sum">
            <span class="amount-expense">{{ isCredit ? '消费' : '流出' }} ¥{{ fenToYuan(m.outAmount) }}</span>
            <span class="amount-income">{{ isCredit ? '还款/退款' : '流入' }} ¥{{ fenToYuan(m.inAmount) }}</span>
            <span v-if="m.revalued" :class="m.revalued > 0 ? 'amount-income' : 'amount-expense'">
              市值变动 {{ formatMoney(m.revalued, 'revalue') }}
            </span>
          </div>
        </button>
        <div v-if="ledger.creditTracking && isOpen(m.month) && !statusFilter" class="month-status">
          <span v-if="m.billedUnpaid" class="st billed_unpaid">未还 ¥{{ fenToYuan(m.billedUnpaid) }}</span>
          <span v-if="m.paid" class="st paid">已还 ¥{{ fenToYuan(m.paid) }}</span>
          <span v-if="m.unbilled" class="st unbilled">未出账 ¥{{ fenToYuan(m.unbilled) }}</span>
        </div>

        <v-list v-if="isOpen(m.month)" bg-color="transparent" class="py-0">
          <v-list-item
            v-for="item in m.items"
            :key="item.id"
            :to="item.type === 'revalue' ? undefined : `/transactions/${item.id}`"
            rounded="lg"
            class="tx-row surface-row"
            @click="item.type === 'revalue' && openRevalue(item)"
          >
            <v-list-item-title class="tx-title">
              <span class="tx-name">{{ txTitle(item) }}</span>
              <span v-if="item.creditStatus" class="st" :class="item.creditStatus">{{ statusLabel[item.creditStatus] }}</span>
              <span v-else-if="isCredit && item.direction === 'in'" class="st repay">{{ item.type === 'income' ? '退款' : '还款' }}</span>
              <span v-else-if="item.type === 'revalue'" class="st unbilled">市值校准</span>
            </v-list-item-title>
            <v-list-item-subtitle>
              {{ dayjs(item.happenedAt).format('MM-DD HH:mm') }}
              <template v-if="txSubCategory(item)"> · {{ txSubCategory(item) }}</template>
              <template v-if="item.type === 'transfer'">
                · {{ item.direction === 'out' ? `转至 ${item.toAccount?.name || ''}` : `来自 ${item.account?.name || ''}` }}
              </template>
            </v-list-item-subtitle>
            <div v-if="item.creditStatus" class="credit-detail">
              <template v-if="(item.installmentPeriods || 1) > 1">
                分期 {{ item.billedPeriods || 0 }}/{{ item.installmentPeriods }} 期已出账 ·
              </template>
              <template v-else-if="item.creditStatus === 'unbilled'">将于 {{ item.statementDate }} 出账</template>
              <template v-else>{{ item.statementDate }} 出账</template>
              <template v-if="isSplit(item)">
                <span v-if="item.paidAmount"> · 已还 ¥{{ fenToYuan(item.paidAmount) }}</span>
                <span v-if="item.billedUnpaidAmount" class="warn"> · 未还 ¥{{ fenToYuan(item.billedUnpaidAmount) }}</span>
                <span v-if="item.unbilledAmount"> · 未出账 ¥{{ fenToYuan(item.unbilledAmount) }}</span>
              </template>
            </div>
            <template #append>
              <span :class="item.direction === 'in' ? 'amount-income' : 'amount-expense'">
                {{ item.direction === 'in' ? '+' : '-' }}¥{{ fenToYuan(Math.abs(item.amount)) }}
              </span>
            </template>
          </v-list-item>
        </v-list>
      </section>
      </div>
    </div>

    <RevalueDialog
      v-model="revalueOpen"
      :account="ledger ? { id: ledger.accountId, name: ledger.accountName, balance: ledger.balance } : null"
      :edit-tx="revalueTx"
      @saved="reload"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import dayjs from 'dayjs'
import {
  useAccountStore,
  type AccountLedger,
  type AccountLedgerItem,
  type AccountLedgerMonth,
  type CreditTxStatus,
} from '../stores/account'
import type { Transaction } from '../stores/transaction'
import RevalueDialog from '../components/RevalueDialog.vue'
import { fenToYuan, formatMoney, txSubCategory, txTitle } from '../utils/money'

const route = useRoute()
const accounts = useAccountStore()

const ledger = ref<AccountLedger | null>(null)
const loading = ref(true)
const error = ref('')
const statusFilter = ref<CreditTxStatus | ''>('')
const openMonths = ref<Set<string>>(new Set())

const statusLabel: Record<CreditTxStatus, string> = {
  billed_unpaid: '已出账未还',
  paid: '已还',
  unbilled: '未出账',
}

const isCredit = computed(() => ledger.value?.accountType === 'credit')
const isInvest = computed(() => ledger.value?.accountType === 'invest')
const revalueOpen = ref(false)
const revalueTx = ref<Transaction | null>(null)

const returnRate = computed(() => {
  const inv = ledger.value?.invest
  if (!inv || inv.netInvested <= 0) return '—'
  const pct = (inv.gain / inv.netInvested) * 100
  return `${pct >= 0 ? '+' : ''}${pct.toFixed(2)}%`
})

function openRevalue(tx: Transaction | null) {
  revalueTx.value = tx
  revalueOpen.value = true
}

const statusTiles = computed(() => {
  const s = ledger.value?.summary
  return [
    { key: 'billed_unpaid' as const, label: '已出账未还', amount: s?.billedUnpaid || 0 },
    { key: 'paid' as const, label: '已还', amount: s?.paid || 0 },
    { key: 'unbilled' as const, label: '未出账', amount: s?.unbilled || 0 },
  ]
})

const months = computed<AccountLedgerMonth[]>(() => {
  const list = ledger.value?.months || []
  if (!statusFilter.value) return list
  return list
    .map((m) => ({ ...m, items: m.items.filter((i) => i.creditStatus === statusFilter.value) }))
    .filter((m) => m.items.length)
})

function isOpen(month: string) {
  return !!statusFilter.value || openMonths.value.has(month)
}

function toggleMonth(month: string) {
  if (statusFilter.value) return
  const next = new Set(openMonths.value)
  if (next.has(month)) next.delete(month)
  else next.add(month)
  openMonths.value = next
}

function toggleFilter(key: CreditTxStatus) {
  statusFilter.value = statusFilter.value === key ? '' : key
}

function monthLabel(month: string) {
  const d = dayjs(`${month}-01`)
  return d.year() === dayjs().year() ? `${d.month() + 1} 月` : d.format('YYYY 年 M 月')
}

function isSplit(item: AccountLedgerItem) {
  return [item.paidAmount, item.billedUnpaidAmount, item.unbilledAmount].filter((v) => v > 0).length > 1
}

async function reload() {
  try {
    ledger.value = await accounts.ledger(Number(route.params.id))
    const first = ledger.value.months[0]?.month
    if (first && !openMonths.value.size) openMonths.value = new Set([first])
  } catch (e: any) {
    error.value = e?.message || '加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(reload)
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; }
.toolbar h1 { font-size: 1.15rem; margin: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.toolbar-gap { width: 48px; }
.err { color: #c45c3e; }

.bal-card {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  padding: 12px 14px;
  margin-bottom: 10px;
  border-radius: 12px;
  border: 1px solid var(--surface-border);
  background: var(--surface-solid);
}
.bal-card span { color: var(--muted); font-size: 0.8rem; }
.bal-card strong { font-family: var(--font-display); font-variant-numeric: tabular-nums; font-size: 1.15rem; }

.status-tiles { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.tile {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px;
  border-radius: 12px;
  border: 1px solid var(--surface-border);
  background: var(--surface-solid);
  color: inherit;
  text-align: left;
  cursor: pointer;
  min-width: 0;
}
.tile span { font-size: 0.72rem; color: var(--muted); white-space: nowrap; }
.tile strong {
  font-family: var(--font-display);
  font-variant-numeric: tabular-nums;
  font-size: clamp(0.8rem, 3.4vw, 0.95rem);
  white-space: nowrap;
}
.tile.static { cursor: default; }
.tile.billed_unpaid strong { color: #c45c3e; }
.tile.paid strong { color: var(--primary); }
.tile.active { border-color: currentColor; box-shadow: 0 0 0 1px var(--primary); }
.tile.billed_unpaid.active { box-shadow: 0 0 0 1px #c45c3e; }

.ledger-tip { margin: 8px 2px 4px; font-size: 0.74rem; color: var(--muted); line-height: 1.45; }

.month { margin-top: 12px; }
.month-head {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  padding: 6px 4px;
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
}
.month-title { display: flex; align-items: center; gap: 4px; font-weight: 700; font-size: 0.92rem; }
.month-count { font-weight: 500; font-size: 0.72rem; color: var(--muted); margin-left: 4px; }
.month-sum { display: flex; flex-direction: column; align-items: flex-end; font-size: 0.72rem; font-variant-numeric: tabular-nums; }
.month-status { display: flex; flex-wrap: wrap; gap: 6px; margin: 0 4px 6px; }

.st {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 6px;
  font-size: 0.68rem;
  font-weight: 600;
  line-height: 1.5;
  white-space: nowrap;
}
.st.billed_unpaid { color: #c45c3e; background: rgba(196, 92, 62, 0.14); }
.st.paid { color: var(--primary); background: var(--primary-soft, rgba(27, 127, 90, 0.12)); }
.st.unbilled { color: var(--muted); background: rgba(128, 128, 128, 0.14); }
.st.repay { color: #3d8fad; background: rgba(61, 143, 173, 0.14); }

.tx-row { margin-bottom: 6px; min-height: 52px !important; }
.tx-row :deep(.v-list-item-subtitle) { font-size: 0.76rem; }
.tx-title { display: flex; align-items: center; gap: 6px; }
.tx-name { overflow: hidden; text-overflow: ellipsis; }
.credit-detail { margin-top: 2px; font-size: 0.72rem; color: var(--muted); }
.credit-detail .warn { color: #c45c3e; }

/* 平板：月份汇总横排，少占一行 */
@media (min-width: 600px) {
  .month-sum { flex-direction: row; gap: 14px; font-size: 0.78rem; }
}

/* 宽屏：左侧余额与状态汇总吸顶，右侧按月流水 */
@media (min-width: 960px) {
  .ledger-layout {
    display: grid;
    grid-template-columns: 300px minmax(0, 1fr);
    gap: 24px;
    align-items: start;
  }
  .ledger-side { position: sticky; top: 16px; }
  .bal-card { flex-direction: column; gap: 4px; padding: 18px; }
  .bal-card strong { font-size: 1.6rem; }
  .ledger-side .status-tiles { grid-template-columns: 1fr; }
  .ledger-side .tile { flex-direction: row; justify-content: space-between; align-items: baseline; padding: 12px 14px; }
  .ledger-side .tile span { font-size: 0.8rem; }
  .ledger-side .tile strong { font-size: 1rem; }
  .ledger-main .month:first-child { margin-top: 0; }
  .month {
    padding: 6px 10px 8px;
    border-radius: var(--radius-md);
    border: 1px solid var(--surface-border);
    background: color-mix(in srgb, var(--surface) 55%, transparent);
  }
}
</style>
