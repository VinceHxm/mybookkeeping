<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>用户管理</h1>
      <span class="spacer" />
    </div>

    <p class="page-tip">仅管理员可见。可调整角色、停用或删除注册用户（不可操作自己；须保留至少一名可用管理员）。</p>

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
        <v-chip v-if="u.disabled" size="x-small" variant="tonal" color="error">已停用</v-chip>
        <v-chip
          size="x-small"
          variant="tonal"
          :color="u.role === 'admin' ? 'primary' : undefined"
        >{{ u.role === 'admin' ? '管理员' : '普通用户' }}</v-chip>
        <v-chip v-if="u.id === selfId" size="x-small" variant="outlined">我</v-chip>
      </template>
      <template #meta>
        <div>{{ u.email || '未填邮箱' }}</div>
        <div v-if="u.createdAt" class="meta-sub">注册 {{ formatTime(u.createdAt) }}</div>
      </template>
    </EntityCard>
    <div v-if="!users.list.length && !loading" class="empty-tip">暂无用户</div>

    <v-dialog v-model="dialog" max-width="440">
      <v-card>
        <v-card-title>编辑用户</v-card-title>
        <v-card-text class="form-fields">
          <v-text-field
            :model-value="editing?.username || ''"
            label="用户名"
            variant="outlined"
            hide-details
            readonly
            class="field"
          />
          <v-select
            v-model="form.role"
            :items="roleItems"
            label="角色"
            variant="outlined"
            hide-details
            class="field"
            :disabled="isSelf"
          />
          <v-switch
            v-model="form.disabled"
            label="停用账户（无法登录，已登录会话立即失效）"
            hide-details
            color="error"
            class="field"
            :disabled="isSelf"
          />
          <p v-if="isSelf" class="field-tip">不能修改自己的角色或停用自己。</p>
          <p v-else class="field-tip">停用后可再启用；删除会清空该用户全部账本数据且不可恢复。</p>

          <div v-if="editing && !isSelf" class="danger-zone">
            <div class="danger-title">危险操作</div>
            <p class="danger-tip">删除用户将级联清除其账户、流水、模板等全部数据。</p>
            <v-btn color="error" variant="outlined" block @click="askDelete(editing)">删除此用户</v-btn>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">取消</v-btn>
          <v-btn color="primary" :loading="saving" :disabled="isSelf" @click="onSave">保存</v-btn>
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
import dayjs from 'dayjs'
import EntityCard from '../components/EntityCard.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import { useAuthStore } from '../stores/auth'
import { useAdminUsersStore, type AdminUser } from '../stores/adminUsers'

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
const roleItems = [
  { title: '普通用户', value: 'user' },
  { title: '管理员', value: 'admin' },
]

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
  dialog.value = true
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
    alert(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

function askDelete(u: AdminUser) {
  if (u.id === selfId.value) {
    alert('不能删除自己')
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
    alert(e?.message || '删除失败')
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
}
.meta-sub { margin-top: 2px; font-size: 0.75rem; opacity: 0.9; }
.form-fields .field { margin-bottom: 14px; }
.field-tip {
  margin: -6px 0 14px;
  padding: 0 2px;
  font-size: 0.78rem;
  color: var(--muted);
  line-height: 1.45;
}
.danger-zone {
  margin-top: 8px;
  padding: 16px;
  border-radius: 14px;
  border: 1px solid rgba(198, 40, 40, 0.35);
  background: rgba(198, 40, 40, 0.08);
}
.danger-title { font-weight: 700; color: #c62828; margin-bottom: 6px; }
.danger-tip { margin: 0 0 12px; font-size: 0.8rem; color: var(--muted); line-height: 1.4; }
</style>
