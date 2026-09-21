import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('./views/Login.vue'), meta: { public: true } },
    { path: '/register', name: 'register', component: () => import('./views/Register.vue'), meta: { public: true } },
    { path: '/forgot-password', name: 'forgot', component: () => import('./views/ForgotPassword.vue'), meta: { public: true } },
    { path: '/', name: 'home', component: () => import('./views/Home.vue') },
    { path: '/transactions', name: 'transactions', component: () => import('./views/TransactionList.vue') },
    { path: '/transactions/new', name: 'transaction-edit', component: () => import('./views/TransactionEdit.vue') },
    { path: '/transactions/recognize', name: 'text-recognize', component: () => import('./views/TextRecognize.vue') },
    { path: '/transactions/:id', name: 'transaction-edit-id', component: () => import('./views/TransactionEdit.vue'), props: true },
    { path: '/stats', name: 'stats', component: () => import('./views/Stats.vue') },
    { path: '/mine', name: 'mine', component: () => import('./views/Mine.vue') },
    { path: '/accounts', name: 'accounts', component: () => import('./views/Accounts.vue') },
    { path: '/accounts/new', name: 'account-create', component: () => import('./views/AccountEdit.vue'), meta: { hideNav: true } },
    { path: '/accounts/:id/edit', name: 'account-edit', component: () => import('./views/AccountEdit.vue'), meta: { hideNav: true } },
    { path: '/categories', name: 'categories', component: () => import('./views/Categories.vue') },
    { path: '/categories/new', name: 'category-create', component: () => import('./views/CategoryEdit.vue'), meta: { hideNav: true } },
    { path: '/categories/:id/edit', name: 'category-edit', component: () => import('./views/CategoryEdit.vue'), meta: { hideNav: true } },
    { path: '/tags', name: 'tags', component: () => import('./views/Tags.vue') },
    { path: '/templates', name: 'templates', component: () => import('./views/Templates.vue') },
    { path: '/templates/new', name: 'template-create', component: () => import('./views/TemplateEdit.vue'), meta: { hideNav: true } },
    { path: '/templates/:id/edit', name: 'template-edit', component: () => import('./views/TemplateEdit.vue'), meta: { hideNav: true } },
    { path: '/fare-rules', name: 'fare-rules', component: () => import('./views/FareRules.vue') },
    { path: '/fare-rules/new', name: 'fare-rule-create', component: () => import('./views/FareRuleEdit.vue'), meta: { hideNav: true } },
    { path: '/fare-rules/:id/edit', name: 'fare-rule-edit', component: () => import('./views/FareRuleEdit.vue'), meta: { hideNav: true } },
    { path: '/schedules', name: 'schedules', component: () => import('./views/Schedules.vue') },
    { path: '/schedules/new', name: 'schedule-create', component: () => import('./views/ScheduleEdit.vue'), meta: { hideNav: true } },
    { path: '/schedules/:id/edit', name: 'schedule-edit', component: () => import('./views/ScheduleEdit.vue'), meta: { hideNav: true } },
    { path: '/settings', name: 'settings', component: () => import('./views/Settings.vue') },
    { path: '/admin/users', name: 'admin-users', component: () => import('./views/UsersAdmin.vue'), meta: { admin: true } },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.token) return { name: 'login', query: { redirect: to.fullPath } }
  if ((to.name === 'login' || to.name === 'register') && auth.token) return { name: 'home' }
  if (to.meta.admin) {
    if (!auth.profile) {
      try {
        await auth.fetchMe()
      } catch {
        return { name: 'login', query: { redirect: to.fullPath } }
      }
    }
    if (!auth.isAdmin()) return { name: 'mine' }
  }
  return true
})

export default router
