export const CUSTOMER_STATUS_META: Record<string, { label: string; className: string }> = {
  ACTIVE: {
    label: 'Active',
    className: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
  },
  ISOLATED: {
    label: 'Isolated',
    className: 'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30',
  },
  SUSPENDED: {
    label: 'Suspended',
    className: 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30',
  },
  TERMINATED: {
    label: 'Terminated',
    className: 'bg-slate-500/15 text-slate-700 dark:text-slate-400 border-slate-500/30',
  },
}

export const IMPORT_FORMATS = [
  { value: 0, label: 'CSV (Comma Separated Values)' },
  { value: 1, label: 'Excel (XLSX Spreadsheet)' },
] as const

export function customerStatusBadge(status: string) {
  return CUSTOMER_STATUS_META[status] || { label: status, className: 'bg-slate-100 text-slate-800' }
}
