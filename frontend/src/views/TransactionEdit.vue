<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" @click="$router.back()" />
      <h1>{{ isEdit ? '编辑流水' : '记一笔' }}</h1>
      <v-btn v-if="isEdit" icon="mdi-delete" variant="text" color="error" @click="onDelete" />
      <span v-else style="width:40px" />
    </div>

    <div v-if="!isEdit" class="ai-row mb-3">
      <v-btn variant="tonal" color="secondary" class="flex1" to="/transactions/recognize">
        <v-icon start>mdi-robot</v-icon>
        文字识别
      </v-btn>
      <v-btn variant="tonal" color="secondary" class="flex1" @click="openAI('image')">
        <v-icon start>mdi-camera</v-icon>
        票据识图
      </v-btn>
    </div>

    <div v-if="!isEdit && templates.length" class="tpls mb-3">
      <v-chip
        v-for="t in templates"
        :key="t.id"
        class="ma-1"
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
    <p v-else-if="uiMode === 'repay'" class="mode-hint mb-3">
      信用账户还款写这里：从银行卡 / 支付宝等转入信用卡、花呗等，<strong>不计入本月支出</strong>。
    </p>
    <p v-else class="mode-hint mb-3">账户之间挪钱（还信用卡 / 花呗请用「还款」）。</p>

    <section v-if="!isEdit && uiMode === 'flow' && flowKind === 'expense'" class="form-section mb-3">
      <div class="form-section__body">
      <div class="meta-head">
        <div class="meta-title-row">
          <v-icon size="20" color="primary">mdi-ticket-percent-outline</v-icon>
          <div class="meta-copy">
            <div class="meta-title">计费规则</div>
            <div class="meta-sub">主要用于累计折扣标记；金额以手填为准</div>
          </div>
        </div>
        <v-switch v-model="fareEnabled" color="primary" density="compact" hide-details inset @update:model-value="onFareToggle" />
      </div>
      <template v-if="fareEnabled">
        <v-select
          v-model="form.fareRuleId"
          :items="fareRuleItems"
          item-title="title"
          item-value="value"
          label="选择规则"
          variant="outlined"
          density="compact"
          hide-details
          class="field"
          @update:model-value="onFareRulePicked"
        />
        <div class="fare-actions">
          <v-btn size="small" variant="tonal" color="primary" :loading="farePreviewLoading" :disabled="!form.fareRuleId" @click="applySuggestedFare(true)">
            填入建议价
          </v-btn>
          <span v-if="fareHint" class="fare-hint">{{ fareHint }}</span>
        </div>
      </template>
      </div>
    </section>

    <v-text-field
      v-model="amountYuan"
      label="金额（元）"
      type="number"
      inputmode="decimal"
      variant="outlined"
      hide-details
      class="amount-field mb-3"
      prefix="¥"
      @update:model-value="onAmountManualEdit"
    />

    <section v-if="showCreditExtras" class="form-section mb-3">
      <div class="form-section__body">
      <div class="meta-head">
        <div class="meta-title-row">
          <v-icon size="20" color="primary">mdi-credit-card-clock-outline</v-icon>
          <div class="meta-copy">
            <div class="meta-title">{{ isCreditExpense ? '刷卡分期 / 利息' : '还款分期 / 利息' }}</div>
            <div class="meta-sub">
              {{ isCreditExpense
                ? '仅影响账单展示拆分；账户余额按上方金额变动'
                : '建议金额=本期本金份额+利息；可改，保存以手填为准' }}
            </div>
          </div>
        </div>
      </div>

      <div v-if="repayStatement" class="repay-stmt mb-2">
        <div class="repay-stmt-row">
          <span>本期已还</span><strong>¥{{ fenToYuan(repayStatement.repaidAmount) }}</strong>
        </div>
        <div class="repay-stmt-row">
          <span>本期待还</span><strong>¥{{ fenToYuan(repayStatement.periodRemaining) }}</strong>
        </div>
        <div class="repay-stmt-row">
          <span>{{ repayStatement.periodRemaining === 0 && (repayStatement.remainingAfterPay || 0) > 0 ? '下期待还' : '还后待还' }}</span>
          <strong class="warn">¥{{ fenToYuan(repayStatement.remainingAfterPay) }}</strong>
        </div>
        <div class="repay-stmt-row">
          <span>整体待还</span><strong>¥{{ fenToYuan(repayTotalOutstanding) }}</strong>
        </div>
        <div class="repay-stmt-hint">
          <template v-if="repayStatement.periodRemaining === 0 && (repayStatement.remainingAfterPay || 0) > 0">
            本期已还清；剩余已用为下期/未出账。默认按本期待还，提前还款可加载整体待还
          </template>
          <template v-else>
            默认按本期待还预填；提前还款可加载整体待还
          </template>
        </div>
      </div>

      <div class="field mb-2">
        <v-text-field
          v-model.number="form.installmentPeriods"
          :label="isCreditExpense ? '分期期数' : '还款分几期'"
          type="number"
          min="1"
          max="60"
          variant="outlined"
          density="compact"
          hide-details
          hint="1 = 一次还清 / 不分期"
          persistent-hint
          @update:model-value="onRepayPlanChanged"
        />
      </div>
      <v-text-field
        v-model="interestYuan"
        label="利息 / 手续费（元）"
        type="number"
        inputmode="decimal"
        variant="outlined"
        density="compact"
        hide-details
        prefix="¥"
        class="mb-1"
        @update:model-value="onRepayPlanChanged"
      />
      <div v-if="creditExtraHint" class="fare-hint">{{ creditExtraHint }}</div>
      <div v-if="isCreditRepay || isRepayMode" class="repay-actions mt-2">
        <v-btn
          size="small"
          variant="tonal"
          color="primary"
          :loading="repayStmtLoading"
          @click="applyRepaySuggestion(true)"
        >按本期待还</v-btn>
        <v-btn
          size="small"
          variant="outlined"
          color="primary"
          :disabled="!repayTotalOutstanding"
          @click="applyEarlyRepay"
        >提前还款</v-btn>
      </div>
      </div>
    </section>

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
      :label="isRepayMode || isCreditRepay ? '信用卡账户' : '转入账户'"
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
      <v-btn
        icon="mdi-plus"
        variant="tonal"
        color="primary"
        title="新增分类"
        @click="openQuickCategory"
      />
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
      <v-btn
        icon="mdi-plus"
        variant="tonal"
        color="primary"
        title="新增标签"
        @click="openQuickTag"
      />
    </div>

    <v-text-field v-model="form.remark" label="备注" variant="outlined" hide-details class="field" />
    <NativeDateField v-model="happenedLocal" label="时间" type="datetime-local" variant="outlined" hide-details class="field field--last" />
      </div>
    </section>

    <section v-if="isEdit" class="form-section mb-3">
      <div class="form-section__body">
      <div class="meta-head">
        <div class="meta-title-row">
          <v-icon size="20" color="primary">mdi-ticket-percent-outline</v-icon>
          <div class="meta-copy">
            <div class="meta-title">计费规则</div>
            <div class="meta-sub">累计折扣标记；金额以手填为准</div>
          </div>
        </div>
        <v-switch v-model="fareEnabled" color="primary" density="compact" hide-details inset @update:model-value="onFareToggle" />
      </div>
      <template v-if="fareEnabled">
        <v-select
          v-model="form.fareRuleId"
          :items="fareRuleItems"
          item-title="title"
          item-value="value"
          label="选择规则"
          variant="outlined"
          density="compact"
          hide-details
          class="field"
          @update:model-value="onFareRulePicked"
        />
        <div class="fare-actions">
          <v-btn size="small" variant="tonal" color="primary" :loading="farePreviewLoading" :disabled="!form.fareRuleId" @click="applySuggestedFare(true)">
            填入建议价
          </v-btn>
          <span v-if="fareHint" class="fare-hint">{{ fareHint }}</span>
        </div>
      </template>
      </div>
    </section>

    <AttachmentUpload v-model="attachmentIds" :existing="existingAttachments" />

    <section class="form-section mb-4">
      <div class="form-section__body">
      <div class="meta-head">
        <div class="meta-title-row">
          <v-icon size="20" color="primary">
            {{ form.geoMode === 'route' ? 'mdi-map-marker-path' : 'mdi-map-marker-outline' }}
          </v-icon>
          <div class="meta-copy">
            <div class="meta-title">地点</div>
            <div class="meta-sub">消费位置或路费起终点，可选</div>
          </div>
        </div>
        <v-chip v-if="geoLabel" size="small" color="primary" variant="tonal">
          {{ form.geoMode === 'route' ? '行程' : '位置' }}
        </v-chip>
      </div>

      <button v-if="!geoLabel" type="button" class="loc-empty" @click="showMap = true">
        <v-icon size="28" color="primary">mdi-map-search-outline</v-icon>
        <div class="loc-empty-text">
          <strong>添加地点</strong>
          <span>点选地图，或一键定位当前位置</span>
        </div>
        <v-icon size="20" color="primary">mdi-chevron-right</v-icon>
      </button>

      <div v-else class="loc-filled">
        <div class="loc-preview" @click="showMap = true">
          <div class="loc-kicker">{{ form.geoMode === 'route' ? '已选行程' : '已选位置' }}</div>
          <div class="loc-main">{{ geoPrimary }}</div>
          <div v-if="geoSecondary" class="loc-second">{{ geoSecondary }}</div>
        </div>
        <div class="loc-actions">
          <v-btn variant="tonal" color="primary" @click="showMap = true">
            <v-icon start>mdi-pencil-outline</v-icon>
            修改
          </v-btn>
          <v-btn variant="text" color="error" @click="clearGeo">
            <v-icon start>mdi-close</v-icon>
            清除
          </v-btn>
        </div>
      </div>
      </div>
    </section>

    <v-alert v-if="error" type="error" class="mb-3" density="compact">{{ error }}</v-alert>
    <v-btn color="primary" block size="large" class="hero-cta" :loading="saving" @click="onSave">保存</v-btn>

    <v-dialog
      v-model="showMap"
      :fullscreen="mapFullscreen"
      :max-width="mapFullscreen ? undefined : 960"
      transition="dialog-bottom-transition"
    >
      <div class="map-dialog-shell" :class="{ 'is-full': mapFullscreen }">
        <AmapPicker
          v-if="showMap"
          :initial-mode="form.geoMode === 'route' ? 'route' : 'point'"
          :initial-lng="form.geoLng"
          :initial-lat="form.geoLat"
          :initial-name="form.geoName"
          :initial-end-lng="form.geoEndLng"
          :initial-end-lat="form.geoEndLat"
          :initial-end-name="form.geoEndName"
          @pick="onPick"
          @close="showMap = false"
        />
      </div>
    </v-dialog>

    <v-dialog v-model="showQuickCategory" max-width="400">
      <v-card>
        <v-card-title>新增分类</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="quickCategoryName"
            label="分类名称"
            variant="outlined"
            autofocus
            @keyup.enter="saveQuickCategory"
          />
          <p class="hint">将创建为当前「{{ form.type === 'income' ? '收入' : '支出' }}」分类，可稍后在分类管理里调整层级。</p>
          <v-alert v-if="quickError" type="error" density="compact" class="mt-2">{{ quickError }}</v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="showQuickCategory = false">取消</v-btn>
          <v-btn color="primary" :loading="quickSaving" @click="saveQuickCategory">创建并选用</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="showQuickTag" max-width="400">
      <v-card>
        <v-card-title>新增标签</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="quickTagName"
            label="标签名称"
            variant="outlined"
            autofocus
            @keyup.enter="saveQuickTag"
          />
          <v-alert v-if="quickError" type="error" density="compact" class="mt-2">{{ quickError }}</v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="showQuickTag = false">取消</v-btn>
          <v-btn color="primary" :loading="quickSaving" @click="saveQuickTag">创建并选用</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="showAI" max-width="520">
      <v-card>
        <v-card-title>票据识图</v-card-title>
        <v-card-text>
          <p class="hint">上传小票 / 支付截图，由 DeepSeek 多模态模型识别</p>
          <input
            ref="fileInput"
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp,image/*"
            capture="environment"
            class="hidden-file"
            @change="onFileChange"
          />
          <div v-if="aiPreview" class="preview mb-3">
            <img :src="aiPreview" alt="预览" />
          </div>
          <v-btn variant="tonal" block class="mb-3" @click="fileInput?.click()">
            <v-icon start>mdi-image</v-icon>
            {{ aiFile ? '重新选择图片' : '选择 / 拍摄图片' }}
          </v-btn>
          <v-text-field v-model="aiHint" label="补充说明（可选）" variant="outlined" hint="如：用招商信用卡付的" persistent-hint />
          <v-alert v-if="aiError" type="error" density="compact" class="mt-2">{{ aiError }}</v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="showAI = false">取消</v-btn>
          <v-btn color="primary" :loading="aiLoading" :disabled="!aiFile" @click="onAI">
            识别并填入
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import dayjs from 'dayjs'
import http from '../api/http'
import { useAccountStore, type CreditStatement } from '../stores/account'
import { useCategoryStore } from '../stores/category'
import { useTagStore } from '../stores/tag'
import { useTransactionStore } from '../stores/transaction'
import { useAuthStore } from '../stores/auth'
import { yuanToFen, fenToYuan } from '../utils/money'
import { cityAllowsAuto, getCachedCity } from '../utils/fare'
import AmapPicker, { type GeoPickResult } from '../components/AmapPicker.vue'
import AttachmentUpload from '../components/AttachmentUpload.vue'
import NativeDateField from '../components/NativeDateField.vue'

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
const showMap = ref(false)
const mapFullscreen = ref(typeof window === 'undefined' ? true : window.innerWidth < 960)
const saving = ref(false)
const error = ref('')

function onMapViewport() {
  mapFullscreen.value = window.innerWidth < 960
}
const templates = ref<any[]>([])
const fareRules = ref<any[]>([])
const fareEnabled = ref(false)
const farePreviewLoading = ref(false)
const fareHint = ref('')
const showAI = ref(false)
const aiHint = ref('')
const aiFile = ref<File | null>(null)
const aiPreview = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const aiLoading = ref(false)
const aiError = ref('')

const showQuickCategory = ref(false)
const showQuickTag = ref(false)
const quickCategoryName = ref('')
const quickTagName = ref('')
const quickSaving = ref(false)
const quickError = ref('')

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
})
const interestYuan = ref('0')
const repayStatement = ref<CreditStatement | null>(null)
const repayStmtLoading = ref(false)
/** 还款金额是否仍由账单/期数自动回填（手改金额后关闭） */
const repayAmountAuto = ref(true)
/** 用于计算还款建议本金的基准（分）：优先本期待还 */
const repayPrincipalBaseFen = ref(0)
let applyingRepaySuggestion = false
let syncingUiMode = false

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
const showCreditExtras = computed(() => isCreditExpense.value || isCreditRepay.value || isRepayMode.value)

