<template>
  <div class="llm-view">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h3>Konfigurasi Provider LLM</h3>
        <p class="sub-text">Pilih provider AI (Gemini, Claude, GPT, Groq), atur API key, dan aktifkan model tanpa ubah kode.</p>
      </div>
      <button class="btn btn-primary shadow-glow" @click="openCreateModal">
        <Plus class="w-4 h-4" />
        <span>Tambah Konfigurasi LLM</span>
      </button>
    </div>

    <!-- Active LLM Highlight Banner -->
    <div v-if="llmStore.activeConfig" class="active-card glass-panel">
      <div class="active-badge-tag">
        <CheckCircle2 class="w-4 h-4 text-emerald-500" />
        <span>Provider Aktif Saat Ini</span>
      </div>
      <div class="active-details">
        <div class="provider-logo-box">
          <Cpu class="w-8 h-8 text-indigo-500" />
        </div>
        <div>
          <h4 class="text-main font-bold">{{ llmStore.activeConfig.provider.toUpperCase() }}</h4>
          <p class="model-name">{{ llmStore.activeConfig.model }}</p>
          <p class="tokens-info">Max Output Tokens: {{ llmStore.activeConfig.max_output_tokens || 512 }}</p>
        </div>
      </div>
    </div>

    <!-- Configs List -->
    <div class="configs-grid">
      <div v-if="llmStore.loading" class="loading-box glass-panel col-span-full">
        <Loader2 class="w-8 h-8 spin text-indigo-500 mb-2" />
        <span>Memuat konfigurasi LLM...</span>
      </div>

      <div v-else-if="llmStore.configs.length === 0" class="empty-box glass-panel col-span-full">
        <Cpu class="w-12 h-12 text-slate-400 mb-2" />
        <h4>Belum ada konfigurasi LLM</h4>
        <p>Klik tombol "Tambah Konfigurasi LLM" untuk memasukkan API Key Gemini, Claude, atau OpenAI.</p>
      </div>

      <div
        v-for="cfg in llmStore.configs"
        :key="cfg.id"
        :class="['config-card', 'glass-panel', { 'config-card-active': cfg.is_active }]"
      >
        <div class="card-top">
          <div class="provider-info">
            <div :class="['provider-icon', getProviderClass(cfg.provider)]">
              <Cpu class="w-5 h-5" />
            </div>
            <div>
              <strong class="provider-title">{{ getProviderName(cfg.provider) }}</strong>
              <span class="model-code">{{ cfg.model }}</span>
            </div>
          </div>
          <span :class="['badge', cfg.is_active ? 'badge-online' : 'badge-offline']">
            {{ cfg.is_active ? 'Aktif' : 'Non-Aktif' }}
          </span>
        </div>

        <div class="card-meta">
          <div class="meta-row">
            <span>API Key:</span>
            <code>••••••••••••••••</code>
          </div>
          <div class="meta-row">
            <span>Max Output Tokens:</span>
            <strong>{{ cfg.max_output_tokens || 512 }}</strong>
          </div>
          <div class="meta-row">
            <span>Tarif (per 1M Token):</span>
            <span class="rate-badge">In: ${{ cfg.cost_per_1m_input || 0 }} | Out: ${{ cfg.cost_per_1m_output || 0 }}</span>
          </div>
        </div>

        <!-- Token Usage & Cost Analytics Card -->
        <div class="analytics-card">
          <div class="analytics-header">
            <Coins class="w-3.5 h-3.5 text-amber-500" />
            <span>Penggunaan Token & Estimasi Biaya</span>
          </div>
          <div class="analytics-grid">
            <div class="analytics-item">
              <span class="analytics-label">Total Token</span>
              <strong class="analytics-val text-indigo-500">
                {{ formatNumber((cfg.total_input_tokens || 0) + (cfg.total_output_tokens || 0)) }}
              </strong>
              <span class="analytics-sub">In: {{ formatNumber(cfg.total_input_tokens || 0) }} | Out: {{ formatNumber(cfg.total_output_tokens || 0) }}</span>
            </div>
            <div class="analytics-item">
              <span class="analytics-label">Respon AI</span>
              <strong class="analytics-val text-cyan-600">
                {{ formatNumber(cfg.total_messages || 0) }}
              </strong>
              <span class="analytics-sub">Balasan Chat</span>
            </div>
            <div class="analytics-item">
              <span class="analytics-label">Est. Biaya (USD)</span>
              <strong class="analytics-val text-emerald-600">
                ${{ (cfg.total_cost_usd || 0).toFixed(4) }}
              </strong>
              <span class="analytics-sub">Kalkulasi 1M Token</span>
            </div>
            <div class="analytics-item">
              <span class="analytics-label">Est. Biaya (IDR)</span>
              <strong class="analytics-val text-emerald-600">
                Rp {{ formatRupiah(cfg.total_cost_idr || 0) }}
              </strong>
              <span class="analytics-sub">1 USD = Rp 18.022</span>
            </div>
          </div>
        </div>

        <!-- Test Result Alert -->
        <div v-if="testResults[cfg.id]" :class="['test-alert', testResults[cfg.id].status === 'success' ? 'alert-success' : 'alert-danger']">
          <CheckCircle2 v-if="testResults[cfg.id].status === 'success'" class="w-4 h-4" />
          <AlertCircle v-else class="w-4 h-4" />
          <span>{{ testResults[cfg.id].message }}</span>
        </div>

        <!-- Card Action Controls -->
        <div class="card-actions">
          <button
            class="btn btn-secondary btn-sm"
            :disabled="llmStore.testingId === cfg.id"
            @click="handleTest(cfg.id)"
          >
            <Loader2 v-if="llmStore.testingId === cfg.id" class="w-3.5 h-3.5 spin" />
            <Zap v-else class="w-3.5 h-3.5 text-amber-500" />
            <span>Test Connection</span>
          </button>

          <button
            v-if="!cfg.is_active"
            class="btn btn-success btn-sm"
            @click="handleActivate(cfg.id)"
          >
            <Check class="w-3.5 h-3.5" />
            <span>Aktifkan</span>
          </button>

          <!-- Edit Button -->
          <button
            class="btn btn-secondary btn-sm icon-only-btn"
            title="Edit Konfigurasi"
            @click="openEditModal(cfg)"
          >
            <Pencil class="w-4 h-4" />
          </button>

          <!-- Delete Button -->
          <button
            class="btn btn-danger btn-sm icon-only-btn"
            title="Hapus Konfigurasi"
            @click="handleDelete(cfg.id)"
          >
            <Trash2 class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>

    <!-- Create / Edit LLM Config Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
      <div class="modal-card glass-panel">
        <div class="modal-header">
          <h3>{{ isEditing ? 'Edit Konfigurasi LLM' : 'Tambah Konfigurasi Provider LLM' }}</h3>
          <button class="close-btn" @click="showModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <form @submit.prevent="handleSubmit">
          <div class="input-group">
            <label class="input-label">Pilih Provider AI</label>
            <select v-model="formProvider" class="form-select" @change="updateDefaultModel">
              <option value="gemini">Google Gemini</option>
              <option value="groq">Groq Cloud (LPU Ultra Fast)</option>
              <option value="claude">Anthropic Claude</option>
              <option value="openai">OpenAI (GPT)</option>
            </select>
          </div>

          <div class="input-group">
            <label class="input-label">Model AI</label>
            <select v-model="formModel" class="form-select">
              <template v-if="formProvider === 'gemini'">
                <option value="gemini-2.0-flash">gemini-2.0-flash (Rekomendasi Cepat & Hemat)</option>
                <option value="gemini-1.5-pro">gemini-1.5-pro (Penalaran Dalam)</option>
              </template>
              <template v-else-if="formProvider === 'groq'">
                <option value="llama-3.3-70b-versatile">llama-3.3-70b-versatile (Rekomendasi 70B Ultra Fast)</option>
                <option value="llama-3.1-8b-instant">llama-3.1-8b-instant (Super Cepat ~500 tok/s)</option>
                <option value="mixtral-8x7b-32768">mixtral-8x7b-32768 (MoE 32k Context)</option>
                <option value="gemma2-9b-it">gemma2-9b-it (Google Gemma 2)</option>
              </template>
              <template v-else-if="formProvider === 'claude'">
                <option value="claude-sonnet-4-20250514">claude-sonnet-4 (Cepat & Responsif)</option>
                <option value="claude-haiku-3-5">claude-3-5-haiku (Ultra Hemat Token)</option>
              </template>
              <template v-else-if="formProvider === 'openai'">
                <option value="gpt-4o-mini">gpt-4o-mini (Cepat & Murah)</option>
                <option value="gpt-4o">gpt-4o (Reasoning Tinggi)</option>
              </template>
            </select>
          </div>

          <div class="input-group">
            <label class="input-label">API Key Provider</label>
            <input
              v-model="formApiKey"
              type="password"
              class="form-input"
              :placeholder="isEditing ? 'Kosongkan jika tidak ingin mengubah API Key' : 'Masukkan API Key yang valid'"
              :required="!isEditing"
            />
            <span class="text-xs text-muted mt-1">Disimpan terenkripsi kuat dengan AES-256-GCM.</span>
          </div>

          <div class="input-group">
            <label class="input-label">Max Output Tokens</label>
            <input
              v-model.number="formMaxTokens"
              type="number"
              class="form-input"
              placeholder="512"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="input-group">
              <label class="input-label">Tarif Input / 1M Token (USD)</label>
              <input
                v-model.number="formCostInput"
                type="number"
                step="0.0001"
                class="form-input"
                placeholder="0.075"
              />
            </div>
            <div class="input-group">
              <label class="input-label">Tarif Output / 1M Token (USD)</label>
              <input
                v-model.number="formCostOutput"
                type="number"
                step="0.0001"
                class="form-input"
                placeholder="0.300"
              />
            </div>
          </div>
          <div class="mb-4 flex items-center justify-between">
            <span class="text-xs text-muted">Tarif acuan per 1 Juta Token resmi.</span>
            <button type="button" class="btn btn-secondary btn-xs flex items-center gap-1" @click="autoFillRates">
              <RefreshCw class="w-3 h-3 text-indigo-500" />
              <span>Isi Tarif Resmi</span>
            </button>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showModal = false">Batal</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              <Loader2 v-if="submitting" class="w-4 h-4 spin" />
              <span>{{ isEditing ? 'Perbarui Konfigurasi' : 'Simpan Konfigurasi' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <ConfirmModal
      :show="showDeleteModal"
      title="Hapus Konfigurasi LLM"
      subtitle="Tindakan ini akan menghapus API Key dan pengaturan provider."
      message="Apakah Anda yakin ingin menghapus konfigurasi LLM ini? Pilihan yang sedang aktif tidak akan dapat merespons jika dihapus."
      confirmText="Ya, Hapus Konfigurasi"
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="showDeleteModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import {
  Plus,
  Cpu,
  CheckCircle2,
  AlertCircle,
  Zap,
  Check,
  Loader2,
  X,
  Pencil,
  Trash2,
  Coins,
  RefreshCw,
} from 'lucide-vue-next'

import ConfirmModal from '../components/ConfirmModal.vue'
import type { LLMConfig } from '../types'
import { useLLMStore } from '../stores/llm'

const llmStore = useLLMStore()

const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)

