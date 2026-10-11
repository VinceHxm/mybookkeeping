<template>
  <div class="page edit-page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" @click="goBack" />
      <h1>{{ editing ? t('fareEdit.titleEdit') : t('fareEdit.titleNew') }}</h1>
      <v-btn color="primary" variant="text" @click="onSave">{{ t('common.save') }}</v-btn>
    </div>

    <div v-if="pageLoading" class="loading-tip">{{ t('common.loading') }}</div>
    <div v-else class="edit-body hide-scrollbar form-sections">

          <v-alert
            v-if="saveError"
            type="warning"
            variant="tonal"
            density="compact"
            class="mb-0"
            closable
            @click:close="saveError = ''"
          >
            {{ saveError }}
          </v-alert>

          <p class="guide-line">{{ t('fareEdit.guide') }}</p>

          <!-- 场景起步 -->
          <section v-if="!editing" class="form-section">
            <header class="form-section__head">
              <h3 class="form-section__title">{{ t('fareEdit.quickStart') }}</h3>
            </header>
            <div class="form-section__body">
              <p class="form-section__tip form-section__tip--alone">{{ t('fareEdit.quickStartTip') }}</p>
              <div class="scenario-grid">
                <button
                  v-for="p in presets"
                  :key="p.id"
                  type="button"
                  class="scenario-card"
                  @click="applyPreset(p)"
                >
                  <span class="scenario-title">{{ p.title }}</span>
                  <span class="scenario-hint">{{ p.hint }}</span>
                </button>
              </div>
            </div>
          </section>

          <!-- 基础：始终可见 -->
          <section class="form-section form-section--soft">
            <header class="form-section__head">
              <h3 class="form-section__title">{{ t('fareEdit.basic') }}</h3>
            </header>
            <div class="form-section__body">
              <IconPicker v-model="form.icon" :title="t('fareEdit.iconTitle')" fallback="mdi-ticket-percent" />
              <v-text-field v-model="form.name" :label="t('fareEdit.name')" variant="outlined" hide-details class="field" />
              <v-switch v-model="form.enabled" :label="t('fareEdit.enabled')" color="primary" hide-details class="field" />

              <v-text-field
                v-model="form.baseYuan"
                :label="t('fareEdit.baseYuan')"
                type="number"
                inputmode="decimal"
                variant="outlined"
                hide-details
                class="field"
              />
              <v-text-field
                v-model="form.cardZhe"
                :label="t('fareEdit.cardZhe')"
                type="number"
                inputmode="decimal"
                variant="outlined"
                hide-details
                class="field"
              />
              <v-select
                v-model="form.cycleType"
                :items="cycleItems"
                :label="t('fareEdit.cycle')"
                variant="outlined"
                hide-details
                :class="form.cycleType === 'from_day' ? 'field' : 'field field--last'"
              />
              <v-text-field
                v-if="form.cycleType === 'from_day'"
                v-model.number="form.cycleStartDay"
                :label="t('fareEdit.cycleStartDay')"
                type="number"
                min="1"
                max="28"
                variant="outlined"
                hide-details
                class="field field--last"
              />
              <p class="form-section__tip">
                {{ t('fareEdit.basicTip') }}
              </p>
            </div>
          </section>

          <!-- 模块开关 -->
          <section class="form-section">
            <header class="form-section__head">
              <h3 class="form-section__title">{{ t('fareEdit.modules') }}</h3>
            </header>
            <div class="form-section__body">
            <p class="form-section__tip form-section__tip--alone">{{ t('fareEdit.modulesTip') }}</p>

            <div
              v-for="m in moduleDefs"
              :key="m.key"
              class="mod-row"
              :class="{ on: mod[m.key] }"
            >
              <div class="mod-row-main">
                <div>
                  <div class="mod-name">{{ m.title }}</div>
                  <div class="mod-hint">{{ m.hint }}</div>
                </div>
                <v-switch
                  v-model="mod[m.key]"
                  color="primary"
                  hide-details
                  density="compact"
                  @update:model-value="(v: boolean | null) => onModuleToggle(m.key, !!v)"
                />
              </div>

              <!-- 免费日 -->
              <div v-if="m.key === 'freeDays' && mod.freeDays" class="mod-body">
                <div class="block-head inner">
                  <span class="inner-label">{{ t('fareEdit.free.config') }}</span>
                  <v-btn size="small" variant="tonal" color="primary" prepend-icon="mdi-plus" @click="addFree">{{ t('fareEdit.add') }}</v-btn>
                </div>
                <p v-if="!form.freePeriods.length" class="form-section__tip form-section__tip--alone">{{ t('fareEdit.free.emptyTip') }}</p>
                <div v-for="(p, i) in form.freePeriods" :key="'f'+i" class="item-card">
                  <div class="item-card-head">
                    <span>{{ t('fareEdit.free.item', { n: i + 1 }) }}</span>
                    <v-btn icon="mdi-delete-outline" size="small" variant="text" color="error" @click="form.freePeriods.splice(i, 1)" />
                  </div>
                  <v-select v-model="p.recur" :items="recurItems" :label="t('fareEdit.free.recur')" variant="outlined" hide-details class="mb-3" />
                  <template v-if="p.recur !== 'weekly'">
                    <div class="pair">
                      <NativeDateField v-model="p.startDate" :label="t('fareEdit.free.startDate')" variant="outlined" hide-details />
                      <NativeDateField v-model="p.endDate" :label="t('fareEdit.free.endDate')" variant="outlined" hide-details />
                    </div>
                  </template>
                  <template v-else>
                    <v-select
                      v-model="p.weekdays"
                      :items="weekdayItems"
                      :label="t('fareEdit.free.weekdays')"
                      multiple
                      chips
                      closable-chips
                      variant="outlined"
                      hide-details
                      class="mb-3"
                    />
                    <div class="pair">
                      <NativeDateField v-model="p.startDate" :label="t('fareEdit.free.validFrom')" variant="outlined" hide-details />
                      <NativeDateField v-model="p.endDate" :label="t('fareEdit.free.validTo')" variant="outlined" hide-details />
                    </div>
                  </template>
                </div>
              </div>

              <!-- 节假日 -->
              <div v-if="m.key === 'holiday' && mod.holiday" class="mod-body">
                <v-switch
                  v-model="form.freeOnHoliday"
                  :label="t('fareEdit.holiday.freeOnHoliday')"
                  color="primary"
                  hide-details
                  class="mb-2"
                />
                <div class="holiday-actions">
                  <v-btn
                    size="small"
                    variant="tonal"
                    color="primary"
                    :loading="holidayLoading"
                    prepend-icon="mdi-eye-outline"
                    @click="loadHolidays"
                  >
                    {{ holidayYearBlocks.length ? t('fareEdit.holiday.refreshList') : t('fareEdit.holiday.view') }}
                  </v-btn>
                  <v-btn
                    v-if="auth.isAdmin()"
                    size="small"
                    variant="text"
                    color="primary"
                    :loading="holidayRefreshing"
                    prepend-icon="mdi-cloud-download-outline"
                    @click="refreshHolidays"
                  >
                    {{ t('fareEdit.holiday.adminRefresh') }}
                  </v-btn>
                </div>
                <p class="form-section__tip">
                  {{ t('fareEdit.holiday.tip') }}
                  <template v-if="holidayMeta"> {{ holidayMeta }}</template>
                </p>
                <div v-if="holidayYearBlocks.length" class="holiday-legend">
                  <span class="leg past"><i />{{ t('fareEdit.holiday.past') }}</span>
                  <span class="leg now"><i />{{ t('fareEdit.holiday.now') }}</span>
                  <span class="leg soon"><i />{{ t('fareEdit.holiday.soon') }}</span>
                </div>
                <div v-if="holidayYearBlocks.length" class="holiday-years">
                  <div v-for="yb in holidayYearBlocks" :key="yb.year" class="holiday-year-block">
                    <div class="hy-head">
                      <span class="hy-year">{{ t('fareEdit.holiday.year', { year: yb.year }) }}</span>
                      <span class="hy-tag">{{ yb.tag }}</span>
                    </div>
                    <div v-if="!yb.groups.length" class="hy-empty">{{ t('fareEdit.holiday.empty') }}</div>
                    <div v-else class="holiday-groups">
                      <div
                        v-for="g in yb.groups"
                        :key="yb.year + g.name + g.range"
                        class="holiday-group"
                        :class="g.status"
                      >
                        <div class="hg-name">{{ g.name }}</div>
                        <div class="hg-range">{{ g.range }}</div>
                        <div class="hg-status">{{ g.statusLabel }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 日内时段 -->
              <div v-if="m.key === 'timeWin' && mod.timeWin" class="mod-body">
                <div class="block-head inner">
                  <span class="inner-label">{{ t('fareEdit.time.config') }}</span>
                  <v-btn size="small" variant="tonal" color="primary" prepend-icon="mdi-plus" @click="addTimeWin">{{ t('fareEdit.add') }}</v-btn>
                </div>
                <p class="form-section__tip form-section__tip--alone">{{ t('fareEdit.time.tip') }}</p>
                <div v-for="(w, i) in form.timeWindows" :key="'w'+i" class="item-card">
                  <div class="item-card-head">
                    <span>{{ t('fareEdit.time.item', { n: i + 1 }) }}</span>
                    <v-btn icon="mdi-delete-outline" size="small" variant="text" color="error" @click="form.timeWindows.splice(i, 1)" />
                  </div>
                  <v-text-field v-model="w.name" :label="t('fareEdit.time.name')" variant="outlined" hide-details class="mb-3" />
                  <div class="pair mb-3">
                    <v-text-field v-model="w.startHm" :label="t('fareEdit.time.start')" placeholder="06:00" variant="outlined" hide-details />
                    <v-text-field v-model="w.endHm" :label="t('fareEdit.time.end')" placeholder="07:00" variant="outlined" hide-details />
                  </div>
                  <v-select
                    v-model="w.weekdays"
                    :items="weekdayItems"
                    :label="t('fareEdit.time.weekdays')"
                    multiple
                    chips
                    closable-chips
                    variant="outlined"
                    hide-details
                    class="mb-3"
                  />
                  <v-switch v-model="w.free" :label="t('fareEdit.time.free')" color="primary" hide-details class="mb-3" />
                  <div v-if="!w.free" class="pair">
                    <v-select v-model="w.base" :items="tierBaseItems" :label="t('fareEdit.base')" variant="outlined" hide-details />
                    <v-text-field v-model="w.zhe" :label="t('fareEdit.zhe')" type="number" inputmode="decimal" variant="outlined" hide-details />
                  </div>
                </div>
              </div>

              <!-- 金额阶梯 -->
              <div v-if="m.key === 'amountTier' && mod.amountTier" class="mod-body">
                <div class="block-head inner">
                  <span class="inner-label">{{ t('fareEdit.tier.config') }}</span>
                  <v-btn size="small" variant="tonal" color="primary" prepend-icon="mdi-plus" @click="addTier">{{ t('fareEdit.tier.add') }}</v-btn>
                </div>
                <p class="form-section__tip form-section__tip--alone">{{ t('fareEdit.tier.tip') }}</p>
                <div v-for="(tier, i) in form.tiers" :key="'t'+i" class="item-card">
                  <div class="item-card-head">
                    <span>{{ t('fareEdit.tier.item', { n: i + 1 }) }}</span>
                    <v-btn icon="mdi-delete-outline" size="small" variant="text" color="error" @click="form.tiers.splice(i, 1)" />
                  </div>
                  <div class="pair mb-3">
                    <v-text-field v-model="tier.minYuan" :label="t('fareEdit.tier.min')" type="number" inputmode="decimal" variant="outlined" hide-details />
                    <v-text-field v-model="tier.maxYuan" :label="t('fareEdit.tier.max')" type="number" inputmode="decimal" variant="outlined" hide-details />
                  </div>
                  <div class="pair">
                    <v-select v-model="tier.base" :items="tierBaseItems" :label="t('fareEdit.base')" variant="outlined" hide-details />
                    <v-text-field v-model="tier.zhe" :label="t('fareEdit.zhe')" type="number" inputmode="decimal" variant="outlined" hide-details />
                  </div>
                </div>
              </div>

              <!-- 乘次阶梯 -->
              <div v-if="m.key === 'countTier' && mod.countTier" class="mod-body">
                <div class="block-head inner">
                  <span class="inner-label">{{ t('fareEdit.count.config') }}</span>
                  <v-btn size="small" variant="tonal" color="primary" prepend-icon="mdi-plus" @click="addCountTier">{{ t('fareEdit.add') }}</v-btn>
                </div>
                <p class="form-section__tip form-section__tip--alone">{{ t('fareEdit.count.tip') }}</p>
                <div v-for="(ct, i) in form.countTiers" :key="'c'+i" class="item-card">
                  <div class="item-card-head">
                    <span>{{ t('fareEdit.count.item', { n: i + 1 }) }}</span>
                    <v-btn icon="mdi-delete-outline" size="small" variant="text" color="error" @click="form.countTiers.splice(i, 1)" />
                  </div>
                  <div class="pair mb-3">
                    <v-text-field v-model.number="ct.minCount" :label="t('fareEdit.count.min')" type="number" variant="outlined" hide-details />
                    <v-text-field v-model.number="ct.maxCount" :label="t('fareEdit.count.max')" type="number" variant="outlined" hide-details />
                  </div>
                  <v-switch v-model="ct.free" :label="t('fareEdit.count.free')" color="primary" hide-details class="mb-3" />
                  <template v-if="!ct.free">
                    <div class="pair mb-3">
                      <v-select v-model="ct.base" :items="tierBaseItems" :label="t('fareEdit.base')" variant="outlined" hide-details />
                      <v-text-field v-model="ct.zhe" :label="t('fareEdit.zhe')" type="number" inputmode="decimal" variant="outlined" hide-details />
                    </div>
                    <v-text-field v-model="ct.offYuan" :label="t('fareEdit.count.off')" type="number" inputmode="decimal" variant="outlined" hide-details />
                  </template>
                </div>
              </div>
            </div>
            </div>
          </section>

          <!-- 更多选项 -->
          <section class="form-section">
            <div class="form-section__body">
              <button type="button" class="more-toggle" @click="mod.more = !mod.more">
                <span>{{ mod.more ? t('fareEdit.more.collapse') : t('fareEdit.more.expand') }}</span>
                <span class="more-sub">{{ t('fareEdit.more.sub') }}</span>
                <v-icon size="18">{{ mod.more ? 'mdi-chevron-up' : 'mdi-chevron-down' }}</v-icon>
              </button>
              <div v-if="mod.more" class="more-body">
                <v-text-field
                  v-model="form.city"
                  :label="t('fareEdit.more.city')"
                  :placeholder="t('fareEdit.more.cityPlaceholder')"
                  variant="outlined"
                  hide-details
                  class="field"
                />
                <v-text-field
                  v-model="form.amountOffYuan"
                  :label="t('fareEdit.more.amountOff')"
                  type="number"
                  inputmode="decimal"
                  variant="outlined"
                  hide-details
                  class="field"
                />
                <v-select
                  v-model="form.stackMode"
                  :items="stackItems"
                  :label="t('fareEdit.more.stack')"
                  variant="outlined"
                  hide-details
                  class="field"
                />
                <div class="pair field">
                  <NativeDateField v-model="form.validFrom" :label="t('fareEdit.more.validFrom')" variant="outlined" hide-details />
                  <NativeDateField v-model="form.validTo" :label="t('fareEdit.more.validTo')" variant="outlined" hide-details />
                </div>
                <v-textarea v-model="form.note" :label="t('fareEdit.more.note')" rows="2" auto-grow variant="outlined" hide-details class="field field--last" />
                <p class="form-section__tip">
                  {{ t('fareEdit.more.tip') }}
                </p>
              </div>
            </div>
          </section>

          <!-- 预览 / 补录 -->
          <section v-if="editing" class="form-section">
            <header class="form-section__head">
              <h3 class="form-section__title">{{ t('fareEdit.trial.title') }}</h3>
            </header>
            <div class="form-section__body">
              <div class="form-section__group">
                <v-text-field
                  v-model="cycleSeedYuan"
                  :label="t('fareEdit.trial.seedAmount')"
                  type="number"
                  inputmode="decimal"
                  variant="outlined"
                  hide-details
                  class="field"
                />
                <v-text-field
                  v-model.number="cycleCountSeed"
                  :label="t('fareEdit.trial.seedCount')"
                  type="number"
                  variant="outlined"
                  hide-details
                  class="field field--last"
                />
                <p class="form-section__tip">{{ t('fareEdit.trial.seedTip') }}</p>
              </div>
              <div class="form-section__group">
                <v-text-field
                  v-model="trialBaseYuan"
                  :label="t('fareEdit.trial.base')"
                  type="number"
                  inputmode="decimal"
                  variant="outlined"
                  hide-details
                  class="field field--last"
                />
                <v-btn
                  block
                  variant="tonal"
                  color="primary"
                  class="mt-3"
                  :loading="previewLoading"
                  prepend-icon="mdi-calculator-variant-outline"
                  @click="refreshPreview"
                >
                  {{ t('fareEdit.trial.preview') }}
                </v-btn>
                <div v-if="preview" class="preview-body">
                  <div class="preview-amount">{{ t('fareEdit.trial.amount', { amount: `¥${fenToYuan(preview.amountFen)}` }) }}</div>
                  <div class="preview-reason">{{ preview.reason }}</div>
                  <div class="preview-meta">
                    {{ t('fareEdit.trial.meta', {
                      tx: `¥${fenToYuan(preview.txSpentFen || 0)}`,
                      seed: `¥${fenToYuan(preview.seedFen || 0)}`,
                      total: `¥${fenToYuan(preview.monthSpentFen)}`,
                      n: preview.thisRideNo || 1,
                      start: preview.cycleStart,
                      end: preview.cycleEnd,
                    }) }}
                  </div>
                </div>
              </div>
            </div>
          </section>

          <div v-if="editing" class="danger-zone">
            <div class="danger-title">{{ t('fareEdit.danger.title') }}</div>
            <p class="danger-tip">{{ t('fareEdit.danger.tip') }}</p>
            <v-btn color="error" variant="outlined" block @click="editing && askDelete(editing)">{{ t('fareEdit.danger.delete') }}</v-btn>
          </div>
        
    </div>

    <DeleteConfirmDialog
      v-model="confirmOpen"
      :name="pendingDelete?.name || ''"
      :loading="deleting"
      @confirm="doDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import http from '../api/http'
import IconPicker from '../components/IconPicker.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import NativeDateField from '../components/NativeDateField.vue'
import { fenToYuan, yuanToFen } from '../utils/money'
import { rateToZhe, zheToRate, type FareRule } from '../utils/fare'
import { getFarePresets, type FarePreset } from '../utils/farePresets'
import { useAuthStore } from '../stores/auth'

type ModKey = 'freeDays' | 'holiday' | 'timeWin' | 'amountTier' | 'countTier'

const { t, locale } = useI18n()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const list = ref<FareRule[]>([])
const editing = ref<FareRule | null>(null)
const pageLoading = ref(true)
const preview = ref<any>(null)
const previewLoading = ref(false)
const trialBaseYuan = ref('')
const cycleSeedYuan = ref('')
const cycleCountSeed = ref(0)
const saveError = ref('')
const confirmOpen = ref(false)
const pendingDelete = ref<FareRule | null>(null)
const deleting = ref(false)
const presets = computed(() => getFarePresets())
const holidayLoading = ref(false)
const holidayRefreshing = ref(false)
const holidayMeta = ref('')
/** 按年存放：{ year, days }[] */
const holidayByYear = ref<{ year: number; days: { date: string; name: string; holiday: boolean }[]; fromDb?: boolean }[]>([])

function visibleHolidayYears(at = new Date()) {
  const y = at.getFullYear()
  // 12 月交接：展示今年 + 明年；其余时间只展示当年
  if (at.getMonth() === 11) return [y, y + 1]
  return [y]
}

function todayYmd(at = new Date()) {
  const m = String(at.getMonth() + 1).padStart(2, '0')
  const d = String(at.getDate()).padStart(2, '0')
  return `${at.getFullYear()}-${m}-${d}`
}

type HolidayGroupRow = {
  name: string
  range: string
  count: number
  status: 'past' | 'now' | 'soon'
  statusLabel: string
}

const HOLIDAY_STATUS_KEY = { past: 'fareEdit.holiday.past', now: 'fareEdit.holiday.now', soon: 'fareEdit.holiday.soon' } as const

function groupHolidayDays(
  days: { date: string; name: string; holiday: boolean }[],
  today: string,
): HolidayGroupRow[] {
  const rest = days
    .filter((d) => d.holiday)
    .slice()
    .sort((a, b) => a.date.localeCompare(b.date))
  if (!rest.length) return []

  type Seg = { name: string; dates: string[] }
  const segs: Seg[] = []
  const fallback = t('fareEdit.holiday.restFallback')
  for (const d of rest) {
    // 节假日名是接口数据（中文），只去掉「补班」后缀，不翻译
    const name = (d.name || fallback).replace(/(前|后)?补班/, '').trim() || fallback
    const last = segs[segs.length - 1]
    if (last && last.name === name) last.dates.push(d.date)
    else segs.push({ name, dates: [d.date] })
  }

  return segs.map((s) => {
    const start = s.dates[0]
    const end = s.dates[s.dates.length - 1]
    let status: 'past' | 'now' | 'soon' = 'soon'
    if (end < today) status = 'past'
    else if (start <= today && today <= end) status = 'now'
    const r0 = start.slice(5)
    const r1 = end.slice(5)
    return {
      name: s.name,
      range: s.dates.length === 1 ? r0 : `${r0} ~ ${r1}`,
      count: s.dates.length,
      status,
      statusLabel: t(HOLIDAY_STATUS_KEY[status]),
    }
  })
}

const holidayYearBlocks = computed(() => {
  const today = todayYmd()
  const now = new Date()
  const cur = now.getFullYear()
  return holidayByYear.value.map((yb) => {
    let tag = ''
    if (yb.year === cur) tag = t('fareEdit.holiday.thisYear')
    else if (yb.year === cur + 1) tag = t('fareEdit.holiday.nextYear')
    else if (yb.year === cur - 1) tag = t('fareEdit.holiday.lastYear')
    return {
      year: yb.year,
      tag,
      groups: groupHolidayDays(yb.days || [], today),
    }
  })
})

const mod = reactive({
  freeDays: false,
  holiday: false,
  timeWin: false,
  amountTier: false,
  countTier: false,
  more: false,
})

const MODULE_KEYS: ModKey[] = ['freeDays', 'holiday', 'timeWin', 'amountTier', 'countTier']
const moduleDefs = computed(() =>
  MODULE_KEYS.map((key) => ({
    key,
    title: t(`fareEdit.mod.${key}.title`),
    hint: t(`fareEdit.mod.${key}.hint`),
  })),
)

const cycleItems = computed(() => [
  { title: t('fareEdit.cycleItems.calendar'), value: 'calendar_month' },
  { title: t('fareEdit.cycleItems.fromDay'), value: 'from_day' },
])
const stackItems = computed(() => [
  { title: t('fareEdit.stackItems.prefer'), value: 'prefer' },
  { title: t('fareEdit.stackItems.lowest'), value: 'lowest' },
  { title: t('fareEdit.stackItems.stack'), value: 'stack' },
])
const recurItems = computed(() => [
  { title: t('fareEdit.recurItems.none'), value: 'none' },
  { title: t('fareEdit.recurItems.yearly'), value: 'yearly' },
  { title: t('fareEdit.recurItems.weekly'), value: 'weekly' },
])
const weekdayItems = computed(() => {
  // dayjs 的 locale 不是响应式的，读一次 locale 让切换语言时重算
  void locale.value
  return [0, 1, 2, 3, 4, 5, 6].map((d) => ({ title: dayjs().day(d).format('dd'), value: d }))
})
const tierBaseItems = computed(() => [
  { title: t('fareEdit.baseItems.full'), value: 'full' },
  { title: t('fareEdit.baseItems.card'), value: 'card' },
])

const form = reactive({
  name: '',
  city: '',
  note: '',
  validFrom: '',
  validTo: '',
  enabled: true,
  icon: 'mdi-ticket-percent',
  baseYuan: '',
  cardZhe: '10',
  amountOffYuan: '',
  cycleType: 'calendar_month' as 'calendar_month' | 'from_day',
  cycleStartDay: 1,
  stackMode: 'prefer' as 'prefer' | 'lowest' | 'stack',
  freeOnHoliday: false,
  freePeriods: [] as { startDate: string; endDate: string; recur: string; weekdays: number[] }[],
  tiers: [] as { minYuan: string; maxYuan: string; base: 'full' | 'card'; zhe: string }[],
  timeWindows: [] as {
    name: string
    startHm: string
    endHm: string
    weekdays: number[]
    free: boolean
    base: 'full' | 'card'
    zhe: string
  }[],
  countTiers: [] as {
    minCount: number
    maxCount: number
    free: boolean
    base: 'full' | 'card'
    zhe: string
    offYuan: string
  }[],
})

function resetModules() {
  mod.freeDays = false
  mod.holiday = false
  mod.timeWin = false
  mod.amountTier = false
  mod.countTier = false
  mod.more = false
}

function syncModulesFromForm() {
  mod.freeDays = form.freePeriods.length > 0
  mod.holiday = !!form.freeOnHoliday
  mod.timeWin = form.timeWindows.length > 0
  mod.amountTier = form.tiers.length > 0
  mod.countTier = form.countTiers.length > 0
  mod.more = !!(
    form.city ||
    form.note ||
    form.validFrom ||
    form.validTo ||
    form.amountOffYuan ||
    (form.stackMode && form.stackMode !== 'prefer')
  )
}

function onModuleToggle(key: ModKey, on: boolean) {
  if (!on) {
    if (key === 'freeDays') form.freePeriods = []
    if (key === 'holiday') {
      form.freeOnHoliday = false
      holidayByYear.value = []
      holidayMeta.value = ''
    }
    if (key === 'timeWin') form.timeWindows = []
    if (key === 'amountTier') form.tiers = []
    if (key === 'countTier') form.countTiers = []
    return
  }
  if (key === 'freeDays' && !form.freePeriods.length) addFree()
  if (key === 'holiday') form.freeOnHoliday = true
  if (key === 'timeWin' && !form.timeWindows.length) addTimeWin()
  if (key === 'amountTier' && !form.tiers.length) addTier()
  if (key === 'countTier' && !form.countTiers.length) addCountTier()
}

async function reload() {
  const { data } = await http.get('/fare-rules')
  list.value = data || []
}

function loadIntoForm(r: FareRule | null) {
  editing.value = r
  saveError.value = ''
  resetForm()
  if (!r) return
  fillFromRule(r)
  form.tiers = (r.tiers || []).map((tier: any) => {
    const minFen = Number(tier.minFen ?? tier.thresholdFen ?? 0)
    const maxFen = Number(tier.maxFen ?? 0)
    return {
      minYuan: fenToYuan(minFen),
      maxYuan: maxFen > 0 ? fenToYuan(maxFen) : '',
      base: (tier.base === 'full' ? 'full' : 'card') as 'full' | 'card',
      zhe: rateToZhe(tier.rate ?? 1),
    }
  })
  syncModulesFromForm()
  preview.value = null
  trialBaseYuan.value = r.baseFen ? fenToYuan(r.baseFen) : ''
  cycleSeedYuan.value = r.cycleSeedFen && r.cycleSeedFen > 0 ? fenToYuan(r.cycleSeedFen) : ''
  cycleCountSeed.value = r.cycleCountSeed && r.cycleCountSeed > 0 ? r.cycleCountSeed : 0
  void refreshPreview().then(() => {
    if (preview.value) {
      cycleSeedYuan.value = preview.value.seedFen > 0 ? fenToYuan(preview.value.seedFen) : ''
      cycleCountSeed.value = preview.value.seedRideCount > 0 ? preview.value.seedRideCount : 0
    }
  })
}

onMounted(async () => {
  pageLoading.value = true
  try {
    if (auth.token && !auth.profile) {
      auth.fetchMe().catch(() => {})
    }
    await reload()
    await loadHolidays()
    const raw = route.params.id
    if (!raw || raw === 'new') {
      loadIntoForm(null)
    } else {
      const id = Number(raw)
      const r = list.value.find((x) => x.id === id) || null
      if (!r) {
        router.replace('/fare-rules')
        return
      }
      loadIntoForm(r)
    }
  } finally {
    pageLoading.value = false
  }
})

function resetForm() {
  form.name = ''
  form.city = ''
  form.note = ''
  form.validFrom = ''
  form.validTo = ''
  form.enabled = true
  form.icon = 'mdi-ticket-percent'
  form.baseYuan = ''
  form.cardZhe = '10'
  form.amountOffYuan = ''
  form.cycleType = 'calendar_month'
  form.cycleStartDay = 1
  form.stackMode = 'prefer'
  form.freeOnHoliday = false
  form.freePeriods = []
  form.tiers = []
  form.timeWindows = []
  form.countTiers = []
  preview.value = null
  trialBaseYuan.value = ''
  cycleSeedYuan.value = ''
  cycleCountSeed.value = 0
  holidayByYear.value = []
  holidayMeta.value = ''
  resetModules()
}

function fillFromRule(r: Partial<FareRule>) {
  if (r.name != null) form.name = r.name
  if (r.city != null) form.city = r.city
  if (r.note != null) form.note = r.note
  if (r.validFrom != null) form.validFrom = r.validFrom
  if (r.validTo != null) form.validTo = r.validTo
  if (r.enabled != null) form.enabled = r.enabled !== false
  if (r.icon) form.icon = r.icon
  if (r.baseFen != null) form.baseYuan = r.baseFen ? fenToYuan(r.baseFen) : ''
  if (r.cardRate != null) form.cardZhe = rateToZhe(r.cardRate ?? 1)
  if (r.amountOffFen != null) form.amountOffYuan = r.amountOffFen ? fenToYuan(r.amountOffFen) : ''
  if (r.cycleType) form.cycleType = r.cycleType as any
  if (r.cycleStartDay) form.cycleStartDay = r.cycleStartDay
  if (r.stackMode) form.stackMode = r.stackMode
  if (r.freeOnHoliday != null) form.freeOnHoliday = !!r.freeOnHoliday
  if (r.freePeriods) {
    form.freePeriods = r.freePeriods.map((p) => ({
      startDate: p.startDate || '',
      endDate: p.endDate || '',
      recur: p.recur || 'none',
      weekdays: [...(p.weekdays || [])],
    }))
  }
  if (r.tiers) {
    form.tiers = r.tiers.map((tier: any) => {
      const minFen = Number(tier.minFen ?? tier.thresholdFen ?? 0)
      const maxFen = Number(tier.maxFen ?? 0)
      return {
        minYuan: fenToYuan(minFen),
        maxYuan: maxFen > 0 ? fenToYuan(maxFen) : '',
        base: (tier.base === 'full' ? 'full' : 'card') as 'full' | 'card',
        zhe: rateToZhe(tier.rate ?? 1),
      }
    })
  }
  if (r.timeWindows) {
    form.timeWindows = r.timeWindows.map((w) => ({
      name: w.name || '',
      startHm: w.startHm || '',
      endHm: w.endHm || '',
      weekdays: [...(w.weekdays || [])],
      free: !!w.free,
      base: (w.base === 'full' ? 'full' : 'card') as 'full' | 'card',
      zhe: rateToZhe(w.rate ?? 1),
    }))
  }
  if (r.countTiers) {
    form.countTiers = r.countTiers.map((ct) => ({
      minCount: ct.minCount ?? 0,
      maxCount: ct.maxCount ?? 0,
      free: !!ct.free,
      base: (ct.base === 'full' ? 'full' : 'card') as 'full' | 'card',
      zhe: rateToZhe(ct.rate ?? 1),
      offYuan: ct.amountOffFen ? fenToYuan(ct.amountOffFen) : '',
    }))
  }
}

function applyPreset(p: FarePreset) {
  fillFromRule(p.rule)
  if (!form.name) form.name = p.title
  syncModulesFromForm()
  if (p.modules) {
    if (p.modules.freeDays) mod.freeDays = true
    if (p.modules.holiday) mod.holiday = true
    if (p.modules.timeWin) mod.timeWin = true
    if (p.modules.amountTier) mod.amountTier = true
    if (p.modules.countTier) mod.countTier = true
  }
}

function addFree() {
  form.freePeriods.push({ startDate: '', endDate: '', recur: 'weekly', weekdays: [6, 0] })
}
function addTier() {
  form.tiers.push({ minYuan: '80', maxYuan: '150', base: 'full', zhe: '7' })
}
function addTimeWin() {
  form.timeWindows.push({
    name: t('farePresets.morning.window'),
    startHm: '00:00',
    endHm: '07:00',
    weekdays: [1, 2, 3, 4, 5],
    free: false,
    base: 'full',
    zhe: '7',
  })
}
function addCountTier() {
  form.countTiers.push({
    minCount: 39,
    maxCount: 40,
    free: true,
    base: 'card',
    zhe: '10',
    offYuan: '',
  })
}

async function loadHolidays() {
  holidayLoading.value = true
  try {
    const years = visibleHolidayYears()
    const rows = await Promise.all(
      years.map(async (year) => {
        const { data } = await http.get(`/holidays/${year}`)
        return {
          year,
          days: data?.days || [],
          fromDb: !!data?.fromDb,
        }
      }),
    )
    holidayByYear.value = rows
    const hasAny = rows.some((r) => r.fromDb && (r.days || []).some((d: any) => d.holiday))
    if (!hasAny) {
      holidayMeta.value = t('fareEdit.holiday.metaNoData')
    } else if (years.length > 1) {
      holidayMeta.value = t('fareEdit.holiday.metaTwoYears', { y1: years[0], y2: years[1] })
    } else {
      holidayMeta.value = t('fareEdit.holiday.metaOneYear', { year: years[0] })
    }
  } catch {
    holidayByYear.value = []
    holidayMeta.value = ''
    saveError.value = t('fareEdit.holiday.loadFailed')
  } finally {
    holidayLoading.value = false
  }
}

async function refreshHolidays() {
  if (!auth.isAdmin()) return
  holidayRefreshing.value = true
  try {
    const years = visibleHolidayYears()
    const rows = await Promise.all(
      years.map(async (year) => {
        const { data } = await http.post(`/holidays/${year}/refresh`)
        return {
          year,
          days: data?.days || [],
          fromDb: true,
        }
      }),
    )
    holidayByYear.value = rows
    const n = holidayYearBlocks.value.reduce((s, y) => s + y.groups.length, 0)
    holidayMeta.value = t('fareEdit.holiday.refreshed', { years: years.join(t('fareEdit.holiday.yearSep')), n })
  } catch (e: any) {
    saveError.value = e?.response?.data?.error || e?.message || t('fareEdit.holiday.refreshFailed')
  } finally {
    holidayRefreshing.value = false
  }
}

function buildPayload() {
  const seedFen =
    cycleSeedYuan.value !== '' && cycleSeedYuan.value != null
      ? yuanToFen(cycleSeedYuan.value)
      : 0
  return {
    name: form.name,
    city: form.city,
    note: form.note,
    validFrom: form.validFrom,
    validTo: form.validTo,
    enabled: form.enabled,
    icon: form.icon,
    baseFen: form.baseYuan ? yuanToFen(form.baseYuan) : 0,
    cardRate: zheToRate(form.cardZhe),
    amountOffFen: form.amountOffYuan ? yuanToFen(form.amountOffYuan) : 0,
    stackMode: form.stackMode,
    freeOnHoliday: mod.holiday ? form.freeOnHoliday : false,
    cycleType: form.cycleType,
    cycleStartDay: form.cycleStartDay,
    cycleSeedFen: seedFen,
    cycleCountSeed: Math.max(0, Number(cycleCountSeed.value) || 0),
    freePeriods: mod.freeDays
      ? form.freePeriods.map((p) => ({
          startDate: p.startDate,
          endDate: p.endDate,
          recur: p.recur,
          weekdays: p.recur === 'weekly' ? p.weekdays : [],
        }))
      : [],
    tiers: mod.amountTier
      ? form.tiers.map((tier) => ({
          minFen: tier.minYuan !== '' && tier.minYuan != null ? yuanToFen(tier.minYuan) : 0,
          maxFen: tier.maxYuan !== '' && tier.maxYuan != null ? yuanToFen(tier.maxYuan) : 0,
          base: tier.base || 'card',
          rate: zheToRate(tier.zhe),
        }))
      : [],
    timeWindows: mod.timeWin
      ? form.timeWindows.map((w) => ({
          name: w.name,
          startHm: w.startHm,
          endHm: w.endHm,
          weekdays: w.weekdays || [],
          free: w.free,
          base: w.base || 'card',
          rate: zheToRate(w.zhe),
        }))
      : [],
    countTiers: mod.countTier
      ? form.countTiers.map((ct) => ({
          minCount: Number(ct.minCount) || 0,
          maxCount: Number(ct.maxCount) || 0,
          free: ct.free,
          base: ct.base || 'card',
          rate: zheToRate(ct.zhe),
          amountOffFen: ct.offYuan ? yuanToFen(ct.offYuan) : 0,
        }))
      : [],
  }
}

function validateForm(): string {
  if (!form.name.trim()) return t('fareEdit.validate.name')
  if (mod.amountTier) {
    for (let i = 0; i < form.tiers.length; i++) {
      const tier = form.tiers[i]
      if (tier.minYuan === '' || tier.minYuan == null) return t('fareEdit.validate.tierMin', { n: i + 1 })
      if (tier.zhe === '' || tier.zhe == null) return t('fareEdit.validate.tierZhe', { n: i + 1 })
    }
  }
  if (mod.timeWin) {
    for (let i = 0; i < form.timeWindows.length; i++) {
      const w = form.timeWindows[i]
      if (!w.startHm || !w.endHm) return t('fareEdit.validate.timeRange', { n: i + 1 })
    }
  }
  return ''
}

async function refreshPreview() {
  if (!editing.value?.id) return
  previewLoading.value = true
  try {
    const body: { baseFen?: number; monthSpentFen?: number; rideCount?: number } = {}
    if (trialBaseYuan.value !== '' && trialBaseYuan.value != null) {
      const fen = yuanToFen(trialBaseYuan.value)
      if (fen > 0) body.baseFen = fen
    }
    const { data } = await http.post(`/fare-rules/${editing.value.id}/preview`, body)
    const tx = Number(data.txSpentFen || 0)
    const savedSeed = Number(data.seedFen || 0)
    const txRides = Number(data.txRideCount || 0)
    const savedRideSeed = Number(data.seedRideCount || 0)
    const localSeed =
      cycleSeedYuan.value !== '' && cycleSeedYuan.value != null
        ? Math.max(0, yuanToFen(cycleSeedYuan.value))
        : 0
    const localRideSeed = Math.max(0, Number(cycleCountSeed.value) || 0)
    if (localSeed !== savedSeed || localRideSeed !== savedRideSeed) {
      const { data: d2 } = await http.post(`/fare-rules/${editing.value.id}/preview`, {
        ...body,
        monthSpentFen: tx + localSeed,
        rideCount: txRides + localRideSeed,
      })
      preview.value = {
        ...d2,
        txSpentFen: tx,
        seedFen: localSeed,
        monthSpentFen: tx + localSeed,
        txRideCount: txRides,
        seedRideCount: localRideSeed,
        rideCount: txRides + localRideSeed,
        thisRideNo: txRides + localRideSeed + 1,
      }
    } else {
      preview.value = data
    }
  } catch {
    preview.value = null
  } finally {
    previewLoading.value = false
  }
}

async function onSave() {
  saveError.value = validateForm()
  if (saveError.value) return
  const payload = buildPayload()
  if (editing.value?.id) await http.put(`/fare-rules/${editing.value.id}`, payload)
  else await http.post('/fare-rules', payload)
  router.replace('/fare-rules')
}

function askDelete(r: FareRule) {
  pendingDelete.value = r
  confirmOpen.value = true
}

async function doDelete() {
  if (!pendingDelete.value?.id) return
  deleting.value = true
  try {
    await http.delete(`/fare-rules/${pendingDelete.value.id}`)
    confirmOpen.value = false
    pendingDelete.value = null
    router.replace('/fare-rules')
  } finally {
    deleting.value = false
  }
}

function goBack() {
  if (window.history.length > 1) router.back()
  else router.push('/fare-rules')
}
</script>


<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; }
.toolbar h1 { font-size: 1.15rem; margin: 0; font-family: var(--font-display); }
.page-hint { color: var(--muted); font-size: 0.85rem; margin: 0 0 12px; line-height: 1.45; }
.empty { color: var(--muted); text-align: center; margin-top: 32px; }
.guide-line {
  margin: 0;
  padding: 10px 12px;
  border-radius: 12px;
  background: var(--primary-soft);
  border: 1px solid color-mix(in srgb, var(--primary) 18%, transparent);
  color: var(--muted);
  font-size: 0.8rem;
  line-height: 1.45;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.scenario-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}
