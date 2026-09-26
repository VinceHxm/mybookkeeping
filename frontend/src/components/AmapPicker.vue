<template>
  <div class="picker">
    <header class="bar">
      <v-btn icon="mdi-close" variant="text" aria-label="关闭" @click="$emit('close')" />
      <div class="bar-center">
        <div class="bar-title">选择地点</div>
        <div class="bar-sub">{{ barSub }}</div>
      </div>
      <v-btn
        icon="mdi-crosshairs-gps"
        variant="tonal"
        color="primary"
        :loading="locating"
        :aria-label="mode === 'transit' ? '定位当前城市' : '定位到当前位置'"
        @click="mode === 'transit' ? resolveCityFromGps() : locateAndPick()"
      />
    </header>

    <div class="modes">
      <button
        type="button"
        class="mode-card"
        :class="{ active: mode === 'point' }"
        @click="onModeChange('point')"
      >
        <v-icon size="20">mdi-map-marker</v-icon>
        <div class="mode-text">
          <strong>单点</strong>
          <span>门店 / 当前位置</span>
        </div>
      </button>
      <button
        type="button"
        class="mode-card"
        :class="{ active: mode === 'route' }"
        @click="onModeChange('route')"
      >
        <v-icon size="20">mdi-map-marker-path</v-icon>
        <div class="mode-text">
          <strong>起终点</strong>
          <span>打车 / 路费</span>
        </div>
      </button>
      <button
        type="button"
        class="mode-card"
        :class="{ active: mode === 'transit' }"
        @click="onModeChange('transit')"
      >
        <v-icon size="20">mdi-subway-variant</v-icon>
        <div class="mode-text">
          <strong>公交地铁</strong>
          <span>搜线路选站</span>
        </div>
      </button>
    </div>

    <div v-if="mode !== 'transit'" class="search-box">
      <v-text-field
        v-model="searchKeyword"
        :label="searchLabel"
        variant="outlined"
        density="compact"
        hide-details
        clearable
        prepend-inner-icon="mdi-magnify"
        :loading="searching"
        @update:model-value="onSearchInput"
        @keyup.enter="runSearch"
      />
      <div v-if="searchTips.length" class="tips">
        <button
          v-for="(tip, i) in searchTips"
          :key="i"
          type="button"
          class="tip"
          @click="pickTip(tip)"
        >
          <strong>{{ tip.name }}</strong>
          <span>{{ tip.district || tip.address }}</span>
        </button>
      </div>
    </div>

    <div v-else class="search-box is-transit">
      <div class="city-row">
        <v-text-field
          v-model="cityInput"
          label="城市（必填）"
          placeholder="如 上海、杭州"
          variant="outlined"
          density="compact"
          hide-details="auto"
          :error="cityError"
          :error-messages="cityError ? '请先填写城市再搜线路' : ''"
          prepend-inner-icon="mdi-city-variant-outline"
          clearable
          class="city-field"
          @update:model-value="onCityEdit"
          @keyup.enter="commitCity"
        />
        <v-btn
          variant="tonal"
          color="primary"
          class="city-locate"
          :loading="locating"
          @click="resolveCityFromGps"
        >
          定位
        </v-btn>
      </div>
      <p v-if="!selectedLine" class="city-hint">
        {{ cityReady ? `已限定：${cityReady}` : '先标定城市，避免搜到外地线路、浪费配额' }}
      </p>
      <v-text-field
        v-if="!selectedLine"
        v-model="lineKeyword"
        label="搜线路名（如 2号线、301路）"
        variant="outlined"
        density="compact"
        hide-details
        clearable
        prepend-inner-icon="mdi-bus"
        :loading="lineSearching"
        :disabled="!cityReady"
        @keyup.enter="searchLines"
      >
        <template #append-inner>
          <v-btn
            size="small"
            variant="text"
            color="primary"
            :loading="lineSearching"
            :disabled="!cityReady"
            @click="searchLines"
          >搜索</v-btn>
        </template>
      </v-text-field>
      <div v-if="lineError" class="line-error">{{ lineError }}</div>

      <!-- 已选线路：收成一行，腾出地图空间 -->
      <div v-if="selectedLine" class="line-picked">
        <div class="line-picked-text">
          <strong>{{ selectedLine.name }}</strong>
          <span>{{ selectedLine.startStop }} → {{ selectedLine.endStop }}</span>
        </div>
        <v-btn size="small" variant="text" color="primary" @click="clearSelectedLine">更换</v-btn>
      </div>

      <div v-if="!selectedLine && lineResults.length" class="tips line-tips">
        <button
          v-for="(line, i) in lineResults"
          :key="i"
          type="button"
          class="tip"
          @click="selectLine(line)"
        >
          <strong>{{ line.name }}</strong>
          <span>{{ line.startStop }} → {{ line.endStop }}</span>
        </button>
      </div>
      <div v-if="selectedLine && transitStations.length" class="station-panel">
        <template v-if="transitStart && transitEnd">
          <div class="station-done">
            <div class="station-done-text">
              <strong>{{ transitStart.name }} → {{ transitEnd.name }}</strong>
              <span>上下站已选，可看地图确认</span>
            </div>
            <v-btn size="small" variant="text" color="primary" @click="resetStations">重选</v-btn>
          </div>
        </template>
        <template v-else>
          <div class="station-hint">点选上车站 → 下车站（列表可滚动）</div>
          <div class="station-list">
            <button
              v-for="(st, i) in transitStations"
              :key="i"
              type="button"
              class="station"
              :class="{
                start: transitStart?.name === st.name && transitStart?.index === i,
                end: transitEnd?.name === st.name && transitEnd?.index === i,
              }"
              @click="pickStation(st, i)"
            >
              <span class="st-idx">{{ i + 1 }}</span>
              <span class="st-name">{{ st.name }}</span>
            </button>
          </div>
        </template>
      </div>
    </div>

    <div v-if="recentPlaces.length && mode !== 'transit' && !searchTips.length" class="recent">
      <div class="recent-label">
        <v-icon size="14">mdi-history</v-icon>
        <span>最近</span>
      </div>
      <div class="recent-chips">
        <button
          v-for="(r, i) in recentPlaces"
          :key="i"
          type="button"
          class="recent-chip"
          :title="recentLabel(r)"
          @click="applyRecent(r)"
        >
          <v-icon v-if="r.mode === 'route'" size="13" class="recent-chip-ico">mdi-map-marker-path</v-icon>
          <span class="recent-chip-text">{{ recentShortLabel(r) }}</span>
        </button>
      </div>
    </div>

    <div v-if="mode === 'route'" class="route-steps">
      <button
        type="button"
        class="step"
        :class="{ active: routeStep === 'start', done: !!start }"
        @click="routeStep = 'start'"
      >
        <span class="step-dot start">起</span>
        <span class="step-label">{{ start ? '已选起点' : '选起点' }}</span>
      </button>
      <div class="step-line" aria-hidden="true" />
      <button
        type="button"
        class="step"
        :class="{ active: routeStep === 'end', done: !!end }"
        @click="routeStep = 'end'"
      >
        <span class="step-dot end">终</span>
        <span class="step-label">{{ end ? '已选终点' : '选终点' }}</span>
      </button>
    </div>

    <div v-if="mode === 'transit' && transitStart && !transitEnd" class="route-steps">
      <div class="step done">
        <span class="step-dot start">上</span>
        <span class="step-label">{{ transitStart.name }}</span>
      </div>
      <div class="step-line" aria-hidden="true" />
      <div class="step">
        <span class="step-dot end">下</span>
        <span class="step-label">请选下车站</span>
      </div>
    </div>

    <div class="map-wrap">
      <div ref="mapEl" class="map" />
      <div v-if="loadError" class="err-overlay">{{ loadError }}</div>
      <div v-else-if="locating && !hintPrimary" class="map-toast">正在定位…</div>
    </div>

    <footer class="sheet">
      <div class="sheet-info">
        <div class="sheet-kicker">{{ sheetKicker }}</div>
        <div class="sheet-primary">{{ hintPrimary }}</div>
        <div v-if="hintSecondary" class="sheet-secondary">{{ hintSecondary }}</div>
      </div>
      <v-btn
        color="primary"
        size="large"
        class="confirm-btn"
        :disabled="!canConfirm"
        @click="confirm"
      >
        {{ confirmLabel }}
      </v-btn>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import http from '../api/http'

