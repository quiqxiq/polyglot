import { describe, expect, it } from 'vitest'
import { render } from 'vitest-browser-react'
import { Customer } from '@/gen/v1/customer_pb'
import { CustomersProvider } from './customers-provider'
import { CustomersTable } from './customers-table'

const customers = [
  new Customer({
    id: 'c-001',
    customerCode: 'PLG-001',
    name: 'Budi Santoso',
    phone: '081234567890',
    email: 'budi@example.com',
    status: 'ACTIVE',
    portalAccessCode: 'PRT-777',
    registeredAtUnix: 1_750_000_000n,
  }),
  new Customer({
    id: 'c-002',
    customerCode: 'PLG-002',
    name: 'Siti Rahayu',
    phone: '089876543210',
    status: 'SUSPENDED',
    registeredAtUnix: 0n,
  }),
]

function renderTable(props: { data: Customer[]; isLoading?: boolean }) {
  return render(
    <CustomersProvider>
      <CustomersTable {...props} />
    </CustomersProvider>
  )
}

describe('CustomersTable', () => {
  it('renders both customers with their code and name', async () => {
    const { getByText } = await renderTable({ data: customers })

    await expect.element(getByText('Budi Santoso')).toBeInTheDocument()
    await expect.element(getByText('Siti Rahayu')).toBeInTheDocument()
    await expect.element(getByText('PLG-001')).toBeInTheDocument()
  })

  it('renders status badges from customer meta', async () => {
    const { getByText } = await renderTable({ data: customers })

    await expect
      .element(getByText('Active', { exact: true }))
      .toBeInTheDocument()
    await expect
      .element(getByText('Suspended', { exact: true }))
      .toBeInTheDocument()
  })

  it('shows placeholder dash for missing optional fields', async () => {
    const { getByText } = await renderTable({ data: [customers[1]] })

    // Siti has no email and no registered date → at least two '-' cells
    await expect.element(getByText('-').all()[0]).toBeInTheDocument()
    expect(getByText('-').all().length).toBeGreaterThanOrEqual(2)
  })

  it('shows the empty state when there is no data', async () => {
    const { getByText } = await renderTable({ data: [] })

    await expect.element(getByText('Belum ada pelanggan')).toBeInTheDocument()
  })
})
