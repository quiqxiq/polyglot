import { useEffect } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { render } from 'vitest-browser-react'
import { userEvent } from 'vitest/browser'
import { PlansProvider, usePlans } from './plans-provider'
import { PlansMutateDialog } from './plans-mutate-dialog'

const { createMutateAsync, updateMutateAsync } = vi.hoisted(() => ({
  createMutateAsync: vi.fn(),
  updateMutateAsync: vi.fn(),
}))

vi.mock('../api/use-plans', async (orig) => {
  const actual = await orig<typeof import('../api/use-plans')>()
  return {
    ...actual,
    useCreatePlanMutation: () => ({
      mutateAsync: createMutateAsync,
      isPending: false,
    }),
    useUpdatePlanMutation: () => ({
      mutateAsync: updateMutateAsync,
      isPending: false,
    }),
  }
})

vi.mock('@/features/devices/api/use-devices', () => ({
  useDevicesQuery: () => ({
    data: [{ id: 'dev-1', name: 'Router-Utama', host: '192.168.88.1' }],
    isLoading: false,
    isPending: false,
  }),
}))

vi.mock('@/features/hotspot/api/use-router-resources', () => ({
  useParentQueuesQuery: () => ({
    data: ['pq-utama', 'pq-backup'],
    isFetching: false,
  }),
  useIpPoolsQuery: () => ({
    data: ['pool-a', 'pool-b'],
    isFetching: false,
  }),
}))

vi.mock('@/stores/device-store', () => ({
  useDeviceStore: (sel: (s: { selectedDeviceId: string }) => unknown) =>
    sel({ selectedDeviceId: 'dev-1' }),
}))

function CreateHarness() {
  const { setOpen } = usePlans()
  useEffect(() => {
    setOpen('create')
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])
  return <PlansMutateDialog />
}

