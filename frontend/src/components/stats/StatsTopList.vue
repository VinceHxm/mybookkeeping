<template>
  <section class="top surface-card">
    <header class="top-head">
      <span class="top-title">{{ t(`stats.top.title.${kind}`, { n: limit }) }}</span>
      <span class="top-sub">{{ t('stats.top.sub') }}</span>
    </header>
    <div v-if="!list.length" class="top-empty">{{ loading ? t('common.loading') : t('stats.top.empty') }}</div>
    <ol v-else class="top-list" :style="{ '--rows': Math.ceil(list.length / 2) }">
      <li v-for="(tx, i) in list" :key="tx.id">
        <router-link :to="`/transactions/${tx.id}`" class="top-row">
          <span class="rank" :class="{ 'rank--hot': i < 3 }">{{ i + 1 }}</span>
          <div class="top-main">
            <div class="top-name">{{ txTitle(tx) }}</div>
            <div class="top-meta">
              {{ dayjs(tx.happenedAt).format(t('stats.fmt.monthDay')) }}
              <template v-if="tx.category?.name"> · {{ tx.category.name }}</template>
              <template v-if="tx.account?.name"> · {{ tx.account.name }}</template>
            </div>
          </div>
          <span :class="kind === 'income' ? 'amount-income' : 'amount-expense'" class="top-amt">{{ yuan(tx.amount) }}</span>
        </router-link>
      </li>
    </ol>
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import http from '../../api/http'
import type { Transaction } from '../../stores/transaction'
import { txTitle } from '../../utils/money'
import { statsParams, yuan, type StatsFilter } from '../../utils/stats'

const props = withDefaults(defineProps<{ filter: StatsFilter; kind: 'expense' | 'income'; limit?: number }>(), { limit: 10 })
const { t } = useI18n()

const list = ref<Transaction[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const { data } = await http.get('/stats/top', {
      params: statsParams(props.filter, { kind: props.kind, limit: props.limit }),
    })
    list.value = data || []
  } finally {
    loading.value = false
  }
}

watch(() => [props.filter.from, props.filter.to, props.filter.accountId, props.kind], load, { immediate: true })
</script>

<style scoped>
.top { padding: 12px 6px 6px; }
.top-head { display: flex; align-items: baseline; gap: 8px; padding: 0 10px 6px; }
.top-title { font-weight: 700; font-size: 0.92rem; }
.top-sub { font-size: 0.74rem; color: var(--muted); }
.top-empty { color: var(--muted); text-align: center; padding: 20px 0; font-size: 0.85rem; }
.top-list { list-style: none; margin: 0; padding: 0; }
.top-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 10px;
  border-radius: 10px;
  color: inherit;
  text-decoration: none;
  min-width: 0;
}
.top-row:hover { background: var(--primary-soft); }
.rank {
  width: 22px;
  height: 22px;
  border-radius: 7px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.74rem;
  font-weight: 700;
  color: var(--muted);
  background: rgba(var(--v-theme-on-surface), 0.06);
  flex-shrink: 0;
}
.rank--hot { color: #fff; background: var(--primary); }
.top-main { flex: 1; min-width: 0; }
.top-name { font-size: 0.88rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.top-meta { font-size: 0.72rem; color: var(--muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.top-amt { font-size: 0.9rem; flex-shrink: 0; }

/* 宽屏两列排布，减少纵向滚动 */
@media (min-width: 960px) {
  .top-list {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    grid-auto-flow: column;
    grid-template-rows: repeat(var(--rows), auto);
    column-gap: 12px;
  }
}
</style>
