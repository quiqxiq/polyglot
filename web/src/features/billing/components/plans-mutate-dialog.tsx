import { useEffect, useMemo } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm, useWatch } from 'react-hook-form'
import { toast } from 'sonner'
import { Plan, CreatePlanRequest, UpdatePlanRequest } from '@/gen/v1/plan_pb'
import { useDeviceStore } from '@/stores/device-store'
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
import { SelectDropdown } from '@/components/select-dropdown'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  useIpPoolsQuery,
  useParentQueuesQuery,
} from '@/features/hotspot/api/use-router-resources'
import {
  useCreatePlanMutation,
  useUpdatePlanMutation,
} from '../api/use-plans'
import { isFieldHidden, isFieldVisible } from '../data/plan-fields'
import { planFormSchema, type PlanFormValues } from '../data/schema'
import { usePlans } from './plans-provider'

const SERVICE_TYPES = [
  { label: 'PPPoE (Fiber / Broadband)', value: 'PPPOE' },
  { label: 'Hotspot Permanent (Langganan Tetap)', value: 'HOTSPOT' },
  { label: 'Dedicated (CIR Guaranteed)', value: 'DEDICATED' },
] as const

function toNumber(v: unknown, def = 0): number {
  if (typeof v === 'number') return isNaN(v) ? def : v
  if (typeof v === 'string') {
    const n = Number(v)
    return isNaN(n) ? def : n
  }
  return def
}

function defaultFormValues(currentRow: Plan | null | undefined): PlanFormValues {
  if (!currentRow) {
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
      installationFee: 0,
      taxPercent: 0,
      ipPoolName: '',
      remoteAddressPool: '',
      parentQueue: 'none',
      addressList: '',
      sessionTimeout: '',
      idleTimeout: '',
      sharedUsers: 1,
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
    installationFee: toNumber(currentRow.installationFee),
    taxPercent: toNumber(currentRow.taxPercent),
    ipPoolName: currentRow.ipPoolName || currentRow.hotspotConfig?.ipPoolName || '',
    remoteAddressPool: currentRow.remoteAddressPool || currentRow.pppoeConfig?.remoteAddressPool || '',
    parentQueue: currentRow.parentQueue || 'none',
    addressList: currentRow.addressList || currentRow.pppoeConfig?.addressList || currentRow.hotspotConfig?.addressList || '',
    sessionTimeout: currentRow.sessionTimeout || currentRow.pppoeConfig?.sessionTimeout || currentRow.hotspotConfig?.sessionTimeout || '',
    idleTimeout: currentRow.idleTimeout || currentRow.pppoeConfig?.idleTimeout || currentRow.hotspotConfig?.idleTimeout || '',
    sharedUsers: toNumber(currentRow.sharedUsers || currentRow.hotspotConfig?.sharedUsers, 1),
    isActive: currentRow.isActive ?? true,
    description: currentRow.description ?? '',
  }
}

function SectionHeading({ children }: { children: React.ReactNode }) {
  return (
    <p className='text-sm font-semibold text-muted-foreground uppercase tracking-wider'>{children}</p>
  )
}

