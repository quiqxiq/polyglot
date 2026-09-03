import { describe, expect, it } from 'vitest'
import { render } from 'vitest-browser-react'
import { Plan } from '@/gen/v1/plan_pb'
import { PlansProvider } from './plans-provider'
import { PlansTable } from './plans-table'

const plans = [
  new Plan({
    id: 'p-001',
    name: 'Home 20M',
    serviceType: 'PPPOE',
    bandwidthDownloadKbps: 20480,
    bandwidthUploadKbps: 20480,
    burstDownloadKbps: 0,
    price: 250000,
    isActive: true,
  }),
  new Plan({
    id: 'p-002',
    name: 'Voucher Hotspot 3 Jam',
    serviceType: 'HOTSPOT',
    bandwidthDownloadKbps: 4096,
    bandwidthUploadKbps: 2048,
    burstDownloadKbps: 6144,
    price: 15000,
    isActive: false,
  }),
]

function renderTable(props: { data: Plan[]; isLoading?: boolean }) {
  return render(
    <PlansProvider>
      <PlansTable {...props} />
    </PlansProvider>
  )
}

describe('PlansTable', () => {
  it('renders both plans with their names', async () => {
    const { getByText } = await renderTable({ data: plans })

    await expect.element(getByText('Home 20M')).toBeInTheDocument()
    await expect.element(getByText('Voucher Hotspot 3 Jam')).toBeInTheDocument()
  })

  it('renders service type badges per plan meta', async () => {
    const { getByText } = await renderTable({ data: plans })

    await expect
      .element(getByText('PPPOE', { exact: true }))
      .toBeInTheDocument()
    await expect
      .element(getByText('HOTSPOT', { exact: true }))
      .toBeInTheDocument()
  })

  it('formats bandwidth as compact rate pairs', async () => {
    const { getByText } = await renderTable({ data: plans })

    await expect.element(getByText('20M / 20M')).toBeInTheDocument()
    await expect.element(getByText('4M / 2M')).toBeInTheDocument()
  })

  it('shows the burst indicator only for plans with burst configured', async () => {
    const { getByText } = await renderTable({ data: plans })

    // exact:true so the 'Burst' column header doesn't collide with the badge
    const burstBadges = getByText('burst', { exact: true })
    await expect.element(burstBadges).toBeInTheDocument()
    expect(burstBadges.all().length).toBe(1)
  })

  it('shows the empty state when there is no data', async () => {
    const { getByText } = await renderTable({ data: [] })

    await expect.element(getByText('Belum ada paket layanan')).toBeInTheDocument()
  })
})