const repayTotalOutstanding = computed(() => {
  const s = repayStatement.value
  if (!s) return 0
  return s.totalOutstanding ?? s.outstandingBalance ?? 0
})

function firstShareFen(total: number, periods: number) {
  const n = Math.max(1, Math.min(60, Math.floor(periods) || 1))
  const t = Math.max(0, Math.floor(total))
  const base = Math.floor(t / n)
  return base + (t % n)
}

const creditExtraHint = computed(() => {
  const periods = Math.max(1, Number(form.installmentPeriods) || 1)
  if (isCreditExpense.value) {
    if (periods <= 1 || !amountYuan.value) return ''
    const total = yuanToFen(amountYuan.value)
    const first = firstShareFen(total, periods)
    const rest = Math.floor(total / periods)
    return `约首期本金 ¥${fenToYuan(first)}，其后每期 ¥${fenToYuan(rest)}（利息计入首期出账）`
  }
  if (!(isCreditRepay.value || isRepayMode.value)) return ''
  const base = repayPrincipalBaseFen.value
  if (base <= 0) return '暂无本期待还，可手填金额；利息单独填写。'
  const principal = firstShareFen(base, periods)
  const interest = yuanToFen(interestYuan.value || '0')
  const total = principal + interest
  if (periods <= 1) {
    return interest > 0
      ? `建议一次还清：本金 ¥${fenToYuan(principal)} + 利息 ¥${fenToYuan(interest)} = ¥${fenToYuan(total)}`
      : `建议还清待还本金 ¥${fenToYuan(principal)}`
  }
  const rest = Math.floor(base / periods)
  return `本次建议：本金 ¥${fenToYuan(principal)} + 利息 ¥${fenToYuan(interest)} = ¥${fenToYuan(total)}；其后约 ${periods - 1} 期、每期本金约 ¥${fenToYuan(rest)}（下期再记一笔转账）`
})

