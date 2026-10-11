<template>
  <v-dialog v-model="open" max-width="420" scrollable>
    <v-card>
      <v-card-title>{{ editTx ? t('revalue.editTitle') : t('revalue.title') }}</v-card-title>
      <v-card-text>
        <div v-if="account" class="orig">
          <div class="orig-title">{{ account.name }}</div>
          <div class="orig-sub">{{ t('revalue.bookBalance', { amount: fenToYuan(account.balance) }) }}</div>
        </div>

        <template v-if="!editTx">
          <v-text-field
            v-model="marketYuan"
            :label="t('revalue.marketValue')"
            type="number"
            inputmode="decimal"
            prefix="¥"
            variant="outlined"
            hide-details
            class="field"
            autofocus
          />
          <p class="delta" :class="deltaFen >= 0 ? 'amount-income' : 'amount-expense'">
            <template v-if="deltaFen === 0">{{ t('revalue.same') }}</template>
            <template v-else>{{ t(deltaFen > 0 ? 'revalue.gain' : 'revalue.loss', { amount: formatMoney(deltaFen, 'revalue') }) }}</template>
          </p>
        </template>
        <v-text-field
          v-else
          v-model="deltaYuan"
          :label="t('revalue.delta')"
          type="number"
          inputmode="decimal"
          prefix="¥"
          variant="outlined"
          hide-details
          class="field"
        />

        <NativeDateField
          v-model="happenedLocal"
          :label="t('revalue.time')"
          type="datetime-local"
          variant="outlined"
          hide-details
          class="field"
        />
        <v-text-field v-model="remark" :label="t('revalue.remark')" variant="outlined" hide-details class="field" />
        <p class="hint">
          {{ t('revalue.hint') }}
        </p>
        <v-alert v-if="error" type="error" density="compact" class="mt-2">{{ error }}</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-btn v-if="editTx" color="error" variant="text" :loading="deleting" @click="remove">{{ t('revalue.delete') }}</v-btn>
        <v-spacer />
        <v-btn variant="text" @click="open = false">{{ t('common.cancel') }}</v-btn>
        <v-btn color="primary" :loading="saving" @click="submit">{{ t('common.save') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import { useAccountStore } from '../stores/account'
import { useTransactionStore, type Transaction } from '../stores/transaction'
import { fenToYuan, formatMoney, yuanToFen } from '../utils/money'
import NativeDateField from './NativeDateField.vue'

const props = defineProps<{
  account: { id: number; name: string; balance: number } | null
  /** 传入时编辑已有的校准流水（改差额 / 删除） */
  editTx?: Transaction | null
}>()
const emit = defineEmits<{ saved: [] }>()

const { t } = useI18n()
const open = defineModel<boolean>({ required: true })
const accounts = useAccountStore()
const tx = useTransactionStore()

const marketYuan = ref('')
const deltaYuan = ref('')
const happenedLocal = ref('')
const remark = ref('')
const saving = ref(false)
const deleting = ref(false)
const error = ref('')

const deltaFen = computed(() => yuanToFen(marketYuan.value || '0') - (props.account?.balance || 0))

watch(open, (v) => {
  if (!v) return
  error.value = ''
  const cur = props.editTx
  if (cur) {
    deltaYuan.value = fenToYuan(cur.amount)
    happenedLocal.value = dayjs(cur.happenedAt).format('YYYY-MM-DDTHH:mm')
    remark.value = cur.remark || ''
  } else {
    marketYuan.value = props.account ? fenToYuan(props.account.balance) : ''
    happenedLocal.value = dayjs().format('YYYY-MM-DDTHH:mm')
    remark.value = ''
  }
})

async function submit() {
  if (!props.account) return
  const happenedAt = happenedLocal.value ? dayjs(happenedLocal.value).toISOString() : undefined
  saving.value = true
  error.value = ''
  try {
    const cur = props.editTx
    if (cur) {
      const amount = yuanToFen(deltaYuan.value || '0')
      if (amount === 0) throw new Error(t('revalue.deltaZero'))
      await tx.update(cur.id, {
        type: 'revalue',
        amount,
        accountId: cur.accountId,
        remark: remark.value.trim(),
        happenedAt,
      })
    } else {
      if (marketYuan.value === '' || yuanToFen(marketYuan.value) < 0) throw new Error(t('revalue.needMarketValue'))
      await accounts.revalue(props.account.id, {
        marketValue: yuanToFen(marketYuan.value),
        happenedAt,
        remark: remark.value.trim(),
      })
    }
    open.value = false
    emit('saved')
  } catch (e: any) {
    error.value = e?.message || t('revalue.saveFailed')
  } finally {
    saving.value = false
  }
}

async function remove() {
  const cur = props.editTx
  if (!cur) return
  deleting.value = true
  error.value = ''
  try {
    await tx.remove(cur.id)
    open.value = false
    emit('saved')
  } catch (e: any) {
    error.value = e?.message || t('revalue.deleteFailed')
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.hint { color: var(--muted); font-size: 0.8rem; line-height: 1.5; margin: 8px 0 0; }
.field { margin-bottom: 12px; }
.delta { margin: -4px 2px 12px; font-size: 0.85rem; font-weight: 600; font-variant-numeric: tabular-nums; }
.orig {
  padding: 10px 12px;
  margin-bottom: 14px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--muted) 10%, transparent);
}
.orig-title { font-weight: 700; }
.orig-sub { font-size: 0.78rem; color: var(--muted); margin-top: 2px; }
</style>
