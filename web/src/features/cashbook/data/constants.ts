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