const transferToOptions = computed(() => {
  const list = accounts.list.filter((a) => a.id !== form.accountId && !a.archived)
  if (isRepayMode.value) {
    const credits = list.filter((a) => a.type === 'credit')
    return credits.length ? credits : list
  }
  return list
})

function onAmountManualEdit() {
  if (applyingRepaySuggestion) return
  if (isCreditRepay.value || isRepayMode.value) repayAmountAuto.value = false
}

function onRepayPlanChanged() {
  if (isCreditRepay.value || isRepayMode.value) applyRepaySuggestion(false)
}

function applyRepaySuggestion(force: boolean) {
  if (!(isCreditRepay.value || isRepayMode.value)) return
  if (!force && !repayAmountAuto.value) return
  const periods = Math.max(1, Number(form.installmentPeriods) || 1)
  const base = repayPrincipalBaseFen.value
  if (base <= 0 && !force) return
  const principal = firstShareFen(base, periods)
  const interest = yuanToFen(interestYuan.value || '0')
  applyingRepaySuggestion = true
  amountYuan.value = fenToYuan(principal + interest)
  repayAmountAuto.value = true
  queueMicrotask(() => { applyingRepaySuggestion = false })
}

/** 提前还款：预填整体待还（已用额度） */
function applyEarlyRepay() {
  const total = repayTotalOutstanding.value
  if (total <= 0) return
  repayPrincipalBaseFen.value = total
  form.installmentPeriods = 1
  applyRepaySuggestion(true)
}

