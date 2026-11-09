import { useMemo } from 'react'
import { Outlet } from '@tanstack/react-router'
import {
  Monitor,
  Palette,
  ReceiptText,
  Wrench,
  UserCog,
  CreditCard,
  RadioTower,
  WalletCards,
} from 'lucide-react'
import { useI18n } from '@/context/i18n-provider'
import { useCan } from '@/hooks/use-can'
import { Separator } from '@/components/ui/separator'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { SidebarNav } from './components/sidebar-nav'

export function Settings() {
  const { t } = useI18n()
  const canReadOnt = useCan('ont:read')

  const sidebarNavItems = useMemo(
    () => [
      {
        title: t('nav.profile'),
        href: '/settings',
        icon: <UserCog size={18} />,
      },
      {
        title: t('nav.account_security'),
        href: '/settings/account',
        icon: <Wrench size={18} />,
      },
      ...(canReadOnt
        ? [
            {
              title: t('nav.acs'),
              href: '/settings/acs',
              icon: <RadioTower size={18} />,
            },
          ]
        : []),
      {
        title: t('nav.payment_gateway'),
        href: '/settings/gateway',
        icon: <CreditCard size={18} />,
      },
      {
        title: t('nav.billing_settings'),
        href: '/settings/billing',
        icon: <ReceiptText size={18} />,
      },
      {
        title: t('nav.payment_methods'),
        href: '/settings/payment-methods',
        icon: <WalletCards size={18} />,
      },
      {
        title: t('nav.appearance'),
        href: '/settings/appearance',
        icon: <Palette size={18} />,
      },
      {
        title: t('nav.display'),
        href: '/settings/display',
        icon: <Monitor size={18} />,
      },
    ],
    [t, canReadOnt]
  )

  return (
    <>
      {/* ===== Top Heading ===== */}
      <Header>
        <Search className='me-auto' />
        <LanguageSwitch />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main fixed>
        <div className='space-y-0.5'>
          <h1 className='text-2xl font-bold tracking-tight md:text-3xl'>
            {t('settings.title')}
          </h1>
          <p className='text-muted-foreground'>{t('settings.display_desc')}</p>
        </div>
        <Separator className='my-4 lg:my-6' />
        <div className='flex flex-1 flex-col space-y-2 overflow-hidden md:space-y-2 lg:flex-row lg:space-y-0 lg:space-x-12'>
          <aside className='top-0 lg:sticky lg:w-1/5'>
            <SidebarNav items={sidebarNavItems} />
          </aside>
          <div className='flex w-full overflow-y-auto p-1 pr-4'>
            <Outlet />
          </div>
        </div>
      </Main>
    </>
  )
}
