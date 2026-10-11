<template>
  <div class="page te-page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" @click="$router.back()" />
      <h1>{{ isEdit ? t('txEdit.titleEdit') : t('txEdit.titleNew') }}</h1>
      <v-btn v-if="isEdit" icon="mdi-delete" variant="text" color="error" @click="onDelete" />
      <span v-else style="width:40px" />
    </div>

    <div class="te-layout">
    <div class="te-main">

    <div v-if="!isEdit" class="ai-row mb-3">
      <v-btn variant="tonal" color="secondary" class="flex1" to="/transactions/recognize">
        <v-icon start>mdi-robot</v-icon>
        {{ t('txEdit.textRecognize') }}
      </v-btn>
      <v-btn variant="tonal" color="secondary" class="flex1" @click="showAI = true">
        <v-icon start>mdi-camera</v-icon>
        {{ t('txEdit.receiptRecognize') }}
      </v-btn>
    </div>

    <div v-if="!isEdit && templates.length" class="tpls mb-3">
      <v-chip
        v-for="tpl in templates"
        :key="tpl.id"
        class="tpl-chip"
        color="primary"
        variant="outlined"
        @click="applyTemplate(tpl)"
      >{{ tpl.name }}</v-chip>
    </div>

    <v-btn-toggle v-model="uiMode" mandatory color="primary" class="mb-2 w100" divided>
      <v-btn value="flow" class="flex1">{{ t('txEdit.modeFlow') }}</v-btn>
      <v-btn value="transfer" class="flex1">{{ t('txType.transfer') }}</v-btn>
      <v-btn value="repay" class="flex1">{{ t('txEdit.modeRepay') }}</v-btn>
    </v-btn-toggle>

    <div v-if="uiMode === 'flow'" class="flow-sub mb-3">
      <v-btn-toggle v-model="flowKind" mandatory color="primary" density="comfortable" class="flow-kind" divided>
        <v-btn value="expense" class="flex1" size="small">{{ t('txType.expense') }}</v-btn>
        <v-btn value="income" class="flex1" size="small">{{ t('txType.income') }}</v-btn>
      </v-btn-toggle>
    </div>
    <div v-if="!isEdit && uiMode === 'flow' && flowKind === 'income'" class="mb-3">
      <v-btn variant="tonal" color="info" block @click="refundOpen = true">
        <v-icon start>mdi-ticket-percent-outline</v-icon>
        {{ t('txEdit.refundFromVoucher') }}
      </v-btn>
    </div>
    <i18n-t v-else-if="uiMode === 'repay'" keypath="txEdit.repayHint" tag="p" class="mode-hint mb-3">
      <template #bold><strong>{{ t('txEdit.repayHintBold') }}</strong></template>
    </i18n-t>
    <p v-else class="mode-hint mb-3">{{ t('txEdit.transferHint') }}</p>

    <FareRuleSection
      v-if="!isEdit && uiMode === 'flow' && flowKind === 'expense'"
      v-model:enabled="fareEnabled"
      v-model:rule-id="form.fareRuleId"
      :items="fareRuleItems"
      :loading="farePreviewLoading"
      :hint="fareHint"
      @toggle="onFareToggle"
      @picked="onFareRulePicked"
      @apply="applySuggestedFare(true)"
    />

    <div v-if="wasPending" class="pending-box mb-3">
      <div class="pending-title">
        <v-icon size="18">mdi-timer-sand</v-icon>
        {{ t('txEdit.pendingTitle') }}
      </div>
      <p class="pending-tip">{{ t('txEdit.pendingTip') }}</p>
      <v-checkbox
        v-model="keepPending"
        :label="t('txEdit.keepPending')"
        color="primary"
        density="compact"
        hide-details
      />
    </div>

    <v-text-field
      v-model="amountYuan"
      :label="wasPending && keepPending ? t('txEdit.amountEstimated') : t('txEdit.amount')"
      type="number"
      inputmode="decimal"
      variant="outlined"
      hide-details
      class="amount-field mb-3"
      prefix="¥"
      @update:model-value="repay.onAmountManualEdit"
    />

    <section v-if="form.type !== 'transfer'" class="form-section mb-3">
      <div class="form-section__body">
        <ForeignPaymentFields
          v-model:enabled="fxEnabled"
          v-model:currency="fxCurrency"
          v-model:amount-yuan="fxAmountYuan"
          :cny-fen="yuanToFen(amountYuan || '0')"
          :estimated="wasPending && keepPending"
        />
      </div>
    </section>

    <CreditExtrasPanel
      v-if="showCreditExtras"
      v-model:periods="installmentPeriods"
      v-model:interest-yuan="interestYuan"
      v-model:defer-cycles="form.installmentDeferCycles"
      :billing-day="creditBillingDay"
      :happened-at="happenedLocal"
      :statement="repay.statement.value"
      :total-outstanding="repay.totalOutstanding.value"
      :is-credit-expense="isCreditExpense"
      :show-repay-actions="repayActive"
      :repay-loading="repay.loading.value"
      :amount-fen="yuanToFen(amountYuan || '0')"
      :repay-base-fen="repay.baseFen.value"
      @plan-changed="repay.onPlanChanged"
      @suggest="repay.apply(true)"
      @early-repay="repay.applyEarlyRepay"
    />

    <section class="form-section mb-3">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('txEdit.details') }}</h3>
      </header>
      <div class="form-section__body">
        <v-select
          v-model="form.accountId"
          :items="accounts.list"
          item-title="name"
          item-value="id"
          :label="form.type === 'transfer' ? t('txEdit.fromAccount') : t('txEdit.account')"
          variant="outlined"
          hide-details
          class="field"
        />
        <v-select
          v-if="form.type === 'transfer'"
          v-model="form.toAccountId"
          :items="transferToOptions"
          item-title="name"
          item-value="id"
          :label="repayActive ? t('txEdit.creditAccount') : t('txEdit.toAccount')"
          variant="outlined"
          hide-details
          class="field"
        />
        <div v-else class="form-field-row field">
          <v-select
            v-model="form.categoryId"
            :items="form.type === 'expense' ? categories.expenseOptions : categories.incomeOptions"
            item-title="name"
            item-value="id"
            :label="t('txEdit.category')"
            variant="outlined"
            class="flex1"
            hide-details
          />
          <v-btn icon="mdi-plus" variant="tonal" color="primary" :title="t('txEdit.addCategory')" @click="showQuickCategory = true" />
        </div>

        <div class="form-field-row field">
          <v-select
            v-model="form.tagIds"
            :items="tags.list"
            item-title="name"
            item-value="id"
            :label="t('txEdit.tag')"
            multiple
            chips
            closable-chips
            variant="outlined"
            class="flex1"
            hide-details
          />
          <v-btn icon="mdi-plus" variant="tonal" color="primary" :title="t('txEdit.addTag')" @click="showQuickTag = true" />
        </div>

        <v-text-field v-model="form.remark" :label="t('txEdit.remark')" variant="outlined" hide-details class="field" />
        <NativeDateField v-model="happenedLocal" :label="t('txEdit.time')" type="datetime-local" variant="outlined" hide-details class="field field--last" />
      </div>
    </section>

    <section v-if="form.type === 'expense'" class="form-section mb-3">
      <header class="form-section__head voucher-head">
        <h3 class="form-section__title">
          {{ t('txEdit.voucher') }}
          <v-chip
            v-if="isEdit && voucherStatus"
            size="x-small"
            :color="voucherStatusColor[voucherStatus]"
            variant="tonal"
            class="ml-1"
          >{{ voucherStatusLabel(voucherStatus) }}</v-chip>
        </h3>
        <v-switch
          v-model="voucherOn"
          color="primary"
          density="compact"
          hide-details
          inset
          :disabled="voucherStatus === 'refunded'"
        />
      </header>
      <div v-if="voucherOn" class="form-section__body">
        <NativeDateField
          v-model="voucherExpireOn"
          :label="t('txEdit.voucherExpire')"
          variant="outlined"
          hide-details
          class="field"
        />
        <p v-if="voucherStatus === 'refunded'" class="form-section__tip">{{ t('txEdit.voucherRefundedTip') }}</p>
        <p v-else class="form-section__tip">{{ t('txEdit.voucherTip') }}</p>
        <div v-if="isEdit && savedVoucherStatus && voucherStatus !== 'refunded'" class="voucher-ops">
          <v-btn
            v-if="voucherStatus === 'pending'"
            size="small"
            variant="tonal"
            color="success"
            :loading="voucherBusy"
            @click="toggleUsed('used')"
          >{{ t('txEdit.markUsed') }}</v-btn>
          <v-btn
            v-else
            size="small"
            variant="text"
            :loading="voucherBusy"
            @click="toggleUsed('pending')"
          >{{ t('txEdit.unmarkUsed') }}</v-btn>
          <v-btn size="small" variant="tonal" color="info" @click="refundOpen = true">{{ t('txEdit.refund') }}</v-btn>
        </div>
      </div>
    </section>

    <div v-if="refundOfId" class="refund-link mb-3">
      <v-icon size="16">mdi-ticket-percent-outline</v-icon>
      <span class="flex1">{{ t('txEdit.refundLink') }}</span>
      <v-btn size="small" variant="text" color="primary" :to="`/transactions/${refundOfId}`">{{ t('txEdit.viewVoucher') }}</v-btn>
    </div>

    <FareRuleSection
      v-if="isEdit"
      v-model:enabled="fareEnabled"
      v-model:rule-id="form.fareRuleId"
      :items="fareRuleItems"
      :loading="farePreviewLoading"
      :hint="fareHint"
      @toggle="onFareToggle"
      @picked="onFareRulePicked"
      @apply="applySuggestedFare(true)"
    />
    </div>

    <aside class="te-side">
    <AttachmentUpload v-model="attachmentIds" :existing="existingAttachments" />

    <LocationSection v-model:geo="geo" />
    </aside>
    </div>

    <v-alert v-if="error" type="error" class="mb-3" density="compact">{{ error }}</v-alert>
    <div class="save-bar">
      <v-btn color="primary" block size="large" class="hero-cta" :loading="saving" @click="onSave">{{ t('common.save') }}</v-btn>
    </div>

    <QuickNameDialog
      v-model="showQuickCategory"
      :title="t('txEdit.addCategory')"
      :label="t('txEdit.categoryName')"
      :hint="t('txEdit.categoryHint', { kind: form.type === 'income' ? t('txType.income') : t('txType.expense') })"
      :save="createCategory"
    />
    <QuickNameDialog
      v-model="showQuickTag"
      :title="t('txEdit.addTag')"
      :label="t('txEdit.tagName')"
      :save="createTag"
    />
    <ReceiptRecognizeDialog v-model="showAI" @recognized="applyRecognized" @attached="onReceiptAttached" />
    <VoucherRefundDialog v-model="refundOpen" :voucher="refundVoucher" @refunded="onRefunded" />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import http from '../api/http'
