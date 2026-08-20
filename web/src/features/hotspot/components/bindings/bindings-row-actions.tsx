import { type Row } from '@tanstack/react-table'
import { MoreHorizontal, Edit, Trash2, ShieldCheck, ShieldAlert } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useHotspot } from '../../context/hotspot-context'
import { useUpdateHotspotIPBindingMutation } from '../../api/use-hotspot-bindings'
import { useDeviceStore } from '@/stores/device-store'
import type { HotspotIPBinding } from '@/gen/v1/hotspot_pb'

type BindingsRowActionsProps = {
  row: Row<HotspotIPBinding>
}

export function BindingsRowActions({ row }: BindingsRowActionsProps) {
  const binding = row.original
  const { setOpen, setCurrentBinding } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()
  const updateMutation = useUpdateHotspotIPBindingMutation()

  const handleToggleDisabled = async () => {
    if (!selectedDeviceId) return

    toast.promise(
      updateMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: binding.id,
        macAddress: binding.macAddress,
        address: binding.address,
        toAddress: binding.toAddress,
        server: binding.server,
        type: binding.type,
        comment: binding.comment,
        disabled: !binding.disabled,
      }),
      {
        loading: `${binding.disabled ? 'Enabling' : 'Disabling'} IP binding...`,
        success: `IP binding ${binding.disabled ? 'enabled' : 'disabled'}.`,
        error: 'Failed to update IP binding.',
      }
    )
  }

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          className='flex size-8 p-0 data-[state=open]:bg-muted'
        >
          <MoreHorizontal className='size-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[160px]'>
        <DropdownMenuItem
          onClick={() => {
            setCurrentBinding(binding)
            setOpen('binding-update')
          }}
          className='gap-2'
        >
          <Edit className='size-4' />
          Edit
        </DropdownMenuItem>

        <DropdownMenuItem onClick={handleToggleDisabled} className='gap-2'>
          {binding.disabled ? (
            <>
              <ShieldCheck className='size-4 text-emerald-500' />
              Enable
            </>
          ) : (
            <>
              <ShieldAlert className='size-4 text-amber-500' />
              Disable
            </>
          )}
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem
          onClick={() => {
            setCurrentBinding(binding)
            setOpen('binding-delete')
          }}
          className='gap-2 text-destructive focus:text-destructive'
        >
          <Trash2 className='size-4' />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
