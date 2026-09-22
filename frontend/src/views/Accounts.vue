<template>
  <div class="page">
    <div class="toolbar">
      <v-btn icon="mdi-arrow-left" variant="text" to="/mine" />
      <h1>账户管理</h1>
      <v-btn icon="mdi-plus" variant="text" color="primary" @click="openCreate" />
    </div>

    <v-switch v-model="showArchived" label="显示已归档" hide-details class="mb-2" density="compact" @update:model-value="reload" />

    <EntityCard
      v-for="a in accounts.list"
      :key="a.id"
      :title="a.name"
      :icon="accountTypeIcon(a.type)"
      :dim="a.archived"
      @edit="openEdit(a)"
      @delete="askDelete(a)"
    >
      <template #badges>
        <v-chip v-if="a.archived" size="x-small" variant="tonal">已归档</v-chip>
        <v-chip size="x-small" variant="tonal" color="primary">{{ typeMap[a.type] || a.type }}</v-chip>
      </template>
      <template #meta>
        <div v-if="hasProfile(a)" class="profile-row">
          <span v-if="a.institution" class="profile-item">{{ a.institution }}</span>
          <span v-if="revealCache[a.id]?.cardNo || a.cardNo" class="profile-item mono">
            {{ revealCache[a.id]?.cardNo || displayCardNo(a.cardNo) }}
          </span>
          <span v-if="revealCache[a.id]?.holderName || a.holderName" class="profile-item">
            {{ revealCache[a.id]?.holderName || displayHolderName(a.holderName) }}
          </span>
          <button
            v-if="a.cardNo || a.holderName"
            type="button"
            class="reveal-btn no-card-nav"
            :aria-label="revealCache[a.id] ? '隐藏完整信息' : '显示完整信息'"
            :disabled="revealLoading[a.id]"
            @click.stop="toggleReveal(a)"
            @pointerdown.stop
          >
            <v-icon size="16">{{ revealCache[a.id] ? 'mdi-eye-off-outline' : 'mdi-eye-outline' }}</v-icon>
          </button>
        </div>

        <template v-if="a.type === 'credit'">
          <div class="bal-line">已用额度 ¥{{ fenToYuan(creditUsedFen(a)) }}</div>
          <div v-if="a.creditLimit" class="meta-sub">额度 ¥{{ fenToYuan(a.creditLimit) }} · 可用 ¥{{ fenToYuan(Math.max(0, (a.creditLimit || 0) - creditUsedFen(a))) }}</div>
          <div class="split-line">
            <span>已出账 ¥{{ fenToYuan(displayBilledFen(a)) }}</span>
            <span>未出账 ¥{{ fenToYuan(displayUnbilledFen(a)) }}</span>
          </div>
          <div v-if="a.billingDay" class="meta-sub">账单日 {{ a.billingDay }} 号<template v-if="a.paymentDueDay"> · 还款日 {{ a.paymentDueDay }} 号</template></div>
          <div
            v-if="a.billingDay && statements[a.id]"
            class="stmt-box"
          >
            <div class="stmt-head">
              <span>本期账单</span>
              <span class="stmt-range">{{ statements[a.id].cycleStart }} ~ {{ statements[a.id].cycleEnd }}</span>
            </div>
            <div class="stmt-grid">
              <div class="stmt-cell">
                <span>本期已还</span>
                <strong>¥{{ fenToYuan(statements[a.id].repaidAmount) }}</strong>
              </div>
              <div class="stmt-cell">
                <span>本期待还</span>
                <strong>¥{{ fenToYuan(statements[a.id].periodRemaining) }}</strong>
              </div>
              <div class="stmt-cell">
                <span>{{ statements[a.id].periodRemaining === 0 && (statements[a.id].remainingAfterPay || 0) > 0 ? '下期待还' : '还后待还' }}</span>
                <strong class="warn">¥{{ fenToYuan(statements[a.id].remainingAfterPay) }}</strong>
              </div>
              <div class="stmt-cell">
                <span>整体待还</span>
                <strong>¥{{ fenToYuan(statements[a.id].totalOutstanding ?? statements[a.id].outstandingBalance) }}</strong>
              </div>
            </div>
            <div class="stmt-sub">
              <template v-if="statements[a.id].periodRemaining === 0 && (statements[a.id].remainingAfterPay || 0) > 0">
                本期已还清；剩余已用为下期/未出账
              </template>
              <template v-else>
                本期待还 = 本期账单尚未还清部分；还后待还 = 未出账；整体待还 = 已用额度
              </template>
              <template v-if="statements[a.id].futureInstallment">
                · 后续分期 ¥{{ fenToYuan(statements[a.id].futureInstallment) }}
              </template>
            </div>
          </div>
          <div v-else-if="a.billingDay && stmtError[a.id]" class="stmt-err">
            {{ stmtError[a.id] }}
          </div>
          <router-link
            v-if="!a.archived"
            class="repay-link no-card-nav"
            :to="repayLink(a)"
            @click.stop
            @pointerdown.stop
          >去还款</router-link>
        </template>
        <template v-else>
          <div class="bal-line">余额 ¥{{ fenToYuan(a.balance) }}</div>
        </template>

        <div v-if="a.storageNote || (a.attachments && a.attachments.length)" class="storage-block">
          <div v-if="a.storageNote" class="storage-note">
            <v-icon size="14" class="storage-ico">mdi-map-marker-outline</v-icon>
            <span>{{ a.storageNote }}</span>
          </div>
          <div v-if="a.attachments?.length" class="storage-thumbs">
            <button
              v-for="(att, idx) in a.attachments"
              :key="att.id"
              type="button"
              class="storage-thumb no-card-nav"
              @click.stop="openStoragePreview(a, idx)"
              @pointerdown.stop
            >
              <img :src="att.url" alt="存放位置" />
            </button>
          </div>
        </div>
      </template>
    </EntityCard>
    <div v-if="!accounts.list.length" class="empty-tip">暂无账户，点右上角添加</div>

    <!-- 宽屏：弹窗编辑 -->
    <v-dialog v-if="!isMobile" v-model="dialog" max-width="480" scrollable>
      <v-card>
        <v-card-title>{{ editing ? '编辑账户' : '新建账户' }}</v-card-title>
        <v-card-text>
          <AccountEditForm
            v-if="dialog"
            ref="dialogFormRef"
            :account-id="editing?.id ?? null"
            :initial="editing"
            @saved="onDialogSaved"
            @reconciled="onDialogSaved"
            @delete="editing && askDelete(editing)"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">取消</v-btn>
          <v-btn color="primary" :loading="dialogFormRef?.saving" @click="onDialogSave">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <Teleport to="body">
      <div
        v-if="storagePreview"
        class="lightbox"
        role="dialog"
        aria-modal="true"
        @click.self="storagePreview = null"
      >
        <button type="button" class="lb-close" aria-label="关闭" @click="storagePreview = null">
          <v-icon size="22">mdi-close</v-icon>
        </button>
        <img class="lb-img" :src="storagePreview.url" alt="存放位置大图" />
      </div>
    </Teleport>

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
import { useRouter } from 'vue-router'
import EntityCard from '../components/EntityCard.vue'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog.vue'
import AccountEditForm from '../components/AccountEditForm.vue'
import {
  useAccountStore,
  creditUsedFen,
  creditBilledFenOf,
  creditUnbilledFen,
  type Account,
  type CreditStatement,
} from '../stores/account'
import { fenToYuan } from '../utils/money'
import { accountTypeIcon } from '../utils/icons'
import { displayCardNo, displayHolderName } from '../utils/mask'
import { useBreakpoint } from '../composables/useBreakpoint'

