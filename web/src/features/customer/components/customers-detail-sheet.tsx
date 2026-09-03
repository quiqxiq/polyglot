import { useState } from 'react'
import {
  Copy,
  MessageCircle,
  Pencil,
  Plus,
  User,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'
import { toast } from 'sonner'
import type { Customer } from '@/gen/v1/customer_pb'
import type { Subscription } from '@/gen/v1/subscription_pb'
import type { Invoice } from '@/gen/v1/billing_pb'
import { useSubscriptionsQuery, useInvoicesQuery } from '@/features/billing/api/use-billing'
import { usePlansQuery } from '@/features/billing/api/use-plans'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { CashierDialog } from '@/features/invoices/components/cashier-dialog'
import { SubscriptionsEditDialog } from '@/features/subscriptions/components/subscriptions-edit-dialog'
import { customerStatusBadge } from '../data/constants'
import { useCustomers } from './customers-provider'
import { CustomerOverviewTab } from './customer-overview-tab'
import { CustomerSubscriptionsTab } from './customer-subscriptions-tab'
import { CustomerInvoicesTab } from './customer-invoices-tab'

interface CustomersDetailSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  customer: Customer | null
}

export function CustomersDetailSheet({
  open,
  onOpenChange,
  customer,
}: CustomersDetailSheetProps) {
  const { setOpen, setCurrentRow } = useCustomers()
  const customerId = customer?.id || ''
  const isEnabled = open && Boolean(customerId)

  const [selectedInvoice, setSelectedInvoice] = useState<Invoice | null>(null)
  const [selectedSubscription, setSelectedSubscription] = useState<Subscription | null>(null)

  const { data: subscriptions = [], isLoading: isLoadingSubs } =
    useSubscriptionsQuery(customerId, { enabled: isEnabled })
  const { data: invoices = [], isLoading: isLoadingInvoices } =
    useInvoicesQuery(customerId, '', { enabled: isEnabled })
  const { data: plans = [] } = usePlansQuery(false)
  const { data: devices = [] } = useDevicesQuery()

  if (!customer) return null

  const statusMeta = customerStatusBadge(customer.status)
  const cleanPhone = customer.phone.replace(/\D/g, '')
  const waPhone = cleanPhone.startsWith('0') ? `62${cleanPhone.slice(1)}` : cleanPhone
  const unpaidCount = invoices.filter((i) => i.status !== 'PAID').length

  const handleAddSubscription = () => {
    setCurrentRow(customer)
    setOpen('create-subscription')
  }

  const handleEdit = () => {
    setCurrentRow(customer)
    setOpen('update')
  }

  const copyCustomerCode = () => {
    const code = customer.customerCode || customer.id
    navigator.clipboard.writeText(code)
    toast.success('Kode pelanggan berhasil disalin')
  }

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className='flex w-full flex-col sm:max-w-xl md:max-w-2xl lg:max-w-3xl p-0 gap-0 border-l'>
          {/* ─── Hero Dossier Header ─── */}
          <SheetHeader className='border-b p-6 pb-5 bg-muted/10'>
            <div className='flex flex-col sm:flex-row sm:items-center justify-between gap-4'>
              <div className='flex items-center gap-3.5 min-w-0'>
                {/* Avatar Inisial */}
                <div className='flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-primary border border-primary/20 shadow-xs'>
                  <User className='h-6 w-6' />
                </div>

                <div className='min-w-0'>
                  <div className='flex items-center gap-2'>
                    <SheetTitle className='text-lg sm:text-xl font-bold truncate text-foreground'>
                      {customer.name}
                    </SheetTitle>
                    <Badge variant='outline' className={`text-[10px] px-2 py-0.5 font-semibold ${statusMeta.className}`}>
                      {statusMeta.label}
                    </Badge>
                  </div>

                  <SheetDescription className='mt-0.5 flex items-center gap-1.5 font-mono text-xs text-muted-foreground'>
                    <span>{customer.customerCode || customer.id}</span>
                    <Button
                      size='icon'
                      variant='ghost'
                      className='h-5 w-5 text-muted-foreground hover:text-foreground'
                      onClick={copyCustomerCode}
                      title='Salin kode'
                    >
                      <Copy className='h-3 w-3' />
                    </Button>
                  </SheetDescription>
                </div>
              </div>

              {/* Quick Actions Bar */}
              <div className='flex items-center gap-2 shrink-0 self-start sm:self-center'>
                {customer.phone && (
                  <Button
                    size='sm'
                    variant='outline'
                    className='h-8 gap-1.5 text-xs text-emerald-600 hover:text-emerald-700 hover:bg-emerald-500/10 border-emerald-500/30'
                    asChild
                  >
                    <a
                      href={`https://wa.me/${waPhone}?text=Halo%20${encodeURIComponent(customer.name)}`}
                      target='_blank'
                      rel='noopener noreferrer'
                    >
                      <MessageCircle className='h-3.5 w-3.5' />
                      WhatsApp
                    </a>
                  </Button>
                )}
                <Button size='sm' variant='outline' className='h-8 gap-1.5 text-xs' onClick={handleEdit}>
                  <Pencil className='h-3.5 w-3.5' />
                  Edit
                </Button>
                <Button size='sm' className='h-8 gap-1.5 text-xs shadow-xs' onClick={handleAddSubscription}>
                  <Plus className='h-3.5 w-3.5' />
                  Langganan
                </Button>
              </div>
            </div>
          </SheetHeader>

          {/* ─── Tab Navigation Bar ─── */}
          <Tabs defaultValue='overview' className='flex flex-1 flex-col overflow-hidden'>
            <div className='border-b px-6 bg-background'>
              <TabsList className='h-11 w-full justify-start rounded-none bg-transparent p-0 gap-6'>
                <TabsTrigger
                  value='overview'
                  className='relative h-11 rounded-none border-b-2 border-transparent px-2 pb-3 pt-2 font-medium text-xs sm:text-sm text-muted-foreground hover:text-foreground data-[state=active]:border-primary data-[state=active]:text-foreground data-[state=active]:font-bold'
                >
                  Ringkasan
                </TabsTrigger>
                <TabsTrigger
                  value='subscriptions'
                  className='relative h-11 rounded-none border-b-2 border-transparent px-2 pb-3 pt-2 font-medium text-xs sm:text-sm text-muted-foreground hover:text-foreground data-[state=active]:border-primary data-[state=active]:text-foreground data-[state=active]:font-bold'
                >
                  Layanan & Langganan
                  {subscriptions.length > 0 && (
                    <span className='ml-1.5 rounded-full bg-primary/15 px-2 py-0.5 text-[10px] font-semibold text-primary'>
                      {subscriptions.length}
                    </span>
                  )}
                </TabsTrigger>
                <TabsTrigger
                  value='invoices'
                  className='relative h-11 rounded-none border-b-2 border-transparent px-2 pb-3 pt-2 font-medium text-xs sm:text-sm text-muted-foreground hover:text-foreground data-[state=active]:border-primary data-[state=active]:text-foreground data-[state=active]:font-bold'
                >
                  Tagihan
                  {unpaidCount > 0 ? (
                    <span className='ml-1.5 rounded-full bg-amber-500/15 px-2 py-0.5 text-[10px] font-semibold text-amber-600 dark:text-amber-400'>
                      {unpaidCount}
                    </span>
                  ) : invoices.length > 0 ? (
                    <span className='ml-1.5 rounded-full bg-muted px-2 py-0.5 text-[10px] font-semibold text-muted-foreground'>
                      {invoices.length}
                    </span>
                  ) : null}
                </TabsTrigger>
              </TabsList>
            </div>

            {/* ─── Tab Content Views ─── */}
            <ScrollArea className='flex-1'>
              <TabsContent value='overview' className='m-0 focus-visible:outline-hidden'>
                <CustomerOverviewTab
                  customer={customer}
                  subscriptions={subscriptions}
                  invoices={invoices}
                />
              </TabsContent>

              <TabsContent value='subscriptions' className='m-0 focus-visible:outline-hidden'>
                <CustomerSubscriptionsTab
                  subscriptions={subscriptions}
                  plans={plans}
                  devices={devices}
                  isLoading={isLoadingSubs}
                  onAddSubscription={handleAddSubscription}
                  onEditSubscription={(sub) => setSelectedSubscription(sub)}
                />
              </TabsContent>

              <TabsContent value='invoices' className='m-0 focus-visible:outline-hidden'>
                <CustomerInvoicesTab
                  invoices={invoices}
                  isLoading={isLoadingInvoices}
                  onPayInvoice={(inv) => setSelectedInvoice(inv)}
                />
              </TabsContent>
            </ScrollArea>
          </Tabs>
        </SheetContent>
      </Sheet>

      {/* ─── Modal Kasir Pembayaran Cepat ─── */}
      {selectedInvoice && (
        <CashierDialog
          open={Boolean(selectedInvoice)}
          onOpenChange={(isOpen) => !isOpen && setSelectedInvoice(null)}
          currentInvoice={selectedInvoice}
        />
      )}

      {/* ─── Modal Edit Langganan ─── */}
      {selectedSubscription && (
        <SubscriptionsEditDialog
          open={Boolean(selectedSubscription)}
          onOpenChange={(isOpen) => !isOpen && setSelectedSubscription(null)}
          currentRow={selectedSubscription}
        />
      )}
    </>
  )
}
