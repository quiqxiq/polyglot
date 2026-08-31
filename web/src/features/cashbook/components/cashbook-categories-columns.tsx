import { type ColumnDef } from '@tanstack/react-table'
import { ArrowDownLeft, ArrowUpRight, Edit } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { type CashCategory } from '@/gen/v1/cashbook_pb'
import { useCashbook } from './cashbook-provider'

export function useCategoryColumns(): ColumnDef<CashCategory>[] {
  const { setOpen, setCurrentCategory } = useCashbook()

  return [
    {
      accessorKey: 'name',
      header: 'Nama Kategori',
      cell: ({ row }) => (
        <span className='text-xs font-semibold text-foreground'>{row.original.name}</span>
      ),
    },
    {
      accessorKey: 'type',
      header: 'Tipe Arus Kas',
      cell: ({ row }) => {
        const isIncome = row.original.type === 'INCOME'
        return (
          <Badge
            variant='outline'
            className={`gap-1 text-[11px] ${
              isIncome
                ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/30'
                : 'bg-rose-500/10 text-rose-700 dark:text-rose-400 border-rose-500/30'
            }`}
          >
            {isIncome ? <ArrowDownLeft className='h-3 w-3' /> : <ArrowUpRight className='h-3 w-3' />}
            {isIncome ? 'Pendapatan (INCOME)' : 'Pengeluaran (EXPENSE)'}
          </Badge>
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
              setCurrentCategory(row.original)
              setOpen('edit-category')
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
