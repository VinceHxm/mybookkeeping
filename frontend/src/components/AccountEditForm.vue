<template>
  <div class="form-fields">
    <p class="section-label">基本信息</p>
    <v-text-field v-model="form.name" label="名称" variant="outlined" hide-details class="field" />
    <v-select
      v-model="form.type"
      :items="typeItems"
      label="类型"
      variant="outlined"
      hide-details
      class="field"
    />
    <v-text-field v-model.number="form.sort" label="排序" type="number" variant="outlined" hide-details class="field" />

    <template v-if="form.type === 'bank' || form.type === 'credit'">
      <p class="section-label">卡片信息</p>
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
      <p class="field-tip">
        默认脱敏；眼睛向服务器拉明文（不常驻）。改卡号请清空后重填完整号码，否则保存不会覆盖原值。
      </p>
      <v-text-field
        v-model="form.holderName"
        label="户名 / 持卡人"
        variant="outlined"
        hide-details
        class="field"
        :append-inner-icon="accountId && originalHolderMasked ? (secretsRevealed ? 'mdi-eye-off-outline' : 'mdi-eye-outline') : undefined"
        @click:append-inner="toggleFormReveal"
        @focus="onSecretFocus('holderName')"
        @blur="onSecretBlur('holderName')"
      />
    </template>

    <p class="section-label">{{ form.type === 'credit' ? '额度与账单' : '余额' }}</p>
    <template v-if="form.type === 'credit'">
      <v-text-field
        v-model="usedCreditYuan"
        :label="accountId ? '已用额度（元）' : '初始已用额度（元）'"
        type="number"
        inputmode="decimal"
        variant="outlined"
        hide-details
        class="field"
        @update:model-value="onUsedCreditChange"
      />
      <p class="field-tip">刷卡欠款总额；可用额度 = 信用额度 − 已用</p>
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
      <p class="field-tip">本期账单已出账、待还款部分</p>
      <v-text-field
        :model-value="unbilledYuan"
        label="未出账（元）"
        type="number"
        variant="outlined"
        readonly
        hide-details
        class="field"
      />
      <p class="field-tip">自动计算：已用额度 − 已出账</p>
      <v-text-field v-model="creditLimitYuan" label="信用额度（元）" type="number" variant="outlined" hide-details class="field" />
      <v-text-field v-model.number="form.billingDay" label="账单日（1-28）" type="number" variant="outlined" hide-details class="field" />
      <v-text-field v-model.number="form.paymentDueDay" label="还款日（1-28，可选）" type="number" variant="outlined" hide-details class="field" />
      <p class="field-tip">设置账单日后可看本期应还；填写还款日后，临近还款日会在首页提醒（无欠款不提醒）。</p>
    </template>
    <template v-else>
      <v-text-field
        v-model="balanceYuan"
        :label="accountId ? '余额（元）' : '初始余额（元）'"
        type="number"
        inputmode="decimal"
        variant="outlined"
        hide-details
        class="field"
      />
      <p v-if="accountId" class="field-tip">可直接改余额，不自动生成流水</p>
    </template>

    <p class="section-label">存放位置</p>
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

    <v-text-field
      v-model="form.remark"
      label="备注（可选）"
      variant="outlined"
      hide-details
      class="field mt-3"
    />

    <template v-if="accountId">
      <v-switch v-model="form.archived" label="归档（隐藏，保留流水）" hide-details class="field" />
      <v-btn class="mt-1 mb-2" variant="tonal" size="small" :loading="reconciling" @click="markReconciled">
        标记已对账
      </v-btn>

      <div class="danger-zone">
        <div class="danger-title">危险操作</div>
        <p class="danger-tip">删除账户不可恢复；若仍有流水关联可能无法删除。</p>
        <v-btn color="error" variant="outlined" block @click="emit('delete')">删除此账户</v-btn>
      </div>
    </template>
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
} from '../stores/account'
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
const saving = ref(false)
const reconciling = ref(false)
const revealLoading = ref(false)
const secretsRevealed = ref(false)
/** 打开表单时的脱敏快照，用于判断「未改动」 */
const originalCardMasked = ref('')
const originalHolderMasked = ref('')
const balanceYuan = ref('0')
const usedCreditYuan = ref('0')
const billedYuan = ref('0')
const creditLimitYuan = ref('0')
const attachmentIds = ref<number[]>([])
const formAttachments = ref<{ id: number; url: string }[]>([])

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
    billedYuan.value = '0'
    creditLimitYuan.value = '0'
    attachmentIds.value = []
    formAttachments.value = []
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
}

async function loadAccount() {
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

function onUsedCreditChange() {
  const used = yuanToFen(usedCreditYuan.value)
  const billed = yuanToFen(billedYuan.value)
  if (billed > used) billedYuan.value = fenToYuan(used)
}

function onBilledChange() {
  const used = yuanToFen(usedCreditYuan.value)
  const billed = yuanToFen(billedYuan.value)
  if (billed < 0) billedYuan.value = '0'
  else if (billed > used) billedYuan.value = fenToYuan(used)
}

async function save() {
  if (saving.value) return
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
    } else {
      payload.balance = yuanToFen(balanceYuan.value)
    }
    if (props.accountId) {
      payload.archived = form.archived
      await accounts.update(props.accountId, payload as any)
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
.section-label {
  margin: 4px 0 10px;
  font-size: 0.8rem;
  font-weight: 700;
  color: rgb(var(--v-theme-on-surface));
  letter-spacing: 0.02em;
}
.form-fields .field {
  margin-bottom: 14px;
}
.field-tip {
  margin: -6px 0 14px;
  padding: 0 2px;
  font-size: 0.78rem;
  color: var(--muted);
  line-height: 1.45;
}
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
