'use client'

import { DeleteUserRequest, type User } from '@/gen/v1/users_pb'
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
import { useDeleteUserMutation } from '../api/use-users'

type UsersDeleteDialogProps = {
  user?: User | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UsersDeleteDialog({
  user,
  open,
  onOpenChange,
}: UsersDeleteDialogProps) {
  const deleteMutation = useDeleteUserMutation()

  async function onDelete() {
    if (!user) return
    try {
      await deleteMutation.mutateAsync(new DeleteUserRequest({ id: user.id }))
      toast.success(`User ${user.username} deleted`)
      onOpenChange(false)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete user')
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={(v) => !v && onOpenChange(false)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. The account{' '}
            <strong>{user?.username}</strong> ({user?.email}) will be deleted
            permanently, along with its role assignments.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <Button
            variant='destructive'
            onClick={onDelete}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
