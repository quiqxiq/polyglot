import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { PPPoEActiveSession, HotspotActiveSession, DHCPLease } from '../types'
import {
  listPPPoEActiveApi,
  listHotspotActiveApi,
  listDHCPLeasesApi,
  kickPPPoESessionApi,
  kickHotspotSessionApi,
  getWSBaseUrl,
} from '../api/client'

export const useNetworkStore = defineStore('network', () => {
  const selectedDeviceId = ref<string>('')
  const pppoeActive = ref<PPPoEActiveSession[]>([])
  const hotspotActive = ref<HotspotActiveSession[]>([])
  const dhcpLeases = ref<DHCPLease[]>([])
  const loading = ref<boolean>(false)
  const error = ref<string | null>(null)

  async function fetchPPPoEActive() {
    if (!selectedDeviceId.value) {
      pppoeActive.value = []
      return
    }
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
    if (!selectedDeviceId.value) {
      hotspotActive.value = []
      return
    }
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
    if (!selectedDeviceId.value) {
      dhcpLeases.value = []
      return
    }
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
    if (!selectedDeviceId.value) return
    try {
      await kickPPPoESessionApi(selectedDeviceId.value, rosId)
      pppoeActive.value = pppoeActive.value.filter((s) => s.id !== rosId)
    } catch (e: any) {
      throw new Error(e.message || 'Failed to disconnect PPPoE session')
    }
  }

  async function kickHotspotSession(rosId: string) {
    if (!selectedDeviceId.value) return
    try {
      await kickHotspotSessionApi(selectedDeviceId.value, rosId)
      hotspotActive.value = hotspotActive.value.filter((s) => s.id !== rosId)
    } catch (e: any) {
      throw new Error(e.message || 'Failed to disconnect Hotspot session')
    }
  }

  async function fetchAll() {
    if (!selectedDeviceId.value) {
      pppoeActive.value = []
      hotspotActive.value = []
      dhcpLeases.value = []
      return
    }
    await Promise.allSettled([fetchPPPoEActive(), fetchHotspotActive(), fetchDHCPLeases()])
  }

  let hotspotActiveEventSource: EventSource | null = null
  let pppActiveEventSource: EventSource | null = null

  function startHotspotActiveStream() {
    stopHotspotActiveStream()
    if (!selectedDeviceId.value) return

    const streamUrl = `${getWSBaseUrl()}/ws/devices/${selectedDeviceId.value}/mikhmon/hotspot-active`
    hotspotActiveEventSource = new EventSource(streamUrl)

    hotspotActiveEventSource.addEventListener('hotspot_active', (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data)
        if (Array.isArray(data)) {
          hotspotActive.value = data
        }
      } catch (e) {
        console.error('Error parsing hotspot_active SSE frame:', e)
      }
    })
  }

  function stopHotspotActiveStream() {
    if (hotspotActiveEventSource) {
      hotspotActiveEventSource.close()
      hotspotActiveEventSource = null
    }
  }

  function startPPPActiveStream() {
    stopPPPActiveStream()
    if (!selectedDeviceId.value) return

    const streamUrl = `${getWSBaseUrl()}/ws/devices/${selectedDeviceId.value}/mikhmon/ppp-active`
    pppActiveEventSource = new EventSource(streamUrl)

    pppActiveEventSource.addEventListener('ppp_active', (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data)
        if (Array.isArray(data)) {
          pppoeActive.value = data
        }
      } catch (e) {
        console.error('Error parsing ppp_active SSE frame:', e)
      }
    })
  }

  function stopPPPActiveStream() {
    if (pppActiveEventSource) {
      pppActiveEventSource.close()
      pppActiveEventSource = null
    }
  }

  function stopAllStreams() {
    stopHotspotActiveStream()
    stopPPPActiveStream()
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
    startHotspotActiveStream,
    stopHotspotActiveStream,
    startPPPActiveStream,
    stopPPPActiveStream,
    stopAllStreams,
  }
})
