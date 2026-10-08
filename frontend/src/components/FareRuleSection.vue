<template>
  <section class="form-section mb-3">
    <div class="form-section__body">
      <MetaHead
        icon="mdi-ticket-percent-outline"
        title="计费规则"
        subtitle="主要用于累计折扣标记；金额以手填为准"
        :class="{ 'mb-3': enabled }"
      >
        <v-switch
          v-model="enabled"
          color="primary"
          density="compact"
          hide-details
          inset
          @update:model-value="emit('toggle', !!$event)"
        />
      </MetaHead>
      <template v-if="enabled">
        <v-select
          v-model="ruleId"
          :items="items"
          item-title="title"
          item-value="value"
          label="选择规则"
          variant="outlined"
          density="compact"
          hide-details
          class="field"
          @update:model-value="emit('picked')"
        />
        <div class="fare-actions">
          <v-btn size="small" variant="tonal" color="primary" :loading="loading" :disabled="!ruleId" @click="emit('apply')">
            填入建议价
          </v-btn>
          <span v-if="hint" class="fare-hint">{{ hint }}</span>
        </div>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import MetaHead from './MetaHead.vue'

defineProps<{
  items: { title: string; value: number }[]
  loading?: boolean
  hint?: string
}>()

const enabled = defineModel<boolean>('enabled', { required: true })
const ruleId = defineModel<number | null>('ruleId', { required: true })

const emit = defineEmits<{
  toggle: [on: boolean]
  picked: []
  apply: []
}>()
</script>

<style scoped>
.fare-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.fare-hint {
  font-size: 0.75rem;
  color: var(--muted);
  line-height: 1.35;
  overflow-wrap: anywhere;
  word-break: break-word;
  min-width: 0;
  flex: 1 1 140px;
}
</style>
