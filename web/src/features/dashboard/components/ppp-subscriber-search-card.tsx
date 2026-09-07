import { useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { usePPPSecretsQuery } from '@/features/ppp/api/use-ppp-secrets'
import { useStreamPPPActiveSessions } from '@/features/ppp/api/use-ppp-stream'
import {
  ArrowUpRight,
  Check,
  Clock,
  Copy,
  Globe,
  Network,
  Search,
  User,
  X,
} from 'lucide-react'

interface PPPSubscriberSearchCardProps {
  deviceId: string
}

type SubscriberStatus = 'active' | 'inactive' | 'disabled'

interface UnifiedPPPSubscriber {
  id: string
  name: string
  profile: string
  service: string
  status: SubscriberStatus
  callerId: string
  ipAddress: string
  uptime?: string
  lastLogout?: string
  comment?: string
  disabled: boolean
}

export function PPPSubscriberSearchCard({ deviceId }: PPPSubscriberSearchCardProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedSubscriber, setSelectedSubscriber] = useState<UnifiedPPPSubscriber | null>(null)
  const [copiedField, setCopiedField] = useState<string | null>(null)

  const secretsQuery = usePPPSecretsQuery(deviceId)
  const secrets = secretsQuery.data ?? []

  const { sessions: activeSessions = [], isLoading: isActiveLoading } =
    useStreamPPPActiveSessions(deviceId, Boolean(deviceId))

  // Gabungkan data secret dengan data active sessions secara realtime
  const subscribers = useMemo<UnifiedPPPSubscriber[]>(() => {
    if (!secrets || secrets.length === 0) return []

    // Buat map active sessions by username (name)
    const activeMap = new Map<string, (typeof activeSessions)[0]>()
    activeSessions.forEach((s) => {
      if (s.name) {
        activeMap.set(s.name.toLowerCase(), s)
      }
    })

    return secrets.map((sec) => {
      const active = activeMap.get(sec.name.toLowerCase())
      let status: SubscriberStatus = 'inactive'
      if (sec.disabled) {
        status = 'disabled'
      } else if (active) {
        status = 'active'
      }

      const callerId = active?.callerId || sec.callerId || ''
      const ipAddress = active?.address || sec.remoteAddress || sec.localAddress || ''

      return {
        id: sec.id || sec.name,
        name: sec.name,
        profile: sec.profile || 'default',
        service: sec.service || 'pppoe',
        status,
        callerId,
        ipAddress,
        uptime: active?.uptime,
        lastLogout: sec.lastLoggedOut,
        comment: sec.comment,
        disabled: sec.disabled,
      }
    })
  }, [secrets, activeSessions])

  // Filter hanya berjalan jika ada input query pencarian
  const filteredSubscribers = useMemo(() => {
    const q = searchQuery.trim().toLowerCase()
    if (!q) return []

    return subscribers.filter((sub) => {
      const nameMatch = sub.name.toLowerCase().includes(q)
      const macMatch = sub.callerId.toLowerCase().includes(q)
      const ipMatch = sub.ipAddress.toLowerCase().includes(q)
      const commentMatch = (sub.comment || '').toLowerCase().includes(q)
      const profileMatch = sub.profile.toLowerCase().includes(q)

      return nameMatch || macMatch || ipMatch || commentMatch || profileMatch
    })
  }, [subscribers, searchQuery])

  const handleCopy = (text: string, field: string) => {
    if (navigator?.clipboard) {
      navigator.clipboard.writeText(text)
      setCopiedField(field)
      setTimeout(() => setCopiedField(null), 2000)
    }
  }

  const isSearching = searchQuery.trim().length > 0
  const isLoading = secretsQuery.isLoading && isActiveLoading

  return (
    <>
      <Card className='shadow-xs flex flex-col justify-between overflow-hidden'>
        <div>
          <CardHeader className='flex flex-row items-center justify-between pb-3'>
            <div>
              <CardTitle className='text-base font-semibold flex items-center gap-2'>
                <Network className='size-4 text-primary' />
                Pencarian Subscriber PPPoE
              </CardTitle>
              <CardDescription>
                Cari pelanggan realtime berdasarkan Nama (Username), MAC Address, atau IP
              </CardDescription>
            </div>
            <Button asChild size='sm' variant='ghost' className='h-8 gap-1 text-xs'>
              <Link to='/ppp' search={{ tab: 'secrets' }}>
                Kelola PPP <ArrowUpRight className='size-3.5' />
              </Link>
            </Button>
          </CardHeader>

          <CardContent className='space-y-3 pt-0'>
            {/* Input Pencarian */}
            <div className='relative'>
              <Search className='absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground' />
              <Input
                type='text'
                placeholder='Cari username, MAC address, IP, atau profile...'
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className='h-9 pl-8 pr-8 text-xs sm:text-sm font-mono'
              />
              {searchQuery && (
                <button
                  type='button'
                  onClick={() => setSearchQuery('')}
                  className='absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground'
                  title='Hapus pencarian'
                >
                  <X className='size-4' />
                </button>
              )}
            </div>

            {/* Hasil Pencarian: Hanya tampil saat sedang mencari */}
            {isSearching && (
              <div className='space-y-2 max-h-80 overflow-y-auto pr-0.5 pt-1'>
                {isLoading ? (
                  <div className='p-3 space-y-2.5 rounded-md border bg-muted/20'>
                    {Array.from({ length: 3 }).map((_, i) => (
                      <div key={i} className='flex items-center justify-between gap-3'>
                        <div className='flex items-center gap-2'>
                          <Skeleton className='size-7 rounded-full' />
                          <div className='space-y-1'>
                            <Skeleton className='h-3.5 w-28' />
                            <Skeleton className='h-2.5 w-20' />
                          </div>
                        </div>
                        <Skeleton className='h-4 w-16' />
                      </div>
                    ))}
                  </div>
                ) : filteredSubscribers.length === 0 ? (
                  <div className='flex flex-col items-center justify-center p-6 text-center text-xs text-muted-foreground rounded-md border bg-muted/20'>
                    <User className='size-7 text-muted-foreground/40 mb-1.5' />
                    <p>Tidak ada subscriber yang cocok dengan &quot;{searchQuery}&quot;</p>
                  </div>
                ) : (
                  filteredSubscribers.map((sub) => {
                    const isDisabled = sub.status === 'disabled'
                    const isActive = sub.status === 'active'

                    return (
                      <div
                        key={sub.id}
                        className={`p-2.5 sm:p-3 flex flex-col sm:flex-row sm:items-center justify-between gap-2.5 rounded-lg border transition-all ${
                          isDisabled
                            ? 'border-muted bg-muted/40 text-muted-foreground opacity-60 grayscale'
                            : 'border-border/60 hover:bg-muted/40 hover:border-border'
                        }`}
                      >
                        <div className='flex items-center gap-3 min-w-0'>
                          {/* Warna CUMA pada dot: Hijau untuk aktif, Merah untuk tidak aktif, Abu-abu untuk disable */}
                          <div className='shrink-0 flex items-center justify-center'>
                            {isActive ? (
                              <span
                                className='size-2.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)] animate-pulse'
                                title='Aktif / Online'
                              />
                            ) : isDisabled ? (
                              <span
                                className='size-2.5 rounded-full bg-slate-400 dark:bg-slate-600'
                                title='Disabled (Dinonaktifkan)'
                              />
                            ) : (
                              <span
                                className='size-2.5 rounded-full bg-rose-500 shadow-[0_0_5px_rgba(244,63,94,0.6)]'
                                title='Tidak Aktif / Offline'
                              />
                            )}
                          </div>

                          <div className='min-w-0 space-y-0.5'>
                            <div className='flex items-center gap-1.5 flex-wrap'>
                              <span
                                className={`font-mono font-semibold text-xs sm:text-sm truncate ${
                                  isDisabled ? 'text-muted-foreground' : 'text-foreground'
                                }`}
                              >
                                {sub.name}
                              </span>
                              <Badge
                                variant='outline'
                                className='font-mono text-[10px] px-1.5 py-0 h-4 font-normal'
                              >
                                {sub.profile}
                              </Badge>
                              <Badge
                                variant='secondary'
                                className='text-[9px] px-1.5 py-0 h-4 uppercase font-mono font-normal'
                              >
                                {isActive ? 'Online' : isDisabled ? 'Disabled' : 'Offline'}
                              </Badge>
                            </div>

                            <div className='flex items-center gap-2.5 text-[11px] font-mono text-muted-foreground flex-wrap'>
                              <div className='flex items-center gap-1' title='MAC Address / Caller ID'>
                                <Network className='size-3 text-muted-foreground/70' />
                                <span>{sub.callerId || 'No MAC'}</span>
                              </div>
                              {sub.ipAddress && (
                                <div className='flex items-center gap-1' title='IP Address'>
                                  <Globe className='size-3 text-muted-foreground/70' />
                                  <span>{sub.ipAddress}</span>
                                </div>
                              )}
                            </div>
                          </div>
                        </div>

                        <div className='flex items-center justify-between sm:justify-end gap-3 text-xs shrink-0 self-end sm:self-center'>
                          <div className='text-right font-mono text-[11px] text-muted-foreground flex items-center gap-1'>
                            {isActive ? (
                              <>
                                <Clock className='size-3' />
                                <span>{sub.uptime || 'Baru aktif'}</span>
                              </>
                            ) : isDisabled ? (
                              <span>Akun Nonaktif</span>
                            ) : (
                              <span>{sub.lastLogout ? `Logout: ${sub.lastLogout}` : 'Offline'}</span>
                            )}
                          </div>

                          <Button
                            type='button'
                            size='sm'
                            variant='outline'
                            className='h-7 px-2.5 text-[11px] gap-1 font-mono'
                            onClick={() => setSelectedSubscriber(sub)}
                          >
                            Detail
                          </Button>
                        </div>
                      </div>
                    )
                  })
                )}
              </div>
            )}
          </CardContent>
        </div>
      </Card>

      {/* Modal Dialog Detail Subscriber */}
      <Dialog
        open={Boolean(selectedSubscriber)}
        onOpenChange={(open) => {
          if (!open) {
            setSelectedSubscriber(null)
            setCopiedField(null)
          }
        }}
      >
        <DialogContent className='sm:max-w-md'>
          {selectedSubscriber && (
            <>
              <DialogHeader>
                <div className='flex items-center gap-2'>
                  <DialogTitle className='font-mono text-base font-bold flex items-center gap-2'>
                    <User className='size-4 text-primary' />
                    {selectedSubscriber.name}
                  </DialogTitle>
                  <Badge
                    variant='outline'
                    className='text-[10px] px-1.5 py-0 h-4 uppercase font-mono font-normal flex items-center gap-1.5'
                  >
                    <span
                      className={`size-1.5 rounded-full ${
                        selectedSubscriber.status === 'active'
                          ? 'bg-emerald-500 animate-pulse'
                          : selectedSubscriber.status === 'disabled'
                          ? 'bg-slate-400'
                          : 'bg-rose-500'
                      }`}
                    />
                    {selectedSubscriber.status === 'active'
                      ? 'Online'
                      : selectedSubscriber.status === 'disabled'
                      ? 'Disabled'
                      : 'Offline'}
                  </Badge>
                </div>
                <DialogDescription className='text-xs'>
                  Informasi rinci subscriber PPPoE pada router terpilih.
                </DialogDescription>
              </DialogHeader>

              {/* Status Banner */}
              <div className='space-y-3 py-1 text-xs'>
                {selectedSubscriber.status === 'active' ? (
                  <div className='flex items-center gap-2.5 p-2.5 rounded-md border border-border bg-muted/40 text-foreground'>
                    <span className='size-2.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)] animate-pulse shrink-0' />
                    <div className='min-w-0 flex-1'>
                      <div className='font-semibold'>Sedang Terhubung (Aktif)</div>
                      <div className='text-[11px] font-mono text-muted-foreground flex items-center gap-1 mt-0.5'>
                        <Clock className='size-3' />
                        <span>Uptime: {selectedSubscriber.uptime || 'Baru aktif'}</span>
                      </div>
                    </div>
                  </div>
                ) : selectedSubscriber.status === 'disabled' ? (
                  <div className='flex items-center gap-2.5 p-2.5 rounded-md border border-muted bg-muted/50 text-muted-foreground'>
                    <span className='size-2.5 rounded-full bg-slate-400 shrink-0' />
                    <div className='min-w-0 flex-1'>
                      <div className='font-semibold'>Akun Dinonaktifkan (Disabled)</div>
                      <div className='text-[11px]'>Akun ini dinonaktifkan di MikroTik dan tidak dapat dial.</div>
                    </div>
                  </div>
                ) : (
                  <div className='flex items-center gap-2.5 p-2.5 rounded-md border border-border bg-muted/40 text-foreground'>
                    <span className='size-2.5 rounded-full bg-rose-500 shrink-0' />
                    <div className='min-w-0 flex-1'>
                      <div className='font-semibold'>Tidak Terhubung (Offline)</div>
                      <div className='text-[11px] font-mono text-muted-foreground mt-0.5'>
                        {selectedSubscriber.lastLogout ? `Logout terakhir: ${selectedSubscriber.lastLogout}` : 'Belum pernah login'}
                      </div>
                    </div>
                  </div>
                )}

                {/* Grid Rincian Informasi */}
                <div className='rounded-md border divide-y divide-border/60 bg-muted/20 font-mono'>
                  <div className='p-2.5 flex items-center justify-between'>
                    <span className='text-muted-foreground'>Profile Paket</span>
                    <Badge variant='outline' className='font-mono text-xs'>
                      {selectedSubscriber.profile}
                    </Badge>
                  </div>

                  <div className='p-2.5 flex items-center justify-between'>
                    <span className='text-muted-foreground'>Service</span>
                    <span className='font-semibold uppercase'>{selectedSubscriber.service}</span>
                  </div>

                  <div className='p-2.5 flex items-center justify-between'>
                    <span className='text-muted-foreground'>IP Address</span>
                    <div className='flex items-center gap-1.5'>
                      <span className={selectedSubscriber.ipAddress ? 'font-semibold text-foreground' : 'text-muted-foreground'}>
                        {selectedSubscriber.ipAddress || '-'}
                      </span>
                      {selectedSubscriber.ipAddress && (
                        <button
                          type='button'
                          onClick={() => handleCopy(selectedSubscriber.ipAddress, 'ip')}
                          className='p-1 rounded hover:bg-muted text-muted-foreground hover:text-foreground'
                          title='Salin IP'
                        >
                          {copiedField === 'ip' ? (
                            <Check className='size-3 text-emerald-500' />
                          ) : (
                            <Copy className='size-3' />
                          )}
                        </button>
                      )}
                    </div>
                  </div>

                  <div className='p-2.5 flex items-center justify-between'>
                    <span className='text-muted-foreground'>Caller ID / MAC</span>
                    <div className='flex items-center gap-1.5'>
                      <span className={selectedSubscriber.callerId ? 'font-semibold text-foreground' : 'text-muted-foreground'}>
                        {selectedSubscriber.callerId || '-'}
                      </span>
                      {selectedSubscriber.callerId && (
                        <button
                          type='button'
                          onClick={() => handleCopy(selectedSubscriber.callerId, 'mac')}
                          className='p-1 rounded hover:bg-muted text-muted-foreground hover:text-foreground'
                          title='Salin MAC'
                        >
                          {copiedField === 'mac' ? (
                            <Check className='size-3 text-emerald-500' />
                          ) : (
                            <Copy className='size-3' />
                          )}
                        </button>
                      )}
                    </div>
                  </div>

                  {selectedSubscriber.comment && (
                    <div className='p-2.5 flex flex-col gap-1'>
                      <span className='text-muted-foreground'>Catatan / Komentar</span>
                      <span className='text-foreground break-all'>{selectedSubscriber.comment}</span>
                    </div>
                  )}
                </div>
              </div>

              <DialogFooter className='flex-row items-center justify-between gap-2 sm:justify-between pt-2'>
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  onClick={() => setSelectedSubscriber(null)}
                >
                  Tutup
                </Button>

                <Button asChild size='sm' className='gap-1.5 font-mono text-xs'>
                  <Link
                    to='/ppp'
                    search={{
                      tab: selectedSubscriber.status === 'active' ? 'active' : 'secrets',
                      filter: selectedSubscriber.name,
                    }}
                    onClick={() => setSelectedSubscriber(null)}
                  >
                    Kelola di Menu PPP <ArrowUpRight className='size-3.5' />
                  </Link>
                </Button>
              </DialogFooter>
            </>
          )}
        </DialogContent>
      </Dialog>
    </>
  )
}
