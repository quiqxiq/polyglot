'use client'

import { Row } from '@tanstack/react-table'
import { MoreHorizontal, Edit, Trash2, Zap } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useDevicesContext } from './devices-provider'
import { Device } from '@/gen/v1/device_pb'

interface DataTableRowActionsProps {
  row: Row<Device>
}

export function DataTableRowActions({ row }: DataTableRowActionsProps) {
  const { setOpen, setCurrentRow } = useDevicesContext()
  const device = row.original

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' className='flex h-8 w-8 p-0 data-[state=open]:bg-muted'>
          <MoreHorizontal className='h-4 w-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[160px]'>
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(device)
            setOpen('test')
          }}
        >
          <Zap className='me-2 h-3.5 w-3.5 text-amber-500' />
          Test Connection
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(device)
            setOpen('edit')
          }}
        >
          <Edit className='me-2 h-3.5 w-3.5' />
          Edit
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className='text-red-600 focus:text-red-600'
          onClick={() => {
            setCurrentRow(device)
            setOpen('delete')
          }}
        >
          <Trash2 className='me-2 h-3.5 w-3.5' />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
