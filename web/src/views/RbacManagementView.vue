<template>
  <div class="rbac-page animate-fade-in">
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1 class="page-title gradient-text">User & RBAC Access Control</h1>
        <p class="page-desc">Manajemen hak akses role-based (Casbin v3) dan penugasan role pengguna</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-secondary" :disabled="rbacStore.loading" @click="refreshData">
          <RefreshCw :class="['w-4 h-4 mr-2', { 'animate-spin': rbacStore.loading }]" />
          Refresh
        </button>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="rbacStore.error" class="alert alert-error mb-6">
      <AlertCircle class="w-5 h-5 flex-shrink-0" />
      <span>{{ rbacStore.error }}</span>
    </div>

    <!-- Tabs Navigation -->
    <div class="tab-header mb-6">
      <button
        :class="['tab-btn', { active: activeTab === 'policies' }]"
        @click="activeTab = 'policies'"
      >
        <ShieldCheck class="w-4 h-4 inline mr-2" />
        Matriks Permisi Role (Policies)
      </button>
      <button
        :class="['tab-btn', { active: activeTab === 'assignments' }]"
        @click="activeTab = 'assignments'"
      >
        <Users class="w-4 h-4 inline mr-2" />
        Penugasan Role Pengguna (User Roles)
      </button>
    </div>

    <!-- Tab 1: Policies Matrix -->
    <div v-if="activeTab === 'policies'" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Add Policy Form -->
      <div class="glass-panel p-6 lg:col-span-1">
        <h2 class="text-lg font-bold mb-4 text-main flex items-center gap-2">
          <PlusCircle class="w-5 h-5 text-indigo-500" />
          Tambah Aturan Policy
        </h2>

        <form @submit.prevent="handleAddPolicy">
          <div class="form-group mb-4">
            <label class="form-label">Role Subjek</label>
            <input
              v-model="policyForm.role"
              type="text"
              placeholder="misal: admin, teknisi, finance"
              class="form-control"
              required
            />
          </div>

          <div class="form-group mb-4">
            <label class="form-label">Path API / Resource</label>
            <input
              v-model="policyForm.path"
              type="text"
              placeholder="misal: /api/v1/devices"
              class="form-control"
              required
            />
          </div>

          <div class="form-group mb-6">
            <label class="form-label">Method HTTP</label>
            <select v-model="policyForm.method" class="form-control" required>
              <option value="*">SEMUA ( * )</option>
              <option value="GET">GET (Read)</option>
              <option value="POST">POST (Create)</option>
              <option value="PUT">PUT (Update)</option>
              <option value="DELETE">DELETE (Remove)</option>
            </select>
          </div>

          <button type="submit" class="btn btn-primary w-full justify-center" :disabled="submitting">
            <PlusCircle class="w-4 h-4 mr-2" />
            Simpan Policy
          </button>
        </form>
      </div>

      <!-- Policies Table -->
      <div class="glass-panel p-6 lg:col-span-2">
        <h2 class="text-lg font-bold mb-4 text-main flex items-center gap-2">
          <ShieldCheck class="w-5 h-5 text-emerald-500" />
          Daftar Permission Rules ({{ rbacStore.policies.length }})
        </h2>

        <div class="table-responsive">
          <table class="data-table">
            <thead>
              <tr>
                <th>Role</th>
                <th>Resource Path</th>
                <th>HTTP Method</th>
                <th>Tindakan</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="rbacStore.policies.length === 0">
                <td colspan="4" class="text-center py-8 text-muted">
                  Belum ada aturan policy yang terdaftar.
                </td>
              </tr>
              <tr v-for="(p, idx) in rbacStore.policies" :key="idx">
                <td>
                  <span :class="['badge', getRoleBadgeClass(p.role)]">{{ p.role }}</span>
                </td>
                <td class="font-mono text-cyan font-bold">{{ p.path }}</td>
                <td>
                  <span :class="['badge', getMethodBadgeClass(p.method)]">{{ p.method }}</span>
                </td>
                <td>
                  <button class="btn btn-sm btn-danger-outline" @click="handleRemovePolicy(p)">
                    <Trash2 class="w-3.5 h-3.5 mr-1" />
                    Hapus
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Tab 2: User Role Assignments -->
    <div v-if="activeTab === 'assignments'" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Assign Role Form -->
      <div class="glass-panel p-6 lg:col-span-1">
        <h2 class="text-lg font-bold mb-4 text-main flex items-center gap-2">
          <UserPlus class="w-5 h-5 text-indigo-500" />
          Tugaskan Role ke User
        </h2>

        <form @submit.prevent="handleAssignRole">
          <div class="form-group mb-4">
            <label class="form-label">User / Email / Subject</label>
            <input
              v-model="assignForm.user"
              type="text"
              placeholder="misal: admin@gnet.id atau 1"
              class="form-control"
              required
            />
          </div>

          <div class="form-group mb-6">
            <label class="form-label">Role Yang Diberikan</label>
            <input
              v-model="assignForm.role"
              type="text"
              placeholder="misal: admin, agent, teknisi"
              class="form-control"
              required
            />
          </div>

          <button type="submit" class="btn btn-primary w-full justify-center" :disabled="submitting">
            <UserPlus class="w-4 h-4 mr-2" />
            Tugaskan Role
          </button>
        </form>
      </div>

      <!-- Assignments Table -->
      <div class="glass-panel p-6 lg:col-span-2">
        <h2 class="text-lg font-bold mb-4 text-main flex items-center gap-2">
          <Users class="w-5 h-5 text-emerald-500" />
          Penugasan Role Aktif ({{ rbacStore.roleAssignments.length }})
        </h2>

        <div class="table-responsive">
          <table class="data-table">
            <thead>
              <tr>
                <th>Pengguna (User)</th>
                <th>Role</th>
                <th>Tindakan</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="rbacStore.roleAssignments.length === 0">
                <td colspan="3" class="text-center py-8 text-muted">
                  Belum ada penugasan role pengguna.
                </td>
              </tr>
              <tr v-for="(ra, idx) in rbacStore.roleAssignments" :key="idx">
                <td class="font-bold text-main">
                  <span class="user-avatar mr-2">{{ ra.user.charAt(0).toUpperCase() }}</span>
                  {{ ra.user }}
                </td>
                <td>
                  <span :class="['badge', getRoleBadgeClass(ra.role)]">{{ ra.role }}</span>
                </td>
                <td>
                  <button class="btn btn-sm btn-danger-outline" @click="handleUnassignRole(ra)">
                    <UserX class="w-3.5 h-3.5 mr-1" />
                    Pencabutan
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  ShieldCheck,
  Users,
  PlusCircle,
  UserPlus,
  RefreshCw,
  Trash2,
  UserX,
  AlertCircle,
} from 'lucide-vue-next'
import { useRBACStore } from '../stores/rbac'
import type { RBACPolicy, RBACRoleAssignment } from '../types'

