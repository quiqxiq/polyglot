import { useMemo } from 'react'
import { AlertCircle, Receipt, Search as SearchIcon, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { useInvoiceListQuery } from '@/features/billing/api/use-billing-ops'
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
import { InvoicesDialogs } from './components/invoices-dialogs'
import { InvoicesPrimaryButtons } from './components/invoices-primary-buttons'
import { InvoicesProvider, useInvoices } from './components/invoices-provider'
import { InvoicesSummaryCards } from './components/invoices-summary-cards'
import { InvoicesTable } from './components/invoices-table'

function InvoicesContent() {
  const { selectedDeviceId } = useDeviceStore()
  const { data: devices = [] } = useDevicesQuery()
  const { filters, setFilters, resetFilters } = useInvoices()
  const {
    data: invoiceData,
    isLoading: isLoadingInvoices,
  } = useInvoiceListQuery({
    customerId: filters.customerId,
    status: filters.status,
    period: filters.period,
    search: filters.search,
    overdueOnly: filters.overdueOnly,
    limit: 500,
  })
  const invoices = useMemo(() => invoiceData?.invoices ?? [], [invoiceData])
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

        {/* Filter Tagihan */}
        <div className='flex flex-wrap items-center gap-2 rounded-lg border bg-muted/20 p-2.5'>
          <div className='relative'>
            <SearchIcon className='absolute left-2 top-2.5 h-4 w-4 text-muted-foreground' />
            <Input
              className='h-9 w-60 pl-8'
              placeholder='Cari no. invoice / pelanggan'
              value={filters.search}
              onChange={(e) => setFilters((f) => ({ ...f, search: e.target.value }))}
            />
          </div>
          <Select
            value={filters.status || 'ALL'}
            onValueChange={(v) => setFilters((f) => ({ ...f, status: v === 'ALL' ? '' : v }))}
          >
            <SelectTrigger className='h-9 w-40'>
              <SelectValue placeholder='Semua status' />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='ALL'>Semua status</SelectItem>
              <SelectItem value='UNPAID'>Belum Bayar</SelectItem>
              <SelectItem value='PARTIAL'>Dibayar Sebagian</SelectItem>
              <SelectItem value='PAID'>Lunas</SelectItem>
              <SelectItem value='OVERDUE'>Lewat Jatuh Tempo</SelectItem>
              <SelectItem value='CANCELLED'>Dibatalkan</SelectItem>
            </SelectContent>
          </Select>
          <Input
            className='h-9 w-36'
            placeholder='Periode YYYY-MM'
            value={filters.period}
            onChange={(e) => setFilters((f) => ({ ...f, period: e.target.value }))}
          />
          <label className='flex items-center gap-2 text-sm text-muted-foreground'>
            <Switch
              checked={filters.overdueOnly}
              onCheckedChange={(checked) =>
                setFilters((f) => ({ ...f, overdueOnly: checked }))
              }
            />
            Hanya lewat jatuh tempo
          </label>
          <Button variant='ghost' size='sm' className='gap-1' onClick={resetFilters}>
            <X className='h-4 w-4' /> Reset
          </Button>
          <span className='ms-auto text-xs text-muted-foreground'>
            {invoiceData?.total ?? 0} tagihan
          </span>
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
