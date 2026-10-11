<template>
  <div class="form-sections">
    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('templates.form.basic') }}</h3>
      </header>
      <div class="form-section__body">
        <IconPicker v-model="form.icon" :title="t('templates.form.iconTitle')" fallback="mdi-flash" />
        <v-text-field v-model="form.name" :label="t('templates.form.name')" variant="outlined" hide-details class="field" />
        <v-select v-model="form.type" :items="typeItems" :label="t('templates.form.type')" variant="outlined" hide-details class="field" />
        <v-text-field
          v-model="amountYuan"
          :label="t('templates.form.amount')"
          type="number"
          inputmode="decimal"
          variant="outlined"
          hide-details
          class="field"
        />
        <v-select
          v-model="form.accountId"
          :items="accounts.list"
          item-title="name"
          item-value="id"
          :label="t('templates.form.account')"
          clearable
          variant="outlined"
          hide-details
          class="field"
        />
        <v-text-field v-model="form.remark" :label="t('templates.form.remark')" variant="outlined" hide-details class="field field--last" />
        <p class="form-section__tip">{{ t('templates.form.amountTip') }}</p>
      </div>
    </section>

    <section v-if="form.type !== 'transfer'" class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('templates.form.categoryAndTags') }}</h3>
      </header>
      <div class="form-section__body">
        <div class="form-field-row field">
          <v-select
            v-model="form.categoryId"
            :items="form.type === 'expense' ? categories.expenseOptions : categories.incomeOptions"
            item-title="name"
            item-value="id"
            :label="t('templates.form.category')"
            clearable
            variant="outlined"
            hide-details
            class="flex1"
          />
          <v-btn icon="mdi-plus" variant="tonal" color="primary" :aria-label="t('templates.form.newCategory')" @click="openQuickCategory" />
        </div>
        <div class="form-field-row field field--last">
          <v-select
            v-model="form.tagIds"
            :items="tags.list"
            item-title="name"
            item-value="id"
            :label="t('templates.form.tags')"
            multiple
            chips
            closable-chips
            variant="outlined"
            hide-details
            class="flex1"
          />
          <v-btn icon="mdi-plus" variant="tonal" color="primary" :aria-label="t('templates.form.newTag')" @click="openQuickTag" />
        </div>
      </div>
    </section>

    <section v-else class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('templates.form.tags') }}</h3>
      </header>
      <div class="form-section__body">
        <div class="form-field-row field field--last">
          <v-select
            v-model="form.tagIds"
            :items="tags.list"
            item-title="name"
            item-value="id"
            :label="t('templates.form.tags')"
            multiple
            chips
            closable-chips
            variant="outlined"
            hide-details
            class="flex1"
          />
          <v-btn icon="mdi-plus" variant="tonal" color="primary" :aria-label="t('templates.form.newTag')" @click="openQuickTag" />
        </div>
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('templates.form.fareRule') }}</h3>
      </header>
      <div class="form-section__body">
        <v-select
          v-model="form.fareRuleId"
          :items="fareRuleItems"
          item-title="title"
          item-value="value"
          :label="t('templates.form.bindRule')"
          clearable
          variant="outlined"
          hide-details
          class="field field--last"
        />
        <p class="form-section__tip">
          {{ t('templates.form.ruleTip') }}
          <router-link class="inline-link" to="/fare-rules">{{ t('templates.form.manageRules') }}</router-link>
        </p>
      </div>
    </section>

    <section class="form-section">
      <header class="form-section__head">
        <h3 class="form-section__title">{{ t('templates.form.place') }}</h3>
      </header>
      <div class="form-section__body">
        <button v-if="!formGeoLabel" type="button" class="loc-tile" @click="showMap = true">
          <v-icon size="24" color="primary">mdi-map-marker-plus-outline</v-icon>
          <div class="loc-copy">
            <strong>{{ t('templates.form.addPlace') }}</strong>
            <span>{{ t('templates.form.addPlaceHint') }}</span>
          </div>
          <v-icon size="18" class="loc-chevron">mdi-chevron-right</v-icon>
        </button>
        <div v-else class="loc-filled">
          <div class="loc-preview" @click="showMap = true">
            <v-chip size="small" color="primary" variant="tonal" class="mb-2">
              {{ form.geoMode === 'route' ? t('templates.form.route') : t('templates.form.point') }}
            </v-chip>
            <div class="loc-value">{{ formGeoLabel }}</div>
          </div>
          <div class="loc-actions">
            <v-btn variant="tonal" color="primary" size="small" @click="showMap = true">{{ t('templates.form.edit') }}</v-btn>
            <v-btn variant="text" color="error" size="small" @click="clearGeo">{{ t('templates.form.clear') }}</v-btn>
          </div>
        </div>
      </div>
    </section>

    <div v-if="templateId" class="danger-zone">
      <div class="danger-title">{{ t('templates.form.dangerTitle') }}</div>
      <p class="danger-tip">{{ t('templates.form.dangerTip') }}</p>
      <v-btn color="error" variant="outlined" block @click="emit('delete')">{{ t('templates.form.delete') }}</v-btn>
    </div>

    <v-dialog v-model="showQuickCat" max-width="360">
      <v-card>
        <v-card-title>{{ t('templates.form.newCategory') }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="quickName" :label="t('templates.form.quickName')" variant="outlined" hide-details autofocus @keyup.enter="saveQuickCategory" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="showQuickCat = false">{{ t('common.cancel') }}</v-btn>
          <v-btn color="primary" :loading="quickSaving" @click="saveQuickCategory">{{ t('templates.form.createAndUse') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="showQuickTag" max-width="360">
      <v-card>
        <v-card-title>{{ t('templates.form.newTag') }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="quickName" :label="t('templates.form.quickName')" variant="outlined" hide-details autofocus @keyup.enter="saveQuickTag" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="showQuickTag = false">{{ t('common.cancel') }}</v-btn>
          <v-btn color="primary" :loading="quickSaving" @click="saveQuickTag">{{ t('templates.form.createAndUse') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog
      v-model="showMap"
      :fullscreen="isMobile"
      :max-width="isMobile ? undefined : 960"
      transition="dialog-bottom-transition"
    >
      <div class="map-dialog-shell" :class="{ 'is-full': isMobile }">
        <AmapPicker
          v-if="showMap"
          :initial-mode="form.geoMode === 'route' ? 'route' : 'point'"
          :initial-lng="form.geoLng"
          :initial-lat="form.geoLat"
          :initial-name="form.geoName"
          :initial-end-lng="form.geoEndLng"
          :initial-end-lat="form.geoEndLat"
          :initial-end-name="form.geoEndName"
          @pick="onPick"
          @close="showMap = false"
        />
      </div>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import http from '../api/http'
import IconPicker from './IconPicker.vue'
import AmapPicker, { type GeoPickResult } from './AmapPicker.vue'
import { useAccountStore } from '../stores/account'
import { useCategoryStore } from '../stores/category'
import { useTagStore } from '../stores/tag'
import { fenToYuan, yuanToFen } from '../utils/money'
import { txTypeIcon } from '../utils/icons'
import { useBreakpoint } from '../composables/useBreakpoint'

export interface Template {
  id: number
  name: string
  type: string
  amount: number
  accountId?: number | null
  toAccountId?: number | null
  categoryId?: number | null
  tagIds?: number[]
  remark?: string
  fareRuleId?: number | null
  fareRule?: { name: string; city?: string } | null
  icon?: string
  geoMode?: string
  geoLng?: number | null
  geoLat?: number | null
  geoName?: string
  geoEndLng?: number | null
  geoEndLat?: number | null
  geoEndName?: string
}

const props = defineProps<{
  templateId?: number | null
  initial?: Template | null
}>()

const emit = defineEmits<{
  saved: []
  delete: []
}>()

const { t } = useI18n()
const accounts = useAccountStore()
const categories = useCategoryStore()
const tags = useTagStore()
const { isMobile } = useBreakpoint()

const saving = ref(false)
const amountYuan = ref('')
const fareRules = ref<any[]>([])
const showMap = ref(false)
const showQuickCat = ref(false)
const showQuickTag = ref(false)
const quickName = ref('')
const quickSaving = ref(false)

const typeItems = computed(() => [
  { title: t('txType.expense'), value: 'expense' },
  { title: t('txType.income'), value: 'income' },
  { title: t('txType.transfer'), value: 'transfer' },
])

const form = reactive({
  name: '',
  type: 'expense',
  accountId: null as number | null,
  categoryId: null as number | null,
  tagIds: [] as number[],
  remark: '',
  fareRuleId: null as number | null,
  icon: 'mdi-flash',
  geoMode: '' as '' | 'point' | 'route',
  geoLng: null as number | null,
  geoLat: null as number | null,
  geoName: '',
  geoEndLng: null as number | null,
  geoEndLat: null as number | null,
  geoEndName: '',
})

const formGeoLabel = computed(() => {
  if (form.geoMode === 'route' && form.geoName && form.geoEndName) return `${form.geoName} → ${form.geoEndName}`
  return form.geoName || (form.geoLng != null ? t('templates.pickedLocation') : '')
})

const fareRuleItems = computed(() =>
  (fareRules.value || [])
    .filter((r) => r.enabled !== false)
    .map((r) => ({ title: r.city ? t('templates.form.ruleWithCity', { name: r.name, city: r.city }) : r.name, value: r.id })),
)

function resetGeo() {
  form.geoMode = ''
  form.geoLng = null
  form.geoLat = null
  form.geoName = ''
  form.geoEndLng = null
  form.geoEndLat = null
  form.geoEndName = ''
}

function clearGeo() {
  resetGeo()
}

function fillFrom(tpl: Template | null | undefined) {
  if (!tpl) {
    form.name = ''
    form.type = 'expense'
    form.accountId = null
    form.categoryId = null
    form.tagIds = []
    form.remark = ''
    form.fareRuleId = null
    form.icon = 'mdi-flash'
    resetGeo()
    amountYuan.value = ''
    return
  }
  form.name = tpl.name
  form.type = tpl.type
  form.accountId = tpl.accountId ?? null
  form.categoryId = tpl.categoryId ?? null
  form.tagIds = tpl.tagIds || []
  form.remark = tpl.remark || ''
  form.fareRuleId = tpl.fareRuleId || null
  form.icon = tpl.icon || txTypeIcon(tpl.type)
  form.geoMode = (tpl.geoMode as 'point' | 'route') || (tpl.geoLng != null ? 'point' : '')
  form.geoLng = tpl.geoLng ?? null
  form.geoLat = tpl.geoLat ?? null
  form.geoName = tpl.geoName || ''
  form.geoEndLng = tpl.geoEndLng ?? null
  form.geoEndLat = tpl.geoEndLat ?? null
  form.geoEndName = tpl.geoEndName || ''
  amountYuan.value = tpl.amount ? fenToYuan(tpl.amount) : ''
}

async function loadTemplate() {
  if (props.initial) {
    fillFrom(props.initial)
    return
  }
  if (!props.templateId) {
    fillFrom(null)
    return
  }
  const { data } = await http.get('/templates')
  const tpl = (data || []).find((x: Template) => x.id === props.templateId)
  fillFrom(tpl || null)
}

watch(
  () => [props.templateId, props.initial] as const,
  () => {
    loadTemplate()
  },
)

onMounted(async () => {
  await Promise.all([
    accounts.ensure(),
    categories.ensure(),
    tags.ensure(),
    http.get('/fare-rules').then(({ data }) => {
      fareRules.value = data || []
    }),
    loadTemplate(),
  ])
})

function onPick(p: GeoPickResult) {
  form.geoMode = p.mode
  form.geoLng = p.lng
  form.geoLat = p.lat
  form.geoName = p.name
  if (p.mode === 'route') {
    form.geoEndLng = p.endLng ?? null
    form.geoEndLat = p.endLat ?? null
    form.geoEndName = p.endName || ''
  } else {
    form.geoEndLng = null
    form.geoEndLat = null
    form.geoEndName = ''
  }
  showMap.value = false
}

function openQuickCategory() {
  quickName.value = ''
  showQuickCat.value = true
}

function openQuickTag() {
  quickName.value = ''
  showQuickTag.value = true
}

async function saveQuickCategory() {
  const name = quickName.value.trim()
  if (!name) return
  quickSaving.value = true
  try {
    const cat = await categories.create({
      name,
      kind: form.type === 'income' ? 'income' : 'expense',
      icon: 'mdi-shape',
      sort: 0,
    })
    form.categoryId = cat.id
    showQuickCat.value = false
  } finally {
    quickSaving.value = false
  }
}

async function saveQuickTag() {
  const name = quickName.value.trim()
  if (!name) return
  quickSaving.value = true
  try {
    const tag = await tags.create({ name, color: '#1b7f5a', sort: 0 })
    if (!form.tagIds.includes(tag.id)) form.tagIds = [...form.tagIds, tag.id]
    showQuickTag.value = false
  } finally {
    quickSaving.value = false
  }
}

async function save() {
  if (saving.value) return
  saving.value = true
  try {
    const amount = amountYuan.value ? yuanToFen(amountYuan.value) : 0
    const payload = {
      ...form,
      amount,
      icon: form.icon || 'mdi-flash',
      fareRuleId: form.fareRuleId || null,
      geoMode: form.geoLng != null ? (form.geoMode || 'point') : '',
    }
    if (props.templateId) await http.put(`/templates/${props.templateId}`, payload)
    else await http.post('/templates', payload)
    emit('saved')
  } finally {
    saving.value = false
  }
}

defineExpose({ save, saving })
</script>

<style scoped>
.inline-link {
  color: var(--primary);
  font-weight: 600;
  text-decoration: none;
  margin-left: 4px;
  white-space: nowrap;
}
.flex1 { flex: 1; min-width: 0; }

.loc-tile {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 12px;
  border-radius: 14px;
  border: 1.5px dashed var(--primary);
  background: var(--primary-soft);
  color: inherit;
  text-align: left;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  min-width: 0;
}
.loc-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.loc-copy strong { font-size: 0.95rem; color: rgb(var(--v-theme-on-surface)); }
.loc-copy span {
  font-size: 0.78rem;
  color: var(--muted);
  line-height: 1.35;
  overflow-wrap: anywhere;
}
.loc-chevron { color: var(--muted); flex-shrink: 0; }

.loc-filled {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
  border-radius: 14px;
  border: 1px solid var(--surface-border);
  background: var(--primary-soft);
  min-width: 0;
}
.loc-preview { cursor: pointer; min-width: 0; }
.loc-value {
  font-size: 0.92rem;
  font-weight: 600;
  line-height: 1.4;
  overflow-wrap: anywhere;
  word-break: break-word;
  color: rgb(var(--v-theme-on-surface));
}
.loc-actions { display: flex; flex-wrap: wrap; gap: 8px; }

.map-dialog-shell {
  height: min(80vh, 720px);
  background: var(--bg);
  overflow: hidden;
  border-radius: 18px;
}
.map-dialog-shell.is-full {
  height: 100%;
  min-height: var(--app-height, 100dvh);
  border-radius: 0;
}
.map-dialog-shell .picker { height: 100%; }
</style>
