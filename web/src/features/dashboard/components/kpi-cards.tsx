import { useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { useHotspotReportsQuery } from '@/features/reports/api/use-reports'
import { useHotspotActiveSessionsQuery } from '@/features/hotspot/api/use-hotspot-sessions'
import { useHotspotUsersQuery } from '@/features/hotspot/api/use-hotspot-users'
import { usePPPActiveSessionsQuery } from '@/features/ppp/api/use-ppp-active'
import { usePPPSecretsQuery } from '@/features/ppp/api/use-ppp-secrets'
import { useWASessionsQuery, useConversationsQuery } from '@/features/chats/api/use-chats'
import { Banknote, MessageCircleMore, Network, Wifi, TrendingUp } from 'lucide-react'

interface KPICardsProps {
  deviceId: string
}

function formatIDR(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

export function KPICards({ deviceId }: KPICardsProps) {
  // 1. Hotspot Sales Reports
  const reportsQuery = useHotspotReportsQuery(deviceId, '', '', '', Boolean(deviceId))
  const reportsData = reportsQuery.data

  const { todayIncome, todayCount, totalMonthIncome, totalMonthCount } = useMemo(() => {
    if (!reportsData?.reports) {
      return {
        todayIncome: 0,
        todayCount: 0,
        totalMonthIncome: reportsData?.totalIncome || 0,
        totalMonthCount: reportsData?.total || 0,
      }
    }

    const now = new Date()
    const months = ['jan', 'feb', 'mar', 'apr', 'may', 'jun', 'jul', 'aug', 'sep', 'oct', 'nov', 'dec']
    const currentMonthShort = months[now.getMonth()]
    const currentDay = now.getDate()
    const currentYear = now.getFullYear()

    let tIncome = 0
    let tCount = 0

    for (const r of reportsData.reports) {
      // Format date di mikrotik report biasanya "aug/20/2026" atau "2026-08-20"
      const dLower = r.date.toLowerCase()
      const isToday =
        dLower.includes(`${currentMonthShort}/${currentDay}/`) ||
        dLower.includes(`${currentMonthShort}/${String(currentDay).padStart(2, '0')}/`) ||
        dLower.includes(`${currentYear}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(currentDay).padStart(2, '0')}`)

      if (isToday) {
        tIncome += r.price
        tCount++
      }
    }

    return {
      todayIncome: tIncome,
      todayCount: tCount,
      totalMonthIncome: reportsData.totalIncome || 0,
      totalMonthCount: reportsData.total || 0,
    }
  }, [reportsData])

  // 2. Hotspot Users
  const activeSessionsQuery = useHotspotActiveSessionsQuery(deviceId, Boolean(deviceId))
  const activeSessions = activeSessionsQuery.data ?? []
  const hotspotUsersQuery = useHotspotUsersQuery(deviceId, '', '', Boolean(deviceId))
  const hotspotUsers = hotspotUsersQuery.data ?? []

  // 3. PPPoE Sessions
  const pppActiveQuery = usePPPActiveSessionsQuery(deviceId)
  const pppActive = pppActiveQuery.data ?? []
  const pppSecretsQuery = usePPPSecretsQuery(deviceId)
  const pppSecrets = pppSecretsQuery.data ?? []

  // 4. WhatsApp & Bot
  const waSessionsQuery = useWASessionsQuery()
  const waSessions = waSessionsQuery.data ?? []
  const connectedSessions = waSessions.filter((s) => s.status === 'connected')
  const conversationsQuery = useConversationsQuery()
  const conversations = conversationsQuery.data ?? []
  const escalatedCount = conversations.filter((c) => c.status === 'escalation').length

  const isLoading =
    reportsQuery.isLoading ||
    activeSessionsQuery.isLoading ||
    pppActiveQuery.isLoading ||
    waSessionsQuery.isLoading

  return (
    <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
      {/* ── 1. Omset Penjualan Hotspot ──────────────────────── */}
      <Card className='shadow-xs hover:border-primary/40 transition-colors'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Omset Hotspot Hari Ini</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'>
            <Banknote className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          {isLoading && !reportsData ? (
            <Skeleton className='h-8 w-32' />
          ) : (
            <>
              <div className='text-2xl font-bold text-foreground'>
                {formatIDR(todayIncome > 0 ? todayIncome : totalMonthIncome)}
              </div>
              <p className='mt-1 text-xs text-muted-foreground flex items-center gap-1'>
                <TrendingUp className='size-3 text-emerald-500' />
                <span>
                  {todayCount > 0
                    ? `${todayCount} voucher terjual hari ini`
                    : `${totalMonthCount} voucher bulan ini (${formatIDR(totalMonthIncome)})`}
                </span>
              </p>
            </>
          )}
        </CardContent>
      </Card>

      {/* ── 2. Hotspot User Online ─────────────────────────── */}
      <Card className='shadow-xs hover:border-primary/40 transition-colors'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>Hotspot Online Users</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-blue-500/15 text-blue-600 dark:text-blue-400'>
            <Wifi className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          {isLoading && !activeSessions ? (
            <Skeleton className='h-8 w-24' />
          ) : (
            <>
              <div className='text-2xl font-bold text-foreground'>
                {activeSessions.length}{' '}
                <span className='text-sm font-normal text-muted-foreground'>User</span>
              </div>
              <p className='mt-1 text-xs text-muted-foreground'>
                {hotspotUsers.length > 0
                  ? `dari ${hotspotUsers.length} total voucher terdaftar`
                  : 'Belum ada data voucher di router ini'}
              </p>
            </>
          )}
        </CardContent>
      </Card>

      {/* ── 3. PPPoE Active Sessions ───────────────────────── */}
      <Card className='shadow-xs hover:border-primary/40 transition-colors'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>PPPoE Sessions Aktif</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-purple-500/15 text-purple-600 dark:text-purple-400'>
            <Network className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          {isLoading && !pppActive ? (
            <Skeleton className='h-8 w-24' />
          ) : (
            <>
              <div className='text-2xl font-bold text-foreground'>
                {pppActive.length}{' '}
                <span className='text-sm font-normal text-muted-foreground'>
                  / {pppSecrets.length || '-'} Secret
                </span>
              </div>
              <p className='mt-1 text-xs text-muted-foreground'>
                {pppSecrets.length > 0
                  ? `${Math.round((pppActive.length / pppSecrets.length) * 100) || 0}% pelanggan rumahan online`
                  : 'Belum ada secret PPPoE'}
              </p>
            </>
          )}
        </CardContent>
      </Card>

      {/* ── 4. WhatsApp AI Assistant & CS ─────────────────── */}
      <Card className='shadow-xs hover:border-primary/40 transition-colors'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-sm font-medium'>WhatsApp AI Assistant</CardTitle>
          <div className='flex size-8 items-center justify-center rounded-md bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'>
            <MessageCircleMore className='size-4' />
          </div>
        </CardHeader>
        <CardContent>
          {isLoading && !waSessions ? (
            <Skeleton className='h-8 w-28' />
          ) : (
            <>
              <div className='text-2xl font-bold text-foreground flex items-center gap-2'>
                {connectedSessions.length > 0 ? (
                  <>
                    <span className='size-2.5 rounded-full bg-emerald-500 animate-pulse' />
                    <span>{connectedSessions.length} Online</span>
                  </>
                ) : (
                  <>
                    <span className='size-2.5 rounded-full bg-muted-foreground' />
                    <span className='text-muted-foreground text-lg'>Offline</span>
                  </>
                )}
              </div>
              <p className='mt-1 text-xs text-muted-foreground'>
                {conversations.length} chat ({escalatedCount} eskalasi CS aktif)
              </p>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
