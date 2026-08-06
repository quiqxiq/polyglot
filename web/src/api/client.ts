// API client for communicating with GNET Bot Admin API backend (REST + SSE).
import type {
  WASession,
  LLMConfig,
  Conversation,
  Message,
  KnowledgeEntry,
  Technician,
  User,
  PPPoEActiveSession,
  HotspotActiveSession,
  DHCPLease,
  HotspotProfile,
  VoucherBatchRequest,
  VoucherData,
  VoucherReport,
  Device,
  DevicePayload,
  DeviceTestResult,
} from '../types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

export function getWSBaseUrl(): string {
  return API_BASE_URL.replace('/api/v1', '')
}

function getAuthHeader(): Record<string, string> {
  const token = localStorage.getItem('gnet_token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export async function apiFetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const headers = {
    'Content-Type': 'application/json',
    ...getAuthHeader(),
    ...options.headers,
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  })

  if (response.status === 401) {
    localStorage.removeItem('gnet_token')
    localStorage.removeItem('gnet_user')
    if (!window.location.pathname.includes('/login')) {
      window.location.href = '/login'
    }
  }

  let data: any
  const text = await response.text()
  try {
    data = text ? JSON.parse(text) : {}
  } catch {
    data = { error: text || `HTTP error ${response.status}: ${response.statusText}` }
  }

  if (!response.ok) {
    throw new Error(data.error || `HTTP error ${response.status}: ${response.statusText}`)
  }

  return data as T
}

// --- Auth API ---

export async function loginApi(email: string, password: string): Promise<{ token: string; user: User }> {
  return apiFetch('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export async function registerApi(email: string, password: string, role?: string): Promise<{ token: string; user: User }> {
  return apiFetch('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password, role }),
  })
}

export async function getMeApi(): Promise<{ user: User }> {
  return apiFetch('/auth/me')
}

// --- WA Sessions API (§5.1) ---

export async function listSessionsApi(): Promise<{ sessions: WASession[] }> {
  return apiFetch('/sessions')
}

export async function createSessionApi(deviceName: string, phoneNumber?: string): Promise<{ session: WASession; qr_code?: string }> {
  return apiFetch('/sessions', {
    method: 'POST',
    body: JSON.stringify({ device_name: deviceName, phone_number: phoneNumber }),
  })
}

export async function getQRCodeApi(id: number): Promise<{ session_id: number; qr_code: string }> {
  return apiFetch(`/sessions/${id}/qr`)
}

export async function getPairingCodeApi(id: number, phoneNumber: string): Promise<{ session_id: number; pairing_code: string }> {
  return apiFetch(`/sessions/${id}/pairing-code`, {
    method: 'POST',
    body: JSON.stringify({ phone_number: phoneNumber }),
  })
}

export async function toggleBotApi(id: number, isBotEnabled: boolean): Promise<{ message: string; session: WASession }> {
  return apiFetch(`/sessions/${id}/toggle-bot`, {
    method: 'PUT',
    body: JSON.stringify({ is_bot_enabled: isBotEnabled }),
  })
}

export async function logoutSessionApi(id: number): Promise<{ message: string }> {
  return apiFetch(`/sessions/${id}/logout`, {
    method: 'POST',
  })
}

export async function deleteSessionApi(id: number): Promise<{ message: string }> {
  return apiFetch(`/sessions/${id}`, {
    method: 'DELETE',
  })
}

export async function reconnectSessionApi(id: number): Promise<{ message: string; qr_code?: string }> {
  return apiFetch(`/sessions/${id}/reconnect`, {
    method: 'POST',
  })
}

// --- LLM Config API (§5.2) ---

export async function listConfigsApi(): Promise<{ configs: LLMConfig[] }> {
  return apiFetch('/llm-configs')
}

export async function createConfigApi(configData: {
  provider: string
  model: string
  api_key: string
  params?: string
  max_output_tokens?: number
  cost_per_1m_input?: number
  cost_per_1m_output?: number
}): Promise<{ message: string; config: LLMConfig }> {
  return apiFetch('/llm-configs', {
    method: 'POST',
    body: JSON.stringify(configData),
  })
}

export async function updateConfigApi(
  id: number,
  configData: Partial<LLMConfig> & { api_key?: string }
): Promise<{ message: string; config: LLMConfig }> {
  return apiFetch(`/llm-configs/${id}`, {
    method: 'PUT',
    body: JSON.stringify(configData),
  })
}

export async function activateConfigApi(id: number): Promise<{ message: string }> {
  return apiFetch(`/llm-configs/${id}/activate`, {
    method: 'POST',
  })
}

export async function testConfigApi(id: number): Promise<{ status: string; message: string }> {
  return apiFetch(`/llm-configs/${id}/test`, {
    method: 'POST',
  })
}

export async function deleteConfigApi(id: number): Promise<{ message: string }> {
  return apiFetch(`/llm-configs/${id}`, {
    method: 'DELETE',
  })
}

