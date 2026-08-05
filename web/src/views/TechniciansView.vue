<template>
  <div class="technicians-view">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h3>Tim Teknisi Gangguan</h3>
        <p class="sub-text">Kelola daftar teknisi lapangan untuk menerima notifikasi WhatsApp saat pelanggan melaporkan gangguan.</p>
      </div>
      <button class="btn btn-primary" @click="openCreateModal">
        <UserPlus class="w-4 h-4" />
        <span>Tambah Teknisi</span>
      </button>
    </div>

    <!-- Summary & Search Bar -->
    <div class="filter-bar glass-panel">
      <div class="search-box">
        <Search class="search-icon" />
        <input
          v-model="techStore.searchQuery"
          type="text"
          class="form-input search-input"
          placeholder="Cari berdasarkan nama, username, no. WA, atau spesialisasi..."
        />
      </div>

      <div class="filter-actions">
        <label class="toggle-label">
          <input
            v-model="techStore.filterActiveOnly"
            type="checkbox"
            class="toggle-checkbox"
          />
          <span class="toggle-text">Hanya Aktif ({{ techStore.activeTechniciansCount }})</span>
        </label>
      </div>
    </div>

    <!-- Technicians Grid / List -->
    <div class="technicians-grid">
      <div v-if="techStore.loading" class="loading-box glass-panel col-span-full">
        <Loader2 class="w-8 h-8 spin text-indigo-400" />
        <span>Memuat data teknisi...</span>
      </div>

      <div v-else-if="techStore.filteredTechnicians.length === 0" class="empty-box glass-panel col-span-full">
        <Wrench class="w-12 h-12 text-slate-600 mb-2" />
        <h4>Belum ada data teknisi</h4>
        <p>Silakan tambah teknisi baru agar bot dapat meneruskan laporan gangguan pelanggan.</p>
      </div>

      <div
        v-for="tech in techStore.filteredTechnicians"
        :key="tech.id"
        :class="['tech-card glass-panel', { 'inactive-card': !tech.is_active }]"
      >
        <div class="tech-card-header">
          <div class="tech-avatar">
            <UserCheck v-if="tech.is_active" class="w-5 h-5 text-emerald-400" />
            <UserX v-else class="w-5 h-5 text-slate-500" />
          </div>
          <div class="tech-main-info">
            <h4 class="tech-name">{{ tech.full_name }}</h4>
            <span class="tech-username">@{{ tech.username }}</span>
          </div>
          <span :class="['badge', tech.is_active ? 'badge-success' : 'badge-secondary']">
            {{ tech.is_active ? 'Aktif' : 'Non-Aktif' }}
          </span>
        </div>

        <div class="tech-card-body">
          <div class="info-row">
            <Phone class="w-4 h-4 text-slate-400" />
            <span class="info-text">{{ tech.phone_number }}</span>
          </div>
          <div v-if="tech.specialization" class="info-row">
            <Wrench class="w-4 h-4 text-slate-400" />
            <span class="info-text">{{ tech.specialization }}</span>
          </div>
        </div>

        <div class="tech-card-footer">
          <label class="switch">
            <input
              type="checkbox"
              :checked="tech.is_active"
              @change="handleToggleActive(tech)"
            />
            <span class="slider round"></span>
          </label>
          <span class="switch-label">{{ tech.is_active ? 'Menerima Alert' : 'Di-Pause' }}</span>

          <div class="card-actions">
            <button class="action-btn" title="Edit Teknisi" @click="openEditModal(tech)">
              <Edit2 class="w-4 h-4 text-slate-300" />
            </button>
            <button class="action-btn danger" title="Hapus Teknisi" @click="handleDelete(tech)">
              <Trash2 class="w-4 h-4 text-rose-400" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal Form (Tambah / Edit) -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-content glass-panel">
        <div class="modal-header">
          <h4>{{ isEditing ? 'Edit Data Teknisi' : 'Tambah Teknisi Baru' }}</h4>
          <button class="close-btn" @click="closeModal">&times;</button>
        </div>

        <form @submit.prevent="handleSubmit" class="modal-form">
          <div class="form-group">
            <label class="form-label">Nama Lengkap *</label>
            <input
              v-model="form.full_name"
              type="text"
              class="form-input"
              placeholder="Contoh: Ahmad Fauzi"
              required
            />
          </div>

          <div class="form-group">
            <label class="form-label">Username (Internal) *</label>
            <input
              v-model="form.username"
              type="text"
              class="form-input"
              placeholder="Contoh: ahmad_fauzi"
              required
            />
          </div>

          <div class="form-group">
            <label class="form-label">Nomor WhatsApp *</label>
            <input
              v-model="form.phone_number"
              type="text"
              class="form-input"
              placeholder="Contoh: 6281234567890"
              required
            />
            <span class="form-hint">Format e.164 tanpa tanda + (dimulai dengan 62).</span>
          </div>

          <div class="form-group">
            <label class="form-label">Spesialisasi / Area (Opsional)</label>
            <input
              v-model="form.specialization"
              type="text"
              class="form-input"
              placeholder="Contoh: Fiber Optik & Modems"
            />
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="form.is_active" type="checkbox" />
              <span>Status Aktif (Langsung menerima notifikasi WA gangguan)</span>
            </label>
          </div>

          <div v-if="errorMsg" class="error-banner">
            {{ errorMsg }}
          </div>

          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="closeModal">Batal</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              <Loader2 v-if="submitting" class="w-4 h-4 spin" />
              <span>{{ isEditing ? 'Simpan Perubahan' : 'Tambah Teknisi' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
  UserPlus,
  Search,
  Wrench,
  Loader2,
  Phone,
  Edit2,
  Trash2,
  UserCheck,
  UserX,
} from 'lucide-vue-next'
import { useTechniciansStore } from '../stores/technicians'
import type { Technician } from '../types'

const techStore = useTechniciansStore()

const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)
const errorMsg = ref('')

