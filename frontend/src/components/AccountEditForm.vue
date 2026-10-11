<template>
  <div class="form-sections">
    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('accountEdit.basic') }}</h3>
      </header>
      <div class="form-section__body">
        <v-text-field v-model="form.name" :label="t('accountEdit.name')" variant="outlined" hide-details class="field" />
        <v-select
          v-model="form.type"
          :items="typeItems"
          :label="t('accountEdit.type')"
          variant="outlined"
          hide-details
          class="field"
        />
        <v-text-field
          v-model.number="form.sort"
          :label="t('accountEdit.sort')"
          type="number"
          variant="outlined"
          hide-details
          class="field field--last"
        />
      </div>
    </section>

    <section v-if="form.type === 'invest'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('accountEdit.platform') }}</h3>
      </header>
      <div class="form-section__body">
        <v-text-field
          v-model="form.institution"
          :label="t('accountEdit.platformLabel')"
          variant="outlined"
          hide-details
          class="field field--last"
          :placeholder="t('accountEdit.platformPlaceholder')"
        />
      </div>
    </section>

    <section v-if="form.type === 'bank' || form.type === 'credit'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('accountEdit.card') }}</h3>
      </header>
      <div class="form-section__body">
        <v-text-field
          v-model="form.institution"
          :label="form.type === 'credit' ? t('accountEdit.issuer') : t('accountEdit.bankBranch')"
          variant="outlined"
          hide-details
          class="field"
          :placeholder="t('accountEdit.bankPlaceholder')"
        />
        <v-text-field
          v-model="form.cardNo"
          :label="t('accountEdit.cardNo')"
          variant="outlined"
          hide-details
          class="field"
          inputmode="numeric"
          autocomplete="off"
          :append-inner-icon="accountId && originalCardMasked ? (secretsRevealed ? 'mdi-eye-off-outline' : 'mdi-eye-lock-outline') : undefined"
          @click:append-inner="toggleFormReveal"
          @focus="onSecretFocus('cardNo')"
          @blur="onSecretBlur('cardNo')"
        />
        <v-text-field
          v-model="form.holderName"
          :label="t('accountEdit.holderName')"
          variant="outlined"
          hide-details
          class="field field--last"
          :append-inner-icon="accountId && originalHolderMasked ? (secretsRevealed ? 'mdi-eye-off-outline' : 'mdi-eye-lock-outline') : undefined"
          @click:append-inner="toggleFormReveal"
          @focus="onSecretFocus('holderName')"
          @blur="onSecretBlur('holderName')"
        />
        <p class="form-section__tip">
          {{ t('accountEdit.cardTip') }}
        </p>
      </div>
    </section>

    <section v-if="form.type === 'credit'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('accountEdit.limit') }}</h3>
      </header>
      <div class="form-section__body">
        <v-text-field
          v-model="creditLimitYuan"
          :label="t('accountEdit.creditLimit')"
          type="number"
          inputmode="decimal"
          variant="outlined"
          hide-details
          class="field"
          @update:model-value="onCreditLimitChange"
        />
        <v-text-field
          v-model="availableCreditYuan"
          :label="t('accountEdit.availableCredit')"
          type="number"
          inputmode="decimal"
          variant="outlined"
          hide-details
          class="field"
          @update:model-value="onAvailableCreditChange"
        />
        <v-text-field
          v-model="usedCreditYuan"
          :label="accountId ? t('accountEdit.usedCredit') : t('accountEdit.initialUsedCredit')"
          type="number"
          inputmode="decimal"
          variant="outlined"
          hide-details
          class="field field--last"
          @update:model-value="onUsedCreditChange"
        />
        <p class="form-section__tip">
          {{ t('accountEdit.limitTip') }}
        </p>
      </div>
    </section>

    <section v-if="form.type === 'credit'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('accountEdit.statement') }}</h3>
      </header>
      <div class="form-section__body">
        <div class="form-section__group">
          <v-text-field
            v-model="billedYuan"
            :label="t('accountEdit.billed')"
            type="number"
            inputmode="decimal"
            variant="outlined"
            hide-details
            class="field"
            @update:model-value="onBilledChange"
          />
          <v-text-field
            :model-value="unbilledYuan"
            :label="t('accountEdit.unbilled')"
            type="number"
            variant="outlined"
            readonly
            hide-details
            class="field field--last"
          />
          <i18n-t keypath="accountEdit.billedTip" tag="p" class="form-section__tip">
            <template #zero><strong>0</strong></template>
          </i18n-t>
        </div>
        <div class="form-section__group">
          <v-text-field
            v-model.number="form.billingDay"
            :label="t('accountEdit.billingDay')"
            type="number"
            variant="outlined"
            hide-details
            class="field"
          />
          <v-text-field
            v-model.number="form.paymentDueDay"
            :label="t('accountEdit.dueDay')"
            type="number"
            variant="outlined"
            hide-details
            class="field field--last"
          />
          <p class="form-section__tip">
            {{ t('accountEdit.billingTip') }}
          </p>
        </div>
      </div>
    </section>

    <section v-if="form.type === 'credit'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('accountEdit.plans') }}</h3>
      </header>
      <div class="form-section__body">
        <p class="form-section__tip form-section__tip--alone">
          <i18n-t keypath="accountEdit.plansTip1" tag="span">
            <template #unbilled><strong>{{ t('accountEdit.plansTipUnbilled') }}</strong></template>
          </i18n-t>
          {{ t('accountEdit.plansTip2') }}
          <i18n-t keypath="accountEdit.plansTip3" tag="span">
            <template #not><strong>{{ t('accountEdit.plansTipNot') }}</strong></template>
          </i18n-t>
          {{ t('accountEdit.plansTip4') }}
        </p>
        <div
          v-for="(row, idx) in planRows"
          :key="row.key"
          class="plan-row"
        >
          <v-text-field
            v-model="row.name"
            :label="t('accountEdit.planName')"
            variant="outlined"
            hide-details
            density="compact"
            class="field"
            :placeholder="t('accountEdit.planNamePlaceholder')"
          />
          <div class="plan-row__grid">
            <v-text-field
              v-model="row.principalYuan"
              :label="t('accountEdit.planPrincipal')"
              type="number"
              inputmode="decimal"
              variant="outlined"
              hide-details
              density="compact"
              prefix="¥"
              class="field"
            />
            <v-text-field
              v-model.number="row.periods"
              :label="t('accountEdit.planPeriods')"
              type="number"
              min="1"
              :max="row.firstDueMode === 'current' ? 59 : 60"
              variant="outlined"
              hide-details
              density="compact"
              class="field"
            />
            <v-text-field
              v-model="row.interestYuan"
              :label="t('accountEdit.planInterest')"
              type="number"
              inputmode="decimal"
              variant="outlined"
              hide-details
              density="compact"
              prefix="¥"
              class="field"
            />
            <v-select
              v-model="row.firstDueMode"
              :items="firstDueModeItems"
              :label="t('accountEdit.planFirstDue')"
              variant="outlined"
              hide-details
              density="compact"
              class="field field--last"
            />
          </div>
          <div class="plan-row__actions">
            <span v-if="planShareHint(row)" class="plan-share">{{ t('accountEdit.planShare', { amount: planShareHint(row) }) }}</span>
            <v-btn size="small" variant="text" color="error" @click="removePlanRow(idx)">{{ t('accountEdit.remove') }}</v-btn>
          </div>
        </div>
        <v-btn size="small" variant="tonal" color="primary" class="mt-1" @click="addPlanRow">
          {{ t('accountEdit.addPlan') }}
        </v-btn>
        <div v-if="planSummary.futurePrincipal > 0 || planSummary.includedInBilled > 0 || planRows.length" class="plan-summary">
          <div v-if="planSummary.includedInBilled > 0">
            {{ t('accountEdit.sumIncluded', { amount: fenToYuan(planSummary.includedInBilled) }) }}
          </div>
          <div v-if="planSummary.futurePrincipal > 0">
            {{ t('accountEdit.sumFuture', { amount: fenToYuan(planSummary.futurePrincipal) }) }}
            <template v-if="planSummary.nextShare > 0">
              {{ t('accountEdit.sumNextShare', { amount: fenToYuan(planSummary.nextShare) }) }}
            </template>
          </div>
          <div>
            {{ t('accountEdit.sumNonInstallment', { amount: fenToYuan(planSummary.nonInstallmentUnbilled) }) }}
            <template v-if="planSummary.nextShare > 0">
              {{ t('accountEdit.sumNextTotal', { amount: fenToYuan(planSummary.nonInstallmentUnbilled + planSummary.nextShare) }) }}
            </template>
          </div>
          <div v-if="planSummary.warnBilledZeroButIncluded" class="plan-warn">
            {{ t('accountEdit.warnBilledZero') }}
          </div>
          <div v-if="planSummary.warnAccruingAsBilled" class="plan-warn">
            {{ t('accountEdit.warnAccruing') }}
          </div>
          <div v-if="planSummary.overUnbilled" class="plan-warn">{{ t('accountEdit.warnOverUnbilled') }}</div>
          <div v-if="!form.billingDay" class="plan-warn">{{ t('accountEdit.warnNoBillingDay') }}</div>
        </div>
      </div>
    </section>

    <section v-else-if="form.type !== 'credit'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ form.type === 'invest' ? t('accountEdit.marketValue') : t('accountEdit.balance') }}</h3>
      </header>
      <div class="form-section__body">
        <v-text-field
          v-model="balanceYuan"
          :label="form.type === 'invest'
            ? (accountId ? t('accountEdit.marketValueYuan') : t('accountEdit.currentMarketValueYuan'))
            : (accountId ? t('accountEdit.balanceYuan') : t('accountEdit.initialBalanceYuan'))"
          type="number"
          inputmode="decimal"
          variant="outlined"
          hide-details
          class="field field--last"
        />
        <p v-if="form.type === 'invest'" class="form-section__tip">
          {{ t('accountEdit.investTip') }}
        </p>
        <p v-else-if="accountId" class="form-section__tip">{{ t('accountEdit.balanceTip') }}</p>
      </div>
    </section>

    <section v-if="accountId" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('accountEdit.defaults') }}</h3>
      </header>
      <div class="form-section__body">
        <v-switch
          v-model="asDefaultExpense"
          :label="t('accountEdit.defaultExpense')"
          color="primary"
          hide-details
          density="compact"
          class="field"
          :disabled="form.archived"
        />
        <v-switch
          v-if="form.type !== 'credit'"
          v-model="asDefaultRepay"
          :label="t('accountEdit.defaultRepay')"
          color="primary"
          hide-details
          density="compact"
          class="field field--last"
          :disabled="form.archived"
        />
        <p v-else class="form-section__tip">{{ t('accountEdit.creditNoRepay') }}</p>
        <p class="form-section__tip">{{ t('accountEdit.defaultsTip') }}</p>
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('accountEdit.storage') }}</h3>
      </header>
      <div class="form-section__body">
        <v-textarea
          v-model="form.storageNote"
          :label="t('accountEdit.storageNote')"
          variant="outlined"
          hide-details
          class="field"
          rows="2"
          auto-grow
          :placeholder="t('accountEdit.storagePlaceholder')"
        />
        <AttachmentUpload
          v-model="attachmentIds"
          :existing="formAttachments"
          :title="t('accountEdit.storagePhotos')"
          :subtitle="t('accountEdit.storagePhotosSub')"
        />
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('accountEdit.other') }}</h3>
      </header>
      <div class="form-section__body">
        <v-text-field
          v-model="form.remark"
          :label="t('accountEdit.remark')"
          variant="outlined"
          hide-details
          :class="accountId ? 'field' : 'field field--last'"
        />
        <template v-if="accountId">
          <v-switch v-model="form.archived" :label="t('accountEdit.archive')" hide-details class="field" />
          <v-btn class="mt-1" variant="tonal" size="small" :loading="reconciling" @click="markReconciled">
            {{ t('accountEdit.markReconciled') }}
          </v-btn>
        </template>
      </div>
    </section>

    <div v-if="accountId" class="danger-zone">
      <div class="danger-title">{{ t('accountEdit.danger') }}</div>
      <p class="danger-tip">{{ t('accountEdit.dangerTip') }}</p>
      <v-btn color="error" variant="outlined" block @click="emit('delete')">{{ t('accountEdit.deleteAccount') }}</v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import AttachmentUpload from './AttachmentUpload.vue'
