import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
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
import { profileFormSchema, type ProfileFormValues } from '../../data/schema'
import { EXPIRE_MODES } from '../../data/constants'
import {
  useCreateHotspotProfileMutation,
  useUpdateHotspotProfileMutation,
} from '../../api/use-hotspot-profiles'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function ProfileMutateDialog() {
  const { open, setOpen, currentProfile, setCurrentProfile } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()

  const isEdit = open === 'profile-update'
  const isOpen = open === 'profile-create' || open === 'profile-update'

  const createMutation = useCreateHotspotProfileMutation()
  const updateMutation = useUpdateHotspotProfileMutation()

  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(profileFormSchema),
    defaultValues: {
      name: '',
      addressPool: 'none',
      sharedUsers: '1',
      rateLimit: '',
      parentQueue: 'none',
      price: '0',
      sellingPrice: '0',
      validity: '',
      expireMode: '0',
      lockUser: false,
      lockServer: false,
      enableRecording: false,
      comment: '',
    },
  })

  useEffect(() => {
    if (isEdit && currentProfile) {
      form.reset({
        name: currentProfile.name,
        addressPool: currentProfile.addressPool || 'none',
        sharedUsers: currentProfile.sharedUsers || '1',
        rateLimit: currentProfile.rateLimit || '',
        parentQueue: currentProfile.parentQueue || 'none',
        price: String(currentProfile.price || 0),
        sellingPrice: String(currentProfile.sellingPrice || 0),
        validity: currentProfile.validity || '',
        expireMode: (currentProfile.modeExpire as any) || '0',
        lockUser: currentProfile.lockUser === 'Enable',
        lockServer: currentProfile.lockServer === 'Enable',
        enableRecording: false,
        comment: currentProfile.comment || '',
      })
    } else if (open === 'profile-create') {
      form.reset({
        name: '',
        addressPool: 'none',
        sharedUsers: '1',
        rateLimit: '',
        parentQueue: 'none',
        price: '0',
        sellingPrice: '0',
        validity: '',
        expireMode: '0',
        lockUser: false,
        lockServer: false,
        enableRecording: false,
        comment: '',
      })
    }
  }, [open, isEdit, currentProfile, form])

  const onSubmit = async (values: ProfileFormValues) => {
    try {
      const payload = {
        name: values.name.trim(),
        addressPool: values.addressPool === 'none' ? '' : values.addressPool,
        sharedUsers: values.sharedUsers,
        rateLimit: values.rateLimit.trim(),
        parentQueue: values.parentQueue === 'none' ? '' : values.parentQueue,
        price: values.price,
        sellingPrice: values.sellingPrice,
        validity: values.validity.trim(),
        expireMode: values.expireMode,
        lockUser: values.lockUser,
        lockServer: values.lockServer,
        enableRecording: values.enableRecording,
        comment: values.comment.trim(),
      }

      if (isEdit && currentProfile) {
        await updateMutation.mutateAsync({
          deviceId: selectedDeviceId,
          rosId: currentProfile.id,
          profile: payload,
        })
        toast.success(`Profile "${values.name}" updated successfully!`)
      } else {
        await createMutation.mutateAsync({
          deviceId: selectedDeviceId,
          profile: payload,
        })
        toast.success(`Profile "${values.name}" created successfully!`)
      }
      setOpen(null)
      setCurrentProfile(null)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to save user profile')
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={(val) => !val && setOpen(null)}>
      <DialogContent className='sm:max-w-[560px]'>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit User Profile' : 'Add User Profile'}</DialogTitle>
          <DialogDescription>
            Configure bandwidth rate limits, expiry modes, prices, and session parameters for this profile.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4'>
            <Tabs defaultValue='general' className='w-full'>
              <TabsList className='grid w-full grid-cols-2'>
                <TabsTrigger value='general'>General & Network</TabsTrigger>
                <TabsTrigger value='details'>Mikhmon & Expiry</TabsTrigger>
              </TabsList>

              {/* ===== TAB GENERAL ===== */}
              <TabsContent value='general' className='space-y-3 pt-2'>
                <FormField
                  control={form.control}
                  name='name'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Profile Name *</FormLabel>
                      <FormControl>
                        <Input placeholder='e.g. 1Jam-3K or VIP-Monthly' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='sharedUsers'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Shared Users</FormLabel>
                        <FormControl>
                          <Input type='number' min='1' max='999' {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name='rateLimit'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Rate Limit [rx/tx]</FormLabel>
                        <FormControl>
                          <Input placeholder='e.g. 1M/2M or 512k/1M' {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='addressPool'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Address Pool</FormLabel>
                        <FormControl>
                          <Input placeholder='none / pool name' {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name='parentQueue'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Parent Queue</FormLabel>
                        <FormControl>
                          <Input placeholder='none / parent queue' {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </TabsContent>

              {/* ===== TAB DETAILS / MIKHMON ===== */}
              <TabsContent value='details' className='space-y-3 pt-2'>
                <FormField
                  control={form.control}
                  name='expireMode'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Expire Mode</FormLabel>
                      <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder='Select expire mode' />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {EXPIRE_MODES.map((mode) => (
                            <SelectItem key={mode.value} value={mode.value}>
                              {mode.label} ({mode.desc})
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
                  name='validity'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Validity</FormLabel>
                      <FormControl>
                        <Input placeholder='e.g. 1d, 3h, 30d' {...field} />
                      </FormControl>
                      <FormDescription className='text-[11px]'>
                        Duration before voucher expires (e.g. 3h, 1d, 30d).
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='price'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Price (Rp)</FormLabel>
                        <FormControl>
                          <Input type='number' placeholder='2000' {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name='sellingPrice'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Selling Price (Rp)</FormLabel>
                        <FormControl>
                          <Input type='number' placeholder='3000' {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <div className='grid grid-cols-2 gap-4 rounded-lg border p-3 bg-muted/30'>
                  <FormField
                    control={form.control}
                    name='lockUser'
                    render={({ field }) => (
                      <FormItem className='flex flex-row items-center justify-between space-y-0'>
                        <FormLabel className='text-xs'>Lock to MAC</FormLabel>
                        <FormControl>
                          <Switch checked={field.value} onCheckedChange={field.onChange} />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name='lockServer'
                    render={({ field }) => (
                      <FormItem className='flex flex-row items-center justify-between space-y-0'>
                        <FormLabel className='text-xs'>Lock to Server</FormLabel>
                        <FormControl>
                          <Switch checked={field.value} onCheckedChange={field.onChange} />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                </div>
              </TabsContent>
            </Tabs>

            <DialogFooter className='pt-2'>
              <Button type='button' variant='outline' onClick={() => setOpen(null)}>
                Cancel
              </Button>
              <Button
                type='submit'
                disabled={createMutation.isPending || updateMutation.isPending}
              >
                {createMutation.isPending || updateMutation.isPending ? 'Saving...' : 'Save Profile'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
