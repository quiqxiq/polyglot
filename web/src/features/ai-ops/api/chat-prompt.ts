import type { ModelMessage, UIMessage } from '@tanstack/ai/client'

export type ChatMessage = UIMessage | ModelMessage

function partText(part: unknown): string {
  if (!part || typeof part !== 'object') return ''
  const value = part as { type?: string; content?: unknown }
  if (value.type === 'text' && typeof value.content === 'string') {
    return value.content
  }
  return ''
}

/**
 * Extracts the prompt of the turn being sent. `useChat` hands the full
 * transcript to the transport, and the backend already owns conversation
 * history, so only the newest user message travels on the wire.
 */
export function lastUserText(messages: ReadonlyArray<ChatMessage>): string {
  for (let i = messages.length - 1; i >= 0; i -= 1) {
    const message = messages[i]
    if (!message || message.role !== 'user') continue

    if ('parts' in message && Array.isArray(message.parts)) {
      const text = message.parts.map(partText).join('').trim()
      if (text) return text
    }

    const content = (message as { content?: unknown }).content
    if (typeof content === 'string') return content.trim()
    if (Array.isArray(content)) {
      return content.map(partText).join('').trim()
    }
  }
  return ''
}