async function loadRepayStatement(accountId?: number) {
  const id = accountId || form.toAccountId
  if (!id) {
    repayStatement.value = null
    return
  }
  const acc = accounts.list.find((a) => a.id === id)
  if (!acc || acc.type !== 'credit' || !acc.billingDay) {
    repayStatement.value = null
    return
  }
  repayStmtLoading.value = true
  try {
    const stmt = await accounts.creditStatement(id)
    repayStatement.value = stmt
    const base = stmt.periodRemaining > 0
      ? stmt.periodRemaining
      : (stmt.dueAmount > 0 ? stmt.dueAmount : 0)
    repayPrincipalBaseFen.value = base
    applyRepaySuggestion(false)
  } catch {
    repayStatement.value = null
  } finally {
    repayStmtLoading.value = false
  }
}

const geoLabel = computed(() => {
  if (form.geoMode === 'route' && form.geoName && form.geoEndName) {
    return `${form.geoName} → ${form.geoEndName}`
  }
  return form.geoName || ''
})

const geoPrimary = computed(() => {
  if (form.geoMode === 'route') return form.geoName || ''
  return form.geoName || ''
})

const geoSecondary = computed(() => {
  if (form.geoMode === 'route' && form.geoEndName) return `→ ${form.geoEndName}`
  return ''
})

