<template>
  <div class="page home-page">
    <section class="home-top anim-in" style="--d: 0ms">
      <header class="summary surface-card">
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
    </section>

    <section class="home-mid anim-in" style="--d: 80ms">
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
          <v-icon size="48">mdi-plus</v-icon>
          <span class="fab-label">记一笔</span>
        </span>
      </router-link>

      <div class="tpl-block">
        <div class="tpl-head">
          <span class="tpl-title">快捷模板</span>
          <router-link class="tpl-manage" to="/templates">管理</router-link>
        </div>
        <div v-if="!templates.length" class="tpl-empty">
          暂无模板，去
          <router-link to="/templates">我的 · 记账模板</router-link>
          添加后可一键入账
        </div>
        <div v-else class="tpl-grid">
          <button
            v-for="t in templates"
            :key="t.id"
            type="button"
            class="tpl-btn"
            :disabled="quickBusy != null"
            @click="askQuickUse(t)"
          >
            <v-icon size="22" class="tpl-ico">{{ t.icon || 'mdi-flash' }}</v-icon>
            <span class="tpl-name">{{ t.name }}</span>
            <span v-if="t.amount > 0" class="tpl-amt">¥{{ fenToYuan(t.amount) }}</span>
            <span v-else class="tpl-amt muted-amt">待填</span>
          </button>
        </div>
        <p class="tpl-hint">点模板会先确认再入账；要改内容请用「记一笔」</p>
      </div>
    </section>

    <section class="home-bottom anim-in" style="--d: 140ms">
      <div class="bottom-head">
        <h2 class="section-label tight">最近流水</h2>
        <router-link v-if="tx.recent.length" class="link-more" to="/transactions">全部</router-link>
      </div>

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
    </section>

    <div v-if="flash" class="home-flash" role="status">{{ flash }}</div>

    <v-dialog v-model="confirmOpen" max-width="360">
      <v-card class="confirm-card">
        <v-card-title class="confirm-title">确认记账</v-card-title>
        <v-card-text class="confirm-body">
          <template v-if="pendingTpl">
            <p class="confirm-msg">
              使用模板「<strong>{{ pendingTpl.name }}</strong>」立即记一笔？
            </p>
            <ul class="confirm-meta">
              <li>类型：{{ typeLabel(pendingTpl.type as any) }}</li>
              <li v-if="pendingTpl.amount > 0">金额：¥{{ fenToYuan(pendingTpl.amount) }}</li>
              <li v-else-if="pendingTpl.fareRuleId">金额：按计费规则计算</li>
              <li v-else>金额：未设（将打开编辑页）</li>
            </ul>
          </template>
        </v-card-text>
        <v-card-actions class="confirm-actions">
          <v-btn variant="text" :disabled="quickBusy != null" @click="cancelQuick">取消</v-btn>
          <v-spacer />
          <v-btn
            variant="tonal"
            color="primary"
            :disabled="quickBusy != null || !pendingTpl"
            @click="editInstead"
          >去编辑</v-btn>
          <v-btn
            color="primary"
            variant="flat"
            :loading="quickBusy != null"
            :disabled="!pendingTpl"
            @click="confirmQuick"
          >确认入账</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import dayjs from 'dayjs'
import http from '../api/http'
import { useTransactionStore } from '../stores/transaction'
import { useAccountStore, type CreditRepayReminder } from '../stores/account'
import { fenToYuan, formatMoney, typeLabel } from '../utils/money'
import type { Template } from '../components/TemplateEditForm.vue'

const tx = useTransactionStore()
const accounts = useAccountStore()
const router = useRouter()
const summary = ref<{ monthExpense: number; monthIncome: number } | null>(null)
const repayReminders = ref<CreditRepayReminder[]>([])
const templates = ref<Template[]>([])
const quickBusy = ref<number | null>(null)
const confirmOpen = ref(false)
const pendingTpl = ref<Template | null>(null)
const flash = ref('')
let flashTimer: ReturnType<typeof setTimeout> | null = null

function repayRoute(r: CreditRepayReminder) {
  const q: Record<string, string> = {
    mode: 'repay',
    toAccountId: String(r.accountId),
    amountFen: String(r.amountFen),
  }
  if (r.fromAccountId) q.fromAccountId = String(r.fromAccountId)
  return { path: '/transactions/new', query: q }
}

