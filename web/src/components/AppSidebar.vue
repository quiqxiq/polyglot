<template>
  <aside :class="['sidebar', { 'sidebar-open': mobileOpen }]">
    <!-- Brand Logo -->
    <div class="sidebar-brand">
      <div class="brand-icon">
        <Bot class="w-6 h-6 text-indigo-500" />
      </div>
      <div class="brand-text">
        <span class="brand-name gradient-text-accent">GNET Bot</span>
        <span class="brand-sub">WhatsApp Assistant</span>
      </div>
    </div>

    <!-- Navigation Items -->
    <nav class="sidebar-nav">
      <router-link
        v-for="item in navItems"
        :key="item.path"
        :to="item.path"
        class="nav-link"
        active-class="nav-link-active"
        @click="mobileOpen = false"
      >
        <component :is="item.icon" class="nav-icon" />
        <span class="nav-label">{{ item.label }}</span>
        <span v-if="item.badge" class="nav-badge">{{ item.badge }}</span>
      </router-link>
    </nav>

    <!-- Footer Profile & Role -->
    <div class="sidebar-footer">
      <div class="user-card">
        <div class="avatar">
          {{ userEmail.charAt(0).toUpperCase() }}
        </div>
        <div class="user-info">
          <div class="user-email" :title="userEmail">{{ userEmail }}</div>
          <span class="badge badge-primary user-role">{{ userRole }}</span>
        </div>
        <button class="logout-btn" title="Keluar" @click="handleLogout">
          <LogOut class="w-4 h-4" />
        </button>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  LayoutDashboard,
  Smartphone,
  Cpu,
  BookOpen,
  MessageSquare,
  Wrench,
  LogOut,
  Bot,
  Activity,
  Ticket,
  ShieldCheck,
} from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'

defineProps<{ mobileOpen?: boolean }>()
defineEmits(['update:mobileOpen'])

const router = useRouter()
const authStore = useAuthStore()

const userEmail = computed(() => authStore.userEmail || 'Admin GNET')
const userRole = computed(() => (authStore.isAdmin ? 'Admin' : 'Agent'))

interface NavItem {
  path: string
  label: string
  icon: any
  badge?: string
}

const navItems: NavItem[] = [
  { path: '/', label: 'Overview', icon: LayoutDashboard },
  { path: '/active-sessions', label: 'Sesi Aktif', icon: Activity },
  { path: '/vouchers', label: 'Voucher & Hotspot', icon: Ticket },
  { path: '/sessions', label: 'Koneksi WA', icon: Smartphone },
  { path: '/llm-config', label: 'Konfigurasi LLM', icon: Cpu },
  { path: '/knowledge', label: 'Basis Pengetahuan', icon: BookOpen },
  { path: '/technicians', label: 'Tim Teknisi', icon: Wrench },
  { path: '/rbac-users', label: 'User & Hak Akses', icon: ShieldCheck },
  { path: '/conversations', label: 'Percakapan Live', icon: MessageSquare },
]

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.sidebar {
  width: 260px;
  height: 100vh;
  position: fixed;
  top: 0;
  left: 0;
  z-index: 40;
  background: var(--bg-card-solid);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  padding: 20px 16px;
  transition: transform 0.3s ease, background-color 0.3s ease;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px 24px;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 20px;
}

.brand-icon {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  background: rgba(99, 102, 241, 0.12);
  border: 1px solid rgba(99, 102, 241, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-text {
  display: flex;
  flex-direction: column;
}

.brand-name {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.5px;
}

.brand-sub {
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 500;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.2s ease;
}

.nav-link:hover {
  background: rgba(99, 102, 241, 0.1);
  color: var(--text-main);
  transform: translateX(2px);
}

.nav-link-active {
  background: linear-gradient(90deg, rgba(99, 102, 241, 0.15) 0%, rgba(99, 102, 241, 0.05) 100%);
  color: var(--primary);
  border-left: 3px solid var(--primary);
  font-weight: 700;
}

.nav-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.sidebar-footer {
  margin-top: auto;
  padding-top: 16px;
}

.user-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-main);
}

.avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary) 0%, var(--accent-cyan) 100%);
  color: #ffffff;
  font-weight: 700;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.3);
  flex-shrink: 0;
}

.user-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

.user-email {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-role {
  font-size: 10px;
  padding: 2px 6px;
  align-self: flex-start;
  margin-top: 2px;
}

.logout-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 6px;
  border-radius: 6px;
  transition: color 0.2s, background 0.2s;
  flex-shrink: 0;
}

.logout-btn:hover {
  color: var(--color-danger);
  background: var(--color-danger-bg);
}

@media (max-width: 768px) {
  .sidebar {
    transform: translateX(-100%);
  }
  .sidebar-open {
    transform: translateX(0);
    box-shadow: 10px 0 30px rgba(0, 0, 0, 0.7);
  }
}
</style>
