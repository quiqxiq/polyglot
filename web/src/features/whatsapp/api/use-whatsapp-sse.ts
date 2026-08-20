import { useEffect, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { GetWASessionQRResponse, type WASession } from '@/gen/v1/whatsapp_pb'
import type { WARealtimeStatus } from '@/lib/realtime'
import { botKeys } from '@/features/chats/api/keys'
import { waDeviceKeys } from './keys'

// Endpoint /events ter-register publik di backend (tanpa guard JWT), sehingga
// EventSource native bisa dipakai tanpa header Authorization.
const SSE_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

interface SessionStatusPayload {
  session_id: number
  status: string
  qr_code: string
  jid: string
  phone_number: string
  is_logged_in: boolean
}

interface ChatUpdatePayload {
  session_id: number
  chat_jid: string
}

interface ConversationStatusPayload {
  conversation_id: number
  session_id: number
  customer_number: string
  status: string
}

export interface ChatPresencePayload {
  session_id: number
  chat_jid: string
  sender_jid: string
  state: string // "composing" | "paused"
  media: string // "" | "audio"
  is_group: boolean
}

// State typing/recording satu chat — dipakai indikator "mengetik…" di Inbox.
export interface ChatPresence {
  state: 'composing' | 'paused'
  media: string
  senderJid: string
  isGroup: boolean
  // Expiry epoch ms — fallback bila event "paused" terlewat (WhatsApp tidak
  // menjamin event penutup tiba; pruner interval menghapus entri kadaluarsa).
  until: number
}

export type WARealtimeState = {
  status: WARealtimeStatus
  typing: Record<string, ChatPresence>
}

// Key entri typing: `${session_id}:${chat_jid}` — sama di semua halaman.
export function typingKey(sessionId: string | number, chatJid: string): string {
  return `${sessionId}:${chatJid}`
}

// Masa aktif satu indikator typing — fallback bila event "paused" terlewat.
export const TYPING_TTL_MS = 8000

// Pure reducer state typing: menerapkan satu event chat_presence ke map.
// "composing" menambah/me-refresh entri; "paused" menghapusnya. Diekstrak
// agar logika bisa diuji unit tanpa browser/EventSource.
export function applyChatPresence(
  prev: Record<string, ChatPresence>,
  payload: ChatPresencePayload,
  now: number
): Record<string, ChatPresence> {
  const key = typingKey(payload.session_id, payload.chat_jid)
  if (payload.state === 'composing') {
    return {
      ...prev,
      [key]: {
        state: 'composing',
        media: payload.media || '',
        senderJid: payload.sender_jid || '',
        isGroup: Boolean(payload.is_group),
        until: now + TYPING_TTL_MS,
      },
    }
  }
  if (!(key in prev)) return prev
  const next = { ...prev }
  delete next[key]
  return next
}

// Pure reducer: hapus entri typing yang kadaluarsa (indikator tidak boleh
// menempel selamanya bila event penutup tidak pernah tiba).
export function pruneTyping(
  prev: Record<string, ChatPresence>,
  now: number
): Record<string, ChatPresence> {
  let changed = false
  const next: Record<string, ChatPresence> = {}
  for (const [k, v] of Object.entries(prev)) {
    if (v.until > now) next[k] = v
    else changed = true
  }
  return changed ? next : prev
}

// Query key sessions dipakai di dua fitur dengan key space berbeda:
// - /whatsapp (device management)  → waDeviceKeys.sessions()
// - /chats (inbox device dropdown) → botKeys.waSessions()
// Keduanya di-update agar status terasa live di semua halaman.
const SESSION_KEYS = [waDeviceKeys.sessions(), botKeys.waSessions()]

// Prefix key context percakapan — dipakai invalidate prefix saat chat_update
// agar konteks (status bot/agen, token) ikut refresh.
const BOT_CONTEXT_PREFIX = ['bot', 'conversation-context'] as const

/**
 * Konsumen SSE realtime WhatsApp.
 *
 * Backend menyiarkan event ke `GET /events` (internal/adapter/ws/sse_hub.go +
 * internal/driver/whatsapp/event_handler.go):
 *   - `session_status`: status device (online/connecting/needs_rescan/offline,
 *     jid, phone) → di-merge ke query cache sessions (kedua key space) dan —
 *     bila payload membawa QR baru — ke cache QR per-session.
 *   - `chat_update`: mirror chat berubah (pesan masuk/keluar) → invalidate
 *     query chats/messages/conversations/context untuk session & chat tsb,
 *     sehingga Inbox ter-update instan tanpa polling.
 *   - `conversation_status`: status percakapan berubah (take-over/return bot/
 *     close/escalation) → invalidate daftar percakapan & konteksnya.
 *
 * Hook juga mengekspos status koneksi SSE (connecting/open/reconnecting/
 * closed) untuk indikator realtime di UI.
 *
 * Saat koneksi terputus lalu tersambung ulang (auto-reconnect bawaan
 * EventSource), query di-invalidate sekali agar state tersinkron dengan event
 * yang mungkin terlewat.
 */
export function useWARealtimeStream(): WARealtimeState {
  const queryClient = useQueryClient()
  const [sseStatus, setSseStatus] = useState<WARealtimeStatus>('connecting')
  const [typing, setTyping] = useState<Record<string, ChatPresence>>({})

  useEffect(() => {
    const es = new EventSource(`${SSE_BASE}/events`)
    let hasConnectedOnce = false

    const applyStatus = (raw: string) => {
      let payload: SessionStatusPayload
      try {
        payload = JSON.parse(raw) as SessionStatusPayload
      } catch {
        return
      }

      const sessionId = String(payload.session_id)

      // QR baru tersedia — langsung masukkan ke cache QR session tsb.
      // Format sama dengan response GetQRCode (data URI base64), sehingga
      // modal QR langsung menampilkan tanpa menunggu polling.
      if (payload.qr_code) {
        queryClient.setQueryData(
          waDeviceKeys.qr(sessionId),
          new GetWASessionQRResponse({ qrCodeBase64: payload.qr_code })
        )
      }

      for (const key of SESSION_KEYS) {
        queryClient.setQueryData<WASession[]>(key, (old) => {
          if (!old) return old
          return old.map((s) => {
            if (s.id !== sessionId) return s
            // clone() protobuf-es v1 tanpa argumen — mutate field setelahnya.
            const next = s.clone()
            if (payload.status) next.status = payload.status
            if (payload.jid) next.jid = payload.jid
            if (payload.phone_number) next.phoneNumber = payload.phone_number
            return next
          })
        })
      }
    }

    const applyChatUpdate = (raw: string) => {
      let payload: ChatUpdatePayload
      try {
        payload = JSON.parse(raw) as ChatUpdatePayload
      } catch {
        return
      }

      const sessionId = String(payload.session_id)
      const chatJid = payload.chat_jid
      if (!chatJid) return

      // Daftar chat session tsb (preview/urutan/unread) + pesan chat tsb bila
      // sedang dipilih (key lengkap hanya match query yang aktif) + daftar
      // percakapan bot + konteks percakapan yang sedang dibuka + status rate limit.
      queryClient.invalidateQueries({ queryKey: botKeys.chats(sessionId) })
      queryClient.invalidateQueries({
        queryKey: botKeys.chatMessages(sessionId, chatJid),
      })
      queryClient.invalidateQueries({
        queryKey: botKeys.conversations(sessionId),
      })
      queryClient.invalidateQueries({ queryKey: BOT_CONTEXT_PREFIX })
      queryClient.invalidateQueries({
        queryKey: [...botKeys.all, 'rate-limit-status'],
      })
    }

    const applyConversationStatus = (raw: string) => {
      let payload: ConversationStatusPayload
      try {
        payload = JSON.parse(raw) as ConversationStatusPayload
      } catch {
        return
      }

      const sessionId = String(payload.session_id)
      const convId = String(payload.conversation_id)
      // Daftar percakapan session tsb + konteks percakapan yang sedang dibuka
      // (bar status/tombol ambil alih di kanan) refresh instan — termasuk bila
      // perubahan status dilakukan dari perangkat/tab lain.
      queryClient.invalidateQueries({
        queryKey: botKeys.conversations(sessionId),
      })
      if (convId && convId !== '0') {
        queryClient.invalidateQueries({
          queryKey: botKeys.conversation(convId),
        })
        queryClient.invalidateQueries({
          queryKey: botKeys.conversationContext(convId),
        })
      }
      queryClient.invalidateQueries({ queryKey: BOT_CONTEXT_PREFIX })
    }

    const handleSessionStatus = (e: MessageEvent) => applyStatus(e.data)
    const handleChatUpdate = (e: MessageEvent) => applyChatUpdate(e.data)
    const handleConversationStatus = (e: MessageEvent) =>
      applyConversationStatus(e.data)

    // `chat_presence` (typing/recording dari kontak) → map typing per chat.
    const handleChatPresence = (e: MessageEvent) => {
      let payload: ChatPresencePayload
      try {
        payload = JSON.parse(e.data) as ChatPresencePayload
      } catch {
        return
      }
      if (!payload.chat_jid) return
      setTyping((prev) => applyChatPresence(prev, payload, Date.now()))
    }

    // Pruner: hapus entri typing yang kadaluarsa (fallback bila "paused"
    // tidak pernah tiba, mis. kontak berhenti mengetik tanpa event penutup).
    const pruner = setInterval(() => {
      setTyping((prev) => pruneTyping(prev, Date.now()))
    }, 1000)

    es.addEventListener('session_status', handleSessionStatus)
    es.addEventListener('chat_update', handleChatUpdate)
    es.addEventListener('conversation_status', handleConversationStatus)
    es.addEventListener('chat_presence', handleChatPresence)
    es.onopen = () => {
      setSseStatus('open')
      // Open pertama: query awal sudah fetch saat mount. Namun bila fetch awal
      // gagal (mis. backend sempat down) query berstatus error/never-loaded —
      // invalidate agar tidak menunggu refetch manual.
      if (hasConnectedOnce) {
        // Reconnect: backend tidak me-replay state — hanya menyiarkan event
        // baru. Invalidate semua key device (sessions + semua QR per-session)
        // + key sessions chats agar state tersinkron penuh.
        queryClient.invalidateQueries({ queryKey: waDeviceKeys.all })
        queryClient.invalidateQueries({ queryKey: botKeys.waSessions() })
      } else {
        const state = queryClient.getQueryState(waDeviceKeys.sessions())
        if (state?.status === 'error' || state?.dataUpdatedAt === 0) {
          queryClient.invalidateQueries({ queryKey: waDeviceKeys.all })
        }
      }
      hasConnectedOnce = true
    }
    es.onerror = () => {
      // EventSource auto-reconnect; selama percobaan ulang readyState kembali
      // ke CONNECTING. CLOSED berarti fatal/terminated (mis. es.close() lokal
      // atau kegagalan permanen). Bila belum pernah terhubung sama sekali,
      // labelnya 'connecting' bukan 'reconnecting'.
      setSseStatus(
        es.readyState === EventSource.CLOSED
          ? 'closed'
          : hasConnectedOnce
            ? 'reconnecting'
            : 'connecting'
      )
    }

    return () => {
      es.removeEventListener('session_status', handleSessionStatus)
      es.removeEventListener('chat_update', handleChatUpdate)
      es.removeEventListener('conversation_status', handleConversationStatus)
      es.removeEventListener('chat_presence', handleChatPresence)
      clearInterval(pruner)
      es.close()
    }
  }, [queryClient])

  return { status: sseStatus, typing }
}
