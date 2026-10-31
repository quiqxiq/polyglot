import { memo } from 'react'
import { cjk } from '@streamdown/cjk'
import { code } from '@streamdown/code'
import { math } from '@streamdown/math'
import { mermaid } from '@streamdown/mermaid'
import { Streamdown, type StreamdownProps } from 'streamdown'
import { cn } from '@/lib/utils'

const plugins = { cjk, code, math, mermaid }

/**
 * Markdown renderer for AI answers. Streamdown tolerates partial markdown, so
 * an unfinished code fence or table never breaks the layout while a reply is
 * still streaming in.
 */
export const Markdown = memo(function Markdown({
  className,
  ...props
}: StreamdownProps) {
  return (
    <Streamdown
      className={cn(
        'size-full text-sm leading-relaxed [&>*:first-child]:mt-0 [&>*:last-child]:mb-0',
        className
      )}
      plugins={plugins}
      {...props}
    />
  )
})
