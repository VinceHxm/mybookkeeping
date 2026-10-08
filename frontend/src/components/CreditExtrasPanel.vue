<template>
  <section class="form-section mb-3">
    <div class="form-section__body">
      <div v-if="statement" class="repay-stmt mb-2">
        <div class="repay-stmt-row">
          <span>本期已还<template v-if="(statement.refundAmount || 0) > 0">（含退款 ¥{{ fenToYuan(statement.refundAmount || 0) }}）</template></span><strong>¥{{ fenToYuan(statement.repaidAmount) }}</strong>
        </div>
        <div class="repay-stmt-row">
          <span>本期待还</span><strong>¥{{ fenToYuan(statement.periodRemaining) }}</strong>
        </div>
        <div class="repay-stmt-row">
          <span>{{ settledWithUsage ? '剩余已用' : '还后待还' }}</span>
          <strong class="warn">¥{{ fenToYuan(statement.remainingAfterPay) }}</strong>
        </div>
        <div class="repay-stmt-row">
          <span>整体待还</span><strong>¥{{ fenToYuan(totalOutstanding) }}</strong>
        </div>
        <div v-if="statement.periodRemaining === 0 && (statement.nextPeriodEstimate || 0) > 0" class="repay-stmt-row">
          <span>预计下次出账</span><strong>¥{{ fenToYuan(statement.nextPeriodEstimate || 0) }}</strong>
        </div>
        <div class="repay-stmt-hint">
          <template v-if="settledWithUsage">
            本期已还清。「剩余已用」是卡上还欠总额；还款默认按本期待还（现为 0），提前还款可加载整体待还。
          </template>
          <template v-else>默认按本期待还预填；提前还款可加载整体待还</template>
          <template v-if="(statement.futureInstallment || 0) > 0">
            · 尚未出账分期 ¥{{ fenToYuan(statement.futureInstallment) }}
          </template>
        </div>
      </div>

      <button
        type="button"
        class="credit-toggle"
        :class="{ 'is-active': hasPlan, 'is-open': expanded }"
        :aria-expanded="expanded"
        @click="expanded = !expanded"
      >
        <v-icon size="20" color="primary">mdi-credit-card-clock-outline</v-icon>
        <span class="credit-toggle__title">{{ isCreditExpense ? '分期 / 手续费' : '还款分期 / 利息' }}</span>
        <span class="credit-toggle__summary">{{ summary }}</span>
        <v-icon size="18" class="credit-toggle__chevron">mdi-chevron-down</v-icon>
      </button>

      <v-expand-transition>
        <div v-show="expanded" class="credit-panel">
          <div class="panel-sub mb-2">
            {{ isCreditExpense
              ? '仅影响账单展示拆分；账户余额按上方金额变动'
              : '建议金额=本期本金份额+利息；可改，保存以手填为准' }}
          </div>
          <div class="period-row mb-2">
            <v-chip-group
              :model-value="periods"
              selected-class="text-primary"
              class="period-chips"
              @update:model-value="onChipPicked"
            >
              <v-chip v-for="p in PERIOD_PRESETS" :key="p" :value="p" size="small" variant="outlined" filter>
                {{ p === 1 ? (isCreditExpense ? '不分期' : '一次还清') : `${p} 期` }}
              </v-chip>
            </v-chip-group>
            <v-text-field
              v-model.number="periods"
              label="期数"
              type="number"
              min="1"
              max="60"
              variant="outlined"
              density="compact"
              hide-details
              class="period-input"
              @update:model-value="emit('planChanged')"
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
            @update:model-value="emit('planChanged')"
          />
          <div v-if="showDeferOption" class="defer-row mt-2">
            <span class="defer-label">首期出账</span>
            <v-chip-group
              :model-value="deferCycles"
              mandatory
              selected-class="text-primary"
              class="period-chips"
              @update:model-value="(v: unknown) => { deferCycles = Number(v) || 0 }"
            >
              <v-chip :value="0" size="small" variant="outlined" filter>随刷卡 {{ dueLabel(0) }}</v-chip>
              <v-chip :value="1" size="small" variant="outlined" filter>下期 {{ dueLabel(1) }}</v-chip>
            </v-chip-group>
          </div>
          <div v-if="hint" class="panel-hint">{{ hint }}</div>
        </div>
      </v-expand-transition>

      <div v-if="showRepayActions" class="repay-actions mt-2">
        <v-btn size="small" variant="tonal" color="primary" :loading="repayLoading" @click="emit('suggest')">按本期待还</v-btn>
        <v-btn size="small" variant="outlined" color="primary" :disabled="!totalOutstanding" @click="emit('earlyRepay')">
          提前还款
        </v-btn>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import type { CreditStatement } from '../stores/account'
import { fenToYuan, firstShareFen, yuanToFen } from '../utils/money'

const props = defineProps<{
  statement: CreditStatement | null
  totalOutstanding: number
  /** 信用卡刷卡消费（否则为还款） */
  isCreditExpense: boolean
  showRepayActions: boolean
  repayLoading?: boolean
  /** 上方金额（分） */
  amountFen: number
  /** 还款建议本金基准（分） */
  repayBaseFen: number
  /** 刷卡账户账单日（1-28），0=未设 */
  billingDay?: number
  /** 消费时间（本地） */
  happenedAt?: string
}>()