export type GeoPickResult = {
  mode: 'point' | 'route'
  lng: number
  lat: number
  name: string
  endLng?: number
  endLat?: number
  endName?: string
}

type PlacePt = { lng: number; lat: number; name: string }
type SearchTip = { name: string; district?: string; address?: string; location?: string; id?: string }
type LineItem = {
  id: string
  name: string
  startStop: string
  endStop: string
  stations: PlacePt[]
}
type StationPick = PlacePt & { index: number }

const RECENT_KEY = 'mbk_recent_places'
const CITY_KEY = 'mbk_amap_city'
const RECENT_MAX = 8

const props = withDefaults(defineProps<{
  initialMode?: 'point' | 'route'
  initialLng?: number | null
  initialLat?: number | null
  initialName?: string
  initialEndLng?: number | null
  initialEndLat?: number | null
  initialEndName?: string
}>(), {
  initialMode: 'point',
})

const emit = defineEmits<{
  pick: [p: GeoPickResult]
  close: []
}>()

const mapEl = ref<HTMLElement | null>(null)
const mode = ref<'point' | 'route' | 'transit'>(props.initialMode === 'route' ? 'route' : 'point')
const routeStep = ref<'start' | 'end'>('start')
const picked = ref<PlacePt | null>(null)
const start = ref<PlacePt | null>(null)
const end = ref<PlacePt | null>(null)
const loadError = ref('')
const locating = ref(false)

const searchKeyword = ref('')
const searching = ref(false)
const searchTips = ref<SearchTip[]>([])
let searchTimer: ReturnType<typeof setTimeout> | null = null

const lineKeyword = ref('')
const lineSearching = ref(false)
const lineResults = ref<LineItem[]>([])
const lineError = ref('')
const selectedLine = ref<LineItem | null>(null)
const transitStations = computed(() => selectedLine.value?.stations || [])
const transitStart = ref<StationPick | null>(null)
const transitEnd = ref<StationPick | null>(null)
const recentPlaces = ref<GeoPickResult[]>([])
const cityInput = ref('')
const cityError = ref(false)

const cityReady = computed(() => normalizeCity(cityInput.value))

let map: any = null
let pointMarker: any = null
let startMarker: any = null
let endMarker: any = null
let routeLine: any = null
let AMapRef: any = null
let autoComplete: any = null

declare global {
  interface Window {
    _AMapSecurityConfig?: { securityJsCode: string }
    AMap?: any
  }
}

const canConfirm = computed(() => {
  if (mode.value === 'point') return !!picked.value
  if (mode.value === 'route') return !!(start.value && end.value)
  return !!(transitStart.value && transitEnd.value && transitStart.value.index !== transitEnd.value.index)
})

