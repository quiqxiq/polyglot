'use client'

import { useState } from 'react'
import { RoleAssignment } from '@/gen/v1/rbac_pb'
import { type User } from '@/gen/v1/users_pb'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import {
  roleClassName,
  roleLabel,
  ROLE_META,
} from '@/features/users/data/roles'
import { useAssignRoleMutation, useUnassignRoleMutation } from '../api/use-rbac'

type AssignmentDialogProps = {
  user: User | null
  currentRoles: string[]
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AssignmentDialog({
  user,
  currentRoles,
  open,
  onOpenChange,
}: AssignmentDialogProps) {
  const [selected, setSelected] = useState<string[]>([])
  const assignMutation = useAssignRoleMutation()
  const unassignMutation = useUnassignRoleMutation()

  // Sinkronkan checkbox saat dialog dibuka / user berubah.
  const [lastKey, setLastKey] = useState('')
  const key = `${user?.id}-${open}`
  if (key !== lastKey) {
    setLastKey(key)
    setSelected([...(currentRoles ?? [])])
  }

  const primaryRole = user?.role ?? ''

  function toggle(role: string, checked: boolean) {
    setSelected((prev) =>
      checked ? [...new Set([...prev, role])] : prev.filter((r) => r !== role)
    )
  }

  async function onSave() {
    if (!user) return
    try {
      const wanted = new Set(selected)
      const existing = new Set(currentRoles ?? [])
      // Assign role yang baru dicentang.
      for (const role of selected) {
        if (!existing.has(role)) {
          await assignMutation.mutateAsync(
            new RoleAssignment({ user: String(user.id), role })
          )
        }
      }
      // Unassign role yang dicabut — kecuali role utama (users.role) yang
      // selalu disinkronkan EnsureUserRoleAssignments.
      for (const role of existing) {
        if (!wanted.has(role) && role !== primaryRole) {
          await unassignMutation.mutateAsync(
            new RoleAssignment({ user: String(user.id), role })
          )
        }
      }
      toast.success(`Roles updated for ${user.username}`)
      onOpenChange(false)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to update roles')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader className='text-start'>
          <DialogTitle>Assign roles — {user?.username}</DialogTitle>
          <DialogDescription>
            The primary role (from the account) is fixed. Extra roles can be
            assigned or revoked.
          </DialogDescription>
        </DialogHeader>
        <div className='space-y-3'>
          {Object.keys(ROLE_META).map((role) => {
            const isPrimary = role === primaryRole
            const checked = selected.includes(role) || isPrimary
            return (
              <div key={role} className='flex items-center gap-3'>
                <Checkbox
                  id={`role-${role}`}
                  checked={checked}
                  disabled={isPrimary}
                  onCheckedChange={(v) => toggle(role, v === true)}
                />
                <Label
                  htmlFor={`role-${role}`}
                  className='flex cursor-pointer items-center gap-2'
                >
                  <Badge
                    variant='outline'
                    className={cn('capitalize', roleClassName(role))}
                  >
                    {roleLabel(role)}
                  </Badge>
                  {isPrimary && (
                    <span className='text-xs text-muted-foreground'>
                      (primary)
                    </span>
                  )}
                </Label>
              </div>
            )
          })}
        </div>
        <DialogFooter>
          <Button
            onClick={onSave}
            disabled={assignMutation.isPending || unassignMutation.isPending}
          >
            {assignMutation.isPending || unassignMutation.isPending
              ? 'Saving...'
              : 'Save roles'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
