import { useEffect, useMemo, useRef, useState } from 'react'
import {
  MessageCircleMore,
  Plus,
  QrCode,
  RefreshCw,
  LogOut,
  Trash2,
  Smartphone,
  LoaderCircle,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Separator } from '@/components/ui/separator'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { toast } from 'sonner'
import { CreateDeviceDialog } from './components/create-device-dialog'
import { QRModal } from './components/qr-modal'
import {
  useWASessionsQuery,
  useReconnectWASessionMutation,
  useLogoutWASessionMutation,
  usePurgeWASessionMutation,
} from './api/use-whatsapp'
import { useWARealtimeStream } from './api/use-whatsapp-sse'
import { SSEIndicator } from '@/components/sse-indicator'
import type { WASession } from '@/gen/v1/whatsapp_pb'

function statusMeta(status: string): { label: string; className: string; dot: string } {
  switch (status) {
    case 'online':
      return {
        label: 'Online',
        className: 'border-transparent bg-emerald-500/15 text-emerald-600',
        dot: 'bg-emerald-500',
      }
    case 'connecting':
      return {
        label: 'Menghubungkan…',
        className: 'border-transparent bg-amber-500/15 text-amber-600',
        dot: 'bg-amber-500 animate-pulse',
      }
    case 'needs_rescan':
      return {
        label: 'Perlu scan QR',
        className: 'border-transparent bg-orange-500/15 text-orange-600',
        dot: 'bg-orange-500',
      }
    default:
      return {
        label: 'Offline',
        className: 'border-transparent bg-muted text-muted-foreground',
        dot: 'bg-muted-foreground/60',
      }
  }
}

function formatJid(jid: string): string {
  if (!jid) return ''
  return jid.split('@')[0] || jid
}

