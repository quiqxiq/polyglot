import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render } from 'vitest-browser-react'
import { userEvent } from 'vitest/browser'
import { CustomersImportDialog } from './customers-import-dialog'

const { importMutateAsync, importRouterMutateAsync } = vi.hoisted(() => ({
  importMutateAsync: vi.fn(),
  importRouterMutateAsync: vi.fn(),
}))

vi.mock('../api/use-customer', async (orig) => {
  const actual = await orig<typeof import('../api/use-customer')>()
  return {
    ...actual,
    useImportFileMutation: () => ({
      mutateAsync: importMutateAsync,
      isPending: false,
    }),
    useImportRouterMutation: () => ({
      mutateAsync: importRouterMutateAsync,
      isPending: false,
    }),
  }
})

vi.mock('@tanstack/react-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@tanstack/react-router')>()
  return {
    ...actual,
    Link: ({
      children,
      to,
      className,
      ...rest
    }: {
      children: React.ReactNode
      to: string
      className?: string
    }) => (
      <a href={to} className={className} {...rest}>
        {children}
      </a>
    ),
  }
})

vi.mock('@/features/devices/api/use-devices', () => ({
  useDevicesQuery: () => ({
    data: [{ id: 'dev-1', name: 'ROUTER-TEST', host: '192.168.88.1' }],
    isLoading: false,
  }),
}))

vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    info: vi.fn(),
    error: vi.fn(),
  },
}))

const queryClient = new QueryClient()

function Harness() {
  return (
    <QueryClientProvider client={queryClient}>
      <CustomersImportDialog open onOpenChange={() => {}} />
    </QueryClientProvider>
  )
}

describe('CustomersImportDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    importMutateAsync.mockResolvedValue({
      result: {
        rowsTotal: 1,
        customersCreated: 1,
        customersUpdated: 0,
        subscriptionsCreated: 0,
        plansCreated: 0,
        skipped: [],
      },
    })
  })

  it('shows validation message when submitting without a file', async () => {
    const { getByRole, getByText } = await render(<Harness />)

    await expect
      .element(getByRole('heading', { level: 2, name: /Import Pelanggan/i }))
      .toBeInTheDocument()

    await userEvent.click(getByRole('button', { name: /^Import$/i }))

    await expect
      .element(getByText(/Pilih file terlebih dahulu/i))
      .toBeInTheDocument()
    expect(importMutateAsync).not.toHaveBeenCalled()
  })

  it('submits a csv file with format 0 through the import mutation once', async () => {
    const { getByRole } = await render(<Harness />)

    const input = document.querySelector(
      'input[type="file"]'
    ) as HTMLInputElement
    const csv = new File(['a,b', '1,Budi'], 'c.csv', { type: 'text/csv' })
    await userEvent.upload(input, csv)

    // Explicitly pick the CSV format from the dropdown.
    await userEvent.click(getByRole('combobox', { name: /Format file/i }))
    await userEvent.click(
      getByRole('option', { name: /CSV \(Comma Separated Values\)/i })
    )

    await userEvent.click(getByRole('button', { name: /^Import$/i }))

    await vi.waitFor(() => expect(importMutateAsync).toHaveBeenCalledOnce())
    expect(importMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ format: 0 })
    )
    const req = importMutateAsync.mock.calls[0][0]
    expect(req.payload).toBeInstanceOf(Uint8Array)
    expect(req.payload.length).toBeGreaterThan(0)
  })

  it('renders standalone with explicit open props', async () => {
    const { getByRole } = await render(
      <CustomersImportDialog open onOpenChange={() => {}} />
    )
    await expect
      .element(getByRole('heading', { level: 2, name: /Import Pelanggan/i }))
      .toBeInTheDocument()
  })

  it('handles router live pull preview and execution in Metode A', async () => {
    importRouterMutateAsync.mockResolvedValueOnce({
      result: {
        rowsTotal: 3,
        customersCreated: 3,
        customersUpdated: 0,
        subscriptionsCreated: 3,
        plansCreated: 1,
        skipped: [],
      },
      pppoeDetected: 2,
      hotspotPermanentDetected: 1,
      hotspotIpBindingDetected: 0,
      vouchersSkipped: 25,
      previewRows: [
        '[PPPOE] user1 (paket-10m)',
        '[HOTSPOT-MEMBER] user2 (hotspot-5m)',
      ],
      validationErrors: [],
    })

    const { getByRole, getByText } = await render(<Harness />)

    // Switch to tab "Metode A: Tarik dari Router"
    await userEvent.click(
      getByRole('tab', { name: /Metode A: Tarik dari Router/i })
    )

    // Open router select and pick ROUTER-TEST
    await userEvent.click(getByRole('combobox'))
    await userEvent.click(getByRole('option', { name: /ROUTER-TEST/i }))

    // Click "Tarik & Pratinjau Akun"
    await userEvent.click(
      getByRole('button', { name: /Tarik & Pratinjau Akun/i })
    )

    await vi.waitFor(() =>
      expect(importRouterMutateAsync).toHaveBeenCalledOnce()
    )
    expect(importRouterMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        deviceId: 'dev-1',
        dryRun: true,
      })
    )

    // Verify preview stats are rendered
    await expect
      .element(getByText(/Hasil Deteksi Router:/i))
      .toBeInTheDocument()
    await expect.element(getByText('25')).toBeInTheDocument()
  })
})
