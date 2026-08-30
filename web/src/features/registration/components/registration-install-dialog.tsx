import { useEffect } from 'react'
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
import { Textarea } from '@/components/ui/textarea'
import { SelectDropdown } from '@/components/select-dropdown'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useMarkInstalledMutation } from '../api/use-registration'
import { markInstalledSchema, type MarkInstalledValues } from '../data/schema'
import { MarkInstalledRequest, type Registration } from '@/gen/v1/registration_pb'

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

  const form = useForm<MarkInstalledValues>({
    resolver: zodResolver(markInstalledSchema) as never,
    defaultValues: {
      id: registration?.id || '',
      deviceId: registration?.targetDeviceId || '',
      technicianNotes: registration?.technicianNotes || '',
    },
  })

  useEffect(() => {
    if (registration) {
      form.reset({
        id: registration.id,
        deviceId: registration.targetDeviceId || '',
        technicianNotes: registration.technicianNotes || '',
      })
    }
  }, [registration, form])

  const deviceOptions = devices.map((d) => ({
    label: `${d.name} (${d.host || d.id})`,
    value: d.id,
  }))

  const handleSubmit = async (values: MarkInstalledValues) => {
    if (!registration) return
    try {
      await markInstalledMutation.mutateAsync(
        new MarkInstalledRequest({
          id: registration.id,
          deviceId: values.deviceId,
          technicianNotes: values.technicianNotes,
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
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Hasil Pemasangan Lapangan (Teknisi)</DialogTitle>
          <DialogDescription>
            Pilih Router BRAS target yang tersambung ke pelanggan{' '}
            <span className='font-semibold'>{registration?.fullName}</span> dan isi
            catatan teknis pemasangan.
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
                    <SelectDropdown
                      isControlled
                      defaultValue={field.value || undefined}
                      onValueChange={field.onChange}
                      placeholder={
                        devicesLoading ? 'Memuat router...' : 'Pilih Router BRAS'
                      }
                      items={deviceOptions}
                    />
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
            {markInstalledMutation.isPending ? 'Menyimpan...' : 'Simpan Pemasangan'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
