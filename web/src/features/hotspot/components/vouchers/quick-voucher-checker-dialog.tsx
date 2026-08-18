import { useState } from 'react'
import {
  Search,
  AlertCircle,
  Clock,
  HardDrive,
  Lock,
  Wifi,
  Printer,
  RotateCcw,
  UserX,
  ShieldAlert,
  ShieldCheck,
  Radio,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { useCheckVoucherStatusMutation } from '../../api/use-hotspot-checker'
import {
  useResetHotspotUserCountersMutation,
  useUpdateHotspotUserMutation,
  useDeleteHotspotUserMutation,
} from '../../api/use-hotspot-users'
import { useKickHotspotSessionMutation } from '../../api/use-hotspot-sessions'
import { useHotspot } from '../../context/hotspot-context'
import { useDeviceStore } from '@/stores/device-store'
import type { CheckVoucherStatusResponse } from '@/gen/v1/hotspot_pb'

type QuickVoucherCheckerDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function QuickVoucherCheckerDialog({
  open,
  onOpenChange,
}: QuickVoucherCheckerDialogProps) {
  const [username, setUsername] = useState('')
  const [result, setResult] = useState<CheckVoucherStatusResponse | null>(null)
  const { selectedDeviceId } = useDeviceStore()
  const { setOpen, setPrintSingleUserId } = useHotspot()

  const checkMutation = useCheckVoucherStatusMutation()
  const kickMutation = useKickHotspotSessionMutation()
  const resetMutation = useResetHotspotUserCountersMutation()
  const updateMutation = useUpdateHotspotUserMutation()
  const deleteMutation = useDeleteHotspotUserMutation()

  const handleSearch = async (e?: React.FormEvent) => {
    if (e) e.preventDefault()
    if (!selectedDeviceId || !username.trim()) return

    try {
      const res = await checkMutation.mutateAsync({
        deviceId: selectedDeviceId,
        username: username.trim(),
      })
      setResult(res)
    } catch (err: any) {
      toast.error(err?.message || 'Failed to check voucher status.')
    }
  }

  const handleKick = async () => {
    if (!selectedDeviceId || !result?.activeSession) return

    toast.promise(
      kickMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: result.activeSession.id,
      }),
      {
        loading: 'Disconnecting active session...',
        success: 'Session kicked successfully.',
        error: 'Failed to kick session.',
      }
    )
    handleSearch()
  }

  const handleResetCounters = async () => {
    if (!selectedDeviceId || !result?.user) return

    toast.promise(
      resetMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: result.user.id,
      }),
      {
        loading: 'Resetting voucher counters...',
        success: 'Counters reset successfully.',
        error: 'Failed to reset counters.',
      }
    )
    handleSearch()
  }

  const handleToggleDisabled = async () => {
    if (!selectedDeviceId || !result?.user) return

    const newDisabled = !result.user.disabled
    toast.promise(
      updateMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: result.user.id,
        name: result.user.name,
        password: result.user.password,
        profile: result.user.profile,
      }),
      {
        loading: `${newDisabled ? 'Disabling' : 'Enabling'} user...`,
        success: `User ${newDisabled ? 'disabled' : 'enabled'}.`,
        error: 'Failed to update user status.',
      }
    )
    handleSearch()
  }

  const handleDeleteUser = async () => {
    if (!selectedDeviceId || !result?.user) return

    toast.promise(
      deleteMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: result.user.id,
      }),
      {
        loading: 'Deleting user...',
        success: 'User deleted from router.',
        error: 'Failed to delete user.',
      }
    )
    setResult(null)
    setUsername('')
  }

  const handlePrint = () => {
    if (!result?.user) return
    setPrintSingleUserId(result.user.id)
    onOpenChange(false)
    setOpen('voucher-print')
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[580px]'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <Search className='size-5 text-primary' />
            Quick Voucher Checker
          </DialogTitle>
          <DialogDescription>
            Inspect live validity, data usage, uptime, and lock status for any voucher code.
          </DialogDescription>
        </DialogHeader>

        {/* Search Input Bar */}
        <form onSubmit={handleSearch} className='flex items-center gap-2 pt-2'>
          <div className='relative flex-1'>
            <Search className='absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground' />
            <Input
              placeholder='Enter Voucher Code / Username...'
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className='pl-9 font-mono text-sm h-10'
              autoFocus
            />
          </div>
          <Button type='submit' disabled={checkMutation.isPending || !username.trim()} className='h-10 px-4'>
            {checkMutation.isPending ? 'Searching...' : 'Inspect'}
          </Button>
        </form>

        {/* Status Display Area */}
        {result && (
          <div className='space-y-4 pt-2'>
            {!result.found ? (
              <div className='flex flex-col items-center justify-center p-6 text-center rounded-lg border border-dashed text-muted-foreground'>
                <AlertCircle className='size-8 text-amber-500 mb-2' />
                <p className='text-sm font-semibold'>Voucher Not Found</p>
                <p className='text-xs mt-1'>No user matching "{username}" was found on this router.</p>
              </div>
            ) : (
              <>
                {/* Header Summary Card */}
                <Card className='border-primary/20 bg-muted/30'>
                  <CardHeader className='p-4 pb-2'>
                    <div className='flex items-center justify-between'>
                      <div className='flex items-center gap-2'>
                        <span className='text-lg font-bold font-mono tracking-tight text-foreground'>
                          {result.user?.name}
                        </span>
                        {result.user?.password && result.user.password !== result.user.name && (
                          <Badge variant='outline' className='font-mono text-xs'>
                            Pass: {result.user.password}
                          </Badge>
                        )}
                      </div>
                      <Badge
                        variant={
                          result.status === 'active'
                            ? 'default'
                            : result.status === 'expired'
                              ? 'destructive'
                              : 'secondary'
                        }
                        className={`capitalize font-semibold text-xs ${
                          result.status === 'active' ? 'bg-emerald-600 hover:bg-emerald-600' : ''
                        }`}
                      >
                        {result.isOnline && <span className='size-2 rounded-full bg-white mr-1.5 animate-pulse' />}
                        {result.status}
                      </Badge>
                    </div>
                  </CardHeader>
                  <CardContent className='p-4 pt-2 grid grid-cols-2 gap-3 text-xs'>
                    <div>
                      <span className='text-muted-foreground'>Profile: </span>
                      <strong className='text-foreground'>{result.user?.profile || '-'}</strong>
                    </div>
                    <div>
                      <span className='text-muted-foreground'>Server: </span>
                      <strong className='text-foreground'>{result.user?.server || 'all'}</strong>
                    </div>
                    <div>
                      <span className='text-muted-foreground'>Validity / Expire: </span>
                      <strong className='text-foreground'>{result.expireDate || result.profile?.validity || '-'}</strong>
                    </div>
                    <div>
                      <span className='text-muted-foreground'>Comment: </span>
                      <span className='text-muted-foreground font-mono truncate max-w-[150px] inline-block align-bottom'>
                        {result.user?.comment || '-'}
                      </span>
                    </div>
                  </CardContent>
                </Card>

                {/* Metrics Grid */}
                <div className='grid grid-cols-2 gap-3'>
                  <div className='p-3 rounded-lg border bg-card space-y-1'>
                    <div className='flex items-center gap-1.5 text-xs text-muted-foreground font-medium'>
                      <Clock className='size-3.5 text-sky-500' />
                      Uptime / Limit
                    </div>
                    <div className='text-sm font-semibold font-mono'>
                      {result.user?.uptime || '0s'} <span className='text-muted-foreground font-normal'>/ {result.user?.limitUptime || '∞'}</span>
                    </div>
                  </div>

                  <div className='p-3 rounded-lg border bg-card space-y-1'>
                    <div className='flex items-center gap-1.5 text-xs text-muted-foreground font-medium'>
                      <HardDrive className='size-3.5 text-violet-500' />
                      Data Traffic / Limit
                    </div>
                    <div className='text-sm font-semibold font-mono'>
                      {result.user?.bytesIn || '0'} in / {result.user?.bytesOut || '0'} out
                    </div>
                  </div>

                  <div className='p-3 rounded-lg border bg-card space-y-1'>
                    <div className='flex items-center gap-1.5 text-xs text-muted-foreground font-medium'>
                      <Lock className='size-3.5 text-amber-500' />
                      MAC Lock
                    </div>
                    <div className='text-xs font-mono truncate'>
                      {result.macLocked || result.activeSession?.macAddress || 'Unlocked'}
                    </div>
                  </div>

                  <div className='p-3 rounded-lg border bg-card space-y-1'>
                    <div className='flex items-center gap-1.5 text-xs text-muted-foreground font-medium'>
                      <Wifi className='size-3.5 text-emerald-500' />
                      Connected IP
                    </div>
                    <div className='text-xs font-mono truncate'>
                      {result.activeSession?.address || 'Not connected'}
                    </div>
                  </div>
                </div>

                {/* Action Toolbar */}
                <div className='flex flex-wrap items-center gap-2 pt-2 border-t'>
                  {result.isOnline && (
                    <Button
                      variant='outline'
                      size='sm'
                      onClick={handleKick}
                      className='gap-1.5 text-xs h-8 text-amber-600 hover:text-amber-700'
                    >
                      <Radio className='size-3.5' />
                      Kick Session
                    </Button>
                  )}

                  <Button
                    variant='outline'
                    size='sm'
                    onClick={handleResetCounters}
                    className='gap-1.5 text-xs h-8'
                  >
                    <RotateCcw className='size-3.5' />
                    Reset Counters
                  </Button>

                  <Button
                    variant='outline'
                    size='sm'
                    onClick={handlePrint}
                    className='gap-1.5 text-xs h-8'
                  >
                    <Printer className='size-3.5' />
                    Re-print
                  </Button>

                  <Button
                    variant='outline'
                    size='sm'
                    onClick={handleToggleDisabled}
                    className='gap-1.5 text-xs h-8'
                  >
                    {result.user?.disabled ? (
                      <>
                        <ShieldCheck className='size-3.5 text-emerald-500' />
                        Enable
                      </>
                    ) : (
                      <>
                        <ShieldAlert className='size-3.5 text-amber-500' />
                        Disable
                      </>
                    )}
                  </Button>

                  <Button
                    variant='destructive'
                    size='sm'
                    onClick={handleDeleteUser}
                    className='gap-1.5 text-xs h-8 ml-auto'
                  >
                    <UserX className='size-3.5' />
                    Delete
                  </Button>
                </div>
              </>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
