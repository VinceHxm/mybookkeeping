<template>
  <div class="page te-page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" @click="$router.back()" />
      <h1>{{ isEdit ? '编辑流水' : '记一笔' }}</h1>
      <v-btn v-if="isEdit" icon="mdi-delete" variant="text" color="error" @click="onDelete" />
      <span v-else style="width:40px" />
    </div>

    <div class="te-layout">
    <div class="te-main">

    <div v-if="!isEdit" class="ai-row mb-3">
      <v-btn variant="tonal" color="secondary" class="flex1" to="/transactions/recognize">
        <v-icon start>mdi-robot</v-icon>
        文字识别
      </v-btn>
      <v-btn variant="tonal" color="secondary" class="flex1" @click="showAI = true">
        <v-icon start>mdi-camera</v-icon>
        票据识图
      </v-btn>
    </div>

    <div v-if="!isEdit && templates.length" class="tpls mb-3">
      <v-chip
        v-for="t in templates"
        :key="t.id"
        class="tpl-chip"
        color="primary"
        variant="outlined"
        @click="applyTemplate(t)"
      >{{ t.name }}</v-chip>
    </div>

    <v-btn-toggle v-model="uiMode" mandatory color="primary" class="mb-2 w100" divided>
      <v-btn value="flow" class="flex1">收支</v-btn>
      <v-btn value="transfer" class="flex1">转账</v-btn>
      <v-btn value="repay" class="flex1">还款</v-btn>
    </v-btn-toggle>

    <div v-if="uiMode === 'flow'" class="flow-sub mb-3">
      <v-btn-toggle v-model="flowKind" mandatory color="primary" density="comfortable" class="flow-kind" divided>
        <v-btn value="expense" class="flex1" size="small">支出</v-btn>
        <v-btn value="income" class="flex1" size="small">收入</v-btn>
      </v-btn-toggle>
    </div>
    <div v-if="!isEdit && uiMode === 'flow' && flowKind === 'income'" class="mb-3">
      <v-btn variant="tonal" color="info" block @click="refundOpen = true">
        <v-icon start>mdi-ticket-percent-outline</v-icon>
        从待核销团购券退款
      </v-btn>
    </div>
    <p v-else-if="uiMode === 'repay'" class="mode-hint mb-3">
      信用账户还款写这里：从银行卡 / 支付宝等转入信用卡、花呗等，<strong>不计入本月支出</strong>。
    </p>
    <p v-else class="mode-hint mb-3">账户之间挪钱（还信用卡 / 花呗请用「还款」）。</p>

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
        金额待确认
      </div>
      <p class="pending-tip">
        这笔由周期记账生成，当前金额是预估。等信用卡入账后，把下方金额改成账单上的实际人民币金额再保存，即完成确认。
      </p>
      <v-checkbox
        v-model="keepPending"
        label="账单还没出，保存后仍保持待确认"
        color="primary"
        density="compact"
        hide-details
      />
    </div>

    <v-text-field
      v-model="amountYuan"
      :label="wasPending && keepPending ? '预估金额（元）' : '金额（元）'"
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
        <h3 class="form-section__title">明细</h3>
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
          :items="transferToOptions"
          item-title="name"
          item-value="id"
          :label="repayActive ? '信用卡账户' : '转入账户'"
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
            label="分类"
            variant="outlined"
            class="flex1"
            hide-details
          />
          <v-btn icon="mdi-plus" variant="tonal" color="primary" title="新增分类" @click="showQuickCategory = true" />
        </div>

        <div class="form-field-row field">
          <v-select
            v-model="form.tagIds"
            :items="tags.list"
            item-title="name"
            item-value="id"
            label="标签"
            multiple
            chips
            closable-chips
            variant="outlined"
            class="flex1"
            hide-details
          />
          <v-btn icon="mdi-plus" variant="tonal" color="primary" title="新增标签" @click="showQuickTag = true" />
        </div>

        <v-text-field v-model="form.remark" label="备注" variant="outlined" hide-details class="field" />
        <NativeDateField v-model="happenedLocal" label="时间" type="datetime-local" variant="outlined" hide-details class="field field--last" />
      </div>
    </section>

    <section v-if="form.type === 'expense'" class="form-section mb-3">
      <header class="form-section__head voucher-head">
        <h3 class="form-section__title">
          团购券
          <v-chip
            v-if="isEdit && voucherStatus"
            size="x-small"
            :color="voucherStatusColor[voucherStatus]"
            variant="tonal"
            class="ml-1"
          >{{ voucherStatusLabel[voucherStatus] }}</v-chip>
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
          label="过期日（可选）"
          variant="outlined"
          hide-details
          class="field"
        />
        <p v-if="voucherStatus === 'refunded'" class="form-section__tip">
          这张券已退款。要撤销，删除对应的退款流水即可恢复为待核销。
        </p>
        <p v-else class="form-section__tip">
          先买后用的券打开此开关，会标记为待核销；用掉后点「一键核销」，过期自动退款时点「退款」直接生成退款收入。
        </p>
        <div v-if="isEdit && savedVoucherStatus && voucherStatus !== 'refunded'" class="voucher-ops">
          <v-btn
            v-if="voucherStatus === 'pending'"
            size="small"
            variant="tonal"
            color="success"
            :loading="voucherBusy"
            @click="toggleUsed('used')"
          >一键核销</v-btn>
          <v-btn
            v-else
            size="small"
            variant="text"
            :loading="voucherBusy"
            @click="toggleUsed('pending')"
          >撤销核销</v-btn>
          <v-btn size="small" variant="tonal" color="info" @click="refundOpen = true">退款</v-btn>
        </div>
      </div>
    </section>

    <div v-if="refundOfId" class="refund-link mb-3">
      <v-icon size="16">mdi-ticket-percent-outline</v-icon>
      <span class="flex1">这是一笔团购券退款，删除后原券会恢复为待核销。</span>
      <v-btn size="small" variant="text" color="primary" :to="`/transactions/${refundOfId}`">查看原券</v-btn>
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
      <v-btn color="primary" block size="large" class="hero-cta" :loading="saving" @click="onSave">保存</v-btn>
    </div>

    <QuickNameDialog
      v-model="showQuickCategory"
      title="新增分类"
      label="分类名称"
      :hint="`将创建为当前「${form.type === 'income' ? '收入' : '支出'}」分类，可稍后在分类管理里调整层级。`"
      :save="createCategory"
    />
    <QuickNameDialog v-model="showQuickTag" title="新增标签" label="标签名称" :save="createTag" />
    <ReceiptRecognizeDialog v-model="showAI" @recognized="applyRecognized" @attached="onReceiptAttached" />
    <VoucherRefundDialog v-model="refundOpen" :voucher="refundVoucher" @refunded="onRefunded" />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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
    const t = await tx.setVoucherStatus(editId.value, status)
    voucherStatus.value = t.voucherStatus || ''
    savedVoucherStatus.value = voucherStatus.value
    editingTx.value = t
  } catch (e: any) {
    error.value = e?.message || '操作失败'
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
      title: r.city ? `${r.name}（${r.city}）` : r.name,
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

watch(() => form.type, (t, prev) => {
  if (syncingUiMode || t === prev) return
  form.categoryId = undefined
  if (!(isRepayMode.value && t === 'transfer')) {
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
      if (form.remark === '信用卡还款') form.remark = ''
      repay.reset()
      if (useDefaultAccount) applyDefaultAccount(flowKind.value)
    } else if (mode === 'transfer') {
      form.type = 'transfer'
      form.categoryId = undefined
      if (form.remark === '信用卡还款') form.remark = ''
      // 普通转账不强制信用账户
      if (form.toAccountId && accounts.list.some((a) => a.id === form.toAccountId && a.type === 'credit')) {
        form.toAccountId = undefined
      }
      repay.reset()
      if (useDefaultAccount) applyDefaultAccount('transfer')
    } else {
      form.type = 'transfer'
      form.categoryId = undefined
      form.remark = form.remark || '信用卡还款'
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

function applyEditTransaction(t: Awaited<ReturnType<typeof tx.get>>) {
  if (t.type === 'revalue') return
  syncingUiMode = true
  form.type = t.type
  if (t.type === 'expense' || t.type === 'income') {
    uiMode.value = 'flow'
    flowKind.value = t.type
  } else {
    const toCredit = t.toAccountId && accounts.list.some((a) => a.id === t.toAccountId && a.type === 'credit')
    uiMode.value = toCredit ? 'repay' : 'transfer'
  }
  form.accountId = t.accountId
  form.toAccountId = t.toAccountId || undefined
  form.categoryId = t.categoryId || undefined
  form.tagIds = (t.tags || []).map((x) => x.id)
  form.remark = t.remark || ''
  form.geoMode = (t.geoMode as 'point' | 'route') || (t.geoLng != null ? 'point' : '')
  form.geoLng = t.geoLng ?? null
  form.geoLat = t.geoLat ?? null
  form.geoName = t.geoName || ''
  form.geoEndLng = t.geoEndLng ?? null
  form.geoEndLat = t.geoEndLat ?? null
  form.geoEndName = t.geoEndName || ''
  form.fareRuleId = t.fareRuleId || null
  form.installmentPeriods = t.installmentPeriods || 1
  form.installmentDeferCycles = t.installmentDeferCycles || 0
  interestYuan.value = fenToYuan(t.interestFen || 0)
  fareEnabled.value = !!t.fareRuleId
  amountYuan.value = fenToYuan(t.amount)
  fxEnabled.value = !!(t.foreignCurrency && t.foreignAmount)
  fxCurrency.value = t.foreignCurrency || 'USD'
  fxAmountYuan.value = t.foreignAmount ? fenToYuan(t.foreignAmount) : ''
  wasPending.value = !!t.amountPending
  keepPending.value = false
  editingTx.value = t
  voucherStatus.value = t.voucherStatus || ''
  savedVoucherStatus.value = voucherStatus.value
  voucherOn.value = !!t.voucherStatus
  voucherExpireOn.value = t.voucherExpireOn || ''
  refundOfId.value = t.refundOfId || null
  happenedLocal.value = dayjs(t.happenedAt).format('YYYY-MM-DDTHH:mm')
  existingAttachments.value = (t.attachments || []).map((a) => ({ id: a.id, url: a.url }))
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
    const [t] = await Promise.all([txP, storesP])
    if (t.type === 'revalue') {
      void router.replace(`/accounts/${t.accountId}/ledger`)
      return
    }
    applyEditTransaction(t)
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
    const t = templates.value.find((x) => x.id === tplId)
    if (t) await applyTemplate(t)
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

async function applyTemplate(t: any) {
  syncingUiMode = true
  form.type = t.type
  if (t.type === 'expense' || t.type === 'income') {
    uiMode.value = 'flow'
    flowKind.value = t.type
  } else {
    const toCredit = t.toAccountId && accounts.list.some((a) => a.id === t.toAccountId && a.type === 'credit')
    uiMode.value = toCredit ? 'repay' : 'transfer'
  }
  syncingUiMode = false
  suppressDefaultAccountOnce()
  if (t.accountId) form.accountId = t.accountId
  form.toAccountId = t.toAccountId || undefined
  form.categoryId = t.categoryId || undefined
  form.tagIds = t.tagIds || []
  form.remark = t.remark?.trim() || t.name || ''
  if (t.geoLng != null && t.geoLat != null) {
    form.geoMode = (t.geoMode as 'point' | 'route') || (t.geoEndLng != null ? 'route' : 'point')
    form.geoLng = t.geoLng
    form.geoLat = t.geoLat
    form.geoName = t.geoName || ''
    if (form.geoMode === 'route') {
      form.geoEndLng = t.geoEndLng ?? null
      form.geoEndLat = t.geoEndLat ?? null
      form.geoEndName = t.geoEndName || ''
    } else {
      form.geoEndLng = null
      form.geoEndLat = null
      form.geoEndName = ''
    }
  }

  // 先写入模板金额，避免选规则触发 preview 时原价为空被写成 0
  const tplAmount = Number(t.amount) || 0
  if (tplAmount > 0) amountYuan.value = fenToYuan(tplAmount)

  const ruleId = t.fareRuleId || t.fareRule?.id
  const rule = fareRules.value.find((r) => r.id === ruleId) || t.fareRule
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
    fareHint.value = rule.city ? `地市不匹配（规则：${rule.city}），未自动启用` : ''
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
    fareHint.value = e.message || '预览失败'
  } finally {
    farePreviewLoading.value = false
  }
}

async function createCategory(name: string) {
  if (form.type === 'transfer') throw new Error('转账无需分类')
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
  if (amount < 0) { error.value = '金额无效'; return }
  if (amount === 0 && !(fareEnabled.value && form.fareRuleId)) {
    error.value = '请输入金额'
    return
  }
  if (!form.accountId) { error.value = '请选择账户'; return }
  const useFx = fxEnabled.value && form.type !== 'transfer'
  const fxCode = fxCurrency.value.trim().toUpperCase()
  const fxFen = useFx ? yuanToFen(fxAmountYuan.value || '0') : 0
  if (useFx && !/^[A-Z]{3,8}$/.test(fxCode)) { error.value = '请选择或填写外币币种（如 USD）'; return }
  if (useFx && fxFen <= 0) { error.value = '请填写外币金额'; return }
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
    error.value = e.message || '保存失败'
  } finally {
    saving.value = false
  }
}

async function onDelete() {
  if (!confirm('确定删除这笔流水？')) return
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
