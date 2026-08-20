import {
  BookOpen,
  FileText,
  LayoutDashboard,
  MessageCircleMore,
  MessagesSquare,
  Monitor,
  Network,
  ShieldCheck,
  Users,
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
          title: 'WhatsApp Devices',
          url: '/whatsapp',
          icon: MessageCircleMore,
          permission: 'whatsapp:read',
        },
        {
          title: 'Knowledge',
          url: '/knowledge',
          icon: BookOpen,
          permission: 'knowledge:read',
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
      ],
    },
  ],
}