const accounts = useAccountStore()
const router = useRouter()
const { isMobile } = useBreakpoint()

const dialog = ref(false)
const editing = ref<Account | null>(null)
const dialogFormRef = ref<InstanceType<typeof AccountEditForm> | null>(null)
const showArchived = ref(true)
const confirmOpen = ref(false)
const pendingDelete = ref<Account | null>(null)
const deleting = ref(false)
/** 临时明文，隐藏即删除，不写入 list */
const revealCache = reactive<Record<number, { cardNo: string; holderName: string }>>({})
const revealLoading = reactive<Record<number, boolean>>({})
const storagePreview = ref<{ url: string } | null>(null)
const typeMap: Record<string, string> = { cash: '现金', bank: '银行卡', credit: '信用', other: '其他' }
const statements = ref<Record<number, CreditStatement>>({})
const stmtError = ref<Record<number, string>>({})

function toggleReveal(a: Account) {
  if (revealCache[a.id]) {
    delete revealCache[a.id]
    return
  }
  void loadReveal(a.id)
}

async function loadReveal(id: number) {
  if (revealLoading[id]) return
  revealLoading[id] = true
  try {
    const data = await accounts.fetchSensitive(id)
    revealCache[id] = {
      cardNo: data.cardNo || '',
      holderName: data.holderName || '',
    }
  } catch (e: any) {
    alert(e?.message || '获取敏感信息失败')
  } finally {
    revealLoading[id] = false
  }
}

function hasProfile(a: Account) {
  return !!(a.institution || a.cardNo || a.holderName)
}

function openStoragePreview(a: Account, idx: number) {
  const att = a.attachments?.[idx]
  if (att) storagePreview.value = { url: att.url }
}

async function reload() {
  await accounts.load(showArchived.value)
  await loadStatements()
}

async function loadStatements() {
  const next: Record<number, CreditStatement> = {}
  const errs: Record<number, string> = {}
  await Promise.all(
    accounts.list
      .filter((a) => a.type === 'credit' && a.billingDay && !a.archived)
      .map(async (a) => {
        try {
          next[a.id] = await accounts.creditStatement(a.id)
        } catch (e: any) {
          errs[a.id] = e?.message || '账单加载失败'
        }
      }),
  )
  statements.value = next
  stmtError.value = errs
}

onMounted(reload)

