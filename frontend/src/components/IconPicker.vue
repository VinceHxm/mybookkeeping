<template>
  <div class="icon-picker">
    <button type="button" class="current" @click.stop="open = !open">
      <v-avatar color="primary" variant="tonal" size="40" rounded="lg">
        <v-icon :icon="model || fallback" />
      </v-avatar>
      <div class="current-text">
        <strong>{{ title || t('icons.defaultTitle') }}</strong>
        <span>{{ iconLabel(model || fallback) }}</span>
      </div>
      <v-icon>{{ open ? 'mdi-chevron-up' : 'mdi-chevron-down' }}</v-icon>
    </button>

    <div v-if="open" class="panel" @click.stop>
      <v-text-field
        v-model="keyword"
        density="compact"
        variant="outlined"
        hide-details
        clearable
        prepend-inner-icon="mdi-magnify"
        :placeholder="t('icons.search')"
        class="mb-3"
      />
      <div class="chips hide-scrollbar">
        <button
          v-for="g in groups"
          :key="g.key"
          type="button"
          class="chip"
          :class="{ active: group === g.key }"
          @click="group = g.key"
        >{{ g.label }}</button>
      </div>
      <div class="grid hide-scrollbar">
        <button
          v-for="item in filtered"
          :key="item.value"
          type="button"
          class="cell"
          :class="{ active: (model || fallback) === item.value }"
          :title="item.label"
          @click="pick(item.value)"
        >
          <v-icon :icon="item.value" size="22" />
          <span>{{ item.label }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { APP_ICONS, ICON_GROUPS, iconGroupLabel, iconLabel, type IconGroup } from '../utils/icons'

withDefaults(defineProps<{
  fallback?: string
  title?: string
}>(), {
  fallback: 'mdi-shape',
  title: '',
})

const { t } = useI18n()
const model = defineModel<string>({ default: '' })

const open = ref(false)
const keyword = ref('')
const group = ref<IconGroup | 'all'>('all')

const groups = computed(() => [
  { key: 'all' as const, label: t('common.all') },
  ...ICON_GROUPS.map((g) => ({ key: g, label: iconGroupLabel(g) })),
])

const items = computed(() =>
  APP_ICONS.map((i) => ({ ...i, label: iconLabel(i.value), groupLabel: iconGroupLabel(i.group) })),
)

const filtered = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  return items.value.filter((i) => {
    if (group.value !== 'all' && i.group !== group.value) return false
    if (!q) return true
    return i.label.toLowerCase().includes(q) || i.value.toLowerCase().includes(q) || i.groupLabel.toLowerCase().includes(q)
  })
})

function pick(v: string) {
  model.value = v
}
</script>

<style scoped>
.icon-picker { margin-bottom: 14px; }

.current {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 10px 12px;
  border-radius: 14px;
  border: 1px solid var(--surface-border);
  background: var(--primary-soft);
  cursor: pointer;
  color: inherit;
  text-align: left;
}
.current-text { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.current-text strong { font-size: 0.9rem; }
.current-text span { font-size: 0.78rem; color: var(--muted); }

.panel {
  margin-top: 10px;
  padding: 12px;
  border-radius: 14px;
  border: 1px solid var(--surface-border);
  background: rgb(var(--v-theme-surface));
}

.chips {
  display: flex;
  gap: 6px;
  overflow-x: auto;
  padding-bottom: 10px;
  margin-bottom: 4px;
}
.chip {
  flex: 0 0 auto;
  border: 1px solid var(--surface-border);
  background: transparent;
  color: var(--muted);
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 0.78rem;
  cursor: pointer;
}
.chip.active {
  background: var(--primary-soft);
  border-color: var(--primary);
  color: var(--primary);
  font-weight: 700;
}

.grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
  max-height: 220px;
  overflow-y: auto;
}
@media (min-width: 520px) {
  .grid { grid-template-columns: repeat(5, minmax(0, 1fr)); max-height: 260px; }
}

.cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 10px 4px;
  border-radius: 12px;
  border: 1px solid transparent;
  background: var(--primary-soft);
  color: inherit;
  cursor: pointer;
  min-width: 0;
}
.cell span {
  font-size: 0.68rem;
  color: var(--muted);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.cell.active {
  border-color: var(--primary);
  box-shadow: 0 0 0 1px var(--primary);
  color: var(--primary);
}
.cell.active span { color: var(--primary); font-weight: 600; }
</style>
