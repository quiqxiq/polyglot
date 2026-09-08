import { useMemo, useState } from 'react'
import { AlertCircle, ArrowLeftRight, Building2, Tag, Wallet } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useDeviceStore } from '@/stores/device-store'
import { useSubscriptionsQuery, useInvoicesQuery } from '@/features/billing/api/use-billing'
import { useCustomersQuery } from '@/features/customer/api/use-customer'
import {
  useCashAccountsQuery,
  useCashBalancesQuery,
  useCashCategoriesQuery,
  useCashTransactionsQuery,
} from './api/use-cashbook'
import { CashbookDialogs } from './components/cashbook-dialogs'
import { CashbookPrimaryButtons } from './components/cashbook-primary-buttons'
import { CashbookProvider, useCashbook } from './components/cashbook-provider'
import { CashbookSummaryCards } from './components/cashbook-summary-cards'
import { CashbookTransactionsTable } from './components/cashbook-transactions-table'
import { CashbookAccountsTable } from './components/cashbook-accounts-table'
import { CashbookCategoriesTable } from './components/cashbook-categories-table'

function CashbookContent() {
  const [activeTab, setActiveTab] = useState('transactions')
  const { filters } = useCashbook()
  const { selectedDeviceId } = useDeviceStore()
  const { data: devices = [] } = useDevicesQuery()

  const { data: accounts = [], isLoading: isLoadingAccounts } = useCashAccountsQuery(false)
  const { data: categories = [], isLoading: isLoadingCategories } = useCashCategoriesQuery(false)
  const { data: transactions = [], isLoading: isLoadingTransactions } = useCashTransactionsQuery(filters)
  const { data: balances = {} } = useCashBalancesQuery(filters.fromUnix, filters.toUnix)
  const { data: subscriptions = [] } = useSubscriptionsQuery('')
  const { data: invoices = [] } = useInvoicesQuery('', '')
  const { data: customers = [] } = useCustomersQuery()

  const currentDevice = devices.find((d) => d.id === selectedDeviceId)

  // Identifikasi invoice & customer yang berada di router terpilih
  const { routerInvoiceNumbers, routerCustomerCodes } = useMemo(() => {
    if (!selectedDeviceId) {
      return { routerInvoiceNumbers: new Set<string>(), routerCustomerCodes: new Set<string>() }
    }
    const routerSubIds = new Set(
      subscriptions.filter((s) => s.deviceId === selectedDeviceId).map((s) => s.id)
    )
    const routerCustIds = new Set(
      subscriptions.filter((s) => s.deviceId === selectedDeviceId).map((s) => s.customerId)
    )

    const invNums = new Set<string>()
    for (const inv of invoices) {
      if (
        (inv.subscriptionId && routerSubIds.has(inv.subscriptionId)) ||
        (inv.customerId && routerCustIds.has(inv.customerId))
      ) {
        if (inv.invoiceNumber) invNums.add(inv.invoiceNumber.toLowerCase())
      }
    }

    const custCodes = new Set<string>()
    for (const c of customers) {
      if (routerCustIds.has(c.id) && c.customerCode) {
        custCodes.add(c.customerCode.toLowerCase())
      }
    }

    return { routerInvoiceNumbers: invNums, routerCustomerCodes: custCodes }
  }, [subscriptions, invoices, customers, selectedDeviceId])

  // Filter transaksi kas berdasarkan router terpilih
  const filteredTransactions = useMemo(() => {
    if (!selectedDeviceId) return []
    return transactions.filter((tx) => {
      // Transaksi masuk dari pelunasan invoice pelanggan
      if (tx.sourceType === 'PAYMENT') {
        const desc = tx.description.toLowerCase()
        for (const invNum of routerInvoiceNumbers) {
          if (invNum && desc.includes(invNum)) return true
        }
        for (const custCode of routerCustomerCodes) {
          if (custCode && desc.includes(custCode)) return true
        }
        return false
      }
      // Mutasi operasional / transfer: jika menyebut router lain secara spesifik, skip
      const desc = tx.description.toLowerCase()
      for (const dev of devices) {
        if (dev.id !== selectedDeviceId && dev.name && desc.includes(dev.name.toLowerCase())) {
          return false
        }
      }
      return true
    })
  }, [transactions, selectedDeviceId, routerInvoiceNumbers, routerCustomerCodes, devices])

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        {/* Title & Actions */}
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <div className='flex items-center gap-2'>
              <Wallet className='size-6 text-primary' />
              <h2 className='text-2xl font-bold tracking-tight'>Buku Kas & Keuangan</h2>
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
                'Pilih router di sidebar untuk melihat mutasi kas & keuangan.'
              )}
            </p>
          </div>
          <CashbookPrimaryButtons />
        </div>

        {!selectedDeviceId ? (
          <div className='flex flex-col items-center justify-center p-12 text-center rounded-lg border border-dashed'>
            <AlertCircle className='size-10 text-muted-foreground mb-3' />
            <h3 className='text-lg font-semibold'>No Router Selected</h3>
            <p className='text-sm text-muted-foreground max-w-sm mt-1'>
              Silakan pilih router MikroTik dari dropdown di sidebar untuk melihat mutasi kas & keuangan.
            </p>
          </div>
        ) : (
          <>
            {/* Ringkasan Saldo & Arus Kas */}
            <CashbookSummaryCards transactions={filteredTransactions} />

            {/* Tabs: Jurnal Mutasi, Rekening, Kategori */}
            <Tabs value={activeTab} onValueChange={setActiveTab} className='flex flex-1 flex-col gap-4'>
              <div className='flex items-center justify-between border-b pb-2'>
                <TabsList>
                  <TabsTrigger value='transactions' className='gap-1.5 text-xs sm:text-sm'>
                    <ArrowLeftRight className='h-4 w-4' />
                    Jurnal Mutasi Kas ({filteredTransactions.length})
                  </TabsTrigger>
                  <TabsTrigger value='accounts' className='gap-1.5 text-xs sm:text-sm'>
                    <Building2 className='h-4 w-4' />
                    Rekening Kas & Bank ({accounts.length})
                  </TabsTrigger>
                  <TabsTrigger value='categories' className='gap-1.5 text-xs sm:text-sm'>
                    <Tag className='h-4 w-4' />
                    Kategori Pos Kas ({categories.length})
                  </TabsTrigger>
                </TabsList>
              </div>

              <TabsContent value='transactions' className='m-0 flex-1 flex flex-col'>
                <CashbookTransactionsTable
                  data={filteredTransactions}
                  accounts={accounts}
                  categories={categories}
                  isLoading={isLoadingTransactions}
                />
              </TabsContent>

              <TabsContent value='accounts' className='m-0 flex-1 flex flex-col'>
                <CashbookAccountsTable
                  data={accounts}
                  balances={balances}
                  isLoading={isLoadingAccounts}
                />
              </TabsContent>

              <TabsContent value='categories' className='m-0 flex-1 flex flex-col'>
                <CashbookCategoriesTable
                  data={categories}
                  isLoading={isLoadingCategories}
                />
              </TabsContent>
            </Tabs>
          </>
        )}
      </Main>

      <CashbookDialogs />
    </>
  )
}

export function Cashbook() {
  return (
    <CashbookProvider>
      <CashbookContent />
    </CashbookProvider>
  )
}
export default Cashbook
