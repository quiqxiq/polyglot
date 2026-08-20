import { useEffect, useState } from 'react'
import { Printer, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { TEMPLATE_LAYOUTS } from '../../data/constants'
import { useRenderVouchersMutation } from '../../api/use-hotspot-vouchers'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function VoucherPrintDialog() {
  const {
    open,
    setOpen,
    printBatchComment,
    setPrintBatchComment,
    printSingleUserId,
    setPrintSingleUserId,
  } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()

  const isOpen = open === 'voucher-print'
  const [templateName, setTemplateName] = useState<string>('default')
  const [renderedHtml, setRenderedHtml] = useState<string>('')
  const [totalVouchers, setTotalVouchers] = useState<number>(0)

  const renderMutation = useRenderVouchersMutation()

  const fetchVouchersHtml = async (layout = templateName) => {
    if (!selectedDeviceId) return
    try {
      const res = await renderMutation.mutateAsync({
        deviceId: selectedDeviceId,
        templateName: layout,
        comment: printBatchComment,
        userId: printSingleUserId,
        preview: false,
      })
      setRenderedHtml(res.html)
      setTotalVouchers(res.totalVouchers)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to render vouchers')
    }
  }

  useEffect(() => {
    if (isOpen && selectedDeviceId) {
      fetchVouchersHtml(templateName)
    } else {
      setRenderedHtml('')
      setTotalVouchers(0)
    }
  }, [isOpen, selectedDeviceId, printBatchComment, printSingleUserId])

  const handleLayoutChange = (newLayout: string) => {
    setTemplateName(newLayout)
    fetchVouchersHtml(newLayout)
  }

  const handlePrint = () => {
    if (!renderedHtml) return
    const printWindow = window.open('', '_blank', 'width=800,height=600')
    if (!printWindow) {
      toast.error('Pop-up blocked! Please allow pop-ups to print vouchers.')
      return
    }
    printWindow.document.write(renderedHtml)
    printWindow.document.close()
    printWindow.focus()
    setTimeout(() => {
      printWindow.print()
    }, 500)
  }

  const title = printBatchComment
    ? `Print Batch: ${printBatchComment}`
    : 'Print Single Voucher'

  return (
    <Dialog open={isOpen} onOpenChange={(val) => !val && setOpen(null)}>
      <DialogContent className='sm:max-w-[720px] max-h-[90vh] flex flex-col'>
        <DialogHeader>
          <div className='flex items-center justify-between pr-6'>
            <div className='flex items-center gap-2'>
              <Printer className='size-5 text-primary' />
              <DialogTitle>{title}</DialogTitle>
            </div>
            <Select value={templateName} onValueChange={handleLayoutChange}>
              <SelectTrigger className='h-8 w-44 text-xs'>
                <SelectValue placeholder='Select Layout' />
              </SelectTrigger>
              <SelectContent>
                {TEMPLATE_LAYOUTS.map((t) => (
                  <SelectItem key={t.value} value={t.value}>
                    {t.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <DialogDescription>
            Preview and print generated vouchers. Total rendered:{' '}
            <span className='font-semibold text-foreground'>{totalVouchers}</span> vouchers.
          </DialogDescription>
        </DialogHeader>

        {/* Live Voucher Preview Container */}
        <div className='flex-1 min-h-[360px] max-h-[480px] w-full rounded-md border bg-white overflow-hidden relative shadow-inner'>
          {renderMutation.isPending ? (
            <div className='absolute inset-0 flex flex-col items-center justify-center bg-background/80 gap-2'>
              <RefreshCw className='size-6 animate-spin text-primary' />
              <span className='text-xs text-muted-foreground'>Rendering voucher template with QR codes...</span>
            </div>
          ) : renderedHtml ? (
            <iframe
              srcDoc={renderedHtml}
              title='Voucher Preview'
              className='w-full h-full border-0'
            />
          ) : (
            <div className='flex items-center justify-center h-full text-xs text-muted-foreground'>
              No voucher preview available.
            </div>
          )}
        </div>

        <DialogFooter className='flex flex-row justify-between items-center pt-2'>
          <Button
            type='button'
            variant='ghost'
            size='sm'
            onClick={() => fetchVouchersHtml(templateName)}
            disabled={renderMutation.isPending}
          >
            <RefreshCw className='mr-1.5 size-3.5' /> Reload
          </Button>

          <div className='flex gap-2'>
            <Button
              type='button'
              variant='outline'
              size='sm'
              onClick={() => {
                setOpen(null)
                setPrintBatchComment('')
                setPrintSingleUserId('')
              }}
            >
              Close
            </Button>
            <Button
              size='sm'
              onClick={handlePrint}
              disabled={!renderedHtml || renderMutation.isPending}
            >
              <Printer className='mr-1.5 size-3.5' />
              Print ({totalVouchers})
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
