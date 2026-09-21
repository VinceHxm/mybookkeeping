<template>
  <div class="form-root">
    <section class="block">
      <h3 class="block-title">基本信息</h3>
      <IconPicker v-model="form.icon" title="模板图标" fallback="mdi-flash" />
      <v-text-field v-model="form.name" label="名称" variant="outlined" hide-details class="mb-3" />
      <v-select v-model="form.type" :items="typeItems" label="类型" variant="outlined" hide-details class="mb-3" />
      <v-text-field
        v-model="amountYuan"
        label="金额（元）"
        type="number"
        inputmode="decimal"
        variant="outlined"
        hide-details
        class="mb-1"
      />
      <p class="field-note">可空。绑定计费规则时，此金额会作为原价参与折扣计算。</p>
      <v-select
        v-model="form.accountId"
        :items="accounts.list"
        item-title="name"
        item-value="id"
        label="账户"
        clearable
        variant="outlined"
        hide-details
        class="mb-3"
      />
      <v-text-field v-model="form.remark" label="备注" variant="outlined" hide-details />
    </section>

    <section v-if="form.type !== 'transfer'" class="block">
      <h3 class="block-title">分类与标签</h3>
      <div class="field-row mb-3">
        <v-select
          v-model="form.categoryId"
          :items="form.type === 'expense' ? categories.expenseOptions : categories.incomeOptions"
          item-title="name"
          item-value="id"
          label="分类"
          clearable
          variant="outlined"
          hide-details
          class="flex1"
        />
        <v-btn icon="mdi-plus" variant="tonal" color="primary" aria-label="新建分类" @click="openQuickCategory" />
      </div>
      <div class="field-row">
        <v-select
          v-model="form.tagIds"
          :items="tags.list"
          item-title="name"
          item-value="id"
          label="标签"
          multiple
          chips
          closable-chips
          variant="outlined"
          hide-details
          class="flex1"
        />
        <v-btn icon="mdi-plus" variant="tonal" color="primary" aria-label="新建标签" @click="openQuickTag" />
      </div>
    </section>

    <section v-else class="block">
      <h3 class="block-title">标签</h3>
      <div class="field-row">
        <v-select
          v-model="form.tagIds"
          :items="tags.list"
          item-title="name"
          item-value="id"
          label="标签"
          multiple
          chips
          closable-chips
          variant="outlined"
          hide-details
          class="flex1"
        />
        <v-btn icon="mdi-plus" variant="tonal" color="primary" aria-label="新建标签" @click="openQuickTag" />
      </div>
    </section>

    <section class="block">
      <h3 class="block-title">计费规则</h3>
      <v-select
        v-model="form.fareRuleId"
        :items="fareRuleItems"
        item-title="title"
        item-value="value"
        label="绑定规则"
        clearable
        variant="outlined"
        hide-details
        class="mb-1"
      />
      <p class="field-note">
        套用模板时启用规则，按模板金额算建议价。
        <router-link class="inline-link" to="/fare-rules">管理计费规则</router-link>
      </p>
    </section>

    <section class="block">
      <h3 class="block-title">预设地点</h3>
      <button v-if="!formGeoLabel" type="button" class="loc-tile" @click="showMap = true">
        <v-icon size="24" color="primary">mdi-map-marker-plus-outline</v-icon>
        <div class="loc-copy">
          <strong>添加地点</strong>
          <span>常用通勤 / 门店，套用时自动带入</span>
        </div>
        <v-icon size="18" class="loc-chevron">mdi-chevron-right</v-icon>
      </button>
      <div v-else class="loc-filled">
        <div class="loc-preview" @click="showMap = true">
          <v-chip size="small" color="primary" variant="tonal" class="mb-2">
            {{ form.geoMode === 'route' ? '行程' : '位置' }}
          </v-chip>
          <div class="loc-value">{{ formGeoLabel }}</div>
        </div>
        <div class="loc-actions">
          <v-btn variant="tonal" color="primary" size="small" @click="showMap = true">修改</v-btn>
          <v-btn variant="text" color="error" size="small" @click="clearGeo">清除</v-btn>
        </div>
      </div>
    </section>

    <div v-if="templateId" class="danger-zone">
      <div class="danger-title">危险操作</div>
      <p class="danger-tip">删除模板后不可恢复，不会影响已生成的流水。</p>
      <v-btn color="error" variant="outlined" block @click="emit('delete')">删除此模板</v-btn>
    </div>

    <v-dialog v-model="showQuickCat" max-width="360">
      <v-card>
        <v-card-title>新建分类</v-card-title>
        <v-card-text>
          <v-text-field v-model="quickName" label="名称" variant="outlined" hide-details autofocus @keyup.enter="saveQuickCategory" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="showQuickCat = false">取消</v-btn>
          <v-btn color="primary" :loading="quickSaving" @click="saveQuickCategory">创建并选用</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="showQuickTag" max-width="360">
      <v-card>
        <v-card-title>新建标签</v-card-title>
        <v-card-text>
          <v-text-field v-model="quickName" label="名称" variant="outlined" hide-details autofocus @keyup.enter="saveQuickTag" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="showQuickTag = false">取消</v-btn>
          <v-btn color="primary" :loading="quickSaving" @click="saveQuickTag">创建并选用</v-btn>
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

