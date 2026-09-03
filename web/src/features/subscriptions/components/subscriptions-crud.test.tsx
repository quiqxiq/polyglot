import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render } from 'vitest-browser-react'
import { Subscription } from '@/gen/v1/subscription_pb'
import { useAuthStore } from '@/stores/auth-store'
import { SubscriptionsProvider } from './subscriptions-provider'
import { SubscriptionsDialogs } from './subscriptions-dialogs'
import { SubscriptionsPrimaryButtons } from './subscriptions-primary-buttons'
import { SubscriptionsEditDialog } from './subscriptions-edit-dialog'

const mockCustomers = [
  { id: 'c-001', name: 'Budi Santoso' },
  { id: 'c-002', name: 'Siti Rahayu' },
]
const mockPlans = [
  { id: 'p-001', name: 'Home 20M', price: 150000 },
  { id: 'p-002', name: 'Hotspot Bulanan', price: 100000 },
]

vi.mock('@/features/customer/api/use-customer', () => ({
  useCustomersQuery: () => ({ data: mockCustomers, isPending: false }),
}))

vi.mock('@/features/billing/api/use-plans', () => ({
  usePlansQuery: () => ({ data: mockPlans, isPending: false }),
}))

vi.mock('@/features/devices/api/use-devices', () => ({
  useDevicesQuery: () => ({ data: [], isPending: false }),
}))

const createSpy = vi.fn()
const updateSpy = vi.fn()
const deleteSpy = vi.fn()

vi.mock('@/features/billing/api/use-billing', async (importOriginal) => {
  const actual = await importOriginal<
    typeof import('@/features/billing/api/use-billing')
  >()
  return {
    ...actual,
    useSubscriptionsQuery: () => ({ data: [], isLoading: false }),
    useCreateSubscriptionMutation: () => ({
      mutateAsync: createSpy,
      isPending: false,
    }),
    useUpdateSubscriptionMutation: () => ({
      mutateAsync: updateSpy,
      isPending: false,
    }),
    useDeleteSubscriptionMutation: () => ({
      mutateAsync: deleteSpy,
      isPending: false,
    }),
  }
})

beforeEach(() => {
  vi.clearAllMocks()
  useAuthStore.setState({
    auth: {
      user: {
        email: 'admin@example.com',
        role: ['admin'],
        permissions: ['billing:manage'],
      },
      setUser: () => {},
      accessToken: '',
      setAccessToken: () => {},
      resetAccessToken: () => {},
      reset: () => {},
    },
  })
})

describe('Subscriptions CRUD dialogs', () => {
  it('opens the create dialog from primary button and submits with selected customer+plan', async () => {
    const screen = await render(
      <SubscriptionsProvider>
        <SubscriptionsPrimaryButtons />
        <SubscriptionsDialogs />
      </SubscriptionsProvider>
    )

    await screen.getByRole('button', { name: /Tambah Langganan/ }).click()

    const dialog = screen.getByRole('dialog')
    await expect.element(dialog).toBeInTheDocument()

    // Pilih pelanggan
    await dialog.getByText('Pilih pelanggan').click()
    await expect
      .element(screen.getByRole('option', { name: 'Budi Santoso' }))
      .toBeInTheDocument()
    await screen.getByRole('option', { name: 'Budi Santoso' }).click()

    // Pilih paket
    await dialog.getByText('Pilih paket').click()
    await expect
      .element(screen.getByRole('option', { name: /Home 20M/ }))
      .toBeInTheDocument()
    await screen.getByRole('option', { name: /Home 20M/ }).click()

    await screen.getByRole('button', { name: 'Buat Langganan' }).click()

    await vi.waitFor(() => expect(createSpy).toHaveBeenCalledTimes(1))
    const req = createSpy.mock.calls[0][0] as {
      customerId: string
      planId: string
    }
    expect(req.customerId).toBe('c-001')
    expect(req.planId).toBe('p-001')
  })

  it('prefills username from currentRow in the edit dialog', async () => {
    const row = new Subscription({
      id: 's-001',
      customerId: 'c-001',
      planId: 'p-001',
      serviceType: 'PPPOE',
      remoteUsername: 'budi20m',
      customPrice: 155000,
      status: 'ACTIVE',
      billingCycle: 'MONTHLY',
      billingDay: 5,
    })

    const screen = await render(
      <SubscriptionsEditDialog open onOpenChange={() => {}} currentRow={row} />
    )

    // useEffect prefill berjalan setelah mount; tunggu input terisi.
    await vi.waitFor(() => {
      const input = document.querySelector<HTMLInputElement>(
        '#subscriptions-edit-form input'
      )
      expect(input?.value).toBe('budi20m')
    })
    void screen
  })
})
