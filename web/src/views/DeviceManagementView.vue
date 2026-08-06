<template>
  <div class="devices-page animate-fade-in">
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1 class="page-title gradient-text">Manajemen Router & Device</h1>
        <p class="page-desc">Kelola inventaris router MikroTik, Cisco, OLT, dan pengujian koneksi realtime</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-primary" @click="openCreateModal">
          <PlusCircle class="w-4 h-4 mr-2" />
          Tambah Router Baru
        </button>
        <button class="btn btn-secondary" :disabled="deviceStore.loading" @click="refreshData">
          <RefreshCw :class="['w-4 h-4 mr-2', { 'animate-spin': deviceStore.loading }]" />
          Refresh
        </button>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="deviceStore.error" class="alert alert-error mb-6">
      <AlertCircle class="w-5 h-5 flex-shrink-0" />
      <span>{{ deviceStore.error }}</span>
    </div>

    <!-- Stat Cards Overview -->
    <div class="stats-grid mb-6">
      <div class="stat-card glass-panel">
        <div class="stat-icon bg-indigo-500/10 text-indigo-500">
          <RouterIcon class="w-6 h-6" />
        </div>
        <div class="stat-content">
          <span class="stat-label">Total Router</span>
          <span class="stat-value">{{ deviceStore.devices.length }}</span>
        </div>
      </div>

      <div class="stat-card glass-panel">
        <div class="stat-icon bg-emerald-500/10 text-emerald-500">
          <CheckCircle2 class="w-6 h-6" />
        </div>
        <div class="stat-content">
          <span class="stat-label">Aktif / Enabled</span>
          <span class="stat-value">{{ enabledCount }}</span>
        </div>
      </div>

      <div class="stat-card glass-panel">
        <div class="stat-icon bg-cyan-500/10 text-cyan-500">
          <Cpu class="w-6 h-6" />
        </div>
        <div class="stat-content">
          <span class="stat-label">Driver Vendor</span>
          <span class="stat-value">MikroTik / OLT</span>
        </div>
      </div>
    </div>

    <!-- Filter & Search Bar -->
    <div class="filter-bar mb-6">
      <div class="search-input-wrapper">
        <Search class="search-icon w-4 h-4" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari berdasarkan Nama, Host IP, ID, atau Vendor..."
          class="form-control search-input"
        />
      </div>
    </div>

    <!-- Devices Grid Table -->
    <div class="glass-panel overflow-hidden">
      <div class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID & Nama Router</th>
              <th>Vendor / Driver</th>
              <th>Host Address</th>
              <th>Port API</th>
              <th>Status</th>
              <th>Hasil Test Connection</th>
              <th>Tindakan</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredDevices.length === 0">
              <td colspan="7" class="text-center py-8 text-muted">
                Belum ada perangkat router terdaftar. Klik "Tambah Router Baru" untuk mendaftarkan router.
              </td>
            </tr>
            <tr v-for="dev in filteredDevices" :key="dev.id">
              <td>
                <div class="font-bold text-main flex items-center gap-2">
                  <RouterIcon class="w-4 h-4 text-indigo-400" />
                  {{ dev.name }}
                </div>
                <div class="font-mono text-xs text-muted mt-0.5">ID: {{ dev.id }}</div>
              </td>
              <td>
                <span :class="['badge', getVendorBadgeClass(dev.vendor)]">
                  {{ (dev.vendor || 'mikrotik').toUpperCase() }} ({{ dev.driver_type || 'mikrotik' }})
                </span>
              </td>
              <td class="font-mono text-cyan font-bold">{{ dev.host }}</td>
              <td class="font-mono text-xs">{{ dev.port }}</td>
              <td>
                <div v-if="!dev.enabled" class="badge badge-offline">
                  Disabled
                </div>
                <div v-else-if="deviceStore.testResults[dev.id]" class="flex items-center gap-2">
                  <span v-if="deviceStore.testResults[dev.id].success" class="badge badge-online">
                    <span class="pulse-dot pulse-dot-online mr-1"></span>
                    LIVE / ONLINE
                  </span>
                  <span v-else class="badge badge-danger">
                    <XCircle class="w-3.5 h-3.5 inline mr-1" />
                    OFFLINE
                  </span>
                </div>
                <div v-else class="flex items-center gap-1.5 text-xs text-muted">
                  <RefreshCw class="w-3.5 h-3.5 animate-spin text-indigo-400" />
                  Checking Live...
                </div>
              </td>
              <td>
                <div v-if="deviceStore.testResults[dev.id]" class="test-result-box flex flex-wrap gap-1">
                  <span
                    :class="[
                      'badge',
                      deviceStore.testResults[dev.id].success ? 'badge-success' : 'badge-danger',
                    ]"
                  >
                    <CheckCircle2 v-if="deviceStore.testResults[dev.id].success" class="w-3.5 h-3.5 inline mr-1" />
                    <XCircle v-else class="w-3.5 h-3.5 inline mr-1" />
                    {{ deviceStore.testResults[dev.id].message }}
                  </span>
                  <span v-if="deviceStore.testResults[dev.id].latency_ms !== undefined" class="badge badge-info text-xs font-mono">
                    ⚡ {{ deviceStore.testResults[dev.id].latency_ms }}ms
                  </span>
                  <span v-if="deviceStore.testResults[dev.id].identity" class="badge badge-primary text-xs font-mono">
                    🏷️ {{ deviceStore.testResults[dev.id].identity }}
                  </span>
                  <span v-if="deviceStore.testResults[dev.id].version" class="badge badge-warning text-xs font-mono">
                    📦 ROS {{ deviceStore.testResults[dev.id].version }}
                  </span>
                </div>
                <span v-else-if="deviceStore.testingId === dev.id" class="text-xs text-muted animate-pulse">
                  Connecting API...
                </span>
                <span v-else class="text-xs text-muted">Auto-testing...</span>
              </td>
              <td>
                <div class="flex gap-2">
                  <button
                    class="btn btn-sm btn-secondary"
                    :disabled="deviceStore.testingId === dev.id"
                    title="Test Ping/Koneksi API"
                    @click="handleTestConnection(dev.id)"
                  >
                    <RefreshCw v-if="deviceStore.testingId === dev.id" class="w-3.5 h-3.5 mr-1 animate-spin" />
                    <Zap v-else class="w-3.5 h-3.5 mr-1 text-amber-400" />
                    Test Ping
                  </button>
                  <button class="btn btn-sm btn-secondary" title="Edit Router" @click="openEditModal(dev)">
                    <Edit3 class="w-3.5 h-3.5" />
                  </button>
                  <button class="btn btn-sm btn-danger-outline" title="Hapus Router" @click="confirmDelete(dev)">
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create/Edit Device Modal -->
    <div v-if="showModal" class="modal-overlay">
      <div class="modal-content glass-panel max-w-lg w-full p-6">
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-lg font-bold text-main">
            {{ isEditing ? 'Edit Konfigurasi Router' : 'Tambah Router Baru' }}
          </h3>
          <button class="btn btn-sm btn-secondary" @click="showModal = false">
            <X class="w-4 h-4" />
          </button>
        </div>

        <form @submit.prevent="handleSubmit">
          <div class="grid grid-cols-2 gap-4 mb-4">
            <div class="form-group">
              <label class="form-label">ID Router (Slug)</label>
              <input
                v-model="form.id"
                type="text"
                placeholder="misal: mtk-core-01"
                class="form-control"
                :disabled="isEditing"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label">Nama Device</label>
              <input
                v-model="form.name"
                type="text"
                placeholder="misal: Router Core Utama"
                class="form-control"
                required
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4 mb-4">
            <div class="form-group">
              <label class="form-label">Vendor</label>
              <select v-model="form.vendor" class="form-control" required>
                <option value="mikrotik">MikroTik RouterOS</option>
                <option value="cisco">Cisco IOS</option>
                <option value="genieacs">GenieACS TR-069</option>
                <option value="zteolt">ZTE OLT</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">Driver Type</label>
              <input v-model="form.driver_type" type="text" class="form-control" required />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4 mb-4">
            <div class="form-group">
              <label class="form-label">Host / IP Address</label>
              <input
                v-model="form.host"
                type="text"
                placeholder="misal: 192.168.1.1"
                class="form-control"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label">Port API / SSH</label>
              <input v-model.number="form.port" type="number" class="form-control" required />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4 mb-4">
            <div class="form-group">
              <label class="form-label">Username API</label>
              <input v-model="form.username" type="text" placeholder="admin" class="form-control" required />
            </div>
            <div class="form-group">
              <label class="form-label">Password API</label>
              <input
                v-model="form.password"
                type="password"
                :placeholder="isEditing ? '(Tetap simpan lama)' : 'Password'"
                class="form-control"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4 mb-4">
            <div class="form-group">
              <label class="form-label">Timeout (ms)</label>
              <input v-model.number="form.timeout_ms" type="number" class="form-control" required />
            </div>
            <div class="form-group">
              <label class="form-label">Poll Interval (ms)</label>
              <input v-model.number="form.poll_interval_ms" type="number" placeholder="30000" class="form-control" />
            </div>
          </div>

          <div class="form-group mb-6">
            <label class="form-label">Tags (Pisahkan koma)</label>
            <input v-model="tagsInput" type="text" placeholder="core, hotspot, pppoe" class="form-control" />
          </div>

          <div class="form-group mb-6">
            <label class="flex items-center gap-2 cursor-pointer">
              <input v-model="form.enabled" type="checkbox" class="w-4 h-4 accent-indigo-500" />
              <span class="text-main font-bold">Aktifkan Router Ini (Enabled)</span>
            </label>
          </div>

          <div class="flex justify-end gap-2">
            <button type="button" class="btn btn-secondary" @click="showModal = false">Batal</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              <RefreshCw v-if="submitting" class="w-4 h-4 mr-2 animate-spin" />
              <Save v-else class="w-4 h-4 mr-2" />
              Simpan Router
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Confirm Modal -->
    <ConfirmModal
      :show="showDeleteModal"
      title="Hapus Router"
      :message="`Apakah Anda yakin ingin menghapus router '${targetDevice?.name}' (${targetDevice?.id})?`"
      confirmText="Ya, Hapus Router"
      variant="danger"
      @confirm="executeDelete"
      @cancel="showDeleteModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  Router as RouterIcon,
  PlusCircle,
  RefreshCw,
  Search,
  CheckCircle2,
  XCircle,
  Zap,
  Edit3,
  Trash2,
  AlertCircle,
  Cpu,
  Save,
  X,
} from 'lucide-vue-next'
import { useDeviceStore } from '../stores/devices'
import type { Device, DevicePayload } from '../types'
import ConfirmModal from '../components/ConfirmModal.vue'