import { useAccountStore } from '../stores/account'
import { useCategoryStore } from '../stores/category'
import { useTagStore } from '../stores/tag'
import {
  useTransactionStore,
  voucherStatusColor,
  voucherStatusLabel,
  type Transaction,
  type VoucherStatus,
} from '../stores/transaction'
import { useAuthStore } from '../stores/auth'
import { useRepaySuggestion } from '../composables/useRepaySuggestion'
import { yuanToFen, fenToYuan } from '../utils/money'
import { cityAllowsAuto, getCachedCity } from '../utils/fare'
import AttachmentUpload from '../components/AttachmentUpload.vue'
import NativeDateField from '../components/NativeDateField.vue'
import ForeignPaymentFields from '../components/ForeignPaymentFields.vue'
import FareRuleSection from '../components/FareRuleSection.vue'
import CreditExtrasPanel from '../components/CreditExtrasPanel.vue'
import LocationSection, { type GeoValue } from '../components/LocationSection.vue'
import QuickNameDialog from '../components/QuickNameDialog.vue'
import ReceiptRecognizeDialog from '../components/ReceiptRecognizeDialog.vue'
import VoucherRefundDialog from '../components/VoucherRefundDialog.vue'

const props = defineProps<{ id?: string }>()
const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const accounts = useAccountStore()
const categories = useCategoryStore()
const tags = useTagStore()
const auth = useAuthStore()
const tx = useTransactionStore()