function showFlash(msg: string) {
  flash.value = msg
  if (flashTimer) clearTimeout(flashTimer)
  flashTimer = setTimeout(() => { flash.value = '' }, 2200)
}

function canQuickCreate(t: Template) {
  if (!t.accountId) return false
  if (t.type === 'transfer' && !t.toAccountId) return false
  if ((Number(t.amount) || 0) <= 0 && !t.fareRuleId) return false
  return true
}

async function resolveAmount(t: Template): Promise<number> {
  let amount = Number(t.amount) || 0
  if (t.fareRuleId) {
    try {
      const { data } = await http.post(`/templates/${t.id}/preview-fare`, {})
      const fen = Number(data?.amountFen) || 0
      if (fen > 0) amount = fen
    } catch { /* 用模板原价 */ }
  }
  return amount
}

function askQuickUse(t: Template) {
  if (quickBusy.value) return
  pendingTpl.value = t
  confirmOpen.value = true
}

function cancelQuick() {
  if (quickBusy.value) return
  confirmOpen.value = false
  pendingTpl.value = null
}

async function editInstead() {
  const t = pendingTpl.value
  if (!t || quickBusy.value) return
  confirmOpen.value = false
  pendingTpl.value = null
  await router.push({ path: '/transactions/new', query: { templateId: String(t.id) } })
}

async function confirmQuick() {
  const t = pendingTpl.value
  if (!t || quickBusy.value) return
  await quickUse(t)
}

async function quickUse(t: Template) {
  if (quickBusy.value) return
  if (!canQuickCreate(t)) {
    confirmOpen.value = false
    pendingTpl.value = null
    await router.push({ path: '/transactions/new', query: { templateId: String(t.id) } })
    return
  }
  quickBusy.value = t.id
  try {
    const amount = await resolveAmount(t)
    if (amount <= 0 || !t.accountId) {
      confirmOpen.value = false
      pendingTpl.value = null
      await router.push({ path: '/transactions/new', query: { templateId: String(t.id) } })
      return
    }
    await tx.create({
      type: t.type,
      amount,
      accountId: t.accountId,
      toAccountId: t.toAccountId || null,
      categoryId: t.categoryId || null,
      tagIds: t.tagIds || [],
      remark: t.remark || '',
      fareRuleId: t.fareRuleId || null,
      geoMode: t.geoMode || undefined,
      geoLng: t.geoLng ?? null,
      geoLat: t.geoLat ?? null,
      geoName: t.geoName || '',
      geoEndLng: t.geoEndLng ?? null,
      geoEndLat: t.geoEndLat ?? null,
      geoEndName: t.geoEndName || '',
      happenedAt: dayjs().toISOString(),
    })
    confirmOpen.value = false
    pendingTpl.value = null
    showFlash(`已记「${t.name}」¥${fenToYuan(amount)}`)
    await Promise.all([
      tx.loadRecent(),
      http.get('/stats/summary').then((r) => { summary.value = r.data }),
    ])
  } catch (e: any) {
    showFlash(e?.message || '记账失败')
  } finally {
    quickBusy.value = null
  }
}

async function refreshHome() {
  const [{ data: tpls }] = await Promise.all([
    http.get('/templates').catch(() => ({ data: [] })),
    tx.loadRecent(),
    http.get('/stats/summary').then((r) => { summary.value = r.data }),
    accounts.creditRepayReminders().then((list) => { repayReminders.value = list }).catch(() => {}),
  ])
  templates.value = (tpls || []).slice(0, 9)
}

onMounted(refreshHome)
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
  position: relative;
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