const barSub = computed(() => {
  if (mode.value === 'point') return '搜索或点地图选位置'
  if (mode.value === 'route') return '搜索或点选起终点'
  return cityReady.value ? `${cityReady.value} · 搜线路选站` : '先标定城市再搜线路'
})

function normalizeCity(raw: string) {
  return raw.trim().replace(/市$/, '') || ''
}

function persistCity(name: string) {
  const n = normalizeCity(name)
  if (!n) return
  cityInput.value = n
  cityError.value = false
  try {
    localStorage.setItem(CITY_KEY, n)
  } catch { /* ignore */ }
}

function loadCityFromStorage() {
  try {
    const saved = localStorage.getItem(CITY_KEY)
    if (saved) cityInput.value = normalizeCity(saved)
  } catch { /* ignore */ }
}

function onCityEdit() {
  cityError.value = false
  lineError.value = ''
  // 改城市后清空旧线路结果，避免误选外地线
  if (lineResults.value.length || selectedLine.value) {
    lineResults.value = []
    selectedLine.value = null
    transitStart.value = null
    transitEnd.value = null
    clearRouteVisual()
  }
}

function commitCity() {
  const n = normalizeCity(cityInput.value)
  if (!n) {
    cityError.value = true
    return
  }
  persistCity(n)
}

function extractCityFromGeoResult(result: any): string {
  const ac = result?.addressComponent
  const raw = ac?.city || ac?.province || result?.city || ''
  // 直辖市 city 可能是空数组/空串，用 province
  if (Array.isArray(raw) && !raw.length) {
    return normalizeCity(String(ac?.province || ''))
  }
  return normalizeCity(String(raw || ''))
}

const searchLabel = computed(() =>
  mode.value === 'route'
    ? (routeStep.value === 'start' ? '搜索起点' : '搜索终点')
    : '搜索地名 / 门店',
)

const sheetKicker = computed(() => {
  if (mode.value === 'point') return '已选位置'
  if (mode.value === 'transit') {
    if (transitStart.value && transitEnd.value) return selectedLine.value?.name || '行程'
    if (transitStart.value) return '已选上车站，请选下车站'
    return '请选择上车站'
  }
  if (start.value && end.value) return '行程'
  if (routeStep.value === 'start') return '当前：选择起点'
  return '当前：选择终点'
})

const hintPrimary = computed(() => {
  if (mode.value === 'point') {
    return picked.value?.name || (locating.value ? '正在获取当前位置…' : '搜索或在地图上点选')
  }
  if (mode.value === 'transit') {
    if (transitStart.value && transitEnd.value) return transitStart.value.name
    if (transitStart.value) return transitStart.value.name
    if (!cityReady.value) return '请先填写或定位当前城市'
    return '再搜索并选择线路'
  }
  if (start.value && end.value) return start.value.name
  if (start.value) return start.value.name
  return locating.value ? '正在获取当前位置作为起点…' : '先选起点'
})

const hintSecondary = computed(() => {
  if (mode.value === 'transit' && transitStart.value && transitEnd.value) {
    return `→ ${transitEnd.value.name}`
  }
  if (mode.value === 'route' && start.value && end.value) return `→ ${end.value.name}`
  if (mode.value === 'route' && start.value && !end.value) return '下一步：选终点'
  return ''
})

const confirmLabel = computed(() => {
  if (!canConfirm.value) {
    if (mode.value === 'point') return '请先选位置'
    if (mode.value === 'transit') return '请选齐上下车站'
    return '请选齐起终点'
  }
  return mode.value === 'point' ? '使用此位置' : '确认行程'
})

function loadScript(src: string) {
  return new Promise<void>((resolve, reject) => {
    if (document.querySelector(`script[src="${src}"]`)) {
      resolve()
      return
    }
    const s = document.createElement('script')
    s.src = src
    s.async = true
    s.onload = () => resolve()
    s.onerror = () => reject(new Error('高德脚本加载失败'))
    document.head.appendChild(s)
  })
}

function parseLngLat(loc: any): { lng: number; lat: number } | null {
  if (!loc) return null
  if (typeof loc.lng === 'number' && typeof loc.lat === 'number') return { lng: loc.lng, lat: loc.lat }
  if (typeof loc.getLng === 'function') return { lng: loc.getLng(), lat: loc.getLat() }
  if (typeof loc === 'string' && loc.includes(',')) {
    const [a, b] = loc.split(',')
    const lng = Number(a)
    const lat = Number(b)
    if (Number.isFinite(lng) && Number.isFinite(lat)) return { lng, lat }
  }
  return null
}

function reverseGeocode(lng: number, lat: number): Promise<string> {
  return new Promise((resolve) => {
    AMapRef.plugin('AMap.Geocoder', () => {
      const geocoder = new AMapRef.Geocoder()
      geocoder.getAddress([lng, lat], (status: string, result: any) => {
        const name =
          status === 'complete' && result.regeocode
            ? result.regeocode.formattedAddress
            : `${lng.toFixed(5)}, ${lat.toFixed(5)}`
        resolve(name)
      })
    })
  })
}

