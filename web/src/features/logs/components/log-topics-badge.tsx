interface LogTopicsBadgeProps {
  topics: string
  className?: string
}

export function LogTopicsBadge({ topics, className = '' }: LogTopicsBadgeProps) {
  if (!topics || !topics.trim()) {
    return <span className='text-[10px] text-muted-foreground/60 select-none'>-</span>
  }

  const topicList = topics
    .split(',')
    .map((t) => t.trim())
    .filter(Boolean)

  if (topicList.length === 0) {
    return <span className='text-[10px] text-muted-foreground/60 select-none'>-</span>
  }

  return (
    <div className={`flex flex-wrap items-center gap-1 ${className}`}>
      {topicList.map((topic, idx) => (
        <span
          key={idx}
          className='inline-block rounded-xs bg-muted/90 border border-border/40 px-1 py-0.2 font-mono text-[10px] text-muted-foreground select-none'
        >
          {topic}
        </span>
      ))}
    </div>
  )
}
