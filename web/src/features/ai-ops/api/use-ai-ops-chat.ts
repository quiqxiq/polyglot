import { useCallback, useMemo, useState } from 'react'
import { useChat } from '@tanstack/ai-react'
import {
  AI_OPS_CUSTOM,
  parseAiOpsCustomEvent,
  type AiOpsApproval,
  type AiOpsPlanStep,
  type AiOpsTerminal,
  type AiOpsUsage,
} from './ag-ui-mapping'
import { createAiOpsConnection } from './ai-ops-connection'

export interface AiOpsRunTrace {
  planSteps: AiOpsPlanStep[]
  approvals: AiOpsApproval[]
  terminal: AiOpsTerminal[]
  usage: AiOpsUsage | null
}

const EMPTY_TRACE: AiOpsRunTrace = {
  planSteps: [],
  approvals: [],
  terminal: [],
  usage: null,
}

export function isTraceEmpty(trace: AiOpsRunTrace): boolean {
  return (
    trace.planSteps.length === 0 &&
    trace.approvals.length === 0 &&
    trace.terminal.length === 0 &&
    trace.usage === null
  )
}

export interface UseAiOpsChatOptions {
  deviceId: string
  modelConfigId?: string
}

/**
 * AI Router Ops chat state.
 *
 * `useChat` owns messages, run status and cancellation through the AG-UI
 * transport; this hook keeps the AI Ops extras that AG-UI has no part for
 * (plan steps, approvals, terminal output, token usage) plus the conversation
 * cursor the backend returns on `done`.
 */
export function useAiOpsChat({
  deviceId,
  modelConfigId = '',
}: UseAiOpsChatOptions) {
  const [conversationId, setConversationId] = useState('')
  const [lastEventId, setLastEventId] = useState(0)
  const [threadId] = useState(() => crypto.randomUUID())
  const [trace, setTrace] = useState<AiOpsRunTrace>(EMPTY_TRACE)

  // Created once: a new connection would reset the ChatClient and drop the
  // transcript. Request values travel through `forwardedProps` instead.
  const connection = useMemo(() => createAiOpsConnection(), [])

  const forwardedProps = useMemo(
    () => ({ deviceId, modelConfigId, conversationId, lastEventId }),
    [deviceId, modelConfigId, conversationId, lastEventId]
  )

  const chat = useChat({
    connection,
    threadId,
    forwardedProps,
    onChunk: (chunk) => {
      const custom = parseAiOpsCustomEvent(chunk)
      if (!custom) return

      switch (custom.name) {
        case AI_OPS_CUSTOM.conversation:
          setConversationId(custom.value.conversationId)
          setLastEventId(custom.value.lastEventId)
          break
        case AI_OPS_CUSTOM.planStep:
          setTrace((prev) => upsertPlanStep(prev, custom.value))
          break
        case AI_OPS_CUSTOM.approval:
          setTrace((prev) => upsertApproval(prev, custom.value))
          break
        case AI_OPS_CUSTOM.terminal:
          setTrace((prev) => ({
            ...prev,
            terminal: [...prev.terminal, custom.value],
          }))
          break
        case AI_OPS_CUSTOM.usage:
          setTrace((prev) => ({ ...prev, usage: custom.value }))
          break
      }
    },
  })

  const { sendMessage, stop, reload } = chat

  const send = useCallback(
    async (text: string) => {
      const prompt = text.trim()
      if (!prompt) return
      setTrace(EMPTY_TRACE)
      try {
        await sendMessage(prompt)
      } catch {
        // The failure is surfaced through `chat.error` and the error banner.
      }
    },
    [sendMessage]
  )

  const retry = useCallback(async () => {
    setTrace(EMPTY_TRACE)
    await reload()
  }, [reload])

  const isBusy = chat.status === 'submitted' || chat.status === 'streaming'

  return {
    messages: chat.messages,
    status: chat.status,
    error: chat.error,
    isBusy,
    conversationId,
    trace,
    send,
    stop,
    retry,
  }
}

function upsertPlanStep(
  trace: AiOpsRunTrace,
  step: AiOpsPlanStep
): AiOpsRunTrace {
  const index = trace.planSteps.findIndex((item) => item.step === step.step)
  if (index === -1) {
    return { ...trace, planSteps: [...trace.planSteps, step] }
  }
  const planSteps = [...trace.planSteps]
  planSteps[index] = step
  return { ...trace, planSteps }
}

function upsertApproval(
  trace: AiOpsRunTrace,
  approval: AiOpsApproval
): AiOpsRunTrace {
  const index = trace.approvals.findIndex(
    (item) => item.approvalId === approval.approvalId
  )
  if (index === -1) {
    return { ...trace, approvals: [...trace.approvals, approval] }
  }
  const approvals = [...trace.approvals]
  approvals[index] = approval
  return { ...trace, approvals }
}