function ensureMarker(kind: 'point' | 'start' | 'end', lng: number, lat: number) {
  if (!map || !AMapRef) return
  const pos = [lng, lat]
  if (kind === 'point') {
    if (!pointMarker) {
      pointMarker = new AMapRef.Marker({ position: pos })
      map.add(pointMarker)
    } else {
      pointMarker.setPosition(pos)
    }
    return
  }
  const label = kind === 'start' ? '起' : '终'
  const color = kind === 'start' ? '#1b7f5a' : '#c62828'
  const labelHtml = {
    content: `<div style="background:${color};color:#fff;padding:2px 8px;border-radius:999px;font-size:12px;font-weight:600;box-shadow:0 2px 8px rgba(0,0,0,.2)">${label}</div>`,
    direction: 'top',
  }
  if (kind === 'start') {
    if (!startMarker) {
      startMarker = new AMapRef.Marker({ position: pos, label: labelHtml })
      map.add(startMarker)
    } else {
      startMarker.setPosition(pos)
      startMarker.setLabel(labelHtml)
    }
  } else if (!endMarker) {
    endMarker = new AMapRef.Marker({ position: pos, label: labelHtml })
    map.add(endMarker)
  } else {
    endMarker.setPosition(pos)
    endMarker.setLabel(labelHtml)
  }
}

function updateRouteLine() {
  if (!map || !AMapRef) return
  if (routeLine) {
    map.remove(routeLine)
    routeLine = null
  }
  const a = mode.value === 'transit' ? transitStart.value : start.value
  const b = mode.value === 'transit' ? transitEnd.value : end.value
  if (!a || !b) return
  routeLine = new AMapRef.Polyline({
    path: [
      [a.lng, a.lat],
      [b.lng, b.lat],
    ],
    strokeColor: '#1b7f5a',
    strokeWeight: 4,
    strokeOpacity: 0.85,
    strokeStyle: 'solid',
  })
  map.add(routeLine)
  const markers = [startMarker, endMarker, routeLine].filter(Boolean)
  if (markers.length) map.setFitView(markers, false, [60, 60, 60, 60])
}

function clearPointVisual() {
  if (pointMarker && map) {
    map.remove(pointMarker)
    pointMarker = null
  }
}

function clearRouteVisual() {
  if (!map) return
  if (startMarker) { map.remove(startMarker); startMarker = null }
  if (endMarker) { map.remove(endMarker); endMarker = null }
  if (routeLine) { map.remove(routeLine); routeLine = null }
}

async function setPoint(lng: number, lat: number, name?: string) {
  const resolved = name || await reverseGeocode(lng, lat)
  picked.value = { lng, lat, name: resolved }
  ensureMarker('point', lng, lat)
  map?.setCenter([lng, lat])
}

async function setRoutePoint(which: 'start' | 'end', lng: number, lat: number, name?: string) {
  const resolved = name || await reverseGeocode(lng, lat)
  const pt = { lng, lat, name: resolved }
  if (which === 'start') {
    start.value = pt
    ensureMarker('start', lng, lat)
    if (!end.value) routeStep.value = 'end'
  } else {
    end.value = pt
    ensureMarker('end', lng, lat)
  }
  updateRouteLine()
  if (!(start.value && end.value)) map?.setCenter([lng, lat])
}

async function locateAndPick() {
  if (!AMapRef || !map) return
  locating.value = true
  try {
    await new Promise<void>((resolve) => {
      AMapRef.plugin('AMap.Geolocation', () => {
        const geo = new AMapRef.Geolocation({
          enableHighAccuracy: true,
          timeout: 10000,
        })
        geo.getCurrentPosition(async (status: string, result: any) => {
          if (status === 'complete' && result.position) {
            const lng = result.position.lng
            const lat = result.position.lat
            const city = extractCityFromGeoResult(result)
            if (city) persistCity(city)
            if (mode.value === 'point') {
              await setPoint(lng, lat)
            } else if (mode.value === 'route') {
              await setRoutePoint(routeStep.value, lng, lat)
            }
          }
          resolve()
        })
      })
    })
  } finally {
    locating.value = false
  }
}

/** 仅解析城市：GPS → 失败再用 CitySearch（IP），不落点 */
async function resolveCityFromGps() {
  if (!AMapRef) return
  locating.value = true
  lineError.value = ''
  try {
    const fromGps = await new Promise<string>((resolve) => {
      AMapRef.plugin('AMap.Geolocation', () => {
        const geo = new AMapRef.Geolocation({
          enableHighAccuracy: true,
          timeout: 10000,
        })
        geo.getCurrentPosition((status: string, result: any) => {
          if (status === 'complete') {
            resolve(extractCityFromGeoResult(result))
          } else {
            resolve('')
          }
        })
      })
    })
    if (fromGps) {
      persistCity(fromGps)
      return
    }
    const fromIp = await new Promise<string>((resolve) => {
      AMapRef.plugin('AMap.CitySearch', () => {
        const cs = new AMapRef.CitySearch()
        cs.getLocalCity((status: string, result: any) => {
          if (status === 'complete' && result?.city) {
            resolve(normalizeCity(String(result.city)))
          } else {
            resolve('')
          }
        })
      })
    })
    if (fromIp) {
      persistCity(fromIp)
      return
    }
    cityError.value = true
    lineError.value = '未能自动定位城市，请手动填写'
  } finally {
    locating.value = false
  }
}

function onSearchInput() {
  searchTips.value = []
  if (searchTimer) clearTimeout(searchTimer)
  const q = searchKeyword.value.trim()
  if (q.length < 2) return
  searchTimer = setTimeout(() => runSearch(), 400)
}

