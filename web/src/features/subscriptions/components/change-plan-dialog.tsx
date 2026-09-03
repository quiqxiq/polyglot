import { useState } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  ChangePlanRequest,
  type Subscription,
} from '@/gen/v1/subscription_pb'
import { usePlansQuery } from '@/features/billing/api/use-plans'
import { useChangePlanMutation } from '@/features/billing/api/use-billing'

interface ChangePlanDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow: Subscription | null
}

export function ChangePlanDialog({
  open,
  onOpenChange,
  currentRow: subscription,
}: ChangePlanDialogProps) {
  const { data: plans = [], isPending: plansLoading } = usePlansQuery(true)
  const changePlan = useChangePlanMutation()
  const [newPlanId, setNewPlanId] = useState('')

  const items = plans.map((plan) => ({
    label: `${plan.name} — Rp${Number(plan.price).toLocaleString('id-ID')}`,
    value: plan.id,
  }))

  const handleSubmit = async () => {
    if (!subscription || !newPlanId) return
    try {
      await changePlan.mutateAsync(
        new ChangePlanRequest({
          subscriptionId: subscription.id,
          newPlanId,
        })
      )
      toast.success('Paket langganan berhasil diganti')
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal mengganti paket')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Ganti Paket</DialogTitle>
          <DialogDescription>
            Ganti paket untuk langganan{' '}
            <span className='font-mono text-xs'>
              {subscription?.remoteUsername || subscription?.id}
            </span>
            . Profil router dan rate limit akan diperbarui otomatis.
          </DialogDescription>
        </DialogHeader>

        <div className='flex flex-col gap-2 py-2'>
          <Label>Paket Baru</Label>
          <Select
            value={newPlanId}
            onValueChange={setNewPlanId}
            disabled={plansLoading}
          >
            <SelectTrigger>
              <SelectValue
                placeholder={plansLoading ? 'Memuat paket...' : 'Pilih paket'}
              />
            </SelectTrigger>
            <SelectContent>
              {items.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <DialogFooter>
          <Button
            type='button'
            onClick={handleSubmit}
            disabled={!newPlanId || changePlan.isPending}
          >
            {changePlan.isPending ? 'Menyimpan...' : 'Ganti Paket'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
