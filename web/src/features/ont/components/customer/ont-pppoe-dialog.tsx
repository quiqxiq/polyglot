import { useState } from 'react'
import type { Onu } from '@/gen/v1/olt_pb'
import { Loader2, Network } from 'lucide-react'
import { toast } from 'sonner'
import { useI18n } from '@/context/i18n-provider'
import { Button } from '@/components/ui/button'
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
import { positionLabel } from '@/features/olt/lib/format'
import { usePushOntPppoeMutation } from '../../api/use-ont'

interface OntPppoeDialogProps {
  onu: Onu
  onClose: () => void
}

export function OntPppoeDialog({ onu, onClose }: OntPppoeDialogProps) {
  const { t } = useI18n()
  const [username, setUsername] = useState(onu.pppoeUsername || '')
  const [password, setPassword] = useState('')
  const pushMutation = usePushOntPppoeMutation()
  const position = positionLabel(onu.ponIndex, onu.onuIndex)

  const handleSubmit = async () => {
    if (!username.trim()) {
      toast.error(t('ont.pppoe_username'))
      return
    }
    try {
      const res = await pushMutation.mutateAsync({
        ontId: onu.id,
        username: username.trim(),
        password,
      })
      toast.success(t('ont.job_queued', { jobId: res.jobId || '-' }))
      onClose()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('ont.job_failed'))
    }
  }

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !pushMutation.isPending) onClose()
      }}
    >
      <DialogContent className='sm:max-w-md'>
        <DialogHeader className='text-start'>
          <DialogTitle className='flex items-center gap-2'>
            <Network className='h-5 w-5' />
            {t('ont.pppoe_title')}
          </DialogTitle>
          <DialogDescription>
            {t('ont.pppoe_desc', { position })}
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-3 py-1'>
          <div className='space-y-1.5'>
            <Label htmlFor='ont-pppoe-username'>
              {t('ont.pppoe_username')}
            </Label>
            <Input
              id='ont-pppoe-username'
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder='cust01'
            />
          </div>
          <div className='space-y-1.5'>
            <Label htmlFor='ont-pppoe-password'>
              {t('ont.pppoe_password')}
            </Label>
            <Input
              id='ont-pppoe-password'
              type='password'
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete='new-password'
            />
          </div>
        </div>

        <DialogFooter>
          <Button
            variant='outline'
            onClick={onClose}
            disabled={pushMutation.isPending}
          >
            {t('common.cancel')}
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={pushMutation.isPending || !username.trim()}
          >
            {pushMutation.isPending ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : (
              <Network className='h-4 w-4' />
            )}
            {t('common.save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
