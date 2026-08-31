import { useMemo } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { Subscription } from '@/gen/v1/billing_pb'
import { useCustomersQuery } from '@/features/customer/api/use-customer'
import { usePlansQuery } from '@/features/billing/api/use-plans'
import {
  PROVISION_STATUS_META,
  subscriptionStatusBadge,
} from '@/features/billing/data/constants'
import { SubscriptionsRowActions } from './subscriptions-row-actions'

function serviceTypeBadge(serviceType: string) {
  if (serviceType === 'HOTSPOT') {
    return {
      label: 'Hotspot',
      className:
        'bg-purple-500/15 text-purple-700 dark:text-purple-400 border-purple-500/30',
    }
  }
  return {
    label: 'PPPoE',
    className:
      'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
  }
}

export function useSubscriptionsColumns(): ColumnDef<Subscription>[] {
  const { data: customers = [] } = useCustomersQuery()
  const { data: plans = [] } = usePlansQuery(false)

  const customerNameById = useMemo(() => {
    const map = new Map<string, string>()
    for (const customer of customers) map.set(customer.id, customer.name)
    return map
  }, [customers])

  const planNameById = useMemo(() => {
    const map = new Map<string, string>()
    for (const plan of plans) map.set(plan.id, plan.name)
    return map
  }, [plans])

  return useMemo(
    () => [
      {
        accessorKey: 'customerId',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Pelanggan' />
        ),
        cell: ({ row }) => {
          const name = row.original.customerName || customerNameById.get(row.original.customerId) || '-'
          const phone = row.original.customerPhone || ''
          return (
            <div>
              <div className='font-semibold'>{name}</div>
              {phone ? <div className='font-mono text-[11px] text-muted-foreground'>{phone}</div> : null}
            </div>
          )
        },
      },
      {
        accessorKey: 'planId',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Paket' />
        ),
        cell: ({ row }) => {
          const planName = row.original.planName || planNameById.get(row.original.planId) || '-'
          return <span>{planName}</span>
        },
      },
      {
        accessorKey: 'serviceType',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Layanan' />
        ),
        cell: ({ row }) => {
          const badge = serviceTypeBadge(row.original.serviceType)
          return (
            <Badge variant='outline' className={`text-xs ${badge.className}`}>
              {badge.label}
            </Badge>
          )
        },
      },
      {
        accessorKey: 'remoteUsername',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Username Kredensial' />
        ),
        cell: ({ row }) => (
          <span className='font-mono text-xs font-semibold'>
            {row.original.remoteUsername || '-'}
          </span>
        ),
      },
      {
        accessorKey: 'rateLimit',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Rate Limit / Profile' />
        ),
        cell: ({ row }) => (
          <span className='font-mono text-xs'>
            {row.original.rateLimit || row.original.routerProfile || '-'}
          </span>
        ),
        meta: { className: 'hidden lg:table-cell' },
      },
      {
        accessorKey: 'deviceId',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Router' />
        ),
        cell: ({ row }) => {
          const deviceLabel = row.original.deviceName || row.original.deviceId || '-'
          return (
            <span className='text-xs font-medium'>
              {deviceLabel}
            </span>
          )
        },
        meta: { className: 'hidden md:table-cell' },
      },
      {
        accessorKey: 'provisionStatus',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Provisi' />
        ),
        cell: ({ row }) => {
          const meta =
            PROVISION_STATUS_META[row.original.provisionStatus as keyof typeof PROVISION_STATUS_META] ||
            PROVISION_STATUS_META.NONE
          return (
            <Badge variant='outline' className={`text-xs ${meta.className}`}>
              {meta.label}
            </Badge>
          )
        },
      },
      {
        accessorKey: 'status',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Status Jaringan' />
        ),
        cell: ({ row }) => {
          const badge = subscriptionStatusBadge(row.original.status)
          return (
            <Badge variant='outline' className={`text-xs ${badge.className}`}>
              {badge.label}
            </Badge>
          )
        },
      },
      {
        id: 'actions',
        cell: ({ row }) => <SubscriptionsRowActions subscription={row.original} />,
        enableSorting: false,
      },
    ],
    [customerNameById, planNameById]
  )
}
