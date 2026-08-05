<template>
  <div v-if="show" class="modal-overlay" @click.self="closeModal">
    <div class="modal-card glass-panel">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-title">
          <QrCode class="w-6 h-6 text-indigo-400" />
          <h3>Tautkan Perangkat WhatsApp</h3>
        </div>
        <button class="close-btn" @click="closeModal">
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Tab Switcher -->
      <div class="tab-switcher">
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'qr' }"
          @click="activeTab = 'qr'"
        >
          <QrCode class="w-4 h-4" />
          <span>Scan QR Code</span>
        </button>
        <button
          class="tab-btn"
          :class="{ active: activeTab === 'pairing' }"
          @click="activeTab = 'pairing'"
        >
          <KeyRound class="w-4 h-4" />
          <span>Kode Pairing WA</span>
        </button>
      </div>

      <div class="modal-body">
        <!-- TAB 1: SCAN QR CODE -->
        <div v-if="activeTab === 'qr'" class="tab-content">
          <p class="instruction">
            Buka WhatsApp di HP → <strong>Perangkat Tertaut</strong> → <strong>Tautkan Perangkat</strong>, lalu arahkan kamera ke QR Code di bawah:
          </p>

          <div class="qr-container">
            <div v-if="qrCode" class="qr-wrapper">
              <div class="qr-box">
                <img :src="qrImageUrl" alt="WhatsApp QR Code" class="qr-img" />
              </div>

              <!-- Live Countdown Progress Bar -->
              <div class="timer-bar-container">
                <div class="timer-info">
                  <span class="timer-label">
                    <Clock class="w-3.5 h-3.5 text-cyan-400 inline mr-1" />
                    QR Diperbarui Dalam:
                  </span>
                  <span class="timer-seconds">{{ countdown }}s</span>
                </div>
                <div class="progress-track">
                  <div
                    class="progress-fill"
                    :style="{ width: (countdown / 20) * 100 + '%' }"
                  ></div>
                </div>
              </div>
            </div>

            <div v-else class="qr-loading">
              <RefreshCw class="w-8 h-8 spin text-indigo-400" />
              <span>Menyiapkan QR Code WhatsApp...</span>
            </div>
          </div>

          <div class="refresh-action">
            <button class="btn btn-secondary text-xs" @click="$emit('refresh')">
              <RefreshCw class="w-3.5 h-3.5 mr-1" />
              Muat Ulang QR
            </button>
          </div>
        </div>

        <!-- TAB 2: PAIRING CODE VIA PHONE NUMBER -->
        <div v-else class="tab-content">
          <p class="instruction">
            Gunakan Kode Pairing 8-digit jika kamera HP Anda tidak bisa scan QR. Masukkan nomor WhatsApp Anda (format: 628xxx):
          </p>

          <div class="pairing-form">
            <div class="input-group">
              <label class="form-label">Nomor WhatsApp (dengan kode negara 62)</label>
              <input
                v-model="phoneNumber"
                type="text"
                class="form-input"
                placeholder="Contoh: 6281234567890"
              />
            </div>

            <button
              class="btn btn-primary w-full mt-2"
              :disabled="loadingPairing || !phoneNumber"
              @click="handleRequestPairingCode"
            >
              <RefreshCw v-if="loadingPairing" class="w-4 h-4 spin mr-2" />
              <KeyRound v-else class="w-4 h-4 mr-2" />
              <span>{{ loadingPairing ? 'Memproses Kode...' : 'Dapatkan Kode Pairing' }}</span>
            </button>
          </div>

          <!-- Display Pairing Code -->
          <div v-if="pairingCode" class="pairing-result">
            <div class="code-badge-label">Kode Pairing WhatsApp Anda:</div>
            <div class="code-box">
              <span class="code-text">{{ formattedCode }}</span>
              <button class="copy-btn" @click="copyCode">
                <Check v-if="copied" class="w-4 h-4 text-emerald-400" />
                <Copy v-else class="w-4 h-4" />
              </button>
            </div>
            <p class="pairing-guide">
              1. Buka WA di HP → <strong>Perangkat Tertaut</strong> → <strong>Tautkan Perangkat</strong>.<br/>
              2. Pilih <strong>Tautkan dengan nomor telepon saja</strong>.<br/>
              3. Masukkan kode 8 karakter di atas.
            </p>
          </div>
        </div>

        <div class="session-info" v-if="sessionName">
          <span class="text-secondary">Nama Perangkat:</span>
          <strong>{{ sessionName }}</strong>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" @click="closeModal">Tutup</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { X, QrCode, KeyRound, RefreshCw, Clock, Copy, Check } from 'lucide-vue-next'
import { useSessionStore } from '../stores/sessions'

