import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { CheckCircle2, Router, Zap } from 'lucide-react'
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
import { SelectDropdown } from '@/components/select-dropdown'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useConvertRegistrationMutation } from '../api/use-registration'
import {
  convertRegistrationSchema,
  type ConvertRegistrationValues,
} from '../data/schema'
import {
  ConvertRegistrationRequest,
  type Registration,
} from '@/gen/v1/registration_pb'

interface RegistrationConvertDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  registration: Registration | null
}

export function RegistrationConvertDialog({
  open,
  onOpenChange,
  registration,
}: RegistrationConvertDialogProps) {
  const { data: devices = [], isLoading: devicesLoading } = useDevicesQuery()
  const convertMutation = useConvertRegistrationMutation()

  const form = useForm<ConvertRegistrationValues>({
    resolver: zodResolver(convertRegistrationSchema) as never,
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

  const handleSubmit = async (values: ConvertRegistrationValues) => {
    if (!registration) return
    try {
      const res = await convertMutation.mutateAsync(
        new ConvertRegistrationRequest({
          id: registration.id,
          deviceId: values.deviceId,
          technicianNotes: values.technicianNotes || '',
        })
      )
      toast.success(
        `Aktivasi Berhasil! Customer ID: ${res.customerId}, Sub ID: ${res.subscriptionId}`
      )
      onOpenChange(false)
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Gagal melakukan aktivasi'
      )
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <div className='flex items-center gap-2 text-emerald-600 dark:text-emerald-400'>
            <Zap className='h-5 w-5' />
            <DialogTitle>Aktivasi Pelanggan & Provisi Jaringan</DialogTitle>
          </div>
          <DialogDescription>
            Konfirmasi aktivasi untuk calon pelanggan{' '}
            <span className='font-semibold'>{registration?.fullName}</span>.
          </DialogDescription>
        </DialogHeader>

        <div className='rounded-lg border bg-muted/40 p-3.5 text-xs space-y-1.5'>
          <p className='font-semibold text-foreground flex items-center gap-1.5'>
            <CheckCircle2 className='h-4 w-4 text-emerald-500' />
            Tindakan Otomatis yang Dijalankan Sistem:
          </p>
          <ul className='list-disc list-inside space-y-1 text-muted-foreground'>
            <li>Membuat master data <b>Customer</b> & kode akses portal.</li>
            <li>Generate username & password unik berbasis inisial pelanggan.</li>
            <li>Membuat akun <b>Subscription</b> & provisi akun ke Router MikroTik.</li>
            <li>Menerbitkan <b>Invoice Tagihan Pertama</b> (paket + biaya pasang).</li>
          </ul>
        </div>

        <Form {...form}>
          <form
            id='registration-convert-form'
            onSubmit={form.handleSubmit(handleSubmit)}
            className='flex flex-col gap-3 py-1'
          >
            <FormField
              control={form.control}
              name='deviceId'
              render={({ field }) => (
                <FormItem>
                  <FormLabel className='flex items-center gap-1.5'>
                    <Router className='h-3.5 w-3.5' /> Router BRAS Target
                  </FormLabel>
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
            form='registration-convert-form'
            type='submit'
            className='bg-emerald-600 hover:bg-emerald-700 text-white'
            disabled={convertMutation.isPending}
          >
            {convertMutation.isPending ? 'Mengaktivasi...' : 'Aktivasi Sekarang'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