const formProvider = ref<'gemini' | 'groq' | 'claude' | 'openai'>('gemini')
const formModel = ref('gemini-2.0-flash')
const formApiKey = ref('')
const formMaxTokens = ref(512)
const formCostInput = ref(0.075)
const formCostOutput = ref(0.300)

const testResults = reactive<Record<number, { status: string; message: string }>>({})

const officialRates: Record<string, { in: number; out: number }> = {
  'gemini-2.0-flash': { in: 0.075, out: 0.300 },
  'gemini-1.5-flash': { in: 0.075, out: 0.300 },
  'gemini-1.5-pro': { in: 1.250, out: 5.000 },
  'gpt-4o-mini': { in: 0.150, out: 0.600 },
  'gpt-4o': { in: 2.500, out: 10.000 },
  'gpt-3.5-turbo': { in: 0.500, out: 1.500 },
  'claude-sonnet-4-20250514': { in: 3.000, out: 15.000 },
  'claude-3-5-sonnet-20241022': { in: 3.000, out: 15.000 },
  'claude-3-5-haiku-20241022': { in: 0.800, out: 4.000 },
  'llama-3.3-70b-versatile': { in: 0.590, out: 0.790 },
  'llama-3.1-8b-instant': { in: 0.050, out: 0.080 },
  'mixtral-8x7b-32768': { in: 0.240, out: 0.240 },
}

