import {
  ApprovalRequestedEvent,
  ChatEvent,
  DoneEvent,
  ErrorEvent,
  PlanStepEvent,
  TerminalOutputEvent,
  TextDelta,
  ToolCallEvent,
  ToolResultEvent,
  UsageEvent,
} from '@/gen/v1/ai_chat_pb'
import { EventType, type StreamChunk } from '@tanstack/ai/client'
import { describe, expect, it } from 'vitest'
import {
  AI_OPS_CUSTOM,
  closeOpenMessages,
  createMappingState,
  mapChatEvent,
  parseAiOpsCustomEvent,
  startChunk,
  type AiOpsMappingState,
} from './ag-ui-mapping'

let seq = 0

function event(payload: ChatEvent['event'], eventId = ++seq): ChatEvent {
  return new ChatEvent({ eventId: BigInt(eventId), event: payload })
}

function text(content: string, eventId?: number) {
  return event(
    { case: 'text', value: new TextDelta({ text: content }) },
    eventId
  )
}

function reasoning(content: string, eventId?: number) {
  return event(
    { case: 'reasoning', value: new TextDelta({ text: content }) },
    eventId
  )
}

function types(chunks: StreamChunk[]): EventType[] {
  return chunks.map((chunk) => chunk.type as EventType)
}

function freshState(): AiOpsMappingState {
  return createMappingState({ threadId: 'thread-1', runId: 'run-1' })
}

