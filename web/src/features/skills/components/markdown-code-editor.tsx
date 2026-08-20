import React from 'react'
import CodeMirror, { EditorView } from '@uiw/react-codemirror'
import { markdown } from '@codemirror/lang-markdown'
import { useTheme } from '@/context/theme-provider'

interface MarkdownCodeEditorProps {
  value?: string
  content?: string
  onChange: (val: string) => void
}

export const MarkdownCodeEditor: React.FC<MarkdownCodeEditorProps> = ({
  value,
  content,
  onChange,
}) => {
  const { theme } = useTheme()
  const isDark =
    theme === 'dark' ||
    (theme === 'system' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)

  const effectiveValue = value !== undefined ? value : content || ''

  return (
    <div className='h-full w-full overflow-hidden font-mono text-sm'>
      <CodeMirror
        value={effectiveValue}
        height='100%'
        className='h-full [&_.cm-editor]:h-full [&_.cm-scroller]:h-full'
        theme={isDark ? 'dark' : 'light'}
        extensions={[markdown(), EditorView.lineWrapping]}
        onChange={onChange}
        basicSetup={{
          lineNumbers: true,
          highlightActiveLineGutter: true,
          highlightSpecialChars: true,
          history: true,
          foldGutter: true,
          drawSelection: true,
          dropCursor: true,
          allowMultipleSelections: true,
          indentOnInput: true,
          syntaxHighlighting: true,
          bracketMatching: true,
          closeBrackets: true,
          autocompletion: true,
          rectangularSelection: true,
          crosshairCursor: true,
          highlightActiveLine: true,
          highlightSelectionMatches: true,
          closeBracketsKeymap: true,
          defaultKeymap: true,
          searchKeymap: true,
          historyKeymap: true,
          foldKeymap: true,
          completionKeymap: true,
          lintKeymap: true,
        }}
      />
    </div>
  )
}
