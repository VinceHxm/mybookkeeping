<template>
  <section class="meta-block mb-3">
    <div class="meta-head">
      <div class="meta-title-row">
        <v-icon size="20" color="primary">mdi-paperclip</v-icon>
        <div>
          <div class="meta-title">{{ title }}</div>
          <div class="meta-sub">{{ subtitle }}</div>
        </div>
      </div>
      <span v-if="previews.length" class="meta-count">{{ previews.length }} 张</span>
    </div>

    <div v-if="!previews.length" class="empty-actions">
      <label class="action-tile">
        <input type="file" accept="image/*" hidden @change="onFile" />
        <v-icon size="28" color="primary">mdi-image-plus-outline</v-icon>
        <span class="action-title">添加图片</span>
        <span class="action-sub">相册或文件</span>
      </label>
      <label class="action-tile">
        <input type="file" accept="image/*" capture="environment" hidden @change="onFile" />
        <v-icon size="28" color="primary">mdi-camera-outline</v-icon>
        <span class="action-title">拍照</span>
        <span class="action-sub">手机摄像头</span>
      </label>
    </div>

    <div v-else class="thumbs">
      <div
        v-for="(a, idx) in previews"
        :key="a.id"
        class="thumb"
        role="button"
        tabindex="0"
        @click="openPreview(idx)"
        @keydown.enter="openPreview(idx)"
      >
        <img :src="a.url" alt="附件预览" />
        <button type="button" class="rm" aria-label="删除附件" @click.stop="remove(a.id)">
          <v-icon size="14">mdi-close</v-icon>
        </button>
      </div>
      <label class="add" :class="{ busy: uploading }" :aria-busy="uploading">
        <input type="file" accept="image/*" hidden :disabled="uploading" @change="onFile" />
        <v-progress-circular v-if="uploading" indeterminate size="22" width="2" color="primary" />
        <template v-else>
          <v-icon size="22" color="primary">mdi-plus</v-icon>
          <span>继续加</span>
        </template>
      </label>
    </div>

    <div v-if="error" class="err">{{ error }}</div>

    <Teleport to="body">
      <div
        v-if="previewOpen"
        class="lightbox"
        role="dialog"
        aria-modal="true"
        aria-label="附件预览"
        @click.self="closePreview"
      >
        <button type="button" class="lb-close" aria-label="关闭预览" @click="closePreview">
          <v-icon size="22">mdi-close</v-icon>
        </button>
        <button
          v-if="previews.length > 1"
          type="button"
          class="lb-nav lb-prev"
          aria-label="上一张"
          @click.stop="shiftPreview(-1)"
        >
          <v-icon size="28">mdi-chevron-left</v-icon>
        </button>
        <img
          class="lb-img"
          :src="previews[previewIndex]?.url"
          alt="附件大图"
          @click.stop
        />
        <button
          v-if="previews.length > 1"
          type="button"
          class="lb-nav lb-next"
          aria-label="下一张"
          @click.stop="shiftPreview(1)"
        >
          <v-icon size="28">mdi-chevron-right</v-icon>
        </button>
        <div v-if="previews.length > 1" class="lb-counter">
          {{ previewIndex + 1 }} / {{ previews.length }}
        </div>
      </div>
    </Teleport>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import http from '../api/http'

const props = withDefaults(
  defineProps<{
    modelValue: number[]
    existing?: { id: number; url: string }[]
    title?: string
    subtitle?: string
  }>(),
  {
    title: '附件',
    subtitle: '票据 / 小票照片，可选 · 点图可放大',
  },
)
const emit = defineEmits<{ 'update:modelValue': [ids: number[]] }>()

const local = ref<{ id: number; url: string }[]>([...(props.existing || [])])
const uploading = ref(false)
const error = ref('')
const previewOpen = ref(false)
const previewIndex = ref(0)

watch(
  () => props.existing,
  (v) => {
    if (v) local.value = [...v]
  },
)

const previews = computed(() => local.value)

function sync() {
  emit('update:modelValue', local.value.map((a) => a.id))
}

function remove(id: number) {
  local.value = local.value.filter((a) => a.id !== id)
  sync()
  if (previewOpen.value) {
    if (!local.value.length) closePreview()
    else if (previewIndex.value >= local.value.length) {
      previewIndex.value = local.value.length - 1
    }
  }
}

function openPreview(idx: number) {
  previewIndex.value = idx
  previewOpen.value = true
  document.body.style.overflow = 'hidden'
}

function closePreview() {
  previewOpen.value = false
  document.body.style.overflow = ''
}

