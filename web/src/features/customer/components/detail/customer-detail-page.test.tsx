import { beforeEach, describe, expect, it, vi } from 'vitest'
import { render } from 'vitest-browser-react'
import { userEvent } from 'vitest/browser'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Customer } from '@/gen/v1/customer_pb'
import { Subscription } from '@/gen/v1/subscription_pb'
import { Invoice } from '@/gen/v1/billing_pb'
import { useAuthStore } from '@/stores/auth-store'
import { CustomerDetailPage } from './customer-detail-page'

const { customerQuery, subscriptionsQuery, invoicesQuery } = vi.hoisted(() => ({
  customerQuery: vi.fn(),
  subscriptionsQuery: vi.fn(),
  invoicesQuery: vi.fn(),
}))

vi.mock('../../api/use-customer', async (orig) => {
  const actual = await orig<typeof import('../../api/use-customer')>()
  return {
    ...actual,
    useCustomerQuery: (...args: unknown[]) => customerQuery(...args),
    // Dipakai SubscriptionsCreateDialog (tertutup) — hindari call nyata.
    useCustomersQuery: () => ({ data: [], isPending: false }),
  }
})

vi.mock('@/features/billing/api/use-billing', async (orig) => {
  const actual = await orig<typeof import('@/features/billing/api/use-billing')>()
  return {
    ...actual,
    useSubscriptionsQuery: (...args: unknown[]) => subscriptionsQuery(...args),
    useInvoicesQuery: (...args: unknown[]) => invoicesQuery(...args),
  }
})

vi.mock('@/features/billing/api/use-plans', () => ({
  usePlansQuery: () => ({ data: [], isPending: false }),
}))

vi.mock('@/features/devices/api/use-devices', () => ({
  useDevicesQuery: () => ({ data: [], isPending: false }),
}))

vi.mock('@tanstack/react-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@tanstack/react-router')>()
  return {
    ...actual,
    Link: ({
      children,
      to,
      ...rest
    }: {
      children: React.ReactNode
      to: string
      className?: string
    }) => (
      <a href={to} {...rest}>
        {children}
      </a>
    ),
    useNavigate: () => vi.fn(),
    useParams: () => ({ customerId: 'c-001' }),
  }
})

vi.mock('@/components/layout/header', () => ({
  Header: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
}))
vi.mock('@/components/search', () => ({ Search: () => null }))
vi.mock('@/components/theme-switch', () => ({ ThemeSwitch: () => null }))
vi.mock('@/components/config-drawer', () => ({ ConfigDrawer: () => null }))
vi.mock('@/components/profile-dropdown', () => ({ ProfileDropdown: () => null }))

const MOCK_CUSTOMER = new Customer({
  id: 'c-001',
  customerCode: 'PLG-001',
  name: 'Budi Santoso',
  phone: '081234567890',
  email: 'budi@example.com',
  address: 'Jl. Mawar No. 1',
  status: 'ACTIVE',
  portalAccessCode: 'PRT-777',
  registeredAtUnix: BigInt(Date.UTC(2025, 0, 15, 12) / 1000),
  activeSubscriptionsCount: 1,
  unpaidInvoicesCount: 1,
})

const MOCK_SUBSCRIPTION = new Subscription({
  id: 'sub-1',
  customerId: 'c-001',
  planId: 'plan-1',
  planName: 'Home 10M',
  serviceType: 'PPPOE',
  remoteUsername: 'budi_home',
  status: 'ACTIVE',
  provisionStatus: 'OK',
  deviceName: 'RB-01',
  rateLimit: '10M/2M',
  billingDay: 5,
  billingCycle: 'MONTHLY',
  planPrice: 150000,
})

const MOCK_INVOICE = new Invoice({
  id: 'inv-1',
  customerId: 'c-001',
  invoiceNumber: 'INV-2026-001',
  period: '2026-08',
  total: 165000,
  paidAmount: 0,
  status: 'UNPAID',
  dueDateUnix: BigInt(Date.UTC(2026, 7, 10, 12) / 1000),
})

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

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <CustomerDetailPage />
    </QueryClientProvider>
  )
}

describe('CustomerDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setPermissions(['customer:read', 'customer:manage', 'billing:manage'])
    customerQuery.mockReturnValue({
      data: MOCK_CUSTOMER,
      isLoading: false,
      isError: false,
    })
    subscriptionsQuery.mockReturnValue({
      data: [MOCK_SUBSCRIPTION],
      isLoading: false,
    })
    invoicesQuery.mockReturnValue({ data: [MOCK_INVOICE], isLoading: false })
  })

  it('menampilkan identitas pelanggan dan kartu ringkasan', async () => {
    const { getByRole, getByText } = await renderPage()

    await expect
      .element(getByRole('heading', { name: 'Budi Santoso' }))
      .toBeInTheDocument()
    await expect.element(getByText('PLG-001').first()).toBeInTheDocument()
    await expect.element(getByText('Aktif', { exact: true })).toBeInTheDocument()
    await expect.element(getByText('Layanan Aktif')).toBeInTheDocument()
    await expect.element(getByText('Tagihan Tertunggak')).toBeInTheDocument()
    await expect.element(getByText(/165\.000/).first()).toBeInTheDocument()
  })

  it('tab Layanan menampilkan langganan beserta paket dan username', async () => {
    const { getByRole, getByText } = await renderPage()

    await userEvent.click(getByRole('tab', { name: /Layanan/i }))

    await expect.element(getByText('Home 10M')).toBeInTheDocument()
    await expect.element(getByText('budi_home')).toBeInTheDocument()
    await expect
      .element(getByText('PPPoE', { exact: true }))
      .toBeInTheDocument()
    await expect.element(getByText('10M/2M')).toBeInTheDocument()
  })

  it('tab Tagihan menampilkan faktur dengan status belum bayar', async () => {
    const { getByRole, getByText } = await renderPage()

    await userEvent.click(getByRole('tab', { name: /Tagihan/i }))

    await expect.element(getByText('INV-2026-001').first()).toBeInTheDocument()
    await expect
      .element(getByText('Belum Bayar').first())
      .toBeInTheDocument()
  })

  it('menyembunyikan aksi kelola tanpa permission', async () => {
    setPermissions(['customer:read'])
    const { getByRole } = await renderPage()

    const addButtons = getByRole('button', {
      name: /Tambah Layanan/i,
    }).query() as (HTMLElement | SVGElement)[] | null
    expect(addButtons?.length ?? 0).toBe(0)

    const editButtons = getByRole('button', { name: /^Edit$/i }).query() as
      | (HTMLElement | SVGElement)[]
      | null
    expect(editButtons?.length ?? 0).toBe(0)
  })

  it('menampilkan empty state saat pelanggan tidak ditemukan', async () => {
    customerQuery.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    })
    const { getByText } = await renderPage()

    await expect
      .element(getByText('Pelanggan Tidak Ditemukan'))
      .toBeInTheDocument()
  })
})
