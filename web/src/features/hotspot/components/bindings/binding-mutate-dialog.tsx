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
import { bindingFormSchema, type BindingFormValues } from '../../data/schema'
import { useHotspot } from '../../context/hotspot-context'
import {
  useCreateHotspotIPBindingMutation,
  useUpdateHotspotIPBindingMutation,
} from '../../api/use-hotspot-bindings'
import { useHotspotServersQuery } from '../../api/use-hotspot-sessions'
import { useDeviceStore } from '@/stores/device-store'
import type { HotspotServerInfo } from '@/gen/v1/hotspot_pb'

type BindingMutateDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit?: boolean
}

export function BindingMutateDialog({
  open,
  onOpenChange,
  isEdit = false,
}: BindingMutateDialogProps) {
  const { currentBinding, prefillBinding, setPrefillBinding } = useHotspot()
  const { selectedDeviceId } = useDeviceStore()
  const { data: servers = [] } = useHotspotServersQuery(selectedDeviceId || '')

  const createMutation = useCreateHotspotIPBindingMutation()
  const updateMutation = useUpdateHotspotIPBindingMutation()

  const form = useForm<BindingFormValues>({
    resolver: zodResolver(bindingFormSchema),
    defaultValues: {
      macAddress: '',
      address: '',
      toAddress: '',
      server: 'all',
      type: 'bypassed',
      comment: '',
      disabled: false,
    },
  })

  useEffect(() => {
    if (open) {
      if (isEdit && currentBinding) {
        form.reset({
          macAddress: currentBinding.macAddress || '',
          address: currentBinding.address || '',
          toAddress: currentBinding.toAddress || '',
          server: currentBinding.server || 'all',
          type: (currentBinding.type as 'bypassed' | 'blocked' | 'regular') || 'bypassed',
          comment: currentBinding.comment || '',
          disabled: currentBinding.disabled || false,
        })
      } else if (prefillBinding) {
        form.reset({
          macAddress: prefillBinding.macAddress || '',
          address: prefillBinding.address || '',
          toAddress: '',
          server: prefillBinding.server || 'all',
          type: 'bypassed',
          comment: 'Bypassed from host',
          disabled: false,
        })
        setPrefillBinding(null)
      } else {
        form.reset({
          macAddress: '',
          address: '',
          toAddress: '',
          server: 'all',
          type: 'bypassed',
          comment: '',
          disabled: false,
        })
      }
    }
  }, [open, isEdit, currentBinding, prefillBinding, form, setPrefillBinding])

  const onSubmit = async (values: BindingFormValues) => {
    if (!selectedDeviceId) return

    if (!values.macAddress && !values.address) {
      toast.error('Either MAC Address or IP Address must be specified.')
      return
    }

    onOpenChange(false)

    if (isEdit && currentBinding) {
      toast.promise(
        updateMutation.mutateAsync({
          deviceId: selectedDeviceId,
          rosId: currentBinding.id,
          macAddress: values.macAddress,
          address: values.address,
          toAddress: values.toAddress,
          server: values.server === 'all' ? '' : values.server,
          type: values.type,
          comment: values.comment,
          disabled: values.disabled,
        }),
        {
          loading: 'Updating IP binding...',
          success: 'IP binding updated successfully.',
          error: 'Failed to update IP binding.',
        }
      )
    } else {
      toast.promise(
        createMutation.mutateAsync({
          deviceId: selectedDeviceId,
          macAddress: values.macAddress,
          address: values.address,
          toAddress: values.toAddress,
          server: values.server === 'all' ? '' : values.server,
          type: values.type,
          comment: values.comment,
          disabled: values.disabled,
        }),
        {
          loading: 'Creating IP binding...',
          success: 'IP binding created successfully.',
          error: 'Failed to create IP binding.',
        }
      )
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[500px]'>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit IP Binding' : 'Add IP Binding'}</DialogTitle>
          <DialogDescription>
            Configure bypass, block, or regular NAT binding for non-hotspot devices like Smart TVs or CCTVs.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4 py-2'>
            <div className='grid grid-cols-2 gap-4'>
              <FormField
                control={form.control}
                name='macAddress'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className='text-xs'>MAC Address</FormLabel>
                    <FormControl>
                      <Input placeholder='AA:BB:CC:DD:EE:FF' className='font-mono text-xs' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='address'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className='text-xs'>IP Address</FormLabel>
                    <FormControl>
                      <Input placeholder='192.168.88.50' className='font-mono text-xs' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className='grid grid-cols-2 gap-4'>
              <FormField
                control={form.control}
                name='type'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className='text-xs'>Type</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder='Select type' />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value='bypassed'>Bypassed (No login needed)</SelectItem>
                        <SelectItem value='blocked'>Blocked (Denied access)</SelectItem>
                        <SelectItem value='regular'>Regular (Standard login)</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='server'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className='text-xs'>Hotspot Server</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder='Select server' />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value='all'>All Servers</SelectItem>
                        {servers.map((s: HotspotServerInfo) => (
                          <SelectItem key={s.id} value={s.name}>
                            {s.name}
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
              name='toAddress'
              render={({ field }) => (
                <FormItem>
                  <FormLabel className='text-xs'>To Address (Optional translation)</FormLabel>
                  <FormControl>
                    <Input placeholder='Leave empty if same' className='font-mono text-xs' {...field} />
                  </FormControl>
                  <FormDescription className='text-[10px]'>
                    Optional destination translation address.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='comment'
              render={({ field }) => (
                <FormItem>
                  <FormLabel className='text-xs'>Comment</FormLabel>
                  <FormControl>
                    <Input placeholder='e.g. CCTV Pos Satpam' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='disabled'
              render={({ field }) => (
                <FormItem className='flex items-center justify-between rounded-lg border p-3'>
                  <div className='space-y-0.5'>
                    <FormLabel className='text-xs'>Disabled</FormLabel>
                    <FormDescription className='text-[10px]'>
                      Disable this binding rule without deleting it.
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                  </FormControl>
                </FormItem>
              )}
            />

            <DialogFooter className='pt-2'>
              <Button type='button' variant='outline' onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <Button type='submit' disabled={createMutation.isPending || updateMutation.isPending}>
                {isEdit ? 'Save Changes' : 'Add Binding'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