function shiftPreview(delta: number) {
  const n = previews.value.length
  if (n < 2) return
  previewIndex.value = (previewIndex.value + delta + n) % n
}

function onKey(e: KeyboardEvent) {
  if (!previewOpen.value) return
  if (e.key === 'Escape') closePreview()
  else if (e.key === 'ArrowLeft') shiftPreview(-1)
  else if (e.key === 'ArrowRight') shiftPreview(1)
}

watch(previewOpen, (open) => {
  if (open) window.addEventListener('keydown', onKey)
  else window.removeEventListener('keydown', onKey)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey)
  document.body.style.overflow = ''
})

async function onFile(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  error.value = ''
  uploading.value = true
  try {
    const fd = new FormData()
    fd.append('file', file)
    const { data } = await http.post('/attachments', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    local.value.push({ id: data.id, url: data.url })
    sync()
  } catch (err: any) {
    error.value = err.message || '上传失败'
  } finally {
    uploading.value = false
  }
}
</script>

<style scoped>
.meta-block {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-soft);
  padding: 14px;
}

.meta-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.meta-title-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.meta-title {
  font-weight: 700;
  font-size: 0.98rem;
  line-height: 1.2;
}

.meta-sub {
  color: var(--muted);
  font-size: 0.8rem;
  margin-top: 2px;
}

.meta-count {
  flex-shrink: 0;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--primary);
  background: var(--primary-soft);
  padding: 4px 10px;
  border-radius: 999px;
}

.empty-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.action-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  min-height: 96px;
  padding: 14px 10px;
  border-radius: 14px;
  border: 1.5px dashed rgba(27, 127, 90, 0.45);
  background: var(--primary-soft);
  cursor: pointer;
  text-align: center;
  -webkit-tap-highlight-color: transparent;
  transition: transform 0.15s ease, border-color 0.15s ease;
}

.action-tile:active {
  transform: scale(0.98);
}

.action-title {
  font-weight: 700;
  font-size: 0.92rem;
  margin-top: 4px;
}

.action-sub {
  color: var(--muted);
  font-size: 0.75rem;
}

.thumbs {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.thumb,
.add {
  width: 84px;
  height: 84px;
  border-radius: 14px;
  position: relative;
  overflow: hidden;
}

.thumb {
  background: var(--surface-solid);
  border: 1px solid var(--surface-border);
  box-shadow: var(--shadow-soft);
  cursor: zoom-in;
  -webkit-tap-highlight-color: transparent;
}

.thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  pointer-events: none;
}

.add {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  border: 1.5px dashed rgba(27, 127, 90, 0.4);
  background: var(--primary-soft);
  color: var(--primary);
  font-size: 0.72rem;
  font-weight: 600;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}

.add.busy {
  opacity: 0.7;
  pointer-events: none;
}

.rm {
  position: absolute;
  top: 6px;
  right: 6px;
  z-index: 1;
  width: 24px;
  height: 24px;
  padding: 0;
  margin: 0;
  border: 0;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.62);
  color: #fff;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.err {
  color: var(--danger, #c62828);
  font-size: 0.85rem;
  margin-top: 10px;
}

@media (min-width: 960px) {
  .thumb,
  .add {
    width: 96px;
    height: 96px;
  }
  .empty-actions {
    grid-template-columns: repeat(2, minmax(0, 220px));
  }
}
</style>

<style>
/* lightbox 挂到 body，不用 scoped */
.lightbox {
  position: fixed;
  inset: 0;
  z-index: 4000;
  background: rgba(0, 0, 0, 0.88);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px 16px 56px;
  box-sizing: border-box;
  -webkit-tap-highlight-color: transparent;
}
.lightbox .lb-img {
  max-width: min(100%, 960px);
  max-height: calc(100dvh - 120px);
  width: auto;
  height: auto;
  object-fit: contain;
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.45);
  user-select: none;
}
.lightbox .lb-close {
  position: absolute;
  top: max(12px, env(safe-area-inset-top, 0px));
  right: max(12px, env(safe-area-inset-right, 0px));
  width: 40px;
  height: 40px;
  border: 0;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.lightbox .lb-nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 44px;
  height: 44px;
  border: 0;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.lightbox .lb-prev { left: max(8px, env(safe-area-inset-left, 0px)); }
.lightbox .lb-next { right: max(8px, env(safe-area-inset-right, 0px)); }
.lightbox .lb-counter {
  position: absolute;
  bottom: max(16px, env(safe-area-inset-bottom, 0px));
  left: 50%;
  transform: translateX(-50%);
  color: rgba(255, 255, 255, 0.9);
  font-size: 0.85rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.02em;
}
</style>