@media (max-width: 420px) {
  .scenario-grid { grid-template-columns: 1fr; }
}
.scenario-card {
  text-align: left;
  padding: 12px 14px;
  border-radius: 14px;
  border: 1px solid var(--surface-border);
  background: rgb(var(--v-theme-surface));
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s, transform 0.15s;
}
.scenario-card:hover {
  border-color: color-mix(in srgb, var(--primary) 45%, var(--surface-border));
  box-shadow: 0 4px 14px color-mix(in srgb, var(--primary) 12%, transparent);
}
.scenario-card:active { transform: scale(0.98); }
.scenario-title {
  display: block;
  font-weight: 700;
  font-size: 0.9rem;
  color: rgb(var(--v-theme-on-surface));
  margin-bottom: 4px;
}
.scenario-hint {
  display: block;
  font-size: 0.72rem;
  line-height: 1.4;
  color: var(--muted);
}

.mod-row {
  margin-bottom: 8px;
  border-radius: 14px;
  border: 1px solid var(--surface-border);
  background: rgb(var(--v-theme-surface));
  overflow: hidden;
  transition: border-color 0.15s;
  min-width: 0;
}
.mod-row.on {
  border-color: color-mix(in srgb, var(--primary) 40%, var(--surface-border));
}
.mod-row-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  min-width: 0;
}
.mod-row-main > div:first-child {
  min-width: 0;
  flex: 1;
}
.mod-name { font-weight: 700; font-size: 0.92rem; }
.mod-hint {
  font-size: 0.72rem;
  color: var(--muted);
  margin-top: 2px;
  line-height: 1.35;
  overflow-wrap: anywhere;
  word-break: break-word;
}
.mod-body {
  padding: 0 14px 14px;
  border-top: 1px dashed var(--surface-border);
  padding-top: 12px;
}
.block-head.inner {
  margin-bottom: 8px;
}
.inner-label { font-size: 0.8rem; font-weight: 600; color: var(--muted); }

