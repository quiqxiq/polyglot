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
import { useRejectRegistrationMutation } from '../api/use-registration'
import { rejectRegistrationSchema, type RejectRegistrationValues } from '../data/schema'
import { RejectRegistrationRequest, type Registration } from '@/gen/v1/registration_pb'

interface RegistrationRejectDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  registration: Registration | null
}

export function RegistrationRejectDialog({
  open,
  onOpenChange,
  registration,
}: RegistrationRejectDialogProps) {
  const rejectMutation = useRejectRegistrationMutation()

  const form = useForm<RejectRegistrationValues>({
    resolver: zodResolver(rejectRegistrationSchema) as never,
    defaultValues: {
      id: registration?.id || '',
      reason: '',
    },
  })

  const handleSubmit = async (values: RejectRegistrationValues) => {
    if (!registration) return
    try {
      await rejectMutation.mutateAsync(
        new RejectRegistrationRequest({
          id: registration.id,
          reason: values.reason,
        })
      )
      toast.success('Pendaftaran ditolak')
      form.reset()
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal menolak pendaftaran')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Tolak Pendaftaran Calon Pelanggan</DialogTitle>
          <DialogDescription>
            Masukkan alasan penolakan untuk pendaftaran{' '}
            <span className='font-semibold'>{registration?.fullName}</span>.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='registration-reject-form'
            onSubmit={form.handleSubmit(handleSubmit)}
            className='flex flex-col gap-3 py-1'
          >
            <FormField
              control={form.control}
              name='reason'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Alasan Penolakan</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='Contoh: Wilayah di luar jangkauan ODP, kapasitas port penuh'
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
            form='registration-reject-form'
            type='submit'
            variant='destructive'
            disabled={rejectMutation.isPending}
          >
            {rejectMutation.isPending ? 'Menolak...' : 'Tolak Pendaftaran'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
