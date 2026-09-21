<template>
  <div class="page recognize-page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" @click="$router.back()" />
      <h1>文字识别记账</h1>
      <span style="width:40px" />
    </div>

    <p class="hint">可贴多段描述，一次识别多笔；确认后直接提交入库。</p>

    <v-textarea
      v-model="text"
      label="记账描述"
      variant="outlined"
      rows="6"
      auto-grow
      :disabled="recognizing || submitting"
      placeholder="例如：&#10;昨天午饭外卖 36.5 支付宝&#10;地铁 4 元&#10;工资入账 12000 招行&#10;还信用卡 2000 从招行到中信信用卡"
      class="mb-3"
    />

    <v-btn
      color="primary"
      block
      size="large"
      class="mb-4"
      :loading="recognizing"
      :disabled="!text.trim() || submitting"
      @click="onRecognize"
    >
      识别多笔
    </v-btn>

    <v-alert v-if="error" type="error" density="compact" class="mb-3">{{ error }}</v-alert>

    <template v-if="items.length">
      <div class="list-head mb-2">
        <span>识别结果（{{ selectedCount }}/{{ items.length }}）</span>
        <div>
          <v-btn size="small" variant="text" @click="toggleAll(true)">全选</v-btn>
          <v-btn size="small" variant="text" @click="toggleAll(false)">清空</v-btn>
        </div>
      </div>

      <div
        v-for="(row, idx) in items"
        :key="row.key"
        class="item-card surface-card mb-3"
        :class="{ dim: !row.selected }"
      >
        <div class="item-top">
          <v-checkbox v-model="row.selected" hide-details density="compact" color="primary" />
          <v-btn-toggle v-model="row.type" mandatory density="compact" color="primary" divided>
            <v-btn value="expense" size="small">支</v-btn>
            <v-btn value="income" size="small">收</v-btn>
            <v-btn value="transfer" size="small">转</v-btn>
          </v-btn-toggle>
          <v-btn icon="mdi-delete-outline" size="small" variant="text" @click="removeItem(idx)" />
        </div>

        <v-text-field
          v-model="row.amountYuan"
          label="金额（元）"
          type="number"
          inputmode="decimal"
          variant="outlined"
          density="compact"
          prefix="¥"
          class="mb-1"
          hide-details
        />
        <v-select
          v-model="row.accountId"
          :items="accounts.list"
          item-title="name"
          item-value="id"
          :label="row.type === 'transfer' ? '转出账户' : '账户'"
          variant="outlined"
          density="compact"
          class="mb-1"
          hide-details
        />
        <v-select
          v-if="row.type === 'transfer'"
          v-model="row.toAccountId"
          :items="accounts.list.filter((a) => a.id !== row.accountId)"
          item-title="name"
          item-value="id"
          label="转入账户"
          variant="outlined"
          density="compact"
          class="mb-1"
          hide-details
        />
        <v-select
          v-else
          v-model="row.categoryId"
          :items="row.type === 'income' ? categories.incomeOptions : categories.expenseOptions"
          item-title="name"
          item-value="id"
          label="分类"
          variant="outlined"
          density="compact"
          class="mb-1"
          hide-details
        />
        <v-text-field
          v-model="row.remark"
          label="备注"
          variant="outlined"
          density="compact"
          class="mb-1"
          hide-details
        />
        <NativeDateField
          v-model="row.happenedLocal"
          label="时间"
          type="datetime-local"
          variant="outlined"
          density="compact"
          hide-details
        />
        <v-alert
          v-if="row.type === 'transfer' && isCreditAccount(row.toAccountId)"
          type="info"
          variant="tonal"
          density="compact"
          class="mt-2"
        >
          识别为信用卡还款（转账），不计入本月支出
        </v-alert>
      </div>

      <v-btn
        color="primary"
        block
        size="large"
        class="hero-cta mb-6"
        :loading="submitting"
        :disabled="selectedCount === 0"
        @click="onSubmit"
      >
        确认提交 {{ selectedCount }} 笔
      </v-btn>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import dayjs from 'dayjs'
import http from '../api/http'
import { useAccountStore } from '../stores/account'
import { useCategoryStore } from '../stores/category'
import { useAuthStore } from '../stores/auth'
import { useTransactionStore } from '../stores/transaction'
import { yuanToFen } from '../utils/money'
import NativeDateField from '../components/NativeDateField.vue'