const rbacStore = useRBACStore()
const activeTab = ref<'policies' | 'assignments'>('policies')
const submitting = ref(false)

const policyForm = ref({
  role: '',
  path: '',
  method: 'GET',
})

const assignForm = ref({
  user: '',
  role: '',
})

function refreshData() {
  rbacStore.fetchAll()
}

async function handleAddPolicy() {
  if (!policyForm.value.role || !policyForm.value.path) return
  try {
    submitting.value = true
    await rbacStore.addPolicy(
      policyForm.value.role,
      policyForm.value.path,
      policyForm.value.method
    )
    policyForm.value = { role: '', path: '', method: 'GET' }
    alert('Aturan policy berhasil ditambahkan!')
  } catch (e: any) {
    alert(e.message || 'Gagal menambahkan policy')
  } finally {
    submitting.value = false
  }
}

async function handleRemovePolicy(p: RBACPolicy) {
  if (!confirm(`Hapus policy untuk role '${p.role}' pada ${p.method} ${p.path}?`)) return
  try {
    await rbacStore.removePolicy(p.role, p.path, p.method)
  } catch (e: any) {
    alert(e.message || 'Gagal menghapus policy')
  }
}

async function handleAssignRole() {
  if (!assignForm.value.user || !assignForm.value.role) return
  try {
    submitting.value = true
    await rbacStore.assignRole(assignForm.value.user, assignForm.value.role)
    assignForm.value = { user: '', role: '' }
    alert('Role berhasil ditugaskan!')
  } catch (e: any) {
    alert(e.message || 'Gagal menugaskan role')
  } finally {
    submitting.value = false
  }
}

async function handleUnassignRole(ra: RBACRoleAssignment) {
  if (!confirm(`Cabut role '${ra.role}' dari user '${ra.user}'?`)) return
  try {
    await rbacStore.unassignRole(ra.user, ra.role)
  } catch (e: any) {
    alert(e.message || 'Gagal mencabut role')
  }
}

function getRoleBadgeClass(role: string): string {
  switch (role.toLowerCase()) {
    case 'admin':
    case 'superadmin':
      return 'badge-danger'
    case 'teknisi':
    case 'agent':
      return 'badge-primary'
    case 'finance':
      return 'badge-success'
    default:
      return 'badge-info'
  }
}

function getMethodBadgeClass(method: string): string {
  switch (method.toUpperCase()) {
    case 'GET':
      return 'badge-success'
    case 'POST':
      return 'badge-primary'
    case 'PUT':
      return 'badge-warning'
    case 'DELETE':
      return 'badge-danger'
    default:
      return 'badge-info'
  }
}

onMounted(() => {
  rbacStore.fetchAll()
})
</script>

<style scoped>
.rbac-page {
  padding-top: 16px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-title {
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.5px;
}

.page-desc {
  font-size: 14px;
  color: var(--text-muted);
  margin-top: 4px;
}

.tab-header {
  display: flex;
  gap: 8px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 12px;
}

.tab-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  padding: 8px 16px;
  font-size: 14px;
  font-weight: 600;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.05);
}

.tab-btn.active {
  color: #ffffff;
  background: var(--primary);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
}

.user-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--primary);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
}

.btn-danger-outline {
  background: transparent;
  border: 1px solid var(--color-danger);
  color: var(--color-danger);
  padding: 4px 10px;
  font-size: 12px;
}

.btn-danger-outline:hover {
  background: var(--color-danger);
  color: #fff;
}
</style>
