import type { ModelMessage, UIMessage } from '@tanstack/ai/client'
import { describe, expect, it } from 'vitest'
import { lastUserText } from './chat-prompt'

function userMessage(content: string): UIMessage {
  return {
    id: 'user-1',
    role: 'user',
    parts: [{ type: 'text', content }],
  }
}

function assistantMessage(content: string): UIMessage {
  return {
    id: 'assistant-1',
    role: 'assistant',
    parts: [{ type: 'text', content }],
  }
}

describe('lastUserText', () => {
  it('returns the newest user message from a UI transcript', () => {
    const messages = [
      userMessage('halo'),
      assistantMessage('Halo juga'),
      userMessage('cek interface'),
    ]

    expect(lastUserText(messages)).toBe('cek interface')
  })

  it('joins every text part of the same message', () => {
    const message: UIMessage = {
      id: 'user-1',
      role: 'user',
      parts: [
        { type: 'text', content: 'bagian satu' },
        { type: 'text', content: 'bagian dua' },
      ],
    }

    expect(lastUserText([message])).toBe('bagian satubagian dua')
  })

  it('reads a model message with string content', () => {
    const messages: ModelMessage[] = [
      { role: 'user', content: 'status router' },
    ]

    expect(lastUserText(messages)).toBe('status router')
  })

  it('reads a model message with content parts', () => {
    const messages: ModelMessage[] = [
      { role: 'user', content: [{ type: 'text', content: 'dari part' }] },
    ]

    expect(lastUserText(messages)).toBe('dari part')
  })

  it('returns an empty string when there is no user message', () => {
    expect(lastUserText([assistantMessage('sendirian')])).toBe('')
    expect(lastUserText([])).toBe('')
  })
})
