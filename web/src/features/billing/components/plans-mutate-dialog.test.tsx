import { useEffect } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { render } from 'vitest-browser-react'
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
})
