<template>
  <div class="form-fields">
    <v-text-field v-model="form.name" label="名称" variant="outlined" hide-details class="field" />
    <v-select
      v-model="form.type"
      :items="[
        { title: '支出', value: 'expense' },
        { title: '收入', value: 'income' },
        { title: '转账', value: 'transfer' },
      ]"
      label="类型"
      variant="outlined"
      hide-details
      class="field"
    />
    <v-text-field v-model="amountYuan" label="金额（元）" type="number" inputmode="decimal" variant="outlined" hide-details class="field" />
    <v-select
      v-model="form.accountId"
      :items="accounts.list"
      item-title="name"
      item-value="id"
      label="账户"
      variant="outlined"
      hide-details
      class="field"
    />
    <v-select
      v-if="form.type === 'transfer'"
      v-model="form.toAccountId"
      :items="accounts.list"
      item-title="name"
      item-value="id"
      label="转入账户"
      variant="outlined"
      hide-details
      class="field"
    />
    <v-select
      v-else
      v-model="form.categoryId"
      :items="form.type === 'expense' ? categories.expenseOptions : categories.incomeOptions"
      item-title="name"
      item-value="id"
      label="分类"
      variant="outlined"
      hide-details
      class="field"
    />
    <v-select v-model="form.frequency" :items="freqItems" label="频率" variant="outlined" hide-details class="field" />
    <v-text-field
      v-if="form.frequency === 'every_n_days'"
      v-model.number="form.intervalN"
      label="每隔几天"
      type="number"
      variant="outlined"
      hide-details
      class="field"
    />
    <div class="field">
      <NativeDateField v-model="nextLocal" label="下次执行时间" type="datetime-local" variant="outlined" hide-details />
    </div>
    <v-text-field v-model="form.remark" label="备注" variant="outlined" hide-details class="field" />
    <v-switch v-if="scheduleId" v-model="formEnabled" label="启用" color="primary" hide-details class="field" />

    <div v-if="scheduleId" class="danger-zone">
      <div class="danger-title">危险操作</div>
      <p class="danger-tip">删除后不会再自动生成流水，已生成的记录不受影响。</p>
      <v-btn color="error" variant="outlined" block @click="emit('delete')">删除此周期</v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import dayjs from 'dayjs'
import http from '../api/http'
import NativeDateField from './NativeDateField.vue'
import { useAccountStore } from '../stores/account'
import { useCategoryStore } from '../stores/category'
import { fenToYuan, yuanToFen } from '../utils/money'

const props = defineProps<{
  scheduleId?: number | null
  initial?: any | null
}>()

const emit = defineEmits<{ saved: []; delete: [] }>()

const accounts = useAccountStore()
const categories = useCategoryStore()
const saving = ref(false)
const amountYuan = ref('')
const nextLocal = ref(dayjs().format('YYYY-MM-DDTHH:mm'))
const formEnabled = ref(true)
const freqItems = [
  { title: '每天', value: 'daily' },
  { title: '每周', value: 'weekly' },
  { title: '每月', value: 'monthly' },
  { title: '每年', value: 'yearly' },
  { title: '每隔 N 天', value: 'every_n_days' },
]
const form = reactive({
  name: '',
  type: 'expense',
  accountId: undefined as number | undefined,
  toAccountId: undefined as number | undefined,
  categoryId: undefined as number | undefined,
  frequency: 'monthly',
  intervalN: 1,
  remark: '',
})

function fillFrom(s: any | null | undefined) {
  if (!s) {
    form.name = ''
    form.type = 'expense'
    form.accountId = accounts.list[0]?.id
    form.toAccountId = undefined
    form.categoryId = undefined
    form.frequency = 'monthly'
    form.intervalN = 1
    form.remark = ''
    formEnabled.value = true
    amountYuan.value = ''
    nextLocal.value = dayjs().format('YYYY-MM-DDTHH:mm')
    return
  }
  form.name = s.name
  form.type = s.type
  form.accountId = s.accountId
  form.toAccountId = s.toAccountId
  form.categoryId = s.categoryId
  form.frequency = s.frequency
  form.intervalN = s.intervalN
  form.remark = s.remark || ''
  formEnabled.value = !!s.enabled
  amountYuan.value = fenToYuan(s.amount)
  nextLocal.value = dayjs(s.nextRunAt).format('YYYY-MM-DDTHH:mm')
}

async function load() {
  await Promise.all([accounts.load(), categories.load()])
  if (props.initial) fillFrom(props.initial)
  else if (props.scheduleId) {
    const { data } = await http.get('/schedules')
    const s = (data as any[]).find((x) => x.id === props.scheduleId)
    fillFrom(s || null)
  } else fillFrom(null)
}

watch(() => [props.scheduleId, props.initial], load)
onMounted(load)

async function save() {
  if (saving.value) return
  saving.value = true
  try {
    const payload = {
      ...form,
      enabled: props.scheduleId ? formEnabled.value : true,
      amount: yuanToFen(amountYuan.value),
      nextRunAt: dayjs(nextLocal.value).toISOString(),
    }
    if (props.scheduleId) await http.put(`/schedules/${props.scheduleId}`, payload)
    else await http.post('/schedules', payload)
    emit('saved')
  } finally {
    saving.value = false
  }
}

defineExpose({ save, saving })
</script>

<style scoped>
.form-fields .field { margin-bottom: 14px; }
.danger-zone {
  margin-top: 16px;
  padding: 16px;
  border-radius: 14px;
  border: 1px solid rgba(198, 40, 40, 0.35);
  background: rgba(198, 40, 40, 0.08);
}
.danger-title { font-weight: 700; color: #c62828; margin-bottom: 6px; }
.danger-tip { margin: 0 0 12px; font-size: 0.8rem; color: var(--muted); line-height: 1.4; }
</style>
