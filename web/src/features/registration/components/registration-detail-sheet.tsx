import {
  Building,
  CheckCircle2,
  ClipboardCheck,
  ExternalLink,
  History,
  MapPin,
  MessageCircle,
  Package,
  Router,
  Wrench,
  XCircle,
  Zap,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { ScrollArea } from '@/components/ui/scroll-area'
import type { Registration } from '@/gen/v1/registration_pb'
import { usePlansQuery } from '@/features/billing/api/use-plans'
import {
  registrationStatusBadge,
  surveyStatusBadge,
  REGISTRATION_STATUS,
} from '../data/constants'
import { useRegistrationEventsQuery } from '../api/use-registration'
import { useRegistration } from './registration-provider'

interface RegistrationDetailSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  registration: Registration | null
}

function formatUnixDate(unix: bigint | number | undefined) {
  const num = Number(unix || 0)
  if (!num) return '-'
  return new Date(num * 1000).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

export function RegistrationDetailSheet({
  open,
  onOpenChange,
  registration,
}: RegistrationDetailSheetProps) {
  const { setOpen, setCurrentRow } = useRegistration()
  const { data: plans = [] } = usePlansQuery(false)
  const { data: events = [] } = useRegistrationEventsQuery(
    registration?.id || '',
    open,
  )

  if (!registration) return null

  const statusMeta = registrationStatusBadge(registration.status)
  const plan = plans.find((p) => p.id === registration.planId)
  const planLabel = plan ? `${plan.name} — Rp${Number(plan.price).toLocaleString('id-ID')}` : registration.planId

  const cleanPhone = registration.phone.replace(/\D/g, '')
  const waPhone = cleanPhone.startsWith('0') ? `62${cleanPhone.slice(1)}` : cleanPhone

  const handleAction = (dialog: 'schedule' | 'install' | 'convert') => {
    setCurrentRow(registration)
    setOpen(dialog)
  }

  const isPending = registration.status === REGISTRATION_STATUS.PENDING
  const isSurvey = registration.status === REGISTRATION_STATUS.SURVEY
  const isApproved = registration.status === REGISTRATION_STATUS.APPROVED
  const isOnProgress = registration.status === REGISTRATION_STATUS.ON_PROGRESS
  const isInstalled = registration.status === REGISTRATION_STATUS.INSTALLED
  const isActive = registration.status === REGISTRATION_STATUS.ACTIVE

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className='flex w-full flex-col sm:max-w-xl md:max-w-2xl p-0'>
        <SheetHeader className='border-b p-6 pb-4'>
          <div className='flex items-start justify-between gap-4'>
            <div>
              <div className='flex items-center gap-2'>
                <SheetTitle className='text-xl font-bold'>{registration.fullName}</SheetTitle>
                <Badge variant='outline' className={`text-xs ${statusMeta.className}`}>
                  {statusMeta.label}
                </Badge>
              </div>
              <SheetDescription className='mt-1 font-mono text-xs text-muted-foreground'>
                No. Registrasi: {registration.registrationNo}
              </SheetDescription>
            </div>
            {isPending && (
              <div className='flex gap-2'>
                <Button size='sm' variant='outline' onClick={() => setOpen('survey')}>
                  Survey Lokasi
                </Button>
                <Button size='sm' onClick={() => handleAction('schedule')} className='bg-blue-600 hover:bg-blue-700 text-white'>
                  Setujui &amp; Jadwalkan
                </Button>
              </div>
            )}
            {isSurvey && (
              <Button size='sm' onClick={() => handleAction('schedule')} className='bg-blue-600 hover:bg-blue-700 text-white'>
                Setujui &amp; Jadwalkan
              </Button>
            )}
            {isApproved && (
              <Button size='sm' onClick={() => handleAction('install')} className='bg-purple-600 hover:bg-purple-700 text-white'>
                Input Hasil Pasang
              </Button>
            )}
            {isOnProgress && (
              <Button size='sm' onClick={() => handleAction('install')} className='bg-purple-600 hover:bg-purple-700 text-white'>
                Selesai Pasang + Bukti
              </Button>
            )}
            {isInstalled && (
              <Button size='sm' onClick={() => handleAction('convert')} className='bg-emerald-600 hover:bg-emerald-700 text-white gap-1'>
                <Zap className='h-3.5 w-3.5' /> Aktivasi Pelanggan
              </Button>
            )}
          </div>
        </SheetHeader>

        <ScrollArea className='flex-1 p-6 space-y-6'>
          {/* Quick Contact & Plan Cards */}
          <div className='grid grid-cols-1 gap-3 sm:grid-cols-2 mb-6'>
            <Card className='bg-muted/40'>
              <CardContent className='flex items-center gap-3 p-4'>
                <div className='flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-green-500/10 text-green-600 dark:text-green-400'>
                  <MessageCircle className='h-5 w-5' />
                </div>
                <div className='min-w-0 flex-1'>
                  <p className='text-xs text-muted-foreground'>WhatsApp Calon Pelanggan</p>
                  <p className='font-mono text-sm font-semibold truncate'>{registration.phone || '-'}</p>
                </div>
                {registration.phone && (
                  <Button size='sm' variant='ghost' className='h-8 w-8 p-0 text-green-600' asChild>
                    <a href={`https://wa.me/${waPhone}`} target='_blank' rel='noopener noreferrer' title='Chat WhatsApp'>
                      <ExternalLink className='h-4 w-4' />
                    </a>
                  </Button>
                )}
              </CardContent>
            </Card>

            <Card className='bg-muted/40'>
              <CardContent className='flex items-center gap-3 p-4'>
                <div className='flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400'>
                  <Package className='h-5 w-5' />
                </div>
                <div className='min-w-0 flex-1'>
                  <p className='text-xs text-muted-foreground'>Paket yang Dipilih</p>
                  <p className='text-sm font-semibold truncate'>{planLabel}</p>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Location Card */}
          <Card className='mb-6'>
            <CardHeader className='p-4 pb-2'>
              <CardTitle className='text-sm font-semibold flex items-center gap-2'>
                <Building className='h-4 w-4 text-muted-foreground' /> Alamat Pemasangan
              </CardTitle>
            </CardHeader>
            <CardContent className='space-y-3 p-4 pt-2 text-sm'>
              <p>{registration.address || '-'}</p>
              {registration.hasCoordinates && registration.latitude !== undefined && registration.longitude !== undefined && (
                <div className='flex items-center justify-between rounded-md bg-muted/50 p-2.5'>
                  <div className='flex items-center gap-2 font-mono text-xs'>
                    <MapPin className='h-4 w-4 text-rose-500' />
                    <span>
                      {registration.latitude.toFixed(6)}, {registration.longitude.toFixed(6)}
                    </span>
                  </div>
                  <Button size='sm' variant='outline' className='h-7 text-xs' asChild>
                    <a
                      href={`https://www.google.com/maps?q=${registration.latitude},${registration.longitude}`}
                      target='_blank'
                      rel='noopener noreferrer'
                    >
                      Buka di Maps
                    </a>
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Pipeline Details */}
          <div className='space-y-4'>
            <h4 className='text-sm font-bold tracking-tight'>Riwayat & Pipeline Pemasangan</h4>

            {/* Step: Survey Lokasi (PSB v2) */}
            <Card className={isSurvey || isApproved || isOnProgress || isInstalled || isActive ? 'border-cyan-500/40 bg-cyan-500/5' : ''}>
              <CardHeader className='p-4 pb-2'>
                <CardTitle className='text-sm font-semibold flex items-center justify-between'>
                  <span className='flex items-center gap-2'>
                    <ClipboardCheck className='h-4 w-4 text-cyan-600 dark:text-cyan-400' />
                    Survey Lokasi
                  </span>
                  {registration.surveyStatus ? (
                    <Badge variant='outline' className={`text-[10px] ${surveyStatusBadge(registration.surveyStatus).className}`}>
                      {surveyStatusBadge(registration.surveyStatus).label}
                    </Badge>
                  ) : (
                    <Badge variant='outline' className='text-[10px] text-muted-foreground'>Belum Survey</Badge>
                  )}
                </CardTitle>
              </CardHeader>
              <CardContent className='grid grid-cols-2 gap-3 p-4 pt-2 text-xs'>
                <div>
                  <span className='text-muted-foreground'>Tanggal Survey:</span>{' '}
                  <span className='font-semibold'>{formatUnixDate(registration.surveyDateUnix)}</span>
                  {registration.surveyTime && <span> ({registration.surveyTime})</span>}
                </div>
                <div>
                  <span className='text-muted-foreground'>Estimasi Kabel:</span>{' '}
                  <span className='font-semibold'>
                    {registration.estimatedCableLengthMeter
                      ? `${registration.estimatedCableLengthMeter} meter`
                      : '-'}
                  </span>
                </div>
                {registration.surveyorId && (
                  <div>
                    <span className='text-muted-foreground'>Surveyor ID:</span>{' '}
                    <span className='font-semibold'>{registration.surveyorId}</span>
                  </div>
                )}
                {registration.surveyNotes && (
                  <div className='col-span-2 rounded bg-muted p-2 mt-1 text-muted-foreground'>
                    <b>Catatan Survey:</b> {registration.surveyNotes}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Step: Schedule & Technician */}
            <Card className={isApproved || isInstalled || isActive ? 'border-primary/40 bg-primary/5' : ''}>
              <CardHeader className='p-4 pb-2'>
                <CardTitle className='text-sm font-semibold flex items-center justify-between'>
                  <span className='flex items-center gap-2'>
                    <Wrench className='h-4 w-4 text-blue-600 dark:text-blue-400' />
                    Jadwal & Penugasan Teknisi
                  </span>
                  {registration.scheduledInstallDateUnix ? (
                    <Badge variant='outline' className='text-[10px] bg-blue-500/10 text-blue-600 border-blue-500/20'>
                      Terjadwal
                    </Badge>
                  ) : (
                    <Badge variant='outline' className='text-[10px] text-muted-foreground'>Belum Dijadwalkan</Badge>
                  )}
                </CardTitle>
              </CardHeader>
              <CardContent className='grid grid-cols-2 gap-3 p-4 pt-2 text-xs'>
                <div>
                  <span className='text-muted-foreground'>Tanggal Pasang:</span>{' '}
                  <span className='font-semibold'>{formatUnixDate(registration.scheduledInstallDateUnix)}</span>
                  {registration.scheduledInstallTime && <span> ({registration.scheduledInstallTime})</span>}
                </div>
                <div>
                  <span className='text-muted-foreground'>Teknisi ID:</span>{' '}
                  <span className='font-semibold'>{registration.assignedTechnicianId || 'Belum ditugaskan'}</span>
                </div>
                {registration.notes && (
                  <div className='col-span-2 rounded bg-muted p-2 mt-1 text-muted-foreground'>
                    <b>Catatan:</b> {registration.notes}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Step: Physical Installation */}
            <Card className={isInstalled || isActive ? 'border-purple-500/40 bg-purple-500/5' : ''}>
              <CardHeader className='p-4 pb-2'>
                <CardTitle className='text-sm font-semibold flex items-center justify-between'>
                  <span className='flex items-center gap-2'>
                    <Router className='h-4 w-4 text-purple-600 dark:text-purple-400' />
                    Hasil Pemasangan Fisik Lapangan
                  </span>
                  {registration.installedAtUnix ? (
                    <Badge variant='outline' className='text-[10px] bg-purple-500/10 text-purple-600 border-purple-500/20'>
                      Terpasang Fisik
                    </Badge>
                  ) : (
                    <Badge variant='outline' className='text-[10px] text-muted-foreground'>Menunggu Teknisi</Badge>
                  )}
                </CardTitle>
              </CardHeader>
              <CardContent className='grid grid-cols-2 gap-3 p-4 pt-2 text-xs'>
                <div>
                  <span className='text-muted-foreground'>Router BRAS Terpilih:</span>{' '}
                  <span className='font-semibold font-mono'>{registration.targetDeviceName || registration.targetDeviceId || '-'}</span>
                </div>
                <div>
                  <span className='text-muted-foreground'>Waktu Selesai:</span>{' '}
                  <span>{formatUnixDate(registration.installedAtUnix)}</span>
                </div>
                {registration.installationStartedAtUnix ? (
                  <div>
                    <span className='text-muted-foreground'>Mulai Dikerjakan:</span>{' '}
                    <span>{formatUnixDate(registration.installationStartedAtUnix)}</span>
                  </div>
                ) : null}
                {registration.ontMacAddress && (
                  <div>
                    <span className='text-muted-foreground'>MAC ONT:</span>{' '}
                    <span className='font-mono font-semibold'>{registration.ontMacAddress}</span>
                  </div>
                )}
                {registration.ontSerialNumber && (
                  <div>
                    <span className='text-muted-foreground'>SN ONT:</span>{' '}
                    <span className='font-mono font-semibold'>{registration.ontSerialNumber}</span>
                  </div>
                )}
                {registration.installationPhotoUrl && (
                  <div className='col-span-2'>
                    <span className='text-muted-foreground'>Foto Bukti Instalasi:</span>{' '}
                    <a
                      href={registration.installationPhotoUrl}
                      target='_blank'
                      rel='noopener noreferrer'
                      className='text-primary underline break-all'
                    >
                      {registration.installationPhotoUrl}
                    </a>
                  </div>
                )}
                {registration.technicianNotes && (
                  <div className='col-span-2 rounded bg-muted p-2 mt-1 text-muted-foreground'>
                    <b>Catatan Lapangan Teknisi:</b> {registration.technicianNotes}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Step: Activation Artifacts */}
            {isActive && (
              <Card className='border-emerald-500/40 bg-emerald-500/5'>
                <CardHeader className='p-4 pb-2'>
                  <CardTitle className='text-sm font-semibold flex items-center gap-2 text-emerald-600 dark:text-emerald-400'>
                    <CheckCircle2 className='h-4 w-4' />
                    Artefak Aktivasi Pelanggan
                  </CardTitle>
                </CardHeader>
                <CardContent className='grid grid-cols-3 gap-3 p-4 pt-2 text-xs'>
                  <div>
                    <span className='text-muted-foreground'>Customer ID:</span>
                    <p className='font-mono font-bold mt-0.5'>{registration.customerId}</p>
                  </div>
                  <div>
                    <span className='text-muted-foreground'>Subscription ID:</span>
                    <p className='font-mono font-bold mt-0.5'>{registration.subscriptionId}</p>
                  </div>
                  <div>
                    <span className='text-muted-foreground'>Invoice Pertama ID:</span>
                    <p className='font-mono font-bold mt-0.5'>{registration.invoiceId}</p>
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Timeline PSB (registration_events) */}
            <Card>
              <CardHeader className='p-4 pb-2'>
                <CardTitle className='text-sm font-semibold flex items-center gap-2'>
                  <History className='h-4 w-4 text-muted-foreground' />
                  Timeline Perjalanan PSB
                </CardTitle>
              </CardHeader>
              <CardContent className='p-4 pt-2'>
                {events.length === 0 ? (
                  <p className='text-xs text-muted-foreground'>Belum ada event tercatat.</p>
                ) : (
                  <ol className='relative space-y-3 border-l pl-4'>
                    {events.map((ev) => (
                      <li key={ev.id} className='text-xs'>
                        <div className='flex items-center justify-between gap-2'>
                          <span className='font-semibold'>{ev.title || ev.eventType}</span>
                          <span className='text-[10px] text-muted-foreground'>
                            {formatUnixDate(ev.createdAtUnix)}
                          </span>
                        </div>
                        <p className='text-muted-foreground'>{ev.description || ev.eventType}</p>
                        {ev.actorId && (
                          <p className='text-[10px] text-muted-foreground'>oleh user #{ev.actorId}</p>
                        )}
                      </li>
                    ))}
                  </ol>
                )}
              </CardContent>
            </Card>

            {/* Rejection / Cancellation Info */}
            {registration.rejectedReason && (
              <Card className='border-rose-500/40 bg-rose-500/5'>
                <CardHeader className='p-4 pb-2'>
                  <CardTitle className='text-sm font-semibold flex items-center gap-2 text-rose-600'>
                    <XCircle className='h-4 w-4' /> Alasan Penolakan
                  </CardTitle>
                </CardHeader>
                <CardContent className='p-4 pt-2 text-xs text-rose-700 dark:text-rose-400'>
                  {registration.rejectedReason}
                </CardContent>
              </Card>
            )}

            {registration.cancelReason && (
              <Card className='border-zinc-500/40 bg-zinc-500/5'>
                <CardHeader className='p-4 pb-2'>
                  <CardTitle className='text-sm font-semibold flex items-center gap-2 text-zinc-600'>
                    <XCircle className='h-4 w-4' /> Alasan Pembatalan
                  </CardTitle>
                </CardHeader>
                <CardContent className='p-4 pt-2 text-xs text-muted-foreground'>
                  {registration.cancelReason}
                </CardContent>
              </Card>
            )}
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}
