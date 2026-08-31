import { CashierDialog } from './cashier-dialog'
import { GenerateInvoicesDialog } from './generate-invoices-dialog'
import { InvoiceDetailSheet } from './invoice-detail-sheet'
import { InvoicePrintDialog } from './invoice-print-dialog'
import { useInvoices, type InvoicesDialogType } from './invoices-provider'

export function InvoicesDialogs() {
  const { open, setOpen, currentInvoice } = useInvoices()

  const close = (dialog: InvoicesDialogType) => {
    setOpen(dialog)
  }

  return (
    <>
      {/* Dialog Kasir POS / Bayar Cepat */}
      {open === 'cashier' && (
        <CashierDialog
          open={open === 'cashier'}
          onOpenChange={() => close('cashier')}
          currentInvoice={currentInvoice}
        />
      )}

      {/* Dialog Generator Tagihan Bulanan */}
      {open === 'generate' && (
        <GenerateInvoicesDialog
          open={open === 'generate'}
          onOpenChange={() => close('generate')}
        />
      )}

      {/* Sheet Detail Faktur */}
      {currentInvoice && open === 'detail' && (
        <InvoiceDetailSheet
          key={`detail-${currentInvoice.id}`}
          open={open === 'detail'}
          onOpenChange={() => close('detail')}
          invoice={currentInvoice}
        />
      )}

      {/* Dialog Cetak Faktur / Kwitansi */}
      {currentInvoice && open === 'print' && (
        <InvoicePrintDialog
          key={`print-${currentInvoice.id}`}
          open={open === 'print'}
          onOpenChange={() => close('print')}
          invoice={currentInvoice}
        />
      )}
    </>
  )
}
