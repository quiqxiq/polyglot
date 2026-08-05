<template>
  <div class="sessions-view">
    <!-- Action Header -->
    <div class="view-header">
      <div>
        <h3>Manajemen Koneksi WhatsApp</h3>
        <p class="sub-text">Hubungkan nomor WhatsApp, atur balasan otomatis chatbot, dan konfigurasi URL Webhook per nomor.</p>
      </div>
      <button class="btn btn-primary" @click="showAddModal = true">
        <Plus class="w-4 h-4" />
        <span>Tambah Koneksi Baru</span>
      </button>
    </div>

    <!-- Sessions Table / Grid -->
    <div class="sessions-list glass-panel">
      <div v-if="sessionStore.loading" class="loading-box">
        <Loader2 class="w-8 h-8 spin text-indigo-400" />
        <span>Memuat daftar nomor WhatsApp...</span>
      </div>

      <div v-else-if="sessionStore.sessions.length === 0" class="empty-box">
        <Smartphone class="w-12 h-12 text-slate-600 mb-2" />
        <h4>Belum ada nomor WhatsApp tersambung</h4>
        <p>Klik tombol "Tambah Koneksi Baru" untuk menghubungkan nomor baru via Scan QR Code atau Kode Pairing.</p>
      </div>

      <div v-else class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th>Perangkat / Nomor</th>
              <th>Status Koneksi</th>
              <th>Chatbot Automasi</th>
              <th>Webhook Forwarder</th>
              <th>Terhubung Sejak</th>
              <th>Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="session in sessionStore.sessions" :key="session.id">
              <td>
                <div class="device-cell">
                  <div class="device-icon">
                    <Smartphone class="w-5 h-5 text-indigo-400" />
                  </div>
                  <div>
                    <strong class="device-name">{{ session.device_name }}</strong>
                    <span class="device-phone">{{ session.phone_number || 'Belum discan' }}</span>
                  </div>
                </div>
              </td>
              <td>
                <span :class="['badge', getStatusClass(session.status)]">
                  <span :class="['pulse-dot', session.status === 'online' ? 'pulse-dot-online' : '']"></span>
                  {{ getStatusText(session.status) }}
                </span>
              </td>
              <td>
                <label class="switch-label">
                  <span class="switch">
                    <input
                      type="checkbox"
                      :checked="session.is_bot_enabled"
                      @change="handleToggleBot(session.id, !session.is_bot_enabled)"
                    />
                    <span class="slider"></span>
                  </span>
                  <span class="text-xs font-semibold" :class="session.is_bot_enabled ? 'text-emerald-400' : 'text-slate-400'">
                    {{ session.is_bot_enabled ? 'Aktif (Bot Balas Otomatis)' : 'Mati (Manual Only)' }}
                  </span>
                </label>
              </td>
              <td>
                <div class="webhook-cell">
                  <span v-if="session.webhook_url" class="webhook-url" :title="session.webhook_url">
                    <Globe class="w-3.5 h-3.5 text-emerald-400 inline mr-1" />
                    {{ truncateUrl(session.webhook_url) }}
                  </span>
                  <span v-else class="text-xs text-muted">Belum Di-set</span>
                  <button
                    class="icon-btn ml-1"
                    title="Edit Webhook URL"
                    @click="openWebhookModal(session.id, session.webhook_url || '')"
                  >
                    <Edit2 class="w-3.5 h-3.5 text-indigo-400" />
                  </button>
                </div>
              </td>
              <td>
                <span class="text-xs text-muted">{{ formatDate(session.connected_at) }}</span>
              </td>
              <td>
                <div class="actions-cell">
                  <button
                    class="btn btn-secondary btn-sm"
                    title="Hubungkan Ulang Soket WhatsApp"
                    :disabled="reconnectingId === session.id"
                    @click="handleReconnect(session.id, session.device_name)"
                  >
                    <Loader2 v-if="reconnectingId === session.id" class="w-4 h-4 spin" />
                    <RefreshCw v-else class="w-4 h-4 text-cyan-400" />
                    <span>Reconnect</span>
                  </button>
                  <button
                    v-if="session.status === 'needs_rescan'"
                    class="btn btn-secondary btn-sm"
                    @click="handleShowQR(session.id, session.device_name)"
                  >
                    <QrCode class="w-4 h-4" />
                    <span>Scan / Kode</span>
                  </button>
                  <button
                    v-if="session.status === 'online'"
                    class="btn btn-warning btn-sm"
                    title="Logout / Unlink Perangkat"
                    @click="handleLogout(session.id)"
                  >
                    <LogOut class="w-4 h-4" />
                    <span>Logout</span>
                  </button>
                  <button
                    class="btn btn-danger btn-sm"
                    title="Hapus Koneksi"
                    @click="handleDelete(session.id)"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Add New Connection Modal -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="showAddModal = false">
      <div class="modal-card glass-panel">
        <div class="modal-header">
          <h3>Tambah Nomor WA Baru</h3>
          <button class="close-btn" @click="showAddModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <form @submit.prevent="handleCreateSession">
          <div class="input-group">
            <label class="input-label">Nama Perangkat / Label</label>
            <input
              v-model="newDeviceName"
              type="text"
              class="form-input"
              placeholder="mis. WA Customer Service 1"
              required
            />
          </div>
          <div class="input-group">
            <label class="input-label">Nomor HP (Opsional untuk Pairing Code)</label>
            <input
              v-model="newPhoneNumber"
              type="text"
              class="form-input"
              placeholder="mis. 628123456789"
            />
          </div>
          <div class="input-group">
            <label class="input-label">Webhook URL (Opsional)</label>
            <input
              v-model="newWebhookUrl"
              type="url"
              class="form-input"
              placeholder="https://example.com/api/wa-webhook"
            />
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showAddModal = false">Batal</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              <Loader2 v-if="submitting" class="w-4 h-4 spin" />
              <span>Buat & Tampilkan Tautan</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Edit Webhook Modal -->
    <div v-if="showWebhookModal" class="modal-overlay" @click.self="showWebhookModal = false">
      <div class="modal-card glass-panel">
        <div class="modal-header">
          <h3>Konfigurasi Webhook URL</h3>
          <button class="close-btn" @click="showWebhookModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <form @submit.prevent="handleSaveWebhook">
          <div class="input-group">
            <label class="input-label">URL Webhook Forwarder</label>
            <input
              v-model="webhookInputUrl"
              type="url"
              class="form-input"
              placeholder="https://domain-anda.com/webhook"
              required
            />
            <p class="field-hint">Server akan mengirimkan payload HTTP POST JSON setiap kali ada pesan WhatsApp masuk atau perubahan status sesi.</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showWebhookModal = false">Batal</button>
            <button type="submit" class="btn btn-primary" :disabled="savingWebhook">
              <Loader2 v-if="savingWebhook" class="w-4 h-4 spin" />
              <span>Simpan Webhook</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- QR / Pairing Code Modal -->
    <QRModal
      :show="showQRModal"
      :sessionId="selectedSessionId"
      :qrCode="sessionStore.activeQRCode"
      :sessionName="selectedSessionName"
      @close="showQRModal = false"
      @refresh="handleRefreshQR"
    />

    <!-- Logout Confirmation Modal -->
    <ConfirmModal
      :show="showLogoutModal"
      title="Logout WhatsApp Perangkat"
      subtitle="Memutuskan tautan akun WhatsApp dari server WhatsApp."
      message="Apakah Anda yakin ingin melakukan logout? Perangkat ini akan terlepas dari akun WhatsApp dan memerlukan Scan QR / Kode Pairing ulang."
      confirmText="Ya, Logout Perangkat"
      :loading="loggingOut"
      @confirm="confirmLogout"
      @cancel="showLogoutModal = false"
    />

    <!-- Delete Confirmation Modal -->
    <ConfirmModal
      :show="showDeleteModal"
      title="Hapus Koneksi WhatsApp"
      subtitle="Koneksi perangkat nomor WhatsApp ini akan dihapus permanen dari sistem."
      message="Apakah Anda yakin ingin menghapus koneksi nomor WhatsApp ini?"
      confirmText="Ya, Hapus Koneksi"
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="showDeleteModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  Plus,
  Smartphone,
  QrCode,
  Trash2,
  Loader2,
  RefreshCw,
  X,
  Globe,
  Edit2,
  LogOut,
} from 'lucide-vue-next'

