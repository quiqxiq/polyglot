<template>
  <div class="knowledge-view">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h3>Basis Pengetahuan (Grounding FAQ)</h3>
        <p class="sub-text">Entri FAQ dan prosedur layanan GNET yang menjadi acuan bot agar tidak mengarang jawaban (§5.3).</p>
      </div>
      <button class="btn btn-primary" @click="openCreateModal">
        <Plus class="w-4 h-4" />
        <span>Tambah Entri FAQ</span>
      </button>
    </div>

    <!-- Search & Filter Bar -->
    <div class="filter-bar glass-panel">
      <div class="search-box">
        <Search class="search-icon" />
        <input
          v-model="knowledgeStore.searchQuery"
          type="text"
          class="form-input search-input"
          placeholder="Cari berdasarkan judul, konten, atau tag..."
        />
      </div>

      <!-- Tag Filters -->
      <div class="tag-filters">
        <button
          :class="['tag-btn', { active: knowledgeStore.selectedTag === '' }]"
          @click="knowledgeStore.selectedTag = ''"
        >
          Semua Tag
        </button>
        <button
          v-for="tag in knowledgeStore.allTags"
          :key="tag"
          :class="['tag-btn', { active: knowledgeStore.selectedTag === tag }]"
          @click="knowledgeStore.selectedTag = tag"
        >
          #{{ tag }}
        </button>
      </div>
    </div>

    <!-- Knowledge Entries List -->
    <div class="knowledge-grid">
      <div v-if="knowledgeStore.loading" class="loading-box glass-panel col-span-full">
        <Loader2 class="w-8 h-8 spin text-indigo-400" />
        <span>Memuat basis pengetahuan...</span>
      </div>

      <div v-else-if="knowledgeStore.filteredEntries.length === 0" class="empty-box glass-panel col-span-full">
        <BookOpen class="w-12 h-12 text-slate-600 mb-2" />
        <h4>Tidak ada entri FAQ ditemukan</h4>
        <p>Silakan buat entri baru atau sesuaikan kata kunci pencarian Anda.</p>
      </div>

      <div
        v-for="entry in knowledgeStore.filteredEntries"
        :key="entry.id"
        class="knowledge-card glass-panel"
      >
        <div class="card-header">
          <h4 class="entry-title">{{ entry.title }}</h4>
          <div class="card-actions">
            <button class="icon-btn" title="Edit Entri" @click="openEditModal(entry)">
              <Edit2 class="w-4 h-4" />
            </button>
            <button class="icon-btn danger" title="Hapus Entri" @click="handleDelete(entry.id)">
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>

        <p class="entry-content">{{ entry.content }}</p>

        <div class="card-footer">
          <div class="tags-container">
            <span
              v-for="t in getTagList(entry.tags)"
              :key="t"
              class="badge badge-primary tag-pill"
            >
              #{{ t }}
            </span>
          </div>
          <span class="text-xs text-muted">{{ formatDate(entry.updated_at) }}</span>
        </div>
      </div>
    </div>

    <!-- Create / Edit Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
      <div class="modal-card glass-panel">
        <div class="modal-header">
          <h3>{{ editingId ? 'Edit Entri FAQ' : 'Tambah Entri FAQ Baru' }}</h3>
          <button class="close-btn" @click="showModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <form @submit.prevent="handleSubmit">
          <div class="input-group">
            <label class="input-label">Judul Pertanyaan / Prosedur</label>
            <input
              v-model="formTitle"
              type="text"
              class="form-input"
              placeholder="mis. Cara Membayar Tagihan Bulanan GNET"
              required
            />
          </div>

          <div class="input-group">
            <label class="input-label">Isi / Jawaban Resmi (Grounding)</label>
            <textarea
              v-model="formContent"
              rows="5"
              class="form-textarea"
              placeholder="Tuliskan penjelasan lengkap dan resmi dari GNET yang akan menjadi rujukan bot..."
              required
            ></textarea>
          </div>

          <div class="input-group">
            <label class="input-label">Tag / Kata Kunci (Dipisahkan koma)</label>
            <input
              v-model="formTags"
              type="text"
              class="form-input"
              placeholder="mis. bayar, tagihan, transfer, bca, mandiri"
            />
            <span class="text-xs text-muted mt-1">Digunakan oleh bot untuk pencocokan kata kunci cepat.</span>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showModal = false">Batal</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              <Loader2 v-if="submitting" class="w-4 h-4 spin" />
              <span>{{ editingId ? 'Simpan Perubahan' : 'Buat Entri' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <ConfirmModal
      :show="showDeleteModal"
      title="Hapus Entri FAQ"
      subtitle="Data rujukan basis pengetahuan bot akan dihapus."
      message="Apakah Anda yakin ingin menghapus entri basis pengetahuan (FAQ) ini? Bot tidak akan dapat merujuk ke data ini lagi."
      confirmText="Ya, Hapus Entri"
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
  Search,
  BookOpen,
  Edit2,
  Trash2,
  Loader2,
  X,
} from 'lucide-vue-next'

import ConfirmModal from '../components/ConfirmModal.vue'
import type { KnowledgeEntry } from '../types'
import { useKnowledgeStore } from '../stores/knowledge'

const knowledgeStore = useKnowledgeStore()

const showModal = ref(false)
const showDeleteModal = ref(false)
const deleteTargetId = ref<number | null>(null)
const deleting = ref(false)

const editingId = ref<number | null>(null)
const formTitle = ref('')
const formContent = ref('')
const formTags = ref('')
const submitting = ref(false)

onMounted(() => {
  knowledgeStore.fetchKnowledge()
})

function openCreateModal() {
  editingId.value = null
  formTitle.value = ''
  formContent.value = ''
  formTags.value = ''
  showModal.value = true
}

function openEditModal(entry: KnowledgeEntry) {
  editingId.value = entry.id
  formTitle.value = entry.title
  formContent.value = entry.content
  formTags.value = entry.tags
  showModal.value = true
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (editingId.value) {
      await knowledgeStore.updateKnowledge(editingId.value, {
        title: formTitle.value,
        content: formContent.value,
        tags: formTags.value,
      })
    } else {
      await knowledgeStore.createKnowledge({
        title: formTitle.value,
        content: formContent.value,
        tags: formTags.value,
      })
    }
    showModal.value = false
  } finally {
    submitting.value = false
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
    await knowledgeStore.deleteKnowledge(deleteTargetId.value)
    showDeleteModal.value = false
    deleteTargetId.value = null
  } finally {
    deleting.value = false
  }
}