onMounted(() => {
  llmStore.fetchConfigs()
})

watch(formModel, () => {
  autoFillRates()
})

function autoFillRates() {
  const rate = officialRates[formModel.value]
  if (rate) {
    formCostInput.value = rate.in
    formCostOutput.value = rate.out
  } else {
    formCostInput.value = 0.100
    formCostOutput.value = 0.300
  }
}

function updateDefaultModel() {
  if (formProvider.value === 'gemini') formModel.value = 'gemini-2.0-flash'
  else if (formProvider.value === 'groq') formModel.value = 'llama-3.3-70b-versatile'
  else if (formProvider.value === 'claude') formModel.value = 'claude-3-5-sonnet-20241022'
  else if (formProvider.value === 'openai') formModel.value = 'gpt-4o-mini'
  autoFillRates()
}

function openCreateModal() {
  isEditing.value = false
  editingId.value = null
  formProvider.value = 'gemini'
  formModel.value = 'gemini-2.0-flash'
  formApiKey.value = ''
  formMaxTokens.value = 512
  autoFillRates()
  showModal.value = true
}

function openEditModal(cfg: LLMConfig) {
  isEditing.value = true
  editingId.value = cfg.id
  formProvider.value = (cfg.provider as any) || 'gemini'
  formModel.value = cfg.model
  formApiKey.value = ''
  formMaxTokens.value = cfg.max_output_tokens || 512
  formCostInput.value = cfg.cost_per_1m_input || 0.075
  formCostOutput.value = cfg.cost_per_1m_output || 0.300
  showModal.value = true
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (isEditing.value && editingId.value) {
      await llmStore.updateConfig(editingId.value, {
        provider: formProvider.value,
        model: formModel.value,
        api_key: formApiKey.value || undefined,
        max_output_tokens: formMaxTokens.value,
        cost_per_1m_input: formCostInput.value,
        cost_per_1m_output: formCostOutput.value,
      })
    } else {
      await llmStore.createConfig({
        provider: formProvider.value,
        model: formModel.value,
        api_key: formApiKey.value,
        max_output_tokens: formMaxTokens.value,
        cost_per_1m_input: formCostInput.value,
        cost_per_1m_output: formCostOutput.value,
      })
    }
    showModal.value = false
    formApiKey.value = ''
  } finally {
    submitting.value = false
  }
}

