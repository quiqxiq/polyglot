import { CalendarClock, CalendarDays, Receipt, Wifi } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import type { Customer } from '@/gen/v1/customer_pb'
import type { Subscription } from '@/gen/v1/subscription_pb'
import type { Invoice } from '@/gen/v1/billing_pb'
import { formatCurrency, formatUnixDate } from '@/lib/format'

interface CustomerSummaryCardsProps {
  customer: Customer
  subscriptions: Subscription[]
  invoices: Invoice[]
}

export function CustomerSummaryCards({
  customer,
  subscriptions,
  invoices,
}: CustomerSummaryCardsProps) {
  const activeSubscriptions = subscriptions.filter((s) => s.status === 'ACTIVE')
  const unpaidInvoices = invoices.filter((inv) => inv.status !== 'PAID')
  const outstanding = unpaidInvoices.reduce(
    (sum, inv) => sum + Math.max(0, inv.total - inv.paidAmount),
    0
  )

  const nextDue = [...unpaidInvoices]
    .filter((inv) => Number(inv.dueDateUnix) > 0)
    .sort((a, b) => Number(a.dueDateUnix) - Number(b.dueDateUnix))[0]

  const registeredAt = formatUnixDate(
    customer.registeredAtUnix || customer.createdAtUnix
  )

  const cards = [
    {
      label: 'Layanan Aktif',
      value: `${activeSubscriptions.length}`,
      hint: `dari ${subscriptions.length} total langganan`,
      icon: Wifi,
      tone: 'text-emerald-600 dark:text-emerald-400',
    },
    {
      label: 'Tagihan Tertunggak',
      value: formatCurrency(outstanding),
      hint:
        unpaidInvoices.length > 0
          ? `${unpaidInvoices.length} faktur belum lunas`
          : 'Semua tagihan lunas',
      icon: Receipt,
      tone:
        unpaidInvoices.length > 0
          ? 'text-amber-600 dark:text-amber-400'
          : 'text-muted-foreground',
    },
    {
      label: 'Jatuh Tempo Terdekat',
      value: nextDue ? formatUnixDate(nextDue.dueDateUnix) : '-',
      hint: nextDue ? `Periode ${nextDue.period || '-'}` : 'Tidak ada tagihan terbuka',
      icon: CalendarClock,
      tone: 'text-foreground',
    },
    {
      label: 'Terdaftar Sejak',
      value: registeredAt,
      hint: `Kode ${customer.customerCode || customer.id}`,
      icon: CalendarDays,
      tone: 'text-foreground',
    },
  ]

  return (
    <div className='grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4'>
      {cards.map((card) => (
        <Card key={card.label} className='shadow-xs'>
          <CardContent className='p-4'>
            <div className='flex items-center justify-between'>
              <span className='text-xs font-medium text-muted-foreground'>
                {card.label}
              </span>
              <card.icon className={`h-4 w-4 ${card.tone}`} />
            </div>
            <p className='mt-2 truncate font-mono text-lg font-bold tracking-tight'>
              {card.value}
            </p>
            <p className='mt-0.5 text-[11px] text-muted-foreground'>{card.hint}</p>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
