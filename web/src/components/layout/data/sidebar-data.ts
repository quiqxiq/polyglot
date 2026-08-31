import {
  BellRing,
  ClipboardList,
  Contact,
  FileText,
  Globe,
  LayoutDashboard,
  MessageCircleMore,
  MessagesSquare,
  Monitor,
  Network,
  Package,
  Radio,
  Receipt,
  Repeat,
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

export const sidebarData: SidebarData = {
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
      title: 'General',
      items: [
        {
          title: 'Dashboard',
          url: '/',
          icon: LayoutDashboard,
        },
        {
          title: 'Devices',
          url: '/devices',
          icon: Monitor,
          permission: 'device:read',
        },
        {
          title: 'Hotspot',
          url: '/hotspot',
          icon: Wifi,
          permission: 'hotspot:read',
        },
        {
          title: 'PPPoE / PPP',
          url: '/ppp',
          icon: Network,
          permission: 'ppp:read',
        },
        {
          title: 'Router Logs',
          url: '/logs',
          icon: ScrollText,
          permission: 'log:read',
        },
        {
          title: 'Sales Report',
          url: '/reports',
          icon: FileText,
          permission: 'hotspot:read',
        },
        {
          title: 'POP Remote Probes',
          url: '/probes',
          icon: Radio,
          disabled: true,
          badge: 'Pengembangan',
        },
      ],
    },
    {
      title: 'ISP & Billing',
      items: [
        {
          title: 'Customers',
          url: '/customers',
          icon: Contact,
          permission: 'customer:read',
        },
        {
          title: 'Service Plans',
          url: '/plans',
          icon: Package,
          permission: 'billing:read',
        },
        {
          title: 'Subscriptions',
          url: '/subscriptions',
          icon: Repeat,
          permission: 'billing:read',
        },
        {
          title: 'Registrations',
          url: '/registrations',
          icon: UserPlus,
          permission: 'customer:read',
        },
        {
          title: 'Invoices & Billing',
          url: '/invoices',
          icon: Receipt,
          permission: 'billing:read',
        },
        {
          title: 'Cashbook',
          url: '/cashbook',
          icon: Wallet,
          permission: 'cashbook:read',
        },
        {
          title: 'Customer Portal',
          url: '/portal',
          icon: Globe,
          disabled: true,
          badge: 'Pengembangan',
        },
      ],
    },
    {
      title: 'AI & Omnichannel',
      items: [
        {
          title: 'WhatsApp Devices',
          url: '/whatsapp',
          icon: MessageCircleMore,
          permission: 'whatsapp:read',
        },
        {
          title: 'Chats',
          url: '/chats',
          icon: MessagesSquare,
          permission: 'conversation:read',
        },
        {
          title: 'Skills & Prompts',
          url: '/skills',
          icon: Sparkles,
          permission: 'skill:read',
        },
        {
          title: 'Broadcast & Alerts',
          url: '/broadcast',
          icon: BellRing,
          permission: 'notification:read',
          disabled: true,
          badge: 'Pengembangan',
        },
        {
          title: 'Audit Logs',
          url: '/audit-logs',
          icon: ClipboardList,
          permission: 'audit:read',
          disabled: true,
          badge: 'Pengembangan',
        },
      ],
    },
    {
      title: 'System',
      items: [
        {
          title: 'Users',
          url: '/users',
          icon: Users,
          permission: 'user:read',
        },
        {
          title: 'RBAC',
          url: '/rbac',
          icon: ShieldCheck,
          permission: 'rbac:manage',
        },
        {
          title: 'Settings',
          icon: Settings,
          items: [
            {
              title: 'Profile',
              url: '/settings',
            },
            {
              title: 'Account Security',
              url: '/settings/account',
            },
            {
              title: 'Bot & Anti-Spam',
              url: '/settings/bot',
              permission: 'setting:read',
            },
            {
              title: 'Appearance',
              url: '/settings/appearance',
            },
            {
              title: 'Display',
              url: '/settings/display',
            },
            {
              title: 'Notification Templates',
              url: '/settings/notifications',
              disabled: true,
              badge: 'Pengembangan',
            },
            {
              title: 'Payment Gateway',
              url: '/settings/payment-gateway',
              disabled: true,
              badge: 'Pengembangan',
            },
          ],
        },
      ],
    },
  ],
}
