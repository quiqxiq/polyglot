import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render } from 'vitest-browser-react'
import { userEvent } from 'vitest/browser'
import { Customer } from '@/gen/v1/customer_pb'
import { CustomersMutateDrawer } from './customers-mutate-drawer'

const { createMutateAsync, updateMutateAsync } = vi.hoisted(() => ({
  createMutateAsync: vi.fn(),
  updateMutateAsync: vi.fn(),
}))

vi.mock('../api/use-customer', async (orig) => {
  const actual = await orig<typeof import('../api/use-customer')>()
  return {
    ...actual,
    useCreateCustomerMutation: () => ({
      mutateAsync: createMutateAsync,
      isPending: false,
    }),
    useUpdateCustomerMutation: () => ({
      mutateAsync: updateMutateAsync,
      isPending: false,
    }),
  }
})

const MOCK_CUSTOMER = new Customer({
  id: 'c-001',
  name: 'Budi Santoso',
  phone: '081234567890',
  email: 'budi@example.com',
  address: 'Jl. Merdeka No. 1',
  status: 'ACTIVE',
})

describe('CustomersMutateDrawer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    createMutateAsync.mockResolvedValue({})
    window.localStorage.clear()
  })

  it('renders create title and form fields', async () => {
    const { getByRole, getByText } = await render(
      <CustomersMutateDrawer open onOpenChange={() => {}} currentRow={null} />
    )

    const title = getByRole('heading', { level: 2, name: /Tambah Pelanggan/i })
    await expect.element(title).toBeInTheDocument()

    await expect
      .element(getByRole('textbox', { name: /^Nama$/i }))
      .toBeInTheDocument()
    await expect
      .element(getByRole('textbox', { name: /Nomor HP/i }))
      .toBeInTheDocument()
    await expect
      .element(getByText(/Daftarkan pelanggan baru/i))
      .toBeInTheDocument()
  })

  it('shows validation messages when submitting an empty form', async () => {
    const { getByRole, getByText } = await render(
      <CustomersMutateDrawer open onOpenChange={() => {}} currentRow={null} />
    )

    const saveButton = getByRole('button', { name: /^Save$/i })
    await userEvent.click(saveButton)

    await expect
      .element(getByText(/Nama pelanggan wajib diisi/i))
      .toBeInTheDocument()
    await expect
      .element(getByText(/Nomor HP minimal 8 digit/i))
      .toBeInTheDocument()
    expect(createMutateAsync).not.toHaveBeenCalled()
  })

  it('submits create form and calls the create mutation once', async () => {
    const { getByRole } = await render(
      <CustomersMutateDrawer open onOpenChange={() => {}} currentRow={null} />
    )

    await userEvent.fill(getByRole('textbox', { name: /^Nama$/i }), 'Budi')
    await userEvent.fill(
      getByRole('textbox', { name: /Nomor HP/i }),
      '081234567890'
    )

    const saveButton = getByRole('button', { name: /^Save$/i })
    await userEvent.click(saveButton)

    await vi.waitFor(() => expect(createMutateAsync).toHaveBeenCalledOnce())
    expect(createMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        customer: expect.objectContaining({
          name: 'Budi',
          phone: '081234567890',
        }),
      })
    )
    expect(updateMutateAsync).not.toHaveBeenCalled()
  })

  it('prefills form from currentRow and calls the update mutation', async () => {
    const { getByRole } = await render(
      <CustomersMutateDrawer
        open
        onOpenChange={() => {}}
        currentRow={MOCK_CUSTOMER}
      />
    )

    const title = getByRole('heading', { level: 2, name: /Edit Pelanggan/i })
    await expect.element(title).toBeInTheDocument()

    const nameInput = getByRole('textbox', { name: /^Nama$/i })
    await expect.element(nameInput).toHaveValue(MOCK_CUSTOMER.name)

    await userEvent.fill(nameInput, 'Budi Santoso Update')

    const saveButton = getByRole('button', { name: /^Save$/i })
    await userEvent.click(saveButton)

    await vi.waitFor(() => expect(updateMutateAsync).toHaveBeenCalledOnce())
    expect(updateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        customer: expect.objectContaining({
          id: MOCK_CUSTOMER.id,
          name: 'Budi Santoso Update',
        }),
      })
    )
    expect(createMutateAsync).not.toHaveBeenCalled()
  })

  it('renders coordinate inputs directly without switch and submits coordinates', async () => {
    const { getByRole } = await render(
      <CustomersMutateDrawer open onOpenChange={() => {}} currentRow={null} />
    )

    // Latitude and Longitude inputs are always directly visible without switch button
    await expect
      .element(getByRole('spinbutton', { name: /Latitude/i }))
      .toBeInTheDocument()
    await expect
      .element(getByRole('spinbutton', { name: /Longitude/i }))
      .toBeInTheDocument()

    // Ensure there is no switch button in the form
    expect(getByRole('switch').elements()).toHaveLength(0)

    await userEvent.fill(getByRole('textbox', { name: /^Nama$/i }), 'Siti Aminah')
    await userEvent.fill(getByRole('textbox', { name: /Nomor HP/i }), '089876543210')
    await userEvent.fill(getByRole('spinbutton', { name: /Latitude/i }), '-6.2088')
    await userEvent.fill(getByRole('spinbutton', { name: /Longitude/i }), '106.8456')

    const saveButton = getByRole('button', { name: /^Save$/i })
    await userEvent.click(saveButton)

    await vi.waitFor(() => expect(createMutateAsync).toHaveBeenCalledOnce())
    expect(createMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        customer: expect.objectContaining({
          name: 'Siti Aminah',
          phone: '089876543210',
          latitude: -6.2088,
          longitude: 106.8456,
          hasCoordinates: true,
        }),
      })
    )
  })
})
