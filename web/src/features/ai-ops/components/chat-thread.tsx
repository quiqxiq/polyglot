import type { UIMessage } from '@tanstack/ai/client'
import { BotIcon, SparklesIcon } from 'lucide-react'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import {
  MessageScroller,
  MessageScrollerButton,
  MessageScrollerContent,
  MessageScrollerItem,
  MessageScrollerProvider,
  MessageScrollerViewport,
} from '@/components/ui/message-scroller'
import { isTraceEmpty, type AiOpsRunTrace } from '../api/use-ai-ops-chat'
import { ChatMessage } from './chat-message'
import { ChatTrace } from './chat-trace'

const SUGGESTIONS = [
  'Ringkas status interface dan IP router ini.',
  'Periksa konfigurasi DNS dan firewall yang menyimpang.',
  'Siapkan proposal untuk menambah VLAN 120.',
]

export interface ChatThreadProps {
  messages: ReadonlyArray<UIMessage>
  trace: AiOpsRunTrace
  isStreaming: boolean
  onRetry?: () => void
  onSuggestion?: (text: string) => void
}

export function ChatThread({
  messages,
  trace,
  isStreaming,
  onRetry,
  onSuggestion,
}: ChatThreadProps) {
  const lastIndex = messages.length - 1

  return (
    <MessageScrollerProvider autoScroll>
      <MessageScroller className='flex-1'>
        <MessageScrollerViewport className='pr-2'>
          <MessageScrollerContent className='gap-4 pb-2'>
            {messages.length === 0 && (
              <Empty className='my-auto py-12'>
                <EmptyMedia variant='icon'>
                  <BotIcon className='size-6 text-primary' />
                </EmptyMedia>
                <EmptyHeader>
                  <EmptyTitle>AI Router Ops</EmptyTitle>
                  <EmptyDescription>
                    Tanyakan status router, minta inspeksi menu, atau minta AI
                    membuat proposal perubahan. Operasi READ dieksekusi
                    langsung, operasi MUTATION wajib melewati persetujuan di
                    Approval Inbox.
                  </EmptyDescription>
                </EmptyHeader>
                {onSuggestion && (
                  <div className='flex flex-wrap justify-center gap-2'>
                    {SUGGESTIONS.map((suggestion) => (
                      <button
                        key={suggestion}
                        type='button'
                        onClick={() => onSuggestion(suggestion)}
                        className='flex items-center gap-1.5 rounded-full border bg-background px-3 py-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground'
                      >
                        <SparklesIcon className='size-3' />
                        {suggestion}
                      </button>
                    ))}
                  </div>
                )}
              </Empty>
            )}

            {messages.map((message, index) => (
              <MessageScrollerItem
                key={message.id}
                messageId={message.id}
                scrollAnchor={message.role === 'user'}
              >
                <ChatMessage
                  message={message}
                  isStreaming={isStreaming && index === lastIndex}
                  onRetry={
                    index === lastIndex && message.role === 'assistant'
                      ? onRetry
                      : undefined
                  }
                />
              </MessageScrollerItem>
            ))}

            {!isTraceEmpty(trace) && (
              <MessageScrollerItem messageId='ai-ops-trace'>
                <ChatTrace trace={trace} />
              </MessageScrollerItem>
            )}
          </MessageScrollerContent>
        </MessageScrollerViewport>
        <MessageScrollerButton />
      </MessageScroller>
    </MessageScrollerProvider>
  )
}
