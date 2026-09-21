<template>
  <div class="page home-page">
    <section class="home-top">
      <header class="summary surface-card anim-in" style="--d: 0ms">
        <div class="summary-grid">
          <div class="summary-cell">
            <div class="muted">本月支出</div>
            <div class="big amount-expense">¥{{ fenToYuan(summary?.monthExpense || 0) }}</div>
          </div>
          <div class="summary-divider" aria-hidden="true" />
          <div class="summary-cell">
            <div class="muted">本月收入</div>
            <div class="big amount-income">¥{{ fenToYuan(summary?.monthIncome || 0) }}</div>
          </div>
        </div>
      </header>

      <div class="cta-wrap anim-in" style="--d: 80ms">
        <div class="cta-stack">
          <div v-if="repayReminders.length" class="repay-tags">
            <router-link
              v-for="r in repayReminders"
              :key="r.accountId"
              class="repay-tag"
              :to="repayRoute(r)"
            >
              <v-icon size="14" start>mdi-credit-card-clock-outline</v-icon>
              <span class="repay-tag-text">
                {{ r.accountName }}还 ¥{{ fenToYuan(r.amountFen) }}
                <small>{{ r.daysUntilDue === 0 ? '今天到期' : '明天到期' }}</small>
              </span>
            </router-link>
          </div>
          <router-link to="/transactions/new" class="fab-record" aria-label="记一笔">
            <span class="fab-ring" aria-hidden="true" />
            <span class="fab-core">
              <v-icon size="36">mdi-plus</v-icon>
              <span class="fab-label">记一笔</span>
            </span>
          </router-link>
        </div>
      </div>
    </section>

    <section class="home-bottom anim-in" style="--d: 140ms">
      <div class="bottom-head">
        <h2 class="section-label tight">最近流水</h2>
        <router-link v-if="tx.recent.length" class="link-more" to="/transactions">全部</router-link>
      </div>

      <div class="recent-scroll hide-scrollbar">
        <div v-if="!tx.recent.length" class="empty-tip">还没有记录，点上方按钮开始记账</div>
        <v-list v-else bg-color="transparent" class="recent">
          <v-list-item
            v-for="(item, i) in tx.recent"
            :key="item.id"
            :to="`/transactions/${item.id}`"
            rounded="lg"
            class="mb-2 surface-row row-in"
            :style="{ '--d': `${160 + i * 40}ms` }"
          >
            <template #prepend>
              <v-avatar color="primary" variant="tonal" size="40">
                <span>{{ typeLabel(item.type).slice(0, 1) }}</span>
              </v-avatar>
            </template>
            <v-list-item-title>{{ item.category?.name || typeLabel(item.type) }}</v-list-item-title>
            <v-list-item-subtitle>
              {{ item.account?.name }} · {{ dayjs(item.happenedAt).format('MM-DD HH:mm') }}
              <span v-if="item.remark"> · {{ item.remark }}</span>
            </v-list-item-subtitle>
            <template #append>
              <span :class="item.type === 'income' ? 'amount-income' : 'amount-expense'">
                {{ formatMoney(item.amount, item.type) }}
              </span>
            </template>
          </v-list-item>
        </v-list>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import http from '../api/http'
import { useTransactionStore } from '../stores/transaction'
import { useAccountStore, type CreditRepayReminder } from '../stores/account'
import { fenToYuan, formatMoney, typeLabel } from '../utils/money'

const tx = useTransactionStore()
const accounts = useAccountStore()
const summary = ref<{ monthExpense: number; monthIncome: number } | null>(null)
const repayReminders = ref<CreditRepayReminder[]>([])

function repayRoute(r: CreditRepayReminder) {
  const q: Record<string, string> = {
    mode: 'repay',
    toAccountId: String(r.accountId),
    amountFen: String(r.amountFen),
  }
  if (r.fromAccountId) q.fromAccountId = String(r.fromAccountId)
  return { path: '/transactions/new', query: q }
}

onMounted(async () => {
  await Promise.all([
    tx.loadRecent(),
    http.get('/stats/summary').then((r) => { summary.value = r.data }),
    accounts.creditRepayReminders().then((list) => { repayReminders.value = list }).catch(() => {}),
  ])
})
</script>

<style scoped>
.home-page {
  display: flex;
  flex-direction: column;
  height: var(--app-height);
  max-height: var(--app-height);
  min-height: 0;
  padding-top: max(12px, var(--sat));
  padding-bottom: 0;
  overflow: hidden;
  animation: none;
}

