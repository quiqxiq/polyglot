import {
  ChatEvent,
  DoneEvent,
  TextDelta,
  ToolCallEvent,
  ToolResultEvent,
} from '@/gen/v1/ai_chat_pb'
import { ChatClient } from '@tanstack/ai-client'
import { EventType, type StreamChunk } from '@tanstack/ai/client'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createAiOpsConnection } from './ai-ops-connection'

const harness = vi.hoisted(() => ({
  events: [] as Array<unknown>,
  lastRequest: null as Record<string, unknown> | null,
}))

vi.mock('@/lib/api-client', () => ({
  aiChatClient: {
    chat: (request: Record<string, unknown>) => {
      harness.lastRequest = request
      return (async function* () {
        for (const event of harness.events) {
          yield event
        }
      })()
    },
  },
}))

function textEvent(text: string, eventId = 1): ChatEvent {
  return new ChatEvent({
    eventId: BigInt(eventId),
    event: { case: 'text', value: new TextDelta({ text }) },
  })
}

function doneEvent(conversationId: string): ChatEvent {
  return new ChatEvent({
    eventId: BigInt(9),
    event: { case: 'done', value: new DoneEvent({ conversationId }) },
  })
}

const messages = [
  {
    id: 'u1',
    role: 'user' as const,
    parts: [{ type: 'text' as const, content: 'cek interface' }],
  },
]

async function collect(
  forwardedProps: Record<string, unknown>
): Promise<StreamChunk[]> {
  const connection = createAiOpsConnection()
  const chunks: StreamChunk[] = []
  const stream = connection.connect(
    messages,
    undefined,
    new AbortController().signal,
    {
      threadId: 'thread-1',
      runId: 'run-1',
      forwardedProps,
    }
  )
  for await (const chunk of stream) {
    chunks.push(chunk)
  }
  return chunks
}

describe('createAiOpsConnection', () => {
  beforeEach(() => {
    harness.events = [textEvent('Halo'), doneEvent('conv-1')]
    harness.lastRequest = null
  })

  it('turns one chat turn into a complete AG-UI run', async () => {
    const chunks = await collect({
      deviceId: 'dev-1',
      modelConfigId: 'cfg-1',
      conversationId: 'conv-0',
      lastEventId: 7,
    })

    expect(chunks.map((chunk) => chunk.type)).toEqual([
      EventType.RUN_STARTED,
      EventType.TEXT_MESSAGE_START,
      EventType.TEXT_MESSAGE_CONTENT,
      EventType.TEXT_MESSAGE_END,
      EventType.CUSTOM,
      EventType.RUN_FINISHED,
    ])
  })

  it('sends the newest user message and the conversation cursor to Connect', async () => {
    await collect({
      deviceId: 'dev-1',
      modelConfigId: 'cfg-1',
      conversationId: 'conv-0',
      lastEventId: 7,
    })

    expect(harness.lastRequest).toMatchObject({
      deviceId: 'dev-1',
      modelConfigId: 'cfg-1',
      conversationId: 'conv-0',
      message: 'cek interface',
      sinceEventId: BigInt(7),
    })
  })

  it('fails closed when no router is selected', async () => {
    const chunks = await collect({ deviceId: '' })

    expect(chunks).toHaveLength(2)
    expect(chunks[0]?.type).toBe(EventType.RUN_STARTED)
    expect(chunks[1]).toMatchObject({
      type: EventType.RUN_ERROR,
      message: 'Pilih router terlebih dahulu sebelum mengirim pesan.',
    })
    expect(harness.lastRequest).toBeNull()
  })

  it('produces assistant text parts through the real chat client', async () => {
    const client = new ChatClient({
      connection: createAiOpsConnection(),
      threadId: 'thread-1',
      forwardedProps: { deviceId: 'dev-1' },
    })

    await client.sendMessage('cek interface')

    const snapshot = client.getSnapshot()
    const assistant = snapshot.messages.find(
      (message) => message.role === 'assistant'
    )

    expect(assistant?.parts).toContainEqual({
      type: 'text',
      content: 'Halo',
    })
    expect(snapshot.status).toBe('ready')
  })

  it('produces tool-call parts through the real chat client', async () => {
    harness.events = [
      new ChatEvent({
        eventId: BigInt(1),
        event: {
          case: 'toolCall',
          value: new ToolCallEvent({
            id: 'call-1',
            name: 'router_list_interfaces',
            argsJson: '{"limit":1}',
          }),
        },
      }),
      new ChatEvent({
        eventId: BigInt(2),
        event: {
          case: 'toolResult',
          value: new ToolResultEvent({
            toolCallId: 'call-1',
            result: 'ether1 up',
          }),
        },
      }),
      doneEvent('conv-1'),
    ]

    const client = new ChatClient({
      connection: createAiOpsConnection(),
      threadId: 'thread-1',
      forwardedProps: { deviceId: 'dev-1' },
    })

    await client.sendMessage('cek interface')

    const assistant = client
      .getSnapshot()
      .messages.find((message) => message.role === 'assistant')
    const toolCall = assistant?.parts.find((part) => part.type === 'tool-call')

    expect(toolCall).toMatchObject({
      type: 'tool-call',
      id: 'call-1',
      name: 'router_list_interfaces',
    })
  })
})
