import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render } from 'vitest-browser-react'
import { userEvent } from 'vitest/browser'
import { useEffect } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { CustomersProvider, useCustomers } from './customers-provider'
import { CustomersDialogs } from './customers-dialogs'
import { CustomersImportDialog } from './customers-import-dialog'

const { importMutateAsync } = vi.hoisted(() => ({
  importMutateAsync: vi.fn(),
}))

vi.mock('../api/use-customer', async (orig) => {
  const actual = await orig<typeof import('../api/use-customer')>()
  return {
    ...actual,
    useImportFileMutation: () => ({
      mutateAsync: importMutateAsync,
      isPending: false,
    }),
  }
})

vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    info: vi.fn(),
    error: vi.fn(),
  },
}))

function OpenImportOnce() {
  const { setOpen } = useCustomers()
  useEffect(() => {
    setOpen('import')
    // useDialogState's setter toggles — call exactly once, no deps churn.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])
  return null
}

const queryClient = new QueryClient()

function Harness() {
  return (
    <QueryClientProvider client={queryClient}>
      <CustomersProvider>
        <OpenImportOnce />
        <CustomersDialogs />
      </CustomersProvider>
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
    await userEvent.click(getByRole('combobox'))
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
})
