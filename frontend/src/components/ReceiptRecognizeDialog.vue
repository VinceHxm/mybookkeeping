<template>
  <v-dialog v-model="open" max-width="520">
    <v-card>
      <v-card-title>{{ t('recognize.receipt.title') }}</v-card-title>
      <v-card-text>
        <p class="hint">{{ t('recognize.receipt.hint') }}</p>
        <input
          ref="fileInput"
          type="file"
          accept="image/jpeg,image/png,image/gif,image/webp,image/*"
          capture="environment"
          class="hidden-file"
          @change="onFileChange"
        />
        <div v-if="preview" class="preview mb-3">
          <img :src="preview" :alt="t('recognize.receipt.preview')" />
        </div>
        <v-btn variant="tonal" block class="mb-3" @click="fileInput?.click()">
          <v-icon start>mdi-image</v-icon>
          {{ file ? t('recognize.receipt.reselect') : t('recognize.receipt.select') }}
        </v-btn>
        <v-text-field
          v-model="hint"
          :label="t('recognize.receipt.extra')"
          variant="outlined"
          :hint="t('recognize.receipt.extraHint')"
          persistent-hint
        />
        <v-alert v-if="error" type="error" density="compact" class="mt-2">{{ error }}</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="open = false">{{ t('common.cancel') }}</v-btn>
        <v-btn color="primary" :loading="loading" :disabled="!file" @click="recognize">
          {{ t('recognize.receipt.submit') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import http from '../api/http'

const open = defineModel<boolean>({ required: true })
const { t } = useI18n()

const emit = defineEmits<{
  recognized: [data: any]
  /** 识别成功后原图作为附件上传（失败不影响识别结果） */
  attached: [attachment: { id: number; url: string }]
}>()

const hint = ref('')
const file = ref<File | null>(null)
const preview = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const loading = ref(false)
const error = ref('')

watch(open, (v) => { if (v) error.value = '' })

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const f = input.files?.[0]
  if (!f) return
  file.value = f
  if (preview.value) URL.revokeObjectURL(preview.value)
  preview.value = URL.createObjectURL(f)
}

onUnmounted(() => {
  if (preview.value) URL.revokeObjectURL(preview.value)
})

async function recognize() {
  error.value = ''
  if (!file.value) {
    error.value = t('recognize.receipt.needImage')
    return
  }
  loading.value = true
  try {
    const fd = new FormData()
    fd.append('file', file.value)
    if (hint.value.trim()) fd.append('hint', hint.value.trim())
    const { data } = await http.post('/ai/recognize-image', fd, { timeout: 120000 })
    emit('recognized', data)
    try {
      const up = new FormData()
      up.append('file', file.value)
      const att = await http.post('/attachments', up)
      if (att.data?.id) emit('attached', { id: att.data.id, url: att.data.url })
    } catch { /* 附件可选 */ }
    open.value = false
  } catch (e: any) {
    error.value = e.message || t('recognize.failed')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.hint { color: var(--muted); font-size: 0.85rem; margin-bottom: 8px; }
.hidden-file { display: none; }
.preview {
  border-radius: 12px;
  overflow: hidden;
  background: #0001;
  max-height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.preview img { max-width: 100%; max-height: 240px; object-fit: contain; }
</style>
