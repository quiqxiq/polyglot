import {
  CalendarClock,
  CheckCircle2,
  Eye,
  MoreHorizontal,
  Router,
  XCircle,
} from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import type { Registration } from '@/gen/v1/registration_pb'
import { REGISTRATION_STATUS } from '../data/constants'
import { useRegistration } from './registration-provider'

interface RegistrationRowActionsProps {
  registration: Registration
}

export function RegistrationRowActions({ registration }: RegistrationRowActionsProps) {
  const { setOpen, setCurrentRow } = useRegistration()

  const handleAction = (dialog: 'schedule' | 'install' | 'convert' | 'reject' | 'cancel' | 'detail') => {
    setCurrentRow(registration)
    setOpen(dialog)
  }

  const isPending = registration.status === REGISTRATION_STATUS.PENDING
  const isApproved = registration.status === REGISTRATION_STATUS.APPROVED
  const isInstalled = registration.status === REGISTRATION_STATUS.INSTALLED

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='ghost' className='h-8 w-8 p-0'>
          <span className='sr-only'>Open menu</span>
          <MoreHorizontal className='h-4 w-4' />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-56'>
        <DropdownMenuItem onClick={() => handleAction('detail')}>
          <Eye className='mr-2 h-4 w-4' />
          Detail Pendaftaran
        </DropdownMenuItem>

        {isPending && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => handleAction('schedule')}
              className='text-blue-600 focus:text-blue-600 dark:text-blue-400'
            >
              <CalendarClock className='mr-2 h-4 w-4' />
              Setujui & Jadwalkan
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => handleAction('reject')}
              className='text-rose-600 focus:text-rose-600'
            >
              <XCircle className='mr-2 h-4 w-4' />
              Tolak Pendaftaran
            </DropdownMenuItem>
          </>
        )}

        {isApproved && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => handleAction('install')}
              className='text-purple-600 focus:text-purple-600 dark:text-purple-400 font-medium'
            >
              <Router className='mr-2 h-4 w-4' />
              Selesai Pasang (Pilih Router)
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleAction('schedule')}>
              <CalendarClock className='mr-2 h-4 w-4' />
              Ubah Jadwal Pasang
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => handleAction('cancel')}
              className='text-muted-foreground'
            >
              <XCircle className='mr-2 h-4 w-4' />
              Batalkan
            </DropdownMenuItem>
          </>
        )}

        {isInstalled && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => handleAction('convert')}
              className='text-emerald-600 focus:text-emerald-600 dark:text-emerald-400 font-semibold'
            >
              <CheckCircle2 className='mr-2 h-4 w-4' />
              Aktivasi & Terbitkan Tagihan
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => handleAction('install')}>
              <Router className='mr-2 h-4 w-4' />
              Ubah Router Terpasang
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
