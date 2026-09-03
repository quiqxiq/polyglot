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
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  useApproveRegistrationMutation,
  useScheduleInstallMutation,
} from '../api/use-registration'
import {
  scheduleInstallSchema,
  type ScheduleInstallValues,
} from '../data/schema'
import {
  ApproveRegistrationRequest,
  ScheduleInstallRequest,
  type Registration,
} from '@/gen/v1/registration_pb'
import { REGISTRATION_STATUS } from '../data/constants'

interface RegistrationScheduleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  registration: Registration | null
}

export function RegistrationScheduleDialog({
  open,
  onOpenChange,
  registration,
}: RegistrationScheduleDialogProps) {
  const approveMutation = useApproveRegistrationMutation()
  const scheduleMutation = useScheduleInstallMutation()

  const form = useForm<ScheduleInstallValues>({
    resolver: zodResolver(scheduleInstallSchema) as never,
    defaultValues: {
      id: registration?.id || '',
      installDate: new Date().toISOString().split('T')[0],
      installTimeHhmm: '09:00',
      adminNotes: registration?.notes || '',
    },
  })

  useEffect(() => {
    if (registration) {
      let defaultDate = new Date().toISOString().split('T')[0]
      if (registration.scheduledInstallDateUnix) {
        const unix = Number(registration.scheduledInstallDateUnix)
        if (unix) {
          defaultDate = new Date(unix * 1000).toISOString().split('T')[0]
        }
      }
      form.reset({
        id: registration.id,
        installDate: defaultDate,
        installTimeHhmm: registration.scheduledInstallTime || '09:00',
        adminNotes: registration.notes || '',
      })
    }
  }, [registration, form])

  const handleSubmit = async (values: ScheduleInstallValues) => {
    if (!registration) return
    try {
      // 1. Jika status masih PENDING, otomatis approve terlebih dahulu
      if (registration.status === REGISTRATION_STATUS.PENDING) {
        await approveMutation.mutateAsync(
          new ApproveRegistrationRequest({
            id: registration.id,
            adminNotes: values.adminNotes || '',
          })
        )
      }

      // 2. Jadwalkan pemasangan
      const dateObj = new Date(values.installDate)
      const unixDate = Math.floor(dateObj.getTime() / 1000)

      await scheduleMutation.mutateAsync(
        new ScheduleInstallRequest({
          id: registration.id,
          installDateUnix: BigInt(unixDate),
          installTimeHhmm: values.installTimeHhmm || '',
          technicianId: '',
        })
      )

      toast.success('Pemasangan berhasil dijadwalkan & disetujui!')
      onOpenChange(false)
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Gagal menjadwalkan pemasangan'
      )
    }
  }

  const isPending = approveMutation.isPending || scheduleMutation.isPending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Setujui & Jadwalkan Pemasangan</DialogTitle>
          <DialogDescription>
            Tentukan tanggal dan jam rencana pemasangan untuk{' '}
            <span className='font-semibold'>{registration?.fullName}</span>.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='registration-schedule-form'
            onSubmit={form.handleSubmit(handleSubmit)}
            className='flex flex-col gap-3.5 py-1'
          >
            <div className='grid grid-cols-2 gap-3'>
              <FormField
                control={form.control}
                name='installDate'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Tanggal Pasang</FormLabel>
                    <FormControl>
                      <Input type='date' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='installTimeHhmm'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Jam Pasang</FormLabel>
                    <FormControl>
                      <Input type='time' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='adminNotes'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Catatan Admin / Petunjuk Lapangan</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='Contoh: Pelanggan minta kabel hitam, bawa tangga lipat'
                      rows={2}
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
            form='registration-schedule-form'
            type='submit'
            disabled={isPending}
          >
            {isPending ? 'Menyimpan...' : 'Simpan & Setujui'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
