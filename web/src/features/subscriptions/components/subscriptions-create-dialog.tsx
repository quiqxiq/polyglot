import { useEffect } from 'react'
import { z } from 'zod'
import { useForm, useWatch } from 'react-hook-form'
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
import { CreateSubscriptionRequest } from '@/gen/v1/billing_pb'
import { useCustomersQuery } from '@/features/customer/api/use-customer'
import { usePlansQuery } from '@/features/billing/api/use-plans'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useCreateSubscriptionMutation } from '@/features/billing/api/use-billing'

const SERVICE_TYPES = ['PPPOE', 'HOTSPOT', 'DEDICATED'] as const

// Sentinel untuk "tanpa device" — Radix Select melarang value string kosong.
const NO_DEVICE = '__none__'

const createSubscriptionSchema = z.object({
  customerId: z.string().min(1, 'Pelanggan wajib dipilih'),
  planId: z.string().min(1, 'Paket wajib dipilih'),
  deviceId: z.string(),
  serviceType: z.enum(SERVICE_TYPES),
  remoteUsername: z.string(),
  remotePassword: z.string(),
  customPrice: z.coerce.number().min(0),
  notes: z.string(),
  // PPPoE Specific
  localAddress: z.string().optional(),
  remoteAddress: z.string().optional(),
  callerId: z.string().optional(),
  // Hotspot Specific
  server: z.string().optional(),
  macAddress: z.string().optional(),
  ipAddress: z.string().optional(),
})

type CreateSubscriptionValues = z.infer<typeof createSubscriptionSchema>

interface SubscriptionsCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialCustomerId?: string
  lockCustomer?: boolean
}

