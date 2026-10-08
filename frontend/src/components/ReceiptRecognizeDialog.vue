<template>
  <v-dialog v-model="open" max-width="520">
    <v-card>
      <v-card-title>票据识图</v-card-title>
      <v-card-text>
        <p class="hint">上传小票 / 支付截图，由 DeepSeek 多模态模型识别</p>
        <input
          ref="fileInput"
          type="file"
          accept="image/jpeg,image/png,image/gif,image/webp,image/*"
          capture="environment"
          class="hidden-file"
          @change="onFileChange"
        />
        <div v-if="preview" class="preview mb-3">
          <img :src="preview" alt="预览" />
        </div>
        <v-btn variant="tonal" block class="mb-3" @click="fileInput?.click()">
          <v-icon start>mdi-image</v-icon>
          {{ file ? '重新选择图片' : '选择 / 拍摄图片' }}
        </v-btn>
        <v-text-field v-model="hint" label="补充说明（可选）" variant="outlined" hint="如：用招商信用卡付的" persistent-hint />
        <v-alert v-if="error" type="error" density="compact" class="mt-2">{{ error }}</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="open = false">取消</v-btn>
        <v-btn color="primary" :loading="loading" :disabled="!file" @click="recognize">
          识别并填入
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue'
import http from '../api/http'

const open = defineModel<boolean>({ required: true })

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
    error.value = '请先选择图片'
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
    error.value = e.message || '识别失败'
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
