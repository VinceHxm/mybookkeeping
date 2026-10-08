<template>
  <div class="fx">
    <MetaHead icon="mdi-currency-usd" title="外币支付" :subtitle="subtitle">
      <v-switch v-model="enabled" color="primary" density="compact" hide-details inset />
    </MetaHead>
    <template v-if="enabled">
      <div class="fx-row">
        <v-combobox
          v-model="currency"
          :items="currencyItems"
          item-title="title"
          item-value="value"
          :return-object="false"
          label="币种"
          variant="outlined"
          density="compact"
          hide-details
          class="fx-cur"
        />
        <v-text-field
          v-model="amountYuan"
          label="外币金额"
          type="number"
          inputmode="decimal"
          variant="outlined"
          density="compact"
          hide-details
          class="fx-amt"
        />
      </div>
      <div class="fx-rate">
        <template v-if="rate">
          {{ estimated ? '预估汇率' : '支付汇率' }}：1 {{ currencyCode }} = {{ rate }} CNY
        </template>
        <template v-else>填好外币金额与上方人民币金额后，自动算出支付汇率</template>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { FOREIGN_CURRENCIES, payRate, yuanToFen } from '../utils/money'
import MetaHead from './MetaHead.vue'

const props = withDefaults(defineProps<{
  /** 上方实际人民币金额（分） */
  cnyFen: number
  /** 人民币金额仍是预估（待确认） */
  estimated?: boolean
  subtitle?: string
}>(), {
  estimated: false,
  subtitle: '记录外币原价和实际人民币扣款，自动算支付汇率',
})

const enabled = defineModel<boolean>('enabled', { required: true })
const currency = defineModel<string>('currency', { required: true })
const amountYuan = defineModel<string>('amountYuan', { required: true })

const currencyItems = FOREIGN_CURRENCIES.map((c) => ({ title: `${c.code} ${c.name}`, value: c.code }))
const currencyCode = computed(() => String(currency.value || '').trim().toUpperCase())
const rate = computed(() => payRate(props.cnyFen, yuanToFen(amountYuan.value || '0')))
</script>

<style scoped>
.fx-row { display: flex; gap: 8px; margin-top: 12px; min-width: 0; }
.fx-cur { flex: 0 0 132px; min-width: 0; }
.fx-amt { flex: 1; min-width: 0; }
.fx-rate {
  margin-top: 8px;
  font-size: 0.78rem;
  color: var(--muted);
  line-height: 1.4;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}
</style>
