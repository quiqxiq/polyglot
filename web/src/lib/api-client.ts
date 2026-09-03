import { AuthService } from '@/gen/v1/auth_connect'
import { BillingService } from '@/gen/v1/billing_connect'
import { BotService } from '@/gen/v1/bot_connect'
import { CashbookService } from '@/gen/v1/cashbook_connect'
import { CustomerService } from '@/gen/v1/customer_connect'
import { DeviceService } from '@/gen/v1/device_connect'
import { HotspotService } from '@/gen/v1/hotspot_connect'
import { IspAdminService } from '@/gen/v1/ispadmin_connect'
import { LLMConfigService } from '@/gen/v1/llm_connect'
import { NetworkMonitorService } from '@/gen/v1/network_monitor_connect'
import { NetworkService } from '@/gen/v1/network_connect'
import { NotificationService } from '@/gen/v1/notification_connect'
import { PlanService } from '@/gen/v1/plan_connect'
import { PortalService } from '@/gen/v1/portal_connect'
import { PPPService } from '@/gen/v1/ppp_connect'
import { ProbeService } from '@/gen/v1/probe_connect'
import { RBACService } from '@/gen/v1/rbac_connect'
import { RegistrationService } from '@/gen/v1/registration_connect'
import { ReportService } from '@/gen/v1/report_connect'
import { SettingService } from '@/gen/v1/settings_connect'
import { SkillService } from '@/gen/v1/skill_connect'
import { SubscriptionService } from '@/gen/v1/subscription_connect'
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
  return url.includes('/polyglot.v1.AuthService/')
}

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
export const cashbookClient = createPromiseClient(CashbookService, transport)
export const customerClient = createPromiseClient(CustomerService, transport)
export const deviceClient = createPromiseClient(DeviceService, transport)
export const hotspotClient = createPromiseClient(HotspotService, transport)
export const ispAdminClient = createPromiseClient(IspAdminService, transport)
export const llmConfigClient = createPromiseClient(LLMConfigService, transport)
export const networkClient = createPromiseClient(NetworkService, transport)
export const networkMonitorClient = createPromiseClient(NetworkMonitorService, transport)
export const notificationClient = createPromiseClient(NotificationService, transport)
export const planClient = createPromiseClient(PlanService, transport)
export const portalClient = createPromiseClient(PortalService, transport)
export const pppClient = createPromiseClient(PPPService, transport)
export const probeClient = createPromiseClient(ProbeService, transport)
export const rbacClient = createPromiseClient(RBACService, transport)
export const registrationClient = createPromiseClient(RegistrationService, transport)
export const reportClient = createPromiseClient(ReportService, transport)
export const settingClient = createPromiseClient(SettingService, transport)
export const skillClient = createPromiseClient(SkillService, transport)
export const subscriptionClient = createPromiseClient(SubscriptionService, transport)
export const userClient = createPromiseClient(UserService, transport)
export const whatsappClient = createPromiseClient(WhatsAppService, transport)
