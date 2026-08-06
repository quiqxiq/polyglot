import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Device, DevicePayload } from '../types'
import {
  listDevicesApi,
  createDeviceApi,
  updateDeviceApi,
  deleteDeviceApi,
  testDeviceConnectionApi,
} from '../api/client'

export const useDeviceStore = defineStore('devices', () => {
  const devices = ref<Device[]>([])
  const loading = ref<boolean>(false)
  const testingId = ref<string | null>(null)
  const testResults = ref<
    Record<
      string,
      {
        success: boolean
        message: string
        latency_ms?: number
        identity?: string
        version?: string
        board_name?: string
        uptime?: string
      }
    >
  >({})
  const error = ref<string | null>(null)

  async function fetchDevices() {
    try {
      loading.value = true
      error.value = null
      const res = await listDevicesApi()
      devices.value = Array.isArray(res) ? res : []
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch devices'
    } finally {
      loading.value = false
    }
  }

  async function createDevice(payload: DevicePayload) {
    try {
      loading.value = true
      error.value = null
      const res = await createDeviceApi(payload)
      await fetchDevices()
      return res
    } catch (e: any) {
      error.value = e.message || 'Failed to create device'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function updateDevice(id: string, payload: DevicePayload) {
    try {
      loading.value = true
      error.value = null
      const res = await updateDeviceApi(id, payload)
      await fetchDevices()
      return res
    } catch (e: any) {
      error.value = e.message || 'Failed to update device'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteDevice(id: string) {
    try {
      loading.value = true
      error.value = null
      await deleteDeviceApi(id)
      devices.value = devices.value.filter((d) => d.id !== id)
    } catch (e: any) {
      error.value = e.message || 'Failed to delete device'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function testConnection(id: string) {
    try {
      testingId.value = id
      const res = await testDeviceConnectionApi(id)
      testResults.value[id] = {
        success: res.status === 'connected' || res.status === 'ok' || res.status === 'success',
        message: res.message || 'Connection test successful',
        latency_ms: res.latency_ms,
        identity: res.identity,
        version: res.version,
        board_name: res.board_name,
        uptime: res.uptime,
      }
      return res
    } catch (e: any) {
      testResults.value[id] = {
        success: false,
        message: e.message || 'Connection test failed',
      }
      throw e
    } finally {
      testingId.value = null
    }
  }

  async function testAllDevices() {
    for (const dev of devices.value) {
      if (dev.enabled) {
        testConnection(dev.id).catch(() => {})
      }
    }
  }

  return {
    devices,
    loading,
    testingId,
    testResults,
    error,
    fetchDevices,
    createDevice,
    updateDevice,
    deleteDevice,
    testConnection,
    testAllDevices,
  }
})
