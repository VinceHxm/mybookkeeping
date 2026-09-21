<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{
  type?: 'date' | 'datetime-local' | 'time'
}>(), {
  type: 'date',
})

const model = defineModel<string>({ default: '' })
const open = ref(false)
const pickerDate = ref<Date | null>(null)
const hour = ref(0)
const minute = ref(0)

const hourItems = Array.from({ length: 24 }, (_, i) => ({
  title: String(i).padStart(2, '0'),
  value: i,
}))
const minuteItems = Array.from({ length: 60 }, (_, i) => ({
  title: String(i).padStart(2, '0'),
  value: i,
}))

const displayText = computed(() => {
  if (!model.value) return ''
  if (props.type === 'date') {
    const d = dayjs(model.value)
    return d.isValid() ? d.format('YYYY年M月D日') : model.value
  }
  if (props.type === 'time') {
    const d = dayjs(`2000-01-01T${model.value}`)
    return d.isValid() ? d.format('HH:mm') : model.value
  }
  const d = dayjs(model.value)
  return d.isValid() ? d.format('YYYY年M月D日 HH:mm') : model.value
})

const placeholder = computed(() => {
  if (props.type === 'time') return '选择时间'
  if (props.type === 'datetime-local') return '选择日期时间'
  return '选择日期'
})

function syncFromModel() {
  if (props.type === 'time') {
    const d = dayjs(`2000-01-01T${model.value || '00:00'}`)
    hour.value = d.isValid() ? d.hour() : 0
    minute.value = d.isValid() ? d.minute() : 0
    pickerDate.value = null
    return
  }
  const raw = model.value || dayjs().format(props.type === 'date' ? 'YYYY-MM-DD' : 'YYYY-MM-DDTHH:mm')
  const d = dayjs(raw)
  if (!d.isValid()) {
    pickerDate.value = new Date()
    hour.value = dayjs().hour()
    minute.value = dayjs().minute()
    return
  }
  pickerDate.value = d.toDate()
  hour.value = d.hour()
  minute.value = d.minute()
}

watch(open, (v) => {
  if (v) syncFromModel()
})

function onPick(val: unknown) {
  if (val instanceof Date) {
    pickerDate.value = val
    return
  }
  if (typeof val === 'string' && val) {
    const d = dayjs(val)
    if (d.isValid()) pickerDate.value = d.toDate()
  }
}

function clear() {
  model.value = ''
  open.value = false
}

function setToday() {
  const now = dayjs()
  pickerDate.value = now.toDate()
  hour.value = now.hour()
  minute.value = now.minute()
  if (props.type === 'date') {
    model.value = now.format('YYYY-MM-DD')
    open.value = false
  } else if (props.type === 'time') {
    model.value = now.format('HH:mm')
    open.value = false
  }
}

function confirm() {
  if (props.type === 'time') {
    model.value = `${String(hour.value).padStart(2, '0')}:${String(minute.value).padStart(2, '0')}`
    open.value = false
    return
  }
  const base = pickerDate.value ? dayjs(pickerDate.value) : dayjs()
  if (props.type === 'date') {
    model.value = base.format('YYYY-MM-DD')
  } else {
    model.value = base
      .hour(hour.value)
      .minute(minute.value)
      .second(0)
      .format('YYYY-MM-DDTHH:mm')
  }
  open.value = false
}
</script>

<template>
  <div class="app-date-field">
    <v-text-field
      :model-value="displayText"
      :placeholder="placeholder"
      readonly
      append-inner-icon="mdi-calendar-month-outline"
      v-bind="$attrs"
      @click="open = true"
      @click:append-inner="open = true"
    />

    <v-dialog
      v-model="open"
      max-width="360"
      content-class="app-date-dialog"
      scrim="rgba(14, 22, 18, 0.55)"
    >
      <v-card class="picker-card" rounded="xl">
        <div class="picker-title">
          {{ type === 'time' ? '选择时间' : type === 'datetime-local' ? '选择日期时间' : '选择日期' }}
        </div>

        <v-date-picker
          v-if="type !== 'time'"
          :model-value="pickerDate"
          color="primary"
          rounded="lg"
          width="100%"
          show-adjacent-months
          hide-header
          :first-day-of-week="1"
          @update:model-value="onPick"
        />

        <div v-if="type === 'datetime-local' || type === 'time'" class="time-row">
          <v-select
            v-model="hour"
            :items="hourItems"
            label="时"
            variant="outlined"
            density="compact"
            hide-details
          />
          <span class="time-sep">:</span>
          <v-select
            v-model="minute"
            :items="minuteItems"
            label="分"
            variant="outlined"
            density="compact"
            hide-details
          />
        </div>

        <div class="picker-actions">
          <v-btn variant="text" color="primary" @click="clear">清除</v-btn>
          <v-spacer />
          <v-btn v-if="type !== 'time'" variant="text" color="primary" @click="setToday">今天</v-btn>
          <v-btn color="primary" variant="flat" class="confirm-btn" @click="confirm">确定</v-btn>
        </div>
      </v-card>
    </v-dialog>
  </div>
</template>

<style scoped>
.app-date-field {
  width: 100%;
  flex: 1;
  min-width: 0;
}
.app-date-field :deep(input) {
  cursor: pointer;
}

.picker-card {
  background: rgb(var(--v-theme-surface)) !important;
  color: rgb(var(--v-theme-on-surface));
  overflow: hidden;
  border: 1px solid var(--surface-border, rgba(128, 128, 128, 0.22));
  box-shadow: 0 16px 40px rgba(14, 22, 18, 0.28);
}

.picker-title {
  padding: 16px 18px 4px;
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.picker-card :deep(.v-date-picker) {
  background: transparent !important;
}
.picker-card :deep(.v-picker-title),
.picker-card :deep(.v-date-picker-controls) {
  padding-inline: 8px;
}
.picker-card :deep(.v-date-picker-month__day) {
  border-radius: 10px;
}
.picker-card :deep(.v-btn--selected),
.picker-card :deep(.v-date-picker-month__day--selected .v-btn) {
  border-radius: 10px;
}

.time-row {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 8px;
  align-items: center;
  padding: 4px 16px 8px;
}
.time-sep {
  font-weight: 700;
  color: var(--muted, rgba(128, 128, 128, 0.8));
  padding-top: 4px;
}

.picker-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px 14px;
  border-top: 1px solid var(--surface-border, rgba(128, 128, 128, 0.18));
}
.confirm-btn {
  min-width: 76px;
}
</style>
