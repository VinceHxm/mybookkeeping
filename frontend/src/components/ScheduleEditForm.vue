<template>
  <div class="form-sections">
    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">基本信息</h3>
      </header>
      <div class="form-section__body">
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
        <v-text-field
          v-model="amountYuan"
          :label="form.type !== 'transfer' && amountPending ? '预估人民币金额（元）' : '金额（元）'"
          type="number"
          inputmode="decimal"
          variant="outlined"
          hide-details
          class="field field--last"
        />
      </div>
    </section>

    <section v-if="form.type !== 'transfer'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">外币订阅</h3>
      </header>
      <div class="form-section__body">
        <div class="form-section__group">
          <MetaHead
            icon="mdi-timer-sand"
            title="金额待确认"
            subtitle="生成的流水先按上方预估金额入账并标记「待确认」，等账单出来后改成实际金额"
          >
            <v-switch v-model="amountPending" color="primary" density="compact" hide-details inset />
          </MetaHead>
        </div>
        <div class="form-section__group">
          <ForeignPaymentFields
            v-model:enabled="fxEnabled"
            v-model:currency="fxCurrency"
            v-model:amount-yuan="fxAmountYuan"
            :cny-fen="yuanToFen(amountYuan || '0')"
            :estimated="amountPending"
            subtitle="每期外币价格固定时填写，生成的流水会带上外币金额"
          />
        </div>
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ form.type === 'transfer' ? '账户' : '账户与分类' }}</h3>
      </header>
      <div class="form-section__body">
        <v-select
          v-model="form.accountId"
          :items="accounts.list"
          item-title="name"
          item-value="id"
          :label="form.type === 'transfer' ? '转出账户' : '账户'"
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
          class="field field--last"
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
          class="field field--last"
        />
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">周期</h3>
      </header>
      <div class="form-section__body">
        <div class="field">
          <div class="mode-label">重复方式</div>
          <v-btn-toggle v-model="repeatMode" mandatory density="comfortable" color="primary" variant="outlined" divided class="mode-toggle">
            <v-btn value="calendar">按日历</v-btn>
            <v-btn value="fixed">固定天数</v-btn>
          </v-btn-toggle>
        </div>

        <template v-if="repeatMode === 'calendar'">
          <div class="interval-row field">
            <v-text-field
              v-model.number="form.intervalN"
              label="每"
              type="number"
              min="1"
              variant="outlined"
              hide-details
              class="interval-input"
            />
            <v-select
              v-model="form.frequency"
              :items="unitItems"
              variant="outlined"
              hide-details
              class="unit-select"
            />
          </div>
        </template>
        <template v-else>
          <div class="field">
            <v-text-field
              v-model.number="form.intervalN"
              label="间隔天数"
              type="number"
              min="1"
              variant="outlined"
              hide-details
              suffix="天"
            />
            <div class="quick-days">
              <v-chip
                v-for="d in quickDays"
                :key="d"
                size="small"
                :color="form.intervalN === d ? 'primary' : undefined"
                :variant="form.intervalN === d ? 'flat' : 'outlined'"
                @click="form.intervalN = d"
              >{{ d }} 天</v-chip>
            </div>
          </div>
        </template>

        <div class="field field--last">
          <NativeDateField
            v-model="nextLocal"
            :label="scheduleId ? '下次执行时间' : '首次执行时间'"
            type="datetime-local"
            variant="outlined"
            hide-details
          />
          <div class="rule-hint">{{ ruleHint }}</div>
          <div v-if="previewDates.length" class="rule-hint">接下来：{{ previewDates.join('、') }}</div>
        </div>
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">其他</h3>
      </header>
      <div class="form-section__body">
        <v-text-field
          v-model="form.remark"
          label="备注"
          variant="outlined"
          hide-details
          :class="scheduleId ? 'field' : 'field field--last'"
        />
        <v-switch
          v-if="scheduleId"
          v-model="formEnabled"
          label="启用"
          color="primary"
          hide-details
          class="field field--last"
        />
      </div>
    </section>

    <v-alert v-if="error" type="error" density="compact">{{ error }}</v-alert>

    <div v-if="scheduleId" class="danger-zone">
      <div class="danger-title">危险操作</div>
      <p class="danger-tip">删除后不会再自动生成流水，已生成的记录不受影响。</p>
      <v-btn color="error" variant="outlined" block @click="emit('delete')">删除此周期</v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import dayjs from 'dayjs'
import http from '../api/http'
import NativeDateField from './NativeDateField.vue'
import ForeignPaymentFields from './ForeignPaymentFields.vue'
import MetaHead from './MetaHead.vue'
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
const error = ref('')
const amountYuan = ref('')
const nextLocal = ref(dayjs().format('YYYY-MM-DDTHH:mm'))
const formEnabled = ref(true)
const amountPending = ref(false)
const fxEnabled = ref(false)
const fxCurrency = ref('USD')
const fxAmountYuan = ref('')
const unitItems = [
  { title: '天', value: 'daily' },
  { title: '周', value: 'weekly' },
  { title: '月', value: 'monthly' },
  { title: '年', value: 'yearly' },
]
const quickDays = [7, 14, 28, 30, 90]
const WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六']
/** 编辑时载入的下次执行时间与执行日；未改时间则沿用原执行日（避免被月末截断后的日期覆盖） */
const loadedNext = ref('')
const loadedAnchor = ref(0)
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