watch(() => form.type, (t, prev) => {
  if (syncingUiMode || t === prev) return
  form.categoryId = undefined
  if (!(isRepayMode.value && t === 'transfer')) {
    form.toAccountId = undefined
  }
})

watch(flowKind, (k) => {
  if (syncingUiMode || uiMode.value !== 'flow') return
  if (form.type !== k) {
    form.type = k
    form.categoryId = undefined
  }
})

watch(uiMode, (mode, prev) => {
  if (syncingUiMode || mode === prev) return
  enterUiMode(mode)
})

function enterUiMode(mode: 'flow' | 'transfer' | 'repay') {
  syncingUiMode = true
  try {
    if (mode === 'flow') {
      form.type = flowKind.value
      form.toAccountId = undefined
      if (form.remark === '信用卡还款') form.remark = ''
      repayStatement.value = null
    } else if (mode === 'transfer') {
      form.type = 'transfer'
      form.categoryId = undefined
      if (form.remark === '信用卡还款') form.remark = ''
      // 普通转账不强制信用账户
      if (form.toAccountId && accounts.list.some((a) => a.id === form.toAccountId && a.type === 'credit')) {
        form.toAccountId = undefined
      }
      repayStatement.value = null
    } else {
      form.type = 'transfer'
      form.categoryId = undefined
      form.remark = form.remark || '信用卡还款'
      repayAmountAuto.value = true
      form.installmentPeriods = Math.max(1, form.installmentPeriods || 1)
      const credit =
        accounts.list.find((a) => a.id === form.toAccountId && a.type === 'credit') ||
        accounts.list.find((a) => a.type === 'credit' && !a.archived)
      if (credit) form.toAccountId = credit.id
      if (form.accountId === form.toAccountId) {
        form.accountId = accounts.list.find((a) => a.id !== form.toAccountId && a.type !== 'credit')?.id
          || accounts.list.find((a) => a.id !== form.toAccountId)?.id
      }
      void loadRepayStatement(credit?.id)
    }
  } finally {
    void nextTick(() => { syncingUiMode = false })
  }
}

