import {
  AlertCircleIcon,
  CoinsIcon,
  ListTodoIcon,
  TerminalIcon,
} from 'lucide-react'
import { CodeBlock, CodeBlockContent } from '@/components/ui/code-block'
import {
  Trace,
  TraceCollapsible,
  TraceCollapsibleContent,
  TraceCollapsibleTrigger,
  TraceDescription,
  TraceIcon,
  TraceLine,
  TraceTitle,
} from '@/components/ui/trace'
import type { AiOpsTerminal } from '../api/ag-ui-mapping'
import { isTraceEmpty, type AiOpsRunTrace } from '../api/use-ai-ops-chat'
import { ChatApprovalCard } from './chat-approval-card'

type TraceStatus = 'running' | 'error' | 'pending' | undefined

function planStatus(status: string): TraceStatus {
  if (status === 'running') return 'running'
  if (status === 'error') return 'error'
  if (status === 'pending') return 'pending'
  return undefined
}

/**
 * Run-level activity that has no AG-UI part of its own: the plan the agent is
 * working through, commands it ran, token usage, and mutation proposals.
 * Tool calls live on the message itself, not here.
 */
export function ChatTrace({ trace }: { trace: AiOpsRunTrace }) {
  if (isTraceEmpty(trace)) return null

  return (
    <div className='rounded-lg border bg-muted/20 p-3'>
      <div className='mb-2 flex items-center justify-between gap-2'>
        <p className='text-xs font-semibold tracking-wider text-muted-foreground uppercase'>
          Aktivitas AI Ops
        </p>
        {trace.usage && (
          <p className='flex items-center gap-1 text-[11px] text-muted-foreground'>
            <CoinsIcon className='size-3' />
            {trace.usage.inputTokens.toLocaleString('id-ID')} in ·{' '}
            {trace.usage.outputTokens.toLocaleString('id-ID')} out
          </p>
        )}
      </div>

      <Trace>
        {trace.planSteps.map((step, index) => (
          <TraceLine
            key={`${step.step}-${index}`}
            status={planStatus(step.status)}
          >
            <TraceIcon>
              {step.status === 'error' ? (
                <AlertCircleIcon className='size-3.5 text-destructive' />
              ) : (
                <ListTodoIcon className='size-3.5' />
              )}
            </TraceIcon>
            <TraceTitle>{step.step}</TraceTitle>
            <TraceDescription>{step.status}</TraceDescription>
          </TraceLine>
        ))}

        {trace.terminal.map((item, index) => (
          <TerminalTrace key={`${item.command}-${index}`} item={item} />
        ))}
      </Trace>

      {trace.approvals.length > 0 && (
        <div className='mt-3 flex flex-col gap-2'>
          {trace.approvals.map((approval) => (
            <ChatApprovalCard key={approval.approvalId} approval={approval} />
          ))}
        </div>
      )}
    </div>
  )
}

function TerminalTrace({ item }: { item: AiOpsTerminal }) {
  const detail = [item.command, item.output].filter(Boolean).join('\n\n')
  if (!detail) return null

  return (
    <TraceCollapsible>
      <TraceCollapsibleTrigger>
        <TraceLine status={item.isError ? 'error' : undefined}>
          <TraceIcon>
            <TerminalIcon className='size-3.5' />
          </TraceIcon>
          <TraceTitle>{item.command || 'Perintah router'}</TraceTitle>
          {item.isError && <TraceDescription>error</TraceDescription>}
        </TraceLine>
      </TraceCollapsibleTrigger>
      <TraceCollapsibleContent>
        <CodeBlock size='sm'>
          <CodeBlockContent wrap className='max-h-56 font-mono text-[11px]'>
            {detail}
          </CodeBlockContent>
        </CodeBlock>
      </TraceCollapsibleContent>
    </TraceCollapsible>
  )
}
