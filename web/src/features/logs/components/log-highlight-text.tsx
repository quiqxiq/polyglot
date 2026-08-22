interface LogHighlightTextProps {
  text: string
  highlight?: string
  className?: string
}

export function LogHighlightText({
  text,
  highlight = '',
  className = '',
}: LogHighlightTextProps) {
  if (!highlight || !highlight.trim() || !text) {
    return <span className={className}>{text}</span>
  }

  const parts = text.split(
    new RegExp(`(${highlight.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
  )

  return (
    <span className={className}>
      {parts.map((part, i) =>
        part.toLowerCase() === highlight.toLowerCase() ? (
          <mark
            key={i}
            className='rounded-xs bg-amber-400/35 px-0.5 text-foreground font-semibold dark:bg-amber-400/30'
          >
            {part}
          </mark>
        ) : (
          part
        )
      )}
    </span>
  )
}
