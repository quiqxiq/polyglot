import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { render } from 'vitest-browser-react'
import { userEvent } from 'vitest/browser'
import { useAuthStore } from '@/stores/auth-store'
import { CustomersPrimaryButtons } from './customers-primary-buttons'

const { exportMutateAsync, toastSuccessSpy, toastErrorSpy } = vi.hoisted(() => ({
  exportMutateAsync: vi.fn(),
  toastSuccessSpy: vi.fn(),
  toastErrorSpy: vi.fn(),
}))

vi.mock('../api/use-customer', async (orig) => {
  const actual = await orig<typeof import('../api/use-customer')>()
  return {
    ...actual,
    useExportCustomersMutation: () => ({
      mutateAsync: exportMutateAsync,
      isPending: false,
    }),
  }
})

// Pure factory mock: do NOT import the real customers-provider module — it
// transitively imports dialogs/schemas owned by other workstreams and would
// couple this test to their compile state.
vi.mock('./customers-provider', () => ({
  useCustomers: () => ({ setOpen: vi.fn() }),
}))

vi.mock('sonner', () => ({
  toast: {
    success: toastSuccessSpy,
    error: toastErrorSpy,
  },
}))

let clickSpy: ReturnType<typeof vi.spyOn>
let createObjectURLSpy: ReturnType<typeof vi.spyOn>
let revokeObjectURLSpy: ReturnType<typeof vi.spyOn>

function renderButtons() {
  return render(<CustomersPrimaryButtons />)
}

describe('CustomersPrimaryButtons export', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.localStorage.clear()
    // Re-arm spies per test: afterEach restores the originals, so spying must
    // happen here rather than at module level.
    clickSpy = vi
      .spyOn(HTMLAnchorElement.prototype, 'click')
      .mockImplementation(() => {})
    createObjectURLSpy = vi
      .spyOn(URL, 'createObjectURL')
      .mockReturnValue('blob:mock-url')
    revokeObjectURLSpy = vi
      .spyOn(URL, 'revokeObjectURL')
      .mockImplementation(() => {})
    useAuthStore.setState({
      auth: {
        user: {
          email: 'admin@example.com',
          role: ['admin'],
          permissions: ['customer:manage'],
        },
        setUser: () => {},
        accessToken: '',
        setAccessToken: () => {},
        resetAccessToken: () => {},
        reset: () => {},
      },
    })
    createObjectURLSpy.mockReturnValue('blob:mock-url')
  })

  afterEach(() => {
    clickSpy.mockRestore()
    createObjectURLSpy.mockRestore()
    revokeObjectURLSpy.mockRestore()
  })

  it('exports CSV via browser download when the CSV item is clicked', async () => {
    exportMutateAsync.mockResolvedValue({
      payload: new Uint8Array([1, 2]),
      contentType: 'text/csv',
      filename: 'customers.csv',
    })
    const { getByRole } = await renderButtons()

    await userEvent.click(getByRole('button', { name: /Export/i }))
    await userEvent.click(getByRole('menuitem', { name: 'CSV' }))

    await vi.waitFor(() => expect(exportMutateAsync).toHaveBeenCalledOnce())
    expect(exportMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ format: 0 })
    )
    await vi.waitFor(() => expect(createObjectURLSpy).toHaveBeenCalledTimes(1))
    const blob = createObjectURLSpy.mock.calls[0][0] as Blob
    expect(blob.type).toBe('text/csv')
    await vi.waitFor(() => expect(clickSpy).toHaveBeenCalledTimes(1))
    expect(revokeObjectURLSpy).toHaveBeenCalledWith('blob:mock-url')
  })

  it('requests Excel format when the XLSX item is clicked', async () => {
    exportMutateAsync.mockResolvedValue({
      payload: new Uint8Array([3]),
      contentType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      filename: 'customers.xlsx',
    })
    const { getByRole } = await renderButtons()

    await userEvent.click(getByRole('button', { name: /Export/i }))
    await userEvent.click(getByRole('menuitem', { name: 'Excel (XLSX)' }))

    await vi.waitFor(() => expect(exportMutateAsync).toHaveBeenCalledOnce())
    expect(exportMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ format: 1 })
    )
    // Blob URL creation happens after the mutation promise resolves.
    await vi.waitFor(() => expect(createObjectURLSpy).toHaveBeenCalledTimes(1))
  })

  it('shows an error toast when the export mutation fails', async () => {
    exportMutateAsync.mockRejectedValue(new Error('boom'))
    const { getByRole } = await renderButtons()

    await userEvent.click(getByRole('button', { name: /Export/i }))
    await userEvent.click(getByRole('menuitem', { name: 'CSV' }))

    await vi.waitFor(() =>
      expect(toastErrorSpy).toHaveBeenCalledWith(
        expect.stringMatching(/boom|Export gagal/)
      )
    )
    expect(toastSuccessSpy).not.toHaveBeenCalled()
    expect(createObjectURLSpy).not.toHaveBeenCalled()
  })

  it('renders nothing without customer:manage permission', async () => {
    useAuthStore.setState({
      auth: {
        user: {
          email: 'viewer@example.com',
          role: ['viewer'],
          permissions: ['customer:read'],
        },
        setUser: () => {},
        accessToken: '',
        setAccessToken: () => {},
        resetAccessToken: () => {},
        reset: () => {},
      },
    })
    const { container } = await renderButtons()

    expect(container.querySelector('button')).toBeNull()
    expect(container.textContent).toBe('')
  })
})
