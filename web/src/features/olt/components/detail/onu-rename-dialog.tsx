import { useState } from 'react'
import type { Onu } from '@/gen/v1/olt_pb'
import { Loader2, Save } from 'lucide-react'
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
import { useRenameOnuMutation } from '../../api/use-olt'
import { positionLabel } from '../../lib/format'

interface OnuRenameDialogProps {
  onu: Onu
  onClose: () => void
}

export function OnuRenameDialog({ onu, onClose }: OnuRenameDialogProps) {
  const { t } = useI18n()
  const [name, setName] = useState(onu.name)
  const renameMutation = useRenameOnuMutation()

  const handleSubmit = async () => {
    try {
      const res = await renameMutation.mutateAsync({
        onu,
        name: name.trim(),
      })
      toast.success(t('olt.job_queued', { jobId: res.jobId || '-' }))
      onClose()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('olt.job_failed'))
    }
  }

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !renameMutation.isPending) onClose()
      }}
    >
      <DialogContent className='sm:max-w-md'>
        <DialogHeader className='text-start'>
          <DialogTitle>{t('olt.rename_title')}</DialogTitle>
          <DialogDescription>
            {t('olt.rename_desc', {
              position: positionLabel(onu.ponIndex, onu.onuIndex),
            })}
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-1.5 py-2'>
          <Label htmlFor='onu-name'>{t('olt.new_name')}</Label>
          <Input
            id='onu-name'
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') void handleSubmit()
            }}
            autoFocus
          />
        </div>

        <DialogFooter>
          <Button
            variant='outline'
            onClick={onClose}
            disabled={renameMutation.isPending}
          >
            {t('common.cancel')}
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={renameMutation.isPending || !name.trim()}
          >
            {renameMutation.isPending ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : (
              <Save className='h-4 w-4' />
            )}
            {t('common.save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
