<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/accounts" />
      <h1>{{ ledger?.accountName || t('ledger.title') }}</h1>
      <span class="toolbar-gap" />
    </div>

    <div v-if="loading" class="empty-tip">{{ t('common.loading') }}</div>
    <div v-else-if="error" class="empty-tip err">{{ error }}</div>
    <div v-else-if="ledger" class="ledger-layout">
      <aside class="ledger-side">
      <div class="bal-card">
        <span>{{ isCredit ? t('ledger.usedCredit') : isInvest ? t('ledger.marketValue') : t('ledger.balance') }}</span>
        <strong>¥{{ fenToYuan(isCredit ? Math.max(0, -ledger.balance) : ledger.balance) }}</strong>
      </div>

      <template v-if="ledger.invest">
        <div class="status-tiles">
          <div class="tile static">
            <span>{{ t('ledger.netInvested') }}</span>
            <strong>¥{{ fenToYuan(ledger.invest.netInvested) }}</strong>
          </div>
          <div class="tile static">
            <span>{{ t('ledger.gain') }}</span>
            <strong :class="ledger.invest.gain >= 0 ? 'amount-income' : 'amount-expense'">
              {{ formatMoney(ledger.invest.gain, 'revalue') }}
            </strong>
          </div>
          <div class="tile static">
            <span>{{ t('ledger.returnRate') }}</span>
            <strong :class="ledger.invest.gain >= 0 ? 'amount-income' : 'amount-expense'">{{ returnRate }}</strong>
          </div>
        </div>
        <p class="ledger-tip">
          {{ t('ledger.investTip') }}
        </p>
        <v-btn block variant="tonal" color="primary" prepend-icon="mdi-chart-line" class="mb-1" @click="openRevalue(null)">
          {{ t('ledger.revalue') }}
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
          {{ t('ledger.billedUntil', { last: ledger.summary.lastStatementDate, next: ledger.summary.nextStatementDate }) }}
          <template v-if="ledger.summary.openingDebt > 0">
            {{ t('ledger.openingDebt', { amount: fenToYuan(ledger.summary.openingDebt) }) }}<template v-if="ledger.summary.openingDebtUnpaid > 0">{{ t('ledger.openingDebtUnpaid', { amount: fenToYuan(ledger.summary.openingDebtUnpaid) }) }}</template>{{ t('ledger.sentenceEnd') }}
          </template>
          {{ ledgerNote }}
          <template v-if="statusFilter">{{ t('ledger.clearFilter') }}</template>
        </p>
      </template>
      <p v-else-if="ledgerNote" class="ledger-tip">{{ ledgerNote }}</p>
      </aside>

      <div class="ledger-main">
      <div v-if="!months.length" class="empty-tip">{{ t('ledger.empty') }}</div>
      <section v-for="m in months" :key="m.month" class="month">
        <button type="button" class="month-head" @click="toggleMonth(m.month)">
          <div class="month-title">
            <v-icon size="18">{{ isOpen(m.month) ? 'mdi-chevron-down' : 'mdi-chevron-right' }}</v-icon>
            <span>{{ monthLabel(m.month) }}</span>
            <span class="month-count">{{ t('ledger.count', { n: m.items.length }) }}</span>
          </div>
          <div class="month-sum">
            <span class="amount-expense">{{ isCredit ? t('ledger.spent') : t('ledger.outflow') }} ¥{{ fenToYuan(m.outAmount) }}</span>
            <span class="amount-income">{{ isCredit ? t('ledger.repayRefund') : t('ledger.inflow') }} ¥{{ fenToYuan(m.inAmount) }}</span>
            <span v-if="m.revalued" :class="m.revalued > 0 ? 'amount-income' : 'amount-expense'">
              {{ t('ledger.revalued') }} {{ formatMoney(m.revalued, 'revalue') }}
            </span>
          </div>
        </button>
        <div v-if="ledger.creditTracking && isOpen(m.month) && !statusFilter" class="month-status">
          <span v-if="m.billedUnpaid" class="st billed_unpaid">{{ t('ledger.unpaid', { amount: fenToYuan(m.billedUnpaid) }) }}</span>
          <span v-if="m.paid" class="st paid">{{ t('ledger.paid', { amount: fenToYuan(m.paid) }) }}</span>
          <span v-if="m.unbilled" class="st unbilled">{{ t('ledger.unbilled', { amount: fenToYuan(m.unbilled) }) }}</span>
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
              <span v-if="item.creditStatus" class="st" :class="item.creditStatus">{{ t(`ledger.status.${item.creditStatus}`) }}</span>
              <span v-else-if="isCredit && item.direction === 'in'" class="st repay">{{ item.type === 'income' ? t('ledger.refund') : t('ledger.repay') }}</span>
              <span v-else-if="item.type === 'revalue'" class="st unbilled">{{ t('txType.revalue') }}</span>
            </v-list-item-title>
            <v-list-item-subtitle>
              {{ dayjs(item.happenedAt).format('MM-DD HH:mm') }}
              <template v-if="txSubCategory(item)"> · {{ txSubCategory(item) }}</template>
              <template v-if="item.type === 'transfer'">
                · {{ item.direction === 'out'
                  ? t('ledger.transferTo', { name: item.toAccount?.name || '' })
                  : t('ledger.transferFrom', { name: item.account?.name || '' }) }}
              </template>
            </v-list-item-subtitle>
            <div v-if="item.creditStatus" class="credit-detail">
              <template v-if="(item.installmentPeriods || 1) > 1">
                {{ t('ledger.installmentProgress', { billed: item.billedPeriods || 0, total: item.installmentPeriods }) }}
              </template>
              <template v-else-if="item.creditStatus === 'unbilled'">{{ t('ledger.willBill', { date: item.statementDate }) }}</template>
              <template v-else>{{ t('ledger.billedOn', { date: item.statementDate }) }}</template>
              <template v-if="isSplit(item)">
                <span v-if="item.paidAmount"> · {{ t('ledger.paid', { amount: fenToYuan(item.paidAmount) }) }}</span>
                <span v-if="item.billedUnpaidAmount" class="warn"> · {{ t('ledger.unpaid', { amount: fenToYuan(item.billedUnpaidAmount) }) }}</span>
                <span v-if="item.unbilledAmount"> · {{ t('ledger.unbilled', { amount: fenToYuan(item.unbilledAmount) }) }}</span>
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
import { useI18n } from 'vue-i18n'
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

