<template>
  <div
    class="swipe-row"
    :class="{ open: offset < -threshold / 2, revealing: offset < 0, dim }"
  >
    <div class="swipe-behind" aria-hidden="true">
      <button type="button" class="swipe-del" tabindex="-1" @click.stop="onDeleteClick">
        <v-icon icon="mdi-delete" size="22" />
        <span>删除</span>
      </button>
    </div>
    <div
      class="swipe-front"
      :style="{ transform: `translateX(${offset}px)` }"
      @pointerdown="onPointerDown"
      @pointermove="onPointerMove"
      @pointerup="onPointerUp"
      @pointercancel="onPointerUp"
      @click="onFrontClick"
    >
      <article
        class="entity-card"
        :class="{
          'is-default': accent !== 'none',
          'accent-expense': accent === 'expense',
          'accent-repay': accent === 'repay',
          'accent-both': accent === 'both',
        }"
      >
        <div
          v-if="accent !== 'none'"
          class="corner-mark"
          :class="accent"
          :title="cornerTitle"
          aria-hidden="true"
        >
          <v-icon size="15" class="corner-ico">mdi-bookmark</v-icon>
        </div>
        <div class="entity-main">
          <div class="avatar" :class="{ 'has-dot': !icon && !!dot }">
            <v-icon v-if="icon" :icon="icon" size="22" />
            <span v-else-if="dot" class="dot" :style="{ background: dot }" />
            <span v-else class="letter">{{ title.slice(0, 1) }}</span>
          </div>
          <div class="body">
            <div class="title-row">
              <h3 class="title">{{ title }}</h3>
              <slot name="badges" />
            </div>
            <div v-if="$slots.meta || meta" class="meta">
              <slot name="meta">{{ meta }}</slot>
            </div>
          </div>
          <v-icon class="chev" size="18">mdi-chevron-right</v-icon>
        </div>
      </article>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const props = withDefaults(defineProps<{
  title: string
  meta?: string
  icon?: string
  dot?: string
  dim?: boolean
  deleteLabel?: string
  /** 默认账户高亮：消费 / 还款 / 两者 */
  accent?: 'none' | 'expense' | 'repay' | 'both'
}>(), {
  deleteLabel: '删除',
  accent: 'none',
})

const emit = defineEmits<{
  edit: []
  delete: []
}>()

const cornerTitle = computed(() => {
  if (props.accent === 'both') return '默认消费 · 默认还款'
  if (props.accent === 'expense') return '默认消费'
  if (props.accent === 'repay') return '默认还款'
  return ''
})

const ACTION_W = 88
const threshold = 48
const offset = ref(0)
let startX = 0
let startY = 0
let startOffset = 0
let tracking = false
let axis: 'x' | 'y' | '' = ''
let moved = false
let pointerId: number | null = null

const OPEN_EVENT = 'mbk-swipe-close'

function closeOthers() {
  window.dispatchEvent(new CustomEvent(OPEN_EVENT, { detail: closeSelf }))
}

function closeSelf() {
  offset.value = 0
}

function onCloseAll(e: Event) {
  const detail = (e as CustomEvent).detail
  if (detail !== closeSelf) closeSelf()
}

onMounted(() => window.addEventListener(OPEN_EVENT, onCloseAll))
onBeforeUnmount(() => window.removeEventListener(OPEN_EVENT, onCloseAll))

function onPointerDown(e: PointerEvent) {
  if (e.button !== 0) return
  const t = e.target as HTMLElement | null
  if (t?.closest?.('a, button, input, textarea, select, .v-selection-control, .no-card-nav')) return
  tracking = true
  moved = false
  axis = ''
  startX = e.clientX
  startY = e.clientY
  startOffset = offset.value
  pointerId = e.pointerId
}

function onPointerMove(e: PointerEvent) {
  if (!tracking) return
  const dx = e.clientX - startX
  const dy = e.clientY - startY
  if (!axis) {
    if (Math.abs(dx) < 6 && Math.abs(dy) < 6) return
    axis = Math.abs(dx) > Math.abs(dy) ? 'x' : 'y'
    if (axis === 'x') {
      closeOthers()
      const el = e.currentTarget as HTMLElement
      if (pointerId != null) el.setPointerCapture?.(pointerId)
    } else {
      tracking = false
      return
    }
  }
  if (axis !== 'x') return
  e.preventDefault()
  moved = true
  const next = Math.min(0, Math.max(-ACTION_W, startOffset + dx))
  offset.value = next
}

