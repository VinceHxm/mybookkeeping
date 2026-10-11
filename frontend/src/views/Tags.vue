<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>{{ t('tags.title') }}</h1>
      <v-btn icon="mdi-plus" variant="text" color="primary" @click="openCreate" />
    </div>

    <div class="card-grid card-grid--compact">
    <EntityCard
      v-for="tag in tags.list"
      :key="tag.id"
      :title="tag.name"
      :dot="tag.color"
      @edit="openEdit(tag)"
      @delete="askDelete(tag)"
    >
      <template #meta>{{ t('tags.meta') }}</template>
    </EntityCard>
    </div>
    <div v-if="!tags.list.length" class="empty-tip">{{ t('tags.empty') }}</div>

    <v-dialog v-model="dialog" max-width="400">
      <v-card>
        <v-card-title>{{ editing ? t('tags.edit') : t('tags.create') }}</v-card-title>
        <v-card-text>
          <div class="form-sections">
            <section class="form-section">
              <header class="form-section__head">
                <h3 class="form-section__title">{{ t('tags.info') }}</h3>
              </header>
              <div class="form-section__body">
                <v-text-field v-model="form.name" :label="t('ui.name')" variant="outlined" hide-details class="field" />
                <div class="color-row field">
                  <v-text-field v-model="form.color" :label="t('tags.color')" type="color" variant="outlined" hide-details />
                  <div class="swatches">
                    <button
                      v-for="c in swatches"
                      :key="c"
                      type="button"
                      class="swatch"
                      :class="{ active: form.color === c }"
                      :style="{ background: c }"
                      @click="form.color = c"
                    />
                  </div>
                </div>
                <v-text-field
                  v-model.number="form.sort"
                  :label="t('ui.sort')"
                  type="number"
                  variant="outlined"
                  hide-details
                  class="field field--last"
                />
              </div>
            </section>

            <div v-if="editing" class="danger-zone">
              <div class="danger-title">{{ t('ui.dangerZone') }}</div>
              <p class="danger-tip">{{ t('tags.deleteTip') }}</p>
              <v-btn color="error" variant="outlined" block @click="askDelete(editing)">{{ t('tags.deleteThis') }}</v-btn>
            </div>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">{{ t('common.cancel') }}</v-btn>
          <v-btn color="primary" @click="onSave">{{ t('common.save') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <DeleteConfirmDialog
      v-model="confirmOpen"
      :name="pendingDelete?.name || ''"
      :loading="deleting"
      @confirm="doDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import EntityCard from '../components/EntityCard.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import { useTagStore, type Tag } from '../stores/tag'

const { t } = useI18n()
const tags = useTagStore()
const dialog = ref(false)
const editing = ref<Tag | null>(null)
const confirmOpen = ref(false)
const pendingDelete = ref<Tag | null>(null)
const deleting = ref(false)
const form = reactive({ name: '', color: '#1b7f5a', sort: 0 })
const swatches = ['#1b7f5a', '#3d9b74', '#c45c3e', '#d4a017', '#2e7d6f', '#5c6bc0', '#8e24aa', '#546e7a']

onMounted(() => tags.load())

function openCreate() {
  editing.value = null
  form.name = ''
  form.color = '#1b7f5a'
  form.sort = 0
  dialog.value = true
}
function openEdit(tag: Tag) {
  editing.value = tag
  form.name = tag.name
  form.color = tag.color
  form.sort = tag.sort
  dialog.value = true
}
async function onSave() {
  if (editing.value) await tags.update(editing.value.id, { ...form })
  else await tags.create({ ...form })
  dialog.value = false
}

function askDelete(tag: Tag) {
  pendingDelete.value = tag
  confirmOpen.value = true
}

async function doDelete() {
  if (!pendingDelete.value) return
  deleting.value = true
  try {
    await tags.remove(pendingDelete.value.id)
    confirmOpen.value = false
    if (editing.value?.id === pendingDelete.value.id) dialog.value = false
    pendingDelete.value = null
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; }
.toolbar h1 { font-size: 1.15rem; margin: 0; }
.color-row { display: grid; gap: 10px; min-width: 0; }
.swatches { display: flex; flex-wrap: wrap; gap: 8px; }
.swatch {
  width: 28px; height: 28px; border-radius: 50%; border: 2px solid transparent;
  cursor: pointer; padding: 0;
}
.swatch.active { border-color: rgb(var(--v-theme-on-surface)); box-shadow: 0 0 0 2px var(--primary-soft); }
</style>
