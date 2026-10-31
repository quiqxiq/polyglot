export const REGISTRATION_STATUS = {
  PENDING: 'PENDING',
  SURVEY: 'SURVEY',
  APPROVED: 'APPROVED',
  ON_PROGRESS: 'ON_PROGRESS',
  INSTALLED: 'INSTALLED',
  ACTIVE: 'ACTIVE',
  REJECTED: 'REJECTED',
  CANCELLED: 'CANCELLED',
} as const

export type RegistrationStatus =
  (typeof REGISTRATION_STATUS)[keyof typeof REGISTRATION_STATUS]

export const SURVEY_STATUS = {
  ELIGIBLE: 'ELIGIBLE',
  NOT_ELIGIBLE: 'NOT_ELIGIBLE',
} as const

export function surveyStatusBadge(status: string) {
  switch (status) {
    case SURVEY_STATUS.ELIGIBLE:
      return {
        label: 'Layak (ELIGIBLE)',
        className:
          'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
      }
    case SURVEY_STATUS.NOT_ELIGIBLE:
      return {
        label: 'Tidak Layak (NOT_ELIGIBLE)',
        className:
          'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30',
      }
    default:
      return { label: status || '-', className: 'bg-muted text-muted-foreground' }
  }
}

export function registrationStatusBadge(status: string) {
  switch (status) {
    case REGISTRATION_STATUS.PENDING:
      return {
        label: 'Menunggu Review',
        className:
          'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30',
      }
    case REGISTRATION_STATUS.SURVEY:
      return {
        label: 'Survey Lokasi',
        className:
          'bg-cyan-500/15 text-cyan-700 dark:text-cyan-400 border-cyan-500/30',
      }
    case REGISTRATION_STATUS.APPROVED:
      return {
        label: 'Jadwal Pasang',
        className:
          'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
      }
    case REGISTRATION_STATUS.ON_PROGRESS:
      return {
        label: 'Sedang Dikerjakan',
        className:
          'bg-indigo-500/15 text-indigo-700 dark:text-indigo-400 border-indigo-500/30',
      }
    case REGISTRATION_STATUS.INSTALLED:
      return {
        label: 'Terpasang Fisik',
        className:
          'bg-purple-500/15 text-purple-700 dark:text-purple-400 border-purple-500/30',
      }
    case REGISTRATION_STATUS.ACTIVE:
      return {
        label: 'Aktif / Berlangganan',
        className:
          'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
      }
    case REGISTRATION_STATUS.REJECTED:
      return {
        label: 'Ditolak',
        className:
          'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30',
      }
    case REGISTRATION_STATUS.CANCELLED:
      return {
        label: 'Dibatalkan',
        className:
          'bg-zinc-500/15 text-zinc-700 dark:text-zinc-400 border-zinc-500/30',
      }
    default:
      return {
        label: status || 'Unknown',
        className: 'bg-muted text-muted-foreground',
      }
  }
}
