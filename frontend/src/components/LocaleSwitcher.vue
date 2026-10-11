<template>
  <v-menu location="bottom">
    <template #activator="{ props: menuProps }">
      <v-btn v-bind="menuProps" :size="size" variant="text" color="primary" prepend-icon="mdi-translate">
        {{ currentLabel }}
      </v-btn>
    </template>
    <v-list density="compact" min-width="160">
      <v-list-item
        v-for="opt in SUPPORTED_LOCALES"
        :key="opt.value"
        :active="opt.value === locale"
        color="primary"
        :title="opt.label"
        @click="choose(opt.value)"
      />
    </v-list>
  </v-menu>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { SUPPORTED_LOCALES, setLocale, type AppLocale } from '../i18n'
import { useAuthStore } from '../stores/auth'

withDefaults(defineProps<{ size?: string }>(), { size: 'small' })

const { locale } = useI18n()
const auth = useAuthStore()

const currentLabel = computed(() => SUPPORTED_LOCALES.find((o) => o.value === locale.value)?.label || locale.value)

function choose(l: AppLocale) {
  if (l === locale.value) return
  setLocale(l)
  if (auth.token) auth.updateSettings({ locale: l }).catch(() => {})
}
</script>
