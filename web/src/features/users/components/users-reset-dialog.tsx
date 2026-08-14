'use client'

import { useState } from 'react'
import { ResetPasswordRequest, type User } from '@/gen/v1/users_pb'
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
import { PasswordInput } from '@/components/password-input'
import { useResetPasswordMutation } from '../api/use-users'

type UsersResetDialogProps = {
  user?: User | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UsersResetDialog({
  user,
  open,
  onOpenChange,
}: UsersResetDialogProps) {
  const [password, setPassword] = useState('')
  const resetMutation = useResetPasswordMutation()

  async function onReset() {
    if (!user || !password) return
    try {
      await resetMutation.mutateAsync(
        new ResetPasswordRequest({ id: user.id, newPassword: password })
      )
      toast.success(`Password for ${user.username} reset`)
      setPassword('')
      onOpenChange(false)
    } catch (err: unknown) {
      toast.error(
        err instanceof Error ? err.message : 'Failed to reset password'
      )
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(state) => {
        if (!state) setPassword('')
        onOpenChange(state)
      }}
    >
      <DialogContent className='sm:max-w-md'>
        <DialogHeader className='text-start'>
          <DialogTitle>Reset Password</DialogTitle>
          <DialogDescription>
            Set a new password for <strong>{user?.username}</strong>. The user
            can log in with it immediately.
          </DialogDescription>
        </DialogHeader>
        <div className='space-y-2'>
          <Label htmlFor='new-password'>New password</Label>
          <PasswordInput
            id='new-password'
            placeholder='Min. 8 characters'
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoFocus
          />
          {password.length > 0 && password.length < 8 && (
            <p className='text-sm text-destructive'>
              Password must be at least 8 characters.
            </p>
          )}
        </div>
        <DialogFooter>
          <Button
            variant='outline'
            onClick={() => onOpenChange(false)}
            disabled={resetMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={onReset}
            disabled={resetMutation.isPending || password.length < 8}
          >
            {resetMutation.isPending ? 'Resetting...' : 'Reset password'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
