import { describe, expect, it } from 'vitest'
import { render } from 'vitest-browser-react'
import { Invoice } from '@/gen/v1/billing_pb'
import { Customer } from '@/gen/v1/customer_pb'
import { InvoicesProvider } from './invoices-provider'
import { InvoicesTable } from './invoices-table'

const customers = [
  new Customer({
    id: 'c-001',
    customerCode: 'CUST-001',
    name: 'Budi Santoso',
    phone: '081234567890',
  }),
  new Customer({
    id: 'c-002',
    customerCode: 'CUST-002',
    name: 'Siti Rahayu',
    phone: '081298765432',
  }),
]

const invoices = [
  new Invoice({
    id: 'inv-001',
    invoiceNumber: 'INV-202609-001',
    customerId: 'c-001',
    period: '2026-09',
    total: 250000,
    paidAmount: 0,
    status: 'UNPAID',
    manualPaymentCode: '891024',
  }),
  new Invoice({
    id: 'inv-002',
    invoiceNumber: 'INV-202609-002',
    customerId: 'c-002',
    period: '2026-09',
    total: 175000,
    paidAmount: 175000,
    status: 'PAID',
    manualPaymentCode: '891025',
  }),
]

function renderTable(props: {
  data: Invoice[]
  customers: Customer[]
  isLoading?: boolean
}) {
  return render(
    <InvoicesProvider>
      <InvoicesTable {...props} />
    </InvoicesProvider>
  )
}

describe('InvoicesTable', () => {
  it('renders invoices with customer names and codes resolved', async () => {
    const { getByText } = await renderTable({
      data: invoices,
      customers,
    })

    await expect.element(getByText('INV-202609-001')).toBeInTheDocument()
    await expect.element(getByText('INV-202609-002')).toBeInTheDocument()
    await expect.element(getByText('Budi Santoso')).toBeInTheDocument()
    await expect.element(getByText('Siti Rahayu')).toBeInTheDocument()
  })

  it('renders status badges from invoice meta', async () => {
    const { getByText } = await renderTable({
      data: invoices,
      customers,
    })

    await expect.element(getByText('Belum Bayar')).toBeInTheDocument()
    await expect.element(getByText('Lunas')).toBeInTheDocument()
  })

  it('shows empty state when there are no invoices matching filter', async () => {
    const { getByText } = await renderTable({
      data: [],
      customers,
    })

    await expect.element(getByText('Belum ada faktur tagihan yang sesuai filter')).toBeInTheDocument()
  })
})
