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
          :items="typeItems"
          label="类型"
          variant="outlined"
          hide-details
          class="field"
        />
        <v-text-field
          v-model.number="form.sort"
          label="排序"
          type="number"
          variant="outlined"
          hide-details
          class="field field--last"
        />
      </div>
    </section>

    <section v-if="form.type === 'bank' || form.type === 'credit'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">卡片信息</h3>
      </header>
      <div class="form-section__body">
        <v-text-field
          v-model="form.institution"
          :label="form.type === 'credit' ? '发卡机构' : '开户行'"
          variant="outlined"
          hide-details
          class="field"
          placeholder="如：工商银行"
        />
        <v-text-field
          v-model="form.cardNo"
          label="卡号"
          variant="outlined"
          hide-details
          class="field"
          inputmode="numeric"
          autocomplete="off"
          :append-inner-icon="accountId && originalCardMasked ? (secretsRevealed ? 'mdi-eye-off-outline' : 'mdi-eye-outline') : undefined"
          @click:append-inner="toggleFormReveal"
          @focus="onSecretFocus('cardNo')"
          @blur="onSecretBlur('cardNo')"
        />
        <v-text-field
          v-model="form.holderName"
          label="户名 / 持卡人"
          variant="outlined"
          hide-details
          class="field field--last"
          :append-inner-icon="accountId && originalHolderMasked ? (secretsRevealed ? 'mdi-eye-off-outline' : 'mdi-eye-outline') : undefined"
          @click:append-inner="toggleFormReveal"
          @focus="onSecretFocus('holderName')"
          @blur="onSecretBlur('holderName')"
        />
        <p class="form-section__tip">
          卡号与户名默认脱敏；点眼睛向服务器拉取明文（不常驻）。修改请先清空再填完整号码，否则保存不会覆盖原值。
        </p>
      </div>
    </section>

    <section v-if="form.type === 'credit'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">额度</h3>
      </header>
      <div class="form-section__body">
        <v-text-field
          v-model="creditLimitYuan"
          label="信用额度（元）"
          type="number"
          inputmode="decimal"
          variant="outlined"
          hide-details
          class="field"
          @update:model-value="onCreditLimitChange"
        />
        <v-text-field
          v-model="availableCreditYuan"
          label="可用额度（元）"
          type="number"
          inputmode="decimal"
          variant="outlined"
          hide-details
          class="field"
          @update:model-value="onAvailableCreditChange"
        />
        <v-text-field
          v-model="usedCreditYuan"
          :label="accountId ? '已用额度（元）' : '初始已用额度（元）'"
          type="number"
          inputmode="decimal"
          variant="outlined"
          hide-details
          class="field field--last"
          @update:model-value="onUsedCreditChange"
        />
        <p class="form-section__tip">
          银行 App 常见只显示可用：填可用后自动算已用 = 额度 − 可用；也可直接改已用，可用会同步。保存以已用为准。
        </p>
      </div>
    </section>

    <section v-if="form.type === 'credit'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">账单</h3>
      </header>
      <div class="form-section__body">
        <div class="form-section__group">
          <v-text-field
            v-model="billedYuan"
            label="已出账（元）"
            type="number"
            inputmode="decimal"
            variant="outlined"
            hide-details
            class="field"
            @update:model-value="onBilledChange"
          />
          <v-text-field
            :model-value="unbilledYuan"
            label="未出账（元）"
            type="number"
            variant="outlined"
            readonly
            hide-details
            class="field field--last"
          />
          <p class="form-section__tip">
            只填银行/花呗「已出账、当前应还」的金额。上期已还清、下期还在累积时请填
            <strong>0</strong>——花呗里的「本期已累计」不是已出账，应留在未出账里。未出账 = 已用 − 已出账。
          </p>
        </div>
        <div class="form-section__group">
          <v-text-field
            v-model.number="form.billingDay"
            label="账单日（1-28）"
            type="number"
            variant="outlined"
            hide-details
            class="field"
          />
          <v-text-field
            v-model.number="form.paymentDueDay"
            label="还款日（1-28，可选）"
            type="number"
            variant="outlined"
            hide-details
            class="field field--last"
          />
          <p class="form-section__tip">
            设置账单日后可看本期已还/待还；填写还款日后，临近还款日会在首页提醒（无欠款不提醒）。
          </p>
        </div>
      </div>
    </section>

    <section v-if="form.type === 'credit'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">存量分期（可选）</h3>
      </header>
      <div class="form-section__body">
        <p class="form-section__tip form-section__tip--alone">
          按 App「分期待还」抄剩余本金与期数，不改已用额度。系统里的「出账」= 账单日那一次结账，不是花呗正在累积的那一期。
          上期已还清、累计还在滚时：已出账填 0，首期选「下次出账起」，剩余本金含即将入下期账单的那一期。
          仅当上方「已出账」里已经含有该分期的当期款时，才选「已含在已出账」。
        </p>
        <div
          v-for="(row, idx) in planRows"
          :key="row.key"
          class="plan-row"
        >
          <v-text-field
            v-model="row.name"
            label="名称"
            variant="outlined"
            hide-details
            density="compact"
            class="field"
            placeholder="如：京东 12 期"
          />
          <div class="plan-row__grid">
            <v-text-field
              v-model="row.principalYuan"
              label="剩余本金"
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
              label="剩余期数"
              type="number"
              min="1"
              max="60"
              variant="outlined"
              hide-details
              density="compact"
              class="field"
            />
            <v-text-field
              v-model="row.interestYuan"
              label="每期利息"
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
              label="首期如何入账"
              variant="outlined"
              hide-details
              density="compact"
              class="field field--last"
            />
          </div>
          <div class="plan-row__actions">
            <span v-if="planShareHint(row)" class="plan-share">约每期 ¥{{ planShareHint(row) }}</span>
            <v-btn size="small" variant="text" color="error" @click="removePlanRow(idx)">移除</v-btn>
          </div>
        </div>
        <v-btn size="small" variant="tonal" color="primary" class="mt-1" @click="addPlanRow">
          添加分期计划
        </v-btn>
        <div v-if="planSummary.futurePrincipal > 0 || planSummary.includedInBilled > 0 || planRows.length" class="plan-summary">
          <div v-if="planSummary.includedInBilled > 0">
            已含在已出账内的分期首期约 ¥{{ fenToYuan(planSummary.includedInBilled) }}（不会再计入未出账）
          </div>
          <div v-if="planSummary.futurePrincipal > 0">
            尚未出账的分期本金 ¥{{ fenToYuan(planSummary.futurePrincipal) }}
            <template v-if="planSummary.nextShare > 0">
              · 其中下次出账约 ¥{{ fenToYuan(planSummary.nextShare) }}
            </template>
          </div>
          <div>
            未出账里的非分期约 ¥{{ fenToYuan(planSummary.nonInstallmentUnbilled) }}
            <template v-if="planSummary.nextShare > 0">
              · 估下次出账合计约 ¥{{ fenToYuan(planSummary.nonInstallmentUnbilled + planSummary.nextShare) }}
            </template>
          </div>
          <div v-if="planSummary.warnBilledZeroButIncluded" class="plan-warn">
            已出账为 0，但分期选了「已含在已出账」——请改成「下次出账起」，否则首期会对着已关闭的账单周期。
          </div>
          <div v-if="planSummary.warnAccruingAsBilled" class="plan-warn">
            若上期已还清、这里填的是花呗「已累计」而非应还账单，请把已出账改回 0，分期选「下次出账起」。
          </div>
          <div v-if="planSummary.overUnbilled" class="plan-warn">尚未出账的分期本金已超过未出账，请核对已用/已出账或分期金额</div>
          <div v-if="!form.billingDay" class="plan-warn">请先设置账单日，才能正确映射首期出账日</div>
        </div>
      </div>
    </section>

    <section v-else-if="form.type !== 'credit'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">余额</h3>
      </header>
      <div class="form-section__body">
        <v-text-field
          v-model="balanceYuan"
          :label="accountId ? '余额（元）' : '初始余额（元）'"
          type="number"
          inputmode="decimal"
          variant="outlined"
          hide-details
          class="field field--last"
        />
        <p v-if="accountId" class="form-section__tip">可直接改余额，不自动生成流水。</p>
      </div>
    </section>

    <section v-if="accountId" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">默认用途</h3>
      </header>
      <div class="form-section__body">
        <v-switch
          v-model="asDefaultExpense"
          label="默认消费账户（记收支时优先选用；全局仅一个）"
          color="primary"
          hide-details
          density="compact"
          class="field"
          :disabled="form.archived"
        />
        <v-switch
          v-if="form.type !== 'credit'"
          v-model="asDefaultRepay"
          label="默认还款账户（还信用卡时优先作转出；全局仅一个）"
          color="primary"
          hide-details
          density="compact"
          class="field field--last"
          :disabled="form.archived"
        />
        <p v-else class="form-section__tip">信用账户不能设为默认还款账户。</p>
        <p class="form-section__tip">开启后会替换原先的同类型默认账户；可与另一类默认设为同一账户。</p>
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">存放位置</h3>
      </header>
      <div class="form-section__body">
        <v-textarea
          v-model="form.storageNote"
          label="位置说明"
          variant="outlined"
          hide-details
          class="field"
          rows="2"
          auto-grow
          placeholder="如：客厅抽屉第二层 / 钱包夹层 / 保险柜"
        />
        <AttachmentUpload
          v-model="attachmentIds"
          :existing="formAttachments"
          title="位置照片"
          subtitle="柜子、抽屉等现场照片，可选 · 点图可放大"
        />
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">其他</h3>
      </header>
      <div class="form-section__body">
        <v-text-field
          v-model="form.remark"
          label="备注（可选）"
          variant="outlined"
          hide-details
          :class="accountId ? 'field' : 'field field--last'"
        />
        <template v-if="accountId">
          <v-switch v-model="form.archived" label="归档（隐藏，保留流水）" hide-details class="field" />
          <v-btn class="mt-1" variant="tonal" size="small" :loading="reconciling" @click="markReconciled">
            标记已对账
          </v-btn>
        </template>
      </div>
    </section>

    <div v-if="accountId" class="danger-zone">
      <div class="danger-title">危险操作</div>
      <p class="danger-tip">删除账户不可恢复；若仍有流水关联可能无法删除。</p>
      <v-btn color="error" variant="outlined" block @click="emit('delete')">删除此账户</v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import dayjs from 'dayjs'
