import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  MarkInstalledRequest,
  type Registration,
} from '@/gen/v1/registration_pb'
import { Loader2, Search } from 'lucide-react'
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
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useOltsQuery } from '@/features/olt/api/use-olt'
import { useFindOnusMutation } from '@/features/ont/api/use-ont'
import { useMarkInstalledMutation } from '../api/use-registration'
import { markInstalledSchema, type MarkInstalledValues } from '../data/schema'
import { DeviceHealthBadge } from './device-health-badge'

interface RegistrationInstallDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  registration: Registration | null
}

export function RegistrationInstallDialog({
  open,
  onOpenChange,
  registration,
}: RegistrationInstallDialogProps) {
  const { data: devices = [], isLoading: devicesLoading } = useDevicesQuery()
  const markInstalledMutation = useMarkInstalledMutation()
  const { t } = useI18n()
  const { data: olts = [], isLoading: oltsLoading } = useOltsQuery()
  const findOnusMutation = useFindOnusMutation()

  const form = useForm<MarkInstalledValues>({
    resolver: zodResolver(markInstalledSchema) as never,
    defaultValues: {
      id: registration?.id || '',
      deviceId: registration?.targetDeviceId || '',
      technicianNotes: registration?.technicianNotes || '',
      installationPhotoUrl: registration?.installationPhotoUrl || '',
      ontMacAddress: registration?.ontMacAddress || '',
      ontSerialNumber: registration?.ontSerialNumber || '',
      oltDeviceId: registration?.oltDeviceId || '',
      ponIndex: registration?.ponIndex || undefined,
      onuIndex: registration?.onuIndex || undefined,
    },
  })

  useEffect(() => {
    if (registration) {
      form.reset({
        id: registration.id,
        deviceId: registration.targetDeviceId || '',
        technicianNotes: registration.technicianNotes || '',
        installationPhotoUrl: registration.installationPhotoUrl || '',
        ontMacAddress: registration.ontMacAddress || '',
        ontSerialNumber: registration.ontSerialNumber || '',
        oltDeviceId: registration.oltDeviceId || '',
        ponIndex: registration.ponIndex || undefined,
        onuIndex: registration.onuIndex || undefined,
      })
    }
  }, [registration, form])

  const deviceOptions = devices.map((d) => ({
    label: `${d.name} (${d.host || d.id})`,
    value: d.id,
  }))

  const oltOptions = olts.map((olt) => ({
    label: `${olt.name || olt.host} (${olt.host})`,
    value: olt.id,
  }))

  const handleFindOnu = async () => {
    const oltDeviceId = form.getValues('oltDeviceId') || ''
    const search = (
      form.getValues('ontMacAddress') ||
      form.getValues('ontSerialNumber') ||
      ''
    ).trim()
    if (!oltDeviceId || !search) {
      toast.error(t('psb.find_onu_need'))
      return
    }
    try {
      const onus = await findOnusMutation.mutateAsync({
        oltDeviceId,
        search,
      })
      if (onus.length === 0) {
        toast.error(t('psb.find_onu_not_found'))
        return
      }
      const onu = onus[0]
      form.setValue('ponIndex', onu.ponIndex)
      form.setValue('onuIndex', onu.onuIndex)
      if (!form.getValues('ontMacAddress') && onu.macAddress) {
        form.setValue('ontMacAddress', onu.macAddress)
      }
      if (!form.getValues('ontSerialNumber') && onu.serialNumber) {
        form.setValue('ontSerialNumber', onu.serialNumber)
      }
      toast.success(
        t('psb.find_onu_found', { pon: onu.ponIndex, onu: onu.onuIndex })
      )
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('psb.find_onu_failed'))
    }
  }

  const handleSubmit = async (values: MarkInstalledValues) => {
    if (!registration) return
    try {
      await markInstalledMutation.mutateAsync(
        new MarkInstalledRequest({
          id: registration.id,
          deviceId: values.deviceId,
          technicianNotes: values.technicianNotes,
          installationPhotoUrl: values.installationPhotoUrl || '',
          ontMacAddress: values.ontMacAddress || '',
          ontSerialNumber: values.ontSerialNumber || '',
          oltDeviceId: values.oltDeviceId || '',
          ponIndex: values.ponIndex ?? 0,
          onuIndex: values.onuIndex ?? 0,
        })
      )
      toast.success('Pemasangan fisik berhasil dicatat! Status: INSTALLED')
      onOpenChange(false)
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Gagal mencatat pemasangan'
      )
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>Hasil Pemasangan Lapangan (Teknisi)</DialogTitle>
          <DialogDescription>
            Pilih Router BRAS target yang tersambung ke pelanggan{' '}
            <span className='font-semibold'>{registration?.fullName}</span> dan
            isi catatan teknis pemasangan.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='registration-install-form'
            onSubmit={form.handleSubmit(handleSubmit)}
            className='flex flex-col gap-3.5 py-1'
          >
            <FormField
              control={form.control}
              name='deviceId'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Router BRAS Target (MikroTik)</FormLabel>
                  <FormControl>
                    <div className='space-y-2'>
                      <SelectDropdown
                        isControlled
                        defaultValue={field.value || undefined}
                        onValueChange={field.onChange}
                        placeholder={
                          devicesLoading
                            ? 'Memuat router...'
                            : 'Pilih Router BRAS'
                        }
                        items={deviceOptions}
                      />
                      <DeviceHealthBadge deviceId={field.value || ''} />
                    </div>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='technicianNotes'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Catatan Teknis Pemasangan</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='Contoh: Redaman ODP -18.5 dBm, Port 3 ODP-MGR-02, Panjang kabel drop 85 meter, ONT ZTE F609'
                      rows={3}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='installationPhotoUrl'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>URL Foto Bukti Instalasi (opsional)</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='/uploads/instalasi/REG-202609-0001.jpg'
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='grid grid-cols-2 gap-3'>
              <FormField
                control={form.control}
                name='ontMacAddress'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>MAC ONT (opsional)</FormLabel>
                    <FormControl>
                      <Input placeholder='ZTEGC1234567' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='ontSerialNumber'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>SN ONT (opsional)</FormLabel>
                    <FormControl>
                      <Input placeholder='ZTEGC12345678' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className='space-y-3 rounded-lg border p-3'>
              <div className='flex flex-wrap items-start justify-between gap-2'>
                <div>
                  <p className='text-sm font-medium'>{t('psb.ont_section')}</p>
                  <p className='text-xs text-muted-foreground'>
                    {t('psb.ont_section_desc')}
                  </p>
                </div>
                <Button
                  type='button'
                  size='sm'
                  variant='outline'
                  className='h-8'
                  onClick={handleFindOnu}
                  disabled={findOnusMutation.isPending}
                >
                  {findOnusMutation.isPending ? (
                    <Loader2 className='h-3.5 w-3.5 animate-spin' />
                  ) : (
                    <Search className='h-3.5 w-3.5' />
                  )}
                  {t('psb.find_onu')}
                </Button>
              </div>

              <FormField
                control={form.control}
                name='oltDeviceId'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('psb.olt')}</FormLabel>
                    <FormControl>
                      <SelectDropdown
                        isControlled
                        defaultValue={field.value || undefined}
                        onValueChange={field.onChange}
                        placeholder={
                          oltsLoading
                            ? t('common.loading')
                            : t('psb.olt_placeholder')
                        }
                        items={oltOptions}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className='grid grid-cols-2 gap-3'>
                <FormField
                  control={form.control}
                  name='ponIndex'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('psb.pon_index')}</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          min={0}
                          value={field.value ?? ''}
                          onChange={(e) =>
                            field.onChange(
                              e.target.value === ''
                                ? undefined
                                : Number(e.target.value)
                            )
                          }
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='onuIndex'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('psb.onu_index')}</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          min={0}
                          value={field.value ?? ''}
                          onChange={(e) =>
                            field.onChange(
                              e.target.value === ''
                                ? undefined
                                : Number(e.target.value)
                            )
                          }
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </div>
          </form>
        </Form>

        <DialogFooter>
          <Button
            type='button'
            variant='outline'
            onClick={() => onOpenChange(false)}
          >
            Batal
          </Button>
          <Button
            form='registration-install-form'
            type='submit'
            disabled={markInstalledMutation.isPending}
          >
            {markInstalledMutation.isPending
              ? 'Menyimpan...'
              : 'Simpan Pemasangan'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