const { t } = useI18n()
const route = useRoute()
const accounts = useAccountStore()

const ledger = ref<AccountLedger | null>(null)
const loading = ref(true)
const error = ref('')
const statusFilter = ref<CreditTxStatus | ''>('')
const openMonths = ref<Set<string>>(new Set())

const isCredit = computed(() => ledger.value?.accountType === 'credit')
const isInvest = computed(() => ledger.value?.accountType === 'invest')

/** 后端 note 为固定中文说明，信用卡按同样条件在前端出本地化文案 */
const ledgerNote = computed(() => {
  const l = ledger.value
  if (!l) return ''
  if (isCredit.value) return l.creditTracking ? t('ledger.noteFifo') : t('ledger.noteNoBillingDay')
  return l.note || ''
})
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
    { key: 'billed_unpaid' as const, label: t('ledger.status.billed_unpaid'), amount: s?.billedUnpaid || 0 },
    { key: 'paid' as const, label: t('ledger.status.paid'), amount: s?.paid || 0 },
    { key: 'unbilled' as const, label: t('ledger.status.unbilled'), amount: s?.unbilled || 0 },
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
  return d.format(d.year() === dayjs().year() ? t('ledger.monthFmt') : t('ledger.yearMonthFmt'))
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
    error.value = e?.message || t('ledger.loadFailed')
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