function repayLink(a: Account) {
  const stmt = statements.value[a.id]
  const q: Record<string, string> = {
    mode: 'repay',
    toAccountId: String(a.id),
  }
  const periodLeft = stmt?.periodRemaining ?? 0
  if (periodLeft > 0) {
    q.amountFen = String(periodLeft)
  }
  // 本期已还清：不预填已出账/应还，避免把下期未出账当成本期还款额
  return { path: '/transactions/new', query: q }
}

function displayBilledFen(a: Account): number {
  const stmt = statements.value[a.id]
  if (stmt && stmt.periodRemaining === 0 && (stmt.remainingAfterPay || 0) > 0) {
    return 0
  }
  return creditBilledFenOf(a)
}

function displayUnbilledFen(a: Account): number {
  const stmt = statements.value[a.id]
  if (stmt && stmt.periodRemaining === 0) {
    return stmt.remainingAfterPay || stmt.unbilledOutstanding || creditUsedFen(a)
  }
  return creditUnbilledFen(a)
}

function openCreate() {
  if (isMobile.value) {
    router.push('/accounts/new')
    return
  }
  editing.value = null
  dialog.value = true
}

function openEdit(a: Account) {
  if (isMobile.value) {
    router.push(`/accounts/${a.id}/edit`)
    return
  }
  editing.value = a
  dialog.value = true
}

async function onDialogSave() {
  await dialogFormRef.value?.save()
}

async function onDialogSaved() {
  dialog.value = false
  await reload()
}

function askDelete(a: Account) {
  pendingDelete.value = a
  confirmOpen.value = true
}

async function doDelete() {
  if (!pendingDelete.value) return
  deleting.value = true
  try {
    await accounts.remove(pendingDelete.value.id)
    confirmOpen.value = false
    if (editing.value?.id === pendingDelete.value.id) dialog.value = false
    pendingDelete.value = null
    await reload()
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; }
.toolbar h1 { font-size: 1.15rem; margin: 0; }
.bal-line { font-weight: 700; color: rgb(var(--v-theme-on-surface)); }
.meta-sub { font-size: 0.78rem; color: var(--muted); margin-top: 2px; }
.split-line {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 4px;
  font-size: 0.8rem;
  font-variant-numeric: tabular-nums;
}

.profile-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px 8px;
  margin-bottom: 6px;
  font-size: 0.78rem;
  color: var(--muted);
  /* 不拉满整行，避免空白区被父级误当成「禁导航」热区 */
  width: fit-content;
  max-width: 100%;
}
.profile-item.mono {
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.04em;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}
.reveal-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 28px;
  height: 28px;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--primary);
  cursor: pointer;
  border-radius: 6px;
  -webkit-tap-highlight-color: transparent;
}
.reveal-btn:active { opacity: 0.7; }
.reveal-btn:disabled { opacity: 0.45; cursor: wait; }

.storage-block { margin-top: 8px; }
.storage-note {
  display: flex;
  align-items: flex-start;
  gap: 4px;
  font-size: 0.78rem;
  color: var(--muted);
  line-height: 1.4;
}
.storage-ico { flex-shrink: 0; margin-top: 1px; opacity: 0.8; }
.storage-thumbs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 6px;
}
.storage-thumb {
  width: 44px;
  height: 44px;
  padding: 0;
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  overflow: hidden;
  background: transparent;
  cursor: pointer;
}
.storage-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.repay-link {
  display: inline-block;
  margin-top: 6px;
  font-size: 0.8rem;
  color: var(--primary);
  text-decoration: none;
  font-weight: 600;
}
.stmt-box {
  margin-top: 8px;
  padding: 10px 12px;
  border-radius: 12px;
  background: var(--primary-soft, rgba(27, 127, 90, 0.1));
  border: 1px solid var(--surface-border);
}
.stmt-head {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  font-size: 0.78rem;
  font-weight: 700;
  margin-bottom: 8px;
}
.stmt-range { font-weight: 500; color: var(--muted); font-size: 0.72rem; white-space: nowrap; }
.stmt-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 12px;
}
.stmt-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.stmt-cell span {
  color: var(--muted);
  font-size: 0.72rem;
  white-space: nowrap;
}
.stmt-cell strong {
  font-family: var(--font-display);
  font-variant-numeric: tabular-nums;
  font-size: clamp(0.82rem, 3.6vw, 0.95rem);
  white-space: nowrap;
  letter-spacing: -0.02em;
  line-height: 1.25;
}
.stmt-grid .warn { color: #c45c3e; }
.stmt-sub {
  margin-top: 8px;
  font-size: 0.72rem;
  color: var(--muted);
  line-height: 1.4;
}
.stmt-err {
  margin-top: 6px;
  font-size: 0.75rem;
  color: #c45c3e;
}

.lightbox {
  position: fixed;
  inset: 0;
  z-index: 3000;
  background: rgba(0, 0, 0, 0.88);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px 16px 24px;
}
.lb-close {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 40px;
  height: 40px;
  border: none;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}
.lb-img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: 8px;
}
</style>