const deviceStore = useDeviceStore()
const searchQuery = ref('')
const showModal = ref(false)
const isEditing = ref(false)
const submitting = ref(false)

const showDeleteModal = ref(false)
const targetDevice = ref<Device | null>(null)

const tagsInput = ref('')

const form = ref<DevicePayload>({
  id: '',
  name: '',
  vendor: 'mikrotik',
  driver_type: 'mikrotik',
  host: '',
  port: 8728,
  timeout_ms: 10000,
  poll_interval_ms: 30000,
  enabled: true,
  username: 'admin',
  password: '',
})

const enabledCount = computed(() => {
  return deviceStore.devices.filter((d) => d.enabled).length
})

const filteredDevices = computed(() => {
  if (!searchQuery.value) return deviceStore.devices
  const q = searchQuery.value.toLowerCase()
  return deviceStore.devices.filter(
    (d) =>
      d.name.toLowerCase().includes(q) ||
      d.id.toLowerCase().includes(q) ||
      d.host.toLowerCase().includes(q) ||
      d.vendor.toLowerCase().includes(q)
  )
})

function refreshData() {
  deviceStore.fetchDevices()
}

function openCreateModal() {
  isEditing.value = false
  form.value = {
    id: '',
    name: '',
    vendor: 'mikrotik',
    driver_type: 'mikrotik',
    host: '',
    port: 8728,
    timeout_ms: 10000,
    poll_interval_ms: 30000,
    enabled: true,
    username: 'admin',
    password: '',
  }
  tagsInput.value = ''
  showModal.value = true
}

