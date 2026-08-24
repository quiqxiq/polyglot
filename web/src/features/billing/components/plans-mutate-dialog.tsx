import { useEffect, useMemo } from 'react'
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
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { SelectDropdown } from '@/components/select-dropdown'
import { CreatePlanRequest, Plan, UpdatePlanRequest } from '@/gen/v1/billing_pb'
import { planFormSchema, type PlanFormValues } from '../data/schema'
import { EXPIRE_MODES, SERVICE_TYPES, VALIDITY_MODES } from '../data/constants'
import { isFieldVisible } from '../data/plan-fields'
import {
  useCreatePlanMutation,
  useUpdatePlanMutation,
} from '../api/use-plans'
import { usePlans } from './plans-provider'

function toNumber(value: unknown, fallback = 0): number {
  const n = Number(value)
  return Number.isFinite(n) ? n : fallback
}

function defaultFormValues(currentRow: Plan | null): PlanFormValues {
  if (!currentRow) {
    // Explicit defaults instead of handing the schema an empty object:
    // name/serviceType/bandwidth/price have no defaults, so validating an
    // empty input throws — and this runs at render time even while the
    // dialog is closed.
    return {
      id: undefined,
      name: '',
      serviceType: 'PPPOE',
      bandwidthDownloadKbps: 10240,
      bandwidthUploadKbps: 5120,
      burstDownloadKbps: 0,
      burstUploadKbps: 0,
      burstThresholdKbps: 0,
      burstTimeSeconds: 0,
      price: 0,
      sellingPrice: 0,
      installationFee: 0,
      taxPercent: 0,
      validity: '30d',
      validityMode: 'CALENDAR',
      simultaneousUse: 1,
      ipPoolName: '',
      parentQueue: 'none',
      addressList: '',
      sharedUsers: 1,
      expireMode: '0',
      lockUser: false,
      lockServer: false,
      isActive: true,
      description: '',
    }
  }
  return {
    id: currentRow.id,
    name: currentRow.name,
    serviceType: (SERVICE_TYPES.some((t) => t.value === currentRow.serviceType)
      ? currentRow.serviceType
      : 'PPPOE') as PlanFormValues['serviceType'],
    bandwidthDownloadKbps: toNumber(currentRow.bandwidthDownloadKbps),
    bandwidthUploadKbps: toNumber(currentRow.bandwidthUploadKbps),
    burstDownloadKbps: toNumber(currentRow.burstDownloadKbps),
    burstUploadKbps: toNumber(currentRow.burstUploadKbps),
    burstThresholdKbps: toNumber(currentRow.burstThresholdKbps),
    burstTimeSeconds: toNumber(currentRow.burstTimeSeconds),
    price: toNumber(currentRow.price),
    sellingPrice: toNumber(currentRow.sellingPrice),
    installationFee: toNumber(currentRow.installationFee),
    taxPercent: toNumber(currentRow.taxPercent),
    validity: currentRow.validity || '30d',
    validityMode: (VALIDITY_MODES.some((m) => m.value === currentRow.validityMode)
      ? currentRow.validityMode
      : 'CALENDAR') as PlanFormValues['validityMode'],
    simultaneousUse: toNumber(currentRow.simultaneousUse, 1),
    ipPoolName: currentRow.ipPoolName ?? '',
    parentQueue: currentRow.parentQueue || 'none',
    addressList: currentRow.addressList ?? '',
    sharedUsers: toNumber(currentRow.sharedUsers, 1),
    expireMode: currentRow.expireMode || '0',
    lockUser: currentRow.lockUser ?? false,
    lockServer: currentRow.lockServer ?? false,
    isActive: currentRow.isActive ?? true,
    description: currentRow.description ?? '',
  }
}

function SectionHeading({ children }: { children: React.ReactNode }) {
  return (
    <p className='text-sm font-medium text-muted-foreground'>{children}</p>
  )
}

