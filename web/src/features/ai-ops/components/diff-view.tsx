import { useMemo } from 'react'
import { cn } from '@/lib/utils'

interface DiffLine {
  op?: string
  path?: string
  before?: string
  after?: string
  text?: string
}

interface DiffShape {
  menu?: string
  verb?: string
  summary?: string
  lines?: DiffLine[]
}

interface Props {
  raw: string
}

function parseDiff(raw: string): DiffShape[] {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw) as
      | DiffShape
      | DiffShape[]
      | { diff?: DiffShape }[]
    if (Array.isArray(parsed)) {
      return parsed.map((p) =>
        'diff' in p ? (p.diff as DiffShape) : (p as DiffShape)
      )
    }
    return [parsed as DiffShape]
  } catch {
    return []
  }
}

export function DiffView({ raw }: Props) {
  const diffs = useMemo(() => parseDiff(raw), [raw])
  if (diffs.length === 0) {
    return (
      <p className='text-sm text-muted-foreground'>
        Tidak ada diff yang dapat ditampilkan.
      </p>
    )
  }
  return (
    <div className='space-y-3'>
      {diffs.map((diff, i) => (
        <div key={i} className='overflow-hidden rounded-md border'>
          <div className='flex items-center justify-between bg-muted/50 px-3 py-2 text-xs'>
            <span className='font-medium'>
              /{diff.menu || '?'}{' '}
              <span className='text-muted-foreground'>/{diff.verb || '?'}</span>
            </span>
            <span className='text-muted-foreground'>{diff.summary}</span>
          </div>
          <div className='divide-y font-mono text-xs'>
            {(diff.lines || []).length === 0 && (
              <p className='p-3 text-muted-foreground'>Tidak ada baris diff.</p>
            )}
            {(diff.lines || []).map((line, j) => (
              <div
                key={j}
                className={cn(
                  'px-3 py-1.5 whitespace-pre-wrap',
                  line.op === 'add' &&
                    'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400',
                  line.op === 'remove' && 'bg-destructive/10 text-destructive',
                  line.op === 'change' &&
                    'bg-amber-500/10 text-amber-700 dark:text-amber-400'
                )}
              >
                {line.text ||
                  `${line.op} ${line.path}: ${line.before} → ${line.after}`}
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
