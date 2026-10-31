import type { ChatEvent } from '@/gen/v1/ai_chat_pb'
import { EventType, type StreamChunk } from '@tanstack/ai/client'

/**
 * Names of the `CUSTOM` AG-UI events the AI Ops adapter emits. `useChat` only
 * keeps text/reasoning/tool parts, so plan steps, approvals, terminal output
 * and token usage travel as custom events and are collected by the caller.
 */
export const AI_OPS_CUSTOM = {
  conversation: 'ai-ops.conversation',
  planStep: 'ai-ops.plan-step',
  approval: 'ai-ops.approval',
  terminal: 'ai-ops.terminal',
  usage: 'ai-ops.usage',
} as const

export interface AiOpsConversation {
  conversationId: string
  lastEventId: number
}

export interface AiOpsPlanStep {
  step: string
  status: string
  detail: string
}

export interface AiOpsApproval {
  approvalId: string
  summary: string
  restorePointId: string
}

export interface AiOpsTerminal {
  command: string
  output: string
  isError: boolean
}

export interface AiOpsUsage {
  inputTokens: number
  outputTokens: number
}

export type AiOpsCustomEvent =
  | { name: typeof AI_OPS_CUSTOM.conversation; value: AiOpsConversation }
  | { name: typeof AI_OPS_CUSTOM.planStep; value: AiOpsPlanStep }
  | { name: typeof AI_OPS_CUSTOM.approval; value: AiOpsApproval }
  | { name: typeof AI_OPS_CUSTOM.terminal; value: AiOpsTerminal }
  | { name: typeof AI_OPS_CUSTOM.usage; value: AiOpsUsage }

export interface AiOpsMappingState {
  runId: string
  threadId: string
  textMessageId: string | null
  reasoningMessageId: string | null
  textSeq: number
  reasoningSeq: number
  openToolCalls: Set<string>
  lastEventId: number
  done: boolean
}

export function createMappingState(input: {
  threadId: string
  runId: string
}): AiOpsMappingState {
  return {
    runId: input.runId,
    threadId: input.threadId,
    textMessageId: null,
    reasoningMessageId: null,
    textSeq: 0,
    reasoningSeq: 0,
    openToolCalls: new Set(),
    lastEventId: 0,
    done: false,
  }
}

/** Emits the `RUN_STARTED` chunk that must open every AG-UI run. */
export function startChunk(state: AiOpsMappingState): StreamChunk {
  return {
    type: EventType.RUN_STARTED,
    threadId: state.threadId,
    runId: state.runId,
  }
}

function ensureText(state: AiOpsMappingState, out: StreamChunk[]): string {
  if (state.textMessageId) return state.textMessageId
  const messageId = `${state.runId}-text-${state.textSeq++}`
  state.textMessageId = messageId
  out.push({
    type: EventType.TEXT_MESSAGE_START,
    messageId,
    role: 'assistant',
  })
  return messageId
}

function ensureReasoning(state: AiOpsMappingState, out: StreamChunk[]): string {
  if (state.reasoningMessageId) return state.reasoningMessageId
  const messageId = `${state.runId}-reasoning-${state.reasoningSeq++}`
  state.reasoningMessageId = messageId
  out.push({
    type: EventType.REASONING_MESSAGE_START,
    messageId,
    role: 'reasoning',
  })
  return messageId
}

/** Closes an open assistant text message, if any. */
export function closeText(state: AiOpsMappingState): StreamChunk[] {
  if (!state.textMessageId) return []
  const messageId = state.textMessageId
  state.textMessageId = null
  return [{ type: EventType.TEXT_MESSAGE_END, messageId }]
}

/** Closes an open reasoning message, if any. */
export function closeReasoning(state: AiOpsMappingState): StreamChunk[] {
  if (!state.reasoningMessageId) return []
  const messageId = state.reasoningMessageId
  state.reasoningMessageId = null
  return [{ type: EventType.REASONING_MESSAGE_END, messageId }]
}

/**
 * Closes every message the run left open. Called when the transport ends so a
 * cancelled or truncated stream never leaves a part stuck in `streaming`.
 */
export function closeOpenMessages(state: AiOpsMappingState): StreamChunk[] {
  return [...closeText(state), ...closeReasoning(state)]
}

/**
 * Maps one `ChatEvent` from `AIChatService.Chat` to the AG-UI chunks
 * `useChat` understands.
 *
 * The backend currently sends the whole assistant reply as a single `text`
 * event, so this produces a single `TEXT_MESSAGE_CONTENT`. Keep the mapping
 * delta-shaped anyway: when the backend starts emitting real deltas the same
 * code works without change.
 */
