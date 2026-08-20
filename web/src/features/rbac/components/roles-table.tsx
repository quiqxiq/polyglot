'use client'

import { useState, useMemo } from 'react'
import { type Policy } from '@/gen/v1/rbac_pb'
import { ShieldCheck, ShieldAlert, Edit, Trash2, KeyRound } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { roleClassName, roleLabel, ROLE_OPTIONS } from '@/features/users/data/roles'
import { CreateRoleDialog } from './create-role-dialog'
import { EditRoleDialog } from './edit-role-dialog'
import { DeleteRoleDialog } from './delete-role-dialog'

interface RolesTableProps {
  policies: Policy[]
  isLoading?: boolean
}

export function RolesTable({ policies, isLoading }: RolesTableProps) {
  const [editingRole, setEditingRole] = useState<string | null>(null)
  const [deletingRole, setDeletingRole] = useState<string | null>(null)

  // Map role -> list of permission strings (e.g. "device:read")
  const rolePermissionsMap = useMemo(() => {
    const map = new Map<string, string[]>()

    // Ensure default system roles are present in the list
    ROLE_OPTIONS.forEach((r) => {
      map.set(r.value, [])
    })

    // Populate policies
    policies.forEach((p) => {
      if (!p.sub) return
      const current = map.get(p.sub) || []
      if (!current.includes(p.obj)) {
        current.push(p.obj)
      }
      map.set(p.sub, current)
    })

    return map
  }, [policies])

  const roleList = useMemo(() => {
    return Array.from(rolePermissionsMap.keys()).sort((a, b) => {
      if (a === 'owner') return -1
      if (b === 'owner') return 1
      if (a === 'admin') return -1
      if (b === 'admin') return 1
      return a.localeCompare(b)
    })
  }, [rolePermissionsMap])

  return (
    <div className='space-y-4'>
      {/* Top Header Bar */}
      <div className='flex items-center justify-between gap-2'>
        <div>
          <p className='text-sm font-medium text-foreground'>
            System Roles &amp; Permissions
          </p>
          <p className='text-xs text-muted-foreground'>
            Manage permissions matrix per role or create custom roles with tailored access.
          </p>
        </div>
        <CreateRoleDialog existingRoles={roleList} />
      </div>

      {/* Roles List Table */}
      <div className='rounded-md border'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className='w-48'>Role</TableHead>
              <TableHead className='w-36'>Permissions Count</TableHead>
              <TableHead>Granted Permissions Preview</TableHead>
              <TableHead className='w-36 text-right'>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={4} className='h-24 text-center text-xs text-muted-foreground'>
                  Loading roles and policies...
                </TableCell>
              </TableRow>
            ) : roleList.length ? (
              roleList.map((role) => {
                const perms = rolePermissionsMap.get(role) || []
                const isOwner = role === 'owner'
                const isSystemDefault = ['owner', 'admin', 'agent', 'teknisi'].includes(role)

                return (
                  <TableRow key={role}>
                    <TableCell className='font-medium'>
                      <div className='flex items-center gap-2'>
                        <Badge
                          variant='outline'
                          className={cn('capitalize font-mono', roleClassName(role))}
                        >
                          {roleLabel(role)}
                        </Badge>
                        {isOwner && (
                          <span title='Owner has unrestricted access'>
                            <ShieldCheck className='h-3.5 w-3.5 text-amber-500' />
                          </span>
                        )}
                        {!isSystemDefault && (
                          <span className='text-[10px] px-1 py-0.5 rounded bg-muted text-muted-foreground'>
                            custom
                          </span>
                        )}
                      </div>
                    </TableCell>

                    <TableCell>
                      {isOwner ? (
                        <Badge variant='secondary' className='text-xs font-mono font-normal'>
                          All (*)
                        </Badge>
                      ) : (
                        <Badge variant='outline' className='text-xs font-mono'>
                          <KeyRound className='h-3 w-3 mr-1 text-muted-foreground' />
                          {perms.length} perms
                        </Badge>
                      )}
                    </TableCell>

                    <TableCell>
                      {isOwner ? (
                        <span className='text-xs text-muted-foreground italic'>
                          Full administrative access across all modules
                        </span>
                      ) : perms.length === 0 ? (
                        <span className='text-xs text-rose-500/80 italic flex items-center gap-1'>
                          <ShieldAlert className='h-3 w-3' /> No permissions assigned
                        </span>
                      ) : (
                        <div className='flex flex-wrap gap-1 max-h-16 overflow-hidden'>
                          {perms.slice(0, 8).map((p) => (
                            <code
                              key={p}
                              className='rounded bg-muted px-1.5 py-0.5 text-[10px] font-mono text-muted-foreground'
                            >
                              {p}
                            </code>
                          ))}
                          {perms.length > 8 && (
                            <span className='text-[10px] text-muted-foreground self-center font-mono'>
                              +{perms.length - 8} more
                            </span>
                          )}
                        </div>
                      )}
                    </TableCell>

                    <TableCell className='text-right'>
                      <div className='flex items-center justify-end gap-1'>
                        <Button
                          variant='outline'
                          size='sm'
                          className='h-8 text-xs gap-1.5'
                          onClick={() => setEditingRole(role)}
                        >
                          <Edit className='h-3.5 w-3.5' />
                          {isOwner ? 'View' : 'Edit'}
                        </Button>

                        {!isOwner && (
                          <Button
                            variant='ghost'
                            size='icon'
                            className='h-8 w-8 text-muted-foreground hover:text-destructive'
                            title='Delete Role'
                            onClick={() => setDeletingRole(role)}
                          >
                            <Trash2 className='h-3.5 w-3.5' />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })
            ) : (
              <TableRow>
                <TableCell colSpan={4} className='h-24 text-center text-xs text-muted-foreground'>
                  No roles found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Dialogs */}
      <EditRoleDialog
        role={editingRole}
        initialPermissions={editingRole ? rolePermissionsMap.get(editingRole) || [] : []}
        open={Boolean(editingRole)}
        onOpenChange={(open) => {
          if (!open) setEditingRole(null)
        }}
      />

      <DeleteRoleDialog
        role={deletingRole}
        open={Boolean(deletingRole)}
        onOpenChange={(open) => {
          if (!open) setDeletingRole(null)
        }}
      />
    </div>
  )
}
