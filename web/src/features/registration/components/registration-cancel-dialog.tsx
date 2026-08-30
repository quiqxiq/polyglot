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
import { useCancelRegistrationMutation } from '../api/use-registration'
import { cancelRegistrationSchema, type CancelRegistrationValues } from '../data/schema'
import { CancelRegistrationRequest, type Registration } from '@/gen/v1/registration_pb'

interface RegistrationCancelDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  registration: Registration | null
}

export function RegistrationCancelDialog({
  open,
  onOpenChange,
  registration,
}: RegistrationCancelDialogProps) {
  const cancelMutation = useCancelRegistrationMutation()

  const form = useForm<CancelRegistrationValues>({
    resolver: zodResolver(cancelRegistrationSchema) as never,
    defaultValues: {
      id: registration?.id || '',
      reason: '',
    },
  })

  const handleSubmit = async (values: CancelRegistrationValues) => {
    if (!registration) return
    try {
      await cancelMutation.mutateAsync(
        new CancelRegistrationRequest({
          id: registration.id,
          reason: values.reason,
        })
      )
      toast.success('Pendaftaran dibatalkan')
      form.reset()
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal membatalkan pendaftaran')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Batalkan Pendaftaran</DialogTitle>
          <DialogDescription>
            Masukkan alasan pembatalan untuk pendaftaran{' '}
            <span className='font-semibold'>{registration?.fullName}</span>.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='registration-cancel-form'
            onSubmit={form.handleSubmit(handleSubmit)}
            className='flex flex-col gap-3 py-1'
          >
            <FormField
              control={form.control}
              name='reason'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Alasan Pembatalan</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='Contoh: Calon pelanggan membatalkan permintaan pasang'
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
          <Button type='button' variant='outline' onClick={() => onOpenChange(false)}>
            Batal
          </Button>
          <Button
            form='registration-cancel-form'
            type='submit'
            variant='destructive'
            disabled={cancelMutation.isPending}
          >
            {cancelMutation.isPending ? 'Membatalkan...' : 'Batalkan Pendaftaran'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
