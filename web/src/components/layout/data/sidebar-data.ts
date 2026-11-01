import {
  BarChart3,
  BellRing,
  ClipboardList,
  Contact,
  FileText,
  Globe,
  HandCoins,
  LayoutDashboard,
  MessageCircleMore,
  MessagesSquare,
  Monitor,
  Network,
  Package,
  Radio,
  Receipt,
  Repeat,
  Router,
  ScrollText,
  Settings,
  ShieldCheck,
  Sparkles,
  UserPlus,
  Users,
  Wallet,
  Wifi,
} from 'lucide-react'
import { type SidebarData } from '../types'

export const getSidebarData = (t: (key: string) => string): SidebarData => ({
  user: {
    name: 'Operator',
    email: 'operator@polyglot.local',
    avatar: '/avatars/shadcn.jpg',
  },
  teams: [
    {
      name: 'Polyglot',
      logo: LayoutDashboard,
      plan: 'NetOps Engine',
    },
  ],
  navGroups: [
    {
      title: t('nav.general'),
      items: [
        {
          title: t('nav.dashboard'),
          url: '/',
          icon: LayoutDashboard,
        },
        {
          title: t('nav.devices'),
          url: '/devices',
          icon: Monitor,
          permission: 'device:read',
        },
        {
          title: t('nav.hotspot'),
          url: '/hotspot',
          icon: Wifi,
          permission: 'hotspot:read',
        },
        {
          title: t('nav.ppp'),
          url: '/ppp',
          icon: Network,
          permission: 'ppp:read',
        },
        {
          title: t('nav.olt'),
          url: '/olt',
          icon: Router,
          permission: 'olt:read',
        },
        {
          title: t('nav.logs'),
          url: '/logs',
          icon: ScrollText,
          permission: 'log:read',
        },
        {
          title: t('nav.finance_report'),
          url: '/reports/finance',
          icon: BarChart3,
          permission: 'report:read',
        },
        {
          title: t('nav.reports'),
          url: '/reports',
          icon: FileText,
          permission: 'hotspot:read',
        },
        {
          title: t('nav.probes'),
          url: '/probes',
          icon: Radio,
          disabled: true,
          badge: t('common.in_development'),
        },
      ],
    },
    {
      title: t('nav.isp_billing'),
      items: [
        {
          title: t('nav.customers'),
          url: '/customers',
          icon: Contact,
          permission: 'customer:read',
        },
        {
          title: t('nav.plans'),
          url: '/plans',
          icon: Package,
          permission: 'billing:read',
        },
        {
          title: t('nav.subscriptions'),
          url: '/subscriptions',
          icon: Repeat,
          permission: 'billing:read',
        },
        {
          title: t('nav.registrations'),
          url: '/registrations',
          icon: UserPlus,
          permission: 'customer:read',
        },
        {
          title: t('nav.invoices'),
          url: '/invoices',
          icon: Receipt,
          permission: 'billing:read',
        },
        {
          title: t('nav.payments'),
          url: '/payments',
          icon: Wallet,
          permission: 'billing:read',
        },
        {
          title: t('nav.cashbook'),
          url: '/cashbook',
          icon: Wallet,
          permission: 'cashbook:read',
        },
        {
          title: t('nav.debts'),
          url: '/debts',
          icon: HandCoins,
          permission: 'debt:read',
        },
        {
          title: t('nav.portal'),
          url: '/portal',
          icon: Globe,
          disabled: true,
          badge: t('common.in_development'),
        },
      ],
    },
    {
      title: t('nav.ai_omnichannel'),
      items: [
        {
          title: t('nav.whatsapp'),
          url: '/whatsapp',
          icon: MessageCircleMore,
          permission: 'whatsapp:read',
        },
        {
          title: t('nav.chats'),
          url: '/chats',
          icon: MessagesSquare,
          permission: 'conversation:read',
        },
        {
          title: t('nav.skills'),
          url: '/skills',
          icon: Sparkles,
          permission: 'skill:read',
        },
        {
          title: t('nav.broadcast'),
          url: '/broadcast',
          icon: BellRing,
          permission: 'notification:read',
          disabled: true,
          badge: t('common.in_development'),
        },
        {
          title: t('nav.audit_logs'),
          url: '/audit-logs',
          icon: ClipboardList,
          permission: 'audit:read',
          disabled: true,
          badge: t('common.in_development'),
        },
      ],
    },
    {
      title: t('nav.system'),
      items: [
        {
          title: t('nav.users'),
          url: '/users',
          icon: Users,
          permission: 'user:read',
        },
        {
          title: t('nav.rbac'),
          url: '/rbac',
          icon: ShieldCheck,
          permission: 'rbac:manage',
        },
        {
          title: t('nav.settings'),
          icon: Settings,
          items: [
            {
              title: t('nav.profile'),
              url: '/settings',
            },
            {
              title: t('nav.account_security'),
              url: '/settings/account',
            },
            {
              title: t('nav.bot_anti_spam'),
              url: '/settings/bot',
              permission: 'setting:read',
            },
            {
              title: t('nav.appearance'),
              url: '/settings/appearance',
            },
            {
              title: t('nav.display'),
              url: '/settings/display',
            },
            {
              title: t('nav.notifications'),
              url: '/settings/notifications',
              disabled: true,
              badge: t('common.in_development'),
            },
            {
              title: t('nav.payment_gateway'),
              url: '/settings/gateway',
              permission: 'setting:read',
            },
            {
              title: t('nav.billing_settings'),
              url: '/settings/billing',
              permission: 'setting:read',
            },
            {
              title: t('nav.payment_methods'),
              url: '/settings/payment-methods',
              permission: 'billing:read',
            },
          ],
        },
      ],
    },
  ],
})

export const sidebarData = getSidebarData((key: string) => {
  const parts = key.split('.')
  return parts[parts.length - 1] || key
})
