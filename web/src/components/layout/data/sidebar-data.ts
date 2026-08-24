import {
  Contact,
  FileText,
  LayoutDashboard,
  MessageCircleMore,
  MessagesSquare,
  Monitor,
  Network,
  Package,
  ScrollText,
  ShieldCheck,
  Sparkles,
  Users,
  Wifi,
  Settings,
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
          title: 'WhatsApp Devices',
          url: '/whatsapp',
          icon: MessageCircleMore,
          permission: 'whatsapp:read',
        },
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
            // {
            //   title: 'Notifications',
            //   url: '/settings/notifications',
            // },
            {
              title: 'Display',
              url: '/settings/display',
            },
          ],
        },
      ],
    },
  ],
}
