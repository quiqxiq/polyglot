import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import type { Device } from '@/gen/v1/device_pb'
import { ImportFileRequest } from '@/gen/v1/ispadmin_pb'
import { CheckCircle2, AlertCircle, Eye, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useImportFileMutation } from '../api/use-customer'
import { IMPORT_FORMATS } from '../data/constants'
import { importFileFormSchema, type ImportFileFormValues } from '../data/schema'

interface CustomersImportDialogFileTabProps {
  devices: Device[]
  devicesLoading: boolean
  onSuccess: () => void
}

export function CustomersImportDialogFileTab({
  devices,
  devicesLoading,
  onSuccess,
}: CustomersImportDialogFileTabProps) {
  const importFile = useImportFileMutation()
  const [fileDefaultDeviceId, setFileDefaultDeviceId] = useState<string>('')
  const [filePreview, setFilePreview] = useState<{
    rowsTotal: number
    previewRows: string[]
    validationErrors: string[]
  } | null>(null)
  const [isPreviewingFile, setIsPreviewingFile] = useState(false)

  const form = useForm<ImportFileFormValues>({
    resolver: zodResolver(importFileFormSchema),
    defaultValues: { file: undefined, format: 0, defaultDeviceId: '' },
  })

  const fileRef = form.register('file')

  const handleFilePreview = async () => {
    const files = form.getValues('file')
    if (!files || files.length === 0) {
      toast.error('Pilih file CSV atau XLSX terlebih dahulu')
      return
    }
    const file = files[0]
    setIsPreviewingFile(true)
    try {
      const payload = new Uint8Array(await file.arrayBuffer())
      const format = form.getValues('format')
      const res = await importFile.mutateAsync(
        new ImportFileRequest({
          payload,
          format,
          defaultDeviceId: fileDefaultDeviceId,
          dryRun: true,
        })
      )
      setFilePreview({
        rowsTotal: res.result?.rowsTotal || 0,
        previewRows: res.previewRows || [],
        validationErrors: res.validationErrors || [],
      })
      toast.success(
        `Validasi file selesai: ${res.result?.rowsTotal || 0} baris terdeteksi`
      )
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Gagal mempratinjau file'
      )
    } finally {
      setIsPreviewingFile(false)
    }
  }

  const onFileSubmit = async (values: ImportFileFormValues) => {
    const file = values.file[0]
    try {
      const payload = new Uint8Array(await file.arrayBuffer())
      const res = await importFile.mutateAsync(
        new ImportFileRequest({
          payload,
          format: values.format,
          defaultDeviceId: fileDefaultDeviceId,
          dryRun: false,
        })
      )
      const r = res.result
      if (!r) throw new Error('Import gagal')
      const plansNote =
        r.plansCreated > 0 ? `, ${r.plansCreated} paket dibuat` : ''
      toast.success(
        `Import selesai: ${r.customersCreated} pelanggan dibuat, ${r.customersUpdated} diperbarui, ${r.subscriptionsCreated} langganan aktif${plansNote}`
      )
      if (r.skipped.length > 0) {
        const preview = r.skipped.slice(0, 5).join('\n')
        const more = r.skipped.length - Math.min(5, r.skipped.length)
        toast.info(
          `${r.skipped.length} baris dilewati:\n${preview}${more > 0 ? `\n… dan ${more} lainnya` : ''}`
        )
      }
      form.reset()
      onSuccess()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Import gagal')
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onFileSubmit)}
        className='space-y-4 pt-2'
      >
        <FormField
          control={form.control}
          name='file'
          render={() => (
            <FormItem>
              <FormLabel>File Excel / CSV</FormLabel>
              <FormControl>
                <Input
                  type='file'
                  accept='.csv, .xlsx, .xls'
                  {...fileRef}
                  onChange={(e) => {
                    fileRef.onChange(e)
                    setFilePreview(null)
                  }}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='format'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Format file</FormLabel>
              <Select
                value={String(field.value)}
                onValueChange={(val) => {
                  field.onChange(Number(val))
                  setFilePreview(null)
                }}
              >
                <FormControl>
                  <SelectTrigger aria-label='Format file'>
                    <SelectValue placeholder='Pilih format file' />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {IMPORT_FORMATS.map((f) => (
                    <SelectItem key={f.value} value={String(f.value)}>
                      {f.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Router Default Opsional */}
        <div className='space-y-1.5'>
          <label className='text-xs font-semibold text-foreground'>
            Router Default (Opsional)
          </label>
          <Select
            value={fileDefaultDeviceId}
            onValueChange={setFileDefaultDeviceId}
          >
            <SelectTrigger className='w-full'>
              <SelectValue placeholder='Pilih router default bila baris file kosong' />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='none'>Tanpa router default</SelectItem>
              {devicesLoading ? (
                <SelectItem disabled value='loading'>
                  Memuat…
                </SelectItem>
              ) : (
                devices.map((d) => (
                  <SelectItem key={d.id} value={d.id}>
                    {d.name} ({d.host})
                  </SelectItem>
                ))
              )}
            </SelectContent>
          </Select>
        </div>

        {/* Tombol Pratinjau File */}
        <div className='flex justify-end'>
          <Button
            type='button'
            variant='outline'
            size='sm'
            onClick={handleFilePreview}
            disabled={isPreviewingFile || importFile.isPending}
            className='gap-1.5'
          >
            {isPreviewingFile ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : (
              <Eye className='h-4 w-4 text-primary' />
            )}
            <span>Pratinjau File</span>
          </Button>
        </div>

        {/* Hasil Pratinjau File */}
        {filePreview && (
          <div className='space-y-2 rounded-lg border border-primary/20 bg-primary/5 p-3'>
            <div className='flex items-center justify-between text-xs font-semibold text-foreground'>
              <span className='flex items-center gap-1 text-emerald-600 dark:text-emerald-400'>
                <CheckCircle2 className='h-4 w-4' /> Hasil Pratinjau:
              </span>
              <Badge variant='secondary' className='font-mono'>
                {filePreview.rowsTotal} Baris Ditemukan
              </Badge>
            </div>

            {filePreview.previewRows.length > 0 && (
              <ul className='space-y-1 font-mono text-[11px]'>
                {filePreview.previewRows.map((row, i) => (
                  <li
                    key={i}
                    className='truncate rounded border bg-background/80 px-2 py-1'
                  >
                    {row}
                  </li>
                ))}
              </ul>
            )}

            {filePreview.validationErrors.length > 0 && (
              <div className='space-y-1 rounded bg-destructive/10 p-2 text-xs text-destructive'>
                <div className='flex items-center gap-1 font-semibold'>
                  <AlertCircle className='h-3.5 w-3.5' /> Perhatian:
                </div>
                <ul className='list-disc pl-4 text-[11px]'>
                  {filePreview.validationErrors.slice(0, 3).map((err, i) => (
                    <li key={i}>{err}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        )}

        <div className='flex justify-end pt-2'>
          <Button type='submit' disabled={importFile.isPending}>
            {importFile.isPending && (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            )}
            Import
          </Button>
        </div>
      </form>
    </Form>
  )
}
