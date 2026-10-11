<template>
  <div class="page tl-page">
    <h1 class="page-title">{{ t('txList.title') }}</h1>
    <div class="tl-layout">
    <aside class="tl-side">
    <div class="search-row">
      <v-text-field
        v-model="keyword"
        :label="t('txList.keyword')"
        density="compact"
        variant="outlined"
        hide-details
        clearable
        class="f"
        prepend-inner-icon="mdi-magnify"
        @keyup.enter="reload"
        @click:clear="reload"
      />
      <v-btn
        class="filter-toggle"
        :variant="filtersOpen || activeFilterCount ? 'tonal' : 'outlined'"
        color="primary"
        height="40"
        @click="filtersOpen = !filtersOpen"
      >
        <v-icon start size="18">mdi-filter-variant</v-icon>
        {{ t('txList.filter') }}<template v-if="activeFilterCount">·{{ activeFilterCount }}</template>
        <v-icon end size="18">{{ filtersOpen ? 'mdi-chevron-up' : 'mdi-chevron-down' }}</v-icon>
      </v-btn>
    </div>
    <div class="filter-panel" :class="{ open: filtersOpen }">
    <div class="filters">
      <v-select
        v-model="accountId"
        :items="[{ id: 0, name: t('txList.allAccounts') }, ...accounts.list]"
        item-title="name"
        item-value="id"
        density="compact"
        variant="outlined"
        hide-details
        class="f"
      />
      <v-select
        v-model="type"
        :items="typeItems"
        density="compact"
        variant="outlined"
        hide-details
        class="f"
      />
    </div>
    <div class="filters">
      <v-select
        v-model="categoryId"
        :items="[{ id: 0, name: t('txList.allCategories') }, ...catItems]"
        item-title="name"
        item-value="id"
        density="compact"
        variant="outlined"
        hide-details
        class="f"
      />
      <v-select
        v-model="tagId"
        :items="[{ id: 0, name: t('txList.allTags') }, ...tags.list]"
        item-title="name"
        item-value="id"
        density="compact"
        variant="outlined"
        hide-details
        class="f"
      />
    </div>
    <div class="filters">
      <NativeDateField v-model="from" :label="t('txList.from')" density="compact" variant="outlined" hide-details class="f" />
      <NativeDateField v-model="to" :label="t('txList.to')" density="compact" variant="outlined" hide-details class="f" />
    </div>
    <div class="filters">
      <v-select
        v-model="voucher"
        :items="voucherFilterItems"
        density="compact"
        variant="outlined"
        hide-details
        class="f"
      />
    </div>
    <div v-if="activeFilterCount" class="filter-reset">
      <v-btn size="small" variant="text" color="primary" @click="resetFilters">{{ t('txList.clearFilters') }}</v-btn>
    </div>
    </div>
    </aside>

    <div class="tl-main">

    <div v-if="pendingOnly" class="pending-filter">
      <v-icon size="16">mdi-timer-sand</v-icon>
      <span class="f">{{ t('txList.pendingOnly', { n: tx.total }) }}</span>
      <v-btn size="small" variant="text" color="primary" @click="pendingOnly = false">{{ t('txList.viewAll') }}</v-btn>
    </div>

    <div v-if="voucher" class="pending-filter voucher-filter">
      <v-icon size="16">mdi-ticket-percent-outline</v-icon>
      <span class="f">{{ t('txList.filterCount', { label: voucherFilterLabel, n: tx.total }) }}</span>
      <v-btn size="small" variant="text" color="primary" @click="openRefund(null)">{{ t('txEdit.refund') }}</v-btn>
      <v-btn size="small" variant="text" color="primary" @click="voucher = ''">{{ t('txList.viewAll') }}</v-btn>
    </div>

    <div v-if="!groups.length" class="empty-tip">{{ emptyText }}</div>
    <div v-for="g in groups" :key="g.day" class="group">
      <div class="day">
        <span>{{ g.day }}<span class="day-week">{{ weekday(g.day) }}</span></span>
        <span class="day-sum">
          <span v-if="g.income" class="amount-income">{{ t('txList.incomeShort', { amount: fenToYuan(g.income) }) }}</span>
          <span v-if="g.expense" class="amount-expense">{{ t('txList.expenseShort', { amount: fenToYuan(g.expense) }) }}</span>
        </span>
      </div>
      <v-list bg-color="transparent" class="py-0">
        <v-list-item
          v-for="item in g.items"
          :key="item.id"
          :to="item.type === 'revalue' ? `/accounts/${item.accountId}/ledger` : `/transactions/${item.id}`"
          rounded="lg"
          class="tx-row surface-row"
        >
          <v-list-item-title>
            {{ txTitle(item) }}
            <v-chip v-if="item.amountPending" size="x-small" color="warning" variant="tonal" class="ml-1">{{ t('txList.pendingChip') }}</v-chip>
            <v-chip
              v-if="item.voucherStatus"
              size="x-small"
              :color="voucherStatusColor[item.voucherStatus]"
              variant="tonal"
              class="ml-1"
            >{{ voucherStatusLabel(item.voucherStatus) }}</v-chip>
            <v-chip v-if="item.refundOfId" size="x-small" color="info" variant="tonal" class="ml-1">{{ t('txList.refundChip') }}</v-chip>
          </v-list-item-title>
          <v-list-item-subtitle>
            <template v-if="txSubCategory(item)">{{ txSubCategory(item) }} · </template>
            {{ item.account?.name }}
            <template v-if="item.type === 'transfer'"> → {{ item.toAccount?.name }}</template>
            <span v-if="item.tags?.length"> · {{ item.tags.map((tag) => tag.name).join(',') }}</span>
            <span
              v-if="item.voucherStatus === 'pending' && item.voucherExpireOn"
              :class="{ 'expire-soon': expireDays(item.voucherExpireOn) <= 3 }"
            > · {{ expireText(item.voucherExpireOn) }}</span>
          </v-list-item-subtitle>
          <div v-if="item.voucherStatus === 'pending'" class="voucher-actions">
            <v-btn
              size="x-small"
              variant="tonal"
              color="success"
              :loading="voucherBusy === item.id"
              @click.stop.prevent="markUsed(item)"
            >{{ t('txEdit.markUsed') }}</v-btn>
            <v-btn size="x-small" variant="tonal" color="info" @click.stop.prevent="openRefund(item)">{{ t('txEdit.refund') }}</v-btn>
          </div>
          <template #append>
            <div class="amt-col">
              <span :class="amountClass(item)">
                {{ formatMoney(item.amount, item.type) }}
              </span>
              <span v-if="item.foreignCurrency" class="fx-line">
                {{ formatForeign(item.foreignCurrency, item.foreignAmount) }} · {{ payRate(item.amount, item.foreignAmount) }}
              </span>
            </div>
          </template>
        </v-list-item>
      </v-list>
    </div>

    <div v-if="tx.total > tx.list.length" class="more">
      <v-btn variant="tonal" :loading="loadingMore" @click="loadMore">{{ t('txList.loadMore', { loaded: tx.list.length, total: tx.total }) }}</v-btn>
    </div>
    </div>
    </div>

    <VoucherRefundDialog v-model="refundOpen" :voucher="refundTarget" @refunded="onRefunded" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import { useAccountStore } from '../stores/account'
