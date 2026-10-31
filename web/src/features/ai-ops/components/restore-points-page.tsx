import { useMemo, useState } from 'react'
import { RestoreMode, type RestorePoint } from '@/gen/v1/restore_point_pb'
import { ArchiveRestore, Download, Plus, RotateCcw } from 'lucide-react'
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import {
  downloadRestoreArtifact,
  useCreateRestorePointMutation,
  useDeleteRestorePointMutation,
  useRestoreFromPointMutation,
  useRestorePointsQuery,
} from '../api/use-restore-points'

function formatBytes(n: number | bigint | undefined) {
  const v = Number(n || 0)
  if (v < 1024) return `${v} B`
  if (v < 1024 * 1024) return `${(v / 1024).toFixed(1)} KB`
  return `${(v / (1024 * 1024)).toFixed(2)} MB`
}

function fmtTime(unix: number | bigint | undefined) {
  const n = Number(unix || 0)
  if (!n) return '-'
  return new Date(n * 1000).toLocaleString('id-ID', {
    dateStyle: 'short',
    timeStyle: 'short',
  })
}

export function RestorePointsPage() {
  const [deviceId, setDeviceId] = useState('')
  const [selected, setSelected] = useState<RestorePoint | null>(null)
  const [restoreTarget, setRestoreTarget] = useState<RestorePoint | null>(null)
  const [confirmation, setConfirmation] = useState('')
  const [mode, setMode] = useState<RestoreMode>(RestoreMode.INVERSE)
  const { data: devices = [] } = useDevicesQuery()
  const { data, isLoading } = useRestorePointsQuery({ deviceId, limit: 100 })
  const createPoint = useCreateRestorePointMutation()
  const restorePoint = useRestoreFromPointMutation()
  const deletePoint = useDeleteRestorePointMutation()

  const deviceMap = useMemo(() => {
    const m = new Map<string, string>()
    for (const d of devices) m.set(d.id, d.name)
    return m
  }, [devices])

  const restorePoints = data?.restorePoints ?? []

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <div className='flex items-center gap-2'>
              <ArchiveRestore className='h-6 w-6 text-primary' />
              <h1 className='text-2xl font-bold tracking-tight'>
                AI Ops — Restore Points
              </h1>
            </div>
            <p className='mt-0.5 text-sm text-muted-foreground'>
              Setiap mutation menyimpan scoped export, binary backup, dan
              inverse script terenkripsi. Restore butuh konfirmasi nama
              perangkat.
            </p>
          </div>
          <div className='flex items-center gap-2'>
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
            {deviceId && (
              <Button
                size='sm'
                className='gap-1.5'
                disabled={createPoint.isPending}
                onClick={() => createPoint.mutate(deviceId)}
              >
                <Plus className='h-4 w-4' /> Backup Now
              </Button>
            )}
          </div>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className='text-base'>Daftar Restore Point</CardTitle>
            <CardDescription>
              {data?.total ?? 0} bundle tersimpan.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Perangkat</TableHead>
                  <TableHead>Trigger</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Ukuran</TableHead>
                  <TableHead>Dibuat</TableHead>
                  <TableHead>Kedaluwarsa</TableHead>
                  <TableHead className='text-right'>Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  <TableRow>
                    <TableCell
                      colSpan={7}
                      className='py-8 text-center text-sm text-muted-foreground'
                    >
                      Memuat…
                    </TableCell>
                  </TableRow>
                ) : restorePoints.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={7}
                      className='py-8 text-center text-sm text-muted-foreground'
                    >
                      Belum ada restore point.
                    </TableCell>
                  </TableRow>
                ) : (
                  restorePoints.map((rp) => (
                    <TableRow key={rp.id}>
                      <TableCell>
                        {deviceMap.get(rp.deviceId) || rp.deviceId}
                      </TableCell>
                      <TableCell>
                        <Badge variant='outline'>{rp.trigger}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={
                            rp.status === 'AVAILABLE' ? 'secondary' : 'outline'
                          }
                        >
                          {rp.status}
                        </Badge>
                      </TableCell>
                      <TableCell>{formatBytes(rp.totalSizeBytes)}</TableCell>
                      <TableCell>{fmtTime(rp.createdAtUnix)}</TableCell>
                      <TableCell>{fmtTime(rp.expiresAtUnix)}</TableCell>
                      <TableCell className='text-right'>
                        <div className='flex justify-end gap-1'>
                          <Button
                            variant='ghost'
                            size='sm'
                            onClick={() => setSelected(rp)}
                          >
                            Detail
                          </Button>
                          <Button
                            variant='outline'
                            size='sm'
                            className='gap-1'
                            onClick={() => setRestoreTarget(rp)}
                          >
                            <RotateCcw className='h-3.5 w-3.5' /> Restore
                          </Button>
                        </div>
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
        <Dialog open onOpenChange={(open) => !open && setSelected(null)}>
          <DialogContent className='sm:max-w-lg'>
            <DialogHeader>
              <DialogTitle>Restore Point {selected.id}</DialogTitle>
              <DialogDescription>
                {selected.routerosVersion || '-'} ·{' '}
                {selected.scope.join(', ') || '/'}
              </DialogDescription>
            </DialogHeader>
            <div className='space-y-2'>
              {selected.artifacts.map((a) => (
                <div
                  key={`${a.kind}-${a.sha256}`}
                  className='flex items-center justify-between rounded-md border p-2 text-sm'
                >
                  <div>
                    <p className='font-medium'>{a.kind}</p>
                    <p className='font-mono text-[11px] text-muted-foreground'>
                      {a.sha256.slice(0, 16)}… · {formatBytes(a.sizeBytes)}
                    </p>
                  </div>
                  <Button
                    variant='outline'
                    size='sm'
                    className='gap-1'
                    onClick={() => downloadRestoreArtifact(selected.id, a.kind)}
                  >
                    <Download className='h-3.5 w-3.5' /> Unduh
                  </Button>
                </div>
              ))}
            </div>
            <DialogFooter>
              <Button
                variant='destructive'
                disabled={deletePoint.isPending}
                onClick={async () => {
                  if (
                    !window.confirm(
                      'Hapus restore point ini? File backup terenkripsi akan dihapus.'
                    )
                  )
                    return
                  await deletePoint.mutateAsync(selected.id)
                  setSelected(null)
                }}
              >
                Hapus Restore Point
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {restoreTarget && (
        <Dialog open onOpenChange={(open) => !open && setRestoreTarget(null)}>
          <DialogContent className='sm:max-w-md'>
            <DialogHeader>
              <DialogTitle>Restore dari Titik Ini</DialogTitle>
              <DialogDescription>
                Tindakan ini mengubah konfigurasi router dan tidak bisa
                dibatalkan. Ketik nama perangkat untuk konfirmasi.
              </DialogDescription>
            </DialogHeader>
            <div className='space-y-3'>
              <div className='space-y-1.5'>
                <Label>Mode Restore</Label>
                <Select
                  value={String(mode)}
                  onValueChange={(v) => setMode(Number(v) as RestoreMode)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={String(RestoreMode.INVERSE)}>
                      Inverse script (disarankan)
                    </SelectItem>
                    <SelectItem value={String(RestoreMode.SCOPED_EXPORT)}>
                      Scoped export
                    </SelectItem>
                    <SelectItem value={String(RestoreMode.BINARY)}>
                      Binary backup + reboot
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className='space-y-1.5'>
                <Label>Ketik nama perangkat</Label>
                <Input
                  value={confirmation}
                  onChange={(e) => setConfirmation(e.target.value)}
                  placeholder={
                    deviceMap.get(restoreTarget.deviceId) || 'Nama perangkat'
                  }
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant='outline' onClick={() => setRestoreTarget(null)}>
                Batal
              </Button>
              <Button
                variant='destructive'
                disabled={
                  restorePoint.isPending ||
                  confirmation !== (deviceMap.get(restoreTarget.deviceId) || '')
                }
                onClick={async () => {
                  await restorePoint.mutateAsync({
                    id: restoreTarget.id,
                    mode,
                    confirmation,
                  })
                  setRestoreTarget(null)
                  setConfirmation('')
                }}
              >
                {restorePoint.isPending ? 'Memproses…' : 'Restore Sekarang'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </>
  )
}
