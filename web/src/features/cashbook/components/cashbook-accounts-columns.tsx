import { type ColumnDef } from '@tanstack/react-table'
import { Building2, Edit, Landmark } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { type CashAccount } from '@/gen/v1/cashbook_pb'
import { useCashbook } from './cashbook-provider'

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(amount)
}

export function useAccountColumns(balances: Record<string, number>): ColumnDef<CashAccount>[] {
  const { setOpen, setCurrentAccount } = useCashbook()

  return [
    {
      accessorKey: 'accountCode',
      header: 'Kode Akun',
      cell: ({ row }) => (
        <span className='font-mono text-xs font-semibold text-foreground'>
          {row.original.accountCode}
        </span>
      ),
    },
    {
      accessorKey: 'name',
      header: 'Nama Rekening Kas',
      cell: ({ row }) => (
        <span className='text-xs font-semibold text-foreground'>{row.original.name}</span>
      ),
    },
    {
      accessorKey: 'type',
      header: 'Tipe Rekening',
      cell: ({ row }) => {
        const isBank = row.original.type === 'BANK'
        return (
          <Badge
            variant='outline'
            className={`gap-1 text-[11px] ${
              isBank
                ? 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/30'
                : 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/30'
            }`}
          >
            {isBank ? <Building2 className='h-3 w-3' /> : <Landmark className='h-3 w-3' />}
            {isBank ? 'Rekening Bank' : 'Kas Fisik / Kasir'}
          </Badge>
        )
      },
    },
    {
      id: 'balance',
      header: () => <div className='text-right'>Saldo Berjalan (IDR)</div>,
      cell: ({ row }) => {
        const bal = balances[row.original.id] || 0
        return (
          <div className='text-right'>
            <span
              className={`font-mono text-xs font-bold ${
                bal >= 0 ? 'text-foreground' : 'text-rose-600 dark:text-rose-400'
              }`}
            >
              {formatCurrency(bal)}
            </span>
          </div>
        )
      },
    },
    {
      accessorKey: 'isActive',
      header: 'Status',
      cell: ({ row }) => (
        <Badge
          variant='outline'
          className={`text-[11px] ${
            row.original.isActive
              ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30'
              : 'bg-slate-500/15 text-slate-700 dark:text-slate-400 border-slate-500/30'
          }`}
        >
          {row.original.isActive ? 'Aktif' : 'Non-aktif'}
        </Badge>
      ),
    },
    {
      id: 'actions',
      header: () => <div className='text-right'>Aksi</div>,
      cell: ({ row }) => (
        <div className='text-right'>
          <Button
            variant='ghost'
            size='sm'
            onClick={() => {
              setCurrentAccount(row.original)
              setOpen('edit-account')
            }}
            className='h-8 px-2 text-xs'
          >
            <Edit className='h-3.5 w-3.5 mr-1 text-muted-foreground' />
            Ubah
          </Button>
        </div>
      ),
    },
  ]
}
