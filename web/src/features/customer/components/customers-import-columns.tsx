import type { ColumnDef } from '@tanstack/react-table'
import type { CustomerSubscriptionImportRow } from '@/gen/v1/ispadmin_pb'
import type { Plan } from '@/gen/v1/plan_pb'
import { CheckCircle2, Sparkles } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { formatAutoName } from '@/lib/format-name'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { DataTableColumnHeader } from '@/components/data-table'

interface GetCustomersImportColumnsProps {
  plans: Plan[]
  showPasswords: boolean
  onUpdateRow: (
    index: number,
    updated: Partial<CustomerSubscriptionImportRow>
  ) => void
}

export function getCustomersImportColumns({
  plans,
  showPasswords,
  onUpdateRow,
}: GetCustomersImportColumnsProps): ColumnDef<CustomerSubscriptionImportRow>[] {
  return [
    {
      id: 'select',
      header: ({ table }) => (
        <div className='flex justify-center'>
          <Checkbox
            checked={
              table.getIsAllPageRowsSelected() ||
              (table.getIsSomePageRowsSelected() && 'indeterminate')
            }
            onCheckedChange={(value) =>
              table.toggleAllPageRowsSelected(!!value)
            }
            aria-label='Pilih semua di halaman ini'
          />
        </div>
      ),
      cell: ({ row }) => (
        <div className='flex justify-center'>
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label={`Pilih akun ${row.original.username}`}
          />
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'username',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Kredensial Router' />
      ),
      cell: ({ row }) => {
        const r = row.original
        return (
          <div className='space-y-1'>
            <div className='flex items-center gap-1.5'>
              {r.hotspotType === 'IP_BINDING' ? (
                <Badge
                  variant='secondary'
                  className='bg-purple-500/15 px-1.5 py-0 text-[10px] text-purple-700 dark:text-purple-400'
                >
                  IP BINDING
                </Badge>
              ) : r.serviceType === 'PPPOE' ? (
                <Badge
                  variant='secondary'
                  className='bg-sky-500/15 px-1.5 py-0 text-[10px] text-sky-700 dark:text-sky-400'
                >
                  PPPOE
                </Badge>
              ) : (
                <Badge
                  variant='secondary'
                  className='bg-amber-500/15 px-1.5 py-0 text-[10px] text-amber-700 dark:text-amber-400'
                >
                  HOTSPOT
                </Badge>
              )}
              <span
                className='font-mono text-xs font-semibold'
                title='Username Login Router'
              >
                {r.username}
              </span>
            </div>

            {/* Password / MAC info */}
            <div className='font-mono text-[11px] text-muted-foreground'>
              {r.macAddress ? (
                <span>MAC: {r.macAddress}</span>
              ) : showPasswords ? (
                <span>Pass: {r.password || '(empty)'}</span>
              ) : (
                <span>Pass: ••••••</span>
              )}
            </div>
          </div>
        )
      },
    },
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Nama Pelanggan (Wajib)' />
      ),
      cell: ({ row }) => {
        const canFormat =
          row.original.username &&
          (row.original.username.includes('_') ||
            row.original.username.includes('-'))
        return (
          <div className='flex min-w-[170px] items-center gap-1'>
            <Input
              className={`h-8 flex-1 text-xs font-medium ${
                !row.original.name.trim() ? 'border-red-400 bg-red-500/5' : ''
              }`}
              placeholder='Nama Lengkap Pelanggan...'
              value={row.original.name}
              onChange={(e) => onUpdateRow(row.index, { name: e.target.value })}
            />
            {canFormat && (
              <Button
                variant='ghost'
                size='icon'
                className='h-7 w-7 shrink-0 text-muted-foreground hover:text-amber-500'
                title={`Generate nama: ${formatAutoName(row.original.username)}`}
                onClick={() =>
                  onUpdateRow(row.index, {
                    name: formatAutoName(row.original.username),
                  })
                }
              >
                <Sparkles className='h-3.5 w-3.5 text-amber-500' />
              </Button>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: 'phone',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='No. WhatsApp / HP' />
      ),
      cell: ({ row }) => (
        <Input
          className={`h-8 min-w-[120px] text-xs ${
            !row.original.phone ? 'border-amber-400/80 bg-amber-500/5' : ''
          }`}
          placeholder='+628...'
          value={row.original.phone}
          onChange={(e) => onUpdateRow(row.index, { phone: e.target.value })}
        />
      ),
    },
    {
      accessorKey: 'address',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Alamat Lengkap' />
      ),
      cell: ({ row }) => (
        <Input
          className={`h-8 min-w-[160px] text-xs ${
            !row.original.address ? 'border-amber-400/80 bg-amber-500/5' : ''
          }`}
          placeholder='Alamat lengkap...'
          value={row.original.address}
          onChange={(e) =>
            onUpdateRow(row.index, {
              address: e.target.value,
            })
          }
        />
      ),
    },
    {
      accessorKey: 'planName',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Paket Layanan' />
      ),
      cell: ({ row }) => {
        const r = row.original
        const activePlan = plans.find(
          (p) =>
            (r.planId && p.id === r.planId) ||
            (r.planName && p.name.toLowerCase() === r.planName.toLowerCase())
        )
        const hasPlanWarning = !activePlan && !r.planName

        const handlePlanChange = (planId: string) => {
          const selected = plans.find((p) => p.id === planId)
          if (selected) {
            onUpdateRow(row.index, {
              planId: selected.id,
              planName: selected.name,
              price: selected.price,
            })
          }
        }

        return (
          <div className='flex min-w-[170px] flex-col gap-1'>
            <Select
              value={activePlan?.id || r.planId || undefined}
              onValueChange={handlePlanChange}
            >
              <SelectTrigger
                className={`h-8 text-xs ${
                  hasPlanWarning ? 'border-amber-400 bg-amber-500/5' : ''
                }`}
              >
                <SelectValue placeholder={r.planName || 'Pilih Paket'} />
              </SelectTrigger>
              <SelectContent className='max-h-56'>
                {plans.map((p) => (
                  <SelectItem key={p.id} value={p.id} className='text-xs'>
                    <span className='font-medium'>{p.name}</span>
                    <span className='ml-2 font-mono text-muted-foreground'>
                      (Rp {p.price.toLocaleString('id-ID')})
                    </span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {(r.routerProfile || r.planName) && (
              <span className='font-mono text-[10px] text-muted-foreground'>
                Profil:{' '}
                <span className='font-semibold text-foreground/80'>
                  {r.routerProfile || r.planName}
                </span>
              </span>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: 'price',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Harga Tagihan (Rp)' />
      ),
      cell: ({ row }) => (
        <div className='relative min-w-[120px]'>
          <span className='absolute top-2 left-2.5 text-xs text-muted-foreground'>
            Rp
          </span>
          <Input
            type='number'
            className='h-8 pl-8 text-xs font-semibold'
            value={row.original.price || ''}
            placeholder='0'
            onChange={(e) =>
              onUpdateRow(row.index, {
                price: parseFloat(e.target.value) || 0,
              })
            }
          />
        </div>
      ),
    },
    {
      accessorKey: 'billingDay',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Jatuh Tempo' />
      ),
      cell: ({ row }) => (
        <Select
          value={String(row.original.billingDay || 1)}
          onValueChange={(v) =>
            onUpdateRow(row.index, {
              billingDay: parseInt(v, 10) || 1,
            })
          }
        >
          <SelectTrigger className='h-8 w-20 text-center text-xs'>
            <SelectValue />
          </SelectTrigger>
          <SelectContent className='max-h-48'>
            {Array.from({ length: 31 }, (_, i) => i + 1).map((day) => (
              <SelectItem key={day} value={String(day)} className='text-xs'>
                Tgl {day}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      ),
    },
    {
      id: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Status' />
      ),
      filterFn: (row, _columnId, filterValue) => {
        if (filterValue === 'warning') {
          return !row.original.phone || !row.original.address || !row.original.name
        }
        return true
      },
      cell: ({ row }) => {
        const hasPhoneWarning = !row.original.phone
        return hasPhoneWarning ? (
          <Badge
            variant='outline'
            className='border-amber-500/40 bg-amber-500/10 text-[10px] whitespace-nowrap text-amber-600'
          >
            Isi No HP
          </Badge>
        ) : (
          <Badge
            variant='outline'
            className='border-emerald-500/40 bg-emerald-500/10 text-[10px] whitespace-nowrap text-emerald-600'
          >
            <CheckCircle2 className='mr-1 h-3 w-3' />
            Siap Simpan
          </Badge>
        )
      },
    },
  ]
}
