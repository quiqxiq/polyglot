import { Download, Plus, Upload } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ExportCustomersRequest } from '@/gen/v1/ispadmin_pb'
import { useExportCustomersMutation } from '../api/use-customer'
import { useAuthStore } from '@/stores/auth-store'
import { canPermission } from '@/hooks/use-can'
import { useCustomers } from './customers-provider'

export function CustomersPrimaryButtons() {
  const { setOpen } = useCustomers()
  const permissions = useAuthStore((s) => s.auth.user?.permissions)
  const canManage = canPermission(permissions, 'customer:manage')
  const exportCustomers = useExportCustomersMutation()

  if (!canManage) return null

  const handleExport = async (format: number) => {
    try {
      const res = await exportCustomers.mutateAsync(
        new ExportCustomersRequest({ format })
      )
      const blob = new Blob([res.payload as BlobPart], {
        type: res.contentType,
      })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = res.filename || 'customers.csv'
      document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Export gagal')
    }
  }

  return (
    <div className='flex gap-2'>
      <Button
        variant='outline'
        className='space-x-1'
        onClick={() => setOpen('import')}
      >
        <span>Import</span> <Upload size={18} />
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant='outline' className='space-x-1'>
            <span>Export</span> <Download size={18} />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end'>
          <DropdownMenuItem onClick={() => void handleExport(0)}>
            CSV
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => void handleExport(1)}>
            Excel (XLSX)
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <Button className='space-x-1' onClick={() => setOpen('create')}>
        <span>Tambah Pelanggan</span> <Plus size={18} />
      </Button>
    </div>
  )
}