const isEdit = computed(() => !!(props.id || route.params.id))
const editId = computed(() => Number(props.id || route.params.id || 0))
/** 主类型：收支 | 转账 | 还款 */
const uiMode = ref<'flow' | 'transfer' | 'repay'>('flow')
/** 收支子类型 */
const flowKind = ref<'expense' | 'income'>('expense')
const isRepayMode = computed(() => uiMode.value === 'repay')

const amountYuan = ref('')
const happenedLocal = ref(dayjs().format('YYYY-MM-DDTHH:mm'))
const attachmentIds = ref<number[]>([])
const existingAttachments = ref<{ id: number; url: string }[]>([])
const saving = ref(false)
const error = ref('')

const templates = ref<any[]>([])
const fareRules = ref<any[]>([])
const fareEnabled = ref(false)
const farePreviewLoading = ref(false)
const fareHint = ref('')
const showAI = ref(false)
const showQuickCategory = ref(false)
const showQuickTag = ref(false)

const form = reactive({
  type: 'expense' as 'expense' | 'income' | 'transfer',
  accountId: undefined as number | undefined,
  toAccountId: undefined as number | undefined,
  categoryId: undefined as number | undefined,
  tagIds: [] as number[],
  remark: '',
  geoMode: 'point' as 'point' | 'route' | '',
  geoLng: null as number | null,
  geoLat: null as number | null,
  geoName: '',
  geoEndLng: null as number | null,
  geoEndLat: null as number | null,
  geoEndName: '',
  fareRuleId: null as number | null,
  installmentPeriods: 1,
  installmentDeferCycles: 0,
})
const interestYuan = ref('0')
const fxEnabled = ref(false)
const fxCurrency = ref('USD')
const fxAmountYuan = ref('')
/** 编辑的流水原本是否待确认；保存时默认清除，勾选 keepPending 则保留 */
const wasPending = ref(false)
const keepPending = ref(false)
let syncingUiMode = false

const voucherOn = ref(false)
/** 当前券状态（开关打开且原本非券时为 pending） */
const voucherStatus = ref<VoucherStatus>('')
/** 库内已保存的券状态；核销 / 退款按钮只对已保存的券可用 */
const savedVoucherStatus = ref<VoucherStatus>('')
const voucherExpireOn = ref('')
const voucherBusy = ref(false)
const refundOpen = ref(false)
const refundOfId = ref<number | null>(null)
const editingTx = ref<Transaction | null>(null)
/** 编辑页退款：直接退当前这张券；新建收入页：弹窗里挑券 */
const refundVoucher = computed(() => (isEdit.value ? editingTx.value : null))

