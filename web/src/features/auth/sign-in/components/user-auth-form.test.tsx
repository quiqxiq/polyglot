import { beforeEach, describe, expect, it, vi } from 'vitest'
import { render, type RenderResult } from 'vitest-browser-react'
import { type Locator, userEvent } from 'vitest/browser'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { UserAuthForm } from './user-auth-form'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
    mutations: { retry: false },
  },
})

function renderWithQuery(ui: React.ReactElement) {
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>
  )
}

const FORM_MESSAGES = {
  usernameEmpty: 'Please enter your username.',
  passwordEmpty: 'Please enter your password.',
  passwordShort: 'Password must be at least 6 characters long.',
} as const

const mocks = vi.hoisted(() => {
  const setUser = vi.fn()
  const setAccessToken = vi.fn()
  const navigate = vi.fn()
  const mockLogin = vi.fn().mockResolvedValue({
    token: 'mock-access-token',
    expiresAtUnix: 1234567890,
    user: {
      id: '1',
      username: 'admin',
      fullName: 'Administrator',
      email: 'admin@example.com',
      role: 'admin',
      roles: ['admin'],
      permissions: ['*'],
    },
  })
  return {
    setUser,
    setAccessToken,
    navigate,
    mockLogin,
  }
})

vi.mock('@/lib/api-client', () => ({
  authClient: {
    login: (...args: unknown[]) => mocks.mockLogin(...args),
  },
}))

vi.mock('@/stores/auth-store', () => {
  const state = {
    auth: {
      setUser: mocks.setUser,
      setAccessToken: mocks.setAccessToken,
    },
  }
  return {
    useAuthStore: (selector?: (s: typeof state) => unknown) =>
      selector ? selector(state) : state,
  }
})

vi.mock('@tanstack/react-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@tanstack/react-router')>()
  return {
    ...actual,
    useNavigate: () => mocks.navigate,
    Link: ({
      children,
      to,
      className,
      ...rest
    }: {
      children?: React.ReactNode
      to: string
      className?: string
    }) => (
      <a href={to} className={className} {...rest}>
        {children}
      </a>
    ),
  }
})

vi.mock('@/lib/utils', async (orig) => ({
  ...(await orig()),
  sleep: vi.fn(() => Promise.resolve()),
}))

describe('UserAuthForm', () => {
  describe('Rendering without redirectTo', () => {
    let screen: RenderResult
    let usernameInput: Locator
    let passwordInput: Locator
    let signInButton: Locator
    let forgotPasswordLink: Locator

    beforeEach(async () => {
      vi.clearAllMocks()
      screen = await renderWithQuery(<UserAuthForm />)
      usernameInput = screen.getByRole('textbox', { name: /^Username$/i })
      passwordInput = screen.getByLabelText(/^Password$/i)
      signInButton = screen.getByRole('button', { name: /^Sign in$/i })
      forgotPasswordLink = screen.getByText(/^Forgot password\?$/i)
    })

    it('renders fields, submit button, and forgot password link', async () => {
      await expect.element(usernameInput).toBeInTheDocument()
      await expect.element(passwordInput).toBeInTheDocument()
      await expect.element(signInButton).toBeInTheDocument()
      await expect.element(forgotPasswordLink).toBeInTheDocument()
    })

    it('shows validation messages when submitting empty form', async () => {
      await userEvent.click(signInButton)

      await expect
        .element(screen.getByText(FORM_MESSAGES.usernameEmpty))
        .toBeInTheDocument()
      await expect
        .element(screen.getByText(FORM_MESSAGES.passwordEmpty))
        .toBeInTheDocument()
    })

    it('authenticates and navigates to default route on success', async () => {
      await userEvent.fill(usernameInput, 'admin')
      await userEvent.fill(passwordInput, 'admin12345')

      await userEvent.click(signInButton)

      await vi.waitFor(() => expect(mocks.setUser).toHaveBeenCalledOnce())
      expect(mocks.setUser).toHaveBeenCalledWith(
        expect.objectContaining({
          email: 'admin@example.com',
          accountNo: '1',
          username: 'admin',
          role: expect.any(Array),
          exp: expect.any(Number),
        })
      )
      expect(mocks.setAccessToken).toHaveBeenCalledOnce()
      expect(mocks.setAccessToken).toHaveBeenCalledWith('mock-access-token')

      await vi.waitFor(() =>
        expect(mocks.navigate).toHaveBeenCalledWith({ to: '/', replace: true })
      )
    })
  })

  it('navigates to redirectTo when provided', async () => {
    vi.clearAllMocks()

    const { getByRole, getByLabelText } = await renderWithQuery(
      <UserAuthForm redirectTo='/settings' />
    )

    await userEvent.fill(getByRole('textbox', { name: /Username/i }), 'admin')
    await userEvent.fill(getByLabelText('Password'), 'admin12345')

    await userEvent.click(getByRole('button', { name: /Sign in/i }))

    await vi.waitFor(() => expect(mocks.setUser).toHaveBeenCalledOnce())
    expect(mocks.setAccessToken).toHaveBeenCalledOnce()

    await vi.waitFor(() =>
      expect(mocks.navigate).toHaveBeenCalledWith({
        to: '/settings',
        replace: true,
      })
    )
  })
})