/** 主数据：已有缓存则跳过，避免编辑页再等一轮远程库 */
async function ensureCoreStores() {
  const jobs: Promise<unknown>[] = []
  if (!accounts.list.length) jobs.push(accounts.load())
  if (!categories.list.length) jobs.push(categories.load())
  if (!tags.list.length) jobs.push(tags.load())
  if (!auth.profile) jobs.push(auth.fetchMe().catch(() => {}))
  if (jobs.length) await Promise.all(jobs)
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
  interestYuan.value = fenToYuan(t.interestFen || 0)
  fareEnabled.value = !!t.fareRuleId
  amountYuan.value = fenToYuan(t.amount)
  happenedLocal.value = dayjs(t.happenedAt).format('YYYY-MM-DDTHH:mm')
  existingAttachments.value = (t.attachments || []).map((a) => ({ id: a.id, url: a.url }))
  attachmentIds.value = existingAttachments.value.map((a) => a.id)
  // 等 watch 跑完再放开，避免 form.type 变更把刚写入的 categoryId 清掉
  void nextTick(() => { syncingUiMode = false })
}

onMounted(async () => {
  window.addEventListener('resize', onMapViewport)
  onMapViewport()

  // 模板/计费不挡类型与标签回显
  void loadSideData()

  const storesP = ensureCoreStores()
  // 编辑：流水与字典并行拉取，不再串行干等
  const txP = isEdit.value ? tx.get(editId.value) : null

  if (txP) {
    const [t] = await Promise.all([txP, storesP])
    applyEditTransaction(t)
    if (uiMode.value === 'repay' && form.toAccountId) {
      void loadRepayStatement(form.toAccountId)
    }
    return
  }

  await storesP

  if (!form.accountId) {
    const def = auth.profile?.defaultAccountId
    form.accountId = def || accounts.list.find((a) => a.type !== 'credit')?.id || accounts.list[0]?.id
  }
  // 账户页 / 首页还款提醒入口
  if (route.query.mode === 'repay') {
    syncingUiMode = true
    uiMode.value = 'repay'
    syncingUiMode = false
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
    if (qFen > 0) {
      repayPrincipalBaseFen.value = qFen
      repayAmountAuto.value = true
    }
    enterUiMode('repay')
    if (credit) form.toAccountId = credit.id
    if (form.accountId === form.toAccountId) {
      form.accountId = accounts.list.find((a) => a.id !== form.toAccountId && a.type !== 'credit')?.id
        || accounts.list.find((a) => a.id !== form.toAccountId)?.id
    }
    if (qFen > 0) applyRepaySuggestion(true)
    await loadRepayStatement(credit?.id)
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', onMapViewport)
})

function openAI(_mode?: 'text' | 'image') {
  aiError.value = ''
  showAI.value = true
}

watch(
  () => form.toAccountId,
  (id) => {
    if (isEdit.value) return
    if (form.type === 'transfer' && id && accounts.list.some((a) => a.id === id && a.type === 'credit')) {
      repayAmountAuto.value = true
      void loadRepayStatement(id)
    }
  },
)

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  aiFile.value = file
  if (aiPreview.value) URL.revokeObjectURL(aiPreview.value)
  aiPreview.value = URL.createObjectURL(file)
}

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
  }
  if (data.amountYuan) amountYuan.value = String(data.amountYuan)
  if (data.accountId) form.accountId = data.accountId
  if (data.toAccountId) form.toAccountId = data.toAccountId
  if (data.categoryId) form.categoryId = data.categoryId
  if (data.tagIds) form.tagIds = data.tagIds
  if (data.remark) form.remark = data.remark
  if (data.happenedAt) happenedLocal.value = dayjs(data.happenedAt).format('YYYY-MM-DDTHH:mm')
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
  if (t.accountId) form.accountId = t.accountId
  form.toAccountId = t.toAccountId || undefined
  form.categoryId = t.categoryId || undefined
  form.tagIds = t.tagIds || []
  form.remark = t.remark || ''
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