watch(voucherOn, (on) => {
  if (on && !voucherStatus.value) voucherStatus.value = 'pending'
  if (!on && voucherStatus.value !== 'refunded') voucherStatus.value = ''
})

async function toggleUsed(status: 'pending' | 'used') {
  voucherBusy.value = true
  try {
    const updated = await tx.setVoucherStatus(editId.value, status)
    voucherStatus.value = updated.voucherStatus || ''
    savedVoucherStatus.value = voucherStatus.value
    editingTx.value = updated
  } catch (e: any) {
    error.value = e?.message || t('common.failed')
  } finally {
    voucherBusy.value = false
  }
}

async function onRefunded() {
  await router.replace('/transactions')
}

const geo = computed<GeoValue>({
  get: () => ({
    mode: form.geoMode,
    lng: form.geoLng,
    lat: form.geoLat,
    name: form.geoName,
    endLng: form.geoEndLng,
    endLat: form.geoEndLat,
    endName: form.geoEndName,
  }),
  set: (v) => {
    form.geoMode = v.mode
    form.geoLng = v.lng
    form.geoLat = v.lat
    form.geoName = v.name
    form.geoEndLng = v.endLng
    form.geoEndLat = v.endLat
    form.geoEndName = v.endName
  },
})

const installmentPeriods = computed({
  get: () => Math.max(1, Number(form.installmentPeriods) || 1),
  set: (v: number) => { form.installmentPeriods = v },
})

const fareRuleItems = computed(() =>
  (fareRules.value || [])
    .filter((r) => r.enabled !== false)
    .map((r) => ({
      title: r.city ? t('txEdit.fareRuleCity', { name: r.name, city: r.city }) : r.name,
      value: r.id,
    })),
)

const isCreditRepay = computed(() => {
  if (form.type !== 'transfer' || !form.toAccountId) return false
  return accounts.list.some((a) => a.id === form.toAccountId && a.type === 'credit')
})
const isCreditExpense = computed(() => {
  if (form.type !== 'expense' || !form.accountId) return false
  return accounts.list.some((a) => a.id === form.accountId && a.type === 'credit')
})
const creditBillingDay = computed(() =>
  isCreditExpense.value ? accounts.list.find((a) => a.id === form.accountId)?.billingDay || 0 : 0,
)
const repayActive = computed(() => isCreditRepay.value || isRepayMode.value)
const showCreditExtras = computed(() => isCreditExpense.value || repayActive.value)

const repay = useRepaySuggestion({
  amountYuan,
  interestYuan,
  periods: installmentPeriods,
  active: repayActive,
})

const transferToOptions = computed(() => {
  const list = accounts.list.filter((a) => a.id !== form.accountId && !a.archived)
  if (isRepayMode.value) {
    const credits = list.filter((a) => a.type === 'credit')
    return credits.length ? credits : list
  }
  return list
})

watch(() => form.type, (type, prev) => {
  if (syncingUiMode || type === prev) return
  form.categoryId = undefined
  if (!(isRepayMode.value && type === 'transfer')) {
    form.toAccountId = undefined
  }
})

/** 支出用默认消费账户；收入 / 转账 / 还款用默认还款账户。未设置或已归档则不改动 */
function defaultAccountFor(kind: 'expense' | 'income' | 'transfer' | 'repay') {
  const id = kind === 'expense' ? auth.profile?.defaultExpenseAccountId : auth.profile?.defaultAccountId
  if (!id) return undefined
  const acc = accounts.list.find((a) => a.id === id && !a.archived)
  if (!acc) return undefined
  if (kind !== 'expense' && acc.type === 'credit') return undefined
  return acc.id
}

/** 模板 / 识别 / 路由参数已指定账户时，本轮模式切换不再套默认账户 */
let skipDefaultAccount = false
function suppressDefaultAccountOnce() {
  skipDefaultAccount = true
  void nextTick(() => { skipDefaultAccount = false })
}

function applyDefaultAccount(kind: 'expense' | 'income' | 'transfer' | 'repay') {
  if (isEdit.value || skipDefaultAccount) return
  const id = defaultAccountFor(kind)
  if (!id) return
  form.accountId = id
  if (form.toAccountId === id) form.toAccountId = undefined
}

watch(flowKind, (k) => {
  if (syncingUiMode || uiMode.value !== 'flow') return
  if (form.type !== k) {
    form.type = k
    form.categoryId = undefined
  }
  applyDefaultAccount(k)
})

watch(uiMode, (mode, prev) => {
  if (syncingUiMode || mode === prev) return
  enterUiMode(mode)
})

