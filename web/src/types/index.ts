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
  webhook_url?: string
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