export function WhatsAppDevices() {
  // Status device & QR ter-update instan via SSE, bukan polling.
  // Status koneksi untuk indikator header.
  const sseStatus = useWARealtimeStream()

  const [createOpen, setCreateOpen] = useState(false)
  // Modal QR memakai ID + flag snapshot (bukan objek session) supaya status
  // bisa di-derive dari cache live — auto-close saat device jadi online tanpa
  // setState dalam effect (hindari react-hooks/set-state-in-effect).
  const [qrSessionId, setQrSessionId] = useState<string | null>(null)
  const [qrOpenedOnline, setQrOpenedOnline] = useState(false)
  const [purgeTarget, setPurgeTarget] = useState<WASession | null>(null)

  const sessionsQuery = useWASessionsQuery()
  const sessions = useMemo(() => sessionsQuery.data ?? [], [sessionsQuery.data])

  // Session live dari query cache (status ter-update via SSE) untuk modal QR.
  const qrSession = useMemo(
    () => (qrSessionId ? sessions.find((s) => s.id === qrSessionId) ?? null : null),
    [sessions, qrSessionId],
  )
  // Auto-close: modal di-render tertutup (session null) saat device yang tadinya
  // offline berubah jadi online via SSE. Modal yang dibuka untuk device
  // sudah-online (qrOpenedOnline) tetap menampilkan state "sudah terhubung".
  const qrAutoClosed = qrSession !== null && !qrOpenedOnline && qrSession.status === 'online'

  const autoCloseNotifiedRef = useRef(false)
  useEffect(() => {
    if (qrAutoClosed && !autoCloseNotifiedRef.current) {
      autoCloseNotifiedRef.current = true
      toast.success(`${qrSession?.name || 'Perangkat'} berhasil ditautkan`)
    } else if (!qrAutoClosed) {
      autoCloseNotifiedRef.current = false
    }
  }, [qrAutoClosed, qrSession?.name])

  const handleOpenQr = (s: WASession) => {
    setQrSessionId(s.id)
    setQrOpenedOnline(s.status === 'online')
  }

  const reconnectMutation = useReconnectWASessionMutation()
  const logoutMutation = useLogoutWASessionMutation()
  const purgeMutation = usePurgeWASessionMutation()

  const handleReconnect = (s: WASession) => {
    reconnectMutation.mutate({ sessionId: s.id }, {
      onSuccess: () => toast.success(`Reconnect ${s.name} dimulai`),
      onError: (err) => toast.error(`Reconnect gagal: ${err.message}`),
    })
  }

  const handleLogout = (s: WASession) => {
    logoutMutation.mutate({ sessionId: s.id }, {
      onSuccess: () => toast.success(`${s.name} di-logout (slot tetap, bisa di-pair ulang)`),
      onError: (err) => toast.error(`Logout gagal: ${err.message}`),
    })
  }

  const handlePurge = () => {
    if (!purgeTarget) return
    purgeMutation.mutate({ sessionId: purgeTarget.id }, {
      onSuccess: () => {
        toast.success(`${purgeTarget.name} dihapus permanen`)
        setPurgeTarget(null)
      },
      onError: (err) => toast.error(`Hapus gagal: ${err.message}`),
    })
  }

  const pendingPurge = purgeMutation.isPending && purgeTarget !== null

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <SSEIndicator status={sseStatus} />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h1 className='flex items-center gap-2 text-2xl font-bold tracking-tight'>
              <MessageCircleMore className='size-6 text-primary' />
              WhatsApp Devices
            </h1>
            <p className='text-sm text-muted-foreground'>
              Hubungkan, scan QR, atau pairing code untuk setiap perangkat WhatsApp.
            </p>
          </div>
          <Button onClick={() => setCreateOpen(true)} className='gap-1.5'>
            <Plus size={16} /> Tambah Perangkat
          </Button>
        </div>

        <Separator className='shadow-xs' />

        {sessionsQuery.isLoading ? (
          <div className='flex flex-1 items-center justify-center py-24 text-muted-foreground'>
            <LoaderCircle className='size-6 animate-spin' />
          </div>
        ) : sessions.length === 0 ? (
          <div className='flex flex-1 flex-col items-center justify-center gap-4 rounded-lg border border-dashed py-24 text-center'>
            <div className='flex size-16 items-center justify-center rounded-full border-2 border-border'>
              <Smartphone className='size-8' />
            </div>
            <div>
              <h2 className='text-lg font-semibold'>Belum ada perangkat</h2>
              <p className='text-sm text-muted-foreground'>
                Tambah perangkat lalu scan QR atau gunakan pairing code untuk menghubungkan.
              </p>
            </div>
            <Button onClick={() => setCreateOpen(true)} className='gap-1.5'>
              <Plus size={16} /> Tambah Perangkat
            </Button>
          </div>
        ) : (
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3'>
            {sessions.map((s) => {
              const meta = statusMeta(s.status)
              const isOnline = s.status === 'online'
              return (
                <Card key={s.id} className='overflow-hidden transition-shadow hover:shadow-md'>
                  <CardHeader className='flex-row items-center justify-between space-y-0'>
                    <CardTitle className='flex items-center gap-2 text-base'>
                      <Smartphone className='size-4 text-muted-foreground' />
                      <span className='truncate'>{s.name || 'Tanpa nama'}</span>
                    </CardTitle>
                    <Badge variant='outline' className={cn('gap-1.5', meta.className)}>
                      <span className={cn('size-2 rounded-full', meta.dot)} />
                      {meta.label}
                    </Badge>
                  </CardHeader>
                  <CardContent className='space-y-3'>
                    <div className='space-y-1 text-sm'>
                      <div className='flex items-center justify-between'>
                        <span className='text-muted-foreground'>Nomor</span>
                        <span className='font-medium'>
                          {s.phoneNumber || formatJid(s.jid) || '—'}
                        </span>
                      </div>
                      <div className='flex items-center justify-between'>
                        <span className='text-muted-foreground'>JID</span>
                        <span className='max-w-40 truncate font-mono text-xs'>
                          {s.jid || '—'}
                        </span>
                      </div>
                      <div className='flex items-center justify-between'>
                        <span className='text-muted-foreground'>Bot</span>
                        <span>{s.isBotActive ? 'Aktif' : 'Nonaktif'}</span>
                      </div>
                      {s.connectedAt && (
                        <div className='flex items-center justify-between'>
                          <span className='text-muted-foreground'>Terhubung sejak</span>
                          <span className='font-medium'>{s.connectedAt}</span>
                        </div>
                      )}
                      <div className='flex items-center justify-between'>
                        <span className='text-muted-foreground'>Dibuat</span>
                        <span className='font-medium'>{s.createdAt}</span>
                      </div>
                    </div>

                    <Separator />

                    <div className='flex flex-wrap gap-2'>
                      {!isOnline ? (
                        <Button
                          size='sm'
                          variant='outline'
                          className='gap-1.5'
                          disabled={reconnectMutation.isPending}
                          onClick={() => handleReconnect(s)}
                        >
                          <RefreshCw size={14} /> Reconnect
                        </Button>
                      ) : null}
                      <Button
                        size='sm'
                        variant='outline'
                        className='gap-1.5'
                        onClick={() => handleOpenQr(s)}
                      >
                        <QrCode size={14} />
                        {s.status === 'needs_rescan' ? 'Scan QR' : 'QR / Pairing'}
                      </Button>
                      <Button
                        size='sm'
                        variant='outline'
                        className='gap-1.5 text-destructive hover:text-destructive'
                        onClick={() => handleLogout(s)}
                        disabled={logoutMutation.isPending}
                      >
                        <LogOut size={14} /> Logout
                      </Button>
                      <Button
                        size='sm'
                        variant='ghost'
                        className='gap-1.5 text-destructive hover:text-destructive'
                        onClick={() => setPurgeTarget(s)}
                      >
                        <Trash2 size={14} /> Hapus
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        )}
      </Main>

      <CreateDeviceDialog open={createOpen} onOpenChange={setCreateOpen} />

      <QRModal
        key={qrSessionId ?? 'closed'}
        session={qrAutoClosed ? null : qrSession}
        onOpenChange={(open) => !open && setQrSessionId(null)}
      />

      <AlertDialog
        open={purgeTarget !== null}
        onOpenChange={(open) => !open && !pendingPurge && setPurgeTarget(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hapus perangkat permanen?</AlertDialogTitle>
            <AlertDialogDescription>
              {purgeTarget?.name} akan di-unlink dari WhatsApp dan semua data session
              (termasuk riwayat chat mirror) dihapus dari database. Tindakan ini tidak dapat
              dibatalkan.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={pendingPurge}>Batal</AlertDialogCancel>
            <AlertDialogAction
              className='bg-destructive text-white hover:bg-destructive/90'
              disabled={pendingPurge}
              onClick={(e) => {
                e.preventDefault()
                handlePurge()
              }}
            >
              {pendingPurge ? 'Menghapus…' : 'Hapus Permanen'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