function enterUiMode(mode: 'flow' | 'transfer' | 'repay', useDefaultAccount = true) {
  syncingUiMode = true
  try {
    if (mode === 'flow') {
      form.type = flowKind.value
      form.toAccountId = undefined
      if (form.remark === t('txEdit.repayRemark')) form.remark = ''
      repay.reset()
      if (useDefaultAccount) applyDefaultAccount(flowKind.value)
    } else if (mode === 'transfer') {
      form.type = 'transfer'
      form.categoryId = undefined
      if (form.remark === t('txEdit.repayRemark')) form.remark = ''
      // 普通转账不强制信用账户
      if (form.toAccountId && accounts.list.some((a) => a.id === form.toAccountId && a.type === 'credit')) {
        form.toAccountId = undefined
      }
      repay.reset()
      if (useDefaultAccount) applyDefaultAccount('transfer')
    } else {
      form.type = 'transfer'
      form.categoryId = undefined
      form.remark = form.remark || t('txEdit.repayRemark')
      repay.auto.value = true
      form.installmentPeriods = Math.max(1, form.installmentPeriods || 1)
      if (useDefaultAccount) applyDefaultAccount('repay')
      const credit =
        accounts.list.find((a) => a.id === form.toAccountId && a.type === 'credit') ||
        accounts.list.find((a) => a.type === 'credit' && !a.archived)
      if (credit) form.toAccountId = credit.id
      if (form.accountId === form.toAccountId) {
        form.accountId = accounts.list.find((a) => a.id !== form.toAccountId && a.type !== 'credit')?.id
          || accounts.list.find((a) => a.id !== form.toAccountId)?.id
      }
      void repay.load(credit?.id ?? form.toAccountId)
    }
  } finally {
    void nextTick(() => { syncingUiMode = false })
  }
}

/** 主数据：已有缓存则跳过，避免编辑页再等一轮远程库 */
async function ensureCoreStores() {
  const jobs: Promise<unknown>[] = [accounts.ensure(), categories.ensure(), tags.ensure()]
  if (!auth.profile) jobs.push(auth.fetchMe().catch(() => {}))
  await Promise.all(jobs)
}

async function loadSideData() {
  try {
    const [{ data: tpls }, { data: rules }] = await Promise.all([
      http.get('/templates'),
      http.get('/fare-rules'),
    ])
    templates.value = tpls || []
    fareRules.value = rules || []
  } catch { /* ignore */ }
}

function applyEditTransaction(item: Awaited<ReturnType<typeof tx.get>>) {
  if (item.type === 'revalue') return
  syncingUiMode = true
  form.type = item.type
  if (item.type === 'expense' || item.type === 'income') {
    uiMode.value = 'flow'
    flowKind.value = item.type
  } else {
    const toCredit = item.toAccountId && accounts.list.some((a) => a.id === item.toAccountId && a.type === 'credit')
    uiMode.value = toCredit ? 'repay' : 'transfer'
  }
  form.accountId = item.accountId
  form.toAccountId = item.toAccountId || undefined
  form.categoryId = item.categoryId || undefined
  form.tagIds = (item.tags || []).map((x) => x.id)
  form.remark = item.remark || ''
  form.geoMode = (item.geoMode as 'point' | 'route') || (item.geoLng != null ? 'point' : '')
  form.geoLng = item.geoLng ?? null
  form.geoLat = item.geoLat ?? null
  form.geoName = item.geoName || ''
  form.geoEndLng = item.geoEndLng ?? null
  form.geoEndLat = item.geoEndLat ?? null
  form.geoEndName = item.geoEndName || ''
  form.fareRuleId = item.fareRuleId || null
  form.installmentPeriods = item.installmentPeriods || 1
  form.installmentDeferCycles = item.installmentDeferCycles || 0
  interestYuan.value = fenToYuan(item.interestFen || 0)
  fareEnabled.value = !!item.fareRuleId
  amountYuan.value = fenToYuan(item.amount)
  fxEnabled.value = !!(item.foreignCurrency && item.foreignAmount)
  fxCurrency.value = item.foreignCurrency || 'USD'
  fxAmountYuan.value = item.foreignAmount ? fenToYuan(item.foreignAmount) : ''
  wasPending.value = !!item.amountPending
  keepPending.value = false
  editingTx.value = item
  voucherStatus.value = item.voucherStatus || ''
  savedVoucherStatus.value = voucherStatus.value
  voucherOn.value = !!item.voucherStatus
  voucherExpireOn.value = item.voucherExpireOn || ''
  refundOfId.value = item.refundOfId || null
  happenedLocal.value = dayjs(item.happenedAt).format('YYYY-MM-DDTHH:mm')
  existingAttachments.value = (item.attachments || []).map((a) => ({ id: a.id, url: a.url }))
  attachmentIds.value = existingAttachments.value.map((a) => a.id)
  // 等 watch 跑完再放开，避免 form.type 变更把刚写入的 categoryId 清掉
  void nextTick(() => { syncingUiMode = false })
}

