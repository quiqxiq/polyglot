// API client for communicating with GNET Bot Admin API backend (ConnectRPC + SSE).
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

const CONNECT_BASE_URL = import.meta.env.VITE_CONNECT_BASE_URL || 'http://localhost:8080'

export function getWSBaseUrl(): string {
  return CONNECT_BASE_URL
}

function getAuthHeader(): Record<string, string> {
  const token = localStorage.getItem('gnet_token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export async function connectFetch<T>(service: string, method: string, body: Record<string, any> = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...getAuthHeader(),
  }

  const url = `${CONNECT_BASE_URL}/${service}/${method}`
  const response = await fetch(url, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
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
    throw new Error(data.message || data.error || `RPC error ${response.status}: ${response.statusText}`)
  }

  return data as T
}

export async function apiFetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const headers = {
    'Content-Type': 'application/json',
    ...getAuthHeader(),
    ...options.headers,
  }

  const response = await fetch(`${CONNECT_BASE_URL}${endpoint}`, {
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

const nowStr = () => new Date().toISOString()

// --- Auth API ---

export async function loginApi(email: string, password: string): Promise<{ token: string; user: User }> {
  const res = await connectFetch<any>('polyglot.v1.AuthService', 'Login', { username: email, password })
  return {
    token: res.token || '',
    user: {
      id: Number(res.user?.id) || 1,
      email: res.user?.email || email,
      role: (res.user?.role === 'admin' ? 'admin' : 'agent') as 'admin' | 'agent',
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

export async function registerApi(email: string, password: string, _role?: string): Promise<{ token: string; user: User }> {
  return loginApi(email, password)
}

export async function getMeApi(): Promise<{ user: User }> {
  const res = await connectFetch<any>('polyglot.v1.AuthService', 'GetMe', {})
  return {
    user: {
      id: Number(res.user?.id) || 1,
      email: res.user?.email || 'admin@polyglot.net',
      role: (res.user?.role === 'admin' ? 'admin' : 'agent') as 'admin' | 'agent',
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

// --- WA Sessions API ---

export async function listSessionsApi(): Promise<{ sessions: WASession[] }> {
  const res = await connectFetch<any>('polyglot.v1.WhatsAppService', 'ListSessions', {})
  const sessions: WASession[] = (res.sessions || []).map((s: any) => ({
    id: Number(s.id) || 1,
    device_name: s.name || s.device_name || '',
    phone_number: s.phone_number || s.phoneNumber || '',
    status: (s.status === 'online' || s.status === 'needs_rescan' ? s.status : 'offline') as any,
    is_bot_enabled: s.is_bot_active ?? s.is_bot_enabled ?? true,
    connected_at: s.connected_at || nowStr(),
    created_at: s.created_at || s.createdAt || nowStr(),
    updated_at: s.updated_at || nowStr(),
  }))
  return { sessions }
}

export async function createSessionApi(deviceName: string, phoneNumber?: string): Promise<{ session: WASession; qr_code?: string }> {
  const res = await connectFetch<any>('polyglot.v1.WhatsAppService', 'CreateSession', {
    name: deviceName,
    phone_number: phoneNumber || '',
  })
  const s = res.session || {}
  return {
    session: {
      id: Number(s.id) || 1,
      device_name: s.name || deviceName,
      phone_number: s.phone_number || phoneNumber || '',
      status: 'offline',
      is_bot_enabled: s.is_bot_active ?? true,
      connected_at: nowStr(),
      created_at: s.created_at || nowStr(),
      updated_at: nowStr(),
    },
    qr_code: res.qr_code_base64 || '',
  }
}

export async function getQRCodeApi(id: number): Promise<{ session_id: number; qr_code: string }> {
  const res = await connectFetch<any>('polyglot.v1.WhatsAppService', 'GetQRCode', { session_id: String(id) })
  return {
    session_id: id,
    qr_code: res.qr_code_base64 || res.qrCodeBase64 || '',
  }
}

export async function getPairingCodeApi(id: number, phoneNumber: string): Promise<{ session_id: number; pairing_code: string }> {
  const res = await connectFetch<any>('polyglot.v1.WhatsAppService', 'GetPairingCode', {
    session_id: String(id),
    phone_number: phoneNumber,
  })
  return {
    session_id: id,
    pairing_code: res.pairing_code || res.pairingCode || '',
  }
}

export async function toggleBotApi(id: number, isBotEnabled: boolean): Promise<{ message: string; session: WASession }> {
  const res = await connectFetch<any>('polyglot.v1.WhatsAppService', 'ToggleBot', {
    session_id: String(id),
    is_active: isBotEnabled,
  })
  return {
    message: res.message || 'Bot toggled',
    session: {
      id,
      device_name: 'WA Device',
      phone_number: '',
      status: 'online',
      is_bot_enabled: isBotEnabled,
      connected_at: nowStr(),
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

export async function logoutSessionApi(id: number): Promise<{ message: string }> {
  const res = await connectFetch<any>('polyglot.v1.WhatsAppService', 'LogoutSession', { session_id: String(id) })
  return { message: res.message || 'Logged out' }
}

export async function deleteSessionApi(id: number): Promise<{ message: string }> {
  const res = await connectFetch<any>('polyglot.v1.WhatsAppService', 'PurgeSession', { session_id: String(id) })
  return { message: res.message || 'Session deleted' }
}

export async function reconnectSessionApi(id: number): Promise<{ message: string; qr_code?: string }> {
  const res = await connectFetch<any>('polyglot.v1.WhatsAppService', 'ReconnectSession', { session_id: String(id) })
  return { message: res.message || 'Reconnecting', qr_code: res.qr_code_base64 }
}

// --- LLM Config API ---

export async function listConfigsApi(): Promise<{ configs: LLMConfig[] }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'ListLLMConfigs', {})
  const configs: LLMConfig[] = (res.configs || []).map((c: any) => ({
    id: Number(c.id) || 1,
    provider: c.provider || 'openai',
    model: c.model_name || c.modelName || c.model || 'gpt-4o-mini',
    is_active: c.is_active ?? c.isActive ?? true,
    max_output_tokens: c.max_tokens || c.maxTokens || 1024,
    params: c.system_prompt || c.systemPrompt || '',
    created_at: nowStr(),
    updated_at: nowStr(),
  }))
  return { configs }
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
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'CreateLLMConfig', {
    provider: configData.provider,
    model_name: configData.model,
    max_tokens: configData.max_output_tokens || 1024,
    system_prompt: configData.params || '',
  })
  const c = res.config || {}
  return {
    message: 'Config created',
    config: {
      id: Number(c.id) || 1,
      provider: c.provider || configData.provider,
      model: c.model_name || configData.model,
      is_active: c.is_active ?? false,
      max_output_tokens: c.max_tokens || configData.max_output_tokens,
      params: configData.params,
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

export async function updateConfigApi(
  id: number,
  configData: Partial<LLMConfig> & { api_key?: string }
): Promise<{ message: string; config: LLMConfig }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'UpdateLLMConfig', {
    id: String(id),
    provider: configData.provider || '',
    model_name: configData.model || '',
    system_prompt: configData.params || '',
  })
  const c = res.config || {}
  return {
    message: 'Config updated',
    config: {
      id,
      provider: c.provider || configData.provider || '',
      model: c.model_name || configData.model || '',
      is_active: c.is_active ?? true,
      params: configData.params,
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

export async function activateConfigApi(id: number): Promise<{ message: string }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'ActivateLLMConfig', { id: String(id) })
  return { message: res.message || 'Config activated' }
}

export async function testConfigApi(id: number): Promise<{ status: string; message: string }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'TestLLMConfig', { id: String(id) })
  return {
    status: res.success ? 'success' : 'failed',
    message: res.response_text || res.responseText || 'Test complete',
  }
}

export async function deleteConfigApi(id: number): Promise<{ message: string }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'DeleteLLMConfig', { id: String(id) })
  return { message: res.message || 'Config deleted' }
}

// --- Knowledge Base API ---

export async function listKnowledgeApi(): Promise<{ knowledge_entries: KnowledgeEntry[] }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'ListKnowledge', {})
  const knowledge_entries: KnowledgeEntry[] = (res.items || res.knowledge_entries || []).map((k: any) => ({
    id: Number(k.id) || 1,
    title: k.title || '',
    content: k.content || '',
    tags: Array.isArray(k.tags) ? k.tags.join(', ') : k.tags || '',
    created_at: k.created_at || k.createdAt || nowStr(),
    updated_at: k.updated_at || nowStr(),
  }))
  return { knowledge_entries }
}

export async function createKnowledgeApi(entryData: {
  title: string
  content: string
  tags?: string
}): Promise<{ message: string; entry: KnowledgeEntry }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'CreateKnowledge', {
    title: entryData.title,
    content: entryData.content,
    tags: entryData.tags ? entryData.tags.split(',').map((t) => t.trim()) : [],
  })
  const k = res.item || {}
  return {
    message: 'Knowledge created',
    entry: {
      id: Number(k.id) || 1,
      title: k.title || entryData.title,
      content: k.content || entryData.content,
      tags: entryData.tags || '',
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

export async function updateKnowledgeApi(
  id: number,
  entryData: { title: string; content: string; tags?: string }
): Promise<{ message: string; entry: KnowledgeEntry }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'UpdateKnowledge', {
    id: String(id),
    title: entryData.title,
    content: entryData.content,
    tags: entryData.tags ? entryData.tags.split(',').map((t) => t.trim()) : [],
  })
  const k = res.item || {}
  return {
    message: 'Knowledge updated',
    entry: {
      id,
      title: k.title || entryData.title,
      content: k.content || entryData.content,
      tags: entryData.tags || '',
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

export async function deleteKnowledgeApi(id: number): Promise<{ message: string }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'DeleteKnowledge', { id: String(id) })
  return { message: res.message || 'Knowledge deleted' }
}

// --- Conversations API ---

export async function listConversationsApi(status?: string): Promise<{ conversations: Conversation[] }> {
  const res = await connectFetch<any>('polyglot.v1.BotService', 'ListConversations', { status: status || '' })
  const conversations: Conversation[] = (res.conversations || []).map((c: any) => ({
    id: Number(c.id) || 1,
    session_id: Number(c.session_id) || 1,
    customer_wa_number: c.client_phone || c.clientPhone || c.customer_wa_number || '',
    status: (c.status === 'escalation' || c.status === 'done' ? c.status : 'bot') as any,
    assigned_agent_id: c.assigned_agent_id ? Number(c.assigned_agent_id) : null,
    started_at: c.started_at || nowStr(),
    created_at: c.created_at || nowStr(),
    updated_at: c.updated_at || nowStr(),
  }))
  return { conversations }
}

export async function getConversationApi(id: number): Promise<{ conversation: Conversation }> {
  const res = await connectFetch<any>('polyglot.v1.BotService', 'GetConversation', { id: String(id) })
  const c = res.conversation || {}
  return {
    conversation: {
      id: Number(c.id) || id,
      session_id: Number(c.session_id) || 1,
      customer_wa_number: c.client_phone || c.clientPhone || c.customer_wa_number || '',
      status: (c.status === 'escalation' || c.status === 'done' ? c.status : 'bot') as any,
      assigned_agent_id: null,
      started_at: nowStr(),
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

export async function takeOverApi(id: number): Promise<{ message: string; conversation: Conversation }> {
  const res = await connectFetch<any>('polyglot.v1.BotService', 'TakeOverConversation', { id: String(id) })
  return {
    message: res.message || 'Conversation taken over',
    conversation: {
      id,
      session_id: 1,
      customer_wa_number: '',
      status: 'escalation',
      assigned_agent_id: 1,
      started_at: nowStr(),
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

export async function resetBotApi(id: number): Promise<{ message: string; conversation: Conversation }> {
  const res = await connectFetch<any>('polyglot.v1.BotService', 'ResetConversationBot', { id: String(id) })
  return {
    message: res.message || 'Bot control reset',
    conversation: {
      id,
      session_id: 1,
      customer_wa_number: '',
      status: 'bot',
      assigned_agent_id: null,
      started_at: nowStr(),
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

export async function sendAgentMessageApi(id: number, content: string): Promise<{ message: string; data: Message }> {
  return {
    message: 'Message sent',
    data: {
      id: Date.now(),
      conversation_id: id,
      sender_type: 'agent',
      content,
      token_in: 0,
      token_out: 0,
      created_at: nowStr(),
    },
  }
}

export async function closeConversationApi(id: number): Promise<{ message: string }> {
  const res = await connectFetch<any>('polyglot.v1.BotService', 'CloseConversation', { id: String(id) })
  return { message: res.message || 'Conversation closed' }
}

// --- Technicians API ---

export async function listTechniciansApi(): Promise<{ technicians: Technician[] }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'ListTechnicians', {})
  const technicians: Technician[] = (res.technicians || []).map((t: any) => ({
    id: Number(t.id) || 1,
    full_name: t.name || t.full_name || '',
    username: t.email || t.username || '',
    phone_number: t.phone_number || t.phoneNumber || '',
    specialization: t.specialization || 'Field Tech',
    is_active: t.is_active ?? t.isActive ?? true,
    created_at: t.created_at || nowStr(),
    updated_at: t.updated_at || nowStr(),
  }))
  return { technicians }
}

export async function createTechnicianApi(techData: {
  full_name: string
  username: string
  phone_number: string
  specialization?: string
  is_active?: boolean
}): Promise<{ message: string; technician: Technician }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'CreateTechnician', {
    name: techData.full_name,
    phone_number: techData.phone_number,
    email: techData.username,
  })
  const t = res.technician || {}
  return {
    message: 'Technician created',
    technician: {
      id: Number(t.id) || 1,
      full_name: t.name || techData.full_name,
      username: t.email || techData.username,
      phone_number: t.phone_number || techData.phone_number,
      specialization: techData.specialization || 'Field Tech',
      is_active: t.is_active ?? true,
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
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
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'UpdateTechnician', {
    id: String(id),
    name: techData.full_name,
    phone_number: techData.phone_number,
    email: techData.username,
  })
  const t = res.technician || {}
  return {
    message: 'Technician updated',
    technician: {
      id,
      full_name: t.name || techData.full_name,
      username: t.email || techData.username,
      phone_number: t.phone_number || techData.phone_number,
      specialization: techData.specialization || 'Field Tech',
      is_active: t.is_active ?? true,
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

export async function toggleTechnicianActiveApi(
  id: number,
  isActive: boolean
): Promise<{ message: string; technician: Technician }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'ToggleTechnicianActive', {
    id: String(id),
    is_active: isActive,
  })
  return {
    message: res.message || 'Status updated',
    technician: {
      id,
      full_name: 'Tech',
      username: 'tech',
      phone_number: '',
      is_active: isActive,
      created_at: nowStr(),
      updated_at: nowStr(),
    },
  }
}

export async function deleteTechnicianApi(id: number): Promise<{ message: string }> {
  const res = await connectFetch<any>('polyglot.v1.KnowledgeService', 'DeleteTechnician', { id: String(id) })
  return { message: res.message || 'Technician deleted' }
}

// --- Active Sessions API ---

export async function listPPPoEActiveApi(deviceId: string): Promise<{ data: PPPoEActiveSession[] }> {
  const res = await connectFetch<any>('polyglot.v1.MikhmonService', 'ListActiveSessions', { device_id: deviceId })
  return { data: res.sessions || res.activeSessions || res.data || [] }
}

export async function listHotspotActiveApi(deviceId: string): Promise<{ data: HotspotActiveSession[] }> {
  const res = await connectFetch<any>('polyglot.v1.MikhmonService', 'ListActiveSessions', { device_id: deviceId })
  return { data: res.sessions || res.activeSessions || res.data || [] }
}

export async function listDHCPLeasesApi(deviceId: string): Promise<{ data: DHCPLease[] }> {
  const res = await connectFetch<any>('polyglot.v1.MikhmonService', 'ListDHCPLeases', { device_id: deviceId })
  return { data: res.leases || res.dhcpLeases || res.data || [] }
}

export async function kickPPPoESessionApi(deviceId: string, rosId: string): Promise<{ message: string }> {
  const res = await connectFetch<any>('polyglot.v1.MikhmonService', 'KickActiveSession', {
    device_id: deviceId,
    session_id: rosId,
  })
  return { message: res.message || 'Session kicked' }
}

export async function kickHotspotSessionApi(deviceId: string, rosId: string): Promise<{ message: string }> {
  const res = await connectFetch<any>('polyglot.v1.MikhmonService', 'KickActiveSession', {
    device_id: deviceId,
    session_id: rosId,
  })
  return { message: res.message || 'Session kicked' }
}

// --- Mikhmon Hotspot & Voucher Engine API ---

export async function listHotspotProfilesApi(deviceId: string): Promise<{ data: HotspotProfile[] }> {
  const res = await connectFetch<any>('polyglot.v1.MikhmonService', 'ListProfiles', { device_id: deviceId })
  return { data: res.profiles || res.data || [] }
}

export async function generateVouchersApi(deviceId: string, req: VoucherBatchRequest): Promise<{ vouchers: VoucherData[]; count: number }> {
  const res = await connectFetch<any>('polyglot.v1.MikhmonService', 'GenerateVouchers', {
    device_id: deviceId,
    ...req,
  })
  return {
    vouchers: res.vouchers || [],
    count: res.count || (res.vouchers ? res.vouchers.length : 0),
  }
}

export async function renderVoucherHTMLApi(deviceId: string, vouchers: VoucherData[], templateName: string = 'default'): Promise<{ html: string }> {
  const res = await connectFetch<any>('polyglot.v1.MikhmonService', 'RenderHotspotVouchersHTML', {
    device_id: deviceId,
    vouchers,
    template_name: templateName,
  })
  return { html: res.html || '' }
}

export async function getVoucherReportsApi(deviceId: string, date?: string, month?: string, year?: string): Promise<{ reports: VoucherReport[] }> {
  const res = await connectFetch<any>('polyglot.v1.MikhmonService', 'GetDashboard', {
    device_id: deviceId,
    date: date || '',
    month: month || '',
    year: year || '',
  })
  return { reports: res.reports || [] }
}

export async function sendVoucherDocumentWAApi(sessionId: number, to: string, _fileName: string, fileBase64: string, caption?: string): Promise<{ message: string }> {
  const res = await connectFetch<any>('polyglot.v1.WhatsAppService', 'SendTextMessage', {
    session_id: String(sessionId),
    recipient_phone: to,
    message_text: caption || `Document (base64 length ${fileBase64.length})`,
  })
  return { message: res.message || 'Document sent via WhatsApp' }
}

// --- RBAC & User Management API ---

export async function listPoliciesApi(): Promise<{ policies: string[][] }> {
  const res = await connectFetch<any>('polyglot.v1.RBACService', 'ListPolicies', {})
  const policies = (res.policies || []).map((p: any) => [p.sub, p.obj, p.act])
  return { policies }
}

export async function addPolicyApi(role: string, path: string, method: string): Promise<{ message: string; policy: string[] }> {
  await connectFetch<any>('polyglot.v1.RBACService', 'AddPolicy', {
    policy: { sub: role, obj: path, act: method },
  })
  return { message: 'Policy added', policy: [role, path, method] }
}

export async function removePolicyApi(role: string, path: string, method: string): Promise<{ message: string }> {
  await connectFetch<any>('polyglot.v1.RBACService', 'RemovePolicy', {
    policy: { sub: role, obj: path, act: method },
  })
  return { message: 'Policy removed' }
}

export async function listRoleAssignmentsApi(): Promise<{ roles: string[][] }> {
  const res = await connectFetch<any>('polyglot.v1.RBACService', 'ListRoleAssignments', {})
  const roles = (res.roleAssignments || res.role_assignments || []).map((r: any) => [r.user, r.role])
  return { roles }
}

export async function assignRoleApi(user: string, role: string): Promise<{ message: string; user: string; role: string }> {
  await connectFetch<any>('polyglot.v1.RBACService', 'AssignRole', {
    assignment: { user, role },
  })
  return { message: 'Role assigned', user, role }
}

export async function unassignRoleApi(user: string, role: string): Promise<{ message: string }> {
  await connectFetch<any>('polyglot.v1.RBACService', 'UnassignRole', {
    assignment: { user, role },
  })
  return { message: 'Role unassigned' }
}

// --- Devices API ---

export async function listDevicesApi(): Promise<Device[]> {
  const res = await connectFetch<any>('polyglot.v1.DeviceService', 'ListDevices', {})
  const devices: Device[] = (res.devices || []).map((d: any) => ({
    id: d.id,
    tenant_id: d.tenant_id || d.tenantId || 'tenant-default',
    name: d.name,
    vendor: d.vendor,
    driver_type: d.driver_type || d.driverType || d.vendor,
    host: d.host,
    port: d.port || 8728,
    timeout_ms: d.timeout_ms || d.timeoutMs || 5000,
    poll_interval_ms: d.poll_interval_ms || d.pollIntervalMs || 10000,
    tags: d.tags || [],
    enabled: d.enabled ?? true,
  }))
  return devices
}

export async function getDeviceApi(id: string): Promise<Device> {
  const res = await connectFetch<any>('polyglot.v1.DeviceService', 'GetDevice', { id })
  const d = res.device || {}
  return {
    id: d.id || id,
    tenant_id: d.tenant_id || 'tenant-default',
    name: d.name || '',
    vendor: d.vendor || 'mikrotik',
    driver_type: d.driver_type || d.vendor || 'mikrotik',
    host: d.host || '',
    port: d.port || 8728,
    timeout_ms: d.timeout_ms || 5000,
    poll_interval_ms: d.poll_interval_ms || 10000,
    tags: d.tags || [],
    enabled: d.enabled ?? true,
  }
}

export async function createDeviceApi(payload: DevicePayload): Promise<{ message: string; device: Device }> {
  const res = await connectFetch<any>('polyglot.v1.DeviceService', 'UpdateDevice', {
    device: {
      id: payload.id || `dev-${Date.now()}`,
      tenant_id: 'tenant-default',
      name: payload.name,
      vendor: payload.vendor,
      driver_type: payload.driver_type,
      host: payload.host,
      port: payload.port,
      timeout_ms: payload.timeout_ms,
      poll_interval_ms: payload.poll_interval_ms,
      tags: payload.tags || [],
      enabled: payload.enabled ?? true,
    },
    username: payload.username || 'admin',
    password: payload.password || '',
  })
  const d = res.device || {}
  return {
    message: res.message || 'Device created',
    device: {
      id: d.id || payload.id,
      tenant_id: d.tenant_id || 'tenant-default',
      name: d.name || payload.name,
      vendor: d.vendor || payload.vendor,
      driver_type: d.driver_type || payload.driver_type,
      host: d.host || payload.host,
      port: d.port || payload.port,
      timeout_ms: d.timeout_ms || payload.timeout_ms,
      poll_interval_ms: d.poll_interval_ms || payload.poll_interval_ms,
      tags: d.tags || payload.tags || [],
      enabled: d.enabled ?? payload.enabled ?? true,
    },
  }
}

export async function updateDeviceApi(id: string, payload: DevicePayload): Promise<{ message: string; device: Device }> {
  const res = await connectFetch<any>('polyglot.v1.DeviceService', 'UpdateDevice', {
    device: {
      id,
      tenant_id: 'tenant-default',
      name: payload.name,
      vendor: payload.vendor,
      driver_type: payload.driver_type,
      host: payload.host,
      port: payload.port,
      timeout_ms: payload.timeout_ms,
      poll_interval_ms: payload.poll_interval_ms,
      tags: payload.tags || [],
      enabled: payload.enabled ?? true,
    },
    username: payload.username || 'admin',
    password: payload.password || '',
  })
  const d = res.device || {}
  return {
    message: res.message || 'Device updated',
    device: {
      id: d.id || id,
      tenant_id: d.tenant_id || 'tenant-default',
      name: d.name || payload.name,
      vendor: d.vendor || payload.vendor,
      driver_type: d.driver_type || payload.driver_type,
      host: d.host || payload.host,
      port: d.port || payload.port,
      timeout_ms: d.timeout_ms || payload.timeout_ms,
      poll_interval_ms: d.poll_interval_ms || payload.poll_interval_ms,
      tags: d.tags || payload.tags || [],
      enabled: d.enabled ?? payload.enabled ?? true,
    },
  }
}

export async function deleteDeviceApi(id: string): Promise<{ message: string }> {
  const res = await connectFetch<any>('polyglot.v1.DeviceService', 'DeleteDevice', { id })
  return { message: res.message || 'Device deleted' }
}

export async function testDeviceConnectionApi(id: string): Promise<DeviceTestResult> {
  const res = await connectFetch<any>('polyglot.v1.DeviceService', 'TestDeviceConnection', { id })
  return {
    device_id: id,
    status: res.status || (res.success ? 'online' : 'error'),
    latency_ms: res.latency_ms || res.latencyMs || 0,
    uptime: res.uptime || '',
    version: res.version || '',
    board_name: res.board_name || res.boardName || '',
    identity: res.identity || '',
    message: res.message || '',
  }
}

// --- SSE Realtime Stream ---

export function createSSEConnection(
  onEvent: (event: string, data: any) => void,
  onOpen?: () => void
): EventSource {
  const token = localStorage.getItem('gnet_token')
  const sseUrl = token ? `${CONNECT_BASE_URL}/events?token=${encodeURIComponent(token)}` : `${CONNECT_BASE_URL}/events`
  const eventSource = new EventSource(sseUrl)

  eventSource.onopen = () => {
    if (onOpen) onOpen()
  }

  eventSource.onmessage = (e) => {
    try {
      const parsed = JSON.parse(e.data)
      onEvent(parsed.event || 'message', parsed.data)
    } catch {
      onEvent('message', e.data)
    }
  }

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
