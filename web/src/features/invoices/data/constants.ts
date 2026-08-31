export const INVOICE_STATUS_META: Record<string, { label: string; className: string }> = {
  UNPAID: {
    label: 'Belum Bayar',
    className: 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30',
  },
  PARTIAL: {
    label: 'Sebagian',
    className: 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
  },
  PAID: {
    label: 'Lunas',
    className: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
  },
  OVERDUE: {
    label: 'Jatuh Tempo',
    className: 'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30',
  },
  CANCELLED: {
    label: 'Dibatalkan',
    className: 'bg-slate-500/15 text-slate-700 dark:text-slate-400 border-slate-500/30',
  },
}

export const SCAN_METHODS = [
  { value: 'CODE_INPUT', label: 'Input Kode Bayar Manual' },
  { value: 'QR_SCAN', label: 'Scan / Input QRIS QR Code' },
  { value: 'MANUAL', label: 'Input Manual Kasir' },
  { value: 'PAYMENT_GATEWAY', label: 'Payment Gateway Webhook' },
] as const

export const ITEM_TYPE_META: Record<string, string> = {
  SUBSCRIPTION_FEE: 'Biaya Langganan Internet',
  INSTALLATION_FEE: 'Biaya Pasang / Instalasi',
  AD_HOC: 'Biaya Tambahan / Ad-Hoc',
}

export function invoiceStatusBadge(status: string) {
  return INVOICE_STATUS_META[status] || { label: status, className: 'bg-slate-100 text-slate-800' }
}
