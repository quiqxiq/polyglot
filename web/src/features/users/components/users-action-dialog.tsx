'use client'

import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  CreateUserRequest,
  UpdateUserRequest,
  type User,
} from '@/gen/v1/users_pb'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { PasswordInput } from '@/components/password-input'
import { SelectDropdown } from '@/components/select-dropdown'
import { useCreateUserMutation, useUpdateUserMutation } from '../api/use-users'
import { ROLE_OPTIONS } from '../data/roles'

const formSchema = z.object({
  username: z.string().min(1, 'Username is required.'),
  email: z.string().email('Enter a valid email.'),
  role: z.string().min(1, 'Role is required.'),
  password: z.string().optional(),
})

type UserForm = z.infer<typeof formSchema>

type UsersActionDialogProps = {
  currentRow?: User | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UsersActionDialog({
  currentRow,
  open,
  onOpenChange,
}: UsersActionDialogProps) {
  const isEdit = !!currentRow
  const createMutation = useCreateUserMutation()
  const updateMutation = useUpdateUserMutation()

  const isSelf = isEdit && String(currentRow.id) === selfID()

  const form = useForm<UserForm>({
    resolver: zodResolver(formSchema),
    defaultValues: isEdit
      ? {
          username: currentRow.username,
          email: currentRow.email,
          role: currentRow.role,
          password: '',
        }
      : {
          username: '',
          email: '',
          role: '',
          password: '',
        },
  })

  async function onSubmit(values: UserForm) {
    try {
      if (isEdit && currentRow) {
        await updateMutation.mutateAsync(
          new UpdateUserRequest({
            id: currentRow.id,
            username: values.username,
            email: values.email,
            role: values.role,
          })
        )
        toast.success(`User ${values.username} updated`)
      } else {
        await createMutation.mutateAsync(
          new CreateUserRequest({
            username: values.username,
            email: values.email,
            password: values.password ?? '',
            role: values.role,
          })
        )
        toast.success(`User ${values.username} created`)
      }
      onOpenChange(false)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to save user')
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(state) => {
        if (!state) form.reset()
        onOpenChange(state)
      }}
    >
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader className='text-start'>
          <DialogTitle>{isEdit ? 'Edit User' : 'Add New User'}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? 'Update the user profile. Role changes apply immediately.'
              : 'Create a new account. The user can log in right away.'}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form
            id='user-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className='space-y-4 px-0.5'
          >
            <FormField
              control={form.control}
              name='username'
              render={({ field }) => (
                <FormItem className='grid grid-cols-6 items-center space-y-0 gap-x-4 gap-y-1'>
                  <FormLabel className='col-span-2 text-end'>
                    Username
                  </FormLabel>
                  <FormControl>
                    <Input
                      placeholder='john_doe'
                      className='col-span-4'
                      autoComplete='off'
                      {...field}
                    />
                  </FormControl>
                  <FormMessage className='col-span-4 col-start-3' />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='email'
              render={({ field }) => (
                <FormItem className='grid grid-cols-6 items-center space-y-0 gap-x-4 gap-y-1'>
                  <FormLabel className='col-span-2 text-end'>Email</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='john.doe@example.com'
                      className='col-span-4'
                      autoComplete='off'
                      {...field}
                    />
                  </FormControl>
                  <FormMessage className='col-span-4 col-start-3' />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='role'
              render={({ field }) => (
                <FormItem className='grid grid-cols-6 items-center space-y-0 gap-x-4 gap-y-1'>
                  <FormLabel className='col-span-2 text-end'>Role</FormLabel>
                  <SelectDropdown
                    defaultValue={field.value}
                    onValueChange={field.onChange}
                    placeholder='Select a role'
                    className='col-span-4'
                    items={ROLE_OPTIONS}
                    disabled={isSelf}
                  />
                  <FormMessage className='col-span-4 col-start-3' />
                </FormItem>
              )}
            />
            {!isEdit && (
              <FormField
                control={form.control}
                name='password'
                render={({ field }) => (
                  <FormItem className='grid grid-cols-6 items-center space-y-0 gap-x-4 gap-y-1'>
                    <FormLabel className='col-span-2 text-end'>
                      Password
                    </FormLabel>
                    <FormControl>
                      <PasswordInput
                        placeholder='Min. 8 characters'
                        className='col-span-4'
                        {...field}
                      />
                    </FormControl>
                    <FormMessage className='col-span-4 col-start-3' />
                  </FormItem>
                )}
              />
            )}
          </form>
        </Form>
        <DialogFooter>
          <Button
            type='submit'
            form='user-form'
            disabled={createMutation.isPending || updateMutation.isPending}
          >
            {createMutation.isPending || updateMutation.isPending
              ? 'Saving...'
              : 'Save changes'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function selfID(): string {
  return useAuthStore.getState().auth.user?.accountNo ?? ''
}