export function mapChatEvent(
  state: AiOpsMappingState,
  event: ChatEvent
): StreamChunk[] {
  if (state.done || !event.event) return []

  const eventId = Number(event.eventId)
  if (eventId > 0) state.lastEventId = eventId
  const out: StreamChunk[] = []
  const payload = event.event

  switch (payload.case) {
    case 'text': {
      if (!payload.value.text) break
      out.push(...closeReasoning(state))
      const messageId = ensureText(state, out)
      out.push({
        type: EventType.TEXT_MESSAGE_CONTENT,
        messageId,
        delta: payload.value.text,
      })
      break
    }

    case 'reasoning': {
      if (!payload.value.text) break
      out.push(...closeText(state))
      const messageId = ensureReasoning(state, out)
      out.push({
        type: EventType.REASONING_MESSAGE_CONTENT,
        messageId,
        delta: payload.value.text,
      })
      break
    }

    case 'toolCall': {
      out.push(...closeText(state), ...closeReasoning(state))
      const toolCallId =
        payload.value.id || `${state.runId}-tool-${state.lastEventId}`
      state.openToolCalls.add(toolCallId)
      out.push({
        type: EventType.TOOL_CALL_START,
        toolCallId,
        toolCallName: payload.value.name,
      })
      if (payload.value.argsJson) {
        out.push({
          type: EventType.TOOL_CALL_ARGS,
          toolCallId,
          delta: payload.value.argsJson,
        })
      }
      break
    }

    case 'toolResult': {
      out.push(...closeText(state), ...closeReasoning(state))
      const toolCallId = payload.value.toolCallId
      if (state.openToolCalls.delete(toolCallId)) {
        out.push({ type: EventType.TOOL_CALL_END, toolCallId })
      }
      out.push({
        type: EventType.TOOL_CALL_RESULT,
        messageId: `${state.runId}-tool-result-${toolCallId}`,
        toolCallId,
        content: payload.value.isError
          ? `Error: ${payload.value.result}`
          : payload.value.result,
        role: 'tool',
      })
      break
    }

    case 'planStep': {
      out.push(...closeText(state), ...closeReasoning(state))
      const value: AiOpsPlanStep = {
        step: payload.value.step,
        status: payload.value.status,
        detail: payload.value.detail,
      }
      out.push({ type: EventType.CUSTOM, name: AI_OPS_CUSTOM.planStep, value })
      break
    }

    case 'approval': {
      out.push(...closeText(state), ...closeReasoning(state))
      const value: AiOpsApproval = {
        approvalId: payload.value.approvalId,
        summary: payload.value.summary,
        restorePointId: payload.value.restorePointId,
      }
      out.push({ type: EventType.CUSTOM, name: AI_OPS_CUSTOM.approval, value })
      break
    }

    case 'terminal': {
      out.push(...closeText(state), ...closeReasoning(state))
      const value: AiOpsTerminal = {
        command: payload.value.command,
        output: payload.value.output,
        isError: payload.value.isError,
      }
      out.push({ type: EventType.CUSTOM, name: AI_OPS_CUSTOM.terminal, value })
      break
    }

    case 'usage': {
      const value: AiOpsUsage = {
        inputTokens: payload.value.inputTokens,
        outputTokens: payload.value.outputTokens,
      }
      out.push({ type: EventType.CUSTOM, name: AI_OPS_CUSTOM.usage, value })
      break
    }

    case 'done': {
      out.push(...closeOpenMessages(state))
      state.done = true
      const doneLastEventId = Number(payload.value.lastEventId)
      if (doneLastEventId > 0) state.lastEventId = doneLastEventId
      if (payload.value.conversationId) {
        const value: AiOpsConversation = {
          conversationId: payload.value.conversationId,
          lastEventId: state.lastEventId,
        }
        out.push({
          type: EventType.CUSTOM,
          name: AI_OPS_CUSTOM.conversation,
          value,
        })
      }
      out.push({
        type: EventType.RUN_FINISHED,
        threadId: state.threadId,
        runId: state.runId,
      })
      break
    }

    case 'error': {
      out.push(...closeOpenMessages(state))
      state.done = true
      out.push({
        type: EventType.RUN_ERROR,
        message: payload.value.message || 'AI chat failed',
      })
      break
    }

    default:
      break
  }

  return out
}

/** Narrows a `CUSTOM` chunk to the typed AI Ops payload, when it is one. */
export function parseAiOpsCustomEvent(
  chunk: StreamChunk
): AiOpsCustomEvent | null {
  if (chunk.type !== EventType.CUSTOM) return null
  switch (chunk.name) {
    case AI_OPS_CUSTOM.conversation:
      return { name: chunk.name, value: chunk.value as AiOpsConversation }
    case AI_OPS_CUSTOM.planStep:
      return { name: chunk.name, value: chunk.value as AiOpsPlanStep }
    case AI_OPS_CUSTOM.approval:
      return { name: chunk.name, value: chunk.value as AiOpsApproval }
    case AI_OPS_CUSTOM.terminal:
      return { name: chunk.name, value: chunk.value as AiOpsTerminal }
    case AI_OPS_CUSTOM.usage:
      return { name: chunk.name, value: chunk.value as AiOpsUsage }
    default:
      return null
  }
}
