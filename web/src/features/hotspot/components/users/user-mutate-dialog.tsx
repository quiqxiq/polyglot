import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { Eye, EyeOff } from 'lucide-react'
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
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { userFormSchema, type UserFormValues } from '../../data/schema'
import {
  useCreateHotspotUserMutation,
  useUpdateHotspotUserMutation,
} from '../../api/use-hotspot-users'
import { useHotspotProfilesQuery } from '../../api/use-hotspot-profiles'
import { useHotspotServersQuery } from '../../api/use-hotspot-sessions'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function UserMutateDialog() {
  const { open, setOpen, currentUser, setCurrentUser } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()
  const [showPassword, setShowPassword] = useState(false)

  const isEdit = open === 'user-update'
  const isOpen = open === 'user-create' || open === 'user-update'

  const { data: profiles = [] } = useHotspotProfilesQuery(selectedDeviceId, isOpen)
  const { data: servers = [] } = useHotspotServersQuery(selectedDeviceId, isOpen)

  const createMutation = useCreateHotspotUserMutation()
  const updateMutation = useUpdateHotspotUserMutation()

  const form = useForm<UserFormValues>({
    resolver: zodResolver(userFormSchema),
    defaultValues: {
      name: '',
      password: '',
      profile: 'default',
      server: 'all',
      macAddress: '',
      timeLimit: '',
      dataLimit: '',
      comment: '',
      resetCounter: false,
    },
  })

  useEffect(() => {
    if (isEdit && currentUser) {
      form.reset({
        name: currentUser.name,
        password: currentUser.password || '',
        profile: currentUser.profile || 'default',
        server: currentUser.server || 'all',
        macAddress: '',
        timeLimit: currentUser.limitUptime || '',
        dataLimit: currentUser.limitBytes || '',
        comment: currentUser.comment || '',
        resetCounter: false,
      })
    } else if (open === 'user-create') {
      form.reset({
        name: '',
        password: '',
        profile: profiles[0]?.name || 'default',
        server: 'all',
        macAddress: '',
        timeLimit: '',
        dataLimit: '',
        comment: '',
        resetCounter: false,
      })
    }
  }, [open, isEdit, currentUser, form, profiles])

  const onSubmit = async (values: UserFormValues) => {
    try {
      if (isEdit && currentUser) {
        await updateMutation.mutateAsync({
          deviceId: selectedDeviceId,
          rosId: currentUser.id,
          name: values.name.trim(),
          password: values.password.trim(),
          profile: values.profile,
          server: values.server === 'all' ? '' : values.server,
          macAddress: values.macAddress?.trim() || '',
          timeLimit: values.timeLimit?.trim() || '',
          dataLimit: values.dataLimit?.trim() || '',
          comment: values.comment?.trim() || '',
          resetCounter: values.resetCounter,
          expireDate: '',
          userCode: '',
        })
        toast.success(`User "${values.name}" updated successfully!`)
      } else {
        await createMutation.mutateAsync({
          deviceId: selectedDeviceId,
          name: values.name.trim(),
          password: values.password.trim(),
          profile: values.profile,
          server: values.server === 'all' ? '' : values.server,
          macAddress: values.macAddress?.trim() || '',
          timeLimit: values.timeLimit?.trim() || '',
          dataLimit: values.dataLimit?.trim() || '',
          comment: values.comment?.trim() || '',
        })
        toast.success(`User "${values.name}" created successfully!`)
      }
      setOpen(null)
      setCurrentUser(null)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to save hotspot user')
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={(val) => !val && setOpen(null)}>
      <DialogContent className='sm:max-w-[500px]'>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit Hotspot User' : 'Add Hotspot User'}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? 'Update user credentials, limit quotas, or reset traffic counters.'
              : 'Create a single voucher or member user on the MikroTik router.'}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4'>
            <Tabs defaultValue='general' className='w-full'>
              <TabsList className='grid w-full grid-cols-2'>
                <TabsTrigger value='general'>Credentials & Profile</TabsTrigger>
                <TabsTrigger value='limits'>Limits & Quota</TabsTrigger>
              </TabsList>

              <TabsContent value='general' className='space-y-3 pt-2'>
                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='server'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Server</FormLabel>
                        <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder='all' />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value='all'>all (All Servers)</SelectItem>
                            {servers.map((s) => (
                              <SelectItem key={s.id || s.name} value={s.name}>
                                {s.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name='profile'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>User Profile *</FormLabel>
                        <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder='Select profile' />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {profiles.map((p) => (
                              <SelectItem key={p.id || p.name} value={p.name}>
                                {p.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <FormField
                  control={form.control}
                  name='name'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Username / Code *</FormLabel>
                      <FormControl>
                        <Input placeholder='Username or voucher code' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name='password'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Password *</FormLabel>
                      <div className='relative'>
                        <FormControl>
                          <Input
                            type={showPassword ? 'text' : 'password'}
                            placeholder='Password'
                            className='pr-9'
                            {...field}
                          />
                        </FormControl>
                        <Button
                          type='button'
                          variant='ghost'
                          size='sm'
                          className='absolute right-0 top-0 h-full px-3 text-muted-foreground hover:text-foreground'
                          onClick={() => setShowPassword(!showPassword)}
                        >
                          {showPassword ? <EyeOff className='size-4' /> : <Eye className='size-4' />}
                        </Button>
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name='comment'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Comment / Batch Tag</FormLabel>
                      <FormControl>
                        <Input placeholder='e.g. vc-10k-batch1' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </TabsContent>

              <TabsContent value='limits' className='space-y-3 pt-2'>
                <FormField
                  control={form.control}
                  name='timeLimit'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Time Limit (Uptime)</FormLabel>
                      <FormControl>
                        <Input placeholder='e.g. 3h, 30m, 1d' {...field} />
                      </FormControl>
                      <FormDescription className='text-[11px]'>
                        MikroTik uptime limit duration (e.g. 3h or 30m).
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name='dataLimit'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Data Limit (Bytes)</FormLabel>
                      <FormControl>
                        <Input placeholder='e.g. 1000M, 2G' {...field} />
                      </FormControl>
                      <FormDescription className='text-[11px]'>
                        Total quota limit (e.g. 500M, 1000M, 2G).
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name='macAddress'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Lock MAC Address</FormLabel>
                      <FormControl>
                        <Input placeholder='AA:BB:CC:DD:EE:FF (Optional)' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                {isEdit && (
                  <div className='rounded-lg border p-3 bg-muted/30'>
                    <FormField
                      control={form.control}
                      name='resetCounter'
                      render={({ field }) => (
                        <FormItem className='flex flex-row items-center justify-between space-y-0'>
                          <div className='space-y-0.5'>
                            <FormLabel className='text-xs font-semibold'>Reset Uptime & Counters</FormLabel>
                            <FormDescription className='text-[11px]'>
                              Reset uptime and bytes transfer counters to zero on save.
                            </FormDescription>
                          </div>
                          <FormControl>
                            <Switch checked={field.value} onCheckedChange={field.onChange} />
                          </FormControl>
                        </FormItem>
                      )}
                    />
                  </div>
                )}
              </TabsContent>
            </Tabs>

            <DialogFooter className='pt-2'>
              <Button type='button' variant='outline' onClick={() => setOpen(null)}>
                Cancel
              </Button>
              <Button type='submit' disabled={createMutation.isPending || updateMutation.isPending}>
                {createMutation.isPending || updateMutation.isPending ? 'Saving...' : 'Save User'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