import AttachmentUpload from './AttachmentUpload.vue'
import {
  useAccountStore,
  creditUsedFen,
  creditBilledFenOf,
  type Account,
  type CreditInstallmentPlan,
} from '../stores/account'
import { useAuthStore } from '../stores/auth'
import { fenToYuan, yuanToFen } from '../utils/money'
import { looksMaskedSecret, secretFieldForSave } from '../utils/mask'

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
}

let planKeySeq = 1
const planRows = ref<PlanRow[]>([])
const firstDueModeItems = [
  { title: '下次出账起', value: 'next' },
  { title: '已含在已出账', value: 'current' },
]

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

const typeItems = [
  { title: '现金', value: 'cash' },
  { title: '银行卡', value: 'bank' },
  { title: '信用', value: 'credit' },
  { title: '其他', value: 'other' },
]

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

function firstDueOnForMode(mode: 'current' | 'next', billingDay: number) {
  const { stmt, nextStmt } = creditStatementDates(billingDay)
  return (mode === 'current' ? stmt : nextStmt).format('YYYY-MM-DD')
}

function modeFromFirstDueOn(firstDueOn: string, billingDay: number): 'current' | 'next' {
  if (!billingDay || !firstDueOn) return 'next'
  const { stmt, nextStmt } = creditStatementDates(billingDay)
  if (firstDueOn === stmt.format('YYYY-MM-DD')) return 'current'
  if (firstDueOn === nextStmt.format('YYYY-MM-DD')) return 'next'
  // 过去日期视为已开始，尽量按是否等于本期判断
  if (firstDueOn <= stmt.format('YYYY-MM-DD')) return 'current'
  return 'next'
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
    const rem = principal % periods
    const firstShare = base + rem
    if (row.firstDueMode === 'next') {
      futurePrincipal += principal
      nextShare += firstShare
    } else {
      hasIncludedMode = true
      includedInBilled += firstShare
      futurePrincipal += Math.max(0, principal - firstShare)
      if (periods > 1) nextShare += base
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
  const plans = a?.installmentPlans || []
  const billingDay = a?.billingDay || 0
  planRows.value = plans.map((p) => ({
    key: planKeySeq++,
    name: p.name || '',
    principalYuan: fenToYuan(p.principalFen || 0),
    periods: Math.max(1, p.periods || 1),
    interestYuan: fenToYuan(p.interestPerPeriodFen || 0),
    firstDueMode: modeFromFirstDueOn(p.firstDueOn || '', billingDay),
  }))
}

function buildInstallmentPlansPayload(): CreditInstallmentPlan[] {
  const billingDay = form.billingDay || 0
  return planRows.value
    .map((row, i) => {
      const principalFen = Math.max(0, yuanToFen(row.principalYuan))
      const periods = Math.max(1, Math.min(60, Number(row.periods) || 1))
      if (principalFen <= 0 && !row.name.trim()) return null
      return {
        name: row.name.trim() || '存量分期',
        principalFen,
        periods,
        interestPerPeriodFen: Math.max(0, yuanToFen(row.interestYuan)),
        firstDueOn: firstDueOnForMode(row.firstDueMode, billingDay || 1),
        sort: i,
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
  revealLoading.value = true
  try {
    const data = await accounts.fetchSensitive(props.accountId)
    form.cardNo = data.cardNo || ''
    form.holderName = data.holderName || ''
    secretsRevealed.value = true
  } catch (e: any) {
    alert(e?.message || '获取敏感信息失败')
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
    alert('补录存量分期前请先设置账单日（1-28）')
    return
  }
  if (form.type === 'credit' && planSummary.value.overUnbilled) {
    if (!confirm('存量分期本金合计已超过未出账，仍要保存吗？')) return
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
