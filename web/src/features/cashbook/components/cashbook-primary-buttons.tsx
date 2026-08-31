import { Plus, Wallet, Tag } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useCashbook } from './cashbook-provider'

export function CashbookPrimaryButtons() {
  const { setOpen, setCurrentAccount, setCurrentCategory } = useCashbook()

  return (
    <div className='flex items-center gap-2'>
      {/* Menu Tambah Cepat */}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant='outline' size='sm' className='gap-1.5 h-9'>
            <Plus className='h-4 w-4' />
            Master Data
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end' className='w-48'>
          <DropdownMenuItem
            onClick={() => {
              setCurrentAccount(null)
              setOpen('create-account')
            }}
            className='gap-2 cursor-pointer'
          >
            <Wallet className='h-4 w-4 text-emerald-600' />
            Tambah Rekening Kas
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => {
              setCurrentCategory(null)
              setOpen('create-category')
            }}
            className='gap-2 cursor-pointer'
          >
            <Tag className='h-4 w-4 text-blue-600' />
            Tambah Kategori Kas
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Tombol Utama Catat Transaksi */}
      <Button
        size='sm'
        onClick={() => setOpen('create-transaction')}
        className='gap-1.5 h-9'
      >
        <Plus className='h-4 w-4' />
        Catat Transaksi Kas
      </Button>
    </div>
  )
}
