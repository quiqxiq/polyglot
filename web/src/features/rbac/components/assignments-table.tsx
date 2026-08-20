import { useState } from 'react'
import { type RoleAssignment as RoleAssignmentProto } from '@/gen/v1/rbac_pb'
import { type User } from '@/gen/v1/users_pb'
import { Settings2 } from 'lucide-react'
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
import { roleClassName, roleLabel } from '@/features/users/data/roles'
import { AssignmentDialog } from './assignment-dialog'

interface AssignmentsTableProps {
  users: User[]
  assignments: RoleAssignmentProto[]
  availableRoles?: string[]
  isLoading?: boolean
}

export function AssignmentsTable({
  users,
  assignments,
  availableRoles = [],
  isLoading,
}: AssignmentsTableProps) {
  const [editing, setEditing] = useState<User | null>(null)

  const rolesFor = (user: User): string[] => {
    const extra = assignments
      .filter((a) => a.user === String(user.id))
      .map((a) => a.role)
      .filter((r) => r !== user.role)
    return [...new Set([user.role, ...extra])]
  }

  return (
    <div className='space-y-4'>
      <p className='text-sm text-muted-foreground'>
        {users.length} accounts · multi-role per user (primary role is fixed)
      </p>
      <div className='rounded-md border'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className='w-64'>User</TableHead>
              <TableHead>Assigned roles</TableHead>
              <TableHead className='w-24 text-right'>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={3} className='h-24 text-center'>
                  Loading assignments...
                </TableCell>
              </TableRow>
            ) : users.length ? (
              users.map((user) => {
                const roles = rolesFor(user)
                return (
                  <TableRow key={String(user.id)}>
                    <TableCell>
                      <div className='font-medium'>{user.username}</div>
                      <div className='text-xs text-muted-foreground'>
                        {user.email}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className='flex flex-wrap gap-1'>
                        {roles.map((role) => (
                          <Badge
                            key={role}
                            variant='outline'
                            className={cn('capitalize font-mono', roleClassName(role))}
                          >
                            {roleLabel(role)}
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className='text-right'>
                      <Button
                        variant='ghost'
                        size='sm'
                        className='h-8 gap-1.5'
                        onClick={() => setEditing(user)}
                      >
                        <Settings2 className='h-3.5 w-3.5' /> Edit roles
                      </Button>
                    </TableCell>
                  </TableRow>
                )
              })
            ) : (
              <TableRow>
                <TableCell colSpan={3} className='h-24 text-center'>
                  No users found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <AssignmentDialog
        user={editing}
        currentRoles={editing ? rolesFor(editing) : []}
        availableRoles={availableRoles}
        open={!!editing}
        onOpenChange={(v) => !v && setEditing(null)}
      />
    </div>
  )
}
