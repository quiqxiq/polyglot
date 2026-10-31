export const SUBMISSION_STATUS = {
  PENDING: 'PENDING',
  APPROVED: 'APPROVED',
  REJECTED: 'REJECTED',
} as const

export function submissionStatusBadge(status: string) {
  switch (status) {
    case SUBMISSION_STATUS.PENDING:
      return {
        label: 'Menunggu Review',
        className:
          'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30',
      }
    case SUBMISSION_STATUS.APPROVED:
      return {
        label: 'Disetujui',
        className:
          'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
      }
    case SUBMISSION_STATUS.REJECTED:
      return {
        label: 'Ditolak',
        className:
          'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30',
      }
    default:
      return {
        label: status || 'Unknown',
        className: 'bg-muted text-muted-foreground',
      }
  }
}

export function formatIDR(value: number | undefined) {
  return `Rp${Number(value || 0).toLocaleString('id-ID')}`
}
