import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { WASession } from '../types'
import {
  listSessionsApi,
  createSessionApi,
  getQRCodeApi,
  getPairingCodeApi,
  toggleBotApi,
  updateWebhookApi,
  logoutSessionApi,
  deleteSessionApi,
  reconnectSessionApi,
} from '../api/client'

export const useSessionStore = defineStore('sessions', () => {
  const sessions = ref<WASession[]>([])
  const loading = ref(false)
  const activeQRCode = ref<string | null>(null)
  const activePairingCode = ref<string | null>(null)
  const activeSessionId = ref<number | null>(null)

  async function fetchSessions() {
    loading.value = true
    try {
      const res = await listSessionsApi()
      sessions.value = res.sessions || []
    } finally {
      loading.value = false
    }
  }

  async function createSession(deviceName: string, phoneNumber?: string, webhookUrl?: string) {
    const res = await createSessionApi(deviceName, phoneNumber, webhookUrl)
    if (res.qr_code) {
      activeQRCode.value = res.qr_code
      activeSessionId.value = res.session.id
    }
    await fetchSessions()
    return res
  }

  async function fetchQRCode(id: number) {
    const res = await getQRCodeApi(id)
    activeQRCode.value = res.qr_code
    activeSessionId.value = id
    return res.qr_code
  }

  async function fetchPairingCode(id: number, phoneNumber: string) {
    const res = await getPairingCodeApi(id, phoneNumber)
    activePairingCode.value = res.pairing_code
    activeSessionId.value = id
    return res.pairing_code
  }

  async function toggleBot(id: number, isEnabled: boolean) {
    const res = await toggleBotApi(id, isEnabled)
    const idx = sessions.value.findIndex((s) => s.id === id)
    if (idx !== -1) {
      sessions.value[idx] = res.session
    }
    return res
  }

  async function logoutSession(id: number) {
    await logoutSessionApi(id)
    await fetchSessions()
  }

  async function deleteSession(id: number) {
    await deleteSessionApi(id)
    sessions.value = sessions.value.filter((s) => s.id !== id)
  }

  async function reconnectSession(id: number) {
    const res = await reconnectSessionApi(id)
    await fetchSessions()
    return res
  }

  function handleStatusUpdate(data: { session_id: number; status: string; qr_code?: string }) {
    const idx = sessions.value.findIndex((s) => s.id === data.session_id)
    if (idx !== -1) {
      sessions.value[idx].status = data.status as any
    }
    if (data.qr_code) {
      activeQRCode.value = data.qr_code
      activeSessionId.value = data.session_id
    }
  }

  return {
    sessions,
    loading,
    activeQRCode,
    activePairingCode,
    activeSessionId,
    fetchSessions,
    createSession,
    fetchQRCode,
    fetchPairingCode,
    toggleBot,
    logoutSession,
    deleteSession,
    reconnectSession,
    handleStatusUpdate,
  }
})
