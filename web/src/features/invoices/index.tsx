import { useMemo } from 'react'
import { AlertCircle, Receipt } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useDeviceStore } from '@/stores/device-store'
import { useCustomersQuery } from '@/features/customer/api/use-customer'
import { useSubscriptionsQuery } from '@/features/billing/api/use-billing'
import { useInvoicesQuery } from './api/use-invoices'
import { InvoicesDialogs } from './components/invoices-dialogs'
import { InvoicesPrimaryButtons } from './components/invoices-primary-buttons'
import { InvoicesProvider } from './components/invoices-provider'
import { InvoicesSummaryCards } from './components/invoices-summary-cards'
import { InvoicesTable } from './components/invoices-table'

function InvoicesContent() {
  const { selectedDeviceId } = useDeviceStore()
  const { data: devices = [] } = useDevicesQuery()
  const { data: invoices = [], isLoading: isLoadingInvoices } = useInvoicesQuery('', '')
  const { data: customers = [], isLoading: isLoadingCustomers } = useCustomersQuery()
  const { data: subscriptions = [], isLoading: isLoadingSubs } = useSubscriptionsQuery('')

  const currentDevice = devices.find((d) => d.id === selectedDeviceId)

  const { subscriptionMap, customerDeviceMap } = useMemo(() => {
    const subMap = new Map<string, (typeof subscriptions)[0]>()
    const custDevMap = new Map<string, Set<string>>()
    for (const s of subscriptions) {
      subMap.set(s.id, s)
      if (s.customerId && s.deviceId) {
        let set = custDevMap.get(s.customerId)
        if (!set) {
          set = new Set<string>()
          custDevMap.set(s.customerId, set)
        }
        set.add(s.deviceId)
      }
    }
    return { subscriptionMap: subMap, customerDeviceMap: custDevMap }
  }, [subscriptions])

  const filteredInvoices = useMemo(() => {
    if (!selectedDeviceId) return []
    return invoices.filter((inv) => {
      if (inv.subscriptionId) {
        const sub = subscriptionMap.get(inv.subscriptionId)
        return sub?.deviceId === selectedDeviceId
      }
      if (inv.customerId) {
        const devs = customerDeviceMap.get(inv.customerId)
        return devs ? devs.has(selectedDeviceId) : false
      }
      return false
    })
  }, [invoices, selectedDeviceId, subscriptionMap, customerDeviceMap])

  const filteredCustomers = useMemo(() => {
    if (!selectedDeviceId) return []
    return customers.filter((c) => customerDeviceMap.get(c.id)?.has(selectedDeviceId))
  }, [customers, customerDeviceMap, selectedDeviceId])

  const isLoading = isLoadingInvoices || isLoadingCustomers || isLoadingSubs

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        {/* Header Title & Actions */}
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <div className='flex items-center gap-2'>
              <Receipt className='size-6 text-primary' />
              <h2 className='text-2xl font-bold tracking-tight'>Faktur & Tagihan (Invoices)</h2>
            </div>
            <p className='text-sm text-muted-foreground mt-0.5'>
              {currentDevice ? (
                <>
                  Router:{' '}
                  <span className='font-semibold text-foreground'>
                    {currentDevice.name}
                  </span>{' '}
                  ({currentDevice.host})
                </>
              ) : (
                'Pilih router di sidebar untuk melihat faktur & tagihan.'
              )}
            </p>
          </div>
          <InvoicesPrimaryButtons />
        </div>

        {!selectedDeviceId ? (
          <div className='flex flex-col items-center justify-center p-12 text-center rounded-lg border border-dashed'>
            <AlertCircle className='size-10 text-muted-foreground mb-3' />
            <h3 className='text-lg font-semibold'>No Router Selected</h3>
            <p className='text-sm text-muted-foreground max-w-sm mt-1'>
              Silakan pilih router MikroTik dari dropdown di sidebar untuk melihat tagihan & faktur.
            </p>
          </div>
        ) : (
          <>
            {/* Ringkasan Tagihan */}
            <InvoicesSummaryCards invoices={filteredInvoices} />

            {/* Tabel Tagihan */}
            <InvoicesTable
              data={filteredInvoices}
              customers={filteredCustomers}
              isLoading={isLoading}
            />
          </>
        )}
      </Main>

      <InvoicesDialogs />
    </>
  )
}

export function Invoices() {
  return (
    <InvoicesProvider>
      <InvoicesContent />
    </InvoicesProvider>
  )
}
export default Invoices
