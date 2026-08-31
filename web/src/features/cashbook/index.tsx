import { useState } from 'react'
import { ArrowLeftRight, Building2, Tag } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
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

  const { data: accounts = [], isLoading: isLoadingAccounts } = useCashAccountsQuery(false)
  const { data: categories = [], isLoading: isLoadingCategories } = useCashCategoriesQuery(false)
  const { data: transactions = [], isLoading: isLoadingTransactions } = useCashTransactionsQuery(filters)
  const { data: balances = {} } = useCashBalancesQuery(filters.fromUnix, filters.toUnix)

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
            <h2 className='text-2xl font-bold tracking-tight'>Buku Kas & Keuangan</h2>
            <p className='text-muted-foreground'>
              Pencatatan kasir fisik, rekening bank, mutasi pemasukan/pengeluaran operasional, dan arus kas.
            </p>
          </div>
          <CashbookPrimaryButtons />
        </div>

        {/* Ringkasan Saldo & Arus Kas */}
        <CashbookSummaryCards />

        {/* Tabs: Jurnal Mutasi, Rekening, Kategori */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className='flex flex-1 flex-col gap-4'>
          <div className='flex items-center justify-between border-b pb-2'>
            <TabsList>
              <TabsTrigger value='transactions' className='gap-1.5 text-xs sm:text-sm'>
                <ArrowLeftRight className='h-4 w-4' />
                Jurnal Mutasi Kas
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
              data={transactions}
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
