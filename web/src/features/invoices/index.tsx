import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useCustomersQuery } from '@/features/customer/api/use-customer'
import { useInvoicesQuery } from './api/use-invoices'
import { InvoicesDialogs } from './components/invoices-dialogs'
import { InvoicesPrimaryButtons } from './components/invoices-primary-buttons'
import { InvoicesProvider } from './components/invoices-provider'
import { InvoicesSummaryCards } from './components/invoices-summary-cards'
import { InvoicesTable } from './components/invoices-table'

function InvoicesContent() {
  const { data: invoices = [], isLoading: isLoadingInvoices } = useInvoicesQuery('', '')
  const { data: customers = [], isLoading: isLoadingCustomers } = useCustomersQuery()

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
            <h2 className='text-2xl font-bold tracking-tight'>Faktur & Tagihan (Invoices)</h2>
            <p className='text-muted-foreground'>
              Penerbitan tagihan bulanan, pencatatan pembayaran kasir POS, dan riwayat faktur pelanggan.
            </p>
          </div>
          <InvoicesPrimaryButtons />
        </div>

        {/* Ringkasan Tagihan */}
        <InvoicesSummaryCards invoices={invoices} />

        {/* Tabel Tagihan */}
        <InvoicesTable
          data={invoices}
          customers={customers}
          isLoading={isLoadingInvoices || isLoadingCustomers}
        />
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
