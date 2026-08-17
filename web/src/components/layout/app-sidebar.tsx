import { useLayout } from '@/context/layout-provider'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
import { sidebarData } from './data/sidebar-data'
import { NavGroup } from './nav-group'
import { NavUser } from './nav-user'
import { DeviceSwitcher } from './device-switcher'
import { canPermission } from '@/hooks/use-can'
import { useAuthStore } from '@/stores/auth-store'
import { type NavItem } from './types'

export function AppSidebar() {
  const { collapsible, variant } = useLayout()
  const permissions = useAuthStore((s) => s.auth.user?.permissions)

  // Filter item sidebar berdasarkan permission Casbin user — item tanpa
  // permission tampil untuk semua user yang sudah login.
  const filterItems = (items: NavItem[]): NavItem[] =>
    items
      .filter((item) => !item.permission || canPermission(permissions, item.permission))
      .map((item) =>
        item.items
          ? {
              ...item,
              items: item.items.filter(
                (child) => !child.permission || canPermission(permissions, child.permission),
              ),
            }
          : item,
      )

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
