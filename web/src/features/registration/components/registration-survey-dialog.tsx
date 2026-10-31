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
import { SelectDropdown } from '@/components/select-dropdown'
import { useRecordSurveyMutation } from '../api/use-registration'
import { recordSurveySchema, type RecordSurveyValues } from '../data/schema'
import { RecordSurveyRequest, type Registration } from '@/gen/v1/registration_pb'

interface RegistrationSurveyDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  registration: Registration | null
}

export function RegistrationSurveyDialog({
  open,
  onOpenChange,
  registration,
}: RegistrationSurveyDialogProps) {
  const recordSurveyMutation = useRecordSurveyMutation()

  const form = useForm<RecordSurveyValues>({
    resolver: zodResolver(recordSurveySchema) as never,
    defaultValues: {
      id: registration?.id || '',
      surveyDate: new Date().toISOString().split('T')[0],
      surveyTimeHhmm: '09:00',
      surveyStatus: 'ELIGIBLE',
      surveyNotes: '',
      estimatedCableLengthMeter: 0,
    },
  })

  useEffect(() => {
    if (registration) {
      let defaultDate = new Date().toISOString().split('T')[0]
      if (registration.surveyDateUnix) {
        const unix = Number(registration.surveyDateUnix)
        if (unix) defaultDate = new Date(unix * 1000).toISOString().split('T')[0]
      }
      form.reset({
        id: registration.id,
        surveyDate: defaultDate,
        surveyTimeHhmm: registration.surveyTime || '09:00',
        surveyStatus:
          registration.surveyStatus === 'NOT_ELIGIBLE'
            ? 'NOT_ELIGIBLE'
            : 'ELIGIBLE',
        surveyNotes: registration.surveyNotes || '',
        estimatedCableLengthMeter:
          registration.estimatedCableLengthMeter || 0,
      })
    }
  }, [registration, form])

  const handleSubmit = async (values: RecordSurveyValues) => {
    if (!registration) return
    try {
      await recordSurveyMutation.mutateAsync(
        new RecordSurveyRequest({
          id: registration.id,
          surveyDateUnix: values.surveyDate
            ? BigInt(Math.floor(new Date(values.surveyDate).getTime() / 1000))
            : protoZero,
          surveyTimeHhmm: values.surveyTimeHhmm || '',
          surveyStatus: values.surveyStatus,
          surveyNotes: values.surveyNotes || '',
          estimatedCableLengthMeter: values.estimatedCableLengthMeter || 0,
        })
      )
      toast.success('Hasil survey dicatat! Status: SURVEY')
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal mencatat survey')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>Survey Lokasi Calon Pelanggan</DialogTitle>
          <DialogDescription>
            Catat hasil pengecekan lokasi untuk{' '}
            <span className='font-semibold'>{registration?.fullName}</span>:
            kelayakan, estimasi panjang kabel, dan catatan teknis.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id='registration-survey-form'
            onSubmit={form.handleSubmit(handleSubmit)}
            className='flex flex-col gap-3.5 py-1'
          >
            <div className='grid grid-cols-2 gap-3'>
              <FormField
                control={form.control}
                name='surveyDate'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Tanggal Survey</FormLabel>
                    <FormControl>
                      <Input type='date' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='surveyTimeHhmm'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Jam Survey</FormLabel>
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
              name='surveyStatus'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Kelayakan Lokasi</FormLabel>
                  <FormControl>
                    <SelectDropdown
                      isControlled
                      defaultValue={field.value}
                      onValueChange={field.onChange}
                      items={[
                        { label: 'Layak (ELIGIBLE)', value: 'ELIGIBLE' },
                        {
                          label: 'Tidak Layak (NOT_ELIGIBLE)',
                          value: 'NOT_ELIGIBLE',
                        },
                      ]}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='estimatedCableLengthMeter'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Estimasi Panjang Kabel (meter)</FormLabel>
                  <FormControl>
                    <Input type='number' min={0} {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='surveyNotes'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Catatan Survey</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='Kondisi tiang, jarak ODP, port tersedia, redaman, dsb.'
                      rows={3}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter className='pt-2'>
              <Button
                type='button'
                variant='outline'
                onClick={() => onOpenChange(false)}
              >
                Batal
              </Button>
              <Button type='submit' disabled={recordSurveyMutation.isPending}>
                {recordSurveyMutation.isPending ? 'Menyimpan...' : 'Simpan Survey'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}

// protoZero adalah bigint 0 untuk field unix opsional protobuf.
const protoZero = BigInt(0)
