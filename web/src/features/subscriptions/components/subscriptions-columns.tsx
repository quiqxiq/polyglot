import { useMemo } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import type { Subscription } from '@/gen/v1/subscription_pb'
import { useCustomersQuery } from '@/features/customer/api/use-customer'
import { usePlansQuery } from '@/features/billing/api/use-plans'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
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
  if (serviceType === 'DEDICATED') {
    return {
      label: 'Dedicated',
      className:
        'bg-orange-500/15 text-orange-700 dark:text-orange-400 border-orange-500/30',
    }
  }
  return {
    label: 'PPPoE',
    className:
      'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
  }
}

function formatPlanRate(downloadKbps?: number, uploadKbps?: number) {
  if (!downloadKbps && !uploadKbps) return ''
  const formatSide = (kbps?: number) => {
    if (!kbps || kbps <= 0) return ''
    if (kbps >= 1000) return `${Math.round(kbps / 1000)}M`
    return `${kbps}k`
  }
  const dl = formatSide(downloadKbps)
  const ul = formatSide(uploadKbps)
  if (!dl && !ul) return ''
  return `${dl || '0'}/${ul || '0'}`
}

export function useSubscriptionsColumns(): ColumnDef<Subscription>[] {
  const { data: customers = [] } = useCustomersQuery()
  const { data: plans = [] } = usePlansQuery(false)
  const { data: devices = [] } = useDevicesQuery()

  const customerNameById = useMemo(() => {
    const map = new Map<string, string>()
    for (const customer of customers) map.set(customer.id, customer.name)
    return map
  }, [customers])

  const planById = useMemo(() => {
    const map = new Map<string, (typeof plans)[0]>()
    for (const plan of plans) map.set(plan.id, plan)
    return map
  }, [plans])

  /** ID → nama router */
  const deviceNameById = useMemo(() => {
    const map = new Map<string, string>()
    for (const device of devices) {
      if (device.id && device.name) map.set(device.id, device.name)
    }
    return map
  }, [devices])

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
          const plan = planById.get(row.original.planId)
          const planName = row.original.planName || plan?.name || row.original.planId || '-'
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
          <DataTableColumnHeader column={column} title='Username' />
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
          <DataTableColumnHeader column={column} title='Rate Limit' />
        ),
        cell: ({ row }) => {
          const sub = row.original
          const plan = planById.get(sub.planId)
          const rateLimit =
            sub.pppoeConfig?.rateLimit ||
            sub.hotspotConfig?.rateLimit ||
            sub.rateLimit ||
            formatPlanRate(plan?.bandwidthDownloadKbps, plan?.bandwidthUploadKbps) ||
            '-'
          return (
            <span className='font-mono text-xs font-semibold text-foreground'>
              {rateLimit}
            </span>
          )
        },
        meta: { className: 'hidden lg:table-cell' },
      },
      {
        accessorKey: 'deviceId',
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title='Router' />
        ),
        cell: ({ row }) => {
          const sub = row.original
          // Coba deviceName dari backend, lalu dari devices query, lalu deviceId sebagai last resort
          const deviceLabel =
            sub.deviceName ||
            deviceNameById.get(sub.deviceId) ||
            sub.deviceId ||
            '-'
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
          <DataTableColumnHeader column={column} title='Status Router' />
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
          <DataTableColumnHeader column={column} title='Status' />
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
    [customerNameById, planById, deviceNameById]
  )
}