function getTagList(tagsStr?: string) {
  if (!tagsStr) return []
  return tagsStr.split(',').map((t) => t.trim()).filter(Boolean)
}

function formatDate(d: string) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}
</script>

<style scoped>
.knowledge-view {
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

.filter-bar {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 14px;
  width: 18px;
  height: 18px;
  color: var(--text-muted);
}

.search-input {
  padding-left: 42px;
}

.tag-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-btn {
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 600;
  border-radius: 9999px;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--text-main);
  }

  &.active {
    background: var(--primary);
    color: #ffffff;
    border-color: transparent;
  }
}

.knowledge-grid {
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

.knowledge-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.entry-title {
  font-size: 15px;
  font-weight: 700;
  line-height: 1.4;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.icon-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  padding: 6px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    color: var(--text-main);
    background: rgba(255, 255, 255, 0.1);
  }

  &.danger:hover {
    color: var(--color-danger);
    background: var(--color-danger-bg);
  }
}

.entry-content {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
  white-space: pre-line;
  flex: 1;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag-pill {
  font-size: 11px;
  padding: 2px 8px;
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
  max-width: 520px;
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
  .tag-filters {
    overflow-x: auto;
    white-space: nowrap;
    padding-bottom: 4px;
  }
  .tag-btn {
    flex-shrink: 0;
  }
  .knowledge-grid {
    grid-template-columns: 1fr;
  }
}
</style>
