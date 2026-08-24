export const REGISTRATION_STATUS_META: Record<string, { label: string; className: string }> = {
  PENDING: {
    label: 'Pending Review',
    className: 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30',
  },
  APPROVED: {
    label: 'Approved (Ready to Schedule)',
    className: 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
  },
  INSTALLED: {
    label: 'Installed (Ready to Convert)',
    className: 'bg-indigo-500/15 text-indigo-700 dark:text-indigo-400 border-indigo-500/30',
  },
  ACTIVE: {
    label: 'Active Customer',
    className: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
  },
  REJECTED: {
    label: 'Rejected',
    className: 'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30',
  },
  CANCELLED: {
    label: 'Cancelled',
    className: 'bg-slate-500/15 text-slate-700 dark:text-slate-400 border-slate-500/30',
  },
}

export const INSTALL_TIME_SLOTS = [
  { value: '09:00', label: '09:00 WIB (Pagi 1)' },
  { value: '11:00', label: '11:00 WIB (Pagi 2)' },
  { value: '13:00', label: '13:00 WIB (Siang 1)' },
  { value: '15:00', label: '15:00 WIB (Sore 1)' },
] as const

export function registrationStatusBadge(status: string) {
  return REGISTRATION_STATUS_META[status] || { label: status, className: 'bg-slate-100 text-slate-800' }
}
