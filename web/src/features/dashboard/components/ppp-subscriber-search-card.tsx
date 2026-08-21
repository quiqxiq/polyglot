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
import { usePPPSecretsQuery } from '@/features/ppp/api/use-ppp-secrets'
import { useStreamPPPActiveSessions } from '@/features/ppp/api/use-ppp-stream'
import {
  ArrowUpRight,
  Clock,
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
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'inactive' | 'disabled'>('all')

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

  // Metrik status hitungan
  const stats = useMemo(() => {
    let active = 0
    let inactive = 0
    let disabled = 0

    subscribers.forEach((s) => {
      if (s.status === 'active') active++
      else if (s.status === 'disabled') disabled++
      else inactive++
    })

    return { total: subscribers.length, active, inactive, disabled }
  }, [subscribers])

  // Filter berdasarkan search query dan tab status
  const filteredSubscribers = useMemo(() => {
    const q = searchQuery.trim().toLowerCase()

    return subscribers.filter((sub) => {
      // Filter status
      if (statusFilter !== 'all' && sub.status !== statusFilter) {
        return false
      }

      // Filter query pencarian (Username, MAC / CallerId, IP, Comment)
      if (!q) return true

      const nameMatch = sub.name.toLowerCase().includes(q)
      const macMatch = sub.callerId.toLowerCase().includes(q)
      const ipMatch = sub.ipAddress.toLowerCase().includes(q)
      const commentMatch = (sub.comment || '').toLowerCase().includes(q)
      const profileMatch = sub.profile.toLowerCase().includes(q)

      return nameMatch || macMatch || ipMatch || commentMatch || profileMatch
    })
  }, [subscribers, searchQuery, statusFilter])

  const isLoading = secretsQuery.isLoading && isActiveLoading

  return (
    <Card className='shadow-xs flex flex-col justify-between overflow-hidden'>
      <div>
        <CardHeader className='flex flex-row items-center justify-between pb-3'>
          <div>
            <CardTitle className='text-base font-semibold flex items-center gap-2'>
              <Network className='size-4 text-primary' />
              Pencarian Subscriber PPPoE
            </CardTitle>
            <CardDescription>
              Cari pelanggan realtime berdasarkan Nama (Username) atau MAC Address
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

          {/* Quick Filter Chips */}
          <div className='flex flex-wrap items-center gap-1.5 text-xs'>
            <Button
              type='button'
              size='sm'
              variant={statusFilter === 'all' ? 'default' : 'outline'}
              className='h-7 px-2.5 text-[11px] font-normal'
              onClick={() => setStatusFilter('all')}
            >
              Semua ({stats.total})
            </Button>
            <Button
              type='button'
              size='sm'
              variant={statusFilter === 'active' ? 'default' : 'outline'}
              className='h-7 px-2.5 text-[11px] font-normal gap-1.5'
              onClick={() => setStatusFilter('active')}
            >
              <span className='size-2 rounded-full bg-emerald-500 shadow-[0_0_6px_rgba(16,185,129,0.7)] animate-pulse' />
              Aktif ({stats.active})
            </Button>
            <Button
              type='button'
              size='sm'
              variant={statusFilter === 'inactive' ? 'default' : 'outline'}
              className='h-7 px-2.5 text-[11px] font-normal gap-1.5'
              onClick={() => setStatusFilter('inactive')}
            >
              <span className='size-2 rounded-full bg-rose-500 shadow-[0_0_5px_rgba(244,63,94,0.5)]' />
              Tidak Aktif ({stats.inactive})
            </Button>
            <Button
              type='button'
              size='sm'
              variant={statusFilter === 'disabled' ? 'default' : 'outline'}
              className='h-7 px-2.5 text-[11px] font-normal gap-1.5'
              onClick={() => setStatusFilter('disabled')}
            >
              <span className='size-2 rounded-full bg-slate-400 dark:bg-slate-600' />
              Disabled ({stats.disabled})
            </Button>
          </div>

          {/* List Hasil Pencarian */}
          <div className='mt-2 rounded-md border bg-muted/20 divide-y divide-border/60 max-h-72 overflow-y-auto'>
            {isLoading ? (
              <div className='p-3 space-y-2.5'>
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
              <div className='flex flex-col items-center justify-center p-6 text-center text-xs text-muted-foreground'>
                <User className='size-7 text-muted-foreground/40 mb-1.5' />
                <p>
                  {searchQuery
                    ? `Tidak ada subscriber yang cocok dengan "${searchQuery}"`
                    : 'Belum ada data subscriber pada router ini.'}
                </p>
              </div>
            ) : (
              filteredSubscribers.slice(0, 15).map((sub) => (
                <div
                  key={sub.id}
                  className={`p-2.5 sm:p-3 flex flex-col sm:flex-row sm:items-center justify-between gap-2 hover:bg-muted/40 transition-colors ${
                    sub.disabled ? 'opacity-60 bg-muted/10' : ''
                  }`}
                >
                  {/* Left: Indicator, Username, Profile */}
                  <div className='flex items-center gap-2.5 min-w-0'>
                    {/* Status Dot */}
                    <div className='shrink-0 flex items-center justify-center'>
                      {sub.status === 'active' ? (
                        <span
                          className='size-2.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)] animate-pulse'
                          title='Aktif / Online'
                        />
                      ) : sub.status === 'disabled' ? (
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

                    {/* Name and Meta */}
                    <div className='min-w-0'>
                      <div className='flex items-center gap-1.5 flex-wrap'>
                        <span className='font-mono font-semibold text-xs sm:text-sm text-foreground truncate'>
                          {sub.name}
                        </span>
                        <Badge variant='outline' className='font-mono text-[10px] px-1 py-0 h-4 font-normal'>
                          {sub.profile}
                        </Badge>
                        <Badge
                          variant='secondary'
                          className={`text-[9px] px-1 py-0 h-3.5 uppercase font-mono ${
                            sub.status === 'active'
                              ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300'
                              : sub.status === 'disabled'
                              ? 'bg-muted text-muted-foreground'
                              : 'bg-rose-500/15 text-rose-700 dark:text-rose-300'
                          }`}
                        >
                          {sub.status === 'active'
                            ? 'Online'
                            : sub.status === 'disabled'
                            ? 'Disabled'
                            : 'Offline'}
                        </Badge>
                      </div>

                      {/* MAC Address / Caller-ID and IP */}
                      <div className='flex items-center gap-2 text-[11px] font-mono text-muted-foreground mt-0.5 flex-wrap'>
                        <div className='flex items-center gap-1' title='MAC Address / Caller ID'>
                          <Network className='size-3 text-muted-foreground/70' />
                          <span>{sub.callerId || 'No MAC'}</span>
                        </div>
                        {sub.ipAddress && (
                          <div className='flex items-center gap-1 text-foreground/80' title='IP Address'>
                            <Globe className='size-3 text-primary/70' />
                            <span>{sub.ipAddress}</span>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Right: Realtime Uptime or Last Logout & Action Link */}
                  <div className='flex items-center justify-between sm:justify-end gap-3 text-xs shrink-0 self-end sm:self-center'>
                    <div className='text-right font-mono text-[11px] text-muted-foreground'>
                      {sub.status === 'active' ? (
                        <div className='flex items-center gap-1 text-emerald-600 dark:text-emerald-400'>
                          <Clock className='size-3 animate-spin' style={{ animationDuration: '4s' }} />
                          <span>{sub.uptime || 'Baru aktif'}</span>
                        </div>
                      ) : (
                        <span className='text-[10px] text-muted-foreground/80'>
                          {sub.lastLogout ? `Logout: ${sub.lastLogout}` : 'Belum pernah login'}
                        </span>
                      )}
                    </div>

                    <Button asChild size='sm' variant='outline' className='h-6 px-2 text-[10px] gap-1 font-mono'>
                      <Link
                        to='/ppp'
                        search={{ tab: sub.status === 'active' ? 'active' : 'secrets' }}
                      >
                        Detail <ArrowUpRight className='size-2.5' />
                      </Link>
                    </Button>
                  </div>
                </div>
              ))
            )}
          </div>

          {filteredSubscribers.length > 15 && (
            <p className='text-[11px] text-muted-foreground text-center pt-1'>
              Menampilkan 15 dari {filteredSubscribers.length} subscriber. Buka menu PPP untuk melihat seluruh data.
            </p>
          )}
        </CardContent>
      </div>
    </Card>
  )
}