onMounted(async () => {
  // 模板/计费不挡类型与标签回显
  void loadSideData()

  const storesP = ensureCoreStores()
  // 编辑：流水与字典并行拉取，不再串行干等
  const txP = isEdit.value ? tx.get(editId.value) : null

  if (txP) {
    const [item] = await Promise.all([txP, storesP])
    if (item.type === 'revalue') {
      void router.replace(`/accounts/${item.accountId}/ledger`)
      return
    }
    applyEditTransaction(item)
    if (uiMode.value === 'repay' && form.toAccountId) {
      void repay.load(form.toAccountId)
    }
    return
  }

  await storesP

  if (!form.accountId) {
    const kind = uiMode.value === 'flow' ? flowKind.value : uiMode.value
    form.accountId =
      defaultAccountFor(kind) ||
      accounts.list.find((a) => a.type !== 'credit' && !a.archived)?.id ||
      accounts.list[0]?.id
  }
  // 账户页 / 首页还款提醒入口
  if (route.query.mode === 'repay') {
    syncingUiMode = true
    uiMode.value = 'repay'
    syncingUiMode = false
    suppressDefaultAccountOnce()
    const qTo = Number(route.query.toAccountId || 0)
    const qFrom = Number(route.query.fromAccountId || 0)
    const credit =
      accounts.list.find((a) => a.id === qTo && a.type === 'credit') ||
      accounts.list.find((a) => a.type === 'credit' && !a.archived)
    if (qFrom && accounts.list.some((a) => a.id === qFrom && a.type !== 'credit')) {
      form.accountId = qFrom
    } else {
      const def = auth.profile?.defaultAccountId
      const preferred = accounts.list.find((a) => a.id === def && a.type !== 'credit')
        || accounts.list.find((a) => a.type === 'bank' && !a.archived)
        || accounts.list.find((a) => a.type !== 'credit' && !a.archived)
      if (preferred) form.accountId = preferred.id
    }
    const qFen = Number(route.query.amountFen || 0)
    repay.presetBase(qFen)
    enterUiMode('repay', false)
    if (credit) form.toAccountId = credit.id
    if (form.accountId === form.toAccountId) {
      form.accountId = accounts.list.find((a) => a.id !== form.toAccountId && a.type !== 'credit')?.id
        || accounts.list.find((a) => a.id !== form.toAccountId)?.id
    }
    if (qFen > 0) repay.apply(true)
    await repay.load(credit?.id ?? form.toAccountId)
  }

  // 首页快捷模板缺字段时跳转至此，预填模板供编辑
  const tplId = Number(route.query.templateId || 0)
  if (tplId > 0) {
    await loadSideData()
    const tpl = templates.value.find((x) => x.id === tplId)
    if (tpl) await applyTemplate(tpl)
  }
})

watch(
  () => form.toAccountId,
  (id) => {
    if (isEdit.value) return
    if (form.type === 'transfer' && id && accounts.list.some((a) => a.id === id && a.type === 'credit')) {
      repay.auto.value = true
      void repay.load(id)
    }
  },
)

function applyRecognized(data: any) {
  if (data.type) {
    form.type = data.type
    syncingUiMode = true
    if (data.type === 'expense' || data.type === 'income') {
      uiMode.value = 'flow'
      flowKind.value = data.type
    } else if (data.type === 'transfer') {
      const toCredit = data.toAccountId && accounts.list.some((a) => a.id === data.toAccountId && a.type === 'credit')
      uiMode.value = toCredit ? 'repay' : 'transfer'
    }
    syncingUiMode = false
    suppressDefaultAccountOnce()
  }
  if (data.amountYuan) amountYuan.value = String(data.amountYuan)
  if (data.accountId) form.accountId = data.accountId
  if (data.toAccountId) form.toAccountId = data.toAccountId
  if (data.categoryId) form.categoryId = data.categoryId
  if (data.tagIds) form.tagIds = data.tagIds
  if (data.remark) form.remark = data.remark
  if (data.happenedAt) happenedLocal.value = dayjs(data.happenedAt).format('YYYY-MM-DDTHH:mm')
}

function onReceiptAttached(att: { id: number; url: string }) {
  attachmentIds.value = [...attachmentIds.value, att.id]
  existingAttachments.value = [...existingAttachments.value, att]
}

async function applyTemplate(tpl: any) {
  syncingUiMode = true
  form.type = tpl.type
  if (tpl.type === 'expense' || tpl.type === 'income') {
    uiMode.value = 'flow'
    flowKind.value = tpl.type
  } else {
    const toCredit = tpl.toAccountId && accounts.list.some((a) => a.id === tpl.toAccountId && a.type === 'credit')
    uiMode.value = toCredit ? 'repay' : 'transfer'
  }
  syncingUiMode = false
  suppressDefaultAccountOnce()
  if (tpl.accountId) form.accountId = tpl.accountId
  form.toAccountId = tpl.toAccountId || undefined
  form.categoryId = tpl.categoryId || undefined
  form.tagIds = tpl.tagIds || []
  form.remark = tpl.remark?.trim() || tpl.name || ''
  if (tpl.geoLng != null && tpl.geoLat != null) {
    form.geoMode = (tpl.geoMode as 'point' | 'route') || (tpl.geoEndLng != null ? 'route' : 'point')
    form.geoLng = tpl.geoLng
    form.geoLat = tpl.geoLat
    form.geoName = tpl.geoName || ''
    if (form.geoMode === 'route') {
      form.geoEndLng = tpl.geoEndLng ?? null
      form.geoEndLat = tpl.geoEndLat ?? null
      form.geoEndName = tpl.geoEndName || ''
    } else {
      form.geoEndLng = null
      form.geoEndLat = null
      form.geoEndName = ''
    }
  }

  // 先写入模板金额，避免选规则触发 preview 时原价为空被写成 0
  const tplAmount = Number(tpl.amount) || 0
  if (tplAmount > 0) amountYuan.value = fenToYuan(tplAmount)

  const ruleId = tpl.fareRuleId || tpl.fareRule?.id
  const rule = fareRules.value.find((r) => r.id === ruleId) || tpl.fareRule
  if (ruleId && rule) {
    const userCity = getCachedCity()
    if (cityAllowsAuto(rule.city || '', userCity)) {
      fareEnabled.value = true
      applyingTemplateFare.value = true
      try {
        form.fareRuleId = ruleId
        // 模板金额作原价算建议价并填入
        await applySuggestedFare(true, tplAmount > 0 ? tplAmount : undefined)
      } finally {
        applyingTemplateFare.value = false
      }
      return
    }
    fareHint.value = rule.city ? t('txEdit.fareCityMismatch', { city: rule.city }) : ''
  }
}

