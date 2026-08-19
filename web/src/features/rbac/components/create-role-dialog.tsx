'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { PlusIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useSyncRolePermissionsMutation } from '../api/use-rbac'
import { PermissionMatrix } from './permission-matrix'

interface CreateRoleDialogProps {
  existingRoles: string[]
}

export function CreateRoleDialog({ existingRoles }: CreateRoleDialogProps) {
  const [open, setOpen] = useState(false)
  const [roleName, setRoleName] = useState('')
  const [permissions, setPermissions] = useState<string[]>([])
  const syncMutation = useSyncRolePermissionsMutation()

  const handleCreate = async () => {
    const slug = roleName.trim().toLowerCase().replace(/\s+/g, '_')
    if (!slug) {
      toast.error('Role name cannot be empty')
      return
    }
    if (existingRoles.includes(slug)) {
      toast.error(`Role "${slug}" already exists`)
      return
    }
    if (permissions.length === 0) {
      toast.error('Please select at least one permission')
      return
    }

    try {
      await syncMutation.mutateAsync({
        role: slug,
        permissions,
      })
      toast.success(`Role "${slug}" created with ${permissions.length} permissions!`)
      setOpen(false)
      setRoleName('')
      setPermissions([])
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to create role')
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button className='h-9 gap-1.5'>
          <PlusIcon className='h-4 w-4' /> Create Role
        </Button>
      </DialogTrigger>
      <DialogContent className='sm:max-w-3xl max-h-[90vh] flex flex-col'>
        <DialogHeader className='text-start'>
          <DialogTitle>Create New Role</DialogTitle>
          <DialogDescription>
            Define a new role and grant module permissions via the matrix below.
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4 py-2 flex-1 overflow-hidden flex flex-col'>
          <div className='space-y-1.5'>
            <Label htmlFor='role-name' className='text-xs font-semibold'>
              Role Name / Identifier
            </Label>
            <Input
              id='role-name'
              placeholder='e.g. billing_staff, noc_engineer, support_lead'
              value={roleName}
              onChange={(e) => setRoleName(e.target.value)}
              className='h-8 text-xs font-mono max-w-sm'
            />
          </div>

          <div className='space-y-1.5 flex-1 min-h-0 flex flex-col'>
            <Label className='text-xs font-semibold'>Permission Matrix</Label>
            <div className='flex-1 overflow-hidden'>
              <PermissionMatrix
                selectedPermissions={permissions}
                onChange={setPermissions}
                disabled={syncMutation.isPending}
              />
            </div>
          </div>
        </div>

        <DialogFooter className='border-t pt-3'>
          <Button
            type='button'
            variant='outline'
            size='sm'
            onClick={() => setOpen(false)}
            disabled={syncMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            type='button'
            size='sm'
            onClick={handleCreate}
            disabled={syncMutation.isPending || !roleName.trim()}
          >
            {syncMutation.isPending ? 'Creating...' : 'Create Role'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
