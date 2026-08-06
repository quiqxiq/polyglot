import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '../stores/auth'

import DashboardView from '../views/DashboardView.vue'
import LoginView from '../views/LoginView.vue'
import SessionsView from '../views/SessionsView.vue'
import LLMConfigView from '../views/LLMConfigView.vue'
import KnowledgeView from '../views/KnowledgeView.vue'
import ConversationsView from '../views/ConversationsView.vue'
import TechniciansView from '../views/TechniciansView.vue'
import ActiveSessionsView from '../views/ActiveSessionsView.vue'
import VouchersView from '../views/VouchersView.vue'
import RbacManagementView from '../views/RbacManagementView.vue'
import DeviceManagementView from '../views/DeviceManagementView.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: LoginView,
    meta: { public: true },
  },
  {
    path: '/',
    name: 'Dashboard',
    component: DashboardView,
    meta: { requiresAuth: true },
  },
  {
    path: '/devices',
    name: 'DeviceManagement',
    component: DeviceManagementView,
    meta: { requiresAuth: true },
  },
  {
    path: '/active-sessions',
    name: 'ActiveSessions',
    component: ActiveSessionsView,
    meta: { requiresAuth: true },
  },
  {
    path: '/vouchers',
    name: 'Vouchers',
    component: VouchersView,
    meta: { requiresAuth: true },
  },
  {
    path: '/sessions',
    name: 'Sessions',
    component: SessionsView,
    meta: { requiresAuth: true },
  },
  {
    path: '/llm-config',
    name: 'LLMConfig',
    component: LLMConfigView,
    meta: { requiresAuth: true },
  },
  {
    path: '/knowledge',
    name: 'Knowledge',
    component: KnowledgeView,
    meta: { requiresAuth: true },
  },
  {
    path: '/technicians',
    name: 'Technicians',
    component: TechniciansView,
    meta: { requiresAuth: true },
  },
  {
    path: '/rbac-users',
    name: 'RbacManagement',
    component: RbacManagementView,
    meta: { requiresAuth: true },
  },
  {
    path: '/conversations',
    name: 'Conversations',
    component: ConversationsView,
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (to.path === '/login' && authStore.isAuthenticated) {
    next('/')
  } else {
    next()
  }
})

export default router
