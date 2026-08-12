import { createConnectTransport } from '@connectrpc/connect-web'
import { createPromiseClient, Code, ConnectError } from '@connectrpc/connect'
import { AuthService } from '@/gen/v1/auth_connect'
import { BillingService } from '@/gen/v1/billing_connect'
import { BotService } from '@/gen/v1/bot_connect'
import { CustomerService } from '@/gen/v1/customer_connect'
import { DeviceService } from '@/gen/v1/device_connect'
import { HotspotService } from '@/gen/v1/hotspot_connect'
import { KnowledgeService } from '@/gen/v1/knowledge_connect'
import { ProbeService } from '@/gen/v1/probe_connect'
import { RBACService } from '@/gen/v1/rbac_connect'
import { WhatsAppService } from '@/gen/v1/whatsapp_connect'
import { useAuthStore } from '@/stores/auth-store'

const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  interceptors: [
    (next) => async (req) => {
      const token = useAuthStore.getState().auth.accessToken
      if (token) {
        req.header.set('Authorization', `Bearer ${token}`)
      }
      try {
        return await next(req)
      } catch (err: unknown) {
        if (
          (err instanceof ConnectError && err.code === Code.Unauthenticated) ||
          (err instanceof Error && (err.message.includes('401') || err.message.toLowerCase().includes('unauthorized')))
        ) {
          useAuthStore.getState().auth.reset()
          if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/sign-in')) {
            window.location.href = '/sign-in'
          }
        }
        throw err
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
export const whatsappClient = createPromiseClient(WhatsAppService, transport)