import {
  ACCOUNT_TYPES,
  accountTypeLabel,
  useAccountStore,
  creditUsedFen,
  creditBilledFenOf,
  type Account,
  type CreditInstallmentPlan,
} from '../stores/account'
import { useAuthStore } from '../stores/auth'
import { fenToYuan, yuanToFen } from '../utils/money'
import { looksMaskedSecret, secretFieldForSave } from '../utils/mask'
import { confirmWithPassword } from '../composables/usePasswordConfirm'

const props = defineProps<{
  accountId?: number | null
  /** 列表页已持有的账户，可避免再拉一次 */
  initial?: Account | null
}>()

const emit = defineEmits<{
  saved: []
  delete: []
  reconciled: []
}>()

const { t } = useI18n()
const accounts = useAccountStore()
const auth = useAuthStore()
const saving = ref(false)
const reconciling = ref(false)
const revealLoading = ref(false)
const secretsRevealed = ref(false)
const asDefaultExpense = ref(false)
const asDefaultRepay = ref(false)
/** 打开表单时的脱敏快照，用于判断「未改动」 */
const originalCardMasked = ref('')
const originalHolderMasked = ref('')
const balanceYuan = ref('0')
const usedCreditYuan = ref('0')
const availableCreditYuan = ref('0')
const billedYuan = ref('0')
const creditLimitYuan = ref('0')
const attachmentIds = ref<number[]>([])
const formAttachments = ref<{ id: number; url: string }[]>([])