function runSearch() {
  const q = searchKeyword.value.trim()
  if (!q || !AMapRef) return
  searching.value = true
  const city = cityReady.value || '全国'
  const pluginName = AMapRef.AutoComplete ? 'AMap.AutoComplete' : 'AMap.Autocomplete'
  AMapRef.plugin(pluginName, () => {
    const Ctor = AMapRef.AutoComplete || AMapRef.Autocomplete
    if (!autoComplete) {
      autoComplete = new Ctor({ city, citylimit: !!cityReady.value })
    } else {
      if (typeof autoComplete.setCity === 'function') autoComplete.setCity(city)
      if (typeof autoComplete.setCityLimit === 'function') autoComplete.setCityLimit(!!cityReady.value)
    }
    autoComplete.search(q, (status: string, result: any) => {
      searching.value = false
      if (status !== 'complete' || !result?.tips) {
        searchTips.value = []
        return
      }
      searchTips.value = (result.tips as any[])
        .filter((t) => t.name && t.location)
        .slice(0, 8)
        .map((t) => ({
          name: t.name,
          district: t.district,
          address: t.address,
          location: typeof t.location === 'string'
            ? t.location
            : (t.location ? `${t.location.lng},${t.location.lat}` : undefined),
          id: t.id,
        }))
    })
  })
}

async function pickTip(tip: SearchTip) {
  const loc = parseLngLat(tip.location)
  if (!loc) return
  searchTips.value = []
  searchKeyword.value = tip.name
  if (mode.value === 'point') {
    await setPoint(loc.lng, loc.lat, tip.name)
  } else {
    await setRoutePoint(routeStep.value, loc.lng, loc.lat, tip.name)
  }
}

function searchLines() {
  const city = cityReady.value
  if (!city) {
    cityError.value = true
    lineError.value = '请先填写或定位城市'
    return
  }
  const q = lineKeyword.value.trim()
  if (!q || !AMapRef) return
  persistCity(city)
  lineError.value = ''
  lineSearching.value = true
  lineResults.value = []
  selectedLine.value = null
  transitStart.value = null
  transitEnd.value = null
  clearRouteVisual()
  AMapRef.plugin('AMap.LineSearch', () => {
    const linesearch = new AMapRef.LineSearch({
      pageIndex: 1,
      pageSize: 10,
      city,
      extensions: 'all',
    })
    linesearch.search(q, (status: string, result: any) => {
      lineSearching.value = false
      if (status !== 'complete' || !result?.lineInfo?.length) {
        lineResults.value = []
        lineError.value = status === 'no_data' ? `在「${city}」未找到该线路` : '线路搜索失败，请换个关键词'
        return
      }
      lineResults.value = (result.lineInfo as any[]).map((line) => {
        const stops = (line.via_stops || line.viaStops || []) as any[]
        const stations: PlacePt[] = []
        for (const s of stops) {
          const loc = parseLngLat(s.location)
          if (!loc || !s.name) continue
          stations.push({ lng: loc.lng, lat: loc.lat, name: String(s.name) })
        }
        return {
          id: String(line.id || line.name),
          name: String(line.name || q),
          startStop: String(line.start_stop || line.startStop || stations[0]?.name || ''),
          endStop: String(line.end_stop || line.endStop || stations[stations.length - 1]?.name || ''),
          stations,
        }
      }).filter((l) => l.stations.length >= 2)
    })
  })
}

function selectLine(line: LineItem) {
  selectedLine.value = line
  transitStart.value = null
  transitEnd.value = null
  clearRouteVisual()
  if (line.stations.length && map) {
    const mid = line.stations[Math.floor(line.stations.length / 2)]
    map.setCenter([mid.lng, mid.lat])
    map.setZoom(12)
  }
}

function clearSelectedLine() {
  selectedLine.value = null
  transitStart.value = null
  transitEnd.value = null
  clearRouteVisual()
}

function resetStations() {
  transitStart.value = null
  transitEnd.value = null
  clearRouteVisual()
}

function pickStation(st: PlacePt, index: number) {
  const pick: StationPick = { ...st, index }
  if (!transitStart.value || (transitStart.value && transitEnd.value)) {
    transitStart.value = pick
    transitEnd.value = null
    clearRouteVisual()
    ensureMarker('start', st.lng, st.lat)
    map?.setCenter([st.lng, st.lat])
    return
  }
  if (index === transitStart.value.index) return
  transitEnd.value = pick
  ensureMarker('end', st.lng, st.lat)
  updateRouteLine()
}

function recentLabel(r: GeoPickResult) {
  if (r.mode === 'route' && r.endName) return `${r.name} → ${r.endName}`
  return r.name
}

/** 去掉「福建省南平市顺昌县」这类行政区前缀，横滑 chip 里只留有辨识度的部分 */
function shortPlace(name: string) {
  const s = (name || '').trim()
  const province = s.match(/^([^()（）]{2,8}?(省|自治区|特别行政区)|北京市|上海市|天津市|重庆市)/)
  if (!province) return s
  let rest = s.slice(province[0].length)
  if (/(省|自治区|特别行政区)$/.test(province[0])) {
    rest = rest.replace(/^[^()（）]{2,10}?(市|自治州|地区|盟)/, '')
  }
  rest = rest.replace(/^[^()（）]{1,10}?(区|县|市|旗)/, '')
  return rest.length >= 2 ? rest : s
}

function recentShortLabel(r: GeoPickResult) {
  if (r.mode === 'route' && r.endName) return `${shortPlace(r.name)} → ${shortPlace(r.endName)}`
  return shortPlace(r.name)
}

async function applyRecent(r: GeoPickResult) {
  if (r.mode === 'route' && r.endLng != null && r.endLat != null) {
    mode.value = 'route'
    clearPointVisual()
    await setRoutePoint('start', r.lng, r.lat, r.name)
    await setRoutePoint('end', r.endLng, r.endLat, r.endName)
  } else {
    mode.value = 'point'
    clearRouteVisual()
    await setPoint(r.lng, r.lat, r.name)
  }
}

function loadRecentFromStorage() {
  try {
    const raw = localStorage.getItem(RECENT_KEY)
    if (!raw) return
    const arr = JSON.parse(raw)
    if (Array.isArray(arr)) recentPlaces.value = arr.slice(0, RECENT_MAX)
  } catch { /* ignore */ }
}