import { useCategoryStore } from '../stores/category'
import { useTagStore } from '../stores/tag'
import {
  useTransactionStore,
  voucherStatusColor,
  voucherStatusLabel,
  type Transaction,
} from '../stores/transaction'
import { amountClass, fenToYuan, formatForeign, formatMoney, payRate, txSubCategory, txTitle } from '../utils/money'
import NativeDateField from '../components/NativeDateField.vue'
import VoucherRefundDialog from '../components/VoucherRefundDialog.vue'

const { t } = useI18n()
const accounts = useAccountStore()
const categories = useCategoryStore()
const tags = useTagStore()
const tx = useTransactionStore()
const route = useRoute()
const router = useRouter()

const queryNum = (k: string) => Number(route.query[k]) || 0
const queryStr = (k: string) => (typeof route.query[k] === 'string' ? (route.query[k] as string) : '')

const pendingOnly = ref(route.query.pending === '1')
// 统计页下钻会带上账户 / 分类 / 标签 / 类型 / 日期
const accountId = ref(queryNum('accountId'))
const categoryId = ref(queryNum('categoryId'))
const tagId = ref(queryNum('tagId'))
const type = ref(queryStr('type'))
const keyword = ref('')
const from = ref(queryStr('from'))
const to = ref(queryStr('to'))
const voucher = ref(String(route.query.voucher || ''))
const pageSize = 30
const loadingMore = ref(false)
const filtersOpen = ref(
  Boolean(accountId.value || categoryId.value || tagId.value || type.value || from.value || to.value),
)
const voucherBusy = ref<number | null>(null)
const refundOpen = ref(false)
const refundTarget = ref<Transaction | null>(null)