const applyingTemplateFare = ref(false)

function onFareToggle(on: boolean) {
  if (!on) {
    form.fareRuleId = null
    fareHint.value = ''
  }
}

async function onFareRulePicked() {
  // 套用模板时由 applyTemplate 自己 preview，这里只作累计标记不改金额
  if (applyingTemplateFare.value) return
  if (form.fareRuleId) await applySuggestedFare(false)
}

/** force=true 时写入建议价；baseFen 优先作原价，否则用当前金额，再否则用规则内原价 */
async function applySuggestedFare(force = false, baseFen?: number) {
  if (!form.fareRuleId) return
  farePreviewLoading.value = true
  fareHint.value = ''
  try {
    let base = baseFen
    if (base == null || base <= 0) {
      const curFen = amountYuan.value ? yuanToFen(amountYuan.value) : 0
      if (curFen > 0) base = curFen
    }
    const body: Record<string, number> = {}
    if (base && base > 0) body.baseFen = base
    const { data } = await http.post(`/fare-rules/${form.fareRuleId}/preview`, body)
    fareHint.value = data.reason || ''
    // 选规则只刷新说明，不改金额；仅「填入建议价」/套用模板时写入
    if (!force) return
    const suggested = Number(data.amountFen) || 0
    const isFree = !!data.applied?.free
    if (suggested > 0 || isFree) {
      amountYuan.value = fenToYuan(suggested)
    } else if (base && base > 0 && !String(amountYuan.value || '').trim()) {
      // preview 未吃到原价时，至少保留原价，避免写成 0
      amountYuan.value = fenToYuan(base)
    }
  } catch (e: any) {
    fareHint.value = e.message || t('txEdit.previewFailed')
  } finally {
    farePreviewLoading.value = false
  }
}

async function createCategory(name: string) {
  if (form.type === 'transfer') throw new Error(t('txEdit.transferNoCategory'))
  const cat = await categories.create({ name, kind: form.type, icon: 'mdi-shape', sort: 0 })
  form.categoryId = cat.id
}

async function createTag(name: string) {
  const tag = await tags.create({ name, color: '#1b7f5a', sort: 0 })
  if (!form.tagIds.includes(tag.id)) form.tagIds = [...form.tagIds, tag.id]
}

async function onSave() {
  error.value = ''
  const amount = yuanToFen(amountYuan.value || '0')
  if (amount < 0) { error.value = t('txEdit.invalidAmount'); return }
  if (amount === 0 && !(fareEnabled.value && form.fareRuleId)) {
    error.value = t('txEdit.enterAmount')
    return
  }
  if (!form.accountId) { error.value = t('txEdit.selectAccount'); return }
  const useFx = fxEnabled.value && form.type !== 'transfer'
  const fxCode = fxCurrency.value.trim().toUpperCase()
  const fxFen = useFx ? yuanToFen(fxAmountYuan.value || '0') : 0
  if (useFx && !/^[A-Z]{3,8}$/.test(fxCode)) { error.value = t('txEdit.fxCurrencyRequired'); return }
  if (useFx && fxFen <= 0) { error.value = t('txEdit.fxAmountRequired'); return }
  saving.value = true
  try {
    const payload = {
      amountPending: wasPending.value && keepPending.value,
      foreignCurrency: useFx ? fxCode : '',
      foreignAmount: fxFen,
      voucherStatus: (form.type === 'expense' && voucherOn.value ? voucherStatus.value || 'pending' : '') as VoucherStatus,
      voucherExpireOn: form.type === 'expense' && voucherOn.value ? voucherExpireOn.value : '',
      type: form.type,
      amount,
      accountId: form.accountId,
      toAccountId: form.type === 'transfer' ? form.toAccountId : null,
      categoryId: form.type !== 'transfer' ? form.categoryId : null,
      tagIds: form.tagIds,
      remark: form.remark,
      geoMode: form.geoLng != null ? (form.geoMode || 'point') : '',
      geoLng: form.geoLng,
      geoLat: form.geoLat,
      geoName: form.geoName,
      geoEndLng: form.geoMode === 'route' ? form.geoEndLng : null,
      geoEndLat: form.geoMode === 'route' ? form.geoEndLat : null,
      geoEndName: form.geoMode === 'route' ? form.geoEndName : '',
      fareRuleId: fareEnabled.value && form.fareRuleId ? form.fareRuleId : null,
      installmentPeriods: showCreditExtras.value ? installmentPeriods.value : 1,
      interestFen: showCreditExtras.value ? yuanToFen(interestYuan.value || '0') : 0,
      installmentDeferCycles: isCreditExpense.value && installmentPeriods.value > 1 ? form.installmentDeferCycles : 0,
      happenedAt: dayjs(happenedLocal.value).toISOString(),
      attachmentIds: attachmentIds.value,
    }
    if (isEdit.value) await tx.update(editId.value, payload)
    else await tx.create(payload)
    await router.replace('/')
  } catch (e: any) {
    error.value = e.message || t('txEdit.saveFailed')
  } finally {
    saving.value = false
  }
}