function saveRecent(p: GeoPickResult) {
  const key = p.mode === 'route'
    ? `r:${p.name}|${p.endName}|${p.lng},${p.lat}`
    : `p:${p.name}|${p.lng},${p.lat}`
  const next = [p, ...recentPlaces.value.filter((x) => {
    const k = x.mode === 'route'
      ? `r:${x.name}|${x.endName}|${x.lng},${x.lat}`
      : `p:${x.name}|${x.lng},${x.lat}`
    return k !== key
  })].slice(0, RECENT_MAX)
  recentPlaces.value = next
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(next))
  } catch { /* ignore */ }
}

async function loadRecentFromTransactions() {
  try {
    const { data } = await http.get('/transactions', { params: { limit: 40 } })
    const items = data?.items || data || []
    const found: GeoPickResult[] = []
    const seen = new Set<string>()
    for (const t of items) {
      if (t.geoLng == null || t.geoLat == null || !t.geoName) continue
      const isRoute = t.geoMode === 'route' && t.geoEndLng != null && t.geoEndName
      const p: GeoPickResult = isRoute
        ? {
            mode: 'route',
            lng: t.geoLng,
            lat: t.geoLat,
            name: t.geoName,
            endLng: t.geoEndLng,
            endLat: t.geoEndLat,
            endName: t.geoEndName,
          }
        : { mode: 'point', lng: t.geoLng, lat: t.geoLat, name: t.geoName }
      const k = recentLabel(p)
      if (seen.has(k)) continue
      seen.add(k)
      found.push(p)
      if (found.length >= RECENT_MAX) break
    }
    if (!found.length) return
    const merged = [...found]
    for (const r of recentPlaces.value) {
      const k = recentLabel(r)
      if (!seen.has(k)) {
        seen.add(k)
        merged.push(r)
      }
    }
    recentPlaces.value = merged.slice(0, RECENT_MAX)
  } catch { /* ignore */ }
}

function onModeChange(m: 'point' | 'route' | 'transit') {
  if (m === mode.value) return
  mode.value = m
  searchTips.value = []
  searchKeyword.value = ''
  if (m === 'point') {
    clearRouteVisual()
    if (picked.value) ensureMarker('point', picked.value.lng, picked.value.lat)
    else if (map) locateAndPick()
  } else if (m === 'route') {
    clearPointVisual()
    routeStep.value = start.value ? 'end' : 'start'
    if (start.value) ensureMarker('start', start.value.lng, start.value.lat)
    if (end.value) ensureMarker('end', end.value.lng, end.value.lat)
    updateRouteLine()
    if (!start.value && map) locateAndPick()
  } else {
    clearPointVisual()
    clearRouteVisual()
    cityError.value = false
    lineError.value = ''
    if (transitStart.value) ensureMarker('start', transitStart.value.lng, transitStart.value.lat)
    if (transitEnd.value) ensureMarker('end', transitEnd.value.lng, transitEnd.value.lat)
    updateRouteLine()
    // 无城市时尝试自动定位，失败则等用户手填
    if (!cityReady.value && map) resolveCityFromGps()
  }
}

onMounted(async () => {
  loadCityFromStorage()
  loadRecentFromStorage()
  loadRecentFromTransactions()
  try {
    const { data } = await http.get('/amap/config')
    window._AMapSecurityConfig = { securityJsCode: data.securityJsCode }
    await loadScript(`https://webapi.amap.com/maps?v=2.0&key=${data.key}`)
    AMapRef = window.AMap
    map = new AMapRef.Map(mapEl.value, {
      zoom: 15,
      center: [116.397428, 39.90923],
    })
    map.on('click', async (e: any) => {
      if (mode.value === 'transit') return
      const lng = e.lnglat.getLng()
      const lat = e.lnglat.getLat()
      if (mode.value === 'point') {
        await setPoint(lng, lat)
      } else {
        await setRoutePoint(routeStep.value, lng, lat)
      }
    })

    if (props.initialLng != null && props.initialLat != null) {
      if (props.initialMode === 'route' && props.initialEndLng != null && props.initialEndLat != null) {
        mode.value = 'route'
        await setRoutePoint('start', props.initialLng, props.initialLat, props.initialName)
        await setRoutePoint('end', props.initialEndLng, props.initialEndLat, props.initialEndName)
      } else {
        mode.value = 'point'
        await setPoint(props.initialLng, props.initialLat, props.initialName)
      }
    } else {
      await locateAndPick()
    }
  } catch (e: any) {
    loadError.value = e.message || '地图加载失败'
  }
})

onUnmounted(() => {
  if (searchTimer) clearTimeout(searchTimer)
  if (map) map.destroy()
})

function confirm() {
  let result: GeoPickResult | null = null
  if (mode.value === 'point' && picked.value) {
    result = {
      mode: 'point',
      lng: picked.value.lng,
      lat: picked.value.lat,
      name: picked.value.name,
    }
  } else if (mode.value === 'route' && start.value && end.value) {
    result = {
      mode: 'route',
      lng: start.value.lng,
      lat: start.value.lat,
      name: start.value.name,
      endLng: end.value.lng,
      endLat: end.value.lat,
      endName: end.value.name,
    }
  } else if (mode.value === 'transit' && transitStart.value && transitEnd.value) {
    const lineTag = selectedLine.value?.name ? `（${selectedLine.value.name}）` : ''
    result = {
      mode: 'route',
      lng: transitStart.value.lng,
      lat: transitStart.value.lat,
      name: `${transitStart.value.name}${lineTag}`,
      endLng: transitEnd.value.lng,
      endLat: transitEnd.value.lat,
      endName: transitEnd.value.name,
    }
  }
  if (!result) return
  saveRecent(result)
  emit('pick', result)
}
</script>