export function SubscriptionsCreateDialog({
  open,
  onOpenChange,
  initialCustomerId = '',
  lockCustomer = false,
}: SubscriptionsCreateDialogProps) {
  const { data: customers = [], isPending: customersLoading } =
    useCustomersQuery()
  const { data: plans = [], isPending: plansLoading } = usePlansQuery(true)
  const { data: devices = [] } = useDevicesQuery()
  const createMutation = useCreateSubscriptionMutation()

  const form = useForm<CreateSubscriptionValues>({
    resolver: zodResolver(createSubscriptionSchema) as never,
    defaultValues: {
      customerId: initialCustomerId,
      planId: '',
      deviceId: '',
      serviceType: 'PPPOE',
      remoteUsername: '',
      remotePassword: '',
      customPrice: 0,
      notes: '',
      localAddress: '',
      remoteAddress: '',
      callerId: '',
      server: 'all',
      macAddress: '',
      ipAddress: '',
    },
  })

  const selectedPlanId = useWatch({ control: form.control, name: 'planId' })
  const serviceType = useWatch({ control: form.control, name: 'serviceType' })

  // Auto set service type when plan is picked
  useEffect(() => {
    if (selectedPlanId) {
      const selectedPlan = plans.find((p) => p.id === selectedPlanId)
      if (selectedPlan?.serviceType) {
        form.setValue(
          'serviceType',
          selectedPlan.serviceType as (typeof SERVICE_TYPES)[number]
        )
      }
    }
  }, [selectedPlanId, plans, form])

  const customerItems = customers.map((c) => ({
    label: c.name ? `${c.name} (${c.id})` : c.id,
    value: c.id,
  }))
  const planItems = plans.map((p) => ({
    label: `${p.name} — Rp${Number(p.price).toLocaleString('id-ID')} (${p.serviceType})`,
    value: p.id,
  }))
  const deviceItems = [
    { label: '— tanpa device —', value: NO_DEVICE },
    ...devices.map((d) => ({
      label: d.name ? `${d.name} (${d.id})` : d.id,
      value: d.id,
    })),
  ]

  const handleSubmit = async (values: CreateSubscriptionValues) => {
    try {
      await createMutation.mutateAsync(
        new CreateSubscriptionRequest({
          customerId: values.customerId,
          planId: values.planId,
          deviceId: values.deviceId === NO_DEVICE ? '' : values.deviceId,
          serviceType: values.serviceType,
          remoteUsername: values.remoteUsername,
          remotePassword: values.remotePassword,
          customPrice: values.customPrice,
          notes: values.notes,
          pppoeConfig:
            values.serviceType !== 'HOTSPOT'
              ? {
                  localAddress: values.localAddress || '',
                  remoteAddress: values.remoteAddress || '',
                  callerId: values.callerId || '',
                }
              : undefined,
          hotspotConfig:
            values.serviceType === 'HOTSPOT'
              ? {
                  server: values.server || 'all',
                  macAddress: values.macAddress || '',
                  ipAddress: values.ipAddress || '',
                }
              : undefined,
        })
      )
      toast.success('Langganan berhasil dibuat!')
      form.reset()
      onOpenChange(false)
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Gagal membuat langganan'
      )
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-lg max-h-[85vh] overflow-y-auto'>
        <DialogHeader>
          <DialogTitle>Tambah Langganan</DialogTitle>
          <DialogDescription>
            Buat langganan baru. Username/password router dikosongkan untuk
            auto-generate.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='subscriptions-create-form'
            onSubmit={form.handleSubmit(handleSubmit)}
            className='flex flex-col gap-4 py-2'
          >
            <FormField
              control={form.control}
              name='customerId'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Pelanggan</FormLabel>
                  <FormControl>
                    <SelectDropdown
                      isControlled
                      defaultValue={field.value || undefined}
                      onValueChange={field.onChange}
                      placeholder={
                        customersLoading
                          ? 'Memuat pelanggan...'
                          : 'Pilih pelanggan'
                      }
                      isPending={customersLoading}
                      items={customerItems}
                      disabled={lockCustomer}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='planId'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Paket Layanan</FormLabel>
                  <FormControl>
                    <SelectDropdown
                      isControlled
                      defaultValue={field.value || undefined}
                      onValueChange={field.onChange}
                      placeholder={
                        plansLoading ? 'Memuat paket...' : 'Pilih paket'
                      }
                      isPending={plansLoading}
                      items={planItems}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='deviceId'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Router BRAS (Opsional)</FormLabel>
                  <FormControl>
                    <SelectDropdown
                      isControlled
                      defaultValue={field.value}
                      onValueChange={field.onChange}
                      placeholder='Pilih Router BRAS'
                      items={deviceItems}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='serviceType'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Tipe Layanan</FormLabel>
                  <FormControl>
                    <SelectDropdown
                      isControlled
                      defaultValue={field.value}
                      onValueChange={field.onChange}
                      placeholder='Pilih tipe layanan'
                      items={SERVICE_TYPES.map((t) => ({ label: t, value: t }))}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
              <FormField
                control={form.control}
                name='remoteUsername'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Username Kredensial</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder='kosongkan = auto-generate'
                      />
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
                    <FormLabel>Password Kredensial</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        type='password'
                        placeholder='kosongkan = auto-generate'
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            {/* PPPoE Specific Fields */}
            {serviceType !== 'HOTSPOT' && (
              <div className='rounded-lg border bg-muted/30 p-3 space-y-3'>
                <p className='text-xs font-semibold text-foreground'>
                  Konfigurasi Jaringan PPPoE (Opsional)
                </p>
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
              </div>
            )}

            {/* Hotspot Specific Fields */}
            {serviceType === 'HOTSPOT' && (
              <div className='rounded-lg border bg-muted/30 p-3 space-y-3'>
                <p className='text-xs font-semibold text-foreground'>
                  Konfigurasi Hotspot Member (Opsional)
                </p>
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
              </div>
            )}

            <FormField
              control={form.control}
              name='customPrice'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Harga Khusus / Diskon (0 = Ikut Paket)</FormLabel>
                  <FormControl>
                    <Input {...field} type='number' step='any' min={0} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

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
            form='subscriptions-create-form'
            disabled={
              createMutation.isPending ||
              !form.watch('customerId') ||
              !form.watch('planId')
            }
          >
            {createMutation.isPending ? 'Menyimpan...' : 'Buat Langganan'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