// --- Knowledge Base API (§5.3) ---

export async function listKnowledgeApi(): Promise<{ knowledge_entries: KnowledgeEntry[] }> {
  return apiFetch('/knowledge')
}

export async function createKnowledgeApi(entryData: {
  title: string
  content: string
  tags?: string
}): Promise<{ message: string; entry: KnowledgeEntry }> {
  return apiFetch('/knowledge', {
    method: 'POST',
    body: JSON.stringify(entryData),
  })
}

export async function updateKnowledgeApi(
  id: number,
  entryData: { title: string; content: string; tags?: string }
): Promise<{ message: string; entry: KnowledgeEntry }> {
  return apiFetch(`/knowledge/${id}`, {
    method: 'PUT',
    body: JSON.stringify(entryData),
  })
}

export async function deleteKnowledgeApi(id: number): Promise<{ message: string }> {
  return apiFetch(`/knowledge/${id}`, {
    method: 'DELETE',
  })
}

// --- Conversations API (§5.4) ---

export async function listConversationsApi(status?: string): Promise<{ conversations: Conversation[] }> {
  const query = status ? `?status=${encodeURIComponent(status)}` : ''
  return apiFetch(`/conversations${query}`)
}

export async function getConversationApi(id: number): Promise<{ conversation: Conversation }> {
  return apiFetch(`/conversations/${id}`)
}

export async function takeOverApi(id: number): Promise<{ message: string; conversation: Conversation }> {
  return apiFetch(`/conversations/${id}/take-over`, {
    method: 'POST',
  })
}

export async function resetBotApi(id: number): Promise<{ message: string; conversation: Conversation }> {
  return apiFetch(`/conversations/${id}/reset-bot`, {
    method: 'POST',
  })
}

export async function sendAgentMessageApi(id: number, content: string): Promise<{ message: string; data: Message }> {
  return apiFetch(`/conversations/${id}/messages`, {
    method: 'POST',
    body: JSON.stringify({ content }),
  })
}

export async function closeConversationApi(id: number): Promise<{ message: string }> {
  return apiFetch(`/conversations/${id}/close`, {
    method: 'POST',
  })
}

// --- Technicians API ---

export async function listTechniciansApi(): Promise<{ technicians: Technician[] }> {
  return apiFetch('/technicians')
}

export async function createTechnicianApi(techData: {
  full_name: string
  username: string
  phone_number: string
  specialization?: string
  is_active?: boolean
}): Promise<{ message: string; technician: Technician }> {
  return apiFetch('/technicians', {
    method: 'POST',
    body: JSON.stringify(techData),
  })
}

export async function updateTechnicianApi(
  id: number,
  techData: {
    full_name: string
    username: string
    phone_number: string
    specialization?: string
    is_active?: boolean
  }
): Promise<{ message: string; technician: Technician }> {
  return apiFetch(`/technicians/${id}`, {
    method: 'PUT',
    body: JSON.stringify(techData),
  })
}

export async function toggleTechnicianActiveApi(
  id: number,
  isActive: boolean
): Promise<{ message: string; technician: Technician }> {
  return apiFetch(`/technicians/${id}/toggle-active`, {
    method: 'PUT',
    body: JSON.stringify({ is_active: isActive }),
  })
}

export async function deleteTechnicianApi(id: number): Promise<{ message: string }> {
  return apiFetch(`/technicians/${id}`, {
    method: 'DELETE',
  })
}

// --- Active Sessions API ---

export async function listPPPoEActiveApi(deviceId: string): Promise<{ data: PPPoEActiveSession[] }> {
  return apiFetch(`/devices/${encodeURIComponent(deviceId)}/mikhmon/ppp/active`)
}

export async function listHotspotActiveApi(deviceId: string): Promise<{ data: HotspotActiveSession[] }> {
  return apiFetch(`/devices/${encodeURIComponent(deviceId)}/mikhmon/hotspot/active`)
}

export async function listDHCPLeasesApi(deviceId: string): Promise<{ data: DHCPLease[] }> {
  return apiFetch(`/devices/${encodeURIComponent(deviceId)}/mikhmon/dhcp/leases`)
}

export async function kickPPPoESessionApi(deviceId: string, rosId: string): Promise<{ message: string }> {
  return apiFetch(`/devices/${encodeURIComponent(deviceId)}/mikhmon/ppp/active/${rosId}`, {
    method: 'DELETE',
  })
}

export async function kickHotspotSessionApi(deviceId: string, rosId: string): Promise<{ message: string }> {
  return apiFetch(`/devices/${encodeURIComponent(deviceId)}/mikhmon/hotspot/active/${rosId}`, {
    method: 'DELETE',
  })
}

// --- Mikhmon Hotspot & Voucher Engine API ---

export async function listHotspotProfilesApi(deviceId: string): Promise<{ data: HotspotProfile[] }> {
  return apiFetch(`/devices/${encodeURIComponent(deviceId)}/mikhmon/profiles`)
}

