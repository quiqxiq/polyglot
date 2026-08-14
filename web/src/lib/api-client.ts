import { AuthService } from '@/gen/v1/auth_connect'
import { BillingService } from '@/gen/v1/billing_connect'
import { BotService } from '@/gen/v1/bot_connect'
import { CustomerService } from '@/gen/v1/customer_connect'
import { DeviceService } from '@/gen/v1/device_connect'
import { HotspotService } from '@/gen/v1/hotspot_connect'
import { KnowledgeService } from '@/gen/v1/knowledge_connect'
import { ProbeService } from '@/gen/v1/probe_connect'
import { RBACService } from '@/gen/v1/rbac_connect'
import { UserService } from '@/gen/v1/users_connect'
import { WhatsAppService } from '@/gen/v1/whatsapp_connect'
import { createPromiseClient, Code, ConnectError } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { useAuthStore } from '@/stores/auth-store'

function isUnauthenticated(err: unknown): boolean {
  return (
    (err instanceof ConnectError && err.code === Code.Unauthenticated) ||
    (err instanceof Error &&
      (err.message.includes('401') ||
        err.message.toLowerCase().includes('unauthorized')))
  )
}

function isAuthProcedure(url: string): boolean {
  // Login/RefreshToken/Logout tidak boleh memicu silent-refresh sendiri —
  // refresh memakai cookie httpOnly, dan kalau cookie invalid kita tidak
  // boleh masuk loop refresh. Perhatikan separator service di URL Connect
  // adalah titik (…/polyglot.v1.AuthService/RefreshToken), bukan slash.
  return url.includes('/polyglot.v1.AuthService/')
}

// Promise tunggal untuk refresh agar banyak request 401 bersamaan tidak
// memicu N refresh paralel (rotasi token di server membuat yang lama mati).
let refreshPromise: Promise<string | null> | null = null

async function refreshAccessToken(): Promise<string | null> {
  if (refreshPromise) return refreshPromise

  refreshPromise = (async () => {
    try {
      const res = await authClient.refreshToken({})
      if (!res.token) return null
      useAuthStore.getState().auth.setAccessToken(res.token)
      return res.token
    } catch {
      return null
    } finally {
      refreshPromise = null
    }
  })()

  return refreshPromise
}

function handleAuthFailure() {
  useAuthStore.getState().auth.reset()
  if (
    typeof window !== 'undefined' &&
    !window.location.pathname.startsWith('/sign-in')
  ) {
    window.location.href = '/sign-in'
  }
}

const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  // Kirim cookie httpOnly (polyglot_refresh) pada tiap request lintas origin.
  credentials: 'include',
  interceptors: [
    (next) => async (req) => {
      const token = useAuthStore.getState().auth.accessToken
      if (token) {
        req.header.set('Authorization', `Bearer ${token}`)
      }
      try {
        return await next(req)
      } catch (err: unknown) {
        if (!isUnauthenticated(err)) throw err
        if (isAuthProcedure(req.url)) throw err

        // Access token expired → silent refresh via cookie, lalu retry sekali.
        const newToken = await refreshAccessToken()
        if (!newToken) {
          handleAuthFailure()
          throw err
        }
        req.header.set('Authorization', `Bearer ${newToken}`)
        return await next(req)
      }
    },
  ],
})

export const authClient = createPromiseClient(AuthService, transport)
export const billingClient = createPromiseClient(BillingService, transport)
export const botClient = createPromiseClient(BotService, transport)
export const customerClient = createPromiseClient(CustomerService, transport)
export const deviceClient = createPromiseClient(DeviceService, transport)
export const hotspotClient = createPromiseClient(HotspotService, transport)
export const knowledgeClient = createPromiseClient(KnowledgeService, transport)
export const probeClient = createPromiseClient(ProbeService, transport)
export const rbacClient = createPromiseClient(RBACService, transport)
export const userClient = createPromiseClient(UserService, transport)
export const whatsappClient = createPromiseClient(WhatsAppService, transport)