type Row = {
  key: string
  selected: boolean
  type: 'expense' | 'income' | 'transfer'
  amountYuan: string
  accountId?: number
  toAccountId?: number
  categoryId?: number
  remark: string
  tagIds: number[]
  happenedLocal: string
}

const router = useRouter()
const accounts = useAccountStore()
const categories = useCategoryStore()
const auth = useAuthStore()
const tx = useTransactionStore()

const text = ref('')
const items = ref<Row[]>([])
const recognizing = ref(false)
const submitting = ref(false)
const error = ref('')
let keySeq = 0

const selectedCount = computed(() => items.value.filter((r) => r.selected).length)

onMounted(async () => {
  await Promise.all([
    accounts.load(),
    categories.load(),
    auth.fetchMe().catch(() => {}),
  ])
})

function isCreditAccount(id?: number) {
  if (!id) return false
  return accounts.list.some((a) => a.id === id && a.type === 'credit')
}

function defaultAccountId() {
  return auth.profile?.defaultAccountId || accounts.list.find((a) => a.type !== 'credit')?.id || accounts.list[0]?.id
}

function toggleAll(v: boolean) {
  items.value.forEach((r) => { r.selected = v })
}

function removeItem(idx: number) {
  items.value.splice(idx, 1)
}

async function onRecognize() {
  error.value = ''
  recognizing.value = true
  try {
    const { data } = await http.post('/ai/recognize-text-batch', { text: text.value }, { timeout: 120000 })
    const list = (data?.items || []) as any[]
    if (!list.length) {
      error.value = '未识别到可记账条目，请补充金额与账户等信息'
      items.value = []
      return
    }
    const defAcc = defaultAccountId()
    items.value = list.map((it) => {
      keySeq += 1
      const type = (['expense', 'income', 'transfer'].includes(it.type) ? it.type : 'expense') as Row['type']
      return {
        key: `r-${keySeq}`,
        selected: true,
        type,
        amountYuan: it.amountYuan != null ? String(it.amountYuan) : (it.amountFen != null ? String(it.amountFen / 100) : ''),
        accountId: it.accountId || defAcc,
        toAccountId: it.toAccountId || undefined,
        categoryId: it.categoryId || undefined,
        remark: it.remark || '',
        tagIds: it.tagIds || [],
        happenedLocal: it.happenedAt
          ? dayjs(it.happenedAt).format('YYYY-MM-DDTHH:mm')
          : dayjs().format('YYYY-MM-DDTHH:mm'),
      }
    })
  } catch (e: any) {
    error.value = e.message || '识别失败'
  } finally {
    recognizing.value = false
  }
}

async function onSubmit() {
  error.value = ''
  const rows = items.value.filter((r) => r.selected)
  if (!rows.length) return

  for (const r of rows) {
    const amount = yuanToFen(r.amountYuan)
    if (amount <= 0) {
      error.value = '请检查金额：每笔须大于 0'
      return
    }
    if (!r.accountId) {
      error.value = '请为每笔选择账户'
      return
    }
    if (r.type === 'transfer' && !r.toAccountId) {
      error.value = '转账请选择转入账户'
      return
    }
    if (r.type !== 'transfer' && !r.categoryId) {
      error.value = '支出/收入请选择分类'
      return
    }
  }

  submitting.value = true
  let ok = 0
  try {
    for (const r of rows) {
      await tx.create({
        type: r.type,
        amount: yuanToFen(r.amountYuan),
        accountId: r.accountId!,
        toAccountId: r.type === 'transfer' ? r.toAccountId : null,
        categoryId: r.type !== 'transfer' ? r.categoryId : null,
        remark: r.remark,
        tagIds: r.tagIds,
        happenedAt: dayjs(r.happenedLocal).toISOString(),
      })
      ok += 1
    }
    await router.replace('/')
  } catch (e: any) {
    error.value = ok > 0
      ? `已成功 ${ok} 笔，后续失败：${e.message || '请重试'}`
      : (e.message || '提交失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.toolbar h1 { font-size: 1.15rem; margin: 0; }
.hint { color: var(--muted); font-size: 0.85rem; margin: 0 0 12px; line-height: 1.45; }
.list-head {
  display: flex; align-items: center; justify-content: space-between;
  font-weight: 600; font-size: 0.92rem;
}
.item-card { padding: 10px 12px; }
.item-top {
  display: flex; align-items: center; gap: 8px; margin-bottom: 8px;
}
.dim { opacity: 0.55; }
</style>
