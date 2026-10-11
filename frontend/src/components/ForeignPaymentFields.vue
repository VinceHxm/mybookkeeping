<template>
  <div class="fx">
    <MetaHead icon="mdi-currency-usd" :title="t('fx.title')" :subtitle="subtitle ?? t('fx.subtitle')">
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
          :label="t('fx.currency')"
          variant="outlined"
          density="compact"
          hide-details
          class="fx-cur"
        />
        <v-text-field
          v-model="amountYuan"
          :label="t('fx.amount')"
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
          {{ t('fx.rateLine', { label: estimated ? t('fx.estRate') : t('fx.payRate'), code: currencyCode, rate }) }}
        </template>
        <template v-else>{{ t('fx.rateHint') }}</template>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { FOREIGN_CURRENCIES, currencyName, payRate, yuanToFen } from '../utils/money'
import MetaHead from './MetaHead.vue'

const props = withDefaults(defineProps<{
  /** 上方实际人民币金额（分） */
  cnyFen: number
  /** 人民币金额仍是预估（待确认） */
  estimated?: boolean
  subtitle?: string
}>(), {
  estimated: false,
  subtitle: undefined,
})

const { t } = useI18n()

const enabled = defineModel<boolean>('enabled', { required: true })
const currency = defineModel<string>('currency', { required: true })
const amountYuan = defineModel<string>('amountYuan', { required: true })

const currencyItems = computed(() =>
  FOREIGN_CURRENCIES.map((code) => ({ title: `${code} ${currencyName(code)}`, value: code })),
)
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