type PlanRow = {
  key: number
  name: string
  principalYuan: string
  periods: number
  interestYuan: string
  firstDueMode: 'current' | 'next'
  /** 库内首期出账日晚于下次出账日时保留原值，避免编辑时被提前 */
  fixedFirstDueOn?: string
}

let planKeySeq = 1
const planRows = ref<PlanRow[]>([])
const firstDueModeItems = computed(() => [
  { title: t('accountEdit.planFirstNext'), value: 'next' },
  { title: t('accountEdit.planFirstCurrent'), value: 'current' },
])

const form = reactive({
  name: '',
  type: 'cash',
  sort: 0,
  archived: false,
  billingDay: 0,
  paymentDueDay: 0,
  institution: '',
  cardNo: '',
  holderName: '',
  storageNote: '',
  remark: '',
})

const typeItems = computed(() =>
  ACCOUNT_TYPES.map((v) => ({
    title: v === 'invest' ? t('accountEdit.typeInvestLong') : accountTypeLabel(v),
    value: v,
  })),
)

const unbilledYuan = computed(() => {
  const used = yuanToFen(usedCreditYuan.value)
  let billed = yuanToFen(billedYuan.value)
  if (billed < 0) billed = 0
  if (billed > used) billed = used
  return fenToYuan(Math.max(0, used - billed))
})

