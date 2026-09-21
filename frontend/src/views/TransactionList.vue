<template>
  <div class="page">
    <h1 class="page-title">明细</h1>
    <div class="filters">
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
      <v-select
        v-model="type"
        :items="[
          { title: '全部类型', value: '' },
          { title: '支出', value: 'expense' },
          { title: '收入', value: 'income' },
          { title: '转账', value: 'transfer' },
        ]"
        density="compact"
        variant="outlined"
        hide-details
        class="f"
      />
    </div>
    <div class="filters">
      <v-select
        v-model="categoryId"
        :items="[{ id: 0, name: '全部分类' }, ...catItems]"
        item-title="name"
        item-value="id"
        density="compact"
        variant="outlined"
        hide-details
        class="f"
      />
      <v-select
        v-model="tagId"
        :items="[{ id: 0, name: '全部标签' }, ...tags.list]"
        item-title="name"
        item-value="id"
        density="compact"
        variant="outlined"
        hide-details
        class="f"
      />
    </div>
    <div class="filters">
      <NativeDateField v-model="from" label="从" density="compact" variant="outlined" hide-details class="f" />
      <NativeDateField v-model="to" label="到" density="compact" variant="outlined" hide-details class="f" />
    </div>
    <v-text-field
      v-model="keyword"
      label="备注关键词"
      density="compact"
      variant="outlined"
      hide-details
      clearable
      class="mb-3"
      prepend-inner-icon="mdi-magnify"
      @keyup.enter="reload"
    />

    <div v-if="!groups.length" class="empty-tip">暂无流水</div>
    <div v-for="g in groups" :key="g.day" class="group">
      <div class="day">{{ g.day }}</div>
      <v-list bg-color="transparent">
        <v-list-item
          v-for="item in g.items"
          :key="item.id"
          :to="`/transactions/${item.id}`"
          rounded="lg"
          class="mb-2 surface-row"
        >
          <v-list-item-title>{{ item.category?.name || typeLabel(item.type) }}</v-list-item-title>
          <v-list-item-subtitle>
            {{ item.account?.name }}
            <template v-if="item.type === 'transfer'"> → {{ item.toAccount?.name }}</template>
            <span v-if="item.remark"> · {{ item.remark }}</span>
            <span v-if="item.tags?.length"> · {{ item.tags.map((t) => t.name).join(',') }}</span>
          </v-list-item-subtitle>
          <template #append>
            <span :class="item.type === 'income' ? 'amount-income' : 'amount-expense'">
              {{ formatMoney(item.amount, item.type) }}
            </span>
          </template>
        </v-list-item>
      </v-list>
    </div>

    <div v-if="tx.total > tx.list.length" class="more">
      <v-btn variant="tonal" :loading="loadingMore" @click="loadMore">加载更多（{{ tx.list.length }}/{{ tx.total }}）</v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { useAccountStore } from '../stores/account'
import { useCategoryStore } from '../stores/category'
import { useTagStore } from '../stores/tag'
import { useTransactionStore, type Transaction } from '../stores/transaction'
import { formatMoney, typeLabel } from '../utils/money'
import NativeDateField from '../components/NativeDateField.vue'

const accounts = useAccountStore()
const categories = useCategoryStore()
const tags = useTagStore()
const tx = useTransactionStore()

const accountId = ref(0)
const categoryId = ref(0)
const tagId = ref(0)
const type = ref('')
const keyword = ref('')
const from = ref('')
const to = ref('')
const pageSize = 30
const loadingMore = ref(false)

const catItems = computed(() => [
  ...categories.expenseOptions,
  ...categories.incomeOptions,
])

const groups = computed(() => {
  const map = new Map<string, Transaction[]>()
  for (const item of tx.list) {
    const day = dayjs(item.happenedAt).format('YYYY-MM-DD')
    if (!map.has(day)) map.set(day, [])
    map.get(day)!.push(item)
  }
  return [...map.entries()].map(([day, items]) => ({ day, items }))
})

function buildParams(offset = 0) {
  const params: Record<string, unknown> = { limit: pageSize, offset }
  if (accountId.value) params.accountId = accountId.value
  if (categoryId.value) params.categoryId = categoryId.value
  if (tagId.value) params.tagId = tagId.value
  if (type.value) params.type = type.value
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
  await Promise.all([accounts.load(), categories.load(), tags.load()])
  await reload()
})

watch([accountId, type, categoryId, tagId, from, to], reload)
</script>

<style scoped>
.filters { display: flex; gap: 8px; margin-bottom: 8px; }
.f { flex: 1; }
.day { color: var(--muted); font-size: 0.85rem; margin: 12px 4px 4px; font-weight: 600; }
.more { text-align: center; padding: 16px 0 32px; }
@media (max-width: 600px) {
  .filters { flex-direction: column; }
}
</style>