.home-mid {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 8px 0 12px;
  overflow-x: hidden;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

.repay-tags {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
  max-width: min(420px, 100%);
  padding: 0 8px;
  flex: 0 0 auto;
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
  width: clamp(152px, 42vw, 184px);
  height: clamp(152px, 42vw, 184px);
  border-radius: 50%;
  display: grid;
  place-items: center;
  text-decoration: none;
  color: #fff;
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
  outline: none;
  flex: 0 0 auto;
}

.fab-ring {
  position: absolute;
  inset: -4px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(27, 127, 90, 0.22), transparent 70%);
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
  gap: 4px;
  background: linear-gradient(145deg, #24956b 0%, #1b7f5a 48%, #145a41 100%);
  box-shadow:
    0 16px 36px rgba(27, 127, 90, 0.38),
    0 2px 0 rgba(255, 255, 255, 0.18) inset;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.fab-record:active .fab-core,
.fab-record:hover .fab-core {
  transform: scale(0.96);
  box-shadow: 0 10px 24px rgba(27, 127, 90, 0.3);
}

.fab-label {
  font-family: var(--font-display);
  font-size: 1.12rem;
  font-weight: 700;
  letter-spacing: 0.1em;
}

.tpl-block {
  width: min(100%, 380px);
  padding: 0 12px;
  margin-top: 22px;
  flex: 0 0 auto;
}

.tpl-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 8px;
  padding: 0 2px;
}

.tpl-title {
  font-size: 0.86rem;
  font-weight: 700;
  color: var(--muted);
  letter-spacing: 0.04em;
}

.tpl-manage {
  font-size: 0.8rem;
  color: var(--primary);
  text-decoration: none;
  font-weight: 600;
}

.tpl-empty {
  font-size: 0.82rem;
  color: var(--muted);
  text-align: center;
  padding: 10px 8px;
  line-height: 1.5;
}

.tpl-empty a {
  color: var(--primary);
  font-weight: 600;
  text-decoration: none;
}

.tpl-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.tpl-btn {
  position: relative;
  min-height: 72px;
  padding: 10px 8px 8px;
  border: 1px solid color-mix(in srgb, var(--primary) 22%, transparent);
  border-radius: 16px;
  background: color-mix(in srgb, var(--surface-solid, #fff) 88%, var(--primary) 6%);
  box-shadow: var(--shadow-soft);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
  color: inherit;
  font: inherit;
  transition: transform 0.15s ease, border-color 0.15s ease;
}

.tpl-btn:active:not(:disabled) {
  transform: scale(0.96);
}

.tpl-btn:disabled {
  opacity: 0.7;
}

.tpl-ico {
  color: var(--primary);
  margin-bottom: 2px;
}

.tpl-name {
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.2;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tpl-amt {
  font-family: var(--font-display);
  font-size: 0.72rem;
  font-variant-numeric: tabular-nums;
  color: var(--muted);
}

.muted-amt { opacity: 0.75; }

.tpl-spin {
  position: absolute;
  top: 6px;
  right: 6px;
}

.tpl-hint {
  margin: 8px 2px 0;
  font-size: 0.72rem;
  color: var(--muted);
  text-align: center;
  line-height: 1.35;
}

.confirm-card {
  border-radius: 16px !important;
}

.confirm-title {
  font-weight: 700;
  letter-spacing: 0.02em;
}

.confirm-body {
  padding-top: 4px !important;
}

.confirm-msg {
  margin: 0 0 10px;
  line-height: 1.5;
  font-size: 0.95rem;
}

.confirm-meta {
  margin: 0;
  padding-left: 1.1em;
  color: var(--muted);
  font-size: 0.86rem;
  line-height: 1.55;
}

.confirm-actions {
  padding: 8px 12px 14px !important;
}

.home-bottom {
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  padding-bottom: 6px;
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

.recent {
  flex: 0 0 auto;
}

.home-flash {
  position: absolute;
  left: 50%;
  bottom: max(16px, var(--sab, 0px));
  transform: translateX(-50%);
  z-index: 20;
  max-width: min(90%, 320px);
  padding: 10px 16px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--primary) 92%, #000);
  color: #fff;
  font-size: 0.86rem;
  font-weight: 600;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.28);
  pointer-events: none;
  animation: flash-in 0.25s ease both;
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

@keyframes flash-in {
  from { opacity: 0; transform: translate(-50%, 8px); }
  to { opacity: 1; transform: translate(-50%, 0); }
}

@media (prefers-reduced-motion: reduce) {
  .anim-in,
  .row-in,
  .fab-ring,
  .home-flash { animation: none !important; }
}

[data-theme='dark'] .fab-core {
  background: linear-gradient(145deg, #4aaf84 0%, #3d9b74 50%, #2d7a5a 100%);
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.45), 0 2px 0 rgba(255, 255, 255, 0.1) inset;
}

[data-theme='dark'] .tpl-btn {
  background: color-mix(in srgb, var(--surface-solid, #1a1f1c) 90%, var(--primary) 8%);
}

@media (max-width: 360px) {
  .tpl-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
