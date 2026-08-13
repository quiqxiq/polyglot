import { cn } from '@/lib/utils'
import type { WARealtimeStatus } from '@/lib/realtime'

const STATUS_META: Record<WARealtimeStatus, { label: string; text: string; dot: string }> = {
  open: {
    label: 'Live',
    text: 'text-emerald-600',
    dot: 'bg-emerald-500',
  },
  connecting: {
    label: 'Menghubungkan…',
    text: 'text-amber-600',
    dot: 'bg-amber-500 animate-pulse',
  },
  reconnecting: {
    label: 'Menyambung ulang…',
    text: 'text-orange-600',
    dot: 'bg-orange-500 animate-pulse',
  },
  closed: {
    label: 'Terputus',
    text: 'text-muted-foreground',
    dot: 'bg-muted-foreground/60',
  },
}

/**
 * Indikator koneksi realtime (SSE) untuk header halaman. State berasal dari
 * useWARealtimeStream — dot + label kecil yang berubah warna sesuai status
 * koneksi (Live / Menghubungkan / Menyambung ulang / Terputus).
 */
export function SSEIndicator({
  status,
  className,
}: {
  status: WARealtimeStatus
  className?: string
}) {
  const meta = STATUS_META[status] ?? STATUS_META.closed
  return (
    <span
      className={cn(
        'inline-flex shrink-0 items-center gap-1.5 rounded-full border border-border bg-card px-2.5 py-1 text-xs font-medium',
        meta.text,
        className,
      )}
      title={`Koneksi realtime: ${meta.label}`}
    >
      <span className={cn('size-2 rounded-full', meta.dot)} />
      <span className='hidden sm:inline'>{meta.label}</span>
    </span>
  )
}
