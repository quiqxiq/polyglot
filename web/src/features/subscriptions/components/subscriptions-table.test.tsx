import { describe, expect, it, vi } from 'vitest'
import { render } from 'vitest-browser-react'
import { Subscription } from '@/gen/v1/billing_pb'
import { SubscriptionsProvider } from './subscriptions-provider'
import { SubscriptionsTable } from './subscriptions-table'

vi.mock('@/features/customer/api/use-customer', () => ({
  useCustomersQuery: () => ({
    data: [
      { id: 'c-001', name: 'Budi Santoso' },
      { id: 'c-002', name: 'Siti Rahayu' },
    ],
    isPending: false,
  }),
}))

vi.mock('@/features/billing/api/use-plans', () => ({
  usePlansQuery: () => ({
    data: [
      { id: 'p-001', name: 'Home 20M' },
      { id: 'p-002', name: 'Hotspot Bulanan' },
    ],
    isPending: false,
  }),
}))

const subscriptions = [
  new Subscription({
    id: 's-001',
    customerId: 'c-001',
    planId: 'p-001',
    deviceId: 'dev-mikrotik-1',
    serviceType: 'PPPOE',
    remoteUsername: 'budi20m',
    rateLimit: '20M/20M',
    provisionStatus: 'OK',
    status: 'ACTIVE',
    billingCycle: 'MONTHLY',
  }),
  new Subscription({
    id: 's-002',
    customerId: 'c-002',
    planId: 'p-002',
    serviceType: 'HOTSPOT',
    remoteUsername: 'siti.hot',
    provisionStatus: 'FAILED',
    status: 'ISOLATED',
    billingCycle: 'MONTHLY',
  }),
]

function renderTable(props: { data: Subscription[]; isLoading?: boolean }) {
  return render(
    <SubscriptionsProvider>
      <SubscriptionsTable {...props} />
    </SubscriptionsProvider>
  )
}

describe('SubscriptionsTable', () => {
  it('renders both subscriptions with customer and plan names resolved', async () => {
    const { getByText } = await renderTable({ data: subscriptions })

    await expect.element(getByText('Budi Santoso')).toBeInTheDocument()
    await expect.element(getByText('Siti Rahayu')).toBeInTheDocument()
    await expect.element(getByText('Home 20M')).toBeInTheDocument()
    await expect.element(getByText('Hotspot Bulanan')).toBeInTheDocument()
  })

  it('renders status badges from subscription meta', async () => {
    const { getByText } = await renderTable({ data: subscriptions })

    await expect.element(getByText('Active', { exact: true })).toBeInTheDocument()
    await expect.element(getByText('Isolated', { exact: true })).toBeInTheDocument()
  })

  it('shows the empty state when there are no rows', async () => {
    const { getByText } = await renderTable({ data: [] })

    await expect.element(getByText('Belum ada langganan')).toBeInTheDocument()
  })
})
