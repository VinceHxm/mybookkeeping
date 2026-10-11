<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>{{ t('adminUsers.title') }}</h1>
      <span class="spacer" />
    </div>

    <p class="page-tip">{{ t('adminUsers.pageTip') }}</p>

    <div class="card-grid">
    <EntityCard
      v-for="u in users.list"
      :key="u.id"
      :title="u.username"
      icon="mdi-account-outline"
      :dim="u.disabled"
      @edit="openEdit(u)"
      @delete="askDelete(u)"
    >
      <template #badges>
        <v-chip v-if="u.disabled" size="x-small" variant="tonal" color="error">{{ t('adminUsers.disabled') }}</v-chip>
        <v-chip
          size="x-small"
          variant="tonal"
          :color="u.role === 'admin' ? 'primary' : undefined"
        >{{ u.role === 'admin' ? t('adminUsers.admin') : t('adminUsers.user') }}</v-chip>
        <v-chip v-if="u.id === selfId" size="x-small" variant="outlined">{{ t('adminUsers.self') }}</v-chip>
      </template>
      <template #meta>
        <div>{{ u.email || t('adminUsers.noEmail') }}</div>
        <div v-if="u.createdAt" class="meta-sub">{{ t('adminUsers.registeredAt', { time: formatTime(u.createdAt) }) }}</div>
      </template>
    </EntityCard>
    </div>
    <div v-if="!users.list.length && !loading" class="empty-tip">{{ t('adminUsers.empty') }}</div>

    <v-dialog v-model="dialog" max-width="440">
      <v-card>
        <v-card-title>{{ t('adminUsers.edit') }}</v-card-title>
        <v-card-text>
          <div class="form-sections">
            <section class="form-section">
              <header class="form-section__head">
                <h3 class="form-section__title">{{ t('adminUsers.accountSettings') }}</h3>
              </header>
              <div class="form-section__body">
                <v-text-field
                  :model-value="editing?.username || ''"
                  :label="t('auth.username')"
                  variant="outlined"
                  hide-details
                  readonly
                  class="field"
                />
                <v-select
                  v-model="form.role"
                  :items="roleItems"
                  :label="t('adminUsers.role')"
                  variant="outlined"
                  hide-details
                  class="field"
                  :disabled="isSelf"
                />
                <v-switch
                  v-model="form.disabled"
                  :label="t('adminUsers.disableSwitch')"
                  hide-details
                  color="error"
                  class="field field--last"
                  :disabled="isSelf"
                />
                <p v-if="isSelf" class="form-section__tip">{{ t('adminUsers.selfTip') }}</p>
                <p v-else class="form-section__tip">{{ t('adminUsers.tip') }}</p>
              </div>
            </section>

            <section v-if="editing && !isSelf" class="form-section">
              <header class="form-section__head">
                <h3 class="form-section__title">{{ t('adminUsers.resetPassword') }}</h3>
              </header>
              <div class="form-section__body">
                <v-text-field
                  v-model="newPassword"
                  :label="t('adminUsers.newPasswordMin', { n: MIN_PASSWORD_LEN })"
                  type="password"
                  autocomplete="new-password"
                  variant="outlined"
                  hide-details
                  class="field"
                />
                <v-btn
                  color="primary"
                  variant="tonal"
                  block
                  :loading="resetting"
                  :disabled="newPassword.length < MIN_PASSWORD_LEN"
                  @click="onResetPassword"
                >{{ t('adminUsers.resetPassword') }}</v-btn>
                <p class="form-section__tip">{{ t('adminUsers.resetTip') }}</p>
              </div>
            </section>

            <div v-if="editing && !isSelf" class="danger-zone">
              <div class="danger-title">{{ t('ui.dangerZone') }}</div>
              <p class="danger-tip">{{ t('adminUsers.deleteTip') }}</p>
              <v-btn color="error" variant="outlined" block @click="askDelete(editing)">{{ t('adminUsers.deleteThis') }}</v-btn>
            </div>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">{{ t('common.cancel') }}</v-btn>
          <v-btn color="primary" :loading="saving" :disabled="isSelf" @click="onSave">{{ t('common.save') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <DeleteConfirmDialog
      v-model="confirmOpen"
      :name="pendingDelete?.username || ''"
      :loading="deleting"
      @confirm="doDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import EntityCard from '../components/EntityCard.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import { useAuthStore } from '../stores/auth'
import { useAdminUsersStore, type AdminUser } from '../stores/adminUsers'

const MIN_PASSWORD_LEN = 8

const { t } = useI18n()
const auth = useAuthStore()
const users = useAdminUsersStore()
const router = useRouter()
const loading = ref(false)
const dialog = ref(false)
const editing = ref<AdminUser | null>(null)
const saving = ref(false)
const confirmOpen = ref(false)
const pendingDelete = ref<AdminUser | null>(null)
const deleting = ref(false)
const form = reactive({ role: 'user' as 'user' | 'admin', disabled: false })
const newPassword = ref('')
const resetting = ref(false)
const roleItems = computed(() => [
  { title: t('adminUsers.user'), value: 'user' },
  { title: t('adminUsers.admin'), value: 'admin' },
])

const selfId = computed(() => auth.profile?.id)
const isSelf = computed(() => !!editing.value && editing.value.id === selfId.value)

function formatTime(iso: string) {
  return dayjs(iso).format('YYYY-MM-DD HH:mm')
}

async function reload() {
  loading.value = true
  try {
    await users.load()
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  if (!auth.profile) {
    try {
      await auth.fetchMe()
    } catch {
      await router.replace('/login')
      return
    }
  }
  if (!auth.isAdmin()) {
    await router.replace('/mine')
    return
  }
  await reload()
})

function openEdit(u: AdminUser) {
  editing.value = u
  form.role = u.role === 'admin' ? 'admin' : 'user'
  form.disabled = !!u.disabled
  newPassword.value = ''
  dialog.value = true
}

async function onResetPassword() {
  if (!editing.value || isSelf.value) return
  resetting.value = true
  try {
    await users.resetPassword(editing.value.id, newPassword.value)
    newPassword.value = ''
    alert(t('adminUsers.resetDone', { name: editing.value.username }))
  } catch (e: any) {
    alert(e?.message || t('adminUsers.resetFailed'))
  } finally {
    resetting.value = false
  }
}

async function onSave() {
  if (!editing.value || isSelf.value) return
  saving.value = true
  try {
    await users.update(editing.value.id, {
      role: form.role,
      disabled: form.disabled,
    })
    dialog.value = false
  } catch (e: any) {
    alert(e?.message || t('adminUsers.saveFailed'))
  } finally {
    saving.value = false
  }
}

function askDelete(u: AdminUser) {
  if (u.id === selfId.value) {
    alert(t('adminUsers.cannotDeleteSelf'))
    return
  }
  pendingDelete.value = u
  confirmOpen.value = true
}

async function doDelete() {
  if (!pendingDelete.value) return
  deleting.value = true
  try {
    await users.remove(pendingDelete.value.id)
    confirmOpen.value = false
    if (editing.value?.id === pendingDelete.value.id) dialog.value = false
    pendingDelete.value = null
  } catch (e: any) {
    alert(e?.message || t('ui.deleteFailed'))
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.toolbar h1 { font-size: 1.15rem; margin: 0; flex: 1; text-align: center; }
.spacer { width: 40px; }
.page-tip {
  margin: 0 0 12px;
  font-size: 0.82rem;
  color: var(--muted);
  line-height: 1.45;
  overflow-wrap: anywhere;
}
.meta-sub { margin-top: 2px; font-size: 0.75rem; opacity: 0.9; }
</style>
