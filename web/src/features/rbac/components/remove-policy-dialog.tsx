'use client'

import { type Policy } from '@/gen/v1/rbac_pb'
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
import { roleLabel } from '@/features/users/data/roles'
import { useRemovePolicyMutation } from '../api/use-rbac'

type RemovePolicyDialogProps = {
  policy: Policy | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function RemovePolicyDialog({
  policy,
  open,
  onOpenChange,
}: RemovePolicyDialogProps) {
  const removeMutation = useRemovePolicyMutation()

  async function onRemove() {
    if (!policy) return
    try {
      const res = await removeMutation.mutateAsync(policy)
      if (!res.success) {
        toast.error('Policy not found')
        return
      }
      toast.success(`Policy removed: ${roleLabel(policy.sub)} → ${policy.obj}`)
      onOpenChange(false)
    } catch (err: unknown) {
      toast.error(
        err instanceof Error ? err.message : 'Failed to remove policy'
      )
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={(v) => !v && onOpenChange(false)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Remove policy?</AlertDialogTitle>
          <AlertDialogDescription>
            Remove access for{' '}
            <strong>{policy ? roleLabel(policy.sub) : ''}</strong> to{' '}
            <code className='rounded bg-muted px-1.5 py-0.5 text-xs'>
              {policy?.obj}
            </code>
            ? Roles lose this permission immediately.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <Button
            variant='destructive'
            onClick={onRemove}
            disabled={removeMutation.isPending}
          >
            {removeMutation.isPending ? 'Removing...' : 'Remove'}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