describe('PlansMutateDialog', () => {
  it('opens the create dialog without crashing', async () => {
    const { getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket Layanan')).toBeInTheDocument()
    await expect
      .element(getByText(/Buat paket layanan baru/i))
      .toBeInTheDocument()
  })

  it('seeds the create form with safe numeric defaults', async () => {
    const { getByRole } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect
      .element(getByRole('spinbutton', { name: /Bandwidth Download/i }))
      .toHaveValue(10240)
    await expect
      .element(getByRole('spinbutton', { name: /Bandwidth Upload/i }))
      .toHaveValue(5120)
  })

  it('renders safely while the dialog is closed', async () => {
    const screen = await render(
      <PlansProvider>
        <PlansMutateDialog />
      </PlansProvider>
    )

    expect(screen.container).toBeInTheDocument()
    expect(screen.getByText('Tambah Paket Layanan').elements()).toHaveLength(0)
  })

  it('hides hotspot-only fields when service type defaults to PPPOE', async () => {
    const { getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket Layanan')).toBeInTheDocument()

    // HOTSPOT-only fields must be absent on the PPPOE default.
    expect(getByText(/Shared Users/i).elements()).toHaveLength(0)
    expect(getByText(/IP Pool Hotspot/i).elements()).toHaveLength(0)

    // PPPoE-relevant fields stay visible.
    await expect.element(getByText(/Parent Queue/i)).toBeInTheDocument()
    await expect.element(getByText(/Address List/i)).toBeInTheDocument()
    await expect.element(getByText(/Remote Address Pool/i)).toBeInTheDocument()
  })

  it('shows hotspot fields and hides pppoe routing after switching to HOTSPOT', async () => {
    const { getByRole, getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket Layanan')).toBeInTheDocument()

    const serviceTypeSelect = getByRole('combobox', { name: /Tipe Layanan/i })
    await userEvent.click(serviceTypeSelect)
    await userEvent.click(
      getByRole('option', { name: /Hotspot Permanent/i })
    )

    // HOTSPOT-specific field appears once the type switches.
    await expect.element(getByText(/Shared Users/i)).toBeInTheDocument()
    await expect.element(getByText(/IP Pool Hotspot/i)).toBeInTheDocument()

    // PPPoE-only routing fields disappear.
    expect(getByText(/Remote Address Pool/i).elements()).toHaveLength(0)
  })

  it('offers router parent queues and IP pools via select dropdowns on create', async () => {
    const { getByRole, getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket Layanan')).toBeInTheDocument()

    // Parent Queue is a proper Select dropdown seeded with 'none'.
    const pqTrigger = getByRole('combobox', { name: /Parent Queue/i })
    await userEvent.click(pqTrigger)
    await expect.element(getByRole('option', { name: 'pq-utama' })).toBeInTheDocument()
    await expect.element(getByRole('option', { name: 'pq-backup' })).toBeInTheDocument()
    await expect.element(getByRole('option', { name: 'none' })).toBeInTheDocument()
    await userEvent.click(getByRole('option', { name: 'pq-utama' }))
    await expect.element(pqTrigger).toHaveTextContent('pq-utama')

    // Switch to Hotspot
    const serviceTypeSelect = getByRole('combobox', { name: /Tipe Layanan/i })
    await userEvent.click(serviceTypeSelect)
    await userEvent.click(
      getByRole('option', { name: /Hotspot Permanent/i })
    )
    await expect.element(getByText(/Shared Users/i)).toBeInTheDocument()

    // IP Pool dropdown listing router pools.
    const poolTrigger = getByRole('combobox', { name: /IP Pool Hotspot/i })
    await userEvent.click(poolTrigger)
    await expect.element(getByRole('option', { name: 'pool-a' })).toBeInTheDocument()
    await expect.element(getByRole('option', { name: 'pool-b' })).toBeInTheDocument()
    await userEvent.click(getByRole('option', { name: 'pool-a' }))
    await expect.element(poolTrigger).toHaveTextContent('pool-a')
  })

  it('allows switching a queue selection back to none', async () => {
    const { getByRole, getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket Layanan')).toBeInTheDocument()

    const pqTrigger = getByRole('combobox', { name: /Parent Queue/i })
    await userEvent.click(pqTrigger)
    await userEvent.click(getByRole('option', { name: 'pq-backup' }))
    await expect.element(pqTrigger).toHaveTextContent('pq-backup')

    await userEvent.click(pqTrigger)
    await userEvent.click(getByRole('option', { name: 'none' }))
    await expect.element(pqTrigger).toHaveTextContent('none')
  })

  it('renders timeout fields and does not render manual router dropdown', async () => {
    const { getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket Layanan')).toBeInTheDocument()
    await expect.element(getByText(/Session Timeout/i)).toBeInTheDocument()
    await expect.element(getByText(/Idle Timeout/i)).toBeInTheDocument()

    // Ensure manual router dropdown is NOT present in the form
    expect(getByText(/Terapkan ke Router/i).elements()).toHaveLength(0)
  })

  it('submits create plan with global deviceId and timeouts', async () => {
    const { getByRole, getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket Layanan')).toBeInTheDocument()

    const nameInput = getByRole('textbox', { name: /Nama Paket/i })
    await userEvent.fill(nameInput, 'Paket Fiber 20M')

    const sessionTimeoutInput = getByRole('textbox', { name: /Session Timeout/i })
    await userEvent.fill(sessionTimeoutInput, '1d')

    const idleTimeoutInput = getByRole('textbox', { name: /Idle Timeout/i })
    await userEvent.fill(idleTimeoutInput, '10m')

    const saveButton = getByRole('button', { name: /Simpan Paket/i })
    await userEvent.click(saveButton)

    expect(createMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        deviceId: 'dev-1',
        plan: expect.objectContaining({
          name: 'Paket Fiber 20M',
          sessionTimeout: '1d',
          idleTimeout: '10m',
        }),
      })
    )
  })
})