const typeItems = computed(() => [
  { title: t('txList.allTypes'), value: '' },
  ...(['expense', 'income', 'transfer', 'revalue'] as const).map((v) => ({ title: t(`txType.${v}`), value: v })),
])

const VOUCHER_STATUSES = ['pending', 'used', 'refunded'] as const
const voucherFilterItems = computed(() => [
  { title: t('txList.voucherAny'), value: '' },
  { title: t('txList.voucherAll'), value: 'all' },
  ...VOUCHER_STATUSES.map((s) => ({ title: voucherStatusLabel(s), value: s })),
])
const voucherFilterLabel = computed(() => {
  if (voucher.value === 'all') return t('txList.voucherAll')
  const s = VOUCHER_STATUSES.find((x) => x === voucher.value)
  return t('txList.voucherStatus', { status: s ? voucherStatusLabel(s) : '' })
})
const emptyText = computed(() => {
  if (pendingOnly.value) return t('txList.emptyPending')
  if (voucher.value) return t('txList.emptyVoucher')
  return t('txList.empty')
})

const activeFilterCount = computed(() =>
  [accountId.value, type.value, categoryId.value, tagId.value, from.value, to.value, voucher.value].filter(Boolean).length,
)

function resetFilters() {
  accountId.value = 0
  type.value = ''
  categoryId.value = 0
  tagId.value = 0
  from.value = ''
  to.value = ''
  voucher.value = ''
}

function expireDays(d: string) {
  return dayjs(d).startOf('day').diff(dayjs().startOf('day'), 'day')
}

function expireText(d: string) {
  const days = expireDays(d)
  if (days < 0) return t('voucher.expiredDays', { n: -days })
  if (days === 0) return t('voucher.expireToday')
  return t('voucher.expireOn', { date: dayjs(d).format('M/D') })
}

async function markUsed(item: Transaction) {
  voucherBusy.value = item.id
  try {
    await tx.setVoucherStatus(item.id, 'used')
    if (voucher.value === 'pending') await reload()
  } catch (e: any) {
    alert(e?.message || t('txList.markUsedFailed'))
  } finally {
    voucherBusy.value = null
  }
}

function openRefund(item: Transaction | null) {
  refundTarget.value = item
  refundOpen.value = true
}

async function onRefunded() {
  await reload()
}

const catItems = computed(() => {
  const items = [...categories.expenseOptions, ...categories.incomeOptions]
  // 有子类的一级分类不在可选项里，但按它筛选（含子类）是合法的
  if (categoryId.value && !items.some((c) => c.id === categoryId.value)) {
    const c = categories.list.find((x) => x.id === categoryId.value)
    if (c) items.unshift({ id: c.id, name: t('txList.withChildren', { name: c.name }), icon: c.icon })
  }
  return items
})

const groups = computed(() => {
  const map = new Map<string, Transaction[]>()
  for (const item of tx.list) {
    const day = dayjs(item.happenedAt).format('YYYY-MM-DD')
    if (!map.has(day)) map.set(day, [])
    map.get(day)!.push(item)
  }
  return [...map.entries()].map(([day, items]) => ({
    day,
    items,
    expense: items.filter((x) => x.type === 'expense').reduce((s, x) => s + x.amount, 0),
    income: items.filter((x) => x.type === 'income').reduce((s, x) => s + x.amount, 0),
  }))
})

function weekday(day: string) {
  return dayjs(day).format('ddd')
}

function buildParams(offset = 0) {
  const params: Record<string, unknown> = { limit: pageSize, offset }
  if (accountId.value) params.accountId = accountId.value
  if (categoryId.value) params.categoryId = categoryId.value
  if (tagId.value) params.tagId = tagId.value
  if (type.value) params.type = type.value
  if (pendingOnly.value) params.pending = 1
  if (voucher.value) params.voucher = voucher.value
  if (keyword.value) params.keyword = keyword.value
  if (from.value) params.from = from.value
  if (to.value) {
    // 结束日含当天：传次日
    params.to = dayjs(to.value).add(1, 'day').format('YYYY-MM-DD')
  }
  return params
}

async function reload() {
  await tx.load(buildParams(0))
}

async function loadMore() {
  loadingMore.value = true
  try {
    const { data } = await (await import('../api/http')).default.get('/transactions', {
      params: buildParams(tx.list.length),
    })
    const items = Array.isArray(data) ? data : (data.items || [])
    tx.list.push(...items)
    if (!Array.isArray(data)) tx.total = data.total
  } finally {
    loadingMore.value = false
  }
}

