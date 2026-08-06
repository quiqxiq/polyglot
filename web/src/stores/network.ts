import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { PPPoEActiveSession, HotspotActiveSession, DHCPLease } from '../types'
import {
  listPPPoEActiveApi,
  listHotspotActiveApi,
  listDHCPLeasesApi,
  kickPPPoESessionApi,
  kickHotspotSessionApi,
} from '../api/client'

export const useNetworkStore = defineStore('network', () => {
  const selectedDeviceId = ref<string>('mtk-test')
  const pppoeActive = ref<PPPoEActiveSession[]>([])
  const hotspotActive = ref<HotspotActiveSession[]>([])
  const dhcpLeases = ref<DHCPLease[]>([])
  const loading = ref<boolean>(false)
  const error = ref<string | null>(null)

  async function fetchPPPoEActive() {
    try {
      loading.value = true
      error.value = null
      const res = await listPPPoEActiveApi(selectedDeviceId.value)
      pppoeActive.value = res.data || []
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch PPPoE active sessions'
    } finally {
      loading.value = false
    }
  }

  async function fetchHotspotActive() {
    try {
      loading.value = true
      error.value = null
      const res = await listHotspotActiveApi(selectedDeviceId.value)
      hotspotActive.value = res.data || []
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch Hotspot active sessions'
    } finally {
      loading.value = false
    }
  }

  async function fetchDHCPLeases() {
    try {
      loading.value = true
      error.value = null
      const res = await listDHCPLeasesApi(selectedDeviceId.value)
      dhcpLeases.value = res.data || []
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch DHCP leases'
    } finally {
      loading.value = false
    }
  }

  async function kickPPPoESession(rosId: string) {
    try {
      await kickPPPoESessionApi(selectedDeviceId.value, rosId)
      pppoeActive.value = pppoeActive.value.filter((s) => s.id !== rosId)
    } catch (e: any) {
      throw new Error(e.message || 'Failed to disconnect PPPoE session')
    }
  }

  async function kickHotspotSession(rosId: string) {
    try {
      await kickHotspotSessionApi(selectedDeviceId.value, rosId)
      hotspotActive.value = hotspotActive.value.filter((s) => s.id !== rosId)
    } catch (e: any) {
      throw new Error(e.message || 'Failed to disconnect Hotspot session')
    }
  }

  async function fetchAll() {
    await Promise.allSettled([fetchPPPoEActive(), fetchHotspotActive(), fetchDHCPLeases()])
  }

  return {
    selectedDeviceId,
    pppoeActive,
    hotspotActive,
    dhcpLeases,
    loading,
    error,
    fetchPPPoEActive,
    fetchHotspotActive,
    fetchDHCPLeases,
    kickPPPoESession,
    kickHotspotSession,
    fetchAll,
  }
})