const form = reactive({
  full_name: '',
  username: '',
  phone_number: '',
  specialization: '',
  is_active: true,
})

onMounted(() => {
  techStore.fetchTechnicians()
})

function resetForm() {
  form.full_name = ''
  form.username = ''
  form.phone_number = ''
  form.specialization = ''
  form.is_active = true
  errorMsg.value = ''
}

function openCreateModal() {
  resetForm()
  isEditing.value = false
  editingId.value = null
  showModal.value = true
}

function openEditModal(tech: Technician) {
  resetForm()
  isEditing.value = true
  editingId.value = tech.id
  form.full_name = tech.full_name
  form.username = tech.username
  form.phone_number = tech.phone_number
  form.specialization = tech.specialization || ''
  form.is_active = tech.is_active
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

async function handleSubmit() {
  submitting.value = true
  errorMsg.value = ''
  try {
    if (isEditing.value && editingId.value) {
      await techStore.updateTechnician(editingId.value, form)
    } else {
      await techStore.createTechnician(form)
    }
    closeModal()
  } catch (err: any) {
    errorMsg.value = err.message || 'Gagal menyimpan data teknisi'
  } finally {
    submitting.value = false
  }
}

async function handleToggleActive(tech: Technician) {
  try {
    await techStore.toggleActive(tech.id, !tech.is_active)
  } catch (err: any) {
    alert(err.message || 'Gagal memperbarui status teknisi')
  }
}

async function handleDelete(tech: Technician) {
  if (confirm(`Apakah Anda yakin ingin menghapus teknisi ${tech.full_name}?`)) {
    try {
      await techStore.deleteTechnician(tech.id)
    } catch (err: any) {
      alert(err.message || 'Gagal menghapus teknisi')
    }
  }
}
</script>

<style scoped>
.technicians-view {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.view-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.sub-text {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: 4px;
}

.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 18px;
}

.search-box {
  position: relative;
  flex: 1;
  max-width: 420px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  width: 18px;
  height: 18px;
  color: var(--text-muted);
}

.search-input {
  padding-left: 38px;
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
}

.toggle-checkbox {
  accent-color: var(--primary);
  width: 16px;
  height: 16px;
}

.technicians-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.tech-card {
  padding: 18px;
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  gap: 14px;
  transition: all 0.2s ease;
}

.tech-card:hover {
  transform: translateY(-2px);
  border-color: rgba(99, 102, 241, 0.4);
}

.inactive-card {
  opacity: 0.65;
  background: rgba(15, 23, 42, 0.4);
}

.tech-card-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.tech-avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.05);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border-color);
}

.tech-main-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.tech-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-main);
  margin: 0;
}

.tech-username {
  font-size: 12px;
  color: var(--text-muted);
}

.tech-card-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 12px;
  background: rgba(0, 0, 0, 0.2);
  border-radius: var(--radius-sm);
}

.info-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-text {
  font-size: 13px;
  color: var(--text-secondary);
}

.tech-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 4px;
}

.switch {
  position: relative;
  display: inline-block;
  width: 36px;
  height: 20px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #334155;
  transition: 0.3s;
}

.slider:before {
  position: absolute;
  content: "";
  height: 14px;
  width: 14px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
}

input:checked + .slider {
  background-color: #10b981;
}

input:checked + .slider:before {
  transform: translateX(16px);
}

.slider.round {
  border-radius: 20px;
}

.slider.round:before {
  border-radius: 50%;
}

.switch-label {
  font-size: 12px;
  color: var(--text-muted);
  margin-left: 8px;
  flex: 1;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.action-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--border-color);
  padding: 6px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.15);
}

.action-btn.danger:hover {
  background: rgba(244, 63, 94, 0.2);
  border-color: rgba(244, 63, 94, 0.4);
}

.loading-box, .empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px;
  text-align: center;
  color: var(--text-muted);
}

/* Modal Styling */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.modal-content {
  width: 100%;
  max-width: 480px;
  padding: 24px;
  border-radius: var(--radius-lg);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.modal-header h4 {
  font-size: 18px;
  font-weight: 700;
  margin: 0;
}

.close-btn {
  background: transparent;
  border: none;
  font-size: 24px;
  color: var(--text-muted);
  cursor: pointer;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.form-hint {
  font-size: 11px;
  color: var(--text-muted);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;

  input {
    accent-color: var(--primary);
  }
}

.error-banner {
  background: rgba(244, 63, 94, 0.15);
  border: 1px solid rgba(244, 63, 94, 0.3);
  color: #fb7185;
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  font-size: 13px;
}

.modal-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
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
  .filter-bar {
    flex-direction: column;
    align-items: stretch;
  }
  .search-box {
    max-width: none;
  }
  .technicians-grid {
    grid-template-columns: 1fr;
  }
}
</style>
