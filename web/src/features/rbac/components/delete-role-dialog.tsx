'use client'

import { toast } from 'sonner'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { useDeleteRoleMutation } from '../api/use-rbac'

interface DeleteRoleDialogProps {
  role: string | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function DeleteRoleDialog({
  role,
  open,
  onOpenChange,
}: DeleteRoleDialogProps) {
  const deleteMutation = useDeleteRoleMutation()

  if (!role) return null

  const handleDelete = async () => {
    try {
      await deleteMutation.mutateAsync(role)
      toast.success(`Role "${role}" deleted successfully`)
      onOpenChange(false)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete role')
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Role &quot;{role}&quot;?</AlertDialogTitle>
          <AlertDialogDescription>
            This action will remove all permission rules and user assignments for
            the role <strong className='text-foreground'>{role}</strong>. Users
            currently assigned to this role will lose these permissions immediately.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={deleteMutation.isPending}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
            className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
          >
            {deleteMutation.isPending ? 'Deleting...' : 'Delete Role'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