.app-shell--mobile-nav .home-page {
  height: calc(var(--app-height) - var(--nav-total-h));
  max-height: calc(var(--app-height) - var(--nav-total-h));
}

@media (min-width: 960px) {
  .home-page {
    height: auto;
    min-height: calc(var(--app-height) - 24px);
    max-height: calc(var(--app-height) - 24px);
    padding-top: 16px;
  }
}

.home-top {
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.summary {
  padding: 14px 12px;
}

.summary-grid {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: stretch;
  gap: 0;
}

.summary-cell {
  padding: 4px 10px;
  text-align: center;
  min-width: 0;
}

.summary-divider {
  width: 1px;
  background: color-mix(in srgb, var(--muted) 28%, transparent);
  margin: 4px 0;
}

.muted { color: var(--muted); font-size: 0.82rem; }

.big {
  font-family: var(--font-display);
  font-size: clamp(1.35rem, 5vw, 1.75rem);
  margin: 6px 0 0;
  letter-spacing: -0.02em;
  font-variant-numeric: tabular-nums;
  line-height: 1.15;
  word-break: break-all;
}

.cta-wrap {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 8px 0 4px;
}

.cta-stack {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.repay-tags {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
  max-width: min(420px, 100%);
  padding: 0 8px;
}

.repay-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 100%;
  padding: 8px 12px;
  border-radius: 999px;
  text-decoration: none;
  color: #8a3b1d;
  background: color-mix(in srgb, #c45c3e 16%, var(--surface-solid, #fff));
  border: 1px solid color-mix(in srgb, #c45c3e 35%, transparent);
  font-size: 0.82rem;
  font-weight: 700;
  line-height: 1.25;
  box-shadow: var(--shadow-soft);
  -webkit-tap-highlight-color: transparent;
}

.repay-tag-text {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 4px 8px;
  min-width: 0;
}

.repay-tag small {
  font-weight: 600;
  font-size: 0.72rem;
  color: color-mix(in srgb, #8a3b1d 75%, var(--muted));
}

.fab-record {
  position: relative;
  width: clamp(112px, 30vw, 136px);
  height: clamp(112px, 30vw, 136px);
  border-radius: 50%;
  display: grid;
  place-items: center;
  text-decoration: none;
  color: #fff;
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
  outline: none;
}

.fab-ring {
  position: absolute;
  inset: -6px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(27, 127, 90, 0.28), transparent 68%);
  animation: fab-pulse 2.4s ease-in-out infinite;
  pointer-events: none;
}

.fab-core {
  position: relative;
  z-index: 1;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  background: linear-gradient(145deg, #24956b 0%, #1b7f5a 48%, #145a41 100%);
  box-shadow:
    0 14px 32px rgba(27, 127, 90, 0.36),
    0 2px 0 rgba(255, 255, 255, 0.18) inset;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.fab-record:active .fab-core,
.fab-record:hover .fab-core {
  transform: scale(0.96);
  box-shadow: 0 8px 20px rgba(27, 127, 90, 0.3);
}

.fab-label {
  font-family: var(--font-display);
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.home-bottom {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  margin-top: 4px;
}

.bottom-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  flex: 0 0 auto;
  padding-right: 4px;
}

.section-label.tight {
  margin: 4px 0 8px;
}

.link-more {
  color: var(--primary);
  font-size: 0.86rem;
  text-decoration: none;
  font-weight: 600;
}

.recent-scroll {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  -webkit-overflow-scrolling: touch;
  overscroll-behavior: contain;
  padding-bottom: 8px;
  touch-action: pan-y;
}

.anim-in {
  animation: rise-in 0.45s cubic-bezier(0.22, 1, 0.36, 1) both;
  animation-delay: var(--d, 0ms);
}

.row-in {
  animation: rise-in 0.4s cubic-bezier(0.22, 1, 0.36, 1) both;
  animation-delay: var(--d, 0ms);
}

@keyframes rise-in {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: none; }
}

@keyframes fab-pulse {
  0%, 100% { transform: scale(1); opacity: 0.7; }
  50% { transform: scale(1.08); opacity: 1; }
}

@media (prefers-reduced-motion: reduce) {
  .anim-in,
  .row-in,
  .fab-ring { animation: none !important; }
}

[data-theme='dark'] .fab-core {
  background: linear-gradient(145deg, #4aaf84 0%, #3d9b74 50%, #2d7a5a 100%);
  box-shadow: 0 14px 32px rgba(0, 0, 0, 0.45), 0 2px 0 rgba(255, 255, 255, 0.1) inset;
}
</style>
