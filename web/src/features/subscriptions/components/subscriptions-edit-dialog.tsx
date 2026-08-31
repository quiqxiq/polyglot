import { useEffect } from 'react'
import { z } from 'zod'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
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
  PPPoESubscriptionConfig,
  HotspotSubscriptionConfig,
  type Subscription,
} from '@/gen/v1/billing_pb'
import { useUpdateSubscriptionMutation } from '@/features/billing/api/use-billing'
import { useDeviceStore } from '@/stores/device-store'

const editSubscriptionSchema = z.object({
  remoteUsername: z.string().min(1, 'Username tidak boleh kosong'),
  remotePassword: z.string(),
  customPrice: z.coerce.number().min(0),
  billingCycle: z.string(),
  billingDay: z.coerce.number().int().min(1).max(31),
  notes: z.string(),
  // PPPoE
  localAddress: z.string(),
  remoteAddress: z.string(),
  callerId: z.string(),
  rateLimit: z.string(),
  // Hotspot
  server: z.string(),
  macAddress: z.string(),
  ipAddress: z.string(),
  limitUptime: z.string(),
  limitBytes: z.string(),
})

type EditSubscriptionValues = z.infer<typeof editSubscriptionSchema>

interface SubscriptionsEditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow: Subscription
}

function serviceTypeBadge(serviceType: string) {
  if (serviceType === 'HOTSPOT') {
    return { label: 'Hotspot', className: 'bg-purple-500/15 text-purple-700 dark:text-purple-400 border-purple-500/30' }
  }
  if (serviceType === 'DEDICATED') {
    return { label: 'Dedicated', className: 'bg-orange-500/15 text-orange-700 dark:text-orange-400 border-orange-500/30' }
  }
  return { label: 'PPPoE', className: 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30' }
}

export function SubscriptionsEditDialog({
  open,
  onOpenChange,
  currentRow: subscription,
}: SubscriptionsEditDialogProps) {
  const selectedDeviceId = useDeviceStore((s) => s.selectedDeviceId)
  const updateMutation = useUpdateSubscriptionMutation()

  const isHotspot = subscription.serviceType === 'HOTSPOT'
  const isPPPoE = subscription.serviceType !== 'HOTSPOT'
  const stBadge = serviceTypeBadge(subscription.serviceType)

  const form = useForm<EditSubscriptionValues>({
    resolver: zodResolver(editSubscriptionSchema) as never,
    defaultValues: {
      remoteUsername: '',
      remotePassword: '',
      customPrice: 0,
      billingCycle: 'MONTHLY',
      billingDay: 1,
      notes: '',
      localAddress: '',
      remoteAddress: '',
      callerId: '',
      rateLimit: '',
      server: '',
      macAddress: '',
      ipAddress: '',
      limitUptime: '',
      limitBytes: '',
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
      // PPPoE fields
      localAddress: subscription.pppoeConfig?.localAddress ?? '',
      remoteAddress: subscription.pppoeConfig?.remoteAddress ?? '',
      callerId: subscription.pppoeConfig?.callerId ?? '',
      rateLimit:
        subscription.pppoeConfig?.rateLimit ||
        subscription.hotspotConfig?.rateLimit ||
        subscription.rateLimit ||
        '',
      // Hotspot fields
      server: subscription.hotspotConfig?.server ?? '',
      macAddress: subscription.hotspotConfig?.macAddress ?? '',
      ipAddress: subscription.hotspotConfig?.ipAddress ?? '',
      limitUptime: subscription.hotspotConfig?.limitUptime ?? '',
      limitBytes: subscription.hotspotConfig?.limitBytes ?? '',
    })
  }, [open, subscription, form])

  const handleSubmit = async (values: EditSubscriptionValues) => {
    try {
      const pppoeConfig = isPPPoE
        ? new PPPoESubscriptionConfig({
            localAddress: values.localAddress || '',
            remoteAddress: values.remoteAddress || '',
            callerId: values.callerId || '',
            rateLimit: values.rateLimit || '',
          })
        : undefined

      const hotspotConfig = isHotspot
        ? new HotspotSubscriptionConfig({
            server: values.server || '',
            macAddress: values.macAddress || '',
            ipAddress: values.ipAddress || '',
            limitUptime: values.limitUptime || '',
            limitBytes: values.limitBytes || '',
            rateLimit: values.rateLimit || '',
          })
        : undefined

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
          pppoeConfig,
          hotspotConfig,
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
      <DialogContent className='sm:max-w-lg max-h-[85vh] overflow-y-auto'>
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

        {/* Service type info (read-only) */}
        <div className='flex items-center gap-2 rounded-md border bg-muted/40 px-3 py-2 text-xs'>
          <span className='text-muted-foreground'>Tipe Layanan:</span>
          <Badge variant='outline' className={stBadge.className}>
            {stBadge.label}
          </Badge>
          <span className='text-muted-foreground ml-auto'>
            Ubah tipe layanan lewat "Ganti Paket"
          </span>
        </div>

        <Form {...form}>
          <form
            id='subscriptions-edit-form'
            onSubmit={form.handleSubmit(handleSubmit)}
            className='flex flex-col gap-4 py-2'
          >
            {/* Credentials */}
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

            {/* PPPoE Specific */}
            {isPPPoE && (
              <div className='rounded-lg border bg-muted/30 p-3 space-y-3'>
                <p className='text-xs font-semibold text-foreground'>Konfigurasi Jaringan PPPoE</p>
                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='localAddress'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-xs'>Gateway (Local IP)</FormLabel>
                        <FormControl>
                          <Input placeholder='10.0.0.1' {...field} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name='remoteAddress'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-xs'>Static IP (Remote IP)</FormLabel>
                        <FormControl>
                          <Input placeholder='10.0.0.50' {...field} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                </div>
                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='callerId'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-xs'>Caller ID / MAC Binding</FormLabel>
                        <FormControl>
                          <Input placeholder='00:11:22:33:44:55' {...field} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name='rateLimit'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-xs'>Rate Limit Override</FormLabel>
                        <FormControl>
                          <Input placeholder='10M/10M (kosong = ikut paket)' {...field} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                </div>
              </div>
            )}

            {/* Hotspot Specific */}
            {isHotspot && (
              <div className='rounded-lg border bg-muted/30 p-3 space-y-3'>
                <p className='text-xs font-semibold text-foreground'>Konfigurasi Hotspot Member</p>
                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='server'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-xs'>Server Hotspot</FormLabel>
                        <FormControl>
                          <Input placeholder='all' {...field} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name='ipAddress'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-xs'>Static IP</FormLabel>
                        <FormControl>
                          <Input placeholder='192.168.88.50' {...field} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                </div>
                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='macAddress'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-xs'>MAC Address Lock</FormLabel>
                        <FormControl>
                          <Input placeholder='AA:BB:CC:DD:EE:FF' {...field} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name='rateLimit'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-xs'>Rate Limit Override</FormLabel>
                        <FormControl>
                          <Input placeholder='10M/10M (kosong = ikut paket)' {...field} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                </div>
                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='limitUptime'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-xs'>Limit Uptime</FormLabel>
                        <FormControl>
                          <Input placeholder='720h0m0s' {...field} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name='limitBytes'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className='text-xs'>Limit Bytes</FormLabel>
                        <FormControl>
                          <Input placeholder='10737418240' {...field} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                </div>
              </div>
            )}

            {/* Billing */}
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
