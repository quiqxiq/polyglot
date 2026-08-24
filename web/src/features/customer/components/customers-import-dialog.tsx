import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
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
import { SelectDropdown } from '@/components/select-dropdown'
import { ImportFileRequest } from '@/gen/v1/ispadmin_pb'
import { IMPORT_FORMATS } from '../data/constants'
import { importFileFormSchema, type ImportFileFormValues } from '../data/schema'
import { useImportFileMutation } from '../api/use-customer'

const FORMAT_OPTIONS = IMPORT_FORMATS.map((f) => ({
  label: f.label,
  value: String(f.value),
}))

type CustomersImportDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CustomersImportDialog({
  open,
  onOpenChange,
}: CustomersImportDialogProps) {
  const importFile = useImportFileMutation()

  const form = useForm<ImportFileFormValues>({
    resolver: zodResolver(importFileFormSchema),
    defaultValues: { file: undefined, format: 0 },
  })

  const fileRef = form.register('file')

  const onSubmit = async (values: ImportFileFormValues) => {
    const file = values.file[0]
    try {
      const payload = new Uint8Array(await file.arrayBuffer())
      const res = await importFile.mutateAsync(
        new ImportFileRequest({ payload, format: values.format })
      )
      const r = res.result
      if (!r) throw new Error('Import gagal')
      const plansNote =
        r.plansCreated > 0 ? `, ${r.plansCreated} paket dibuat` : ''
      toast.success(
        `Import selesai: ${r.customersCreated} pelanggan dibuat, ${r.customersUpdated} diperbarui${plansNote}`
      )
      if (r.skipped.length > 0) {
        const preview = r.skipped.slice(0, 5).join('\n')
        const more = r.skipped.length - Math.min(5, r.skipped.length)
        toast.info(
          `${r.skipped.length} baris dilewati:\n${preview}${
            more > 0 ? `\n… dan ${more} lainnya` : ''
          }`
        )
      }
      form.reset()
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Import gagal')
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(val) => {
        onOpenChange(val)
        form.reset()
      }}
    >
      <DialogContent className='max-h-[85vh] overflow-y-auto gap-2 sm:max-w-sm'>
        <DialogHeader className='text-start'>
          <DialogTitle>Import Pelanggan</DialogTitle>
          <DialogDescription>
            Unggah file CSV atau XLSX hasil export Mikhmon/polyglot untuk
            membuat atau memperbarui data pelanggan secara massal.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form
            id='customer-import-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className='space-y-4'
          >
            <FormField
              control={form.control}
              name='format'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Format file</FormLabel>
                  <FormControl>
                    <SelectDropdown
                      defaultValue={String(field.value)}
                      onValueChange={(val) => field.onChange(Number(val))}
                      isControlled
                      placeholder='Pilih format'
                      items={FORMAT_OPTIONS}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='file'
              render={() => (
                <FormItem>
                  <FormLabel>File</FormLabel>
                  <FormControl>
                    <Input
                      type='file'
                      accept='.csv,.xlsx'
                      {...fileRef}
                      className='h-8 py-0'
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </form>
        </Form>
        <DialogFooter className='gap-2'>
          <DialogClose asChild>
            <Button variant='outline'>Cancel</Button>
          </DialogClose>
          <Button
            type='submit'
            form='customer-import-form'
            disabled={importFile.isPending}
          >
            {importFile.isPending ? 'Mengimpor…' : 'Import'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