function clampDay(year: number, month: number, day: number) {
  const d = Math.min(28, Math.max(1, day || 1))
  return dayjs(`${year}-${String(month).padStart(2, '0')}-${String(d).padStart(2, '0')}`)
}

/** 对齐后端 lastClosedStatementDate / nextStatementDate */
function creditStatementDates(billingDay: number, at = dayjs()) {
  const bd = Math.min(28, Math.max(1, billingDay || 1))
  const d = at.date()
  let stmt = clampDay(at.year(), at.month() + 1, bd)
  if (d < bd) {
    const prev = at.subtract(1, 'month')
    stmt = clampDay(prev.year(), prev.month() + 1, bd)
  }
  const next = stmt.add(1, 'month')
  const nextStmt = clampDay(next.year(), next.month() + 1, bd)
  return { stmt, nextStmt }
}

/** 对齐后端 splitAmountEvenly：余数并入首期 */
function splitEvenly(total: number, periods: number) {
  const n = Math.max(1, periods)
  const base = Math.floor(total / n)
  const out = Array.from({ length: n }, () => base)
  out[0] += total - base * n
  return out
}

/**
 * 库内计划按今天重新定位：已出账的期数剔除，表单只展示「尚未出账」的剩余本金/期数。
 * 库内「当期」计划含已出账那一期（本金、期数各多一期），表单里扣掉。
 */
