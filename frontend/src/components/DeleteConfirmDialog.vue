<template>
  <v-dialog :model-value="modelValue" max-width="400" @update:model-value="onDialog">
    <v-card class="confirm-card">
      <v-card-title class="title">{{ step === 1 ? t('ui.deleteDialog.title') : t('ui.deleteDialog.titleAgain') }}</v-card-title>
      <v-card-text>
        <i18n-t v-if="step === 1" keypath="ui.deleteDialog.ask" tag="p" class="msg">
          <template #name><strong>{{ name }}</strong></template>
        </i18n-t>
        <i18n-t v-else keypath="ui.deleteDialog.askAgain" tag="p" class="msg warn">
          <template #name><strong>{{ name }}</strong></template>
        </i18n-t>
      </v-card-text>
      <v-card-actions class="actions">
        <v-btn variant="text" @click="cancel">{{ t('common.cancel') }}</v-btn>
        <v-spacer />
        <v-btn v-if="step === 1" color="error" variant="tonal" @click="step = 2">{{ t('ui.deleteDialog.continue') }}</v-btn>
        <v-btn v-else color="error" variant="flat" :loading="loading" @click="confirm">{{ t('ui.deleteDialog.confirm') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  modelValue: boolean
  name: string
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  confirm: []
}>()

const step = ref(1)

watch(
  () => props.modelValue,
  (v) => {
    if (v) step.value = 1
  },
)

function onDialog(v: boolean) {
  emit('update:modelValue', v)
  if (!v) step.value = 1
}

function cancel() {
  emit('update:modelValue', false)
  step.value = 1
}

function confirm() {
  emit('confirm')
}
</script>

<style scoped>
.confirm-card { background: rgb(var(--v-theme-surface)) !important; }
.title { font-size: 1.1rem !important; font-weight: 700; }
.msg { margin: 0; line-height: 1.5; color: rgb(var(--v-theme-on-surface)); }
.msg.warn { color: #c62828; }
.actions { padding: 8px 16px 16px !important; }
</style>
