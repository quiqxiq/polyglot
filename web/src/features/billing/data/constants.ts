export const INVOICE_STATUS_META: Record<string, { label: string; className: string }> = {
  UNPAID: {
    label: 'Unpaid',
    className: 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30',
  },
  PARTIAL: {
    label: 'Partial',
    className: 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
  },
  PAID: {
    label: 'Paid',
    className: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
  },
  OVERDUE: {
    label: 'Overdue',
    className: 'bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/30',
  },
  CANCELLED: {
    label: 'Cancelled',
    className: 'bg-slate-500/15 text-slate-700 dark:text-slate-400 border-slate-500/30',
  },
}

export const SUBSCRIPTION_STATUS_META: Record<string, { label: string; className: string }> = {
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
  PENDING: {
    label: 'Pending',
    className: 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
  },
}

export const PROVISION_STATUS_META: Record<string, { label: string; className: string }> = {
  OK: {
    label: 'Provisioned',
    className: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400',
  },
  PENDING: {
    label: 'Pending Provision',
    className: 'bg-amber-500/15 text-amber-700 dark:text-amber-400',
  },
  FAILED: {
    label: 'Provision Failed',
    className: 'bg-rose-500/15 text-rose-700 dark:text-rose-400',
  },
  NONE: {
    label: 'No Provision',
    className: 'bg-slate-500/15 text-slate-700 dark:text-slate-400',
  },
}

export const SERVICE_TYPES = [
  { value: 'PPPOE', label: 'PPPoE (Dial-up/Fiber)' },
  { value: 'HOTSPOT', label: 'Hotspot (Voucher/Monthly)' },
  { value: 'DEDICATED', label: 'Dedicated / Direct Lease' },
] as const

export const VALIDITY_MODES = [
  { value: 'CALENDAR', label: 'Calendar Month (1st-30th)' },
  { value: 'UPTIME', label: 'Uptime Duration Accumulation' },
] as const

export const EXPIRE_MODES = [
  { value: '0', label: '0 — Flat Subscription (No Auto Expire)' },
  { value: 'ntf', label: 'ntf — Disable User on Expire' },
  { value: 'ntfc', label: 'ntfc — Disable User + Append Comment' },
  { value: 'rem', label: 'rem — Remove User on Expire' },
  { value: 'remc', label: 'remc — Remove User + Append Comment' },
] as const

export const SCAN_METHODS = [
  { value: 'CODE_INPUT', label: 'Manual Payment Code' },
  { value: 'QR_SCAN', label: 'QRIS / QR Code Scan' },
  { value: 'MANUAL', label: 'Cashier Manual Input' },
  { value: 'PAYMENT_GATEWAY', label: 'Payment Gateway Webhook' },
] as const

export function invoiceStatusBadge(status: string) {
  return INVOICE_STATUS_META[status] || { label: status, className: 'bg-slate-100 text-slate-800' }
}

export function subscriptionStatusBadge(status: string) {
  return SUBSCRIPTION_STATUS_META[status] || { label: status, className: 'bg-slate-100 text-slate-800' }
}
