import { useState } from 'react'
import type {
  ToolCallPart,
  ToolResultPart,
  UIMessage,
} from '@tanstack/ai/client'
import {
  BotIcon,
  BrainIcon,
  CheckIcon,
  ChevronRightIcon,
  CopyIcon,
  LoaderCircleIcon,
  RefreshCcwIcon,
  TerminalIcon,
} from 'lucide-react'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Bubble, BubbleContent } from '@/components/ui/bubble'
import { Button } from '@/components/ui/button'
import { CodeBlock, CodeBlockContent } from '@/components/ui/code-block'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  Message,
  MessageAvatar,
  MessageContent,
  MessageFooter,
} from '@/components/ui/message'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { Markdown } from '@/components/markdown'

export interface ChatMessageProps {
  message: UIMessage
  isStreaming?: boolean
  onRetry?: () => void
}

function messageText(message: UIMessage): string {
  return message.parts
    .filter((part) => part.type === 'text')
    .map((part) => part.content)
    .join('\n\n')
    .trim()
}

function isEmptyAssistant(message: UIMessage): boolean {
  return message.parts.every(
    (part) => part.type === 'text' && part.content.trim().length === 0
  )
}

export function ChatMessage({
  message,
  isStreaming,
  onRetry,
}: ChatMessageProps) {
  const isUser = message.role === 'user'
  const text = messageText(message)
  const showTyping = !isUser && isStreaming && isEmptyAssistant(message)

  return (
    <Message align={isUser ? 'end' : 'start'}>
      <MessageAvatar>
        <Avatar className='size-7'>
          <AvatarFallback
            className={
              isUser
                ? 'bg-primary text-xs font-semibold text-primary-foreground'
                : 'bg-primary/10 text-primary'
            }
          >
            {isUser ? 'U' : <BotIcon className='size-4' />}
          </AvatarFallback>
        </Avatar>
      </MessageAvatar>

      <MessageContent>
        {isUser ? (
          <Bubble align='end' variant='muted'>
            <BubbleContent className='whitespace-pre-wrap'>
              {text}
            </BubbleContent>
          </Bubble>
        ) : (
          <AssistantParts message={message} />
        )}

        {showTyping && (
          <div className='flex items-center gap-2 text-xs text-muted-foreground'>
            <LoaderCircleIcon className='size-3.5 animate-spin' />
            Menyusun jawaban…
          </div>
        )}

        {!isStreaming && text.length > 0 && (
          <MessageFooter className={isUser ? undefined : 'px-0'}>
            <TooltipProvider delayDuration={200}>
              <CopyAction text={text} />
              {!isUser && onRetry && <RetryAction onRetry={onRetry} />}
            </TooltipProvider>
          </MessageFooter>
        )}
      </MessageContent>
    </Message>
  )
}

function AssistantParts({ message }: { message: UIMessage }) {
  const results = new Map<string, ToolResultPart>()
  for (const part of message.parts) {
    if (part.type === 'tool-result') {
      results.set(part.toolCallId, part)
    }
  }

  return (
    <>
      {message.parts.map((part, index) => {
        switch (part.type) {
          case 'text':
            return (
              <Markdown key={`text-${index}`} className='max-w-none'>
                {part.content}
              </Markdown>
            )
          case 'thinking':
            return (
              <ThinkingBlock key={`thinking-${index}`} content={part.content} />
            )
          case 'tool-call':
            return (
              <ToolCallBlock
                key={part.id}
                part={part}
                result={results.get(part.id)}
              />
            )
          default:
            return null
        }
      })}
    </>
  )
}

function ThinkingBlock({ content }: { content: string }) {
  if (!content.trim()) return null

  return (
    <Collapsible>
      <CollapsibleTrigger asChild>
        <button
          type='button'
          className='group flex items-center gap-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground'
        >
          <BrainIcon className='size-3.5' />
          <span>Penalaran</span>
          <ChevronRightIcon className='size-3.5 transition-transform group-data-[state=open]:rotate-90' />
        </button>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <p className='mt-1.5 border-l pl-3 text-xs whitespace-pre-wrap text-muted-foreground'>
          {content}
        </p>
      </CollapsibleContent>
    </Collapsible>
  )
}

function toolStatus(result?: ToolResultPart) {
  if (!result) return 'running' as const
  return result.state === 'error' ? ('error' as const) : ('done' as const)
}

function resultText(result: ToolResultPart): string {
  if (result.error) return result.error
  if (typeof result.content === 'string') return result.content
  return JSON.stringify(result.content, null, 2)
}

function prettyJson(raw: string): string {
  if (!raw) return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

function ToolCallBlock({
  part,
  result,
}: {
  part: ToolCallPart
  result?: ToolResultPart
}) {
  const status = toolStatus(result)
  const args = prettyJson(part.arguments)
  const output = result ? resultText(result) : ''

  return (
    <Collapsible className='rounded-lg border bg-muted/20'>
      <CollapsibleTrigger asChild>
        <button
          type='button'
          className='group flex w-full items-center gap-2 px-3 py-2 text-left'
        >
          <TerminalIcon className='size-3.5 shrink-0 text-muted-foreground' />
          <span className='min-w-0 flex-1 truncate font-mono text-xs'>
            {part.name}
          </span>
          <Badge
            variant={status === 'error' ? 'destructive' : 'outline'}
            className='h-5 px-1.5 text-[10px] tracking-wide uppercase'
          >
            {status}
          </Badge>
          <ChevronRightIcon className='size-3.5 shrink-0 text-muted-foreground transition-transform group-data-[state=open]:rotate-90' />
        </button>
      </CollapsibleTrigger>
      <CollapsibleContent className='space-y-2 border-t px-3 py-2'>
        {args && (
          <div className='space-y-1'>
            <p className='text-[10px] font-semibold tracking-wide text-muted-foreground uppercase'>
              Argumen
            </p>
            <CodeBlock size='sm'>
              <CodeBlockContent wrap className='max-h-48 font-mono text-[11px]'>
                {args}
              </CodeBlockContent>
            </CodeBlock>
          </div>
        )}
        {output && (
          <div className='space-y-1'>
            <p className='text-[10px] font-semibold tracking-wide text-muted-foreground uppercase'>
              Hasil
            </p>
            <CodeBlock size='sm'>
              <CodeBlockContent wrap className='max-h-56 font-mono text-[11px]'>
                {output}
              </CodeBlockContent>
            </CodeBlock>
          </div>
        )}
      </CollapsibleContent>
    </Collapsible>
  )
}

function CopyAction({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)

  async function copy() {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 2000)
    } catch {
      setCopied(false)
    }
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant='ghost'
          size='icon-xs'
          aria-label={copied ? 'Tersalin' : 'Salin pesan'}
          onClick={() => void copy()}
        >
          {copied ? <CheckIcon /> : <CopyIcon />}
        </Button>
      </TooltipTrigger>
      <TooltipContent>{copied ? 'Tersalin' : 'Salin'}</TooltipContent>
    </Tooltip>
  )
}

function RetryAction({ onRetry }: { onRetry: () => void }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant='ghost'
          size='icon-xs'
          aria-label='Ulangi jawaban'
          onClick={onRetry}
        >
          <RefreshCcwIcon />
        </Button>
      </TooltipTrigger>
      <TooltipContent>Ulangi</TooltipContent>
    </Tooltip>
  )
}
