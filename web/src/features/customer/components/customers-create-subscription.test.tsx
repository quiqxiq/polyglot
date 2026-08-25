import { beforeEach, describe, expect, it, vi } from 'vitest'
import { render } from 'vitest-browser-react'
import { userEvent } from 'vitest/browser'
import { Customer } from '@/gen/v1/customer_pb'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth-store'
import { CustomersProvider } from './customers-provider'
import { CustomersDialogs } from './customers-dialogs'
import { CustomersRowActions } from './customers-row-actions'

const { plansQuery, customersQuery, createSubAsync } = vi.hoisted(() => ({
  plansQuery: vi.fn(),
  customersQuery: vi.fn(),
  createSubAsync: vi.fn(),
}))

vi.mock('../api/use-customer', async (orig) => {
  const actual = await orig<typeof import('../api/use-customer')>()
  return {
    ...actual,
    useCustomersQuery: (...args: unknown[]) => customersQuery(...args),
  }
})

vi.mock('@/features/billing/api/use-plans', () => ({
  usePlansQuery: (...args: unknown[]) => plansQuery(...args),
}))

vi.mock('@/features/devices/api/use-devices', () => ({
  useDevicesQuery: () => ({ data: [], isPending: false }),
}))

vi.mock('@/features/billing/api/use-billing', async (orig) => {
  const actual = await orig<typeof import('@/features/billing/api/use-billing')>()
  return {
    ...actual,
    useCreateSubscriptionMutation: () => ({
      mutateAsync: createSubAsync,
      isPending: false,
    }),
  }
})

const MOCK_CUSTOMER = new Customer({
  id: 'c-001',
  customerCode: 'PLG-001',
  name: 'Budi Santoso',
  phone: '081234567890',
  status: 'ACTIVE',
})

// Real provider + real dialogs: the row action must flip the shared dialog
// state to 'create-subscription' and mount SubscriptionsCreateDialog.
function renderActions() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <CustomersProvider>
        <CustomersRowActions customer={MOCK_CUSTOMER} />
        <CustomersDialogs />
      </CustomersProvider>
    </QueryClientProvider>
  )
}

function setPermissions(permissions: string[]) {
  useAuthStore.setState({
    auth: {
      user: {
        email: 'admin@example.com',
        role: ['admin'],
        permissions,
      },
      setUser: () => {},
      accessToken: '',
      setAccessToken: () => {},
      resetAccessToken: () => {},
      reset: () => {},
    },
  })
}

describe('inline create subscription from customer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.localStorage.clear()
    // Plans query pending → select shows loading placeholder; not under test.
    plansQuery.mockReturnValue({ data: [], isPending: true })
    customersQuery.mockReturnValue({
      data: [MOCK_CUSTOMER],
      isPending: false,
    })
  })

  it('opens the subscription dialog with the customer locked', async () => {
    setPermissions(['customer:manage', 'billing:manage'])
    const { getByRole } = await renderActions()

    await userEvent.click(getByRole('button', { name: /open menu/i }))
    await userEvent.click(getByRole('menuitem', { name: /Buat Langganan/i }))

    await expect
      .element(getByRole('heading', { name: /Tambah Langganan/i }))
      .toBeInTheDocument()

    // Customer select is disabled and prefilled with this customer's name.
    const customerSelect = getByRole('combobox').elements().find((el) =>
      (el as HTMLElement).textContent?.includes('Budi Santoso')
    ) as HTMLElement | undefined
    expect(customerSelect).toBeDefined()
    expect(customerSelect).toHaveAttribute('data-disabled')
  })

  it('hides the menu item without billing:manage permission', async () => {
    setPermissions(['customer:manage'])
    const { getByRole } = await renderActions()

    await userEvent.click(getByRole('button', { name: /open menu/i }))

    await expect
      .element(getByRole('menuitem', { name: /^Edit$/i }))
      .toBeInTheDocument()
    const subItems = getByRole('menuitem', {
      name: /Buat Langganan/i,
    }).query() as (HTMLElement | SVGElement)[] | null
    expect(subItems?.length ?? 0).toBe(0)
  })
})
