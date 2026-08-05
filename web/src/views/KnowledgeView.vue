<template>
  <div class="knowledge-view">
    <!-- View Header -->
    <div class="view-header">
      <div>
        <div class="header-title-badge">
          <h3>Basis Pengetahuan (Grounding RAG)</h3>
          <span class="badge badge-primary-glow">
            <Sparkles class="w-3.5 h-3.5" />
            AI Acuan Resmi
          </span>
        </div>
        <p class="sub-text">
          Kelola dokumen acuan resmi, FAQ, dan SOP pelayan GNET. Data ini digunakan oleh Bot AI untuk menjawab pertanyaan pelanggan tanpa manipulasi (hallucination).
        </p>
      </div>
      <button class="btn btn-primary shadow-glow" @click="openCreateModal">
        <Plus class="w-4 h-4" />
        <span>Tambah Entri Baru</span>
      </button>
    </div>

    <!-- Stats Overview Cards -->
    <div class="stats-row">
      <div class="stat-mini-card glass-panel">
        <div class="stat-icon-wrapper bg-indigo-500/10 text-indigo-500">
          <BookOpen class="w-5 h-5" />
        </div>
        <div>
          <span class="stat-label">Total Entri Dokumen</span>
          <h4 class="stat-value">{{ knowledgeStore.entries.length }}</h4>
        </div>
      </div>

      <div class="stat-mini-card glass-panel">
        <div class="stat-icon-wrapper bg-purple-500/10 text-purple-500">
          <Tag class="w-5 h-5" />
        </div>
        <div>
          <span class="stat-label">Total Kategori (Tag)</span>
          <h4 class="stat-value">{{ knowledgeStore.allTags.length }}</h4>
        </div>
      </div>

      <div class="stat-mini-card glass-panel">
        <div class="stat-icon-wrapper bg-emerald-500/10 text-emerald-500">
          <Check class="w-5 h-5" />
        </div>
        <div>
          <span class="stat-label">Status Grounding</span>
          <h4 class="stat-value text-emerald-500 font-bold">Aktif & Terhubung</h4>
        </div>
      </div>
    </div>

    <!-- Search Bar & Tag Filter Section -->
    <div class="filter-card glass-panel">
      <div class="search-row">
        <div class="search-box">
          <Search class="search-icon" />
          <input
            v-model="knowledgeStore.searchQuery"
            type="text"
            class="form-input search-input"
            placeholder="Cari berdasarkan judul, kata kunci, atau isi dokumen..."
          />
          <button
            v-if="knowledgeStore.searchQuery"
            class="clear-search-btn"
            @click="knowledgeStore.searchQuery = ''"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- Interactive Tag Filter Chips -->
      <div class="tag-filter-section">
        <div class="tag-label">
          <Filter class="w-3.5 h-3.5" />
          <span>Filter Tag:</span>
        </div>
        <div class="tag-chips-container">
          <button
            :class="['tag-chip', { active: knowledgeStore.selectedTag === '' }]"
            @click="knowledgeStore.selectedTag = ''"
          >
            Semua Tag ({{ knowledgeStore.entries.length }})
          </button>

          <button
            v-for="tag in tagCounts"
            :key="tag.name"
            :class="['tag-chip', { active: knowledgeStore.selectedTag === tag.name }]"
            @click="selectTag(tag.name)"
          >
            #{{ tag.name }}
            <span class="chip-count">{{ tag.count }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Knowledge Entries Grid -->
    <div class="knowledge-grid">
      <div v-if="knowledgeStore.loading" class="loading-box glass-panel col-span-full">
        <Loader2 class="w-8 h-8 spin text-indigo-500 mb-2" />
        <span>Memuat basis pengetahuan...</span>
      </div>

      <div v-else-if="knowledgeStore.filteredEntries.length === 0" class="empty-box glass-panel col-span-full">
        <BookOpen class="w-12 h-12 text-slate-400 mb-3" />
        <h4>Tidak Ada Dokumen FAQ Ditemukan</h4>
        <p>Silakan buat entri baru atau atur ulang filter pencarian Anda.</p>
        <button class="btn btn-secondary mt-3" @click="resetFilters">Reset Filter</button>
      </div>

      <!-- Knowledge Cards -->
      <div
        v-for="entry in knowledgeStore.filteredEntries"
        :key="entry.id"
        class="knowledge-card glass-panel"
      >
        <div class="card-top">
          <h4 class="entry-title" @click="openViewModal(entry)">{{ entry.title }}</h4>
          <div class="card-actions">
            <button class="icon-btn" title="Baca Detail Full Markdown" @click="openViewModal(entry)">
              <Eye class="w-4 h-4" />
            </button>
            <button class="icon-btn" title="Edit Dokumen" @click="openEditModal(entry)">
              <Edit2 class="w-4 h-4" />
            </button>
            <button class="icon-btn danger" title="Hapus Dokumen" @click="handleDelete(entry.id)">
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>

        <!-- Markdown Summary Preview -->
        <div class="content-preview" @click="openViewModal(entry)">
          <v-md-preview :text="getShortContent(entry.content)" />
        </div>

        <div class="card-footer">
          <div class="tags-wrapper">
            <span
              v-for="t in getTagList(entry.tags)"
              :key="t"
              class="tag-pill"
              @click.stop="knowledgeStore.selectedTag = t"
            >
              #{{ t }}
            </span>
          </div>
          <span class="entry-date">{{ formatDate(entry.updated_at) }}</span>
        </div>
      </div>
    </div>

    <!-- Create / Edit Modal with v-md-editor -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeEditorModal">
      <div class="modal-card modal-lg glass-panel">
        <div class="modal-header">
          <div class="flex items-center gap-2">
            <BookOpen class="w-5 h-5 text-indigo-500" />
            <h3>{{ editingId ? 'Edit Dokumen Basis Pengetahuan' : 'Tambah Dokumen FAQ Baru' }}</h3>
          </div>
          <button class="close-btn" @click="closeEditorModal">
            <X class="w-5 h-5" />
          </button>
        </div>

        <form @submit.prevent="handleSubmit" class="editor-form">
          <div class="input-group">
            <label class="input-label">Judul Dokumen / Pertanyaan FAQ</label>
            <input
              v-model="formTitle"
              type="text"
              class="form-input text-base"
              placeholder="misal: Prosedur Pembayaran & Rekening Resmi GNET"
              required
            />
          </div>

          <div class="input-group">
            <div class="flex items-center justify-between mb-1">
              <label class="input-label">Konten Dokumen (Format Markdown Rich-Text)</label>
              <span class="text-xs text-indigo-500 font-semibold">Markdown Editor (@kangc/v-md-editor English)</span>
            </div>
            <div class="v-md-editor-container">
              <v-md-editor
                v-model="formContent"
                height="380px"
                placeholder="Tuliskan jawaban resmi, contoh format, list, atau panduan lengkap..."
              />
            </div>
          </div>

          <div class="input-group">
            <label class="input-label">Tag / Kata Kunci Pencarian (Dipisahkan Koma)</label>
            <input
              v-model="formTags"
              type="text"
              class="form-input"
              placeholder="misal: pembayaran, tagihan, bca, mandiri, rekening"
            />
            <div class="tag-preview-pills mt-2" v-if="formTags">
              <span v-for="t in getTagList(formTags)" :key="t" class="tag-pill-preview">
                #{{ t }}
              </span>
            </div>
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeEditorModal">Batal</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              <Loader2 v-if="submitting" class="w-4 h-4 spin" />
              <span>{{ editingId ? 'Simpan Perubahan' : 'Buat Dokumen' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Read Detail Markdown Modal -->
    <div v-if="showViewModal && viewingEntry" class="modal-overlay" @click.self="showViewModal = false">
      <div class="modal-card modal-lg glass-panel">
        <div class="modal-header">
          <div>
            <h3 class="text-lg font-bold text-main">{{ viewingEntry.title }}</h3>
            <div class="flex items-center gap-2 mt-1">
              <span v-for="t in getTagList(viewingEntry.tags)" :key="t" class="tag-pill">#{{ t }}</span>
              <span class="text-xs text-muted ml-2">Diperbarui: {{ formatDate(viewingEntry.updated_at) }}</span>
            </div>
          </div>
          <button class="close-btn" @click="showViewModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="view-content-body">
          <v-md-preview :text="viewingEntry.content" />
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showViewModal = false">Tutup</button>
          <button class="btn btn-primary" @click="editFromViewModal">
            <Edit2 class="w-4 h-4" />
            <span>Edit Dokumen Ini</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <ConfirmModal
      :show="showDeleteModal"
      title="Hapus Dokumen Basis Pengetahuan"
      subtitle="Data acuan bot AI akan dihapus dari sistem."
      message="Apakah Anda yakin ingin menghapus dokumen FAQ ini? Bot AI tidak akan dapat merujuk ke data ini lagi saat menjawab pertanyaan pelanggan."
      confirmText="Ya, Hapus Dokumen"
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="showDeleteModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Plus,
  Search,
  BookOpen,
  Edit2,
  Trash2,
  Loader2,
  X,
  Tag,
  Eye,
  Sparkles,
  Filter,
  Check,
} from 'lucide-vue-next'

import ConfirmModal from '../components/ConfirmModal.vue'
import type { KnowledgeEntry } from '../types'
import { useKnowledgeStore } from '../stores/knowledge'

const knowledgeStore = useKnowledgeStore()

const showModal = ref(false)
const showViewModal = ref(false)
const viewingEntry = ref<KnowledgeEntry | null>(null)

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

const tagCounts = computed(() => {
  const map = new Map<string, number>()
  knowledgeStore.entries.forEach((e) => {
    if (e.tags) {
      e.tags.split(',').forEach((t) => {
        const trimmed = t.trim()
        if (trimmed) {
          map.set(trimmed, (map.get(trimmed) || 0) + 1)
        }
      })
    }
  })
  return Array.from(map.entries()).map(([name, count]) => ({ name, count }))
})

function selectTag(tagName: string) {
  if (knowledgeStore.selectedTag === tagName) {
    knowledgeStore.selectedTag = ''
  } else {
    knowledgeStore.selectedTag = tagName
  }
}

function resetFilters() {
  knowledgeStore.searchQuery = ''
  knowledgeStore.selectedTag = ''
}

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

function openViewModal(entry: KnowledgeEntry) {
  viewingEntry.value = entry
  showViewModal.value = true
}

function editFromViewModal() {
  if (!viewingEntry.value) return
  const entry = viewingEntry.value
  showViewModal.value = false
  openEditModal(entry)
}

function closeEditorModal() {
  showModal.value = false
}

async function handleSubmit() {
  if (!formTitle.value.trim() || !formContent.value.trim()) return
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

function getShortContent(text: string) {
  if (!text) return ''
  if (text.length <= 220) return text
  return text.substring(0, 220) + '...'
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
  gap: 16px;
}

.header-title-badge {
  display: flex;
  align-items: center;
  gap: 12px;

  h3 {
    font-size: 22px;
    font-weight: 800;
    color: var(--text-main);
  }
}

.badge-primary-glow {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  font-size: 11px;
  font-weight: 700;
  border-radius: 9999px;
  background: rgba(99, 102, 241, 0.15);
  color: var(--primary);
  border: 1px solid rgba(99, 102, 241, 0.3);
  box-shadow: 0 0 12px rgba(99, 102, 241, 0.2);
}

.sub-text {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: 4px;
}

.shadow-glow {
  box-shadow: 0 4px 14px rgba(99, 102, 241, 0.35);
}

/* Stats Row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
}

.stat-mini-card {
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 14px;
  border-radius: var(--radius-md);
}

.stat-icon-wrapper {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 500;
}

.stat-value {
  font-size: 18px;
  font-weight: 800;
  color: var(--text-main);
}

/* Filter Card */
.filter-card {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.search-row {
  width: 100%;
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}

.search-icon {
  position: absolute;
  left: 14px;
  width: 18px;
  height: 18px;
  color: var(--text-muted);
}

.search-input {
  width: 100%;
  padding-left: 42px;
  padding-right: 36px;
  height: 42px;
  background: var(--bg-input);
  border-color: var(--border-color);
  font-size: 14px;
}

.clear-search-btn {
  position: absolute;
  right: 12px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;

  &:hover {
    color: var(--text-main);
  }
}

.tag-filter-section {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.tag-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 700;
  color: var(--text-muted);
  margin-top: 4px;
  white-space: nowrap;
}

.tag-chips-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 14px;
  font-size: 12px;
  font-weight: 600;
  border-radius: 9999px;
  background: var(--bg-card);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    background: rgba(99, 102, 241, 0.15);
    color: var(--text-main);
    border-color: var(--primary-light);
  }

  &.active {
    background: linear-gradient(135deg, var(--primary) 0%, #4f46e5 100%);
    color: #ffffff;
    border-color: transparent;
    box-shadow: 0 2px 10px rgba(99, 102, 241, 0.3);
  }
}

.chip-count {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 9999px;
  background: rgba(255, 255, 255, 0.2);
}

/* Grid & Cards */
.knowledge-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 20px;
}

.col-span-full {
  grid-column: 1 / -1;
}

.loading-box,
.empty-box {
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
  transition: transform 0.2s ease, border-color 0.2s ease;

  &:hover {
    transform: translateY(-2px);
    border-color: var(--border-color-hover);
  }
}

.card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.entry-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-main);
  line-height: 1.4;
  cursor: pointer;

  &:hover {
    color: var(--primary);
  }
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
    background: rgba(99, 102, 241, 0.1);
  }

  &.danger:hover {
    color: #f43f5e;
    background: rgba(239, 68, 68, 0.15);
  }
}

.content-preview {
  cursor: pointer;
  max-height: 140px;
  overflow: hidden;
  position: relative;
  font-size: 13px;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
  margin-top: auto;
}

.tags-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag-pill {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
  background: rgba(99, 102, 241, 0.12);
  color: var(--primary);
  cursor: pointer;

  &:hover {
    background: rgba(99, 102, 241, 0.25);
  }
}

.entry-date {
  font-size: 11px;
  color: var(--text-muted);
}

/* Modals */
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
  max-width: 540px;
  padding: 24px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.modal-card.modal-lg {
  max-width: 920px;
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

.editor-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
}

.v-md-editor-container {
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--border-color);
}

.tag-preview-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag-pill-preview {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.view-content-body {
  max-height: 60vh;
  overflow-y: auto;
  padding: 12px;
  background: var(--bg-input);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.modal-footer {
  margin-top: 20px;
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
  }
  .tag-filter-section {
    flex-direction: column;
  }
  .knowledge-grid {
    grid-template-columns: 1fr;
  }
}
</style>
