import { type ColumnDef } from '@tanstack/react-table'
import { ArrowDownLeft, ArrowUpRight } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { type CashTransaction, type CashAccount, type CashCategory } from '@/gen/v1/cashbook_pb'
import { directionBadge, sourceTypeBadge } from '../data/constants'

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatUnixDate(unix: bigint | number | undefined) {
  const num = Number(unix || 0)
  if (!num) return '-'
  return new Date(num * 1000).toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function createTransactionColumns(
  accountMap: Map<string, CashAccount>,
  categoryMap: Map<string, CashCategory>
): ColumnDef<CashTransaction>[] {
  return [
    {
      accessorKey: 'transactionNo',
      header: 'No. Transaksi',
      cell: ({ row }) => {
        const tx = row.original
        return (
          <div className='flex flex-col'>
            <span className='font-mono text-xs font-semibold text-foreground'>
              {tx.transactionNo || tx.id}
            </span>
            <span className='text-[11px] text-muted-foreground'>
              {formatUnixDate(tx.trxDateUnix)}
            </span>
          </div>
        )
      },
    },
    {
      accessorKey: 'direction',
      header: 'Arah / Tipe',
      cell: ({ row }) => {
        const dir = row.original.direction
        const source = row.original.sourceType
        const meta = directionBadge(dir)
        const srcMeta = sourceTypeBadge(source)
        const isIncome = dir === 'IN'

        return (
          <div className='flex items-center gap-1.5'>
            <Badge variant='outline' className={`gap-1 text-[11px] font-medium ${meta.className}`}>
              {isIncome ? <ArrowDownLeft className='h-3 w-3' /> : <ArrowUpRight className='h-3 w-3' />}
              {meta.label}
            </Badge>
            {source && (
              <Badge variant='outline' className={`text-[10px] ${srcMeta.className}`}>
                {srcMeta.label}
              </Badge>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: 'accountId',
      header: 'Rekening Kas / Bank',
      cell: ({ row }) => {
        const acc = accountMap.get(row.original.accountId)
        if (!acc) return <span className='text-xs text-muted-foreground'>-</span>
        return (
          <div className='flex flex-col'>
            <span className='text-xs font-medium text-foreground'>{acc.name}</span>
            <span className='font-mono text-[10px] text-muted-foreground'>{acc.accountCode}</span>
          </div>
        )
      },
    },
    {
      accessorKey: 'categoryId',
      header: 'Kategori',
      cell: ({ row }) => {
        const cat = categoryMap.get(row.original.categoryId)
        if (!cat) return <span className='text-xs text-muted-foreground'>-</span>
        return <span className='text-xs text-foreground font-medium'>{cat.name}</span>
      },
    },
    {
      accessorKey: 'description',
      header: 'Deskripsi / Catatan',
      cell: ({ row }) => {
        const desc = row.original.description
        return (
          <span className='text-xs text-muted-foreground line-clamp-1 max-w-[280px]' title={desc}>
            {desc || '-'}
          </span>
        )
      },
    },
    {
      accessorKey: 'amount',
      header: () => <div className='text-right'>Nominal (IDR)</div>,
      cell: ({ row }) => {
        const tx = row.original
        const isIncome = tx.direction === 'IN'
        return (
          <div className='text-right'>
            <span
              className={`font-mono text-xs font-bold ${
                isIncome
                  ? 'text-emerald-600 dark:text-emerald-400'
                  : 'text-rose-600 dark:text-rose-400'
              }`}
            >
              {isIncome ? '+ ' : '- '}
              {formatCurrency(tx.amount)}
            </span>
          </div>
        )
      },
    },
  ]
}
