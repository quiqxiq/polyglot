import { PORTAL_STORAGE_KEY } from '../data/constants'

// Data tagihan publik (tanpa sesi) dari GET /api/portal/bill.
export interface PublicBill {
  customer_name: string
  invoice_id: string
  invoice_number: string
  period: string
  total: number
  paid_amount: number
  outstanding: number
  due_date: string
  status: string
  manual_payment_code: string
}

export interface ChargeResult {
  gateway?: string
  external_id?: string
  payment_url?: string
  qr_string?: string
  va_number?: string
  status?: string
  amount?: number
}

export interface PortalHttpError extends Error {
  status?: number
}

export function getPortalToken(): string {
  if (typeof window === 'undefined') return ''
  return window.localStorage.getItem(PORTAL_STORAGE_KEY) || ''
}

export function setPortalToken(token: string) {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(PORTAL_STORAGE_KEY, token)
}

export function clearPortalToken() {
  if (typeof window === 'undefined') return
  window.localStorage.removeItem(PORTAL_STORAGE_KEY)
}

function baseUrl() {
  return import.meta.env.VITE_API_URL || 'http://localhost:8080'
}

function toHttpError(status: number, fallback: string, data: { message?: string } | null): PortalHttpError {
  const err: PortalHttpError = new Error(data?.message || fallback)
  err.status = status
  return err
}

// fetchPublicBill mencari tagihan terbuka via kode bayar / no HP / kode portal.
export async function fetchPublicBill(identifier: string): Promise<PublicBill> {
  const res = await fetch(`${baseUrl()}/api/portal/bill?identifier=${encodeURIComponent(identifier)}`)
  if (!res.ok) {
    const data = await res.json().catch(() => null)
    throw toHttpError(res.status, 'Tagihan tidak ditemukan atau sudah lunas', data)
  }
  return res.json()
}

// chargePublicBill membuat tagihan online untuk pelanggan yang SUDAH login
// portal (F5-4): endpoint wajib bearer token sesi pelanggan.
export async function chargePortalInvoice(
  token: string,
  payload: { invoice_id: string; gateway?: string; channel?: string; expire_minutes?: number }
): Promise<ChargeResult> {
  const res = await fetch(`${baseUrl()}/api/portal/charge`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  })
  if (!res.ok) {
    const data = await res.json().catch(() => null)
    throw toHttpError(res.status, 'Gagal membuat tagihan online', data)
  }
  return res.json()
}

// isAuthError melaporkan error 401/403 dari endpoint bersesi.
export function isAuthError(err: unknown): boolean {
  const status = (err as PortalHttpError | undefined)?.status
  return status === 401 || status === 403
}
