import { useMemo, useState } from 'react'
import type { Approval } from '@/gen/v1/approval_pb'
import { ShieldCheck } from 'lucide-react'
import { useCan } from '@/hooks/use-can'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useApprovalsQuery } from '../api/use-approvals'
import { ApprovalDetailDialog } from './approval-detail-dialog'

const STATUS_TABS: { value: string; label: string }[] = [
  { value: 'PENDING', label: 'Pending' },
  { value: 'EXECUTED', label: 'Executed' },
  { value: 'FAILED', label: 'Failed' },
  { value: 'REJECTED', label: 'Rejected' },
  { value: '', label: 'Semua' },
]

function statusVariant(
  status: string
): 'default' | 'secondary' | 'destructive' | 'outline' {
  if (status === 'PENDING') return 'default'
  if (status === 'EXECUTED') return 'secondary'
  if (status === 'FAILED' || status === 'REJECTED') return 'destructive'
  return 'outline'
}

function fmtTime(unix: number | bigint | undefined) {
  const n = Number(unix || 0)
  if (!n) return '-'
  return new Date(n * 1000).toLocaleString('id-ID', {
    dateStyle: 'short',
    timeStyle: 'short',
  })
}

export function ApprovalInbox() {
  const [status, setStatus] = useState('PENDING')
  const [deviceId, setDeviceId] = useState('')
  const [selected, setSelected] = useState<Approval | null>(null)
  const canDecide = useCan('approval:decide')
  const { data: devices = [] } = useDevicesQuery()
  const { data, isLoading } = useApprovalsQuery({
    status,
    deviceId,
    limit: 100,
  })
  const approvals = data?.approvals ?? []

  const deviceName = useMemo(() => {
    const m = new Map<string, string>()
    for (const d of devices) m.set(d.id, d.name)
    return m
  }, [devices])

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div>
          <div className='flex items-center gap-2'>
            <ShieldCheck className='h-6 w-6 text-primary' />
            <h1 className='text-2xl font-bold tracking-tight'>
              AI Ops — Approval Inbox
            </h1>
          </div>
          <p className='mt-0.5 text-sm text-muted-foreground'>
            Setiap mutation router wajib disetujui manusia setelah melihat diff
            dan restore point. AI tidak pernah mengeksekusi mutation sendiri.
          </p>
        </div>

        <Card>
          <CardHeader className='flex flex-wrap items-center justify-between gap-3 space-y-0'>
            <div>
              <CardTitle className='text-base'>Approval Requests</CardTitle>
              <CardDescription>
                {data?.total ?? 0} request ditemukan.
              </CardDescription>
            </div>
            <div className='flex flex-wrap items-center gap-2'>
              <Select
                value={deviceId || 'ALL'}
                onValueChange={(v) => setDeviceId(v === 'ALL' ? '' : v)}
              >
                <SelectTrigger className='h-9 w-52'>
                  <SelectValue placeholder='Semua perangkat' />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='ALL'>Semua perangkat</SelectItem>
                  {devices.map((d) => (
                    <SelectItem key={d.id} value={d.id}>
                      {d.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </CardHeader>
          <CardContent className='space-y-3'>
            <Tabs value={status} onValueChange={setStatus}>
              <TabsList className='flex h-auto flex-wrap justify-start'>
                {STATUS_TABS.map((tab) => (
                  <TabsTrigger key={tab.value || 'ALL'} value={tab.value}>
                    {tab.label}
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>

            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Perangkat</TableHead>
                  <TableHead>Command</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Rev</TableHead>
                  <TableHead>Kedaluwarsa</TableHead>
                  <TableHead className='text-right'>Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  <TableRow>
                    <TableCell
                      colSpan={6}
                      className='py-8 text-center text-sm text-muted-foreground'
                    >
                      Memuat approval…
                    </TableCell>
                  </TableRow>
                ) : approvals.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={6}
                      className='py-8 text-center text-sm text-muted-foreground'
                    >
                      Tidak ada approval pada filter ini.
                    </TableCell>
                  </TableRow>
                ) : (
                  approvals.map((a) => (
                    <TableRow key={a.id}>
                      <TableCell>
                        {deviceName.get(a.deviceId) || a.deviceId}
                      </TableCell>
                      <TableCell className='max-w-[320px] truncate font-mono text-xs'>
                        {a.command}
                      </TableCell>
                      <TableCell>
                        <Badge variant={statusVariant(a.status)}>
                          {a.status}
                        </Badge>
                      </TableCell>
                      <TableCell>{a.revision}</TableCell>
                      <TableCell>{fmtTime(a.expiresAtUnix)}</TableCell>
                      <TableCell className='text-right'>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => setSelected(a)}
                        >
                          Review
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </Main>

      {selected && (
        <ApprovalDetailDialog
          approval={selected}
          open
          onOpenChange={(open) => !open && setSelected(null)}
          canDecide={canDecide}
        />
      )}
    </>
  )
}
