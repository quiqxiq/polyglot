import { useState } from 'react'
import type { Plan } from '@/gen/v1/plan_pb'
import {
  Layers,
  Calendar,
  MapPin,
  CheckSquare,
  Square,
  ChevronDown,
  Sparkles,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'

interface CustomersImportBatchBarProps {
  plans: Plan[]
  selectedCount: number
  totalCount: number
  allSelected: boolean
  onApplyPlan: (planId: string, planName: string, price: number) => void
  onApplyBillingDay: (day: number) => void
  onApplyDefaultAddress: (address: string) => void
  onAutoGenerateNames: (mode: 'name_only' | 'split_address') => void
  onToggleSelectAll: (select: boolean) => void
}

export function CustomersImportBatchBar({
  plans,
  selectedCount,
  totalCount,
  allSelected,
  onApplyPlan,
  onApplyBillingDay,
  onApplyDefaultAddress,
  onAutoGenerateNames,
  onToggleSelectAll,
}: CustomersImportBatchBarProps) {
  const [addressInput, setAddressInput] = useState('')
  const [isAddressOpen, setIsAddressOpen] = useState(false)

  const handleAddressSubmit = () => {
    if (addressInput.trim()) {
      onApplyDefaultAddress(addressInput.trim())
      setAddressInput('')
      setIsAddressOpen(false)
    }
  }

  return (
    <div className='flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-card p-3 shadow-xs'>
      {/* Selection controls */}
      <div className='flex items-center gap-2'>
        <Button
          variant='ghost'
          size='sm'
          onClick={() => onToggleSelectAll(!allSelected)}
          className='h-8 text-xs font-medium'
        >
          {allSelected ? (
            <CheckSquare className='mr-1.5 h-4 w-4 text-primary' />
          ) : (
            <Square className='mr-1.5 h-4 w-4 text-muted-foreground' />
          )}
          {allSelected ? 'Batal Pilih Semua' : 'Pilih Semua'}
        </Button>
        <span className='text-xs text-muted-foreground'>
          <strong className='text-foreground'>{selectedCount}</strong> dari{' '}
          {totalCount} akun dipilih
        </span>
      </div>

      {/* Batch Operations */}
      <div className='flex flex-wrap items-center gap-2'>
        {/* Batch Auto Generate Names */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant='outline'
              size='sm'
              disabled={totalCount === 0}
              className='h-8 text-xs'
            >
              <Sparkles className='mr-1.5 h-3.5 w-3.5 text-amber-500' />
              Auto Generate Nama
              <ChevronDown className='ml-1 h-3 w-3 opacity-60' />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end' className='w-64'>
            <DropdownMenuItem
              onClick={() => onAutoGenerateNames('name_only')}
              className='cursor-pointer flex-col items-start gap-0.5 text-xs'
            >
              <span className='font-medium'>Format Nama Penuh</span>
              <span className='text-[10px] text-muted-foreground'>
                Ubah &quot;susanto_madek&quot; &rarr; &quot;Susanto Madek&quot;
              </span>
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => onAutoGenerateNames('split_address')}
              className='cursor-pointer flex-col items-start gap-0.5 text-xs'
            >
              <span className='font-medium'>Pisahkan Nama &amp; Alamat</span>
              <span className='text-[10px] text-muted-foreground'>
                Nama: &quot;Susanto&quot;, Alamat: &quot;Madek&quot;
              </span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Batch Assign Plan */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant='outline'
              size='sm'
              disabled={selectedCount === 0}
              className='h-8 text-xs'
            >
              <Layers className='mr-1.5 h-3.5 w-3.5 text-sky-500' />
              Set Paket Terpilih
              <ChevronDown className='ml-1 h-3 w-3 opacity-60' />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            align='end'
            className='max-h-60 w-56 overflow-y-auto'
          >
            {plans.length === 0 ? (
              <div className='p-2 text-xs text-muted-foreground'>
                Belum ada paket terdaftar.
              </div>
            ) : (
              plans.map((p) => (
                <DropdownMenuItem
                  key={p.id}
                  onClick={() => onApplyPlan(p.id, p.name, p.price)}
                  className='flex justify-between text-xs'
                >
                  <span className='font-medium'>{p.name}</span>
                  <span className='font-mono text-muted-foreground'>
                    Rp {p.price.toLocaleString('id-ID')}
                  </span>
                </DropdownMenuItem>
              ))
            )}
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Batch Set Billing Day */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant='outline'
              size='sm'
              disabled={selectedCount === 0}
              className='h-8 text-xs'
            >
              <Calendar className='mr-1.5 h-3.5 w-3.5 text-amber-500' />
              Set Tgl Jatuh Tempo
              <ChevronDown className='ml-1 h-3 w-3 opacity-60' />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            align='end'
            className='max-h-60 w-44 overflow-y-auto'
          >
            {[1, 5, 10, 15, 20, 25].map((day) => (
              <DropdownMenuItem
                key={day}
                onClick={() => onApplyBillingDay(day)}
                className='text-xs'
              >
                Tanggal {day} setiap bulan
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Batch Set Address */}
        <div className='relative flex items-center'>
          {isAddressOpen ? (
            <div className='flex items-center gap-1'>
              <Input
                placeholder='Masukkan alamat...'
                className='h-8 w-48 text-xs'
                value={addressInput}
                onChange={(e) => setAddressInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleAddressSubmit()}
                autoFocus
              />
              <Button
                variant='secondary'
                size='sm'
                onClick={handleAddressSubmit}
                className='h-8 text-xs'
              >
                Terapkan
              </Button>
              <Button
                variant='ghost'
                size='sm'
                onClick={() => setIsAddressOpen(false)}
                className='h-8 px-2 text-xs'
              >
                Batal
              </Button>
            </div>
          ) : (
            <Button
              variant='outline'
              size='sm'
              disabled={selectedCount === 0}
              onClick={() => setIsAddressOpen(true)}
              className='h-8 text-xs'
            >
              <MapPin className='mr-1.5 h-3.5 w-3.5 text-rose-500' />
              Set Alamat Massal
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
