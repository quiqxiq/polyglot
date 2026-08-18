import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { PPPSecret } from '@/gen/v1/ppp_pb'
import { useDeviceStore } from '@/stores/device-store'
import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { CheckCircle2, Edit2, PowerOff, Trash2 } from 'lucide-react'
import { useSetPPPSecretDisabledMutation } from '../../api/use-ppp-secrets'
import { usePPP } from '../../context/ppp-context'

interface SecretsRowActionsProps {
  row: PPPSecret
}

export function SecretsRowActions({ row }: SecretsRowActionsProps) {
  const { setOpen, setCurrentSecret } = usePPP()
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const setDisabledMutation = useSetPPPSecretDisabledMutation()

  const handleEdit = () => {
    setCurrentSecret(row)
    setOpen('secret-update')
  }

  const handleDelete = () => {
    setCurrentSecret(row)
    setOpen('secret-delete')
  }

  const handleToggleDisabled = () => {
    if (!selectedDeviceId) return
    setDisabledMutation.mutate({
      deviceId: selectedDeviceId,
      rosId: row.id,
      disabled: !row.disabled,
    })
  }

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="flex h-8 w-8 p-0 data-[state=open]:bg-muted"
        >
          <DotsHorizontalIcon className="h-4 w-4" />
          <span className="sr-only">Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[160px]">
        <DropdownMenuItem onClick={handleEdit}>
          <Edit2 className="mr-2 h-4 w-4" />
          Edit Secret
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleToggleDisabled}>
          {row.disabled ? (
            <>
              <CheckCircle2 className="mr-2 h-4 w-4 text-emerald-500" />
              Enable Secret
            </>
          ) : (
            <>
              <PowerOff className="mr-2 h-4 w-4 text-amber-500" />
              Disable Secret
            </>
          )}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={handleDelete}
          className="text-destructive focus:text-destructive"
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
