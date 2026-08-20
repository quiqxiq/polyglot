import React from 'react'
import MarkdownPreview from '@uiw/react-markdown-preview'
import { useTheme } from '@/context/theme-provider'

interface MarkdownViewerProps {
  content: string
}

function stripFrontmatter(raw: string): string {
  const trimmed = raw.trimStart()
  if (!trimmed.startsWith('---')) {
    return raw
  }
  const rest = trimmed.slice(3)
  const closingIdx = rest.indexOf('---')
  if (closingIdx === -1) {
    return raw
  }
  return rest.slice(closingIdx + 3).trimStart()
}

export const MarkdownViewer: React.FC<MarkdownViewerProps> = ({ content }) => {
  const { theme } = useTheme()
  const isDark =
    theme === 'dark' ||
    (theme === 'system' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)

  const cleanContent = stripFrontmatter(content || '')

  return (
    <div className='w-full px-6 py-4'>
      <div
        className='prose prose-neutral dark:prose-invert max-w-none text-foreground leading-relaxed'
        data-color-mode={isDark ? 'dark' : 'light'}
      >
        <MarkdownPreview
          source={cleanContent || '*Berkas ini kosong.*'}
          style={{
            backgroundColor: 'transparent',
            color: 'inherit',
            fontFamily: 'inherit',
          }}
          wrapperElement={{
            'data-color-mode': isDark ? 'dark' : 'light',
          }}
        />
      </div>
    </div>
  )
}
