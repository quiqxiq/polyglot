import { useState } from 'react'
import { CheckCircle2, AlertTriangle, Pause, Trash2, Clock, ShieldCheck } from 'lucide-react'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  useExpireMonitorStatusQuery,
  useSetupExpireMonitorMutation,
  useDisableExpireMonitorMutation,
  useRemoveExpireMonitorMutation,
} from '../../api/use-hotspot-expire'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function ExpireMonitorDialog() {
  const { open, setOpen } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()
  const [interval, setInterval] = useState('00:01:00')

  const { data: status, isLoading, refetch } = useExpireMonitorStatusQuery(
    selectedDeviceId,
    open === 'expire-monitor'
  )

  const setupMutation = useSetupExpireMonitorMutation()
  const disableMutation = useDisableExpireMonitorMutation()
  const removeMutation = useRemoveExpireMonitorMutation()

  const isOpen = open === 'expire-monitor'

  const handleSetup = async () => {
    try {
      await setupMutation.mutateAsync({
        deviceId: selectedDeviceId,
        interval: interval.trim() || '00:01:00',
      })
      toast.success('Expire Monitor Scheduler successfully configured!')
      refetch()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to setup expire monitor')
    }
  }

  const handleDisable = async () => {
    try {
      await disableMutation.mutateAsync({ deviceId: selectedDeviceId })
      toast.success('Expire Monitor disabled')
      refetch()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to disable expire monitor')
    }
  }

  const handleRemove = async () => {
    try {
      await removeMutation.mutateAsync({ deviceId: selectedDeviceId })
      toast.success('Expire Monitor scheduler removed')
      refetch()
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to remove expire monitor')
    }
  }

  const isOk = status?.status === 'ok' && status?.isEnabled

  return (
    <Dialog open={isOpen} onOpenChange={(val) => !val && setOpen(null)}>
      <DialogContent className='sm:max-w-[480px]'>
        <DialogHeader>
          <div className='flex items-center gap-2'>
            <Clock className='size-5 text-primary' />
            <DialogTitle>Mikhmon Expire Monitor</DialogTitle>
          </div>
          <DialogDescription>
            Automatic background scheduler to track expired vouchers, lock users, or generate sales reports.
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4 py-2'>
          {/* Status Box */}
          <div className='rounded-lg border p-4 space-y-3 bg-muted/40'>
            <div className='flex items-center justify-between'>
              <span className='text-sm font-medium'>Current Status:</span>
              {isLoading ? (
                <Badge variant='outline'>Checking...</Badge>
              ) : isOk ? (
                <Badge className='bg-emerald-500 hover:bg-emerald-600 gap-1.5'>
                  <CheckCircle2 className='size-3.5' /> Active (OK)
                </Badge>
              ) : status?.isInstalled ? (
                <Badge variant='outline' className='text-amber-500 border-amber-500 gap-1.5'>
                  <AlertTriangle className='size-3.5' /> Disabled / Not Ready
                </Badge>
              ) : (
                <Badge variant='destructive' className='gap-1.5'>
                  <AlertTriangle className='size-3.5' /> Not Installed
                </Badge>
              )}
            </div>

            {status?.schedulerName && (
              <div className='flex justify-between text-xs text-muted-foreground'>
                <span>Scheduler Name:</span>
                <span className='font-mono font-medium text-foreground'>{status.schedulerName}</span>
              </div>
            )}
          </div>

          {/* Config form */}
          <div className='space-y-2'>
            <Label htmlFor='exp-interval' className='text-xs font-semibold'>
              Check Interval (HH:MM:SS)
            </Label>
            <Input
              id='exp-interval'
              value={interval}
              onChange={(e) => setInterval(e.target.value)}
              placeholder='00:01:00'
              className='font-mono text-sm'
            />
            <p className='text-[11px] text-muted-foreground'>
              Default interval is 1 minute (<code className='font-mono'>00:01:00</code>).
            </p>
          </div>
        </div>

        <DialogFooter className='flex flex-col sm:flex-row gap-2 sm:justify-between'>
          {status?.isInstalled && (
            <div className='flex gap-1.5'>
              {status.isEnabled ? (
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  onClick={handleDisable}
                  disabled={disableMutation.isPending}
                  className='text-amber-600'
                >
                  <Pause className='mr-1.5 size-3.5' /> Disable
                </Button>
              ) : null}
              <Button
                type='button'
                variant='outline'
                size='sm'
                onClick={handleRemove}
                disabled={removeMutation.isPending}
                className='text-destructive'
              >
                <Trash2 className='mr-1.5 size-3.5' /> Remove
              </Button>
            </div>
          )}

          <div className='flex gap-2 justify-end'>
            <Button variant='outline' size='sm' onClick={() => setOpen(null)}>
              Close
            </Button>
            <Button
              size='sm'
              onClick={handleSetup}
              disabled={setupMutation.isPending || !selectedDeviceId}
            >
              <ShieldCheck className='mr-1.5 size-3.5' />
              {setupMutation.isPending ? 'Applying...' : status?.isInstalled ? 'Update / Enable' : 'Install & Setup'}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