const periods = defineModel<number>('periods', { required: true })
const interestYuan = defineModel<string>('interestYuan', { required: true })
const deferCycles = defineModel<number>('deferCycles', { default: 0 })

const emit = defineEmits<{
  planChanged: []
  suggest: []
  earlyRepay: []
}>()

const PERIOD_PRESETS = [1, 3, 6, 12, 24]
const expanded = ref(false)

const periodCount = computed(() => Math.max(1, Number(periods.value) || 1))
const interestFen = computed(() => yuanToFen(interestYuan.value || '0'))
const hasPlan = computed(() => periodCount.value > 1 || interestFen.value > 0)
watch(hasPlan, (v) => { if (v) expanded.value = true }, { immediate: true })

const showDeferOption = computed(() => props.isCreditExpense && periodCount.value > 1 && !!props.billingDay)

/** 与后端 firstDueStatementForPurchase 一致：消费日 ≤ 账单日记当月账单，否则下月 */
function dueLabel(defer: number) {
  const day = Math.min(28, Math.max(1, props.billingDay || 1))
  const at = dayjs(props.happenedAt || undefined)
  let due = at.date(day).startOf('day')
  if (at.date() > day) due = due.add(1, 'month')
  return due.add(defer, 'month').format('M/D')
}

const settledWithUsage = computed(() =>
  !!props.statement && props.statement.periodRemaining === 0 && (props.statement.remainingAfterPay || 0) > 0,
)

function onChipPicked(v: unknown) {
  if (v == null) return
  periods.value = Number(v)
  emit('planChanged')
}

const summary = computed(() => {
  const parts: string[] = []
  if (periodCount.value > 1) {
    parts.push(`${periodCount.value} 期`)
    if (props.isCreditExpense && props.amountFen > 0) {
      parts.push(`每期约 ¥${fenToYuan(Math.floor(props.amountFen / periodCount.value))}`)
    }
  } else {
    parts.push(props.isCreditExpense ? '不分期' : '一次还清')
  }
  if (interestFen.value > 0) parts.push(`${props.isCreditExpense ? '手续费' : '利息'} ¥${fenToYuan(interestFen.value)}`)
  if (showDeferOption.value) parts.push(`首期 ${dueLabel(deferCycles.value || 0)}`)
  return parts.join(' · ')
})

const hint = computed(() => {
  const n = periodCount.value
  if (props.isCreditExpense) {
    if (n <= 1 || props.amountFen <= 0) return ''
    const first = firstShareFen(props.amountFen, n)
    const rest = Math.floor(props.amountFen / n)
    return `约首期本金 ¥${fenToYuan(first)}，其后每期 ¥${fenToYuan(rest)}（利息计入首期出账）`
  }
  const base = props.repayBaseFen
  if (base <= 0) return '暂无本期待还，可手填金额；利息单独填写。'
  const principal = firstShareFen(base, n)
  const total = principal + interestFen.value
  if (n <= 1) {
    return interestFen.value > 0
      ? `建议一次还清：本金 ¥${fenToYuan(principal)} + 利息 ¥${fenToYuan(interestFen.value)} = ¥${fenToYuan(total)}`
      : `建议还清待还本金 ¥${fenToYuan(principal)}`
  }
  const rest = Math.floor(base / n)
  return `本次建议：本金 ¥${fenToYuan(principal)} + 利息 ¥${fenToYuan(interestFen.value)} = ¥${fenToYuan(total)}；其后约 ${n - 1} 期、每期本金约 ¥${fenToYuan(rest)}（下期再记一笔转账）`
})
</script>

<style scoped>
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
.credit-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  min-width: 0;
  padding: 2px 0;
  background: none;
  border: 0;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
}
.credit-toggle__title { font-weight: 600; font-size: 0.92rem; flex-shrink: 0; }
.credit-toggle__summary {
  flex: 1;
  min-width: 0;
  text-align: right;
  font-size: 0.82rem;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.credit-toggle.is-active .credit-toggle__summary {
  color: rgb(var(--v-theme-primary));
  font-weight: 600;
}
.credit-toggle__chevron { color: var(--muted); transition: transform 0.2s; }
.credit-toggle.is-open .credit-toggle__chevron { transform: rotate(180deg); }
.credit-panel { padding-top: 10px; }
.panel-sub {
  color: var(--muted);
  font-size: 0.8rem;
  line-height: 1.4;
  overflow-wrap: anywhere;
}
.panel-hint {
  font-size: 0.75rem;
  color: var(--muted);
  line-height: 1.35;
  overflow-wrap: anywhere;
}
.period-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.period-chips { flex: 1; min-width: 0; padding: 0; }
.period-input { flex: 0 0 84px; }
.defer-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.defer-label { flex-shrink: 0; font-size: 0.8rem; color: var(--muted); }
.repay-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
