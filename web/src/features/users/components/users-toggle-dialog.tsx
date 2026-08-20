'use client'

import { toast } from 'sonner'
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { ToggleActiveRequest, type User } from '@/gen/v1/users_pb'
import { useToggleActiveMutation } from '../api/use-users'

type UsersToggleDialogProps = {
  user?: User | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UsersToggleDialog({
  user,
  open,
  onOpenChange,
}: UsersToggleDialogProps) {
  const toggleMutation = useToggleActiveMutation()
  const deactivating = !!user?.isActive

  async function onToggle() {
    if (!user) return
    try {
      await toggleMutation.mutateAsync(
        new ToggleActiveRequest({ id: user.id })
      )
      toast.success(
        deactivating
          ? `User ${user.username} deactivated`
          : `User ${user.username} activated`
      )
      onOpenChange(false)
    } catch (err: unknown) {
      toast.error(
        err instanceof Error ? err.message : 'Failed to update user status'
      )
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={(v) => !v && onOpenChange(false)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            {deactivating ? 'Deactivate account?' : 'Activate account?'}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {deactivating ? (
              <>
                <strong>{user?.username}</strong> will no longer be able to log
                in. The account data is kept — you can reactivate it later.
              </>
            ) : (
              <>
                <strong>{user?.username}</strong> will be able to log in again
                immediately.
              </>
            )}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <Button
            variant={deactivating ? 'destructive' : 'default'}
            onClick={onToggle}
            disabled={toggleMutation.isPending}
          >
            {toggleMutation.isPending
              ? 'Saving...'
              : deactivating
                ? 'Deactivate'
                : 'Activate'}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
