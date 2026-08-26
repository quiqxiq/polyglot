'use client'

import { useEffect } from 'react'
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
import { Checkbox } from '@/components/ui/checkbox'
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
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useCreateUserMutation, useUpdateUserMutation } from '../api/use-users'
import { ROLE_OPTIONS } from '../data/roles'
import { cn } from '@/lib/utils'

const formSchema = z.object({
  username: z.string().min(1, 'Username is required.'),
  email: z.string().email('Enter a valid email.'),
  role: z.string().min(1, 'Role is required.'),
  fullName: z.string().optional(),
  phoneNumber: z.string().optional(),
  specialization: z.string().optional(),
  password: z.string().optional(),
  assignedDeviceIds: z.array(z.string()),
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
  const { data: devices = [] } = useDevicesQuery()

  const currentUser = useAuthStore((s) => s.auth.user)
  const isCurrentUserOwner = Boolean(currentUser?.role?.includes('owner'))
  const isSelf = isEdit && String(currentRow.id) === (currentUser?.accountNo ?? '')

  // Filter available role options based on actor role and hierarchy:
  // - Owner: can create/assign any role (owner, admin, agent, teknisi)
  // - Admin (non-owner):
  //     - Create mode: ONLY agent, teknisi (cannot create admin or owner)
  //     - Edit self: admin, agent, teknisi (cannot promote self to owner)
  //     - Edit other: ONLY agent, teknisi (cannot promote other to admin or owner)
  const availableRoleOptions = ROLE_OPTIONS.filter((opt) => {
    if (isCurrentUserOwner) return true
    if (opt.value === 'owner') return false
    if (!isSelf && opt.value === 'admin') return false
    return true
  })

  const form = useForm<UserForm>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      email: '',
      role: '',
      fullName: '',
      phoneNumber: '',
      specialization: '',
      password: '',
      assignedDeviceIds: [],
    },
  })

  useEffect(() => {
    if (open) {
      if (currentRow) {
        form.reset({
          username: currentRow.username,
          email: currentRow.email,
          role: currentRow.role,
          fullName: currentRow.fullName ?? '',
          phoneNumber: currentRow.phoneNumber ?? '',
          specialization: currentRow.specialization ?? '',
          password: '',
          assignedDeviceIds: currentRow.assignedDeviceIds ?? [],
        })
      } else {
        form.reset({
          username: '',
          email: '',
          role: '',
          fullName: '',
          phoneNumber: '',
          specialization: '',
          password: '',
          assignedDeviceIds: [],
        })
      }
    }
  }, [open, currentRow, form])

  const selectedRole = form.watch('role')

  async function onSubmit(values: UserForm) {
    try {
      const assigned = values.role === 'owner' ? [] : (values.assignedDeviceIds ?? [])
      if (isEdit && currentRow) {
        await updateMutation.mutateAsync(
          new UpdateUserRequest({
            id: currentRow.id,
            username: values.username,
            email: values.email,
            role: values.role,
            fullName: values.fullName ?? '',
            phoneNumber: values.phoneNumber ?? '',
            specialization: values.specialization ?? '',
            assignedDeviceIds: assigned,
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
            fullName: values.fullName ?? '',
            phoneNumber: values.phoneNumber ?? '',
            specialization: values.specialization ?? '',
            assignedDeviceIds: assigned,
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
                    isControlled
                    placeholder='Select a role'
                    className='col-span-4'
                    items={availableRoleOptions}
                  />
                  <FormMessage className='col-span-4 col-start-3' />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='fullName'
              render={({ field }) => (
                <FormItem className='grid grid-cols-6 items-center space-y-0 gap-x-4 gap-y-1'>
                  <FormLabel className='col-span-2 text-end'>
                    Nama Lengkap
                  </FormLabel>
                  <FormControl>
                    <Input
                      placeholder='Budi Santoso'
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
              name='phoneNumber'
              render={({ field }) => (
                <FormItem className='grid grid-cols-6 items-center space-y-0 gap-x-4 gap-y-1'>
                  <FormLabel className='col-span-2 text-end'>
                    No. WhatsApp / HP
                  </FormLabel>
                  <FormControl>
                    <Input
                      placeholder='08123456789 atau 6281...'
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
              name='specialization'
              render={({ field }) => (
                <FormItem className='grid grid-cols-6 items-center space-y-0 gap-x-4 gap-y-1'>
                  <FormLabel className='col-span-2 text-end'>
                    Spesialisasi
                  </FormLabel>
                  <FormControl>
                    <Input
                      placeholder='Fiber Optic, Splicing, Wireless, dll'
                      className='col-span-4'
                      autoComplete='off'
                      {...field}
                    />
                  </FormControl>
                  <FormMessage className='col-span-4 col-start-3' />
                </FormItem>
              )}
            />
            {selectedRole === 'owner' ? (
              <div className='rounded-lg bg-purple-500/10 p-3 text-xs text-purple-700 dark:text-purple-300 border border-purple-200 dark:border-purple-800/40 space-y-1'>
                <p className='font-semibold'>Hak Akses Global Owner</p>
                <p className='text-[11px] leading-relaxed text-muted-foreground dark:text-purple-300/80'>
                  Akun dengan role Owner secara otomatis memiliki akses penuh ke seluruh router MikroTik (Global Scope).
                </p>
              </div>
            ) : (
              <FormField
                control={form.control}
                name='assignedDeviceIds'
                render={({ field }) => (
                  <FormItem className='space-y-2 rounded-lg border p-3 bg-muted/20'>
                    <div className='flex items-center justify-between'>
                      <div>
                        <FormLabel className='text-xs font-semibold uppercase tracking-wider text-muted-foreground'>
                          Assigned MikroTik Routers
                        </FormLabel>
                        <p className='text-[11px] text-muted-foreground'>
                          {isCurrentUserOwner
                            ? 'Pilih router mana saja yang dapat diakses oleh user ini.'
                            : 'Pilih router dari daftar router yang Anda kelola.'}
                        </p>
                      </div>
                      {devices.length > 0 && (
                        <Button
                          type='button'
                          variant='ghost'
                          size='sm'
                          className='h-6 px-2 text-[11px]'
                          onClick={() => {
                            const allIds = devices.map((d) => d.id)
                            const isAllSelected = allIds.every((id) => field.value?.includes(id))
                            field.onChange(isAllSelected ? [] : allIds)
                          }}
                        >
                          {devices.every((d) => field.value?.includes(d.id)) ? 'Deselect All' : 'Select All'}
                        </Button>
                      )}
                    </div>
                    {devices.length === 0 ? (
                      <div className='py-2 text-center text-xs text-muted-foreground italic'>
                        Tidak ada router yang tersedia untuk di-assign.
                      </div>
                    ) : (
                      <div className='grid grid-cols-1 sm:grid-cols-2 gap-2 pt-1 max-h-40 overflow-y-auto pr-1'>
                        {devices.map((device) => {
                          const isChecked = field.value?.includes(device.id)
                          return (
                            <label
                              key={device.id}
                              className={cn(
                                'flex items-start gap-2.5 p-2 rounded-md border text-xs cursor-pointer transition-colors',
                                isChecked
                                  ? 'border-primary/50 bg-primary/5 font-medium'
                                  : 'border-border/60 hover:bg-muted/40'
                              )}
                            >
                              <Checkbox
                                checked={isChecked}
                                onCheckedChange={(checked) => {
                                  const current = field.value ?? []
                                  if (checked) {
                                    field.onChange([...current, device.id])
                                  } else {
                                    field.onChange(current.filter((id) => id !== device.id))
                                  }
                                }}
                                className='mt-0.5'
                              />
                              <div className='flex flex-col min-w-0 flex-1'>
                                <span className='truncate text-foreground'>{device.name}</span>
                                <span className='truncate text-[10px] text-muted-foreground'>
                                  {device.host} ({device.vendor || 'mikrotik'})
                                </span>
                              </div>
                            </label>
                          )
                        })}
                      </div>
                    )}
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
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
