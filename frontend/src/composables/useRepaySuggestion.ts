import { computed, ref, type Ref } from 'vue'
import { useAccountStore, type CreditStatement } from '../stores/account'
import { fenToYuan, firstShareFen, yuanToFen } from '../utils/money'

/**
 * 信用卡还款建议金额：按账单「本期待还」或「整体待还」+ 期数 + 利息回填金额。
 * 用户手改金额后不再自动覆盖，直到点「按本期待还」/「提前还款」。
 */
export function useRepaySuggestion(opts: {
  amountYuan: Ref<string>
  interestYuan: Ref<string>
  periods: Ref<number>
  /** 当前是否处于还款场景（还款模式或转入信用卡） */
  active: Ref<boolean>
}) {
  const accounts = useAccountStore()
  const statement = ref<CreditStatement | null>(null)
  const loading = ref(false)
  /** 金额是否仍由账单/期数自动回填 */
  const auto = ref(true)
  /** 建议本金的基准（分）：优先本期待还 */
  const baseFen = ref(0)
  let applying = false

  const totalOutstanding = computed(() => {
    const s = statement.value
    if (!s) return 0
    return s.totalOutstanding ?? s.outstandingBalance ?? 0
  })

  function apply(force: boolean) {
    if (!opts.active.value) return
    if (!force && !auto.value) return
    if (baseFen.value <= 0 && !force) return
    const principal = firstShareFen(baseFen.value, opts.periods.value)
    const interest = yuanToFen(opts.interestYuan.value || '0')
    applying = true
    opts.amountYuan.value = fenToYuan(principal + interest)
    auto.value = true
    queueMicrotask(() => { applying = false })
  }

  /** 金额输入框手动修改时调用 */
  function onAmountManualEdit() {
    if (applying) return
    if (opts.active.value) auto.value = false
  }

  function onPlanChanged() {
    if (opts.active.value) apply(false)
  }

  /** 提前还款：预填整体待还（已用额度） */
  function applyEarlyRepay() {
    const total = totalOutstanding.value
    if (total <= 0) return
    baseFen.value = total
    opts.periods.value = 1
    apply(true)
  }

  /** 路由带入的建议金额（如首页还款提醒） */
  function presetBase(fen: number) {
    if (fen <= 0) return
    baseFen.value = fen
    auto.value = true
  }

  function reset() {
    statement.value = null
  }

  async function load(accountId?: number) {
    const acc = accountId ? accounts.list.find((a) => a.id === accountId) : undefined
    if (!acc || acc.type !== 'credit' || !acc.billingDay) {
      statement.value = null
      return
    }
    loading.value = true
    try {
      const stmt = await accounts.creditStatement(acc.id)
      statement.value = stmt
      baseFen.value = stmt.periodRemaining > 0 ? stmt.periodRemaining : Math.max(0, stmt.dueAmount)
      apply(false)
    } catch {
      statement.value = null
    } finally {
      loading.value = false
    }
  }

  return {
    statement,
    loading,
    auto,
    baseFen,
    totalOutstanding,
    apply,
    applyEarlyRepay,
    onAmountManualEdit,
    onPlanChanged,
    presetBase,
    reset,
    load,
  }
}
