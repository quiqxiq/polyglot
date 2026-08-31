export const ACCOUNT_TYPES = [
  { value: 'CASH', label: 'Kas Fisik / Kasir Tunai' },
  { value: 'BANK', label: 'Rekening Bank / Transfer' },
] as const

export const CATEGORY_TYPES = [
  { value: 'INCOME', label: 'Pendapatan (Income)' },
  { value: 'EXPENSE', label: 'Pengeluaran (Expense)' },
] as const

export const DIRECTION_TYPES = [
  { value: 'IN', label: 'Masuk (Debet)' },
  { value: 'OUT', label: 'Keluar (Kredit)' },
] as const

export const TRANSACTION_SOURCE_TYPES = [
  { value: 'PAYMENT', label: 'Pembayaran Tagihan' },
  { value: 'EXPENSE', label: 'Biaya Operasional' },
  { value: 'TRANSFER', label: 'Mutasi Antar Kas' },
] as const

export const DIRECTION_META: Record<string, { label: string; className: string }> = {
  IN: {
    label: 'Masuk',
    className: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
  },
  OUT: {
    label: 'Keluar',
    className: 'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30',
  },
}

export const SOURCE_TYPE_META: Record<string, { label: string; className: string }> = {
  PAYMENT: {
    label: 'Tagihan',
    className: 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
  },
  EXPENSE: {
    label: 'Operasional',
    className: 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30',
  },
  TRANSFER: {
    label: 'Mutasi Kas',
    className: 'bg-purple-500/15 text-purple-700 dark:text-purple-400 border-purple-500/30',
  },
}

export function directionBadge(direction: string) {
  return DIRECTION_META[direction] || { label: direction, className: 'bg-slate-100 text-slate-800' }
}

export function sourceTypeBadge(sourceType: string) {
  return SOURCE_TYPE_META[sourceType] || { label: sourceType, className: 'bg-slate-100 text-slate-800' }
}