export function PlansMutateDialog() {
  const { open, setOpen, currentRow } = usePlans()
  const isOpen = open === 'create' || open === 'update'
  const isUpdate = !!currentRow
  const createMutation = useCreatePlanMutation()
  const updateMutation = useUpdatePlanMutation()

  // TFieldValues/TTransformedValues are inferred from the resolver: zod
  // `.default()` fields are optional on input but guaranteed on output.
  const form = useForm({
    resolver: zodResolver(planFormSchema),
    defaultValues: defaultFormValues(currentRow),
  })

  // The dialog stays mounted while the provider switches create/update rows;
  // useForm only reads defaultValues once at mount, so resync explicitly.
  const initialValues = useMemo(() => defaultFormValues(currentRow), [currentRow])
  useEffect(() => {
    if (isOpen) form.reset(initialValues)
  }, [isOpen, initialValues, form])

  // Visibility follows the service-type matrix (data/plan-fields.ts) and
  // re-renders live as the user switches Tipe Layanan. Fallback 'PPPOE'
  // matches the create default while the watch has not hydrated yet.
  const watchedServiceType = useWatch({
    control: form.control,
    name: 'serviceType',
  })
  const serviceType = watchedServiceType ?? 'PPPOE'

  const MIKROTIK_FIELDS = [
    'validity',
    'validityMode',
    'expireMode',
    'sharedUsers',
    'simultaneousUse',
    'ipPoolName',
    'parentQueue',
    'addressList',
    'lockUser',
    'lockServer',
  ] as const
  const BURST_FIELDS = [
    'burstDownloadKbps',
    'burstUploadKbps',
    'burstThresholdKbps',
    'burstTimeSeconds',
  ] as const
  const mikrotikVisibleCount = MIKROTIK_FIELDS.filter((f) =>
    isFieldVisible(f, serviceType)
  ).length
  const burstVisible = BURST_FIELDS.some((f) => isFieldVisible(f, serviceType))

  const onSubmit = async (values: PlanFormValues) => {
    try {
      if (isUpdate && currentRow) {
        await updateMutation.mutateAsync(
          new UpdatePlanRequest({
            plan: new Plan({ ...values, id: currentRow.id }),
          })
        )
        toast.success('Paket diperbarui')
      } else {
        await createMutation.mutateAsync(
          new CreatePlanRequest({
            plan: new Plan({ ...values }),
          })
        )
        toast.success('Paket disimpan')
      }
      setOpen(null)
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : 'Gagal menyimpan paket'
      toast.error(errorMessage)
    }
  }

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(v) => {
        setOpen(v ? (isUpdate ? 'update' : 'create') : null)
      }}
    >
      <DialogContent className='max-h-[85vh] overflow-y-auto sm:max-w-2xl'>
        <DialogHeader className='text-start'>
          <DialogTitle>{isUpdate ? 'Edit Paket' : 'Tambah Paket'}</DialogTitle>
          <DialogDescription>
            {isUpdate
              ? 'Perbarui paket layanan beserta parameter MikroTik-nya.'
              : 'Buat paket layanan baru dengan parameter MikroTik lengkap.'}{' '}
            Klik simpan setelah selesai.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form
            id='plans-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className='space-y-6'
          >
            <div className='space-y-4'>
              <SectionHeading>Dasar</SectionHeading>
              <div className='grid gap-4 sm:grid-cols-2'>
                <FormField
                  control={form.control}
                  name='name'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Nama Paket</FormLabel>
                      <FormControl>
                        <Input {...field} placeholder='Home 20M' />
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
                      <SelectDropdown
                        defaultValue={field.value}
                        onValueChange={field.onChange}
                        placeholder='Pilih tipe layanan'
                        items={[...SERVICE_TYPES]}
                      />
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='bandwidthDownloadKbps'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Bandwidth Download (Kbps)</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          {...field}
                          onChange={(e) => field.onChange(e.target.valueAsNumber)}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='bandwidthUploadKbps'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Bandwidth Upload (Kbps)</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          {...field}
                          onChange={(e) => field.onChange(e.target.valueAsNumber)}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='price'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Harga</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          step='any'
                          {...field}
                          onChange={(e) => field.onChange(e.target.valueAsNumber)}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='sellingPrice'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Harga Jual</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          step='any'
                          {...field}
                          onChange={(e) => field.onChange(e.target.valueAsNumber)}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='installationFee'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Biaya Pasang</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          step='any'
                          {...field}
                          onChange={(e) => field.onChange(e.target.valueAsNumber)}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='taxPercent'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Pajak (%)</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          step='any'
                          {...field}
                          onChange={(e) => field.onChange(e.target.valueAsNumber)}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <FormField
                control={form.control}
                name='description'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Deskripsi</FormLabel>
                    <FormControl>
                      <Textarea
                        {...field}
                        value={field.value ?? ''}
                        placeholder='Deskripsi paket (opsional)'
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='isActive'
                render={({ field }) => (
                  <FormItem className='flex flex-row items-center justify-between rounded-lg border p-3 shadow-xs'>
                    <div className='space-y-0.5'>
                      <FormLabel>Aktif</FormLabel>
                      <p className='text-muted-foreground text-xs'>
                        Paket nonaktif tidak muncul di pendaftaran langganan baru.
                      </p>
                    </div>
                    <FormControl>
                      <Switch checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                  </FormItem>
                )}
              />
            </div>

            <div className='space-y-4'>
              {mikrotikVisibleCount > 0 && (
                <SectionHeading>MikroTik</SectionHeading>
              )}
              <div className='grid gap-4 sm:grid-cols-2'>
                {isFieldVisible('validity', serviceType) && (
                  <FormField
                    control={form.control}
                    name='validity'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Validity</FormLabel>
                        <FormControl>
                          <Input {...field} placeholder='30d' />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
                {isFieldVisible('validityMode', serviceType) && (
                  <FormField
                    control={form.control}
                    name='validityMode'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Validity Mode</FormLabel>
                        <SelectDropdown
                          defaultValue={field.value}
                          onValueChange={field.onChange}
                          placeholder='Pilih mode validity'
                          items={[...VALIDITY_MODES]}
                        />
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
                {isFieldVisible('expireMode', serviceType) && (
                  <FormField
                    control={form.control}
                    name='expireMode'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Expire Mode</FormLabel>
                        <SelectDropdown
                          defaultValue={field.value}
                          onValueChange={field.onChange}
                          placeholder='Pilih mode expire'
                          items={[...EXPIRE_MODES]}
                        />
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
                {isFieldVisible('sharedUsers', serviceType) && (
                  <FormField
                    control={form.control}
                    name='sharedUsers'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Shared Users</FormLabel>
                        <FormControl>
                          <Input
                            type='number'
                            {...field}
                            onChange={(e) => field.onChange(e.target.valueAsNumber)}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
                {isFieldVisible('simultaneousUse', serviceType) && (
                  <FormField
                    control={form.control}
                    name='simultaneousUse'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Simultaneous Use</FormLabel>
                        <FormControl>
                          <Input
                            type='number'
                            {...field}
                            onChange={(e) => field.onChange(e.target.valueAsNumber)}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
                {isFieldVisible('ipPoolName', serviceType) && (
                  <FormField
                    control={form.control}
                    name='ipPoolName'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>IP Pool</FormLabel>
                        <FormControl>
                          <Input {...field} value={field.value ?? ''} placeholder='pool-hotspot' />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
                {isFieldVisible('parentQueue', serviceType) && (
                  <FormField
                    control={form.control}
                    name='parentQueue'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Parent Queue</FormLabel>
                        <FormControl>
                          <Input {...field} value={field.value ?? ''} placeholder='none' />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
                {isFieldVisible('addressList', serviceType) && (
                  <FormField
                    control={form.control}
                    name='addressList'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Address List</FormLabel>
                        <FormControl>
                          <Input {...field} value={field.value ?? ''} placeholder='paid-users' />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
              </div>
              <div className='grid gap-4 sm:grid-cols-2'>
                {isFieldVisible('lockUser', serviceType) && (
                  <FormField
                    control={form.control}
                    name='lockUser'
                    render={({ field }) => (
                      <FormItem className='flex flex-row items-center justify-between rounded-lg border p-3 shadow-xs'>
                        <div className='space-y-0.5'>
                          <FormLabel>Lock User</FormLabel>
                          <p className='text-muted-foreground text-xs'>
                            Kunci user ke router tempat pertama kali login.
                          </p>
                        </div>
                        <FormControl>
                          <Switch checked={field.value} onCheckedChange={field.onChange} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                )}
                {isFieldVisible('lockServer', serviceType) && (
                  <FormField
                    control={form.control}
                    name='lockServer'
                    render={({ field }) => (
                      <FormItem className='flex flex-row items-center justify-between rounded-lg border p-3 shadow-xs'>
                        <div className='space-y-0.5'>
                          <FormLabel>Lock Server</FormLabel>
                          <p className='text-muted-foreground text-xs'>
                            Kunci user ke interface server tertentu.
                          </p>
                        </div>
                        <FormControl>
                          <Switch checked={field.value} onCheckedChange={field.onChange} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                )}
              </div>
            </div>

            {burstVisible && (
              <div className='space-y-4'>
                <SectionHeading>Burst (Opsional)</SectionHeading>
                <div className='grid gap-4 sm:grid-cols-2'>
                  {isFieldVisible('burstDownloadKbps', serviceType) && (
                    <FormField
                      control={form.control}
                      name='burstDownloadKbps'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Burst Download (Kbps)</FormLabel>
                          <FormControl>
                            <Input
                              type='number'
                              {...field}
                              onChange={(e) => field.onChange(e.target.valueAsNumber)}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {isFieldVisible('burstUploadKbps', serviceType) && (
                    <FormField
                      control={form.control}
                      name='burstUploadKbps'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Burst Upload (Kbps)</FormLabel>
                          <FormControl>
                            <Input
                              type='number'
                              {...field}
                              onChange={(e) => field.onChange(e.target.valueAsNumber)}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {isFieldVisible('burstThresholdKbps', serviceType) && (
                    <FormField
                      control={form.control}
                      name='burstThresholdKbps'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Burst Threshold (Kbps)</FormLabel>
                          <FormControl>
                            <Input
                              type='number'
                              {...field}
                              onChange={(e) => field.onChange(e.target.valueAsNumber)}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {isFieldVisible('burstTimeSeconds', serviceType) && (
                    <FormField
                      control={form.control}
                      name='burstTimeSeconds'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Burst Time (Detik)</FormLabel>
                          <FormControl>
                            <Input
                              type='number'
                              {...field}
                              onChange={(e) => field.onChange(e.target.valueAsNumber)}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                </div>
                <FormDescription>
                  Kosongkan = tanpa burst (kolom diisi = fitur aktif)
                </FormDescription>
              </div>
            )}
          </form>
        </Form>
        <DialogFooter className='gap-2'>
          <Button
            type='button'
            variant='outline'
            onClick={() => setOpen(null)}
          >
            Cancel
          </Button>
          <Button
            form='plans-form'
            type='submit'
            disabled={createMutation.isPending || updateMutation.isPending}
          >
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
