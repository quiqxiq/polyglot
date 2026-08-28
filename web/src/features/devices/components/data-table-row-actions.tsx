'use client'

import type { Row } from '@tanstack/react-table'
import {
  MoreHorizontal,
  Edit,
  Trash2,
  Zap,
  BarChart2,
  Settings2,
  Terminal,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useDevicesContext } from './devices-provider'
import type { Device } from '@/gen/v1/device_pb'

interface DataTableRowActionsProps {
  row: Row<Device>
}

export function DataTableRowActions({ row }: DataTableRowActionsProps) {
  const { setOpen, setCurrentRow } = useDevicesContext()
  const device = row.original

  const handleSelectAction = (
    action:
      | 'terminal'
      | 'test'
      | 'edit'
      | 'delete'
      | 'ping-analytics'
      | 'ping-settings'
  ) => {
    setCurrentRow(device)
    setOpen(action)
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          className='flex h-8 w-8 p-0 data-[state=open]:bg-muted'
        >
          <MoreHorizontal className='h-4 w-4' />
          <span className='sr-only'>Buka menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[180px]'>
        <DropdownMenuItem onClick={() => handleSelectAction('ping-analytics')}>
          <BarChart2 className='me-2 h-3.5 w-3.5 text-blue-500' />
          Ping Analytics
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => handleSelectAction('ping-settings')}>
          <Settings2 className='me-2 h-3.5 w-3.5 text-indigo-500' />
          Ping Settings
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => handleSelectAction('terminal')}>
          <Terminal className='me-2 h-3.5 w-3.5 text-emerald-500' />
          SSH Terminal
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => handleSelectAction('test')}>
          <Zap className='me-2 h-3.5 w-3.5 text-amber-500' />
          Test Connection
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => handleSelectAction('edit')}>
          <Edit className='me-2 h-3.5 w-3.5' />
          Edit
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className='text-red-600 focus:text-red-600'
          onClick={() => handleSelectAction('delete')}
        >
          <Trash2 className='me-2 h-3.5 w-3.5' />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
