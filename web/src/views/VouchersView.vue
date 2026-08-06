<template>
  <div class="vouchers-page animate-fade-in">
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1 class="page-title gradient-text">Mikhmon Hotspot & Voucher Engine</h1>
        <p class="page-desc">Generate voucher massal, cetak layout HTML/PDF dengan QR Code, dan laporan penjualan</p>
      </div>

      <div class="header-actions">
        <select v-model="mikhmonStore.selectedDeviceId" class="device-select" @change="handleDeviceChange">
          <option value="" disabled v-if="deviceStore.devices.length === 0">
            -- Belum Ada Router Terdaftar --
          </option>
          <option v-for="dev in deviceStore.devices" :key="dev.id" :value="dev.id">
            {{ dev.name }} ({{ dev.id }})
          </option>
        </select>
      </div>
    </div>

    <!-- Tabs Navigation -->
    <div class="tab-header mb-6">
      <button
        :class="['tab-btn', { active: activeTab === 'generate' }]"
        @click="activeTab = 'generate'"
      >
        <Ticket class="w-4 h-4 inline mr-2" />
        Generate & Cetak Voucher
      </button>
      <button
        :class="['tab-btn', { active: activeTab === 'reports' }]"
        @click="activeTab = 'reports'"
      >
        <TrendingUp class="w-4 h-4 inline mr-2" />
        Laporan Penjualan (Income Report)
      </button>
    </div>

    <!-- Tab 1: Generate Voucher Form -->
    <div v-if="activeTab === 'generate'" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Form Input -->
      <div class="glass-panel p-6 lg:col-span-1">
        <h2 class="text-lg font-bold mb-4 text-main flex items-center gap-2">
          <PlusCircle class="w-5 h-5 text-indigo-500" />
          Form Batch Voucher
        </h2>

        <form @submit.prevent="handleGenerate">
          <div class="form-group mb-4">
            <label class="form-label">Jumlah Voucher (Qty)</label>
            <input
              v-model.number="form.qty"
              type="number"
              min="1"
              max="500"
              class="form-control"
              required
            />
          </div>

          <div class="form-group mb-4">
            <label class="form-label">Hotspot Profile</label>
            <select v-model="form.profile" class="form-control" required>
              <option value="" disabled>-- Pilih Profile --</option>
              <option v-for="prof in mikhmonStore.profiles" :key="prof.id" :value="prof.name">
                {{ prof.name }} {{ prof.price ? `(Rp ${prof.price})` : '' }}
              </option>
            </select>
          </div>

          <div class="grid grid-cols-2 gap-4 mb-4">
            <div class="form-group">
              <label class="form-label">Mode Server</label>
              <select v-model="form.server_mode" class="form-control">
                <option value="all">all</option>
                <option value="hotspot1">hotspot1</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">Tipe User</label>
              <select v-model="form.user_type" class="form-control">
                <option value="vc">Voucher (Username=Password)</option>
                <option value="up">User & Password (Berbeda)</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4 mb-4">
            <div class="form-group">
              <label class="form-label">Prefix Kode</label>
              <input v-model="form.prefix" type="text" placeholder="misal: VC" class="form-control" />
            </div>
            <div class="form-group">
              <label class="form-label">Panjang Karakter</label>
              <select v-model.number="form.char_len" class="form-control">
                <option :value="4">4 Karakter</option>
                <option :value="6">6 Karakter</option>
                <option :value="8">8 Karakter</option>
              </select>
            </div>
          </div>

          <div class="form-group mb-6">
            <label class="form-label">Set Karakter Acak</label>
            <select v-model="form.char_set" class="form-control">
              <option value="mix">Campuran (Huruf Kecil & Angka)</option>
              <option value="num">Hanya Angka (123456)</option>
              <option value="alpha">Hanya Huruf (abcde)</option>
            </select>
          </div>

          <button
            type="submit"
            class="btn btn-primary w-full justify-center"
            :disabled="mikhmonStore.loading"
          >
            <RefreshCw v-if="mikhmonStore.loading" class="w-4 h-4 mr-2 animate-spin" />
            <Sparkles v-else class="w-4 h-4 mr-2" />
            Generate Batch Voucher
          </button>
        </form>
      </div>

      <!-- Generated Batch Results & Quick Print Options -->
      <div class="glass-panel p-6 lg:col-span-2 flex flex-col">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-bold text-main flex items-center gap-2">
            <Ticket class="w-5 h-5 text-emerald-500" />
            Hasil Batch ({{ mikhmonStore.generatedVouchers.length }} Voucher)
          </h2>

          <div v-if="mikhmonStore.generatedVouchers.length > 0" class="flex gap-2">
            <button class="btn btn-secondary btn-sm" @click="openPrintModal">
              <Printer class="w-4 h-4 mr-1" />
              Preview & Cetak
            </button>
            <button class="btn btn-primary btn-sm" @click="openSendWAModal">
              <Send class="w-4 h-4 mr-1" />
              Kirim via WA
            </button>
          </div>
        </div>

        <div v-if="mikhmonStore.generatedVouchers.length === 0" class="empty-state flex-1 flex flex-col items-center justify-center py-12 text-center text-muted">
          <Ticket class="w-12 h-12 mb-3 text-muted/50" />
          <p>Belum ada batch voucher yang di-generate.</p>
          <p class="text-xs text-muted mt-1">Gunakan form di sebelah kiri untuk membuat kode voucher baru.</p>
        </div>

        <div v-else class="table-responsive flex-1 max-h-[500px] overflow-y-auto">
          <table class="data-table">
            <thead>
              <tr>
                <th>#</th>
                <th>Username</th>
                <th>Password</th>
                <th>Profile</th>
                <th>Harga</th>
                <th>Masa Aktif</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(v, index) in mikhmonStore.generatedVouchers" :key="index">
                <td class="text-muted">{{ index + 1 }}</td>
                <td class="font-mono font-bold text-main">{{ v.username }}</td>
                <td class="font-mono text-cyan">{{ v.password }}</td>
                <td><span class="badge badge-primary">{{ v.profile }}</span></td>
                <td class="text-emerald-400 font-bold">Rp {{ v.price || '-' }}</td>
                <td class="text-xs text-muted">{{ v.validity || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Tab 2: Sales Report -->
    <div v-if="activeTab === 'reports'" class="glass-panel p-6">
      <div class="flex flex-wrap justify-between items-center gap-4 mb-6">
        <h2 class="text-lg font-bold text-main flex items-center gap-2">
          <TrendingUp class="w-5 h-5 text-emerald-500" />
          Laporan Transaksi Voucher
        </h2>

        <div class="flex gap-2">
          <input v-model="filterDate" type="date" class="form-control" @change="loadReports" />
          <button class="btn btn-secondary" @click="loadReports">
            <RefreshCw class="w-4 h-4 mr-1" />
            Muat Laporan
          </button>
        </div>
      </div>

      <!-- Report Summary -->
      <div class="stats-grid mb-6">
        <div class="stat-card glass-panel">
          <div class="stat-icon bg-emerald-500/10 text-emerald-500">
            <DollarSign class="w-6 h-6" />
          </div>
          <div class="stat-content">
            <span class="stat-label">Total Pendapatan</span>
            <span class="stat-value text-emerald-400">Rp {{ totalIncome.toLocaleString('id-ID') }}</span>
          </div>
        </div>

        <div class="stat-card glass-panel">
          <div class="stat-icon bg-indigo-500/10 text-indigo-500">
            <Ticket class="w-6 h-6" />
          </div>
          <div class="stat-content">
            <span class="stat-label">Total Voucher Terjual</span>
            <span class="stat-value">{{ mikhmonStore.reports.length }}</span>
          </div>
        </div>
      </div>

      <!-- Reports Table -->
      <div class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th>Tanggal</th>
              <th>User Voucher</th>
              <th>Profile</th>
              <th>Harga (Rp)</th>
              <th>Keterangan</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="mikhmonStore.reports.length === 0">
              <td colspan="5" class="text-center py-8 text-muted">
                Tidak ada data laporan penjualan pada tanggal ini.
              </td>
            </tr>
            <tr v-for="(r, idx) in mikhmonStore.reports" :key="idx">
              <td class="font-mono text-xs text-muted">{{ r.date }} {{ r.time || '' }}</td>
              <td class="font-mono font-bold text-main">{{ r.user }}</td>
              <td><span class="badge badge-primary">{{ r.profile }}</span></td>
              <td class="text-emerald-400 font-bold">Rp {{ r.price ? r.price.toLocaleString('id-ID') : 0 }}</td>
              <td class="text-xs text-muted">{{ r.comment || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Print Modal -->
    <div v-if="showPrintModal" class="modal-overlay">
      <div class="modal-content glass-panel max-w-4xl w-full p-6">
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-lg font-bold text-main">Preview Cetak Voucher</h3>
          <button class="btn btn-sm btn-secondary" @click="showPrintModal = false">Tutup</button>
        </div>

        <div class="preview-container border border-white/10 rounded-lg p-4 max-h-[60vh] overflow-y-auto bg-white text-black mb-4">
          <iframe :srcdoc="mikhmonStore.renderedHTML" class="w-full h-[500px] border-0"></iframe>
        </div>

        <div class="flex justify-end gap-2">
          <button class="btn btn-secondary" @click="showPrintModal = false">Batal</button>
          <button class="btn btn-primary" @click="printIframe">
            <Printer class="w-4 h-4 mr-1" />
            Cetak Sekarang
          </button>
        </div>
      </div>
    </div>

    <!-- Send WA Modal -->
    <div v-if="showSendWAModal" class="modal-overlay">
      <div class="modal-content glass-panel max-w-md w-full p-6">
        <h3 class="text-lg font-bold text-main mb-4 flex items-center gap-2">
          <Send class="w-5 h-5 text-emerald-500" />
          Kirim Voucher via WA
        </h3>

        <div class="form-group mb-4">
          <label class="form-label">Sesi WA Pengirim</label>
          <select v-model.number="sendWAForm.sessionId" class="form-control" required>
            <option v-for="s in waSessions" :key="s.id" :value="s.id">
              {{ s.device_name }} ({{ s.phone_number || s.status }})
            </option>
          </select>
        </div>

        <div class="form-group mb-4">
          <label class="form-label">Nomor WhatsApp Penerima</label>
          <input
            v-model="sendWAForm.recipient"
            type="text"
            placeholder="misal: 081234567890"
            class="form-control"
            required
          />
        </div>

        <div class="form-group mb-6">
          <label class="form-label">Pesan / Catatan</label>
          <input
            v-model="sendWAForm.caption"
            type="text"
            placeholder="Berikut kode voucher hotspot Anda"
            class="form-control"
          />
        </div>

        <div class="flex justify-end gap-2">
          <button class="btn btn-secondary" @click="showSendWAModal = false">Batal</button>
          <button class="btn btn-primary" :disabled="sendingWA" @click="executeSendWA">
            <RefreshCw v-if="sendingWA" class="w-4 h-4 mr-1 animate-spin" />
            <Send v-else class="w-4 h-4 mr-1" />
            Kirim Dokumen
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Ticket,
  TrendingUp,
  PlusCircle,
  Sparkles,
  Printer,
  Send,
  RefreshCw,
  DollarSign,
} from 'lucide-vue-next'
import { useMikhmonStore } from '../stores/mikhmon'
import { useSessionStore } from '../stores/sessions'
import { useDeviceStore } from '../stores/devices'
import type { VoucherBatchRequest } from '../types'

const mikhmonStore = useMikhmonStore()
const sessionsStore = useSessionStore()
const deviceStore = useDeviceStore()

const activeTab = ref<'generate' | 'reports'>('generate')
const showPrintModal = ref(false)
const showSendWAModal = ref(false)
const sendingWA = ref(false)

const filterDate = ref(new Date().toISOString().split('T')[0])

const form = ref<VoucherBatchRequest>({
  qty: 10,
  profile: '',
  server_mode: 'all',
  user_type: 'vc',
  prefix: 'VC',
  char_len: 4,
  char_set: 'mix',
})

const sendWAForm = ref({
  sessionId: 1,
  recipient: '',
  caption: 'Berikut file cetak voucher hotspot Anda.',
})

const waSessions = computed(() => sessionsStore.sessions)

const totalIncome = computed(() => {
  return mikhmonStore.reports.reduce((sum, r) => sum + (r.price || 0), 0)
})

function handleDeviceChange() {
  mikhmonStore.fetchProfiles()
}

async function handleGenerate() {
  try {
    const vouchers = await mikhmonStore.generateVouchers(form.value)
    if (vouchers && vouchers.length > 0) {
      await mikhmonStore.renderHTML()
    }
  } catch (e: any) {
    alert(e.message || 'Gagal generate voucher')
  }
}

async function openPrintModal() {
  if (!mikhmonStore.renderedHTML) {
    await mikhmonStore.renderHTML()
  }
  showPrintModal.value = true
}

function printIframe() {
  const iframe = document.querySelector('iframe')
  if (iframe && iframe.contentWindow) {
    iframe.contentWindow.focus()
    iframe.contentWindow.print()
  }
}

async function openSendWAModal() {
  await sessionsStore.fetchSessions()
  if (sessionsStore.sessions.length > 0) {
    sendWAForm.value.sessionId = sessionsStore.sessions[0].id
  }
  showSendWAModal.value = true
}

async function executeSendWA() {
  if (!sendWAForm.value.recipient) {
    alert('Nomor penerima wajib diisi')
    return
  }

  try {
    sendingWA.value = true
    if (!mikhmonStore.renderedHTML) {
      await mikhmonStore.renderHTML()
    }
    const fileName = `Voucher-Batch-${Date.now()}.html`
    await mikhmonStore.sendVouchersToWA(
      sendWAForm.value.sessionId,
      sendWAForm.value.recipient,
      fileName,
      mikhmonStore.renderedHTML,
      sendWAForm.value.caption
    )
    alert('Voucher berhasil dikirim ke WhatsApp!')
    showSendWAModal.value = false
  } catch (e: any) {
    alert('Gagal mengirim WA: ' + (e.message || e))
  } finally {
    sendingWA.value = false
  }
}

function loadReports() {
  mikhmonStore.fetchReports(filterDate.value)
}

onMounted(async () => {
  await deviceStore.fetchDevices()
  if (deviceStore.devices.length > 0 && !mikhmonStore.selectedDeviceId) {
    mikhmonStore.selectedDeviceId = deviceStore.devices[0].id
  }
  if (mikhmonStore.selectedDeviceId) {
    mikhmonStore.fetchProfiles()
    loadReports()
  }
})
</script>

<style scoped>
.vouchers-page {
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

.device-select {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-main);
  padding: 8px 12px;
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 600;
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

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
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
