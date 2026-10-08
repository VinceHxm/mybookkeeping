<template>
  <v-dialog v-model="open" max-width="420" scrollable>
    <v-card>
      <v-card-title>{{ editTx ? '编辑市值校准' : '校准市值' }}</v-card-title>
      <v-card-text>
        <div v-if="account" class="orig">
          <div class="orig-title">{{ account.name }}</div>
          <div class="orig-sub">账上余额 ¥{{ fenToYuan(account.balance) }}</div>
        </div>

        <template v-if="!editTx">
          <v-text-field
            v-model="marketYuan"
            label="当前市值（元）"
            type="number"
            inputmode="decimal"
            prefix="¥"
            variant="outlined"
            hide-details
            class="field"
            autofocus
          />
          <p class="delta" :class="deltaFen >= 0 ? 'amount-income' : 'amount-expense'">
            <template v-if="deltaFen === 0">与账上一致，只记录校准时间</template>
            <template v-else>本次{{ deltaFen > 0 ? '浮盈' : '浮亏' }} {{ formatMoney(deltaFen, 'revalue') }}</template>
          </p>
        </template>
        <v-text-field
          v-else
          v-model="deltaYuan"
          label="校准差额（元，浮亏填负数）"
          type="number"
          inputmode="decimal"
          prefix="¥"
          variant="outlined"
          hide-details
          class="field"
        />

        <NativeDateField
          v-model="happenedLocal"
          label="校准时间"
          type="datetime-local"
          variant="outlined"
          hide-details
          class="field"
        />
        <v-text-field v-model="remark" label="备注（可选）" variant="outlined" hide-details class="field" />
        <p class="hint">
          照着支付宝 / 券商 App 的持仓市值填。差额记为一笔「市值校准」，只改这个账户的余额，不计入收支统计。
        </p>
        <v-alert v-if="error" type="error" density="compact" class="mt-2">{{ error }}</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-btn v-if="editTx" color="error" variant="text" :loading="deleting" @click="remove">删除</v-btn>
        <v-spacer />
        <v-btn variant="text" @click="open = false">取消</v-btn>
        <v-btn color="primary" :loading="saving" @click="submit">保存</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
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
  const t = props.editTx
  if (t) {
    deltaYuan.value = fenToYuan(t.amount)
    happenedLocal.value = dayjs(t.happenedAt).format('YYYY-MM-DDTHH:mm')
    remark.value = t.remark || ''
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
    const t = props.editTx
    if (t) {
      const amount = yuanToFen(deltaYuan.value || '0')
      if (amount === 0) throw new Error('差额不能为 0，不需要这笔校准可直接删除')
      await tx.update(t.id, {
        type: 'revalue',
        amount,
        accountId: t.accountId,
        remark: remark.value.trim(),
        happenedAt,
      })
    } else {
      if (marketYuan.value === '' || yuanToFen(marketYuan.value) < 0) throw new Error('请填写当前市值')
      await accounts.revalue(props.account.id, {
        marketValue: yuanToFen(marketYuan.value),
        happenedAt,
        remark: remark.value.trim(),
      })
    }
    open.value = false
    emit('saved')
  } catch (e: any) {
    error.value = e?.message || '保存失败'
  } finally {
    saving.value = false
  }
}

async function remove() {
  const t = props.editTx
  if (!t) return
  deleting.value = true
  error.value = ''
  try {
    await tx.remove(t.id)
    open.value = false
    emit('saved')
  } catch (e: any) {
    error.value = e?.message || '删除失败'
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
