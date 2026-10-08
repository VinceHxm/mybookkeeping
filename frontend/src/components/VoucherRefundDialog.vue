<template>
  <v-dialog v-model="open" max-width="440" scrollable>
    <v-card>
      <v-card-title>{{ picked ? '团购券退款' : '选择要退款的团购券' }}</v-card-title>
      <v-card-text>
        <template v-if="!picked">
          <div v-if="listLoading" class="hint">加载中…</div>
          <div v-else-if="!candidates.length" class="hint">没有待核销的团购券。记支出时打开「团购券」开关即可标记。</div>
          <v-list v-else bg-color="transparent" class="py-0">
            <v-list-item
              v-for="v in candidates"
              :key="v.id"
              rounded="lg"
              class="pick-row surface-row"
              @click="pick(v)"
            >
              <v-list-item-title>{{ txTitle(v) }}</v-list-item-title>
              <v-list-item-subtitle>
                {{ dayjs(v.happenedAt).format('YYYY-MM-DD') }} · {{ v.account?.name }}
                <template v-if="v.voucherExpireOn"> · {{ expireText(v.voucherExpireOn) }}</template>
              </v-list-item-subtitle>
              <template #append>
                <span class="amount-expense">¥{{ fenToYuan(v.amount) }}</span>
              </template>
            </v-list-item>
          </v-list>
        </template>

        <template v-else>
          <div class="orig">
            <div class="orig-title">{{ txTitle(picked) }}</div>
            <div class="orig-sub">
              {{ dayjs(picked.happenedAt).format('YYYY-MM-DD') }} · {{ picked.account?.name }} · 原支付 ¥{{ fenToYuan(picked.amount) }}
            </div>
          </div>
          <v-text-field
            v-model="amountYuan"
            label="退款金额（元）"
            type="number"
            inputmode="decimal"
            prefix="¥"
            variant="outlined"
            hide-details
            class="field"
          />
          <v-select
            v-model="accountId"
            :items="accounts.list.filter((a) => !a.archived || a.id === accountId)"
            item-title="name"
            item-value="id"
            label="退回账户"
            variant="outlined"
            hide-details
            class="field"
          />
          <NativeDateField
            v-model="happenedLocal"
            label="退款到账时间"
            type="datetime-local"
            variant="outlined"
            hide-details
            class="field"
          />
          <p class="hint">
            信用卡 / 花呗：时间填银行账单上的<strong>记账日</strong>，这样会落到正确的账期，并计入当期已还。
            保存后会生成一笔「退款」收入，原券标记为已退款；删掉这笔退款可恢复为待核销。
          </p>
        </template>
        <v-alert v-if="error" type="error" density="compact" class="mt-2">{{ error }}</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-btn v-if="picked && !voucher" variant="text" @click="picked = null">重选</v-btn>
        <v-spacer />
        <v-btn variant="text" @click="open = false">取消</v-btn>
        <v-btn v-if="picked" color="primary" :loading="saving" @click="submit">确认退款</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import dayjs from 'dayjs'
import http from '../api/http'
import { useAccountStore } from '../stores/account'
import { useTransactionStore, type Transaction } from '../stores/transaction'
import { fenToYuan, txTitle, yuanToFen } from '../utils/money'
import NativeDateField from './NativeDateField.vue'

const props = defineProps<{
  /** 指定要退的券；不传则先列出待核销券供选择 */
  voucher?: Transaction | null
}>()
const emit = defineEmits<{ refunded: [refund: Transaction, voucherId: number] }>()

const open = defineModel<boolean>({ required: true })
const accounts = useAccountStore()
const tx = useTransactionStore()

const candidates = ref<Transaction[]>([])
const listLoading = ref(false)
const picked = ref<Transaction | null>(null)
const amountYuan = ref('')
const accountId = ref<number | undefined>()
const happenedLocal = ref('')
const saving = ref(false)
const error = ref('')

function expireText(d: string) {
  const days = dayjs(d).startOf('day').diff(dayjs().startOf('day'), 'day')
  if (days < 0) return `已过期 ${-days} 天`
  if (days === 0) return '今天过期'
  return `${days} 天后过期`
}

function pick(v: Transaction) {
  picked.value = v
  amountYuan.value = fenToYuan(v.amount)
  accountId.value = v.accountId
  happenedLocal.value = dayjs().format('YYYY-MM-DDTHH:mm')
  error.value = ''
}

async function loadCandidates() {
  listLoading.value = true
  try {
    const { data } = await http.get('/transactions', { params: { voucher: 'pending', limit: 200 } })
    const items: Transaction[] = data?.items || []
    // 快过期 / 已过期的排前面，没填过期日的放最后
    candidates.value = items.sort((a, b) =>
      (a.voucherExpireOn || '9999-99-99').localeCompare(b.voucherExpireOn || '9999-99-99'),
    )
  } finally {
    listLoading.value = false
  }
}

watch(open, (v) => {
  if (!v) return
  error.value = ''
  if (!accounts.list.length) void accounts.load()
  if (props.voucher) {
    pick(props.voucher)
  } else {
    picked.value = null
    void loadCandidates()
  }
})

async function submit() {
  const v = picked.value
  if (!v) return
  const amount = yuanToFen(amountYuan.value || '0')
  if (amount <= 0) { error.value = '请输入退款金额'; return }
  if (amount > v.amount) { error.value = '退款金额不能超过原支付金额'; return }
  saving.value = true
  error.value = ''
  try {
    const refund = await tx.refundVoucher(v.id, {
      amount,
      accountId: accountId.value,
      happenedAt: happenedLocal.value ? dayjs(happenedLocal.value).toISOString() : undefined,
    })
    open.value = false
    emit('refunded', refund, v.id)
  } catch (e: any) {
    error.value = e?.message || '退款失败'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.hint { color: var(--muted); font-size: 0.8rem; line-height: 1.5; margin: 8px 0 0; }
.field { margin-bottom: 12px; }
.pick-row { margin-bottom: 6px; cursor: pointer; }
.orig {
  padding: 10px 12px;
  margin-bottom: 14px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--muted) 10%, transparent);
}
.orig-title { font-weight: 700; }
.orig-sub { font-size: 0.78rem; color: var(--muted); margin-top: 2px; }
</style>
