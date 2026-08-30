import { Eye, MoreHorizontal, Pencil, Repeat, Trash2 } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import type { Customer } from '@/gen/v1/customer_pb'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { useCustomers } from './customers-provider'

interface CustomersRowActionsProps {
  customer: Customer
}

export function CustomersRowActions({ customer }: CustomersRowActionsProps) {
  const { setOpen, setCurrentRow } = useCustomers()
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canManage = canPermission(permissions, 'customer:manage')
  const canBilling = canPermission(permissions, 'billing:manage')

  if (!canManage) return null

  const handleDetail = () => {
    setCurrentRow(customer)
    setOpen('detail')
  }

  const handleEdit = () => {
    setCurrentRow(customer)
    setOpen('update')
  }

  const handleCreateSubscription = () => {
    setCurrentRow(customer)
    setOpen('create-subscription')
  }

  const handleDelete = () => {
    setCurrentRow(customer)
    setOpen('delete')
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' className='h-8 w-8 p-0'>
          <span className='sr-only'>Open menu</span>
          <MoreHorizontal className='h-4 w-4' />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-48'>
        <DropdownMenuItem onClick={handleDetail}>
          <Eye className='mr-2 h-4 w-4' />
          Detail Pelanggan
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleEdit}>
          <Pencil className='mr-2 h-4 w-4' />
          Edit
        </DropdownMenuItem>
        {canBilling && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleCreateSubscription}>
              <Repeat className='mr-2 h-4 w-4' />
              Buat Langganan
            </DropdownMenuItem>
          </>
        )}
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleDelete} className='text-destructive focus:text-destructive'>
          <Trash2 className='mr-2 h-4 w-4' />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
