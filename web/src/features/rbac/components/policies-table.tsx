import { useState } from 'react'
import { type Policy } from '@/gen/v1/rbac_pb'
import { PlusIcon, Trash2 } from 'lucide-react'
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
import { ALL, splitObject } from '../data/catalog'
import { AddPolicyDialog } from './add-policy-dialog'
import { RemovePolicyDialog } from './remove-policy-dialog'

interface PoliciesTableProps {
  policies: Policy[]
  isLoading?: boolean
}

export function PoliciesTable({ policies, isLoading }: PoliciesTableProps) {
  const [addOpen, setAddOpen] = useState(false)
  const [toRemove, setToRemove] = useState<Policy | null>(null)

  // Policy owner tidak boleh dihapus — owner adalah role tertinggi.
  const isLocked = (p: Policy) => p.sub === 'owner'

  return (
    <div className='space-y-4'>
      <div className='flex items-center justify-between'>
        <p className='text-sm text-muted-foreground'>
          {policies.length} policy rules — object format{' '}
          <code className='rounded bg-muted px-1.5 py-0.5 text-xs'>
            resource:action
          </code>
        </p>
        <Button onClick={() => setAddOpen(true)} className='h-9 gap-1.5'>
          <PlusIcon className='h-4 w-4' /> Add Policy
        </Button>
      </div>

      <div className='rounded-md border'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className='w-40'>Role</TableHead>
              <TableHead>Resource</TableHead>
              <TableHead className='w-32'>Action</TableHead>
              <TableHead className='w-24 text-right'>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={4} className='h-24 text-center'>
                  Loading policies...
                </TableCell>
              </TableRow>
            ) : policies.length ? (
              policies.map((p) => {
                const { resource, action } = splitObject(p.obj)
                const locked = isLocked(p)
                return (
                  <TableRow key={`${p.sub}-${p.obj}-${p.act}`}>
                    <TableCell>
                      <Badge
                        variant='outline'
                        className={cn('capitalize', roleClassName(p.sub))}
                      >
                        {roleLabel(p.sub)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <code className='rounded bg-muted px-1.5 py-0.5 text-xs'>
                        {resource === ALL ? 'all' : resource}
                      </code>
                    </TableCell>
                    <TableCell>
                      <code className='rounded bg-muted px-1.5 py-0.5 text-xs'>
                        {action || p.act || '*'}
                      </code>
                    </TableCell>
                    <TableCell className='text-right'>
                      <Button
                        variant='ghost'
                        size='icon'
                        className='h-8 w-8'
                        disabled={locked}
                        title={
                          locked ? 'Owner policies are locked' : 'Remove policy'
                        }
                        onClick={() => setToRemove(p)}
                      >
                        <Trash2 className='h-4 w-4' />
                      </Button>
                    </TableCell>
                  </TableRow>
                )
              })
            ) : (
              <TableRow>
                <TableCell colSpan={4} className='h-24 text-center'>
                  No policies found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <AddPolicyDialog open={addOpen} onOpenChange={setAddOpen} />
      <RemovePolicyDialog
        policy={toRemove}
        open={!!toRemove}
        onOpenChange={(v) => !v && setToRemove(null)}
      />
    </div>
  )
}
