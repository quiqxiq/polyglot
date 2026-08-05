import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Technician } from '../types'
import {
  listTechniciansApi,
  createTechnicianApi,
  updateTechnicianApi,
  toggleTechnicianActiveApi,
  deleteTechnicianApi,
} from '../api/client'

export const useTechniciansStore = defineStore('technicians', () => {
  const technicians = ref<Technician[]>([])
  const loading = ref(false)
  const searchQuery = ref('')
  const filterActiveOnly = ref(false)

  const filteredTechnicians = computed(() => {
    let result = technicians.value

    if (filterActiveOnly.value) {
      result = result.filter((t) => t.is_active)
    }

    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      result = result.filter(
        (t) =>
          t.full_name.toLowerCase().includes(q) ||
          t.username.toLowerCase().includes(q) ||
          t.phone_number.includes(q) ||
          (t.specialization && t.specialization.toLowerCase().includes(q))
      )
    }

    return result
  })

  const activeTechniciansCount = computed(() => {
    return technicians.value.filter((t) => t.is_active).length
  })

  async function fetchTechnicians() {
    loading.value = true
    try {
      const res = await listTechniciansApi()
      technicians.value = res.technicians || []
    } finally {
      loading.value = false
    }
  }

  async function createTechnician(data: {
    full_name: string
    username: string
    phone_number: string
    specialization?: string
    is_active?: boolean
  }) {
    const res = await createTechnicianApi(data)
    await fetchTechnicians()
    return res
  }

  async function updateTechnician(
    id: number,
    data: {
      full_name: string
      username: string
      phone_number: string
      specialization?: string
      is_active?: boolean
    }
  ) {
    const res = await updateTechnicianApi(id, data)
    await fetchTechnicians()
    return res
  }

  async function toggleActive(id: number, isActive: boolean) {
    const res = await toggleTechnicianActiveApi(id, isActive)
    const found = technicians.value.find((t) => t.id === id)
    if (found) {
      found.is_active = isActive
    }
    return res
  }

  async function deleteTechnician(id: number) {
    await deleteTechnicianApi(id)
    technicians.value = technicians.value.filter((t) => t.id !== id)
  }

  return {
    technicians,
    loading,
    searchQuery,
    filterActiveOnly,
    filteredTechnicians,
    activeTechniciansCount,
    fetchTechnicians,
    createTechnician,
    updateTechnician,
    toggleActive,
    deleteTechnician,
  }
})