function planRowFromStored(p: CreditInstallmentPlan, billingDay: number): PlanRow | null {
  const principal = Math.max(0, p.principalFen || 0)
  const periods = Math.max(1, p.periods || 1)
  const common = {
    key: planKeySeq++,
    name: p.name || '',
    interestYuan: fenToYuan(p.interestPerPeriodFen || 0),
  }
  if (!billingDay || !p.firstDueOn) {
    return { ...common, principalYuan: fenToYuan(principal), periods, firstDueMode: 'next' }
  }
  const { stmt, nextStmt } = creditStatementDates(billingDay)
  const fd = dayjs(p.firstDueOn)
  const firstDue = clampDay(fd.year(), fd.month() + 1, billingDay)
  if (firstDue.isAfter(stmt, 'day')) {
    return {
      ...common,
      principalYuan: fenToYuan(principal),
      periods,
      firstDueMode: 'next',
      fixedFirstDueOn: firstDue.isAfter(nextStmt, 'day') ? p.firstDueOn : undefined,
    }
  }
  // 最近一次出账对应第 k 期（0 起），之后的才是尚未出账
  const k = stmt.diff(firstDue, 'month')
  const left = periods - k - 1
  if (left <= 0) return null
  const rest = splitEvenly(principal, periods)
    .slice(k + 1)
    .reduce((s, x) => s + x, 0)
  return { ...common, principalYuan: fenToYuan(rest), periods: left, firstDueMode: 'current' }
}

function emptyPlanRow(): PlanRow {
  return {
    key: planKeySeq++,
    name: '',
    principalYuan: '',
    periods: 1,
    interestYuan: '0',
    firstDueMode: 'next',
  }
}

function addPlanRow() {
  planRows.value.push(emptyPlanRow())
}

function removePlanRow(idx: number) {
  planRows.value.splice(idx, 1)
}

function planShareHint(row: PlanRow) {
  const principal = yuanToFen(row.principalYuan)
  const periods = Math.max(1, Number(row.periods) || 1)
  if (principal <= 0) return ''
  return fenToYuan(Math.floor(principal / periods))
}

const planSummary = computed(() => {
  let includedInBilled = 0
  let futurePrincipal = 0
  let nextShare = 0
  let hasIncludedMode = false
  for (const row of planRows.value) {
    const principal = Math.max(0, yuanToFen(row.principalYuan))
    const periods = Math.max(1, Math.min(60, Number(row.periods) || 1))
    if (principal <= 0) continue
    const base = Math.floor(principal / periods)
    if (row.firstDueMode === 'next') {
      futurePrincipal += principal
      nextShare += base + (principal % periods)
    } else {
      // 输入只含尚未出账部分；已出账那期按每期均摊估算，存库时补回（余数落在已出账那期）
      hasIncludedMode = true
      includedInBilled += base
      futurePrincipal += base * periods
      nextShare += base
    }
  }
  const used = Math.max(0, yuanToFen(usedCreditYuan.value))
  let billed = yuanToFen(billedYuan.value)
  if (billed < 0) billed = 0
  if (billed > used) billed = used
  const unbilled = used - billed
  // 未出账里应覆盖「尚未出账的分期」；已含在已出账的首期不占用未出账
  const nonInstallmentUnbilled = Math.max(0, unbilled - futurePrincipal)
  return {
    includedInBilled,
    futurePrincipal,
    nextShare,
    nonInstallmentUnbilled,
    overUnbilled: futurePrincipal > unbilled && unbilled >= 0,
    warnBilledZeroButIncluded: billed === 0 && hasIncludedMode && includedInBilled > 0,
    // 已出账>0 且存在「已含在已出账」分期时，提醒别把「已累计」误当已出账（启发式：未出账≈后续分期）
    warnAccruingAsBilled:
      billed > 0 &&
      hasIncludedMode &&
      futurePrincipal > 0 &&
      Math.abs(unbilled - futurePrincipal) <= 1,
  }
})

