// TypeScript type definitions matching Go backend domain entities.
// Keep in sync with internal/domain/*.go structs.

/** Matches Go: domain.User */
export interface User {
  id: number
  email: string
  role: 'admin' | 'agent'
  created_at: string
  updated_at: string
}

/** Matches Go: domain.WASession */
export interface WASession {
  id: number
  device_name: string
  phone_number: string
  jid?: string
  status: 'online' | 'offline' | 'needs_rescan'
  is_bot_enabled: boolean
  connected_at: string
  created_at: string
  updated_at: string
}

/** Matches Go: domain.LLMConfig */
export interface LLMConfig {
  id: number
  provider: 'gemini' | 'claude' | 'openai' | 'groq' | string
  model: string
  params?: string
  max_output_tokens?: number
  is_active: boolean
  cost_per_1m_input?: number
  cost_per_1m_output?: number
  total_input_tokens?: number
  total_output_tokens?: number
  total_messages?: number
  total_cost_usd?: number
  total_cost_idr?: number
  created_at: string
  updated_at: string
}

/** Matches Go: domain.Conversation */
export interface Conversation {
  id: number
  session_id: number
  customer_wa_number: string
  status: 'bot' | 'escalation' | 'done'
  assigned_agent_id: number | null
  started_at: string
  created_at: string
  updated_at: string
  messages?: Message[]
}

/** Matches Go: domain.Message */
export interface Message {
  id: number
  conversation_id: number
  sender_type: 'customer' | 'bot' | 'agent'
  content: string
  token_in: number
  token_out: number
  created_at: string
}

/** Matches Go: domain.KnowledgeEntry */
export interface KnowledgeEntry {
  id: number
  title: string
  content: string
  tags: string
  created_at: string
  updated_at: string
}

/** Matches Go: domain.Technician */
export interface Technician {
  id: number
  full_name: string
  username: string
  phone_number: string
  specialization?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

/** PPPoE Active Session entity */
export interface PPPoEActiveSession {
  id: string
  name: string
  service: string
  caller_id: string
  address: string
  uptime: string
  bytes_in?: number
  bytes_out?: number
}

/** Hotspot Active Session entity */
export interface HotspotActiveSession {
  id: string
  user: string
  server: string
  domain?: string
  address: string
  mac_address: string
  uptime: string
  bytes_in?: number
  bytes_out?: number
}

/** DHCP Lease entity */
export interface DHCPLease {
  id: string
  address: string
  mac_address: string
  server: string
  active_address?: string
  active_mac_address?: string
  host_name?: string
  status: string
  dynamic: boolean
  disabled: boolean
}

/** Hotspot User Profile entity */
export interface HotspotProfile {
  id: string
  name: string
  shared_users?: string
  rate_limit?: string
  on_login?: string
  price?: string
  validity?: string
}

/** Voucher Generation Request & Batch */
export interface VoucherBatchRequest {
  qty: number
  profile: string
  server_mode?: string
  user_type?: string
  prefix?: string
  char_len?: number
  char_set?: string
}

export interface VoucherData {
  username: string
  password: string
  profile: string
  price: string
  validity: string
  qr_code?: string
}

export interface VoucherReport {
  id?: string
  date: string
  time?: string
  user: string
  profile: string
  price: number
  comment?: string
}

/** RBAC Policy Rule */
export interface RBACPolicy {
  role: string
  path: string
  method: string
}

/** RBAC Role Assignment */
export interface RBACRoleAssignment {
  user: string
  role: string
}

/** Network Device entity */
export interface Device {
  id: string
  tenant_id?: string
  name: string
  vendor: 'mikrotik' | 'cisco' | 'genieacs' | string
  driver_type: string
  host: string
  port: number
  timeout_ms: number
  poll_interval_ms?: number
  extra?: Record<string, string>
  tags?: string[]
  enabled: boolean
}

/** Payload for creating/updating device */
export interface DevicePayload {
  id: string
  name: string
  vendor: string
  driver_type: string
  host: string
  port: number
  timeout_ms: number
  poll_interval_ms?: number
  extra?: Record<string, string>
  tags?: string[]
  enabled: boolean
  username: string
  password?: string
  cred_extra?: Record<string, string>
}

/** Result of connection test */
export interface DeviceTestResult {
  device_id?: string
  status: string
  message: string
  latency_ms?: number
  identity?: string
  version?: string
  board_name?: string
  uptime?: string
  details?: any
}