export function PlansMutateDialog() {
  const { open, setOpen, currentRow } = usePlans()
  const isOpen = open === 'create' || open === 'update'
  const isUpdate = open === 'update' && !!currentRow
  const createMutation = useCreatePlanMutation()
  const updateMutation = useUpdatePlanMutation()

  const selectedDeviceId = useDeviceStore((s) => s.selectedDeviceId)

  const form = useForm<PlanFormValues>({
    resolver: zodResolver(planFormSchema) as never,
    defaultValues: defaultFormValues(currentRow),
  })

  const initialValues = useMemo(() => defaultFormValues(currentRow), [currentRow])

  useEffect(() => {
    if (isOpen) {
      form.reset(initialValues)
    }
  }, [isOpen, initialValues, form])

  const parentQueues = useParentQueuesQuery(selectedDeviceId)
  const ipPools = useIpPoolsQuery(selectedDeviceId)

  const watchedServiceType = useWatch({
    control: form.control,
    name: 'serviceType',
  })
  const serviceType = watchedServiceType ?? 'PPPOE'

  const showField = (field: string) =>
    isFieldVisible(field, serviceType) && !isFieldHidden(field, serviceType)

  const MIKROTIK_FIELDS = [
    'sharedUsers',
    'ipPoolName',
    'parentQueue',
    'addressList',
    'remoteAddressPool',
    'sessionTimeout',
    'idleTimeout',
  ] as const
  const BURST_FIELDS = [
    'burstDownloadKbps',
    'burstUploadKbps',
    'burstThresholdKbps',
    'burstTimeSeconds',
  ] as const
  const mikrotikVisibleCount = MIKROTIK_FIELDS.filter(showField).length
  const burstVisible = BURST_FIELDS.some(showField)

  const onSubmit = async (values: PlanFormValues) => {
    try {
      const planPayload = new Plan({
        ...values,
        id: isUpdate && currentRow ? currentRow.id : '',
        parentQueue: values.parentQueue === 'none' ? '' : (values.parentQueue || ''),
        remoteAddressPool: values.remoteAddressPool === 'none' ? '' : (values.remoteAddressPool || ''),
        ipPoolName: values.ipPoolName === 'none' ? '' : (values.ipPoolName || ''),
        sessionTimeout: values.sessionTimeout || '',
        idleTimeout: values.idleTimeout || '',
        pppoeConfig:
          values.serviceType !== 'HOTSPOT'
            ? {
                remoteAddressPool: values.remoteAddressPool === 'none' ? '' : (values.remoteAddressPool || ''),
                addressList: values.addressList || '',
                sessionTimeout: values.sessionTimeout || '',
                idleTimeout: values.idleTimeout || '',
              }
            : undefined,
        hotspotConfig:
          values.serviceType === 'HOTSPOT'
            ? {
                ipPoolName: values.ipPoolName === 'none' ? '' : (values.ipPoolName || ''),
                addressList: values.addressList || '',
                sharedUsers: values.sharedUsers || 1,
                sessionTimeout: values.sessionTimeout || '',
                idleTimeout: values.idleTimeout || '',
              }
            : undefined,
      })

      if (isUpdate && currentRow) {
        await updateMutation.mutateAsync(
          new UpdatePlanRequest({
            plan: planPayload,
            deviceId: selectedDeviceId || '',
          })
        )
        toast.success('Paket berhasil diperbarui')
      } else {
        await createMutation.mutateAsync(
          new CreatePlanRequest({
            plan: planPayload,
            deviceId: selectedDeviceId || '',
          })
        )
        toast.success('Paket berhasil disimpan')
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
        if (!v) setOpen(null)
      }}
    >
      <DialogContent className='max-h-[85vh] overflow-y-auto sm:max-w-2xl'>
        <DialogHeader className='text-start'>
          <DialogTitle>{isUpdate ? 'Edit Paket Layanan' : 'Tambah Paket Layanan'}</DialogTitle>
          <DialogDescription>
            {isUpdate
              ? 'Perbarui konfigurasi paket langganan dan parameter profil router.'
              : 'Buat paket layanan baru (PPPoE / Hotspot Tetap / Dedicated).'}{' '}
            Profil paket akan disinkronkan otomatis ke router MikroTik yang aktif.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form
            id='plans-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className='space-y-6'
          >
            <div className='space-y-4'>
              <SectionHeading>Informasi Dasar</SectionHeading>
              <div className='grid gap-4 sm:grid-cols-2'>
                <FormField
                  control={form.control}
                  name='name'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Nama Paket</FormLabel>
                      <FormControl>
                        <Input {...field} placeholder='Home Fiber 20M' />
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
                        isControlled
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
                          onChange={(e) => field.onChange(isNaN(e.target.valueAsNumber) ? 0 : e.target.valueAsNumber)}
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
                          onChange={(e) => field.onChange(isNaN(e.target.valueAsNumber) ? 0 : e.target.valueAsNumber)}
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
                      <FormLabel>Biaya Langganan (Rp)</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          step='any'
                          {...field}
                          onChange={(e) => field.onChange(isNaN(e.target.valueAsNumber) ? 0 : e.target.valueAsNumber)}
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
                      <FormLabel>Biaya Pasang Baru (Rp)</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          step='any'
                          {...field}
                          onChange={(e) => field.onChange(isNaN(e.target.valueAsNumber) ? 0 : e.target.valueAsNumber)}
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
                      <FormLabel>Pajak PPN (%)</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          step='any'
                          {...field}
                          onChange={(e) => field.onChange(isNaN(e.target.valueAsNumber) ? 0 : e.target.valueAsNumber)}
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
                        placeholder='Deskripsi atau catatan paket (opsional)'
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
                      <FormLabel>Status Aktif</FormLabel>
                      <p className='text-muted-foreground text-xs'>
                        Paket nonaktif tidak akan muncul pada pilihan registrasi pelanggan baru.
                      </p>
                    </div>
                    <FormControl>
                      <Switch checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                  </FormItem>
                )}
              />
            </div>

            {mikrotikVisibleCount > 0 && (
              <div className='space-y-4'>
                <SectionHeading>Konfigurasi Profil Router MikroTik</SectionHeading>
                <div className='grid gap-4 sm:grid-cols-2'>
                  {showField('sharedUsers') && (
                    <FormField
                      control={form.control}
                      name='sharedUsers'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Shared Users (Batas Login Bersamaan)</FormLabel>
                          <FormControl>
                            <Input
                              type='number'
                              min={1}
                              {...field}
                              onChange={(e) => field.onChange(isNaN(e.target.valueAsNumber) ? 1 : e.target.valueAsNumber)}
                            />
                          </FormControl>
                          <FormDescription className='text-xs'>
                            Jumlah gadget/perangkat yang diizinkan login bersamaan untuk 1 akun.
                          </FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {showField('ipPoolName') && (
                    <FormField
                      control={form.control}
                      name='ipPoolName'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>IP Pool Hotspot</FormLabel>
                          <SelectDropdown
                            isControlled
                            defaultValue={field.value || 'none'}
                            onValueChange={(val) => field.onChange(val === 'none' ? '' : val)}
                            placeholder={
                              ipPools.isFetching
                                ? 'Memuat pool dari router…'
                                : 'Pilih IP Pool Hotspot'
                            }
                            items={[
                              { label: 'none', value: 'none' },
                              ...(ipPools.data ?? []).map((o) => ({
                                label: o,
                                value: o,
                              })),
                            ]}
                          />
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {showField('remoteAddressPool') && (
                    <FormField
                      control={form.control}
                      name='remoteAddressPool'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Remote Address Pool (PPPoE)</FormLabel>
                          <SelectDropdown
                            isControlled
                            defaultValue={field.value || 'none'}
                            onValueChange={(val) => field.onChange(val === 'none' ? '' : val)}
                            placeholder={
                              ipPools.isFetching
                                ? 'Memuat pool dari router…'
                                : 'Pilih Remote Address Pool'
                            }
                            items={[
                              { label: 'none', value: 'none' },
                              ...(ipPools.data ?? []).map((o) => ({
                                label: o,
                                value: o,
                              })),
                            ]}
                          />
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {showField('parentQueue') && (
                    <FormField
                      control={form.control}
                      name='parentQueue'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Parent Queue</FormLabel>
                          <SelectDropdown
                            isControlled
                            defaultValue={field.value || 'none'}
                            onValueChange={(val) => field.onChange(val === 'none' ? '' : val)}
                            placeholder={
                              parentQueues.isFetching
                                ? 'Memuat queue dari router…'
                                : 'Pilih parent queue'
                            }
                            items={[
                              { label: 'none', value: 'none' },
                              ...(parentQueues.data ?? []).map((o) => ({
                                label: o,
                                value: o,
                              })),
                            ]}
                          />
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {showField('addressList') && (
                    <FormField
                      control={form.control}
                      name='addressList'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Address List Firewall (Opsional)</FormLabel>
                          <FormControl>
                            <Input {...field} value={field.value ?? ''} placeholder='paid-users / VIP' />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {showField('sessionTimeout') && (
                    <FormField
                      control={form.control}
                      name='sessionTimeout'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Session Timeout (Opsional)</FormLabel>
                          <FormControl>
                            <Input {...field} value={field.value ?? ''} placeholder='Contoh: 1d atau 12:00:00' />
                          </FormControl>
                          <FormDescription className='text-xs'>
                            Batas maksimum durasi koneksi sebelum re-autentikasi oleh router.
                          </FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {showField('idleTimeout') && (
                    <FormField
                      control={form.control}
                      name='idleTimeout'
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Idle Timeout (Opsional)</FormLabel>
                          <FormControl>
                            <Input {...field} value={field.value ?? ''} placeholder='Contoh: 10m atau 00:05:00' />
                          </FormControl>
                          <FormDescription className='text-xs'>
                            Batas waktu tanpa traffic data sebelum koneksi diputuskan otomatis.
                          </FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                </div>
              </div>
            )}

            {burstVisible && (
              <div className='space-y-4'>
                <SectionHeading>Burst Bandwidth (Opsional)</SectionHeading>
                <div className='grid gap-4 sm:grid-cols-2'>
                  {showField('burstDownloadKbps') && (
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
                              onChange={(e) => field.onChange(isNaN(e.target.valueAsNumber) ? 0 : e.target.valueAsNumber)}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {showField('burstUploadKbps') && (
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
                              onChange={(e) => field.onChange(isNaN(e.target.valueAsNumber) ? 0 : e.target.valueAsNumber)}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {showField('burstThresholdKbps') && (
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
                              onChange={(e) => field.onChange(isNaN(e.target.valueAsNumber) ? 0 : e.target.valueAsNumber)}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                  {showField('burstTimeSeconds') && (
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
                              onChange={(e) => field.onChange(isNaN(e.target.valueAsNumber) ? 0 : e.target.valueAsNumber)}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )}
                </div>
                <FormDescription>
                  Kosongkan/nol = tanpa burst (fitur burst aktif jika seluruh nilai burst terisi).
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
            Batal
          </Button>
          <Button
            form='plans-form'
            type='submit'
            disabled={createMutation.isPending || updateMutation.isPending}
          >
            {isUpdate ? 'Simpan Perubahan' : 'Simpan Paket'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

