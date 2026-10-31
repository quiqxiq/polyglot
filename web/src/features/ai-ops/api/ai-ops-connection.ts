import { stream } from '@tanstack/ai-client'
import { EventType, type StreamChunk } from '@tanstack/ai/client'
import { aiChatClient } from '@/lib/api-client'
import {
  closeOpenMessages,
  createMappingState,
  mapChatEvent,
  startChunk,
} from './ag-ui-mapping'
import { lastUserText } from './chat-prompt'

/**
 * Per-session values `useChat` forwards to the transport on every run. They
 * live in `forwardedProps` instead of a ref so the adapter is created once and
 * never reads React state during render.
 */
export interface AiOpsForwardedProps {
  deviceId?: string
  modelConfigId?: string
  conversationId?: string
  lastEventId?: number
}

function runError(message: string): StreamChunk {
  return { type: EventType.RUN_ERROR, message }
}

/**
 * ConnectRPC → AG-UI transport. Wraps `AIChatService.Chat` (server-streaming
 * proto events) into the `AsyncIterable<StreamChunk>` that `useChat` consumes.
 */
export function createAiOpsConnection() {
  return stream(async function* (messages, _data, signal, input) {
    const props = (input.forwardedProps ?? {}) as AiOpsForwardedProps
    const conversationId = props.conversationId ?? ''
    const deviceId = props.deviceId ?? ''
    const runId = crypto.randomUUID()
    const state = createMappingState({
      threadId: conversationId || input.threadId,
      runId,
    })

    yield startChunk(state)

    const prompt = lastUserText(messages)
    if (!prompt) {
      yield runError('Tidak ada pesan untuk dikirim.')
      return
    }
    if (!deviceId) {
      yield runError('Pilih router terlebih dahulu sebelum mengirim pesan.')
      return
    }

    let failure: string | null = null
    try {
      const events = aiChatClient.chat(
        {
          conversationId,
          message: prompt,
          deviceId,
          modelConfigId: props.modelConfigId ?? '',
          sinceEventId: BigInt(props.lastEventId ?? 0),
        },
        { signal }
      )

      for await (const event of events) {
        for (const chunk of mapChatEvent(state, event)) {
          yield chunk
        }
        if (state.done) break
      }
    } catch (error) {
      if (!signal?.aborted) {
        failure =
          error instanceof Error ? error.message : 'Koneksi AI chat terputus.'
      }
    }

    if (failure) {
      yield runError(failure)
      return
    }

    for (const chunk of closeOpenMessages(state)) {
      yield chunk
    }
    if (!state.done) {
      yield {
        type: EventType.RUN_FINISHED,
        threadId: state.threadId,
        runId: state.runId,
        outcome: signal?.aborted ? { type: 'cancelled' } : undefined,
      }
    }
  })
}