function openEditModal(dev: Device) {
  isEditing.value = true
  form.value = {
    id: dev.id,
    name: dev.name,
    vendor: dev.vendor,
    driver_type: dev.driver_type,
    host: dev.host,
    port: dev.port,
    timeout_ms: dev.timeout_ms,
    poll_interval_ms: dev.poll_interval_ms || 30000,
    enabled: dev.enabled,
    username: 'admin',
    password: '',
  }
  tagsInput.value = (dev.tags || []).join(', ')
  showModal.value = true
}

async function handleSubmit() {
  try {
    submitting.value = true
    const tags = tagsInput.value
      .split(',')
      .map((t) => t.trim())
      .filter((t) => t.length > 0)
    form.value.tags = tags

    if (isEditing.value) {
      await deviceStore.updateDevice(form.value.id, form.value)
      alert('Router berhasil diperbarui!')
    } else {
      await deviceStore.createDevice(form.value)
      alert('Router baru berhasil mendaftar!')
    }
    showModal.value = false
  } catch (e: any) {
    alert('Gagal menyimpan router: ' + (e.message || e))
  } finally {
    submitting.value = false
  }
}

async function handleTestConnection(id: string) {
  try {
    await deviceStore.testConnection(id)
  } catch (e: any) {
    // Error handled in store result badge
  }
}

function confirmDelete(dev: Device) {
  targetDevice.value = dev
  showDeleteModal.value = true
}

async function executeDelete() {
  if (!targetDevice.value) return
  showDeleteModal.value = false
  try {
    await deviceStore.deleteDevice(targetDevice.value.id)
    alert('Router berhasil dihapus')
  } catch (e: any) {
    alert('Gagal menghapus router: ' + (e.message || e))
  }
}

function getVendorBadgeClass(vendor?: string): string {
  if (!vendor) return 'badge-primary'
  switch (vendor.toLowerCase()) {
    case 'mikrotik':
      return 'badge-primary'
    case 'cisco':
      return 'badge-info'
    case 'genieacs':
      return 'badge-success'
    default:
      return 'badge-warning'
  }
}

let autoTestInterval: any = null

onMounted(async () => {
  await deviceStore.fetchDevices()
  deviceStore.testAllDevices()
  autoTestInterval = setInterval(() => {
    deviceStore.testAllDevices()
  }, 15000)
})

onUnmounted(() => {
  if (autoTestInterval) {
    clearInterval(autoTestInterval)
  }
})
</script>

<style scoped>
.devices-page {
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

.header-actions {
  display: flex;
  gap: 12px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
}

.stat-card {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  border-radius: var(--radius-lg);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 500;
}

.stat-value {
  font-size: 22px;
  font-weight: 800;
  color: var(--text-main);
}

.filter-bar {
  display: flex;
  gap: 12px;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  max-width: 450px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
}

.search-input {
  padding-left: 36px;
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

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}
</style>
