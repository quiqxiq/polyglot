export function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

export const REPORT_PERIOD_TYPES = [
  { value: 'DAILY', label: 'Laporan Harian (Daily)' },
  { value: 'MONTHLY', label: 'Laporan Bulanan (Monthly)' },
  { value: 'YEARLY', label: 'Laporan Tahunan (Yearly)' },
] as const
