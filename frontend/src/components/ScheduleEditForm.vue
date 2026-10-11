<template>
  <div class="form-sections">
    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('schedules.form.basic') }}</h3>
      </header>
      <div class="form-section__body">
        <v-text-field v-model="form.name" :label="t('schedules.form.name')" variant="outlined" hide-details class="field" />
        <v-select
          v-model="form.type"
          :items="typeItems"
          :label="t('schedules.form.type')"
          variant="outlined"
          hide-details
          class="field"
        />
        <v-text-field
          v-model="amountYuan"
          :label="form.type !== 'transfer' && amountPending ? t('schedules.form.estimatedAmount') : t('schedules.form.amount')"
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
        <h3 class="form-section__title">{{ t('schedules.form.foreign') }}</h3>
      </header>
      <div class="form-section__body">
        <div class="form-section__group">
          <MetaHead
            icon="mdi-timer-sand"
            :title="t('schedules.form.pendingTitle')"
            :subtitle="t('schedules.form.pendingSub')"
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
            :subtitle="t('schedules.form.foreignSub')"
          />
        </div>
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ form.type === 'transfer' ? t('schedules.form.account') : t('schedules.form.accountAndCategory') }}</h3>
      </header>
      <div class="form-section__body">
        <v-select
          v-model="form.accountId"
          :items="accounts.list"
          item-title="name"
          item-value="id"
          :label="form.type === 'transfer' ? t('schedules.form.fromAccount') : t('schedules.form.account')"
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
          :label="t('schedules.form.toAccount')"
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
          :label="t('schedules.form.category')"
          variant="outlined"
          hide-details
          class="field field--last"
        />
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('schedules.form.cycle') }}</h3>
      </header>
      <div class="form-section__body">
        <div class="field">
          <div class="mode-label">{{ t('schedules.form.repeatMode') }}</div>
          <v-btn-toggle v-model="repeatMode" mandatory density="comfortable" color="primary" variant="outlined" divided class="mode-toggle">
            <v-btn value="calendar">{{ t('schedules.form.calendar') }}</v-btn>
            <v-btn value="fixed">{{ t('schedules.form.fixed') }}</v-btn>
          </v-btn-toggle>
        </div>

        <template v-if="repeatMode === 'calendar'">
          <div class="interval-row field">
            <v-text-field
              v-model.number="form.intervalN"
              :label="t('schedules.form.every')"
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
              :label="t('schedules.form.intervalDays')"
              type="number"
              min="1"
              variant="outlined"
              hide-details
              :suffix="t('schedules.form.daysSuffix')"
            />
            <div class="quick-days">
              <v-chip
                v-for="d in quickDays"
                :key="d"
                size="small"
                :color="form.intervalN === d ? 'primary' : undefined"
                :variant="form.intervalN === d ? 'flat' : 'outlined'"
                @click="form.intervalN = d"
              >{{ t('schedules.form.nDays', { n: d }) }}</v-chip>
            </div>
          </div>
        </template>

        <div class="field field--last">
          <NativeDateField
            v-model="nextLocal"
            :label="scheduleId ? t('schedules.form.nextRunAt') : t('schedules.form.firstRunAt')"
            type="datetime-local"
            variant="outlined"
            hide-details
          />
          <div class="rule-hint">{{ ruleHint }}</div>
          <div v-if="previewDates.length" class="rule-hint">{{ t('schedules.form.upcoming', { dates: previewDates.join(t('schedules.form.dateSep')) }) }}</div>
        </div>
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('schedules.form.other') }}</h3>
      </header>
      <div class="form-section__body">
        <v-text-field
          v-model="form.remark"
          :label="t('schedules.form.remark')"
          variant="outlined"
          hide-details
          :class="scheduleId ? 'field' : 'field field--last'"
        />
        <v-switch
          v-if="scheduleId"
          v-model="formEnabled"
          :label="t('schedules.form.enabled')"
          color="primary"
          hide-details
          class="field field--last"
        />
      </div>
    </section>

    <v-alert v-if="error" type="error" density="compact">{{ error }}</v-alert>

    <div v-if="scheduleId" class="danger-zone">
      <div class="danger-title">{{ t('schedules.form.dangerTitle') }}</div>
      <p class="danger-tip">{{ t('schedules.form.dangerTip') }}</p>
      <v-btn color="error" variant="outlined" block @click="emit('delete')">{{ t('schedules.form.delete') }}</v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
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

const { t, locale } = useI18n()
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
const typeItems = computed(() => [
  { title: t('txType.expense'), value: 'expense' },
  { title: t('txType.income'), value: 'income' },
  { title: t('txType.transfer'), value: 'transfer' },
])
const unitItems = computed(() => [
  { title: t('schedules.form.units.daily'), value: 'daily' },
  { title: t('schedules.form.units.weekly'), value: 'weekly' },
  { title: t('schedules.form.units.monthly'), value: 'monthly' },
  { title: t('schedules.form.units.yearly'), value: 'yearly' },
])
const quickDays = [7, 14, 28, 30, 90]
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
  // dayjs 的 locale 不是响应式的，读一次 locale 让切换语言时重算星期 / 月份名
  void locale.value
  const at = dayjs(nextLocal.value)
  if (!at.isValid()) return ''
  const n = intervalN.value
  const hm = at.format('HH:mm')
  const d = anchorDay.value
  switch (form.frequency) {
    case 'daily': return n > 1 ? t('schedules.rule.dailyN', { n, hm }) : t('schedules.rule.daily', { hm })
    case 'weekly': {
      const p = { n, hm, wd: at.format('ddd'), weekday: at.format('dddd') }
      return n > 1 ? t('schedules.rule.weeklyN', p) : t('schedules.rule.weekly', p)
    }
    case 'monthly':
      return (n > 1 ? t('schedules.rule.monthlyN', { n, d }) : t('schedules.rule.monthly', { d }))
        + (d > 28 ? t('schedules.rule.monthEnd') : '')
    case 'yearly': {
      const p = { n, d, m: at.month() + 1, month: at.format('MMMM') }
      return n > 1 ? t('schedules.rule.yearlyN', p) : t('schedules.rule.yearly', p)
    }
    default: return t('schedules.rule.fixed', { n })
  }
})

const previewDates = computed(() => {
  void locale.value
  let at = dayjs(nextLocal.value)
  if (!at.isValid()) return []
  const thisYear = dayjs().year()
  const out: string[] = []
  for (let i = 0; i < 4; i++) {
    out.push(at.format(at.year() === thisYear ? t('schedules.form.dateFmt') : t('schedules.form.dateFmtYear')))
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
  if (useFx && !/^[A-Z]{3,8}$/.test(fxCode)) { error.value = t('schedules.form.fxCurrencyRequired'); return }
  if (useFx && fxFen <= 0) { error.value = t('schedules.form.fxAmountRequired'); return }
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
    error.value = e?.message || t('schedules.form.saveFailed')
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
