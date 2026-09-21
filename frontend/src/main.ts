import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { zhHans } from 'vuetify/locale'

import './styles.css'
import { getStoredTheme, resolveThemeMode, syncDocumentTheme } from './composables/useAppTheme'

const bootMode = syncDocumentTheme(getStoredTheme())

const vuetify = createVuetify({
  components,
  directives,
  locale: {
    locale: 'zhHans',
    fallback: 'en',
    messages: { zhHans },
  },
  defaults: {
    VBtn: { rounded: 'lg' },
    VCard: { rounded: 'lg', elevation: 0 },
    VList: { rounded: 'lg' },
    VTextField: { rounded: 'lg', color: 'primary' },
    VSelect: { rounded: 'lg', color: 'primary' },
    VChip: { rounded: 'lg' },
    VDatePicker: { color: 'primary', rounded: 'lg' },
  },
  theme: {
    defaultTheme: bootMode,
    themes: {
      light: {
        dark: false,
        colors: {
          primary: '#1B7F5A',
          secondary: '#3D5A4C',
          accent: '#D4A017',
          background: '#EEF5F0',
          surface: '#FFFFFF',
          'surface-bright': '#FFFFFF',
          'surface-variant': '#E2EFE7',
          'on-surface': '#15261F',
          'on-background': '#15261F',
          error: '#C62828',
          info: '#2E7D6F',
          success: '#1B7F5A',
          warning: '#D4A017',
        },
      },
      dark: {
        dark: true,
        colors: {
          primary: '#3D9B74',
          secondary: '#9BB5A8',
          accent: '#E0B84A',
          background: '#0E1612',
          surface: '#1C2620',
          'surface-bright': '#24302A',
          'surface-variant': '#24302A',
          'on-surface': '#E6F0EA',
          'on-background': '#E6F0EA',
          error: '#EF5350',
          info: '#5CB89A',
          success: '#3D9B74',
          warning: '#E0B84A',
        },
      },
    },
  },
})

createApp(App).use(createPinia()).use(router).use(vuetify).mount('#app')

// 确保与 document 同步（防止启动竞态）
vuetify.theme.global.name.value = resolveThemeMode(getStoredTheme())
