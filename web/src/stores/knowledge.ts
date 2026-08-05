import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { KnowledgeEntry } from '../types'
import {
  listKnowledgeApi,
  createKnowledgeApi,
  updateKnowledgeApi,
  deleteKnowledgeApi,
} from '../api/client'

export const useKnowledgeStore = defineStore('knowledge', () => {
  const entries = ref<KnowledgeEntry[]>([])
  const loading = ref(false)
  const searchQuery = ref('')
  const selectedTag = ref('')

  const filteredEntries = computed(() => {
    let result = entries.value

    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      result = result.filter(
        (e) =>
          e.title.toLowerCase().includes(q) ||
          e.content.toLowerCase().includes(q) ||
          e.tags.toLowerCase().includes(q)
      )
    }

    if (selectedTag.value) {
      result = result.filter((e) =>
        e.tags.toLowerCase().includes(selectedTag.value.toLowerCase())
      )
    }

    return result
  })

  const allTags = computed(() => {
    const set = new Set<string>()
    entries.value.forEach((e) => {
      if (e.tags) {
        e.tags.split(',').forEach((t) => {
          const trimmed = t.trim()
          if (trimmed) set.add(trimmed)
        })
      }
    })
    return Array.from(set)
  })

  async function fetchKnowledge() {
    loading.value = true
    try {
      const res = await listKnowledgeApi()
      entries.value = res.knowledge_entries || []
    } finally {
      loading.value = false
    }
  }

  async function createKnowledge(data: { title: string; content: string; tags?: string }) {
    const res = await createKnowledgeApi(data)
    await fetchKnowledge()
    return res
  }

  async function updateKnowledge(id: number, data: { title: string; content: string; tags?: string }) {
    const res = await updateKnowledgeApi(id, data)
    await fetchKnowledge()
    return res
  }

  async function deleteKnowledge(id: number) {
    await deleteKnowledgeApi(id)
    entries.value = entries.value.filter((e) => e.id !== id)
  }

  return {
    entries,
    loading,
    searchQuery,
    selectedTag,
    filteredEntries,
    allTags,
    fetchKnowledge,
    createKnowledge,
    updateKnowledge,
    deleteKnowledge,
  }
})
