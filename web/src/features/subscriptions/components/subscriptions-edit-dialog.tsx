import { useEffect } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
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
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { SelectDropdown } from '@/components/select-dropdown'
import {
  UpdateSubscriptionRequest,
  type Subscription,
} from '@/gen/v1/billing_pb'
import { useUpdateSubscriptionMutation } from '@/features/billing/api/use-billing'
import { useDeviceStore } from '@/stores/device-store'

const editSubscriptionSchema = z.object({
  remoteUsername: z.string(),
  remotePassword: z.string(),
  customPrice: z.coerce.number().min(0),
  billingCycle: z.string(),
  billingDay: z.coerce.number().int().min(1).max(31),
  notes: z.string(),
})

type EditSubscriptionValues = z.infer<typeof editSubscriptionSchema>

interface SubscriptionsEditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow: Subscription
}

export function SubscriptionsEditDialog({
  open,
  onOpenChange,
  currentRow: subscription,
}: SubscriptionsEditDialogProps) {
  const selectedDeviceId = useDeviceStore((s) => s.selectedDeviceId)
  const updateMutation = useUpdateSubscriptionMutation()

  const form = useForm<EditSubscriptionValues>({
    resolver: zodResolver(editSubscriptionSchema) as never,
    defaultValues: {
      remoteUsername: '',
      remotePassword: '',
      customPrice: 0,
      billingCycle: 'MONTHLY',
      billingDay: 1,
      notes: '',
    },
  })

  // Prefill dari currentRow setiap kali dialog dibuka.
  useEffect(() => {
    if (!open || !subscription) return
    form.reset({
      remoteUsername: subscription.remoteUsername ?? '',
      remotePassword: '',
      customPrice: Number(subscription.customPrice) || 0,
      billingCycle: subscription.billingCycle || 'MONTHLY',
      billingDay: subscription.billingDay || 1,
      notes: subscription.notes || '',
    })
  }, [open, subscription, form])

  const handleSubmit = async (values: EditSubscriptionValues) => {
    try {
      await updateMutation.mutateAsync(
        new UpdateSubscriptionRequest({
          id: subscription.id,
          remoteUsername: values.remoteUsername,
          // Kosong = password tidak diubah di backend.
          remotePassword: values.remotePassword,
          customPrice: values.customPrice,
          billingCycle: values.billingCycle,
          billingDay: values.billingDay,
          deviceId: selectedDeviceId || subscription.deviceId || '',
          notes: values.notes,
        })
      )
      toast.success('Langganan diperbarui')
      onOpenChange(false)
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Gagal memperbarui langganan'
      )
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>Edit Langganan</DialogTitle>
          <DialogDescription>
            Perbarui detail langganan{' '}
            <span className='font-mono text-xs'>
              {subscription.remoteUsername || subscription.id}
            </span>
            . Biarkan password kosong bila tidak ingin mengubahnya.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='subscriptions-edit-form'
            onSubmit={form.handleSubmit(handleSubmit)}
            className='flex flex-col gap-4 py-2'
          >
            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
              <FormField
                control={form.control}
                name='remoteUsername'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Username</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='remotePassword'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password Baru</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        type='password'
                        placeholder='kosongkan = tidak diubah'
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className='grid grid-cols-1 gap-4 sm:grid-cols-3'>
              <FormField
                control={form.control}
                name='customPrice'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Harga Custom</FormLabel>
                    <FormControl>
                      <Input {...field} type='number' step='any' min={0} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='billingCycle'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Siklus</FormLabel>
                    <FormControl>
                      <SelectDropdown
                        isControlled
                        defaultValue={field.value}
                        onValueChange={field.onChange}
                        items={['MONTHLY', 'QUARTERLY', 'YEARLY'].map((c) => ({
                          label: c,
                          value: c,
                        }))}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='billingDay'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Tgl Tagihan</FormLabel>
                    <FormControl>
                      <Input {...field} type='number' min={1} max={31} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='notes'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Catatan</FormLabel>
                  <FormControl>
                    <Textarea {...field} rows={2} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </form>
        </Form>

        <DialogFooter>
          <Button
            type='submit'
            form='subscriptions-edit-form'
            disabled={updateMutation.isPending}
          >
            {updateMutation.isPending ? 'Menyimpan...' : 'Simpan'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