import QRModal from '../components/QRModal.vue'
import ConfirmModal from '../components/ConfirmModal.vue'
import { useSessionStore } from '../stores/sessions'

const sessionStore = useSessionStore()

const showAddModal = ref(false)
const showQRModal = ref(false)
const showWebhookModal = ref(false)
const showLogoutModal = ref(false)

const selectedSessionId = ref<number | null>(null)
const selectedSessionName = ref('')

const newDeviceName = ref('')
const newPhoneNumber = ref('')
const newWebhookUrl = ref('')

const webhookInputUrl = ref('')
const webhookSessionId = ref<number | null>(null)
const savingWebhook = ref(false)

const submitting = ref(false)
const reconnectingId = ref<number | null>(null)
const logoutTargetId = ref<number | null>(null)
const loggingOut = ref(false)

const showDeleteModal = ref(false)
const deleteTargetId = ref<number | null>(null)
const deleting = ref(false)

onMounted(() => {
  sessionStore.fetchSessions()
})

async function handleReconnect(id: number, name: string) {
  reconnectingId.value = id
  try {
    const res = await sessionStore.reconnectSession(id)
    let qr = res?.qr_code
    if (!qr) {
      qr = await sessionStore.fetchQRCode(id)
    }
    if (qr) {
      selectedSessionId.value = id
      selectedSessionName.value = name
      showQRModal.value = true
    }
  } catch (err: any) {
    console.error('Reconnect error:', err)
  } finally {
    reconnectingId.value = null
  }
}

