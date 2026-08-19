'use client'

import { useState, useEffect } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { roleLabel, roleClassName } from '@/features/users/data/roles'
import { useSyncRolePermissionsMutation } from '../api/use-rbac'
import { PermissionMatrix } from './permission-matrix'

interface EditRoleDialogProps {
  role: string | null
  initialPermissions: string[]
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function EditRoleDialog({
  role,
  initialPermissions,
  open,
  onOpenChange,
}: EditRoleDialogProps) {
  const [permissions, setPermissions] = useState<string[]>([])
  const syncMutation = useSyncRolePermissionsMutation()

  useEffect(() => {
    if (open && role) {
      setPermissions([...initialPermissions])
    }
  }, [open, role, initialPermissions])

  if (!role) return null

  const isOwner = role === 'owner'

  const handleSave = async () => {
    if (isOwner) {
      toast.error('Owner role permissions cannot be modified')
      return
    }

    try {
      await syncMutation.mutateAsync({
        role,
        permissions,
      })
      toast.success(`Permissions updated for role "${role}"!`)
      onOpenChange(false)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to update permissions')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-3xl max-h-[90vh] flex flex-col'>
        <DialogHeader className='text-start'>
          <div className='flex items-center gap-2'>
            <DialogTitle>Edit Role Permissions</DialogTitle>
            <Badge variant='outline' className={roleClassName(role)}>
              {roleLabel(role)}
            </Badge>
          </div>
          <DialogDescription>
            {isOwner
              ? 'Owner role possesses absolute access to all system resources.'
              : 'Modify granted resource permissions for this role using the matrix.'}
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4 py-2 flex-1 overflow-hidden flex flex-col'>
          <div className='flex-1 min-h-0 flex flex-col'>
            <PermissionMatrix
              selectedPermissions={permissions}
              onChange={setPermissions}
              disabled={isOwner || syncMutation.isPending}
            />
          </div>
        </div>

        <DialogFooter className='border-t pt-3'>
          <Button
            type='button'
            variant='outline'
            size='sm'
            onClick={() => onOpenChange(false)}
            disabled={syncMutation.isPending}
          >
            Cancel
          </Button>
          {!isOwner && (
            <Button
              type='button'
              size='sm'
              onClick={handleSave}
              disabled={syncMutation.isPending}
            >
              {syncMutation.isPending ? 'Saving...' : 'Save Permissions'}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