onMounted(async () => {
  await Promise.all([accounts.ensure(), categories.ensure(), tags.ensure()])
  await reload()
})

watch([accountId, type, categoryId, tagId, from, to, pendingOnly, voucher], reload)
watch(pendingOnly, (v) => {
  if (!v && route.query.pending) router.replace({ query: { ...route.query, pending: undefined } })
})
watch(voucher, (v) => {
  if (!v && route.query.voucher) router.replace({ query: { ...route.query, voucher: undefined } })
})
</script>

<style scoped>
.search-row { display: flex; gap: 8px; align-items: center; margin-bottom: 8px; }
.filter-toggle { flex-shrink: 0; padding: 0 10px !important; letter-spacing: 0; }
.filters { display: flex; gap: 8px; margin-bottom: 8px; }
.f { flex: 1; min-width: 0; }
.filter-reset { display: flex; justify-content: flex-end; margin: -4px 0 4px; }
.day { color: var(--muted); font-size: 0.85rem; margin: 12px 4px 4px; font-weight: 600; }
.tx-row { margin-bottom: 6px; min-height: 52px !important; }
.tx-row :deep(.v-list-item-subtitle) { font-size: 0.76rem; }
.more { text-align: center; padding: 16px 0 32px; }
.amt-col { display: flex; flex-direction: column; align-items: flex-end; gap: 1px; }
.fx-line { font-size: 0.7rem; color: var(--muted); font-variant-numeric: tabular-nums; white-space: nowrap; }
.pending-filter {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 4px 0 8px;
  padding: 6px 4px 6px 10px;
  border-radius: 10px;
  font-size: 0.78rem;
  color: #8a5a10;
  background: color-mix(in srgb, #d99a2b 14%, transparent);
}
[data-theme='dark'] .pending-filter { color: #f0c36a; }
.voucher-filter { color: #2b6aa3; background: color-mix(in srgb, #3b82c4 12%, transparent); }
[data-theme='dark'] .voucher-filter { color: #9cc7ef; }
.voucher-actions { display: flex; gap: 6px; margin-top: 4px; }
.expire-soon { color: #c0562b; font-weight: 600; }

.day { display: flex; justify-content: space-between; align-items: baseline; gap: 8px; }
.day-week { margin-left: 6px; font-weight: 400; }
.day-sum { display: inline-flex; gap: 10px; font-size: 0.78rem; }
.day-sum > span { font-weight: 500; }

/* 宽屏始终展开，窄屏默认折叠 */
@media (min-width: 600px) {
  .filter-toggle { display: none; }
}
@media (max-width: 599px) {
  .filter-panel { display: none; }
  .filter-panel.open { display: block; }
  .filters :deep(.v-field__input) { padding-inline: 10px; }
  .day-week,
  .day-sum { display: none; }
}

/* 平板：筛选项拍平成三列网格，少占几行 */
@media (min-width: 600px) and (max-width: 959px) {
  .filter-panel {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 8px;
    margin-bottom: 8px;
  }
  .filters { display: contents; }
  .filter-reset { grid-column: 1 / -1; margin: 0; }
}

/* 宽屏：左侧筛选栏吸顶，右侧流水按天成卡 */
@media (min-width: 960px) {
  .tl-layout {
    display: grid;
    grid-template-columns: 260px minmax(0, 1fr);
    gap: 20px;
    align-items: start;
  }
  .tl-side {
    position: sticky;
    top: 16px;
    padding: 14px;
    border-radius: var(--radius-md);
    background: var(--surface);
    border: 1px solid var(--surface-border);
    box-shadow: var(--shadow-soft);
  }
  .tl-side .search-row { margin-bottom: 12px; }
  .filters { flex-direction: column; gap: 10px; margin-bottom: 10px; }
  .filter-reset { justify-content: stretch; margin: 2px 0 0; }
  .filter-reset .v-btn { width: 100%; }
  .group {
    padding: 4px 8px 6px;
    margin-bottom: 12px;
    border-radius: var(--radius-md);
    background: color-mix(in srgb, var(--surface) 55%, transparent);
    border: 1px solid var(--surface-border);
  }
  .day { margin: 8px 8px 6px; }
  .tx-row { min-height: 58px !important; }
  .tl-main .pending-filter { margin-top: 0; }
}
</style>
