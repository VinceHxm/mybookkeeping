<template>
  <div class="meta-head">
    <div class="meta-title-row">
      <v-icon v-if="icon" size="20" color="primary">{{ icon }}</v-icon>
      <div class="meta-copy">
        <div class="meta-title">{{ title }}</div>
        <div v-if="subtitle" class="meta-sub">{{ subtitle }}</div>
      </div>
    </div>
    <slot />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  icon?: string
  title: string
  subtitle?: string
}>()
</script>

<style scoped>
.meta-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  min-width: 0;
}
.meta-head > :slotted(*) { flex-shrink: 0; }
.meta-title-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  min-width: 0;
  flex: 1;
}
.meta-copy { min-width: 0; flex: 1; }
.meta-title { font-weight: 700; font-size: 0.98rem; line-height: 1.2; overflow-wrap: anywhere; }
.meta-sub {
  color: var(--muted);
  font-size: 0.8rem;
  margin-top: 2px;
  line-height: 1.4;
  overflow-wrap: anywhere;
  word-break: break-word;
}
</style>
