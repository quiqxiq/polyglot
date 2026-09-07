import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import {
  Server,
  FileSpreadsheet,
  CheckCircle2,
  AlertCircle,
  Eye,
  Loader2,
  Info,
  ShieldCheck,
} from 'lucide-react'
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { SelectDropdown } from '@/components/select-dropdown'
import { ImportFileRequest, ImportRouterRequest } from '@/gen/v1/ispadmin_pb'
import { IMPORT_FORMATS } from '../data/constants'
import { importFileFormSchema, type ImportFileFormValues } from '../data/schema'
import { useImportFileMutation, useImportRouterMutation } from '../api/use-customer'
import { useDevicesQuery } from '@/features/devices/api/use-devices'

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
  const [activeTab, setActiveTab] = useState<'router' | 'file'>('file')

  // Mutations & Queries
  const importFile = useImportFileMutation()
  const importRouter = useImportRouterMutation()
  const { data: devices = [], isLoading: devicesLoading } = useDevicesQuery()

  // ─── State Metode A: Tarik Langsung dari Router ──────────────────────────
  const [selectedDeviceId, setSelectedDeviceId] = useState<string>('')
  const [pullPPPoE, setPullPPPoE] = useState<boolean>(true)
  const [pullHotspotMember, setPullHotspotMember] = useState<boolean>(true)
  const [pullHotspotIPBinding, setPullHotspotIPBinding] = useState<boolean>(true)
  const [pullHotspotVoucher, setPullHotspotVoucher] = useState<boolean>(false)
  const [routerPreview, setRouterPreview] = useState<{
    rowsTotal: number
    pppoeDetected: number
    hotspotPermanentDetected: number
    hotspotIpBindingDetected: number
    vouchersSkipped: number
    previewRows: string[]
    validationErrors: string[]
  } | null>(null)
  const [isPullingPreview, setIsPullingPreview] = useState(false)
  const [isExecutingRouterImport, setIsExecutingRouterImport] = useState(false)

  // ─── State Metode B: Import via File ──────────────────────────────────────
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

  const resetAll = () => {
    form.reset()
    setRouterPreview(null)
    setFilePreview(null)
    setIsPullingPreview(false)
    setIsExecutingRouterImport(false)
    setIsPreviewingFile(false)
  }

  // ─── Handler Metode A: Preview & Import Router ───────────────────────────
  const computeServiceType = () => {
    const hasPPPoE = pullPPPoE
    const hasHotspot = pullHotspotMember || pullHotspotIPBinding || pullHotspotVoucher
    if (hasPPPoE && hasHotspot) return 'ALL'
    if (hasPPPoE) return 'PPPOE'
    if (hasHotspot) return 'HOTSPOT'
    return 'ALL'
  }

  const handleRouterPreview = async () => {
    if (!selectedDeviceId) {
      toast.error('Pilih router target terlebih dahulu')
      return
    }
    const dev = devices.find((d) => d.id === selectedDeviceId)
    setIsPullingPreview(true)
    try {
      const res = await importRouter.mutateAsync(
        new ImportRouterRequest({
          deviceId: selectedDeviceId,
          deviceName: dev?.name || selectedDeviceId,
          dryRun: true,
          serviceType: computeServiceType(),
          includeIpBindings: pullHotspotIPBinding,
          includeVouchers: pullHotspotVoucher,
        })
      )
      setRouterPreview({
        rowsTotal: res.result?.rowsTotal || 0,
        pppoeDetected: res.pppoeDetected,
        hotspotPermanentDetected: res.hotspotPermanentDetected,
        hotspotIpBindingDetected: res.hotspotIpBindingDetected,
        vouchersSkipped: res.vouchersSkipped,
        previewRows: res.previewRows || [],
        validationErrors: res.validationErrors || [],
      })
      toast.success(`Ditemukan ${res.result?.rowsTotal || 0} akun aktif di router`)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal menarik data dari router')
    } finally {
      setIsPullingPreview(false)
    }
  }

  const handleRouterImportExecute = async () => {
    if (!selectedDeviceId) {
      toast.error('Pilih router target terlebih dahulu')
      return
    }
    const dev = devices.find((d) => d.id === selectedDeviceId)
    setIsExecutingRouterImport(true)
    try {
      const res = await importRouter.mutateAsync(
        new ImportRouterRequest({
          deviceId: selectedDeviceId,
          deviceName: dev?.name || selectedDeviceId,
          dryRun: false,
          serviceType: computeServiceType(),
          includeIpBindings: pullHotspotIPBinding,
          includeVouchers: pullHotspotVoucher,
        })
      )
      const r = res.result
      if (!r) throw new Error('Import gagal')
      toast.success(
        `Import Selesai: ${r.customersCreated} pelanggan baru, ${r.subscriptionsCreated} langganan terhubung ke router`
      )
      resetAll()
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal menyimpan data import')
    } finally {
      setIsExecutingRouterImport(false)
    }
  }

  // ─── Handler Metode B: Preview File ──────────────────────────────────────
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
      toast.success(`Validasi file selesai: ${res.result?.rowsTotal || 0} baris terdeteksi`)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal mempratinjau file')
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
      const plansNote = r.plansCreated > 0 ? `, ${r.plansCreated} paket dibuat` : ''
      toast.success(
        `Import selesai: ${r.customersCreated} pelanggan dibuat, ${r.customersUpdated} diperbarui, ${r.subscriptionsCreated} langganan aktif${plansNote}`
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
      resetAll()
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
        resetAll()
      }}
    >
      <DialogContent className='max-h-[90vh] overflow-y-auto gap-4 sm:max-w-2xl'>
        <DialogHeader className='text-start'>
          <div className='flex items-center gap-2'>
            <div className='flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary'>
              <Server className='h-5 w-5' />
            </div>
            <div>
              <DialogTitle className='text-lg'>Import Pelanggan</DialogTitle>
              <DialogDescription>
                Hubungkan data pelanggan dan langganan aktif ke Router MikroTik fisik.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <Tabs
          value={activeTab}
          onValueChange={(val) => setActiveTab(val as 'router' | 'file')}
          className='w-full'
        >
          <TabsList className='grid w-full grid-cols-2'>
            <TabsTrigger value='router' className='gap-2'>
              <Server className='h-4 w-4' />
              <span>Metode A: Tarik dari Router</span>
            </TabsTrigger>
            <TabsTrigger value='file' className='gap-2'>
              <FileSpreadsheet className='h-4 w-4' />
              <span>Metode B: Upload Excel / CSV</span>
            </TabsTrigger>
          </TabsList>

          {/* ══════════════════ TAB 1: METODE A (LIVE PULL ROUTER) ══════════════════ */}
          <TabsContent value='router' className='space-y-4 pt-2'>
            <div className='rounded-lg border bg-muted/30 p-3 text-xs text-muted-foreground flex items-start gap-2'>
              <ShieldCheck className='h-4 w-4 text-emerald-500 shrink-0 mt-0.5' />
              <div>
                <span className='font-semibold text-foreground'>Aman & Non-Destruktif:</span> Penarikan
                data bersifat <span className='underline font-mono'>read-only</span>. Status akun
                diset ke <Badge variant='outline' className='text-[10px] px-1 py-0 border-emerald-500 text-emerald-600'>PROVISION_OK</Badge> sehingga
                tidak memutus atau menimpa kredensial router yang sedang aktif di lapangan.
              </div>
            </div>

            {/* Pemilihan Router */}
            <div className='space-y-1.5'>
              <label className='text-xs font-semibold text-foreground'>Pilih Router Target</label>
              <Select
                value={selectedDeviceId}
                onValueChange={(val) => {
                  setSelectedDeviceId(val)
                  setRouterPreview(null)
                }}
              >
                <SelectTrigger className='w-full'>
                  <SelectValue placeholder='Pilih Router MikroTik tujuan' />
                </SelectTrigger>
                <SelectContent>
                  {devicesLoading ? (
                    <SelectItem disabled value='loading'>
                      Memuat daftar router…
                    </SelectItem>
                  ) : devices.length === 0 ? (
                    <SelectItem disabled value='empty'>
                      Tidak ada router terhubung
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

            {/* Checklist Sumber Akun */}
            <div className='space-y-2 rounded-lg border p-3 bg-background'>
              <div className='text-xs font-semibold text-foreground flex items-center justify-between'>
                <span>Jenis Akun yang Ditarik</span>
                <span className='text-[11px] font-normal text-muted-foreground'>
                  Klasifikasi Otomatis
                </span>
              </div>
              <div className='grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs'>
                <label className='flex items-center gap-2 cursor-pointer p-1.5 rounded hover:bg-muted/50 transition-colors'>
                  <Checkbox
                    checked={pullPPPoE}
                    onCheckedChange={(c) => setPullPPPoE(Boolean(c))}
                  />
                  <span>PPPoE Secrets (/ppp/secret)</span>
                </label>
                <label className='flex items-center gap-2 cursor-pointer p-1.5 rounded hover:bg-muted/50 transition-colors'>
                  <Checkbox
                    checked={pullHotspotMember}
                    onCheckedChange={(c) => setPullHotspotMember(Boolean(c))}
                  />
                  <span>Hotspot Member (Tanpa Script Expire)</span>
                </label>
                <label className='flex items-center gap-2 cursor-pointer p-1.5 rounded hover:bg-muted/50 transition-colors'>
                  <Checkbox
                    checked={pullHotspotIPBinding}
                    onCheckedChange={(c) => setPullHotspotIPBinding(Boolean(c))}
                  />
                  <span>IP Binding Statis (Bypass Portal)</span>
                </label>
                <label className='flex items-center gap-2 cursor-pointer p-1.5 rounded hover:bg-muted/50 transition-colors text-muted-foreground'>
                  <Checkbox
                    checked={pullHotspotVoucher}
                    onCheckedChange={(c) => setPullHotspotVoucher(Boolean(c))}
                  />
                  <span>Voucher Sementara (Mikhmon)</span>
                </label>
              </div>
              {!pullHotspotVoucher && (
                <p className='text-[11px] text-muted-foreground flex items-center gap-1 pt-1'>
                  <Info className='h-3 w-3 text-blue-500 shrink-0' />
                  Voucher sementara berdurasi beberapa jam otomatis dilewati agar database tagihan bulanan tetap bersih.
                </p>
              )}
            </div>

            {/* Tombol Pratinjau Router */}
            <div className='flex justify-end'>
              <Button
                type='button'
                variant='outline'
                size='sm'
                onClick={handleRouterPreview}
                disabled={!selectedDeviceId || isPullingPreview || isExecutingRouterImport}
                className='gap-1.5'
              >
                {isPullingPreview ? (
                  <Loader2 className='h-4 w-4 animate-spin' />
                ) : (
                  <Eye className='h-4 w-4 text-primary' />
                )}
                <span>Tarik & Pratinjau Akun</span>
              </Button>
            </div>

            {/* Hasil Pratinjau Router */}
            {routerPreview && (
              <div className='space-y-3 rounded-lg border border-primary/20 bg-primary/5 p-3'>
                <div className='flex items-center justify-between text-xs font-semibold text-foreground'>
                  <span className='flex items-center gap-1 text-emerald-600 dark:text-emerald-400'>
                    <CheckCircle2 className='h-4 w-4' /> Hasil Deteksi Router:
                  </span>
                  <Badge variant='secondary' className='font-mono'>
                    Total: {routerPreview.rowsTotal} Akun
                  </Badge>
                </div>

                <div className='grid grid-cols-2 sm:grid-cols-4 gap-2 text-center text-xs'>
                  <div className='rounded bg-background p-2 border shadow-xs'>
                    <div className='text-muted-foreground text-[11px]'>PPPoE</div>
                    <div className='text-base font-bold text-blue-600'>{routerPreview.pppoeDetected}</div>
                  </div>
                  <div className='rounded bg-background p-2 border shadow-xs'>
                    <div className='text-muted-foreground text-[11px]'>Hotspot Member</div>
                    <div className='text-base font-bold text-purple-600'>{routerPreview.hotspotPermanentDetected}</div>
                  </div>
                  <div className='rounded bg-background p-2 border shadow-xs'>
                    <div className='text-muted-foreground text-[11px]'>IP Binding</div>
                    <div className='text-base font-bold text-emerald-600'>{routerPreview.hotspotIpBindingDetected}</div>
                  </div>
                  <div className='rounded bg-background p-2 border shadow-xs'>
                    <div className='text-muted-foreground text-[11px]'>Voucher Dilewati</div>
                    <div className='text-base font-bold text-muted-foreground'>{routerPreview.vouchersSkipped}</div>
                  </div>
                </div>

                {routerPreview.previewRows.length > 0 && (
                  <div className='space-y-1.5'>
                    <div className='text-[11px] font-semibold text-muted-foreground'>
                      Sampel Akun Ditemukan ({routerPreview.previewRows.length}):
                    </div>
                    <div className='max-h-36 overflow-y-auto rounded bg-background p-2 border text-[11px] font-mono space-y-1'>
                      {routerPreview.previewRows.map((row, idx) => (
                        <div key={idx} className='text-foreground/90 border-b border-muted/50 pb-0.5 last:border-0'>
                          {row}
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {routerPreview.validationErrors.length > 0 && (
                  <div className='rounded bg-destructive/10 border border-destructive/20 p-2 text-destructive text-xs space-y-1'>
                    <div className='font-semibold flex items-center gap-1'>
                      <AlertCircle className='h-3.5 w-3.5' /> Perhatian:
                    </div>
                    {routerPreview.validationErrors.map((err, idx) => (
                      <div key={idx}>• {err}</div>
                    ))}
                  </div>
                )}

                <div className='pt-2 flex justify-end'>
                  <Button
                    type='button'
                    onClick={handleRouterImportExecute}
                    disabled={isExecutingRouterImport || routerPreview.rowsTotal === 0}
                    className='gap-2 bg-emerald-600 hover:bg-emerald-700 text-white'
                  >
                    {isExecutingRouterImport ? (
                      <>
                        <Loader2 className='h-4 w-4 animate-spin' />
                        <span>Menyimpan ke Database…</span>
                      </>
                    ) : (
                      <>
                        <CheckCircle2 className='h-4 w-4' />
                        <span>Konfirmasi & Simpan {routerPreview.rowsTotal} Akun</span>
                      </>
                    )}
                  </Button>
                </div>
              </div>
            )}
          </TabsContent>

          {/* ══════════════════ TAB 2: METODE B (FILE SPREADSHEET) ══════════════════ */}
          <TabsContent value='file' className='space-y-4 pt-2'>
            <Form {...form}>
              <form
                id='customer-import-form'
                onSubmit={form.handleSubmit(onFileSubmit)}
                className='space-y-4'
              >
                {/* Target Router Default */}
                <div className='space-y-1.5'>
                  <label className='text-xs font-semibold text-foreground'>
                    Target Router Default <span className='text-muted-foreground font-normal'>(Opsional)</span>
                  </label>
                  <Select
                    value={fileDefaultDeviceId}
                    onValueChange={(val) => setFileDefaultDeviceId(val)}
                  >
                    <SelectTrigger className='w-full'>
                      <SelectValue placeholder='Gunakan jika file Excel tidak memuat kolom router' />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value='none'>-- Tanpa Router Default --</SelectItem>
                      {devices.map((d) => (
                        <SelectItem key={d.id} value={d.id}>
                          {d.name} ({d.host})
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <p className='text-[11px] text-muted-foreground'>
                    Jika kolom <span className='font-mono'>server</span> atau <span className='font-mono'>router</span> di file spreadsheet kosong, data akan otomatis dikaitkan ke router ini.
                  </p>
                </div>

                <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
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
                        <FormLabel>File Excel / CSV</FormLabel>
                        <FormControl>
                          <Input
                            type='file'
                            accept='.csv,.xlsx'
                            {...fileRef}
                            className='h-9 py-1'
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <div className='flex justify-between items-center pt-1'>
                  <div className='text-[11px] text-muted-foreground'>
                    Dukungan kolom: <span className='font-mono'>nama, no_hp, tipe, server, paket, harga</span>
                  </div>
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
                      <Eye className='h-4 w-4' />
                    )}
                    <span>Pratinjau File</span>
                  </Button>
                </div>

                {/* Pratinjau File */}
                {filePreview && (
                  <div className='rounded-lg border bg-muted/20 p-3 space-y-2 text-xs'>
                    <div className='flex items-center justify-between font-semibold text-foreground'>
                      <span>Ringkasan Validasi File:</span>
                      <Badge variant='secondary'>{filePreview.rowsTotal} Baris Terdeteksi</Badge>
                    </div>
                    {filePreview.previewRows.length > 0 && (
                      <div className='max-h-28 overflow-y-auto rounded bg-background p-2 border text-[11px] font-mono space-y-0.5'>
                        {filePreview.previewRows.map((r, idx) => (
                          <div key={idx} className='truncate text-foreground/80'>{r}</div>
                        ))}
                      </div>
                    )}
                    {filePreview.validationErrors.length > 0 && (
                      <div className='rounded bg-destructive/10 p-2 text-destructive space-y-0.5 text-[11px]'>
                        {filePreview.validationErrors.map((err, idx) => (
                          <div key={idx}>• {err}</div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </form>
            </Form>
          </TabsContent>
        </Tabs>

        <DialogFooter className='gap-2 pt-2 border-t'>
          <DialogClose asChild>
            <Button variant='outline'>Batal</Button>
          </DialogClose>
          {activeTab === 'file' && (
            <Button
              type='submit'
              form='customer-import-form'
              disabled={importFile.isPending}
            >
              {importFile.isPending ? 'Mengimpor…' : 'Import'}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
