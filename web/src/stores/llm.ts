import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { LLMConfig } from '../types'
import {
  listConfigsApi,
  createConfigApi,
  updateConfigApi,
  activateConfigApi,
  testConfigApi,
  deleteConfigApi,
} from '../api/client'

export const useLLMStore = defineStore('llm', () => {
  const configs = ref<LLMConfig[]>([])
  const loading = ref(false)
  const testingId = ref<number | null>(null)

  const activeConfig = computed(() => configs.value.find((c) => c.is_active))

  async function fetchConfigs() {
    loading.value = true
    try {
      const res = await listConfigsApi()
      configs.value = res.configs || []
    } finally {
      loading.value = false
    }
  }

  async function createConfig(configData: {
    provider: string
    model: string
    api_key: string
    params?: string
    max_output_tokens?: number
    cost_per_1m_input?: number
    cost_per_1m_output?: number
  }) {
    const res = await createConfigApi(configData)
    await fetchConfigs()
    return res
  }

  async function updateConfig(id: number, configData: Partial<LLMConfig> & { api_key?: string }) {
    const res = await updateConfigApi(id, configData)
    await fetchConfigs()
    return res
  }

  async function deleteConfig(id: number) {
    await deleteConfigApi(id)
    configs.value = configs.value.filter((c) => c.id !== id)
  }

  async function activateConfig(id: number) {
    const res = await activateConfigApi(id)
    await fetchConfigs()
    return res
  }

  async function testConfig(id: number) {
    testingId.value = id
    try {
      return await testConfigApi(id)
    } finally {
      testingId.value = null
    }
  }

  return {
    configs,
    loading,
    testingId,
    activeConfig,
    fetchConfigs,
    createConfig,
    updateConfig,
    deleteConfig,
    activateConfig,
    testConfig,
  }
})