<style scoped>
.picker {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg, #eef5f0);
  color: var(--ink, #15261f);
}

.bar {
  display: grid;
  grid-template-columns: 48px 1fr 48px;
  align-items: center;
  gap: 4px;
  padding: 8px 8px calc(8px + env(safe-area-inset-top, 0px));
  padding-top: max(8px, env(safe-area-inset-top, 0px));
  background: var(--surface-solid, #fff);
  border-bottom: 1px solid var(--surface-border, #e6ece8);
}

.bar-center { text-align: center; min-width: 0; }
.bar-title { font-weight: 700; font-size: 1.05rem; line-height: 1.2; }
.bar-sub { color: var(--muted, #5a6f64); font-size: 0.75rem; margin-top: 2px; }

.modes {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 8px;
  padding: 10px 12px;
  background: var(--surface-solid, #fff);
  border-bottom: 1px solid var(--surface-border, #e6ece8);
}

.mode-card {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 10px 8px;
  border-radius: 14px;
  border: 1.5px solid var(--surface-border, #e6ece8);
  background: var(--surface, #fff);
  color: inherit;
  text-align: left;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  transition: border-color 0.15s ease, background 0.15s ease, box-shadow 0.15s ease;
}

.mode-card.active {
  border-color: var(--primary, #1b7f5a);
  background: var(--primary-soft, rgba(27, 127, 90, 0.12));
  box-shadow: 0 0 0 1px var(--primary, #1b7f5a);
}

.mode-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.mode-text strong { font-size: 0.82rem; line-height: 1.2; }
.mode-text span { font-size: 0.68rem; color: var(--muted, #5a6f64); line-height: 1.3; }

.search-box {
  padding: 10px 12px 8px;
  background: var(--surface-solid, #fff);
  border-bottom: 1px solid var(--surface-border, #e6ece8);
  flex-shrink: 0;
}

/* 公交地铁：上方控件可自滚动，绝不吃掉地图 */
.search-box.is-transit {
  flex: 0 1 auto;
  max-height: min(38vh, 280px);
  overflow-y: auto;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
  min-height: 0;
}

.city-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}
.city-field { flex: 1; min-width: 0; }
.city-locate {
  flex-shrink: 0;
  margin-top: 2px;
  min-height: 40px;
}
.city-hint {
  margin: 4px 0 8px;
  font-size: 0.72rem;
  color: var(--muted, #5a6f64);
  line-height: 1.35;
}
.line-error {
  margin: 6px 0 4px;
  font-size: 0.78rem;
  color: #c62828;
}

.line-picked {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 6px 0 8px;
  padding: 8px 10px;
  border-radius: 12px;
  border: 1px solid var(--primary, #1b7f5a);
  background: var(--primary-soft, rgba(27, 127, 90, 0.12));
}
.line-picked-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.line-picked-text strong {
  font-size: 0.82rem;
  line-height: 1.25;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.line-picked-text span {
  font-size: 0.68rem;
  color: var(--muted, #5a6f64);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tips {
  max-height: min(22vh, 120px);
  overflow: auto;
  margin-top: 6px;
  margin-bottom: 4px;
  border-radius: 12px;
  border: 1px solid var(--surface-border, #e6ece8);
}
.tips.line-tips {
  max-height: min(24vh, 132px);
}

.tip {
  display: flex;
  flex-direction: column;
  gap: 2px;
  width: 100%;
  padding: 8px 10px;
  border: 0;
  border-bottom: 1px solid var(--surface-border, #e6ece8);
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
}
.tip:last-child { border-bottom: 0; }
.tip:hover, .tip.active { background: var(--primary-soft, rgba(27, 127, 90, 0.1)); }
.tip strong { font-size: 0.82rem; }
.tip span { font-size: 0.68rem; color: var(--muted, #5a6f64); }

.station-panel {
  margin: 4px 0 0;
  border-radius: 12px;
  border: 1px solid var(--surface-border, #e6ece8);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.station-hint {
  padding: 6px 10px;
  font-size: 0.7rem;
  color: var(--muted, #5a6f64);
  background: var(--primary-soft, rgba(27, 127, 90, 0.08));
  flex-shrink: 0;
}
.station-done {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
}
.station-done-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.station-done-text strong {
  font-size: 0.82rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.station-done-text span {
  font-size: 0.68rem;
  color: var(--muted, #5a6f64);
}
.station-list {
  max-height: min(22vh, 140px);
  overflow: auto;
  -webkit-overflow-scrolling: touch;
}
.station {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 7px 10px;
  border: 0;
  border-bottom: 1px solid var(--surface-border, #e6ece8);
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
}
.station:last-child { border-bottom: 0; }
.station.start { background: rgba(27, 127, 90, 0.14); }
.station.end { background: rgba(196, 92, 62, 0.14); }
.st-idx {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  font-size: 0.68rem;
  font-weight: 700;
  background: var(--surface-border, #e6ece8);
  color: var(--muted, #5a6f64);
  flex-shrink: 0;
}
.station.start .st-idx { background: #1b7f5a; color: #fff; }
.station.end .st-idx { background: #c45c3e; color: #fff; }
.st-name { font-size: 0.84rem; font-weight: 600; }

.recent {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0 6px 12px;
  background: var(--surface-solid, #fff);
  border-bottom: 1px solid var(--surface-border, #e6ece8);
  flex-shrink: 0;
}
.recent-label {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--muted, #5a6f64);
}
.recent-chips {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-wrap: nowrap;
  gap: 6px;
  overflow-x: auto;
  overscroll-behavior-x: contain;
  -webkit-overflow-scrolling: touch;
  scroll-snap-type: x proximity;
  scrollbar-width: none;
  padding-right: 12px;
  -webkit-mask-image: linear-gradient(90deg, #000 calc(100% - 20px), transparent);
  mask-image: linear-gradient(90deg, #000 calc(100% - 20px), transparent);
}
.recent-chips::-webkit-scrollbar { display: none; }
.recent-chip {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  max-width: min(62vw, 240px);
  height: 30px;
  padding: 0 11px;
  border-radius: 999px;
  border: 1px solid var(--surface-border, #e6ece8);
  background: var(--primary-soft, rgba(27, 127, 90, 0.08));
  color: inherit;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  scroll-snap-align: start;
  -webkit-tap-highlight-color: transparent;
}
.recent-chip:active { transform: scale(0.97); }
.recent-chip-ico { color: var(--primary, #1b7f5a); flex-shrink: 0; }
.recent-chip-text {
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.route-steps {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: var(--surface-solid, #fff);
  border-bottom: 1px solid var(--surface-border, #e6ece8);
}

.step {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 999px;
  border: 1px solid var(--surface-border, #e6ece8);
  background: transparent;
  color: inherit;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  min-width: 0;
}
.step.active {
  border-color: var(--primary, #1b7f5a);
  background: var(--primary-soft, rgba(27, 127, 90, 0.12));
}
.step.done:not(.active) { opacity: 0.85; }

.step-dot {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: #fff;
  font-size: 0.72rem;
  font-weight: 700;
  flex-shrink: 0;
}
.step-dot.start { background: #1b7f5a; }
.step-dot.end { background: #c45c3e; }
.step-label {
  font-size: 0.82rem;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.step-line {
  flex: 1;
  height: 2px;
  background: linear-gradient(90deg, #1b7f5a, #c45c3e);
  opacity: 0.35;
  border-radius: 2px;
  min-width: 12px;
}

.map-wrap {
  position: relative;
  flex: 1 1 auto;
  min-height: 160px;
}
.map { width: 100%; height: 100%; }

/* 小屏：再压紧控件，地图至少约 1/4 屏 */
@media (max-width: 600px) {
  .modes {
    gap: 6px;
    padding: 8px 10px;
  }
  .mode-card {
    padding: 7px 6px;
    align-items: center;
    justify-content: center;
    border-radius: 12px;
  }
  .mode-text span { display: none; }
  .search-box { padding: 8px 12px 6px; }
  .search-box.is-transit {
    max-height: min(34vh, 240px);
  }
  .tips,
  .tips.line-tips {
    max-height: min(18vh, 100px);
  }
  .station-list {
    max-height: min(18vh, 120px);
  }
  .map-wrap {
    min-height: max(140px, 26vh);
  }
  .route-steps {
    padding: 8px 10px;
  }
}

.map-toast,
.err-overlay {
  position: absolute;
  left: 50%;
  top: 16px;
  transform: translateX(-50%);
  z-index: 2;
  max-width: min(90%, 420px);
  padding: 8px 14px;
  border-radius: 999px;
  background: rgba(21, 38, 31, 0.82);
  color: #fff;
  font-size: 0.85rem;
  text-align: center;
}
.err-overlay { background: rgba(198, 40, 40, 0.92); border-radius: 12px; }

.sheet {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px 16px calc(14px + env(safe-area-inset-bottom, 0px));
  background: var(--surface-solid, #fff);
  border-top: 1px solid var(--surface-border, #e6ece8);
  box-shadow: 0 -8px 28px rgba(21, 38, 31, 0.08);
}

.sheet-kicker {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--muted, #5a6f64);
  letter-spacing: 0.02em;
}
.sheet-primary {
  margin-top: 4px;
  font-size: 0.98rem;
  font-weight: 700;
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.sheet-secondary {
  margin-top: 4px;
  color: var(--muted, #5a6f64);
  font-size: 0.86rem;
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.confirm-btn {
  width: 100%;
  min-height: 48px;
  font-weight: 700 !important;
  letter-spacing: 0.04em;
}

@media (min-width: 960px) {
  .picker {
    max-width: 920px;
    margin: 0 auto;
    border-left: 1px solid var(--surface-border, #e6ece8);
    border-right: 1px solid var(--surface-border, #e6ece8);
  }
  .sheet {
    flex-direction: row;
    align-items: center;
  }
  .sheet-info { flex: 1; min-width: 0; }
  .confirm-btn {
    width: auto;
    min-width: 180px;
    flex-shrink: 0;
  }
}

/* 小屏底栏左右排布：信息在左、按钮在右，省出一整行给地图 */
@media (max-width: 600px) {
  .sheet {
    flex-direction: row;
    align-items: center;
    gap: 10px;
    padding: 10px 12px calc(10px + env(safe-area-inset-bottom, 0px));
  }
  .sheet-info { flex: 1; min-width: 0; }
  .sheet-kicker { font-size: 0.7rem; }
  .sheet-primary { margin-top: 2px; font-size: 0.9rem; }
  .sheet-secondary { margin-top: 2px; font-size: 0.8rem; -webkit-line-clamp: 1; }
  .confirm-btn {
    width: auto;
    min-width: 108px;
    max-width: 40%;
    min-height: 44px;
    flex-shrink: 0;
    padding: 0 14px !important;
    letter-spacing: 0.02em;
  }
}
</style>