const props = defineProps<{
  show: boolean
  sessionId?: number | null
  qrCode?: string | null
  sessionName?: string
}>()

const emit = defineEmits(['close', 'refresh'])

const sessionStore = useSessionStore()
const activeTab = ref<'qr' | 'pairing'>('qr')
const phoneNumber = ref('')
const pairingCode = ref<string | null>(null)
const loadingPairing = ref(false)
const copied = ref(false)

// Countdown timer state
const countdown = ref(20)
let timerInterval: any = null

function startCountdown() {
  stopCountdown()
  countdown.value = 20
  timerInterval = setInterval(() => {
    if (countdown.value > 1) {
      countdown.value--
    } else {
      // Refresh QR when countdown reaches 0
      countdown.value = 20
      emit('refresh')
    }
  }, 1000)
}

function stopCountdown() {
  if (timerInterval) {
    clearInterval(timerInterval)
    timerInterval = null
  }
}

watch(
  () => props.qrCode,
  (newQR) => {
    if (newQR && props.show) {
      startCountdown()
    }
  },
  { immediate: true }
)

watch(
  () => props.show,
  (isShown) => {
    if (isShown) {
      startCountdown()
    } else {
      stopCountdown()
    }
  }
)

onUnmounted(() => {
  stopCountdown()
})

const qrImageUrl = computed(() => {
  if (!props.qrCode) return ''
  return `https://api.qrserver.com/v1/create-qr-code/?size=240x240&data=${encodeURIComponent(props.qrCode)}`
})

const formattedCode = computed(() => {
  if (!pairingCode.value) return ''
  const clean = pairingCode.value.replace('-', '')
  if (clean.length === 8) {
    return `${clean.slice(0, 4)} - ${clean.slice(4)}`
  }
  return pairingCode.value
})

async function handleRequestPairingCode() {
  if (!props.sessionId || !phoneNumber.value) return
  loadingPairing.value = true
  try {
    const code = await sessionStore.fetchPairingCode(props.sessionId, phoneNumber.value)
    pairingCode.value = code
  } catch (err: any) {
    alert(err.message || 'Gagal mendapatkan kode pairing')
  } finally {
    loadingPairing.value = false
  }
}

function copyCode() {
  if (!formattedCode.value) return
  navigator.clipboard.writeText(formattedCode.value.replace(/\s+/g, ''))
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 2000)
}

function closeModal() {
  stopCountdown()
  emit('close')
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(9, 13, 22, 0.85);
  backdrop-filter: blur(10px);
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
  margin-bottom: 16px;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;

  h3 {
    font-size: 17px;
    font-weight: 700;
    margin: 0;
  }
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;

  &:hover {
    color: var(--text-main);
  }
}

.tab-switcher {
  display: flex;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 4px;
  margin-bottom: 20px;
  gap: 4px;
}

.tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 12px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 13px;
  font-weight: 600;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s ease;

  &.active {
    background: var(--primary);
    color: #ffffff;
    box-shadow: var(--shadow-sm);
  }
}

.instruction {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
  margin-bottom: 16px;
}

.qr-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin: 12px 0;
}

.qr-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  width: 100%;
}

.qr-box {
  background: #ffffff;
  padding: 14px;
  border-radius: var(--radius-lg);
  box-shadow: 0 0 25px rgba(255, 255, 255, 0.1);
}

.qr-img {
  width: 210px;
  height: 210px;
  display: block;
}

.timer-bar-container {
  width: 100%;
  max-width: 250px;
}

.timer-info {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 4px;
}

.timer-seconds {
  font-weight: 700;
  color: var(--accent-cyan);
}

.progress-track {
  height: 4px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--primary), var(--accent-cyan));
  transition: width 1s linear;
}

.refresh-action {
  display: flex;
  justify-content: center;
  margin-top: 10px;
}

.qr-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 40px;
  color: var(--text-secondary);
  font-size: 14px;
}

.pairing-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.pairing-result {
  margin-top: 20px;
  padding: 16px;
  background: rgba(99, 102, 241, 0.08);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: var(--radius-md);
  text-align: center;
}

.code-badge-label {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.code-box {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  background: #090d16;
  padding: 12px 20px;
  border-radius: var(--radius-md);
  border: 1px dashed var(--primary);
}

.code-text {
  font-family: 'JetBrains Mono', monospace;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: 2px;
  color: #6366f1;
}

.copy-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;

  &:hover {
    color: #ffffff;
  }
}

.pairing-guide {
  font-size: 11px;
  color: var(--text-muted);
  text-align: left;
  line-height: 1.6;
  margin-top: 12px;
  margin-bottom: 0;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.session-info {
  margin-top: 16px;
  font-size: 12px;
  text-align: center;
  display: flex;
  justify-content: center;
  gap: 6px;
}

.modal-footer {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