function fillPlansFrom(a: Account | null | undefined) {
  const billingDay = a?.billingDay || 0
  planRows.value = (a?.installmentPlans || [])
    .map((p) => planRowFromStored(p, billingDay))
    .filter((r): r is PlanRow => r !== null)
}

function buildInstallmentPlansPayload(): CreditInstallmentPlan[] {
  const { stmt, nextStmt } = creditStatementDates(form.billingDay || 1)
  return planRows.value
    .map((row, i) => {
      const inputFen = Math.max(0, yuanToFen(row.principalYuan))
      if (inputFen <= 0 && !row.name.trim()) return null
      const common = {
        // 留空由后端填默认名
        name: row.name.trim(),
        interestPerPeriodFen: Math.max(0, yuanToFen(row.interestYuan)),
        sort: i,
      }
      if (row.firstDueMode === 'current') {
        // 库内首期 = 已出账那期：本金、期数各补一期，首期出账日为最近一次出账日
        const periods = Math.max(1, Math.min(59, Number(row.periods) || 1))
        return {
          ...common,
          principalFen: inputFen + Math.floor(inputFen / periods),
          periods: periods + 1,
          firstDueOn: stmt.format('YYYY-MM-DD'),
        } as CreditInstallmentPlan
      }
      return {
        ...common,
        principalFen: inputFen,
        periods: Math.max(1, Math.min(60, Number(row.periods) || 1)),
        firstDueOn: row.fixedFirstDueOn || nextStmt.format('YYYY-MM-DD'),
      } as CreditInstallmentPlan
    })
    .filter(Boolean) as CreditInstallmentPlan[]
}

function fillFrom(a: Account | null | undefined) {
  secretsRevealed.value = false
  if (!a) {
    form.name = ''
    form.type = 'cash'
    form.sort = 0
    form.archived = false
    form.billingDay = 0
    form.paymentDueDay = 0
    form.institution = ''
    form.cardNo = ''
    form.holderName = ''
    form.storageNote = ''
    form.remark = ''
    originalCardMasked.value = ''
    originalHolderMasked.value = ''
    balanceYuan.value = '0'
    usedCreditYuan.value = '0'
    availableCreditYuan.value = '0'
    billedYuan.value = '0'
    creditLimitYuan.value = '0'
    attachmentIds.value = []
    formAttachments.value = []
    planRows.value = []
    asDefaultExpense.value = false
    asDefaultRepay.value = false
    return
  }
  form.name = a.name
  form.type = a.type
  form.sort = a.sort
  form.archived = a.archived
  form.billingDay = a.billingDay || 0
  form.paymentDueDay = a.paymentDueDay || 0
  form.institution = a.institution || ''
  form.cardNo = a.cardNo || ''
  form.holderName = a.holderName || ''
  originalCardMasked.value = a.cardNo || ''
  originalHolderMasked.value = a.holderName || ''
  form.storageNote = a.storageNote || ''
  form.remark = a.remark || ''
  formAttachments.value = [...(a.attachments || [])]
  attachmentIds.value = (a.attachments || []).map((x) => x.id)
  balanceYuan.value = fenToYuan(a.balance)
  usedCreditYuan.value = fenToYuan(creditUsedFen(a))
  billedYuan.value = fenToYuan(creditBilledFenOf(a))
  creditLimitYuan.value = fenToYuan(a.creditLimit || 0)
  syncAvailableFromUsed()
  fillPlansFrom(a)
  asDefaultExpense.value = auth.profile?.defaultExpenseAccountId === a.id
  asDefaultRepay.value = a.type !== 'credit' && auth.profile?.defaultAccountId === a.id
}