async function handleCreateSession() {
  submitting.value = true
  try {
    const res = await sessionStore.createSession(newDeviceName.value, newPhoneNumber.value, newWebhookUrl.value)
    showAddModal.value = false
    newDeviceName.value = ''
    newPhoneNumber.value = ''
    newWebhookUrl.value = ''

    selectedSessionId.value = res.session.id
    selectedSessionName.value = res.session.device_name
    showQRModal.value = true
  } finally {
    submitting.value = false
  }
}

function openWebhookModal(id: number, currentUrl: string) {
  webhookSessionId.value = id
  webhookInputUrl.value = currentUrl
  showWebhookModal.value = true
}

async function handleSaveWebhook() {
  if (!webhookSessionId.value) return
  savingWebhook.value = true
  try {
    await sessionStore.updateWebhook(webhookSessionId.value, webhookInputUrl.value)
    showWebhookModal.value = false
    webhookSessionId.value = null
    webhookInputUrl.value = ''
  } finally {
    savingWebhook.value = false
  }
}

async function handleShowQR(id: number, name: string) {
  selectedSessionId.value = id
  selectedSessionName.value = name
  showQRModal.value = true
  await sessionStore.fetchQRCode(id)
}

async function handleRefreshQR() {
  if (selectedSessionId.value) {
    await sessionStore.fetchQRCode(selectedSessionId.value)
  }
}

async function handleToggleBot(id: number, isEnabled: boolean) {
  await sessionStore.toggleBot(id, isEnabled)
}

function handleLogout(id: number) {
  logoutTargetId.value = id
  showLogoutModal.value = true
}

async function confirmLogout() {
  if (!logoutTargetId.value) return
  loggingOut.value = true
  try {
    await sessionStore.logoutSession(logoutTargetId.value)
    showLogoutModal.value = false
    logoutTargetId.value = null
  } finally {
    loggingOut.value = false
  }
}

function handleDelete(id: number) {
  deleteTargetId.value = id
  showDeleteModal.value = true
}

async function confirmDelete() {
  if (!deleteTargetId.value) return
  deleting.value = true
  try {
    await sessionStore.deleteSession(deleteTargetId.value)
    showDeleteModal.value = false
    deleteTargetId.value = null
  } finally {
    deleting.value = false
  }
}

function truncateUrl(url: string) {
  if (!url) return ''
  if (url.length > 28) {
    return url.substring(0, 25) + '...'
  }
  return url
}

function getStatusClass(status: string) {
  switch (status) {
    case 'online': return 'badge-online'
    case 'needs_rescan': return 'badge-warning'
    default: return 'badge-offline'
  }
}

function getStatusText(status: string) {
  switch (status) {
    case 'online': return 'Online'
    case 'needs_rescan': return 'Butuh Scan QR'
    default: return 'Offline'
  }
}

function formatDate(d: string) {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<style scoped>
.sessions-view {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.view-header {
  display: flex;
  align-items: center;
  justify-content: space-between;

  h3 {
    font-size: 20px;
    font-weight: 800;
  }
}

.sub-text {
  font-size: 13px;
  color: var(--text-muted);
}

.sessions-list {
  padding: 0;
  overflow: hidden;
}

.loading-box, .empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

.table-responsive {
  width: 100%;
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 14px;

  th {
    background: rgba(15, 23, 42, 0.6);
    color: var(--text-secondary);
    font-weight: 600;
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    padding: 14px 20px;
    border-bottom: 1px solid var(--border-color);
  }

  td {
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-color);
  }

  tr:last-child td {
    border-bottom: none;
  }
}

.device-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.device-icon {
  width: 38px;
  height: 38px;
  border-radius: var(--radius-md);
  background: rgba(99, 102, 241, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
}

.device-name {
  display: block;
  font-size: 14px;
}

.device-phone {
  display: block;
  font-size: 12px;
  color: var(--text-muted);
}

.webhook-cell {
  display: flex;
  align-items: center;
  gap: 4px;
}

.webhook-url {
  font-size: 12px;
  color: var(--text-main);
  font-family: monospace;
}

.icon-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;

  &:hover {
    background: rgba(255, 255, 255, 0.1);
  }
}

.field-hint {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 6px;
}

.actions-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(9, 13, 22, 0.85);
  backdrop-filter: blur(8px);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.modal-card {
  width: 100%;
  max-width: 440px;
  padding: 24px;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;

  h3 {
    font-size: 18px;
    font-weight: 700;
  }
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
}

.modal-footer {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@media (max-width: 640px) {
  .view-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  .view-header button {
    width: 100%;
  }
}
</style>