export async function generateVouchersApi(deviceId: string, req: VoucherBatchRequest): Promise<{ vouchers: VoucherData[]; count: number }> {
  return apiFetch(`/devices/${encodeURIComponent(deviceId)}/mikhmon/vouchers/generate`, {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

export async function renderVoucherHTMLApi(deviceId: string, vouchers: VoucherData[], templateName: string = 'default'): Promise<{ html: string }> {
  return apiFetch(`/devices/${encodeURIComponent(deviceId)}/mikhmon/vouchers/render`, {
    method: 'POST',
    body: JSON.stringify({ vouchers, template_name: templateName }),
  })
}

export async function getVoucherReportsApi(deviceId: string, date?: string, month?: string, year?: string): Promise<{ reports: VoucherReport[] }> {
  const params = new URLSearchParams()
  if (date) params.append('date', date)
  if (month) params.append('month', month)
  if (year) params.append('year', year)
  const queryString = params.toString() ? `?${params.toString()}` : ''
  return apiFetch(`/devices/${encodeURIComponent(deviceId)}/mikhmon/reports${queryString}`)
}

export async function sendVoucherDocumentWAApi(sessionId: number, to: string, fileName: string, fileBase64: string, caption?: string): Promise<{ message: string }> {
  return apiFetch(`/sessions/${sessionId}/send-document`, {
    method: 'POST',
    body: JSON.stringify({
      to,
      file_name: fileName,
      file_base64: fileBase64,
      content_type: 'text/html',
      caption,
    }),
  })
}

// --- RBAC & User Management API ---

export async function listPoliciesApi(): Promise<{ policies: string[][] }> {
  return apiFetch('/rbac/policies')
}

export async function addPolicyApi(role: string, path: string, method: string): Promise<{ message: string; policy: string[] }> {
  return apiFetch('/rbac/policies', {
    method: 'POST',
    body: JSON.stringify({ role, path, method }),
  })
}

export async function removePolicyApi(role: string, path: string, method: string): Promise<{ message: string }> {
  return apiFetch('/rbac/policies', {
    method: 'DELETE',
    body: JSON.stringify({ role, path, method }),
  })
}

export async function listRoleAssignmentsApi(): Promise<{ roles: string[][] }> {
  return apiFetch('/rbac/roles')
}

export async function assignRoleApi(user: string, role: string): Promise<{ message: string; user: string; role: string }> {
  return apiFetch('/rbac/roles/assign', {
    method: 'POST',
    body: JSON.stringify({ user, role }),
  })
}

export async function unassignRoleApi(user: string, role: string): Promise<{ message: string }> {
  return apiFetch('/rbac/roles/assign', {
    method: 'DELETE',
    body: JSON.stringify({ user, role }),
  })
}

// --- Devices API ---

export async function listDevicesApi(): Promise<Device[]> {
  const res = await apiFetch<any>('/devices')
  if (Array.isArray(res)) return res
  if (res && Array.isArray(res.devices)) return res.devices
  if (res && Array.isArray(res.data)) return res.data
  return []
}

export async function getDeviceApi(id: string): Promise<Device> {
  return apiFetch(`/devices/${id}`)
}

export async function createDeviceApi(payload: DevicePayload): Promise<{ message: string; device: Device }> {
  return apiFetch('/devices', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export async function updateDeviceApi(id: string, payload: DevicePayload): Promise<{ message: string; device: Device }> {
  return apiFetch(`/devices/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export async function deleteDeviceApi(id: string): Promise<{ message: string }> {
  return apiFetch(`/devices/${id}`, {
    method: 'DELETE',
  })
}

export async function testDeviceConnectionApi(id: string): Promise<DeviceTestResult> {
  return apiFetch(`/devices/${id}/test`, {
    method: 'POST',
  })
}

// --- SSE Realtime Stream ---

export function createSSEConnection(
  onEvent: (event: string, data: any) => void,
  onOpen?: () => void
): EventSource {
  const token = localStorage.getItem('gnet_token')
  const sseUrl = token ? `${API_BASE_URL}/events?token=${encodeURIComponent(token)}` : `${API_BASE_URL}/events`
  const eventSource = new EventSource(sseUrl)

  eventSource.onopen = () => {
    if (onOpen) onOpen()
  }

  eventSource.onmessage = (e) => {
    try {
      const parsed = JSON.parse(e.data)
      onEvent(parsed.event || 'message', parsed.data)
    } catch {
      // Raw string data
      onEvent('message', e.data)
    }
  }

  // Handle specific named events from SSEHub
  const eventTypes = ['new_message', 'session_status', 'conversation_update', 'rate_limit_alert', 'active_sessions_update']
  eventTypes.forEach((type) => {
    eventSource.addEventListener(type, (e: MessageEvent) => {
      try {
        const parsed = JSON.parse(e.data)
        onEvent(type, parsed)
      } catch {
        onEvent(type, e.data)
      }
    })
  })

  return eventSource
}