async function onDelete() {
  if (!confirm(t('txEdit.confirmDelete'))) return
  await tx.remove(editId.value)
  await router.replace('/')
}
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.toolbar h1 { font-size: 1.15rem; margin: 0; }
.w100 { width: 100%; }
.flex1 { flex: 1; min-width: 0; }
.ai-row { display: flex; gap: 8px; }
.amount-field :deep(input) {
  font-size: 2rem;
  font-weight: 700;
  font-family: var(--font-display);
  font-variant-numeric: tabular-nums;
}
.tpls {
  display: flex;
  flex-wrap: nowrap;
  gap: 6px;
  overflow-x: auto;
  overscroll-behavior-x: contain;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  padding: 2px 0;
}
.tpls::-webkit-scrollbar { display: none; }
.tpl-chip { flex-shrink: 0; }
@media (min-width: 960px) {
  .tpls { flex-wrap: wrap; overflow: visible; }
}

/* 平板 / 窄桌面：单列但收窄居中，避免表单横向拉满 */
@media (min-width: 600px) and (max-width: 959px) {
  .te-page { max-width: 760px; }
}
@media (min-width: 960px) and (max-width: 1199px) {
  .te-page { max-width: calc(760px + var(--nav-rail-width)); }
}

/* 宽屏：主表单在左，附件 / 地点在右 */
@media (min-width: 1200px) {
  .te-page { max-width: calc(1160px + var(--nav-rail-width)); }
  .te-layout {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(300px, 360px);
    gap: 24px;
    align-items: start;
  }
}

/* 桌面：保存按钮吸底右对齐 */
@media (min-width: 960px) {
  .save-bar {
    position: sticky;
    bottom: 0;
    z-index: 5;
    display: flex;
    justify-content: flex-end;
    margin-top: 8px;
    padding: 14px 0 18px;
    background: linear-gradient(180deg, transparent, var(--bg) 40%);
  }
  .save-bar .hero-cta { width: 280px !important; }
}

@media (max-width: 959px) {
  .save-bar {
    position: sticky;
    bottom: 0;
    z-index: 5;
    margin: 0 calc(-1 * max(var(--content-pad-x), var(--sar))) -16px;
    padding: 10px max(var(--content-pad-x), var(--sar)) calc(10px + var(--sab));
    background: linear-gradient(180deg, transparent, var(--bg) 30%);
  }
}
.pending-box {
  padding: 10px 12px 4px;
  border-radius: 12px;
  background: color-mix(in srgb, #d99a2b 14%, var(--surface-solid, #fff));
  border: 1px solid color-mix(in srgb, #d99a2b 40%, transparent);
}
.pending-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 700;
  font-size: 0.92rem;
  color: #8a5a10;
}
.pending-tip {
  margin: 4px 0 0;
  font-size: 0.78rem;
  line-height: 1.45;
  color: var(--muted);
  overflow-wrap: anywhere;
}
[data-theme='dark'] .pending-title { color: #f0c36a; }
.voucher-head { display: flex; align-items: center; justify-content: space-between; }
.voucher-head :deep(.v-switch) { flex: 0 0 auto; }
.voucher-ops { display: flex; gap: 8px; margin-top: 8px; }
.refund-link {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 4px 6px 10px;
  border-radius: 10px;
  font-size: 0.8rem;
  color: #2b6aa3;
  background: color-mix(in srgb, #3b82c4 12%, transparent);
}
[data-theme='dark'] .refund-link { color: #9cc7ef; }
.flow-sub {
  display: flex;
  justify-content: stretch;
}
.flow-kind {
  width: 100%;
  border-radius: 12px;
  overflow: hidden;
}
.mode-hint {
  margin: 0;
  font-size: 0.8rem;
  line-height: 1.45;
  color: var(--muted);
  overflow-wrap: anywhere;
  word-break: break-word;
}
</style>