async function loadAccount() {
  if (!auth.profile) {
    await auth.fetchMe().catch(() => {})
  }
  if (props.initial) {
    fillFrom(props.initial)
    return
  }
  if (!props.accountId) {
    fillFrom(null)
    return
  }
  await accounts.load(true)
  const a = accounts.list.find((x) => x.id === props.accountId)
  fillFrom(a || null)
}

watch(
  () => [props.accountId, props.initial] as const,
  () => {
    loadAccount()
  },
)

onMounted(loadAccount)

async function toggleFormReveal() {
  if (!props.accountId) return
  if (secretsRevealed.value) {
    // 收起：恢复脱敏展示，丢弃明文
    form.cardNo = originalCardMasked.value
    form.holderName = originalHolderMasked.value
    secretsRevealed.value = false
    return
  }
  if (revealLoading.value) return
  const id = props.accountId
  revealLoading.value = true
  try {
    const data = await confirmWithPassword(
      { title: t('pwdConfirm.verifyTitle'), message: t('pwdConfirm.revealMessage') },
      (pwd) => accounts.fetchSensitive(id, pwd),
    )
    if (!data) return
    form.cardNo = data.cardNo || ''
    form.holderName = data.holderName || ''
    secretsRevealed.value = true
  } finally {
    revealLoading.value = false
  }
}

/** 聚焦脱敏字段时清空，避免在 ··· 上续写导致误判 */
function onSecretFocus(field: 'cardNo' | 'holderName') {
  if (secretsRevealed.value) return
  if (field === 'cardNo' && looksMaskedSecret(form.cardNo)) form.cardNo = ''
  if (field === 'holderName' && looksMaskedSecret(form.holderName)) form.holderName = ''
}

/** 未输入新值就失焦：恢复脱敏，避免误清空库内原值 */
function onSecretBlur(field: 'cardNo' | 'holderName') {
  if (secretsRevealed.value) return
  if (field === 'cardNo' && !form.cardNo.trim() && originalCardMasked.value) {
    form.cardNo = originalCardMasked.value
  }
  if (field === 'holderName' && !form.holderName.trim() && originalHolderMasked.value) {
    form.holderName = originalHolderMasked.value
  }
}

function syncAvailableFromUsed() {
  const limit = yuanToFen(creditLimitYuan.value)
  const used = Math.max(0, yuanToFen(usedCreditYuan.value))
  availableCreditYuan.value = fenToYuan(Math.max(0, limit - used))
}

function clampBilledToUsed() {
  const used = Math.max(0, yuanToFen(usedCreditYuan.value))
  const billed = yuanToFen(billedYuan.value)
  if (billed > used) billedYuan.value = fenToYuan(used)
}

function onCreditLimitChange() {
  // 改额度：保留已用，重算可用（编辑已有账户改额度时不抹掉欠款）
  syncAvailableFromUsed()
}

function onAvailableCreditChange() {
  const limit = yuanToFen(creditLimitYuan.value)
  let available = yuanToFen(availableCreditYuan.value)
  if (available < 0) {
    available = 0
    availableCreditYuan.value = '0'
  }
  // 可用超过额度时按额度封顶，已用记 0
  if (available > limit) {
    available = Math.max(0, limit)
    availableCreditYuan.value = fenToYuan(available)
  }
  usedCreditYuan.value = fenToYuan(Math.max(0, limit - available))
  clampBilledToUsed()
}

function onUsedCreditChange() {
  if (yuanToFen(usedCreditYuan.value) < 0) usedCreditYuan.value = '0'
  syncAvailableFromUsed()
  clampBilledToUsed()
}

function onBilledChange() {
  const used = Math.max(0, yuanToFen(usedCreditYuan.value))
  const billed = yuanToFen(billedYuan.value)
  if (billed < 0) billedYuan.value = '0'
  else if (billed > used) billedYuan.value = fenToYuan(used)
}

