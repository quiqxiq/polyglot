import { describe, expect, it } from 'vitest'
import {
  applyChatPresence,
  pruneTyping,
  typingKey,
  TYPING_TTL_MS,
  type ChatPresence,
} from './use-whatsapp-sse'

const chatA = '628111111111@s.whatsapp.net'
const sender = '628222222222@s.whatsapp.net'
const now = 1_000_000

function basePresence(
  overrides: Partial<Parameters<typeof applyChatPresence>[1]> = {}
) {
  return {
    session_id: 1,
    chat_jid: chatA,
    sender_jid: sender,
    state: 'composing',
    media: '',
    is_group: false,
    ...overrides,
  }
}

function entry(overrides: Partial<ChatPresence> = {}): ChatPresence {
  return {
    state: 'composing',
    media: '',
    senderJid: sender,
    isGroup: false,
    until: now + TYPING_TTL_MS,
    ...overrides,
  }
}

describe('applyChatPresence', () => {
  it('composing menambahkan entri typing dengan TTL', () => {
    const next = applyChatPresence({}, basePresence(), now)
    expect(next[typingKey(1, chatA)]).toEqual(entry())
  })

  it('composing me-refresh until entri yang sudah ada', () => {
    const prev = { [typingKey(1, chatA)]: entry({ until: now - 5000 }) }
    const next = applyChatPresence(prev, basePresence(), now)
    expect(next[typingKey(1, chatA)]?.until).toBe(now + TYPING_TTL_MS)
  })

  it('composing media audio menandai merekam', () => {
    const next = applyChatPresence({}, basePresence({ media: 'audio' }), now)
    expect(next[typingKey(1, chatA)]?.media).toBe('audio')
  })

  it('paused menghapus entri', () => {
    const prev = { [typingKey(1, chatA)]: entry() }
    const next = applyChatPresence(prev, basePresence({ state: 'paused' }), now)
    expect(next[typingKey(1, chatA)]).toBeUndefined()
  })

  it('paused untuk chat tanpa entri tidak mengubah map (referensi tetap)', () => {
    const prev: Record<string, ChatPresence> = {}
    const next = applyChatPresence(prev, basePresence({ state: 'paused' }), now)
    expect(next).toBe(prev)
  })

  it('group disimpan dengan isGroup true', () => {
    const next = applyChatPresence({}, basePresence({ is_group: true }), now)
    expect(next[typingKey(1, chatA)]?.isGroup).toBe(true)
  })
})

describe('pruneTyping', () => {
  it('menghapus entri kadaluarsa dan mempertahankan yang belum', () => {
    const keyA = typingKey(1, chatA)
    const keyB = typingKey(1, '628333333333@s.whatsapp.net')
    const prev: Record<string, ChatPresence> = {
      [keyA]: entry({ until: 500 }),
      [keyB]: entry({ until: 5000 }),
    }
    const next = pruneTyping(prev, 1000)
    expect(next[keyA]).toBeUndefined()
    expect(next[keyB]).toBeDefined()
  })

  it('tanpa entri kadaluarsa mengembalikan referensi yang sama', () => {
    const prev: Record<string, ChatPresence> = {
      [typingKey(1, chatA)]: entry({ until: 5000 }),
    }
    expect(pruneTyping(prev, 1000)).toBe(prev)
  })
})
