<template>
  <section class="form-section mb-4">
    <div class="form-section__body">
      <MetaHead
        :icon="geo.mode === 'route' ? 'mdi-map-marker-path' : 'mdi-map-marker-outline'"
        :title="t('location.title')"
        :subtitle="t('location.subtitle')"
        class="mb-3"
      >
        <v-chip v-if="hasGeo" size="small" color="primary" variant="tonal">
          {{ geo.mode === 'route' ? t('location.route') : t('location.point') }}
        </v-chip>
      </MetaHead>

      <button v-if="!hasGeo" type="button" class="loc-empty" @click="showMap = true">
        <v-icon size="28" color="primary">mdi-map-search-outline</v-icon>
        <div class="loc-empty-text">
          <strong>{{ t('location.add') }}</strong>
          <span>{{ t('location.addSub') }}</span>
        </div>
        <v-icon size="20" color="primary">mdi-chevron-right</v-icon>
      </button>

      <div v-else class="loc-filled">
        <div class="loc-preview" @click="showMap = true">
          <div class="loc-kicker">{{ geo.mode === 'route' ? t('location.pickedRoute') : t('location.pickedPoint') }}</div>
          <div class="loc-main">{{ geo.name }}</div>
          <div v-if="geo.mode === 'route' && geo.endName" class="loc-second">→ {{ geo.endName }}</div>
        </div>
        <div class="loc-actions">
          <v-btn variant="tonal" color="primary" @click="showMap = true">
            <v-icon start>mdi-pencil-outline</v-icon>
            {{ t('location.edit') }}
          </v-btn>
          <v-btn variant="text" color="error" @click="clear">
            <v-icon start>mdi-close</v-icon>
            {{ t('location.clear') }}
          </v-btn>
        </div>
      </div>
    </div>

    <v-dialog
      v-model="showMap"
      :fullscreen="mapFullscreen"
      :max-width="mapFullscreen ? undefined : 960"
      transition="dialog-bottom-transition"
    >
      <div class="map-dialog-shell" :class="{ 'is-full': mapFullscreen }">
        <AmapPicker
          v-if="showMap"
          :initial-mode="geo.mode === 'route' ? 'route' : 'point'"
          :initial-lng="geo.lng"
          :initial-lat="geo.lat"
          :initial-name="geo.name"
          :initial-end-lng="geo.endLng"
          :initial-end-lat="geo.endLat"
          :initial-end-name="geo.endName"
          @pick="onPick"
          @close="showMap = false"
        />
      </div>
    </v-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AmapPicker, { type GeoPickResult } from './AmapPicker.vue'
import MetaHead from './MetaHead.vue'

const { t } = useI18n()

export interface GeoValue {
  mode: 'point' | 'route' | ''
  lng: number | null
  lat: number | null
  name: string
  endLng: number | null
  endLat: number | null
  endName: string
}

const geo = defineModel<GeoValue>('geo', { required: true })

const showMap = ref(false)
const mapFullscreen = ref(typeof window === 'undefined' ? true : window.innerWidth < 960)

const hasGeo = computed(() => !!geo.value.name)

function onViewport() {
  mapFullscreen.value = window.innerWidth < 960
}
onMounted(() => {
  window.addEventListener('resize', onViewport)
  onViewport()
})
onUnmounted(() => window.removeEventListener('resize', onViewport))

function clear() {
  geo.value = { mode: '', lng: null, lat: null, name: '', endLng: null, endLat: null, endName: '' }
}

function onPick(p: GeoPickResult) {
  const route = p.mode === 'route'
  geo.value = {
    mode: p.mode,
    lng: p.lng,
    lat: p.lat,
    name: p.name,
    endLng: route ? p.endLng ?? null : null,
    endLat: route ? p.endLat ?? null : null,
    endName: route ? p.endName || '' : '',
  }
  showMap.value = false
}
</script>

<style scoped>
.loc-empty {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 12px;
  border-radius: 14px;
  border: 1.5px dashed rgba(27, 127, 90, 0.45);
  background: var(--primary-soft);
  color: inherit;
  text-align: left;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  min-width: 0;
}
.loc-empty-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.loc-empty-text strong { font-size: 0.95rem; overflow-wrap: anywhere; }
.loc-empty-text span {
  font-size: 0.78rem;
  color: var(--muted);
  overflow-wrap: anywhere;
  word-break: break-word;
}

.loc-filled { display: flex; flex-direction: column; gap: 12px; }
.loc-preview {
  padding: 12px 14px;
  border-radius: 12px;
  background: var(--primary-soft);
  border: 1px solid rgba(27, 127, 90, 0.22);
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}
.loc-kicker {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--muted);
  margin-bottom: 4px;
}
.loc-main {
  font-weight: 700;
  font-size: 0.95rem;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.loc-second {
  margin-top: 4px;
  color: var(--muted);
  font-size: 0.86rem;
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.loc-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

@media (min-width: 960px) {
  .loc-filled {
    flex-direction: row;
    align-items: stretch;
  }
  .loc-preview { flex: 1; }
  .loc-actions {
    flex-direction: column;
    justify-content: center;
    min-width: 120px;
  }
}
</style>

<style>
.map-dialog-shell {
  height: min(85vh, 820px);
  overflow: hidden;
  border-radius: 18px;
  background: var(--bg, #eef5f0);
}
.map-dialog-shell.is-full {
  height: 100%;
  min-height: var(--app-height, 100dvh);
  border-radius: 0;
}
.map-dialog-shell .picker {
  height: 100%;
}
</style>
