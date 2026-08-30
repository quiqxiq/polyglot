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
import { SelectDropdown } from '@/components/select-dropdown'
import { usePlansQuery } from '@/features/billing/api/use-plans'
import { useSubmitRegistrationMutation } from '../api/use-registration'
import {
  submitRegistrationSchema,
  type SubmitRegistrationValues,
} from '../data/schema'
import { SubmitRegistrationRequest } from '@/gen/v1/registration_pb'

interface RegistrationCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function RegistrationCreateDialog({
  open,
  onOpenChange,
}: RegistrationCreateDialogProps) {
  const { data: plans = [], isLoading: plansLoading } = usePlansQuery(true)
  const submitMutation = useSubmitRegistrationMutation()

  const form = useForm<SubmitRegistrationValues>({
    resolver: zodResolver(submitRegistrationSchema) as never,
    defaultValues: {
      fullName: '',
      phone: '',
      email: '',
      address: '',
      planId: '',
      latitude: undefined,
      longitude: undefined,
      notes: '',
    },
  })

  const planOptions = plans.map((p) => ({
    label: `${p.name} — Rp${Number(p.price).toLocaleString('id-ID')} (${p.serviceType})`,
    value: p.id,
  }))

  const handleSubmit = async (values: SubmitRegistrationValues) => {
    try {
      const hasCoords =
        values.latitude !== undefined &&
        values.longitude !== undefined &&
        !isNaN(values.latitude) &&
        !isNaN(values.longitude)

      await submitMutation.mutateAsync(
        new SubmitRegistrationRequest({
          fullName: values.fullName,
          phone: values.phone,
          email: values.email || '',
          address: values.address,
          planId: values.planId,
          latitude: values.latitude || 0,
          longitude: values.longitude || 0,
          hasCoordinates: hasCoords,
          notes: values.notes || '',
        })
      )
      toast.success('Pendaftaran calon pelanggan berhasil dikirim!')
      form.reset()
      onOpenChange(false)
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Gagal mengirim pendaftaran'
      )
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>Form Pendaftaran Calon Pelanggan</DialogTitle>
          <DialogDescription>
            Isi data calon pelanggan baru. Status awal adalah PENDING dan akan
            ditinjau sebelum penugasan teknisi.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='registration-submit-form'
            onSubmit={form.handleSubmit(handleSubmit)}
            className='flex flex-col gap-3 py-1'
          >
            <FormField
              control={form.control}
              name='fullName'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Nama Lengkap</FormLabel>
                  <FormControl>
                    <Input placeholder='Contoh: Budi Santoso' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='grid grid-cols-1 gap-3 sm:grid-cols-2'>
              <FormField
                control={form.control}
                name='phone'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>WhatsApp / Telepon</FormLabel>
                    <FormControl>
                      <Input placeholder='0856xxxxxxx' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='email'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email (Opsional)</FormLabel>
                    <FormControl>
                      <Input
                        type='email'
                        placeholder='budi@example.com'
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='planId'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Pilihan Paket Layanan</FormLabel>
                  <FormControl>
                    <SelectDropdown
                      isControlled
                      defaultValue={field.value || undefined}
                      onValueChange={field.onChange}
                      placeholder={
                        plansLoading ? 'Memuat paket...' : 'Pilih paket layanan'
                      }
                      items={planOptions}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='address'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Alamat Lengkap Pemasangan</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='Jl. Merdeka No. 12, RT 02/RW 04, Kelurahan...'
                      rows={2}
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
                name='latitude'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className='text-xs'>Latitude (Opsional)</FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        step='any'
                        placeholder='-7.123456'
                        {...field}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='longitude'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className='text-xs'>Longitude (Opsional)</FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        step='any'
                        placeholder='110.123456'
                        {...field}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='notes'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Catatan Tambahan</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='Contoh: Pasang setelah jam kerja, dekat tiang Telkom'
                      {...field}
                    />
                  </FormControl>
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
            form='registration-submit-form'
            type='submit'
            disabled={submitMutation.isPending}
          >
            {submitMutation.isPending ? 'Mengirim...' : 'Simpan Pendaftaran'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
