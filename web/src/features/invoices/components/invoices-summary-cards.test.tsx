import { describe, expect, it } from 'vitest'
import { render } from 'vitest-browser-react'
import { Invoice } from '@/gen/v1/billing_pb'
import { InvoicesSummaryCards } from './invoices-summary-cards'

const invoices = [
  new Invoice({
    id: 'inv-1',
    total: 300000,
    paidAmount: 300000,
    status: 'PAID',
  }),
  new Invoice({
    id: 'inv-2',
    total: 200000,
    paidAmount: 0,
    status: 'UNPAID',
  }),
  new Invoice({
    id: 'inv-3',
    total: 150000,
    paidAmount: 0,
    status: 'OVERDUE',
  }),
]

describe('InvoicesSummaryCards', () => {
  it('renders all invoice KPI metrics correctly', async () => {
    const { getByText } = await render(
      <InvoicesSummaryCards invoices={invoices} />
    )

    await expect.element(getByText('Total Nilai Faktur')).toBeInTheDocument()
    await expect.element(getByText('Tagihan Terkumpul')).toBeInTheDocument()
    await expect.element(getByText('Belum Lunas (Unpaid)')).toBeInTheDocument()
    await expect.element(getByText('Jatuh Tempo (Overdue)')).toBeInTheDocument()
  })
})
