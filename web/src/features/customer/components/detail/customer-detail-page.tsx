import { useState } from 'react'
import { useNavigate, useParams } from '@tanstack/react-router'
import type { Invoice } from '@/gen/v1/billing_pb'
import type { Subscription } from '@/gen/v1/subscription_pb'
import { SearchX } from 'lucide-react'
import { useI18n } from '@/context/i18n-provider'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  useInvoicesQuery,
  useSubscriptionsQuery,
} from '@/features/billing/api/use-billing'
import { usePlansQuery } from '@/features/billing/api/use-plans'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { InvoicesDialogs } from '@/features/invoices/components/invoices-dialogs'
import {
  InvoicesProvider,
  useInvoices,
} from '@/features/invoices/components/invoices-provider'
import { CustomerOntTab } from '@/features/ont/components/customer/customer-ont-tab'
import { ChangePlanDialog } from '@/features/subscriptions/components/change-plan-dialog'
import { LifecycleConfirmDialog } from '@/features/subscriptions/components/lifecycle-confirm-dialog'
import { SubscriptionsCreateDialog } from '@/features/subscriptions/components/subscriptions-create-dialog'
import { SubscriptionsEditDialog } from '@/features/subscriptions/components/subscriptions-edit-dialog'
import { SuspendDialog } from '@/features/subscriptions/components/suspend-dialog'
import { useCustomer360Query, useCustomerQuery } from '../../api/use-customer'
import { CustomersDeleteDialog } from '../customers-delete-dialog'
import { CustomersMutateDrawer } from '../customers-mutate-drawer'
import { CustomerDetailHeader } from './customer-detail-header'
import { CustomerInvoicesTab } from './customer-invoices-tab'
import { CustomerOverviewTab } from './customer-overview-tab'
import { CustomerPaymentsTab } from './customer-payments-tab'
import { CustomerPSBTab } from './customer-psb-tab'
import { CustomerSubscriptionsTab } from './customer-subscriptions-tab'
import { CustomerSummaryCards } from './customer-summary-cards'
import type { SubscriptionLifecycleAction } from './subscription-card'

interface CustomerDetailContentProps {
  customerId: string
}

type DetailTab =
  | 'ringkasan'
  | 'layanan'
  | 'tagihan'
  | 'pembayaran'
  | 'psb'
  | 'ont'

function PageHeader() {
  return (
    <Header fixed>
      <Search className='me-auto' />
      <ThemeSwitch />
      <ConfigDrawer />
      <ProfileDropdown />
    </Header>
  )
}

export function CustomerDetailPage() {
  const { customerId } = useParams({
    from: '/_authenticated/customers/$customerId',
  })
  return (
    <InvoicesProvider>
      <CustomerDetailContent customerId={customerId} />
    </InvoicesProvider>
  )
}

