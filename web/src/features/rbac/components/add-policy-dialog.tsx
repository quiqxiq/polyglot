'use client'

import { useState } from 'react'
import { Policy } from '@/gen/v1/rbac_pb'
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
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ROLE_OPTIONS } from '@/features/users/data/roles'
import { useAddPolicyMutation } from '../api/use-rbac'
import {
  ALL,
  RBAC_ACTIONS,
  RBAC_RESOURCES,
  composeObject,
} from '../data/catalog'

type AddPolicyDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AddPolicyDialog({ open, onOpenChange }: AddPolicyDialogProps) {
  const [role, setRole] = useState('')
  const [resource, setResource] = useState('')
  const [action, setAction] = useState('')
  const addMutation = useAddPolicyMutation()

  const object =
    role && resource && action ? composeObject(resource, action) : ''

  async function onAdd() {
    if (!role || !object) return
    try {
      const res = await addMutation.mutateAsync(
        new Policy({ sub: role, obj: object, act: '*' })
      )
      if (!res.success) {
        toast.error('Policy already exists')
        return
      }
      toast.success(`Policy added: ${role} → ${object}`)
      setRole('')
      setResource('')
      setAction('')
      onOpenChange(false)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to add policy')
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(state) => {
        if (!state) {
          setRole('')
          setResource('')
          setAction('')
        }
        onOpenChange(state)
      }}
    >
      <DialogContent className='sm:max-w-md'>
        <DialogHeader className='text-start'>
          <DialogTitle>Add Policy</DialogTitle>
          <DialogDescription>
            Grant a role access to a resource. The object becomes{' '}
            <code className='rounded bg-muted px-1 py-0.5 text-xs'>
              {object || 'resource:action'}
            </code>
            .
          </DialogDescription>
        </DialogHeader>
        <div className='space-y-4'>
          <div className='space-y-2'>
            <Label>Role</Label>
            <Select value={role} onValueChange={setRole}>
              <SelectTrigger>
                <SelectValue placeholder='Select a role' />
              </SelectTrigger>
              <SelectContent>
                {ROLE_OPTIONS.map((r) => (
                  <SelectItem key={r.value} value={r.value}>
                    {r.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className='space-y-2'>
            <Label>Resource</Label>
            <Select value={resource} onValueChange={setResource}>
              <SelectTrigger>
                <SelectValue placeholder='Select a resource' />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ALL}>All resources</SelectItem>
                {RBAC_RESOURCES.map((r) => (
                  <SelectItem key={r} value={r}>
                    {r}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className='space-y-2'>
            <Label>Action</Label>
            <Select value={action} onValueChange={setAction}>
              <SelectTrigger>
                <SelectValue placeholder='Select an action' />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ALL}>All actions</SelectItem>
                {RBAC_ACTIONS.map((a) => (
                  <SelectItem key={a} value={a}>
                    {a}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button
            onClick={onAdd}
            disabled={!role || !object || addMutation.isPending}
          >
            {addMutation.isPending ? 'Adding...' : 'Add policy'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
