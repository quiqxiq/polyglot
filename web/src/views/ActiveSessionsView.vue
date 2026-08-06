<template>
  <div class="active-sessions-page animate-fade-in">
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1 class="page-title gradient-text">Active Sessions & Leases</h1>
        <p class="page-desc">Monitoring realtime koneksi PPPoE, Hotspot Active, dan DHCP Leases RouterOS</p>
      </div>

      <div class="header-actions">
        <select v-model="networkStore.selectedDeviceId" class="device-select" @change="handleDeviceChange">
          <option value="mtk-test">MikroTik Test Router (mtk-test)</option>
        </select>
        <button class="btn btn-secondary" :disabled="networkStore.loading" @click="refreshData">
          <RefreshCw :class="['w-4 h-4 mr-2', { 'animate-spin': networkStore.loading }]" />
          Refresh
        </button>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="networkStore.error" class="alert alert-error mb-6">
      <AlertCircle class="w-5 h-5 flex-shrink-0" />
      <span>{{ networkStore.error }}</span>
    </div>

    <!-- Stat Cards Overview -->
    <div class="stats-grid mb-6">
      <div class="stat-card glass-panel">
        <div class="stat-icon bg-indigo-500/10 text-indigo-500">
          <Network class="w-6 h-6" />
        </div>
        <div class="stat-content">
          <span class="stat-label">PPPoE Active</span>
          <span class="stat-value">{{ networkStore.pppoeActive.length }}</span>
        </div>
      </div>

      <div class="stat-card glass-panel">
        <div class="stat-icon bg-emerald-500/10 text-emerald-500">
          <Wifi class="w-6 h-6" />
        </div>
        <div class="stat-content">
          <span class="stat-label">Hotspot Active</span>
          <span class="stat-value">{{ networkStore.hotspotActive.length }}</span>
        </div>
      </div>

      <div class="stat-card glass-panel">
        <div class="stat-icon bg-cyan-500/10 text-cyan-500">
          <HardDrive class="w-6 h-6" />
        </div>
        <div class="stat-content">
          <span class="stat-label">DHCP Leases</span>
          <span class="stat-value">{{ networkStore.dhcpLeases.length }}</span>
        </div>
      </div>
    </div>

    <!-- Tabs Navigation -->
    <div class="tab-header mb-6">
      <button
        :class="['tab-btn', { active: activeTab === 'pppoe' }]"
        @click="activeTab = 'pppoe'"
      >
        <Network class="w-4 h-4 inline mr-2" />
        PPPoE Active ({{ networkStore.pppoeActive.length }})
      </button>
      <button
        :class="['tab-btn', { active: activeTab === 'hotspot' }]"
        @click="activeTab = 'hotspot'"
      >
        <Wifi class="w-4 h-4 inline mr-2" />
        Hotspot Active ({{ networkStore.hotspotActive.length }})
      </button>
      <button
        :class="['tab-btn', { active: activeTab === 'dhcp' }]"
        @click="activeTab = 'dhcp'"
      >
        <HardDrive class="w-4 h-4 inline mr-2" />
        DHCP Leases ({{ networkStore.dhcpLeases.length }})
      </button>
    </div>

    <!-- Search Input -->
    <div class="filter-bar mb-4">
      <div class="search-input-wrapper">
        <Search class="search-icon w-4 h-4" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari berdasarkan User, IP, MAC Address..."
          class="form-control search-input"
        />
      </div>
    </div>

    <!-- Tab 1: PPPoE Active Table -->
    <div v-if="activeTab === 'pppoe'" class="glass-panel overflow-hidden">
      <div class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th>User Name</th>
              <th>Service</th>
              <th>Caller ID (MAC)</th>
              <th>IP Address</th>
              <th>Uptime</th>
              <th>Tindakan</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredPPPoE.length === 0">
              <td colspan="6" class="text-center py-8 text-muted">
                Tidak ada data sesi PPPoE aktif.
              </td>
            </tr>
            <tr v-for="item in filteredPPPoE" :key="item.id">
              <td class="font-bold text-main">
                <span class="user-avatar mr-2">{{ item.name.charAt(0).toUpperCase() }}</span>
                {{ item.name }}
              </td>
              <td><span class="badge badge-primary">{{ item.service || 'pppoe' }}</span></td>
              <td class="font-mono text-xs text-muted">{{ item.caller_id || '-' }}</td>
              <td class="font-mono text-cyan">{{ item.address }}</td>
              <td>
                <span class="uptime-badge">
                  <Clock class="w-3 h-3 inline mr-1" />
                  {{ item.uptime }}
                </span>
              </td>
              <td>
                <button
                  class="btn btn-sm btn-danger-outline"
                  title="Kick Session"
                  @click="confirmKick('pppoe', item.id, item.name)"
                >
                  <UserX class="w-3.5 h-3.5 mr-1" />
                  Disconnect
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Tab 2: Hotspot Active Table -->
    <div v-if="activeTab === 'hotspot'" class="glass-panel overflow-hidden">
      <div class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th>Hotspot User</th>
              <th>Server</th>
              <th>IP Address</th>
              <th>MAC Address</th>
              <th>Uptime</th>
              <th>Tindakan</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredHotspot.length === 0">
              <td colspan="6" class="text-center py-8 text-muted">
                Tidak ada data sesi Hotspot aktif.
              </td>
            </tr>
            <tr v-for="item in filteredHotspot" :key="item.id">
              <td class="font-bold text-main">
                <span class="user-avatar bg-emerald mr-2">{{ item.user.charAt(0).toUpperCase() }}</span>
                {{ item.user }}
              </td>
              <td><span class="badge badge-success">{{ item.server || 'hotspot1' }}</span></td>
              <td class="font-mono text-cyan">{{ item.address }}</td>
              <td class="font-mono text-xs text-muted">{{ item.mac_address }}</td>
              <td>
                <span class="uptime-badge">
                  <Clock class="w-3 h-3 inline mr-1" />
                  {{ item.uptime }}
                </span>
              </td>
              <td>
                <button
                  class="btn btn-sm btn-danger-outline"
                  title="Kick Hotspot Session"
                  @click="confirmKick('hotspot', item.id, item.user)"
                >
                  <UserX class="w-3.5 h-3.5 mr-1" />
                  Disconnect
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Tab 3: DHCP Leases Table -->
    <div v-if="activeTab === 'dhcp'" class="glass-panel overflow-hidden">
      <div class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th>IP Address</th>
              <th>MAC Address</th>
              <th>Host Name</th>
              <th>Server</th>
              <th>Status</th>
              <th>Tipe</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredDHCP.length === 0">
              <td colspan="6" class="text-center py-8 text-muted">
                Tidak ada data DHCP Leases.
              </td>
            </tr>
            <tr v-for="item in filteredDHCP" :key="item.id">
              <td class="font-mono text-cyan font-bold">{{ item.address }}</td>
              <td class="font-mono text-xs text-muted">{{ item.mac_address }}</td>
              <td class="text-main">{{ item.host_name || '-' }}</td>
              <td>{{ item.server || 'dhcp1' }}</td>
              <td>
                <span :class="['badge', item.status === 'bound' ? 'badge-success' : 'badge-warning']">
                  {{ item.status }}
                </span>
              </td>
              <td>
                <span :class="['badge', item.dynamic ? 'badge-primary' : 'badge-info']">
                  {{ item.dynamic ? 'Dynamic' : 'Static' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Confirmation Modal -->
    <ConfirmModal
      :show="showKickModal"
      :title="`Disconnect Sesi ${kickTarget.type.toUpperCase()}`"
      :message="`Apakah Anda yakin ingin memutuskan koneksi pengguna '${kickTarget.name}'?`"
      confirmText="Ya, Disconnect"
      variant="danger"
      @confirm="executeKick"
      @cancel="showKickModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Network,
  Wifi,
  HardDrive,
  RefreshCw,
  Search,
  Clock,
  UserX,
  AlertCircle,
} from 'lucide-vue-next'
import { useNetworkStore } from '../stores/network'
import ConfirmModal from '../components/ConfirmModal.vue'

const networkStore = useNetworkStore()
const activeTab = ref<'pppoe' | 'hotspot' | 'dhcp'>('pppoe')
const searchQuery = ref('')

const showKickModal = ref(false)
const kickTarget = ref<{ type: 'pppoe' | 'hotspot'; id: string; name: string }>({
  type: 'pppoe',
  id: '',
  name: '',
})

function handleDeviceChange() {
  refreshData()
}

function refreshData() {
  networkStore.fetchAll()
}

const filteredPPPoE = computed(() => {
  if (!searchQuery.value) return networkStore.pppoeActive
  const q = searchQuery.value.toLowerCase()
  return networkStore.pppoeActive.filter(
    (s) =>
      s.name.toLowerCase().includes(q) ||
      s.address.toLowerCase().includes(q) ||
      (s.caller_id && s.caller_id.toLowerCase().includes(q))
  )
})

const filteredHotspot = computed(() => {
  if (!searchQuery.value) return networkStore.hotspotActive
  const q = searchQuery.value.toLowerCase()
  return networkStore.hotspotActive.filter(
    (s) =>
      s.user.toLowerCase().includes(q) ||
      s.address.toLowerCase().includes(q) ||
      s.mac_address.toLowerCase().includes(q)
  )
})

const filteredDHCP = computed(() => {
  if (!searchQuery.value) return networkStore.dhcpLeases
  const q = searchQuery.value.toLowerCase()
  return networkStore.dhcpLeases.filter(
    (s) =>
      s.address.toLowerCase().includes(q) ||
      s.mac_address.toLowerCase().includes(q) ||
      (s.host_name && s.host_name.toLowerCase().includes(q))
  )
})

function confirmKick(type: 'pppoe' | 'hotspot', id: string, name: string) {
  kickTarget.value = { type, id, name }
  showKickModal.value = true
}

async function executeKick() {
  showKickModal.value = false
  if (kickTarget.value.type === 'pppoe') {
    await networkStore.kickPPPoESession(kickTarget.value.id)
  } else {
    await networkStore.kickHotspotSession(kickTarget.value.id)
  }
}

onMounted(() => {
  networkStore.fetchAll()
})
</script>

<style scoped>
.active-sessions-page {
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
  align-items: center;
}

.device-select {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-main);
  padding: 8px 12px;
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 600;
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

.filter-bar {
  display: flex;
  gap: 12px;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  max-width: 400px;
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

.user-avatar.bg-emerald {
  background: var(--color-success);
}

.uptime-badge {
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.05);
  padding: 2px 8px;
  border-radius: 6px;
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
