export interface WASession {
  id: string
  name: string
  phoneNumber?: string
  jid?: string
  status: string // 'online' | 'connecting' | 'needs_rescan' | 'offline'
  connectedAt?: string
  createdAt?: string
}

export interface SessionStatusPayload {
  session_id: number
  status: string
  qr_code?: string
  jid?: string
  phone_number?: string
  is_logged_in?: boolean
}
