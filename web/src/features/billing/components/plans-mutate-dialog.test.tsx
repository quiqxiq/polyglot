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
  // Call once: useDialogState's setOpen toggles (same value → null), so
  // re-invoking it on every render would flip the dialog closed again.
  useEffect(() => {
    setOpen('create')
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])
  return <PlansMutateDialog />
}

describe('PlansMutateDialog', () => {
  // Regression: defaultFormValues used planFormSchema.parse({}) for the
  // create case, which threw (name/serviceType/bandwidth/price are required)
  // during render — even while the dialog was closed — crashing the whole
  // Service Plans page.
  it('opens the create dialog without crashing', async () => {
    const { getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket')).toBeInTheDocument()
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
    // The dialog stays mounted with open === null; its render-time
    // defaultFormValues call must never throw.
    const screen = await render(
      <PlansProvider>
        <PlansMutateDialog />
      </PlansProvider>
    )

    expect(screen.container).toBeInTheDocument()
    expect(screen.getByText('Tambah Paket').elements()).toHaveLength(0)
  })

  it('hides hotspot-only fields when service type defaults to PPPOE', async () => {
    const { getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket')).toBeInTheDocument()

    // HOTSPOT-only voucher fields must be absent on the PPPOE default.
    expect(getByText(/Expire Mode/i).elements()).toHaveLength(0)
    expect(getByText(/Shared Users/i).elements()).toHaveLength(0)

    // PPPoE-relevant fields stay visible.
    await expect.element(getByText(/Parent Queue/i)).toBeInTheDocument()
    await expect.element(getByText(/Validity Mode/i)).toBeInTheDocument()
  })

  it('shows hotspot fields and hides pppoe routing after switching to HOTSPOT', async () => {
    const { getByRole, getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket')).toBeInTheDocument()

    const serviceTypeSelect = getByRole('combobox', { name: /Tipe Layanan/i })
    await userEvent.click(serviceTypeSelect)
    await userEvent.click(
      getByRole('option', { name: /Hotspot \(Voucher\/Monthly\)/i })
    )

    // HOTSPOT-specific field appears once the type switches.
    await expect.element(getByText(/Expire Mode/i)).toBeInTheDocument()

    // PPPoE routing fields disappear.
    expect(getByText(/Address List/i).elements()).toHaveLength(0)
    expect(getByText(/Simultaneous Use/i).elements()).toHaveLength(0)
  })

  it('offers router parent queues and IP pools via datalist on create (PPPOE)', async () => {
    const { getByRole, getByPlaceholder, getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket')).toBeInTheDocument()

    // Parent Queue input is linked to a datalist listing the router's
    // simple queues (query by placeholder: list= inputs map to combobox).
    const pqInput = getByPlaceholder('none / parent queue')
    await expect.element(pqInput).toHaveAttribute('list')

    const pqListId = pqInput.element().getAttribute('list')
    expect(pqListId).toBeTruthy()
    const pqDatalist = document.getElementById(pqListId!)
    expect(pqDatalist?.tagName.toLowerCase()).toBe('datalist')
    const queueValues = Array.from(
      pqDatalist!.querySelectorAll('option')
    ).map((opt) => opt.getAttribute('value'))
    expect(queueValues).toContain('pq-utama')
    expect(queueValues).toEqual(['pq-utama', 'pq-backup'])

    // ipPoolName is HOTSPOT-only in the visibility matrix (hidden on the
    // PPPOE default), so switch type before checking its datalist.
    const serviceTypeSelect = getByRole('combobox', { name: /Tipe Layanan/i })
    await userEvent.click(serviceTypeSelect)
    await userEvent.click(
      getByRole('option', { name: /Hotspot \(Voucher\/Monthly\)/i })
    )
    await expect.element(getByText(/Expire Mode/i)).toBeInTheDocument()

    // IP Pool input gets its own datalist with the router's pools.
    const poolInput = getByPlaceholder('none / pool-hotspot')
    await expect.element(poolInput).toHaveAttribute('list')

    const poolListId = poolInput.element().getAttribute('list')
    expect(poolListId).toBeTruthy()
    expect(poolListId).not.toBe(pqListId)
    const poolDatalist = document.getElementById(poolListId!)
    expect(poolDatalist?.tagName.toLowerCase()).toBe('datalist')
    const poolValues = Array.from(
      poolDatalist!.querySelectorAll('option')
    ).map((opt) => opt.getAttribute('value'))
    expect(poolValues).toContain('pool-a')
    expect(poolValues).toEqual(['pool-a', 'pool-b'])
  })

  it('still accepts freely typed values outside the datalist suggestions', async () => {
    const { getByPlaceholder, getByText } = await render(
      <PlansProvider>
        <CreateHarness />
      </PlansProvider>
    )

    await expect.element(getByText('Tambah Paket')).toBeInTheDocument()

    // Free-typing an unknown parent queue must work without errors —
    // datalist only suggests, it never restricts. fill() replaces the
    // seeded 'none' default instead of appending to it.
    const pqInput = getByPlaceholder('none / parent queue')
    await userEvent.fill(pqInput, 'pq-luar-router')
    await expect.element(pqInput).toHaveValue('pq-luar-router')
  })
})