function CustomerDetailContent({ customerId }: CustomerDetailContentProps) {
  const navigate = useNavigate()
  const { t } = useI18n()
  const { setCurrentInvoice, setOpen: setInvoiceDialog } = useInvoices()

  const {
    data: customer,
    isLoading: isLoadingCustomer,
    isError,
  } = useCustomerQuery(customerId)
  const { data: subscriptions = [], isLoading: isLoadingSubscriptions } =
    useSubscriptionsQuery(customerId)
  const { data: invoices = [], isLoading: isLoadingInvoices } =
    useInvoicesQuery(customerId, '')
  const { data: customer360, isLoading: isLoading360 } =
    useCustomer360Query(customerId)
  const { data: plans = [] } = usePlansQuery(false)
  const { data: devices = [] } = useDevicesQuery()

  const [tab, setTab] = useState<DetailTab>('ringkasan')
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [createSubscriptionOpen, setCreateSubscriptionOpen] = useState(false)
  const [editSubscription, setEditSubscription] = useState<Subscription | null>(
    null
  )
  const [changePlanSubscription, setChangePlanSubscription] =
    useState<Subscription | null>(null)
  const [suspendSubscription, setSuspendSubscription] =
    useState<Subscription | null>(null)
  const [lifecycle, setLifecycle] = useState<{
    action: SubscriptionLifecycleAction
    subscription: Subscription
  } | null>(null)

  const openInvoice = (
    invoice: Invoice,
    dialog: 'detail' | 'cashier' | 'print'
  ) => {
    setCurrentInvoice(invoice)
    setInvoiceDialog(dialog)
  }

  if (isLoadingCustomer) {
    return (
      <>
        <PageHeader />
        <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
          <Skeleton className='h-5 w-64' />
          <Skeleton className='h-24 w-full rounded-xl' />
          <div className='grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4'>
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton key={index} className='h-24 w-full rounded-xl' />
            ))}
          </div>
          <Skeleton className='h-10 w-80' />
          <Skeleton className='h-64 w-full rounded-xl' />
        </Main>
      </>
    )
  }

  if (isError || !customer) {
    return (
      <>
        <PageHeader />
        <Main className='flex flex-1 flex-col items-center justify-center gap-3'>
          <Card className='w-full max-w-md border-dashed shadow-none'>
            <CardContent className='flex flex-col items-center py-10 text-center'>
              <SearchX className='mb-3 size-10 text-muted-foreground' />
              <h2 className='text-lg font-semibold'>
                Pelanggan Tidak Ditemukan
              </h2>
              <p className='mt-1 text-sm text-muted-foreground'>
                Data pelanggan mungkin sudah dihapus atau ID tidak valid.
              </p>
              <Button
                className='mt-4'
                onClick={() => navigate({ to: '/customers' })}
              >
                Kembali ke Daftar Pelanggan
              </Button>
            </CardContent>
          </Card>
        </Main>
      </>
    )
  }

  const activeCount = subscriptions.filter((s) => s.status === 'ACTIVE').length
  const unpaidCount = invoices.filter((inv) => inv.status !== 'PAID').length

  return (
    <>
      <PageHeader />

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <CustomerDetailHeader
          customer={customer}
          invoices={invoices}
          onEdit={() => setEditOpen(true)}
          onAddSubscription={() => setCreateSubscriptionOpen(true)}
          onPay={(invoice) => openInvoice(invoice, 'cashier')}
        />

        <CustomerSummaryCards
          customer={customer}
          subscriptions={subscriptions}
          invoices={invoices}
        />

        <Tabs
          value={tab}
          onValueChange={(value) => setTab(value as DetailTab)}
          className='gap-4'
        >
          <TabsList className='h-10 w-full justify-start sm:w-auto'>
            <TabsTrigger value='ringkasan'>Ringkasan</TabsTrigger>
            <TabsTrigger value='layanan' className='gap-2'>
              Layanan
              <Badge variant='secondary' className='px-1.5 text-[10px]'>
                {activeCount}/{subscriptions.length}
              </Badge>
            </TabsTrigger>
            <TabsTrigger value='pembayaran'>Pembayaran</TabsTrigger>
            <TabsTrigger value='psb'>Riwayat PSB</TabsTrigger>
            <TabsTrigger value='ont'>{t('ont.tab_title')}</TabsTrigger>
            <TabsTrigger value='tagihan' className='gap-2'>
              Tagihan
              {unpaidCount > 0 && (
                <Badge
                  variant='outline'
                  className='border-amber-500/30 bg-amber-500/15 px-1.5 text-[10px] text-amber-700 dark:text-amber-400'
                >
                  {unpaidCount}
                </Badge>
              )}
            </TabsTrigger>
          </TabsList>

          <TabsContent value='ringkasan'>
            <CustomerOverviewTab
              customer={customer}
              subscriptions={subscriptions}
              invoices={invoices}
              onOpenTab={setTab}
              onOpenInvoice={(invoice) => openInvoice(invoice, 'detail')}
            />
          </TabsContent>

          <TabsContent value='layanan'>
            <CustomerSubscriptionsTab
              subscriptions={subscriptions}
              plans={plans}
              devices={devices}
              isLoading={isLoadingSubscriptions}
              onAddSubscription={() => setCreateSubscriptionOpen(true)}
              onEdit={setEditSubscription}
              onChangePlan={setChangePlanSubscription}
              onSuspend={setSuspendSubscription}
              onLifecycle={(action, subscription) =>
                setLifecycle({ action, subscription })
              }
            />
          </TabsContent>

          <TabsContent value='tagihan'>
            <CustomerInvoicesTab
              invoices={invoices}
              isLoading={isLoadingInvoices}
              onOpenInvoice={(invoice) => openInvoice(invoice, 'detail')}
              onPayInvoice={(invoice) => openInvoice(invoice, 'cashier')}
              onPrintInvoice={(invoice) => openInvoice(invoice, 'print')}
            />
          </TabsContent>

          <TabsContent value='pembayaran'>
            <CustomerPaymentsTab
              payments={customer360?.payments || []}
              isLoading={isLoading360}
            />
          </TabsContent>

          <TabsContent value='psb'>
            <CustomerPSBTab
              registrations={customer360?.registrations || []}
              isLoading={isLoading360}
            />
          </TabsContent>

          <TabsContent value='ont'>
            <CustomerOntTab
              customerId={customerId}
              subscriptions={subscriptions}
            />
          </TabsContent>
        </Tabs>
      </Main>

      <CustomersMutateDrawer
        key={`customer-edit-${customer.id}`}
        open={editOpen}
        onOpenChange={setEditOpen}
        currentRow={customer}
      />

      <CustomersDeleteDialog
        key={`customer-delete-${customer.id}`}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        currentRow={customer}
        onDeleted={() => navigate({ to: '/customers' })}
      />

      <SubscriptionsCreateDialog
        key={`customer-create-sub-${customer.id}`}
        open={createSubscriptionOpen}
        onOpenChange={setCreateSubscriptionOpen}
        initialCustomerId={customer.id}
        lockCustomer
      />

      {editSubscription && (
        <SubscriptionsEditDialog
          key={`sub-edit-${editSubscription.id}`}
          open
          onOpenChange={(isOpen) => !isOpen && setEditSubscription(null)}
          currentRow={editSubscription}
        />
      )}

      {changePlanSubscription && (
        <ChangePlanDialog
          key={`sub-plan-${changePlanSubscription.id}`}
          open
          onOpenChange={(isOpen) => !isOpen && setChangePlanSubscription(null)}
          currentRow={changePlanSubscription}
        />
      )}

      {suspendSubscription && (
        <SuspendDialog
          key={`sub-suspend-${suspendSubscription.id}`}
          open
          onOpenChange={(isOpen) => !isOpen && setSuspendSubscription(null)}
          currentRow={suspendSubscription}
        />
      )}

      {lifecycle && (
        <LifecycleConfirmDialog
          key={`sub-lifecycle-${lifecycle.subscription.id}-${lifecycle.action}`}
          action={lifecycle.action}
          open
          onOpenChange={(isOpen) => !isOpen && setLifecycle(null)}
          currentRow={lifecycle.subscription}
        />
      )}

      <InvoicesDialogs />
    </>
  )
}