describe('mapChatEvent', () => {
  it('opens one text message and appends every text delta to it', () => {
    const state = freshState()

    const first = mapChatEvent(state, text('Halo'))
    const second = mapChatEvent(state, text(' router'))

    expect(types(first)).toEqual([
      EventType.TEXT_MESSAGE_START,
      EventType.TEXT_MESSAGE_CONTENT,
    ])
    expect(types(second)).toEqual([EventType.TEXT_MESSAGE_CONTENT])

    const start = first[0] as { messageId: string }
    const delta = second[0] as { messageId: string; delta: string }
    expect(delta.messageId).toBe(start.messageId)
    expect(delta.delta).toBe(' router')
  })

  it('closes the text message before a tool call and reports the result', () => {
    const state = freshState()
    mapChatEvent(state, text('Cek interface'))

    const call = mapChatEvent(
      state,
      event({
        case: 'toolCall',
        value: new ToolCallEvent({
          id: 'call-1',
          name: 'router_list_interfaces',
          argsJson: '{}',
        }),
      })
    )

    expect(types(call)).toEqual([
      EventType.TEXT_MESSAGE_END,
      EventType.TOOL_CALL_START,
      EventType.TOOL_CALL_ARGS,
    ])
    expect(state.textMessageId).toBeNull()

    const result = mapChatEvent(
      state,
      event({
        case: 'toolResult',
        value: new ToolResultEvent({
          toolCallId: 'call-1',
          result: 'ether1 up',
        }),
      })
    )

    expect(types(result)).toEqual([
      EventType.TOOL_CALL_END,
      EventType.TOOL_CALL_RESULT,
    ])
    expect(state.openToolCalls.size).toBe(0)
  })

  it('maps reasoning deltas to a reasoning message', () => {
    const state = freshState()

    const chunks = mapChatEvent(state, reasoning('Menganalisis'))

    expect(types(chunks)).toEqual([
      EventType.REASONING_MESSAGE_START,
      EventType.REASONING_MESSAGE_CONTENT,
    ])
  })

  it('forwards plan steps, approvals, terminal output and usage as custom events', () => {
    const state = freshState()

    const plan = mapChatEvent(
      state,
      event({
        case: 'planStep',
        value: new PlanStepEvent({ step: 'Baca konteks', status: 'running' }),
      })
    )
    const approval = mapChatEvent(
      state,
      event({
        case: 'approval',
        value: new ApprovalRequestedEvent({
          approvalId: 'apr-1',
          summary: 'Set DNS',
          restorePointId: 'rp-1',
        }),
      })
    )
    const terminal = mapChatEvent(
      state,
      event({
        case: 'terminal',
        value: new TerminalOutputEvent({
          command: '/ip/dns/print',
          output: 'ok',
        }),
      })
    )
    const usage = mapChatEvent(
      state,
      event({
        case: 'usage',
        value: new UsageEvent({ inputTokens: 10, outputTokens: 20 }),
      })
    )

    expect(plan[0]).toMatchObject({
      type: EventType.CUSTOM,
      name: AI_OPS_CUSTOM.planStep,
      value: { step: 'Baca konteks', status: 'running' },
    })
    expect(approval[0]).toMatchObject({
      type: EventType.CUSTOM,
      name: AI_OPS_CUSTOM.approval,
      value: { approvalId: 'apr-1' },
    })
    expect(terminal[0]).toMatchObject({
      type: EventType.CUSTOM,
      name: AI_OPS_CUSTOM.terminal,
      value: { command: '/ip/dns/print' },
    })
    expect(usage[0]).toMatchObject({
      type: EventType.CUSTOM,
      name: AI_OPS_CUSTOM.usage,
      value: { inputTokens: 10, outputTokens: 20 },
    })
  })

  it('closes the open message, reports the conversation and finishes the run', () => {
    const state = freshState()
    mapChatEvent(state, text('Selesai'))

    const chunks = mapChatEvent(
      state,
      event({
        case: 'done',
        value: new DoneEvent({
          conversationId: 'conv-1',
          lastEventId: BigInt(42),
        }),
      })
    )

    expect(types(chunks)).toEqual([
      EventType.TEXT_MESSAGE_END,
      EventType.CUSTOM,
      EventType.RUN_FINISHED,
    ])
    expect(chunks[1]).toMatchObject({
      name: AI_OPS_CUSTOM.conversation,
      value: { conversationId: 'conv-1', lastEventId: 42 },
    })
    expect(state.done).toBe(true)
    expect(mapChatEvent(state, text('dilewatkan'))).toEqual([])
  })

  it('turns an error event into RUN_ERROR and stops the run', () => {
    const state = freshState()
    mapChatEvent(state, text('Sebagian'))

    const chunks = mapChatEvent(
      state,
      event({ case: 'error', value: new ErrorEvent({ message: 'boom' }) })
    )

    expect(types(chunks)).toEqual([
      EventType.TEXT_MESSAGE_END,
      EventType.RUN_ERROR,
    ])
    expect(state.done).toBe(true)
  })

  it('closes leftover messages when the stream ends without done', () => {
    const textState = freshState()
    mapChatEvent(textState, text('Terputus'))

    expect(types(closeOpenMessages(textState))).toEqual([
      EventType.TEXT_MESSAGE_END,
    ])
    expect(closeOpenMessages(textState)).toEqual([])

    const reasoningState = freshState()
    mapChatEvent(reasoningState, reasoning('masih berpikir'))

    expect(types(closeOpenMessages(reasoningState))).toEqual([
      EventType.REASONING_MESSAGE_END,
    ])
  })

  it('ignores events that carry no payload', () => {
    const state = freshState()
    expect(mapChatEvent(state, new ChatEvent())).toEqual([])
  })

  it('starts a run with RUN_STARTED', () => {
    expect(startChunk(freshState())).toMatchObject({
      type: EventType.RUN_STARTED,
      threadId: 'thread-1',
      runId: 'run-1',
    })
  })
})

describe('parseAiOpsCustomEvent', () => {
  it('narrows the AI Ops custom events it knows', () => {
    const [chunk] = mapChatEvent(
      freshState(),
      event({
        case: 'usage',
        value: new UsageEvent({ inputTokens: 1, outputTokens: 2 }),
      })
    )

    expect(parseAiOpsCustomEvent(chunk)).toEqual({
      name: AI_OPS_CUSTOM.usage,
      value: { inputTokens: 1, outputTokens: 2 },
    })
  })

  it('ignores custom events owned by someone else', () => {
    expect(
      parseAiOpsCustomEvent({
        type: EventType.CUSTOM,
        name: 'other.tool',
        value: {},
      })
    ).toBeNull()
  })
})
