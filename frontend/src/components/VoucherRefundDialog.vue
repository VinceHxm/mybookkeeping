<template>
  <v-dialog v-model="open" max-width="440" scrollable>
    <v-card>
      <v-card-title>{{ picked ? t('voucher.refundTitle') : t('voucher.pickTitle') }}</v-card-title>
      <v-card-text>
        <template v-if="!picked">
          <div v-if="listLoading" class="hint">{{ t('common.loading') }}</div>
          <div v-else-if="!candidates.length" class="hint">{{ t('voucher.noPending') }}</div>
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
              {{ dayjs(picked.happenedAt).format('YYYY-MM-DD') }} · {{ picked.account?.name }} · {{ t('voucher.origPaid', { amount: `¥${fenToYuan(picked.amount)}` }) }}
            </div>
          </div>
          <v-text-field
            v-model="amountYuan"
            :label="t('voucher.refundAmount')"
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
            :label="t('voucher.refundAccount')"
            variant="outlined"
            hide-details
            class="field"
          />
          <NativeDateField
            v-model="happenedLocal"
            :label="t('voucher.refundTime')"
            type="datetime-local"
            variant="outlined"
            hide-details
            class="field"
          />
          <i18n-t keypath="voucher.refundHint" tag="p" class="hint">
            <template #date><strong>{{ t('voucher.refundHintDate') }}</strong></template>
          </i18n-t>
        </template>
        <v-alert v-if="error" type="error" density="compact" class="mt-2">{{ error }}</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-btn v-if="picked && !voucher" variant="text" @click="picked = null">{{ t('voucher.repick') }}</v-btn>
        <v-spacer />
        <v-btn variant="text" @click="open = false">{{ t('common.cancel') }}</v-btn>
        <v-btn v-if="picked" color="primary" :loading="saving" @click="submit">{{ t('voucher.confirm') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
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
const { t } = useI18n()
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
  if (days < 0) return t('voucher.expiredDays', { n: -days })
  if (days === 0) return t('voucher.expireToday')
  return t('voucher.expireIn', { n: days })
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
  if (amount <= 0) { error.value = t('voucher.enterAmount'); return }
  if (amount > v.amount) { error.value = t('voucher.overAmount'); return }
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
    error.value = e?.message || t('voucher.failed')
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