function formatNumber(num: number): string {
  return new Intl.NumberFormat('id-ID').format(num)
}

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount)
}

const showDeleteModal = ref(false)
const deleteTargetId = ref<number | null>(null)
const deleting = ref(false)

function handleDelete(id: number) {
  deleteTargetId.value = id
  showDeleteModal.value = true
}

async function confirmDelete() {
  if (!deleteTargetId.value) return
  deleting.value = true
  try {
    await llmStore.deleteConfig(deleteTargetId.value)
    showDeleteModal.value = false
    deleteTargetId.value = null
  } finally {
    deleting.value = false
  }
}

async function handleActivate(id: number) {
  await llmStore.activateConfig(id)
}

async function handleTest(id: number) {
  delete testResults[id]
  try {
    const res = await llmStore.testConfig(id)
    testResults[id] = { status: 'success', message: res.message }
  } catch (err: any) {
    testResults[id] = { status: 'error', message: err.message || 'Koneksi gagal' }
  }
}

function getProviderName(p: string) {
  switch (p) {
    case 'gemini': return 'Google Gemini'
    case 'groq': return 'Groq Cloud (LPU)'
    case 'claude': return 'Anthropic Claude'
    case 'openai': return 'OpenAI GPT'
    default: return p.toUpperCase()
  }
}

