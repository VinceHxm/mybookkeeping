<template>
  <div class="sm">
    <div class="sm-controls">
      <v-btn-toggle v-model="kind" mandatory density="compact" color="primary" variant="outlined" divided rounded="lg">
        <v-btn value="expense" size="small">支出</v-btn>
        <v-btn value="income" size="small">收入</v-btn>
      </v-btn-toggle>
      <span class="sm-note">
        已标地点 {{ data.locatedCount }} 笔 {{ yuan(data.locatedAmount) }}
        <template v-if="data.missingCount"> · 未标地点 {{ data.missingCount }} 笔</template>
      </span>
    </div>

    <div class="sm-grid">
      <div class="sm-map surface-card">
        <div ref="mapEl" class="map" />
        <div v-if="mapError" class="map-overlay map-overlay--err">{{ mapError }}</div>
        <div v-else-if="!loading && !data.places.length" class="map-overlay">
          该区间没有带地点的{{ kindLabel }}，记一笔时在「地点」里选位置即可出现在这里
        </div>
        <v-btn
          v-if="data.places.length"
          class="fit-btn"
          size="small"
          variant="flat"
          prepend-icon="mdi-fit-to-screen-outline"
          @click="fitAll"
        >全部</v-btn>
      </div>

      <section class="sm-list surface-card">
        <header class="list-head">
          <span>{{ data.places.length }} 个地点<template v-if="data.truncated">（仅显示金额最高的 500 个）</template></span>
        </header>
        <div v-if="!data.places.length" class="list-empty">{{ loading ? '加载中…' : '暂无' }}</div>
        <div ref="listEl" class="list-body">
          <button
            v-for="(p, i) in data.places"
            :key="placeKey(p)"
            :data-key="placeKey(p)"
            type="button"
            class="place"
            :class="{ active: selected === placeKey(p) }"
            @click="focusPlace(p)"
          >
            <span class="rank" :class="{ 'rank--hot': i < 3 }">{{ i + 1 }}</span>
            <div class="place-main">
              <div class="place-line">
                <span class="place-name">{{ p.name || '未命名地点' }}</span>
                <span class="place-amt" :class="kind === 'income' ? 'amount-income' : 'amount-expense'">{{ yuan(p.amount) }}</span>
              </div>
              <div class="place-sub">
                {{ p.count }} 笔 · 最近 {{ dayjs(p.lastAt).format('M月D日') }}
                <span class="place-pct">{{ data.locatedAmount ? percent(p.amount / data.locatedAmount) : '' }}</span>
              </div>
            </div>
          </button>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, onMounted, onUnmounted, ref, watch } from 'vue'
import dayjs from 'dayjs'
import http from '../../api/http'
import { loadAmap } from '../../utils/amap'
import { compactYuan, percent, statsParams, yuan, type StatsFilter } from '../../utils/stats'
import { useChartTheme } from '../../composables/useChartTheme'

interface Place {
  name: string
  lng: number
  lat: number
  amount: number
  count: number
  lastAt: string
}
interface PlaceStats {
  places: Place[]
  locatedAmount: number
  locatedCount: number
  missingAmount: number
  missingCount: number
  truncated: boolean
}

const props = defineProps<{ filter: StatsFilter }>()
const chart = useChartTheme()

const kind = ref<'expense' | 'income'>('expense')
const data = ref<PlaceStats>({ places: [], locatedAmount: 0, locatedCount: 0, missingAmount: 0, missingCount: 0, truncated: false })
const loading = ref(false)
const mapError = ref('')
const selected = ref('')
const mapEl = ref<HTMLElement | null>(null)
const listEl = ref<HTMLElement | null>(null)

const kindLabel = computed(() => (kind.value === 'income' ? '收入' : '支出'))
const color = computed(() => (kind.value === 'income' ? chart.value.income : chart.value.expense))

let AMap: any = null
let map: any = null
let markers: any[] = []
let info: any = null
const markerByKey = new Map<string, any>()

function placeKey(p: Place) {
  return `${p.name}|${p.lng.toFixed(5)}|${p.lat.toFixed(5)}`
}

function escapeHtml(s: string) {
  return s.replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]!)
}

function mapStyle() {
  return chart.value.dark ? 'amap://styles/dark' : 'amap://styles/normal'
}

async function initMap() {
  try {
    AMap = await loadAmap()
    if (!mapEl.value) return
    map = new AMap.Map(mapEl.value, { zoom: 11, mapStyle: mapStyle() })
    info = new AMap.InfoWindow({ offset: new AMap.Pixel(0, -18), closeWhenClickMap: true })
    renderMarkers()
  } catch (e: any) {
    mapError.value = e?.message || '地图加载失败'
  }
}

function renderMarkers() {
  if (!map || !AMap) return
  if (markers.length) map.remove(markers)
  markers = []
  markerByKey.clear()
  info?.close()
  const places = data.value.places
  if (!places.length) return
  const max = Math.max(...places.map((p) => p.amount), 1)
  const c = color.value
  // 金额越大圆越大；按金额升序添加，让大的压在上面
  for (const p of [...places].reverse()) {
    const size = Math.round(26 + Math.sqrt(p.amount / max) * 30)
    const m = new AMap.Marker({
      position: [p.lng, p.lat],
      anchor: 'center',
      zIndex: 100 + Math.round((p.amount / max) * 100),
      content: `<div class="mbk-pin" style="--pin:${c};--size:${size}px">${compactYuan(p.amount)}</div>`,
    })
    m.on('click', () => selectPlace(p, false))
    markers.push(m)
    markerByKey.set(placeKey(p), m)
  }
  map.add(markers)
  fitAll()
}