function onPointerUp() {
  if (!tracking && axis !== 'x') {
    axis = ''
    pointerId = null
    return
  }
  tracking = false
  if (axis === 'x') {
    offset.value = offset.value < -threshold ? -ACTION_W : 0
  }
  axis = ''
  pointerId = null
}

function onFrontClick(e: MouseEvent) {
  if (moved || offset.value < -10) {
    if (offset.value < -10) offset.value = 0
    return
  }
  const t = e.target as HTMLElement | null
  if (t?.closest?.('a, button, input, textarea, select, .v-selection-control, .no-card-nav')) return
  emit('edit')
}

function onDeleteClick() {
  offset.value = 0
  emit('delete')
}
</script>

<style scoped>
.swipe-row {
  position: relative;
  margin-bottom: 10px;
  border-radius: var(--radius-md, 14px);
  overflow: hidden;
  touch-action: pan-y;
}
.swipe-row.dim { opacity: 0.55; }

.swipe-behind {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 88px;
  display: flex;
  justify-content: stretch;
  background: transparent;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.12s ease;
}
.swipe-row.revealing .swipe-behind,
.swipe-row.open .swipe-behind {
  opacity: 1;
  pointer-events: auto;
}
.swipe-del {
  width: 100%;
  height: 100%;
  border: 0;
  border-radius: 0;
  background: #c62828;
  color: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-size: 0.78rem;
  font-weight: 700;
  cursor: pointer;
  pointer-events: auto;
}

.swipe-front {
  position: relative;
  z-index: 1;
  transition: transform 0.18s ease;
  will-change: transform;
}

.entity-card {
  position: relative;
  padding: 14px;
  border-radius: var(--radius-md, 14px);
  border: 1px solid var(--surface-border);
  background: var(--surface-solid, #1c2620);
  box-shadow: var(--shadow-soft);
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  overflow: hidden;
  transition: border-color 0.15s ease, background 0.15s ease, box-shadow 0.15s ease;
}
.entity-card.is-default {
  box-shadow:
    var(--shadow-soft),
    0 0 0 1px color-mix(in srgb, var(--accent-tone, var(--primary)) 35%, transparent);
}
.entity-card.accent-expense {
  --accent-tone: var(--primary);
  border-color: color-mix(in srgb, var(--primary) 55%, transparent);
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--primary) 16%, var(--surface-solid, #1c2620)) 0%,
    var(--surface-solid, #1c2620) 72%
  );
}
.entity-card.accent-repay {
  --accent-tone: #3d8fad;
  border-color: color-mix(in srgb, #3d8fad 55%, transparent);
  background: linear-gradient(
    145deg,
    color-mix(in srgb, #3d8fad 16%, var(--surface-solid, #1c2620)) 0%,
    var(--surface-solid, #1c2620) 72%
  );
}
.entity-card.accent-both {
  --accent-tone: var(--primary);
  border-color: color-mix(in srgb, var(--primary) 60%, #3d8fad);
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--primary) 14%, var(--surface-solid, #1c2620)) 0%,
    color-mix(in srgb, #3d8fad 12%, var(--surface-solid, #1c2620)) 48%,
    var(--surface-solid, #1c2620) 100%
  );
}

.corner-mark {
  position: absolute;
  top: 0;
  right: 0;
  width: 0;
  height: 0;
  border-style: solid;
  border-width: 0 40px 40px 0;
  border-color: transparent var(--accent-tone, var(--primary)) transparent transparent;
  z-index: 2;
  pointer-events: none;
}
.corner-mark.repay {
  --accent-tone: #3d8fad;
}
.corner-mark.both {
  border-width: 0 44px 44px 0;
}
.corner-ico {
  position: absolute;
  top: 5px;
  right: -34px;
  color: #fff;
}

.entity-main {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}
.avatar {
  flex: 0 0 auto;
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  background: var(--primary-soft);
  color: var(--primary);
}
.avatar.has-dot {
  background: transparent;
  border: 1px solid var(--surface-border);
}
.dot { width: 18px; height: 18px; border-radius: 50%; display: block; }
.letter {
  font-family: var(--font-display);
  font-weight: 700;
  font-size: 1.05rem;
  color: var(--primary);
}
.body { flex: 1; min-width: 0; }
.title-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  padding-right: 18px;
}
.title {
  margin: 0;
  font-size: 1.02rem;
  font-weight: 700;
  line-height: 1.3;
  word-break: break-word;
}
.meta {
  margin-top: 6px;
  font-size: 0.82rem;
  line-height: 1.45;
  color: var(--muted);
  word-break: break-word;
  white-space: normal;
}
.chev {
  flex: 0 0 auto;
  color: var(--muted);
  margin-top: 10px;
  opacity: 0.7;
}
</style>