function getProviderClass(p: string) {
  switch (p) {
    case 'gemini': return 'icon-bg-cyan'
    case 'groq': return 'icon-bg-amber'
    case 'claude': return 'icon-bg-primary'
    case 'openai': return 'icon-bg-emerald'
    default: return 'icon-bg-primary'
  }
}
</script>

<style scoped>
.llm-view {
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
    color: var(--text-main);
  }
}

.sub-text {
  font-size: 13px;
  color: var(--text-muted);
}

.active-card {
  padding: 20px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.1) 0%, rgba(6, 182, 212, 0.1) 100%);
  border: 1px solid var(--border-color-hover);
}

.active-badge-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 700;
  color: var(--color-success);
  margin-bottom: 12px;
}

.active-details {
  display: flex;
  align-items: center;
  gap: 16px;
}

.provider-logo-box {
  width: 50px;
  height: 50px;
  border-radius: var(--radius-md);
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
}

.model-name {
  font-size: 14px;
  color: var(--primary);
  font-weight: 600;
}

.tokens-info {
  font-size: 12px;
  color: var(--text-muted);
}

.configs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.col-span-full {
  grid-column: 1 / -1;
}

.loading-box, .empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

.config-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: relative;
}

.config-card-active {
  border-color: var(--primary);
  box-shadow: var(--shadow-glow);
}

.card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.provider-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.provider-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}

.provider-title {
  display: block;
  font-size: 15px;
  color: var(--text-main);
  font-weight: 700;
}

.model-code {
  font-size: 12px;
  color: var(--text-muted);
}

.card-meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  padding: 12px;
  border-radius: var(--radius-sm);
}

.meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--text-secondary);
}

.test-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.alert-success {
  background: var(--color-success-bg);
  color: var(--color-success);
}

.alert-danger {
  background: var(--color-danger-bg);
  color: #f43f5e;
}

.card-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  margin-top: auto;
}

.icon-only-btn {
  padding: 6px 8px;
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
  max-width: 480px;
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
    color: var(--text-main);
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

.modal-footer {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.spin {
  animation: spin 1s linear infinite;
}

.rate-badge {
  font-size: 11px;
  font-family: monospace;
  background: rgba(99, 102, 241, 0.12);
  color: var(--primary);
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid rgba(99, 102, 241, 0.25);
}

.analytics-card {
  margin-top: 14px;
  margin-bottom: 14px;
  padding: 12px;
  background: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}

.analytics-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 700;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 10px;
}

.analytics-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.analytics-item {
  display: flex;
  flex-direction: column;
  background: var(--bg-card);
  padding: 8px 10px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.analytics-label {
  font-size: 10px;
  color: var(--text-muted);
  font-weight: 600;
}

.analytics-val {
  font-size: 14px;
  font-weight: 800;
  margin-top: 2px;
  margin-bottom: 2px;
}

.analytics-sub {
  font-size: 9px;
  color: var(--text-muted);
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
  .analytics-grid {
    grid-template-columns: 1fr;
  }
  .card-actions {
    flex-wrap: wrap;
  }
}
</style>