function fitAll() {
  if (map && markers.length) map.setFitView(markers, false, [48, 48, 48, 48], 16)
}

function selectPlace(p: Place, move: boolean) {
  const key = placeKey(p)
  selected.value = key
  if (move && map) map.setZoomAndCenter(Math.max(map.getZoom(), 15), [p.lng, p.lat])
  if (info && map) {
    info.setContent(
      `<div class="mbk-info"><strong>${escapeHtml(p.name || '未命名地点')}</strong>` +
      `<span>${kindLabel.value} ${yuan(p.amount)} · ${p.count} 笔</span>` +
      `<span>最近 ${dayjs(p.lastAt).format('YYYY年M月D日')}</span></div>`,
    )
    info.open(map, [p.lng, p.lat])
  }
  nextTick(() => {
    const el = listEl.value?.querySelector(`[data-key="${CSS.escape(key)}"]`)
    el?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
  })
}

function focusPlace(p: Place) {
  selectPlace(p, true)
}

async function load() {
  loading.value = true
  try {
    const { data: res } = await http.get('/stats/places', { params: statsParams(props.filter, { kind: kind.value }) })
    data.value = { ...res, places: res.places || [] }
    selected.value = ''
    renderMarkers()
  } finally {
    loading.value = false
  }
}

watch([kind, () => props.filter.from, () => props.filter.to, () => props.filter.accountId], load, { immediate: true })
watch(() => chart.value.dark, () => map?.setMapStyle(mapStyle()))

onMounted(initMap)
// 在 KeepAlive 里切回本页签时容器尺寸可能变了
onActivated(() => map?.resize?.())
onUnmounted(() => {
  map?.destroy()
  map = null
})
</script>

<style scoped>
.sm-controls { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 12px; margin-bottom: 10px; }
.sm-note { font-size: 0.74rem; color: var(--muted); }

.sm-grid { display: grid; gap: 12px; }
.sm-map { position: relative; overflow: hidden; padding: 0; }
.map { width: 100%; height: clamp(300px, 52vh, 460px); }
.map-overlay {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: min(86%, 360px);
  padding: 12px 16px;
  border-radius: 12px;
  background: rgba(21, 38, 31, 0.78);
  color: #fff;
  font-size: 0.84rem;
  line-height: 1.5;
  text-align: center;
  pointer-events: none;
}
.map-overlay--err { background: rgba(198, 40, 40, 0.9); }
.fit-btn { position: absolute; right: 10px; top: 10px; z-index: 5; }

.sm-list { padding: 6px; min-width: 0; display: flex; flex-direction: column; }
.list-head { padding: 6px 10px 8px; font-size: 0.78rem; color: var(--muted); }
.list-empty { color: var(--muted); text-align: center; padding: 20px 0; font-size: 0.85rem; }
.list-body { overflow-y: auto; max-height: 420px; }
.place {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 9px 10px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
}
.place:hover { background: var(--primary-soft); }
.place.active { background: var(--primary-soft); box-shadow: inset 3px 0 0 var(--primary); }
.rank {
  width: 22px;
  height: 22px;
  border-radius: 7px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.74rem;
  font-weight: 700;
  color: var(--muted);
  background: rgba(var(--v-theme-on-surface), 0.06);
  flex-shrink: 0;
}
.rank--hot { color: #fff; background: var(--primary); }
.place-main { flex: 1; min-width: 0; }
.place-line { display: flex; justify-content: space-between; gap: 8px; }
.place-name { font-size: 0.88rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.place-amt { font-size: 0.88rem; flex-shrink: 0; }
.place-sub { display: flex; gap: 6px; font-size: 0.72rem; color: var(--muted); margin-top: 2px; }
.place-pct { margin-left: auto; }

/* 平板 / 宽屏：地图在左、地点榜在右，列表与地图等高独立滚动 */
@media (min-width: 840px) {
  .sm-grid { grid-template-columns: minmax(0, 1fr) 320px; align-items: stretch; }
  .map { height: clamp(420px, 62vh, 600px); }
  .sm-list { max-height: clamp(420px, 62vh, 600px); }
  .list-body { max-height: none; flex: 1; min-height: 0; }
}
</style>

<style>
/* 高德自定义标记插在地图 DOM 里，scoped 样式够不到 */
.mbk-pin {
  width: var(--size);
  height: var(--size);
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: color-mix(in srgb, var(--pin) 82%, transparent);
  border: 2px solid #fff;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
  color: #fff;
  font: 600 11px/1 'Outfit', 'Noto Sans SC', sans-serif;
  white-space: nowrap;
  cursor: pointer;
  transition: transform 0.15s ease;
}
.mbk-pin:hover { transform: scale(1.12); }
.mbk-info { display: flex; flex-direction: column; gap: 3px; padding: 2px 4px; font-size: 12px; color: #15261f; max-width: 240px; }
.mbk-info strong { font-size: 13px; }
</style>