const repeatMode = computed({
  get: () => (form.frequency === 'every_n_days' ? 'fixed' : 'calendar'),
  set: (v: 'fixed' | 'calendar') => {
    if (v === repeatMode.value) return
    if (v === 'fixed') {
      form.frequency = 'every_n_days'
      form.intervalN = 30
    } else {
      form.frequency = 'monthly'
      form.intervalN = 1
    }
  },
})

const intervalN = computed(() => Math.max(1, Math.floor(Number(form.intervalN) || 1)))

const anchorDay = computed(() => {
  if (loadedAnchor.value && nextLocal.value === loadedNext.value) return loadedAnchor.value
  return dayjs(nextLocal.value).date()
})

function addMonthsAnchored(from: dayjs.Dayjs, months: number, anchor: number) {
  const first = from.startOf('month').add(months, 'month')
  return first.date(Math.min(anchor, first.daysInMonth()))
}

function nextRun(from: dayjs.Dayjs) {
  const n = intervalN.value
  switch (form.frequency) {
    case 'daily': return from.add(n, 'day')
    case 'weekly': return from.add(7 * n, 'day')
    case 'monthly': return addMonthsAnchored(from, n, anchorDay.value)
    case 'yearly': return addMonthsAnchored(from, 12 * n, anchorDay.value)
    default: return from.add(n, 'day')
  }
}

const ruleHint = computed(() => {
  const at = dayjs(nextLocal.value)
  if (!at.isValid()) return ''
  const n = intervalN.value
  const hm = at.format('HH:mm')
  switch (form.frequency) {
    case 'daily': return `${n > 1 ? `每 ${n} 天` : '每天'} ${hm} 执行`
    case 'weekly': return `${n > 1 ? `每 ${n} 周` : '每周'}的周${WEEKDAYS[at.day()]} ${hm} 执行`
    case 'monthly': {
      const d = anchorDay.value
      return `${n > 1 ? `每 ${n} 个月` : '每月'} ${d} 日执行${d > 28 ? '，当月没有这天时按月末执行' : ''}`
    }
    case 'yearly': return `${n > 1 ? `每 ${n} 年` : '每年'}的 ${at.month() + 1} 月 ${anchorDay.value} 日执行`
    default: return `从首次执行起每隔 ${n} 天执行一次，不受大小月影响`
  }
})

const previewDates = computed(() => {
  let at = dayjs(nextLocal.value)
  if (!at.isValid()) return []
  const thisYear = dayjs().year()
  const out: string[] = []
  for (let i = 0; i < 4; i++) {
    out.push(at.format(at.year() === thisYear ? 'M月D日' : 'YYYY年M月D日'))
    at = nextRun(at)
  }
  return out
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
    amountPending.value = false
    fxEnabled.value = false
    fxCurrency.value = 'USD'
    fxAmountYuan.value = ''
    nextLocal.value = dayjs().format('YYYY-MM-DDTHH:mm')
    loadedNext.value = ''
    loadedAnchor.value = 0
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
  amountPending.value = !!s.amountPending
  fxEnabled.value = !!(s.foreignCurrency && s.foreignAmount)
  fxCurrency.value = s.foreignCurrency || 'USD'
  fxAmountYuan.value = s.foreignAmount ? fenToYuan(s.foreignAmount) : ''
  nextLocal.value = dayjs(s.nextRunAt).format('YYYY-MM-DDTHH:mm')
  loadedNext.value = nextLocal.value
  loadedAnchor.value = s.anchorDay || 0
}

async function load() {
  await Promise.all([accounts.ensure(), categories.ensure()])
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
  const isFlow = form.type !== 'transfer'
  const useFx = isFlow && fxEnabled.value
  const fxCode = fxCurrency.value.trim().toUpperCase()
  const fxFen = useFx ? yuanToFen(fxAmountYuan.value || '0') : 0
  error.value = ''
  if (useFx && !/^[A-Z]{3,8}$/.test(fxCode)) { error.value = '请选择或填写外币币种（如 USD）'; return }
  if (useFx && fxFen <= 0) { error.value = '请填写外币金额'; return }
  saving.value = true
  try {
    const payload = {
      ...form,
      amountPending: isFlow && amountPending.value,
      foreignCurrency: useFx ? fxCode : '',
      foreignAmount: fxFen,
      intervalN: intervalN.value,
      anchorDay: anchorDay.value,
      enabled: props.scheduleId ? formEnabled.value : true,
      amount: yuanToFen(amountYuan.value),
      nextRunAt: dayjs(nextLocal.value).toISOString(),
    }
    if (props.scheduleId) await http.put(`/schedules/${props.scheduleId}`, payload)
    else await http.post('/schedules', payload)
    emit('saved')
  } catch (e: any) {
    error.value = e?.message || '保存失败'
  } finally {
    saving.value = false
  }
}

defineExpose({ save, saving })
</script>

<style scoped>
.mode-label { font-size: 0.78rem; color: var(--muted); margin-bottom: 6px; }
.mode-toggle { width: 100%; }
.mode-toggle :deep(.v-btn) { flex: 1; }
.interval-row { display: flex; align-items: center; gap: 8px; }
.interval-input { flex: 0 0 96px; }
.unit-select { flex: 1; min-width: 0; }
.quick-days { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 8px; }
.rule-hint {
  margin-top: 6px;
  font-size: 0.76rem;
  color: var(--muted);
  line-height: 1.4;
  overflow-wrap: anywhere;
}
</style>
