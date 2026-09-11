import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  CreditCard,
  LogOut,
  QrCode,
  ReceiptText,
  RefreshCw,
  ShieldAlert,
  Wifi,
  WifiOff,
} from 'lucide-react'
import { toast } from 'sonner'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { PortalLogoutRequest } from '@/gen/v1/portal_pb'
import {
  usePortalLogoutMutation,
  usePortalOverviewQuery,
  usePortalPaymentsQuery,
} from '../api/use-portal'
import {
  chargePortalInvoice,
  clearPortalToken,
  getPortalToken,
  isAuthError,
} from '../api/portal-http'

function formatCurrency(val: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
}

/**
 * Dashboard mandiri pelanggan (F5-2): status layanan, langganan, tagihan
 * menunggak (dengan pembayaran online bersesi), dan riwayat pembayaran.
 */
export function PortalDashboard() {
  const navigate = useNavigate()
  const token = getPortalToken()
  const overview = usePortalOverviewQuery(token)
  const payments = usePortalPaymentsQuery(token, 10)
  const logout = usePortalLogoutMutation()

  const [chargingId, setChargingId] = useState('')
  const [chargeUrls, setChargeUrls] = useState<Record<string, string>>({})

  const handleLogout = () => {
    logout.mutate(new PortalLogoutRequest({ token }), {
      onSettled: () => {
        clearPortalToken()
        navigate({ to: '/portal/login' })
      },
    })
  }

  const goToLogin = () => {
    clearPortalToken()
    navigate({ to: '/portal/login' })
  }

  const handleCharge = async (invoiceId: string, channel: string) => {
    setChargingId(invoiceId)
    try {
      const res = await chargePortalInvoice(token, {
        invoice_id: invoiceId,
        channel,
        expire_minutes: 60,
      })
      if (res.payment_url) {
        setChargeUrls((prev) => ({ ...prev, [invoiceId]: res.payment_url || '' }))
      }
      if (res.va_number) {
        toast.success(`Virtual Account: ${res.va_number}`)
      } else {
        toast.success('Tagihan online berhasil dibuat')
      }
    } catch (err) {
      if (isAuthError(err)) {
        toast.error('Sesi berakhir, silakan login ulang')
        goToLogin()
        return
      }
      toast.error('Gagal membuat tagihan online', {
        description: (err as Error).message,
      })
    } finally {
      setChargingId('')
    }
  }

  if (!token) {
    return null // beforeLoad akan mengalihkan ke /portal/login
  }

  if (overview.isLoading) {
    return (
      <div className='flex min-h-screen items-center justify-center'>
        <RefreshCw className='h-8 w-8 animate-spin text-muted-foreground' />
      </div>
    )
  }

  if (overview.isError) {
    return (
      <div className='flex min-h-screen items-center justify-center p-4'>
        <Card className='w-full max-w-md text-center'>
          <CardHeader>
            <CardTitle className='text-base'>Sesi Portal Tidak Valid</CardTitle>
            <CardDescription>Silakan login ulang untuk melanjutkan.</CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={goToLogin}>Login Ulang</Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  const data = overview.data
  if (!data) return null
  const status = data.status || data.customer?.status || ''
  const isolated = status === 'ISOLATED'

  return (
    <div className='min-h-screen bg-gradient-to-b from-background via-muted/30 to-background p-4 sm:p-6 lg:p-8'>
      <div className='mx-auto w-full max-w-3xl space-y-6'>
        {/* Header */}
        <div className='flex items-center justify-between gap-3'>
          <div>
            <h1 className='text-xl sm:text-2xl font-bold tracking-tight'>Portal Pelanggan</h1>
            <p className='text-xs text-muted-foreground'>
              {data.customer?.customerCode || ''} • {data.customer?.name || ''}
            </p>
          </div>
          <Button variant='outline' size='sm' className='gap-1.5' onClick={handleLogout}>
            <LogOut className='h-4 w-4' /> Keluar
          </Button>
        </div>

        {/* Status layanan */}
        <Card>
          <CardHeader className='pb-3'>
            <CardTitle className='flex items-center gap-2 text-base'>
              {isolated ? (
                <WifiOff className='h-5 w-5 text-red-500' />
              ) : (
                <Wifi className='h-5 w-5 text-emerald-500' />
              )}
              Status Layanan
            </CardTitle>
          </CardHeader>
          <CardContent className='flex flex-wrap items-center gap-3 text-sm'>
            <Badge variant={isolated ? 'destructive' : 'default'} className='uppercase'>
              {status || 'UNKNOWN'}
            </Badge>
            {data.subscription && (
              <span className='text-muted-foreground'>
                {data.subscription.serviceType || 'PPPOE'} • {data.subscription.planId} • Tagihan
                setiap tanggal {data.subscription.billingDay || 1}
              </span>
            )}
          </CardContent>
        </Card>

        {/* Tagihan menunggak */}
        <Card>
          <CardHeader className='pb-3'>
            <CardTitle className='flex items-center gap-2 text-base'>
              <ReceiptText className='h-5 w-5 text-amber-500' />
              Tagihan Belum Dibayar
            </CardTitle>
            <CardDescription>
              Pilih metode pembayaran online; layanan aktif kembali otomatis setelah lunas.
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-4'>
            {data.unpaidInvoices.length === 0 && (
              <p className='text-sm text-muted-foreground'>Tidak ada tagihan menunggak. 🎉</p>
            )}
            {data.unpaidInvoices.map((inv) => (
              <div key={inv.id} className='rounded-lg border p-3 space-y-2'>
                <div className='flex items-start justify-between gap-2'>
                  <div>
                    <p className='text-sm font-semibold'>
                      {inv.invoiceNumber} ({inv.period})
                    </p>
                    <p className='text-xs text-muted-foreground'>Jatuh tempo {inv.dueDate}</p>
                    <p className='text-xs font-mono text-muted-foreground'>
                      Kode bayar: {inv.manualPaymentCode}
                    </p>
                  </div>
                  <p className='font-mono font-bold text-red-600 dark:text-red-400'>
                    {formatCurrency(inv.outstanding)}
                  </p>
                </div>

                {chargeUrls[inv.id] ? (
                  <Button
                    className='w-full gap-2'
                    size='sm'
                    onClick={() => window.open(chargeUrls[inv.id], '_blank')}
                  >
                    <ShieldAlert className='h-4 w-4' /> Buka Halaman Pembayaran
                  </Button>
                ) : (
                  <div className='flex flex-wrap gap-2'>
                    <Button
                      size='sm'
                      className='gap-1.5'
                      disabled={chargingId === inv.id}
                      onClick={() => handleCharge(inv.id, 'QRIS')}
                    >
                      <QrCode className='h-4 w-4' /> Bayar QRIS
                    </Button>
                    <Button
                      size='sm'
                      variant='outline'
                      className='gap-1.5'
                      disabled={chargingId === inv.id}
                      onClick={() => handleCharge(inv.id, 'BRIVA')}
                    >
                      <CreditCard className='h-4 w-4' /> Virtual Account
                    </Button>
                  </div>
                )}
              </div>
            ))}
          </CardContent>
        </Card>

        {/* Riwayat pembayaran */}
        <Card>
          <CardHeader className='pb-3'>
            <CardTitle className='text-base'>Riwayat Pembayaran</CardTitle>
          </CardHeader>
          <CardContent className='space-y-2'>
            {(payments.data ?? []).length === 0 && (
              <p className='text-sm text-muted-foreground'>Belum ada pembayaran tercatat.</p>
            )}
            {(payments.data ?? []).map((pay) => (
              <div key={pay.id}>
                <div className='flex items-center justify-between text-sm'>
                  <span className='text-muted-foreground'>{pay.paymentNo}</span>
                  <span className='font-mono'>{formatCurrency(pay.amount)}</span>
                </div>
                <Separator className='my-1.5' />
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
