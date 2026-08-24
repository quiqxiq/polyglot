export const NOTIFICATION_STATUS_META: Record<string, { label: string; className: string }> = {
  QUEUED: {
    label: 'Dalam Antrean',
    className: 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30',
  },
  SENT: {
    label: 'Terkirim',
    className: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
  },
  FAILED: {
    label: 'Gagal Kirim',
    className: 'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30',
  },
}

export const SYSTEM_TEMPLATE_PRESETS = [
  {
    key: 'BILL_REMINDER',
    name: 'Pengingat Tagihan Bulanan',
    sampleVariables: ['customer_name', 'invoice_number', 'period', 'total', 'due_date', 'payment_url'],
  },
  {
    key: 'PAYMENT_RECEIPT',
    name: 'Kuitansi Pembayaran Lunas',
    sampleVariables: ['customer_name', 'payment_no', 'invoice_number', 'amount', 'payment_date'],
  },
  {
    key: 'ISOLATION_NOTICE',
    name: 'Pemberitahuan Isolasi Layanan',
    sampleVariables: ['customer_name', 'invoice_number', 'outstanding', 'payment_url'],
  },
  {
    key: 'TICKET_UPDATE',
    name: 'Update Tiket Gangguan / Teknisi',
    sampleVariables: ['customer_name', 'ticket_no', 'status', 'technician_name'],
  },
] as const

export function notificationStatusBadge(status: string) {
  return NOTIFICATION_STATUS_META[status] || { label: status, className: 'bg-slate-100 text-slate-800' }
}