async function save() {
  if (saving.value) return
  if (form.type === 'credit' && planRows.value.length && !form.billingDay) {
    alert(t('accountEdit.needBillingDay'))
    return
  }
  if (form.type === 'credit' && planSummary.value.overUnbilled) {
    if (!confirm(t('accountEdit.confirmOverUnbilled'))) return
  }
  saving.value = true
  try {
    const payload: Record<string, unknown> = {
      name: form.name,
      type: form.type,
      sort: form.sort,
      creditLimit: yuanToFen(creditLimitYuan.value),
      billingDay: form.billingDay || 0,
      paymentDueDay: form.paymentDueDay || 0,
      institution: form.institution,
      storageNote: form.storageNote,
      remark: form.remark,
      attachmentIds: attachmentIds.value,
    }
    // 新建：仅明文可写入；编辑：脱敏串不提交
    if (props.accountId) {
      const card = secretFieldForSave(form.cardNo, originalCardMasked.value)
      const holder = secretFieldForSave(form.holderName, originalHolderMasked.value)
      if (card !== undefined) payload.cardNo = card
      if (holder !== undefined) payload.holderName = holder
    } else {
      if (!looksMaskedSecret(form.cardNo) && form.cardNo.trim()) payload.cardNo = form.cardNo.trim()
      if (!looksMaskedSecret(form.holderName) && form.holderName.trim()) {
        payload.holderName = form.holderName.trim()
      }
    }
    if (form.type === 'credit') {
      payload.usedCredit = yuanToFen(usedCreditYuan.value)
      payload.creditBilledFen = yuanToFen(billedYuan.value)
      payload.installmentPlans = buildInstallmentPlansPayload()
    } else {
      payload.balance = yuanToFen(balanceYuan.value)
      payload.installmentPlans = []
    }
    if (props.accountId) {
      payload.archived = form.archived
      await accounts.update(props.accountId, payload as any)
      await syncDefaults(props.accountId)
    } else {
      await accounts.create(payload as any)
    }
    // 保存后丢弃明文
    secretsRevealed.value = false
    emit('saved')
  } finally {
    saving.value = false
  }
}

async function syncDefaults(accountId: number) {
  if (form.archived) {
    await auth.fetchMe().catch(() => {})
    return
  }
  const payload: Record<string, unknown> = {}
  const wasExpense = auth.profile?.defaultExpenseAccountId === accountId
  const wasRepay = auth.profile?.defaultAccountId === accountId
  if (asDefaultExpense.value && !wasExpense) payload.defaultExpenseAccountId = accountId
  if (!asDefaultExpense.value && wasExpense) payload.clearDefaultExpenseAccount = true
  if (form.type !== 'credit') {
    if (asDefaultRepay.value && !wasRepay) payload.defaultAccountId = accountId
    if (!asDefaultRepay.value && wasRepay) payload.clearDefaultAccount = true
  } else if (wasRepay) {
    payload.clearDefaultAccount = true
  }
  if (Object.keys(payload).length) {
    await auth.updateSettings(payload)
  }
}

async function markReconciled() {
  if (!props.accountId) return
  reconciling.value = true
  try {
    await accounts.update(props.accountId, { lastReconciledAt: dayjs().toISOString() } as any)
    emit('reconciled')
  } finally {
    reconciling.value = false
  }
}

defineExpose({ save, saving })
</script>

<style scoped>
.plan-row {
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid color-mix(in srgb, var(--border, #d0d7d3) 70%, transparent);
}
.plan-row:last-of-type {
  border-bottom: none;
}
/* 弹窗内勿用视口断点拉四列，否则 label 会被挤成省略号 */
.plan-row__grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  column-gap: 12px;
  row-gap: 0;
}
.plan-row__actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 2px;
}
.plan-share {
  font-size: 12px;
  color: var(--muted, #6b7280);
}
.plan-summary {
  margin-top: 12px;
  font-size: 13px;
  line-height: 1.55;
  color: var(--text-secondary, #4b5563);
}
.plan-warn {
  color: var(--warn, #c45c3e);
  margin-top: 4px;
}
</style>
