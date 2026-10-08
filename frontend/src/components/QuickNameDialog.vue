<template>
  <v-dialog v-model="open" max-width="400">
    <v-card>
      <v-card-title>{{ title }}</v-card-title>
      <v-card-text>
        <v-text-field
          v-model="name"
          :label="label"
          variant="outlined"
          autofocus
          @keyup.enter="submit"
        />
        <p v-if="hint" class="hint">{{ hint }}</p>
        <v-alert v-if="error" type="error" density="compact" class="mt-2">{{ error }}</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="open = false">取消</v-btn>
        <v-btn color="primary" :loading="saving" @click="submit">创建并选用</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  title: string
  label: string
  hint?: string
  /** 创建并选用；抛错时把 message 显示在弹窗里 */
  save: (name: string) => Promise<void>
}>()

const open = defineModel<boolean>({ required: true })

const name = ref('')
const saving = ref(false)
const error = ref('')

watch(open, (v) => {
  if (!v) return
  name.value = ''
  error.value = ''
})

async function submit() {
  const n = name.value.trim()
  if (!n) { error.value = `请输入${props.label}`; return }
  saving.value = true
  error.value = ''
  try {
    await props.save(n)
    open.value = false
  } catch (e: any) {
    error.value = e?.message || '创建失败'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.hint { color: var(--muted); font-size: 0.85rem; margin-bottom: 8px; }
</style>