const typeItems = [
  { title: '支出', value: 'expense' },
  { title: '收入', value: 'income' },
  { title: '转账', value: 'transfer' },
]

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
  return form.geoName || (form.geoLng != null ? '已选位置' : '')
})

const fareRuleItems = computed(() =>
  (fareRules.value || [])
    .filter((r) => r.enabled !== false)
    .map((r) => ({ title: r.city ? `${r.name}（${r.city}）` : r.name, value: r.id })),
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

function fillFrom(t: Template | null | undefined) {
  if (!t) {
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
  form.name = t.name
  form.type = t.type
  form.accountId = t.accountId ?? null
  form.categoryId = t.categoryId ?? null
  form.tagIds = t.tagIds || []
  form.remark = t.remark || ''
  form.fareRuleId = t.fareRuleId || null
  form.icon = t.icon || txTypeIcon(t.type)
  form.geoMode = (t.geoMode as 'point' | 'route') || (t.geoLng != null ? 'point' : '')
  form.geoLng = t.geoLng ?? null
  form.geoLat = t.geoLat ?? null
  form.geoName = t.geoName || ''
  form.geoEndLng = t.geoEndLng ?? null
  form.geoEndLat = t.geoEndLat ?? null
  form.geoEndName = t.geoEndName || ''
  amountYuan.value = t.amount ? fenToYuan(t.amount) : ''
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
  const t = (data || []).find((x: Template) => x.id === props.templateId)
  fillFrom(t || null)
}

watch(
  () => [props.templateId, props.initial] as const,
  () => {
    loadTemplate()
  },
)

onMounted(async () => {
  await Promise.all([
    accounts.load(),
    categories.load(),
    tags.load(),
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
.block { margin-bottom: 22px; }
.block-title {
  margin: 0 0 12px;
  font-size: 0.82rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--muted);
}
.field-note {
  margin: 0 0 14px;
  padding: 0 2px;
  font-size: 0.78rem;
  line-height: 1.45;
  color: var(--muted);
}
.inline-link {
  color: var(--primary);
  font-weight: 600;
  text-decoration: none;
  margin-left: 4px;
}
.field-row { display: flex; gap: 8px; align-items: flex-start; }
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
}
.loc-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.loc-copy strong { font-size: 0.95rem; color: rgb(var(--v-theme-on-surface)); }
.loc-copy span { font-size: 0.78rem; color: var(--muted); line-height: 1.35; }
.loc-chevron { color: var(--muted); flex-shrink: 0; }

.loc-filled {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
  border-radius: 14px;
  border: 1px solid var(--surface-border);
  background: var(--primary-soft);
}
.loc-preview { cursor: pointer; min-width: 0; }
.loc-value {
  font-size: 0.92rem;
  font-weight: 600;
  line-height: 1.4;
  word-break: break-word;
  color: rgb(var(--v-theme-on-surface));
}
.loc-actions { display: flex; flex-wrap: wrap; gap: 8px; }

.danger-zone {
  margin-top: 12px;
  margin-bottom: 8px;
  padding: 16px;
  border-radius: 14px;
  border: 1px solid rgba(198, 40, 40, 0.35);
  background: rgba(198, 40, 40, 0.08);
}
.danger-title { font-weight: 700; color: #c62828; margin-bottom: 6px; }
.danger-tip { margin: 0 0 12px; font-size: 0.8rem; color: var(--muted); line-height: 1.4; }

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