.more-toggle {
  width: 100%;
  display: grid;
  grid-template-columns: 1fr auto;
  grid-template-rows: auto auto;
  column-gap: 8px;
  align-items: center;
  text-align: left;
  padding: 12px 14px;
  border-radius: 14px;
  border: 1px dashed var(--surface-border);
  background: transparent;
  cursor: pointer;
  color: rgb(var(--v-theme-on-surface));
  font-weight: 600;
  font-size: 0.88rem;
}
.more-toggle .v-icon { grid-row: 1 / span 2; }
.more-sub {
  grid-column: 1;
  font-size: 0.72rem;
  font-weight: 400;
  color: var(--muted);
  margin-top: 2px;
  overflow-wrap: anywhere;
}
.more-body { margin-top: 12px; }

.holiday-actions { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; }
.holiday-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin: 10px 0 6px;
  font-size: 0.72rem;
  color: var(--muted);
}
.holiday-legend .leg {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.holiday-legend .leg i {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}
.holiday-legend .leg.past i { background: #9aa3ae; }
.holiday-legend .leg.now i { background: var(--primary); }
.holiday-legend .leg.soon i { background: #d4893a; }

.holiday-years {
  display: flex;
  flex-direction: column;
  gap: 14px;
  max-height: 280px;
  overflow-y: auto;
  margin-top: 4px;
}
.holiday-year-block {
  border-radius: 12px;
  border: 1px solid var(--surface-border);
  padding: 10px;
  background: color-mix(in srgb, rgb(var(--v-theme-surface)) 92%, var(--primary-soft));
}
.hy-head {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 8px;
}
.hy-year {
  font-weight: 800;
  font-size: 0.95rem;
  font-family: var(--font-display);
}
.hy-tag {
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--primary);
  padding: 1px 7px;
  border-radius: 999px;
  background: var(--primary-soft);
}
.hy-empty {
  font-size: 0.78rem;
  color: var(--muted);
  padding: 4px 2px;
}
.holiday-groups {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.holiday-group {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 8px;
  align-items: baseline;
  padding: 8px 12px;
  border-radius: 10px;
  font-size: 0.82rem;
  border: 1px solid transparent;
}
.holiday-group.past {
  background: color-mix(in srgb, #9aa3ae 14%, transparent);
  color: #7a8490;
}
.holiday-group.past .hg-name { text-decoration: line-through; text-decoration-thickness: 1px; }
.holiday-group.now {
  background: var(--primary-soft);
  border-color: color-mix(in srgb, var(--primary) 35%, transparent);
  color: rgb(var(--v-theme-on-surface));
}
.holiday-group.soon {
  background: color-mix(in srgb, #d4893a 12%, transparent);
  border-color: color-mix(in srgb, #d4893a 28%, transparent);
  color: rgb(var(--v-theme-on-surface));
}
.hg-name { font-weight: 700; }
.hg-range { color: inherit; opacity: 0.75; font-variant-numeric: tabular-nums; font-size: 0.78rem; }
.hg-status { font-weight: 700; font-size: 0.72rem; }
.holiday-group.past .hg-status { color: #8a939e; }
.holiday-group.now .hg-status { color: var(--primary); }
.holiday-group.soon .hg-status { color: #c07830; }

.block-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
  min-width: 0;
}

.item-card {
  margin-bottom: 12px;
  padding: 12px;
  border-radius: 14px;
  border: 1px solid var(--surface-border);
  background: var(--primary-soft);
  min-width: 0;
}
.item-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
  font-size: 0.86rem;
  font-weight: 700;
}

.pair {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
@media (max-width: 420px) {
  .pair { grid-template-columns: 1fr; }
}

.preview-body {
  margin-top: 12px;
  padding: 12px 14px;
  border-radius: 12px;
  background: var(--primary-soft);
  border: 1px solid var(--surface-border);
}
.preview-amount {
  font-size: 1.2rem;
  font-weight: 700;
  font-family: var(--font-display);
  margin-bottom: 4px;
  color: var(--primary);
}
.preview-reason,
.preview-meta {
  font-size: 0.82rem;
  line-height: 1.4;
  color: var(--muted);
  overflow-wrap: anywhere;
  word-break: break-word;
}
.preview-meta { margin-top: 4px; }

.edit-page { padding-bottom: calc(24px + env(safe-area-inset-bottom, 0px)); }
.toolbar {
  position: sticky;
  top: 0;
  z-index: 2;
  background: var(--bg, rgb(var(--v-theme-background)));
  padding: 4px 0;
}
.toolbar h1 { flex: 1; text-align: center; }
.edit-body { padding-bottom: 8px; }
.loading-tip { padding: 24px; text-align: center; color: var(--muted); }
</style>

<style>


</style>

