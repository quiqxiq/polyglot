import { Badge } from '@/components/ui/badge'

type BadgeVariant = 'default' | 'secondary' | 'outline' | 'destructive'

// Status embed per-dokumen, konsisten dengan backend
// (internal/domain/knowledge/knowledge.go):
//   none     → dokumen lokal, tidak di-embed ke AnythingLLM
//   pending  → embed sedang diproses
//   embedded → aktif di vector store AnythingLLM
//   failed   → sinkronisasi gagal, bisa di-retry
const STATUS_META: Record<string, { label: string; variant: BadgeVariant }> = {
  none: { label: 'Local', variant: 'secondary' },
  pending: { label: 'Pending', variant: 'outline' },
  embedded: { label: 'Embedded', variant: 'default' },
  failed: { label: 'Failed', variant: 'destructive' },
}

export function KnowledgeEmbedStatusBadge({ status }: { status: string }) {
  const meta = STATUS_META[status] ?? {
    label: status || 'Unknown',
    variant: 'secondary' as BadgeVariant,
  }
  return <Badge variant={meta.variant}>{meta.label}</Badge>
}