function onFareToggle(on: unknown) {
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

function openQuickCategory() {
  quickCategoryName.value = ''
  quickError.value = ''
  showQuickCategory.value = true
}

function openQuickTag() {
  quickTagName.value = ''
  quickError.value = ''
  showQuickTag.value = true
}

async function saveQuickCategory() {
  const name = quickCategoryName.value.trim()
  if (!name) { quickError.value = '请输入分类名称'; return }
  if (form.type === 'transfer') { quickError.value = '转账无需分类'; return }
  quickSaving.value = true
  quickError.value = ''
  try {
    const cat = await categories.create({
      name,
      kind: form.type,
      icon: 'mdi-shape',
      sort: 0,
    })
    form.categoryId = cat.id
    showQuickCategory.value = false
  } catch (e: any) {
    quickError.value = e.message || '创建失败'
  } finally {
    quickSaving.value = false
  }
}

async function saveQuickTag() {
  const name = quickTagName.value.trim()
  if (!name) { quickError.value = '请输入标签名称'; return }
  quickSaving.value = true
  quickError.value = ''
  try {
    const tag = await tags.create({ name, color: '#1b7f5a', sort: 0 })
    if (!form.tagIds.includes(tag.id)) form.tagIds = [...form.tagIds, tag.id]
    showQuickTag.value = false
  } catch (e: any) {
    quickError.value = e.message || '创建失败'
  } finally {
    quickSaving.value = false
  }
}

function clearGeo() {
  form.geoMode = ''
  form.geoLng = null
  form.geoLat = null
  form.geoName = ''
  form.geoEndLng = null
  form.geoEndLat = null
  form.geoEndName = ''
}

function onPick(p: GeoPickResult) {
  form.geoMode = p.mode
  form.geoLng = p.lng
  form.geoLat = p.lat
  form.geoName = p.name
  if (p.mode === 'route') {
    form.geoEndLng = p.endLng ?? null
    form.geoEndLat = p.endLat ?? null
    form.geoEndName = p.endName || ''
  } else {
    form.geoEndLng = null
    form.geoEndLat = null
    form.geoEndName = ''
  }
  showMap.value = false
}

async function onAI() {
  aiError.value = ''
  aiLoading.value = true
  try {
    if (!aiFile.value) {
      aiError.value = '请先选择图片'
      return
    }
    const fd = new FormData()
    fd.append('file', aiFile.value)
    if (aiHint.value.trim()) fd.append('hint', aiHint.value.trim())
    const { data } = await http.post('/ai/recognize-image', fd, {
      timeout: 120000,
    })
    applyRecognized(data)
    try {
      const up = new FormData()
      up.append('file', aiFile.value)
      const att = await http.post('/attachments', up)
      if (att.data?.id) {
        attachmentIds.value = [...attachmentIds.value, att.data.id]
        existingAttachments.value = [
          ...existingAttachments.value,
          { id: att.data.id, url: att.data.url },
        ]
      }
    } catch { /* 附件可选 */ }
    showAI.value = false
  } catch (e: any) {
    aiError.value = e.message || '识别失败'
  } finally {
    aiLoading.value = false
  }
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
  saving.value = true
  try {
    const payload = {
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
      installmentPeriods: showCreditExtras.value ? Math.max(1, Number(form.installmentPeriods) || 1) : 1,
      interestFen: showCreditExtras.value ? yuanToFen(interestYuan.value || '0') : 0,
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
.tpls { display: flex; flex-wrap: wrap; }
.hint { color: var(--muted); font-size: 0.85rem; margin-bottom: 8px; }
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
.repay-stmt {
  padding: 10px 12px;
  border-radius: 12px;
  background: var(--primary-soft, rgba(27, 127, 90, 0.1));
  border: 1px solid var(--surface-border);
  min-width: 0;
}
.repay-stmt-row {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  font-size: 0.8rem;
  line-height: 1.6;
  min-width: 0;
}
.repay-stmt-row span { color: var(--muted); flex-shrink: 0; }
.repay-stmt-row strong {
  font-family: var(--font-display);
  font-variant-numeric: tabular-nums;
  text-align: right;
  overflow-wrap: anywhere;
}
.repay-stmt-row .warn { color: #c45c3e; }
.repay-stmt-hint {
  margin-top: 6px;
  font-size: 0.72rem;
  color: var(--muted);
  line-height: 1.35;
  overflow-wrap: anywhere;
}
.repay-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.hidden-file { display: none; }
.preview {
  border-radius: 12px;
  overflow: hidden;
  background: #0001;
  max-height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.preview img { max-width: 100%; max-height: 240px; object-fit: contain; }

.meta-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  min-width: 0;
}
.meta-title-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  min-width: 0;
  flex: 1;
}
.meta-copy { min-width: 0; flex: 1; }
.meta-title { font-weight: 700; font-size: 0.98rem; line-height: 1.2; overflow-wrap: anywhere; }
.meta-sub {
  color: var(--muted);
  font-size: 0.8rem;
  margin-top: 2px;
  line-height: 1.4;
  overflow-wrap: anywhere;
  word-break: break-word;
}
.meta-head > .v-switch,
.meta-head > .v-chip {
  flex-shrink: 0;
}
.fare-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.fare-hint {
  font-size: 0.75rem;
  color: var(--muted);
  line-height: 1.35;
  overflow-wrap: anywhere;
  word-break: break-word;
  min-width: 0;
  flex: 1 1 140px;
}

.loc-empty {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 12px;
  border-radius: 14px;
  border: 1.5px dashed rgba(27, 127, 90, 0.45);
  background: var(--primary-soft);
  color: inherit;
  text-align: left;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  min-width: 0;
}
.loc-empty-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.loc-empty-text strong { font-size: 0.95rem; overflow-wrap: anywhere; }
.loc-empty-text span {
  font-size: 0.78rem;
  color: var(--muted);
  overflow-wrap: anywhere;
  word-break: break-word;
}

.loc-filled { display: flex; flex-direction: column; gap: 12px; }
.loc-preview {
  padding: 12px 14px;
  border-radius: 12px;
  background: var(--primary-soft);
  border: 1px solid rgba(27, 127, 90, 0.22);
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}
.loc-kicker {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--muted);
  margin-bottom: 4px;
}
.loc-main {
  font-weight: 700;
  font-size: 0.95rem;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.loc-second {
  margin-top: 4px;
  color: var(--muted);
  font-size: 0.86rem;
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.loc-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

@media (min-width: 960px) {
  .loc-filled {
    flex-direction: row;
    align-items: stretch;
  }
  .loc-preview { flex: 1; }
  .loc-actions {
    flex-direction: column;
    justify-content: center;
    min-width: 120px;
  }
}
</style>

<style>
.map-dialog-shell {
  height: min(85vh, 820px);
  overflow: hidden;
  border-radius: 18px;
  background: var(--bg, #eef5f0);
}
.map-dialog-shell.is-full {
  height: 100%;
  min-height: var(--app-height, 100dvh);
  border-radius: 0;
}
.map-dialog-shell .picker {
  height: 100%;
}
</style>
