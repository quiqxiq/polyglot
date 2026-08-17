import {
  BookOpen,
  FileText,
  LayoutDashboard,
  ListTodo,
  MessageCircleMore,
  MessagesSquare,
  Monitor,
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
          title: 'Sales Report',
          url: '/reports',
          icon: FileText,
          permission: 'hotspot:read',
        },
        {
          title: 'Tasks',
          url: '/tasks',
          icon: ListTodo,
        },
        {
          title: 'Chats',
          url: '/chats',
          badge: '3',
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
