'use client'

import { useEffect } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
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
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { SelectDropdown } from '@/components/select-dropdown'
import { useDevicesContext } from './devices-provider'
import { useUpdateDeviceMutation } from '../api/use-devices'
import { Device, UpdateDeviceRequest } from '@/gen/v1/device_pb'

const formSchema = z.object({
  name: z.string().min(1, 'Name is required.'),
  vendor: z.string().min(1, 'Vendor is required.'),
  driverType: z.string().min(1, 'Driver type is required.'),
  host: z.string().min(1, 'Host is required.'),
  port: z.number().min(1, 'Port is required.'),
  sshPort: z.number(),
  username: z.string().optional(),
  password: z.string().optional(),
  timeoutMs: z.number(),
  pollIntervalMs: z.number(),
  enabled: z.boolean(),
  tags: z.string().optional(),
})

export function DevicesActionDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = useDevicesContext()
  const isEdit = open === 'edit'
  const updateMutation = useUpdateDeviceMutation()

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      vendor: 'mikrotik',
      driverType: 'mikrotik',
      host: '',
      port: 8728,
      sshPort: 22,
      username: '',
      password: '',
      timeoutMs: 5000,
      pollIntervalMs: 10000,
      enabled: true,
      tags: '',
    },
  })

  useEffect(() => {
    if (currentRow && isEdit) {
      form.reset({
        name: currentRow.name,
        vendor: currentRow.vendor,
        driverType: currentRow.driverType,
        host: currentRow.host,
        port: currentRow.port,
        sshPort: currentRow.sshPort || 22,
        username: '',
        password: '',
        timeoutMs: currentRow.timeoutMs || 5000,
        pollIntervalMs: currentRow.pollIntervalMs || 10000,
        enabled: currentRow.enabled,
        tags: currentRow.tags ? currentRow.tags.join(', ') : '',
      })
    } else if (!isEdit) {
      form.reset({
        name: '',
        vendor: 'mikrotik',
        driverType: 'mikrotik',
        host: '',
        port: 8728,
        sshPort: 22,
        username: '',
        password: '',
        timeoutMs: 5000,
        pollIntervalMs: 10000,
        enabled: true,
        tags: '',
      })
    }
  }, [currentRow, isEdit, form])

  async function onSubmit(values: z.infer<typeof formSchema>) {
    try {
      const device = new Device({
        id: isEdit && currentRow ? currentRow.id : '',
        name: values.name,
        vendor: values.vendor,
        driverType: values.driverType,
        host: values.host,
        port: values.port,
        sshPort: values.sshPort,
        timeoutMs: values.timeoutMs,
        pollIntervalMs: values.pollIntervalMs,
        enabled: values.enabled,
        tags: values.tags ? values.tags.split(',').map((t) => t.trim()).filter(Boolean) : [],
      })

      const req = new UpdateDeviceRequest({
        device,
        username: values.username || '',
        password: values.password || '',
      })

      await updateMutation.mutateAsync(req)
      toast.success(isEdit ? 'Device updated successfully' : 'Device created successfully')
      setOpen(null)
      setCurrentRow(null)
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to save device'
      toast.error(errorMessage)
    }
  }

  return (
    <Dialog
      open={open === 'add' || open === 'edit'}
      onOpenChange={() => {
        setOpen(null)
        setCurrentRow(null)
      }}
    >
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit Device' : 'Add New Device'}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? 'Update details for this network device.'
              : 'Add a new network device or router to management inventory.'}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4'>
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Device Name</FormLabel>
                  <FormControl>
                    <Input placeholder='Router-Mikrotik-Main' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='grid grid-cols-2 gap-4'>
              <FormField
                control={form.control}
                name='vendor'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Vendor</FormLabel>
                    <SelectDropdown
                      defaultValue={field.value}
                      onValueChange={field.onChange}
                      items={[
                        { label: 'Mikrotik', value: 'mikrotik' },
                        { label: 'Cisco', value: 'cisco' },
                        { label: 'Huawei', value: 'huawei' },
                        { label: 'GenieACS', value: 'genieacs' },
                      ]}
                    />
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='driverType'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Driver Type</FormLabel>
                    <SelectDropdown
                      defaultValue={field.value}
                      onValueChange={field.onChange}
                      items={[
                        { label: 'Mikrotik API', value: 'mikrotik' },
                        { label: 'GenieACS CWMP', value: 'genieacs' },
                        { label: 'Generic SSH', value: 'ssh' },
                      ]}
                    />
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className='grid grid-cols-3 gap-4'>
              <FormField
                control={form.control}
                name='host'
                render={({ field }) => (
                  <FormItem className='col-span-2'>
                    <FormLabel>Host / IP Address</FormLabel>
                    <FormControl>
                      <Input placeholder='192.168.88.1' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='port'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>API Port</FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        placeholder='8728'
                        {...field}
                        onChange={(e) => field.onChange(e.target.valueAsNumber)}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className='grid grid-cols-2 gap-4'>
              <FormField
                control={form.control}
                name='username'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Username</FormLabel>
                    <FormControl>
                      <Input placeholder='admin' {...field} />
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
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <Input type='password' placeholder='••••••••' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className='grid grid-cols-2 gap-4'>
              <FormField
                control={form.control}
                name='sshPort'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>SSH Port</FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        placeholder='22'
                        {...field}
                        onChange={(e) => field.onChange(e.target.valueAsNumber)}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='tags'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Tags (comma separated)</FormLabel>
                    <FormControl>
                      <Input placeholder='gateway, core' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='enabled'
              render={({ field }) => (
                <FormItem className='flex flex-row items-center justify-between rounded-lg border p-3 shadow-xs'>
                  <div className='space-y-0.5'>
                    <FormLabel>Enable Device</FormLabel>
                    <p className='text-muted-foreground text-xs'>
                      Enable polling and active status monitoring.
                    </p>
                  </div>
                  <FormControl>
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                  </FormControl>
                </FormItem>
              )}
            />

            <DialogFooter>
              <Button
                type='button'
                variant='outline'
                onClick={() => {
                  setOpen(null)
                  setCurrentRow(null)
                }}
              >
                Cancel
              </Button>
              <Button type='submit' disabled={updateMutation.isPending}>
                {updateMutation.isPending ? 'Saving...' : isEdit ? 'Save Changes' : 'Create Device'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
