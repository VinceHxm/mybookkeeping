<template>
  <v-dialog :model-value="modelValue" max-width="400" @update:model-value="onDialog">
    <v-card class="confirm-card">
      <v-card-title class="title">{{ step === 1 ? '确认删除' : '再次确认' }}</v-card-title>
      <v-card-text>
        <p v-if="step === 1" class="msg">
          确定删除「<strong>{{ name }}</strong>」吗？
        </p>
        <p v-else class="msg warn">
          删除后不可恢复。请再次确认要删除「<strong>{{ name }}</strong>」。
        </p>
      </v-card-text>
      <v-card-actions class="actions">
        <v-btn variant="text" @click="cancel">取消</v-btn>
        <v-spacer />
        <v-btn v-if="step === 1" color="error" variant="tonal" @click="step = 2">继续删除</v-btn>
        <v-btn v-else color="error" variant="flat" :loading="loading" @click="confirm">确认删除</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

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
