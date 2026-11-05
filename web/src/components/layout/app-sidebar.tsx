import { useMemo } from 'react'
import { useLayout } from '@/context/layout-provider'
import { useI18n } from '@/context/i18n-provider'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
import { getSidebarData } from './data/sidebar-data'
import { NavGroup } from './nav-group'
import { NavUser } from './nav-user'
import { DeviceSwitcher } from './device-switcher'
import { canAllPermissions, canAnyPermission, canPermission } from '@/hooks/use-can'
import { useAuthStore } from '@/stores/auth-store'
import { type NavItem } from './types'

export function AppSidebar() {
  const { t } = useI18n()
  const { collapsible, variant } = useLayout()
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const sidebarData = useMemo(() => getSidebarData(t), [t])

  const checkItemAccess = (item: {
    permission?: string
    anyOf?: string[]
    allOf?: string[]
  }): boolean => {
    if (item.permission && !canPermission(permissions, item.permission)) {
      return false
    }
    if (item.allOf && item.allOf.length > 0 && !canAllPermissions(permissions, item.allOf)) {
      return false
    }
    if (item.anyOf && item.anyOf.length > 0 && !canAnyPermission(permissions, item.anyOf)) {
      return false
    }
    return true
  }

  // Filter item sidebar berdasarkan permission Casbin user dengan cascading denial:
  // 1. Jika parent tidak lolos permission, hide parent dan semua anaknya.
  // 2. Jika collapsible tidak memiliki anak yang dapat diakses, hide collapsible tersebut.
  const filterItems = (items: NavItem[]): NavItem[] => {
    const result: NavItem[] = []
    for (const item of items) {
      if (!checkItemAccess(item)) {
        continue
      }
      if (item.items) {
        const visibleChildren = item.items.filter((child) => checkItemAccess(child))
        if (visibleChildren.length > 0) {
          result.push({
            ...item,
            items: visibleChildren,
          })
        }
      } else {
        result.push(item)
      }
    }
    return result
  }

  const navGroups = sidebarData.navGroups
    .map((group) => ({ ...group, items: filterItems(group.items) }))
    .filter((group) => group.items.length > 0)

  return (
    <Sidebar collapsible={collapsible} variant={variant}>
      <SidebarHeader>
        <DeviceSwitcher />
      </SidebarHeader>
      <SidebarContent>
        {navGroups.map((props) => (
          <NavGroup key={props.title} {...props} />
        ))}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={sidebarData.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
