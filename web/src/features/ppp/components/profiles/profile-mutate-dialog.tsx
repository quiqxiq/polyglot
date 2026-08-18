import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Shield, ShieldPlus, Zap } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
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
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useDeviceStore } from '@/stores/device-store'
import {
  useCreatePPPProfileMutation,
  useUpdatePPPProfileMutation,
} from '../../api/use-ppp-profiles'
import { usePPP } from '../../context/ppp-context'
import {
  PPP_ONLY_ONE_OPTIONS,
  PPP_RATE_LIMIT_PRESETS,
} from '../../data/constants'
import {
  pppProfileSchema,
  type PPPProfileFormValues,
} from '../../data/schema'

export function ProfileMutateDialog() {
  const { open, setOpen, currentProfile, setCurrentProfile } = usePPP()
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)

  const isEdit = open === 'profile-update'
  const isOpen = open === 'profile-create' || isEdit

  const createMutation = useCreatePPPProfileMutation()
  const updateMutation = useUpdatePPPProfileMutation()

  const form = useForm<PPPProfileFormValues>({
    resolver: zodResolver(pppProfileSchema),
    defaultValues: {
      name: '',
      rateLimit: '',
      localAddress: '',
      remoteAddress: '',
      dnsServer: '',
      parentQueue: '',
      addressList: '',
      comment: '',
      sharedUsers: '1',
      onlyOne: 'default',
    },
  })

  useEffect(() => {
    if (isOpen) {
      if (isEdit && currentProfile) {
        form.reset({
          name: currentProfile.name,
          rateLimit: currentProfile.rateLimit || '',
          localAddress: currentProfile.localAddress || '',
          remoteAddress: currentProfile.remoteAddress || '',
          dnsServer: currentProfile.dnsServer || '',
          parentQueue: currentProfile.parentQueue || '',
          addressList: currentProfile.addressList || '',
          comment: currentProfile.comment || '',
          sharedUsers: currentProfile.sharedUsers || '1',
          onlyOne: (currentProfile.onlyOne as 'default' | 'yes' | 'no') || 'default',
        })
      } else {
        form.reset({
          name: '',
          rateLimit: '10M/10M',
          localAddress: '',
          remoteAddress: '',
          dnsServer: '8.8.8.8,1.1.1.1',
          parentQueue: '',
          addressList: '',
          comment: '',
          sharedUsers: '1',
          onlyOne: 'yes',
        })
      }
    }
  }, [isOpen, isEdit, currentProfile, form])

  const handleClose = () => {
    setOpen(null)
    setCurrentProfile(null)
    form.reset()
  }

  const onSubmit = async (values: PPPProfileFormValues) => {
    if (!selectedDeviceId) return

    if (isEdit && currentProfile) {
      await updateMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: currentProfile.id,
        name: values.name,
        rateLimit: values.rateLimit,
        localAddress: values.localAddress,
        remoteAddress: values.remoteAddress,
        dnsServer: values.dnsServer,
        parentQueue: values.parentQueue,
        addressList: values.addressList,
        comment: values.comment,
        sharedUsers: values.sharedUsers,
        onlyOne: values.onlyOne,
      })
    } else {
      await createMutation.mutateAsync({
        deviceId: selectedDeviceId,
        name: values.name,
        rateLimit: values.rateLimit,
        localAddress: values.localAddress,
        remoteAddress: values.remoteAddress,
        dnsServer: values.dnsServer,
        parentQueue: values.parentQueue,
        addressList: values.addressList,
        comment: values.comment,
        sharedUsers: values.sharedUsers,
        onlyOne: values.onlyOne,
      })
    }

    handleClose()
  }

  return (
    <Dialog open={isOpen} onOpenChange={(val) => !val && handleClose()}>
      <DialogContent className="max-w-xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ShieldPlus className="h-5 w-5 text-primary" />
            {isEdit ? 'Edit PPP Profile' : 'Add New PPP Profile'}
          </DialogTitle>
          <DialogDescription>
            {isEdit
              ? `Update bandwidth shaping, IP assignment, and queue rules for "${currentProfile?.name}".`
              : 'Configure a new bandwidth plan and IP profile for PPPoE subscribers.'}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Profile Name *</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="e.g. 10Mbps / Paket_Rumah / isolir"
                      {...field}
                      className="font-mono text-sm font-medium"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="rateLimit"
              render={({ field }) => (
                <FormItem className="space-y-2">
                  <div className="flex items-center justify-between">
                    <FormLabel>Rate Limit (Upload/Download Rx/Tx)</FormLabel>
                    <span className="text-xs text-muted-foreground font-mono">
                      Format: rx/tx (e.g. 10M/10M)
                    </span>
                  </div>
                  <FormControl>
                    <Input
                      placeholder="e.g. 10M/10M (empty = unlimited)"
                      {...field}
                      className="font-mono text-sm font-semibold"
                    />
                  </FormControl>
                  <div className="flex flex-wrap gap-1.5 pt-1">
                    <span className="text-[11px] text-muted-foreground flex items-center mr-1">
                      <Zap className="mr-1 h-3 w-3 text-amber-500" /> Presets:
                    </span>
                    {PPP_RATE_LIMIT_PRESETS.map((preset) => (
                      <Badge
                        key={preset.value}
                        variant="secondary"
                        className="cursor-pointer text-[11px] font-mono hover:bg-primary/20 hover:text-primary transition-colors"
                        onClick={() => form.setValue('rateLimit', preset.value)}
                      >
                        {preset.value}
                      </Badge>
                    ))}
                  </div>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="remoteAddress"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Remote Address / IP Pool</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g. dhcp_pool1 or 192.168.10.2"
                        {...field}
                        className="font-mono text-sm"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="localAddress"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Local Address (Gateway IP)</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g. 192.168.10.1"
                        {...field}
                        className="font-mono text-sm"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="dnsServer"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>DNS Servers</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g. 8.8.8.8,1.1.1.1"
                        {...field}
                        className="font-mono text-sm"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="onlyOne"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Only-One (Session Policy)</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                      value={field.value}
                    >
                      <FormControl>
                        <SelectTrigger className="text-sm">
                          <SelectValue placeholder="Select policy" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {PPP_ONLY_ONE_OPTIONS.map((opt) => (
                          <SelectItem key={opt.value} value={opt.value}>
                            {opt.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="addressList"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Address List</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g. PPPOE_SUBSCRIBERS"
                        {...field}
                        className="font-mono text-sm"
                      />
                    </FormControl>
                    <FormDescription className="text-[11px]">
                      Auto-register connected client IP to firewall address-list.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="parentQueue"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Parent Queue</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g. TOTAL_BANDWIDTH"
                        {...field}
                        className="font-mono text-sm"
                      />
                    </FormControl>
                    <FormDescription className="text-[11px]">
                      Assign subscriber simple queues to parent hierarchy.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="comment"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Comment / Note</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="e.g. Paket Rumahan 10Mbps / polyglot"
                      {...field}
                      className="text-sm"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter className="pt-2">
              <Button type="button" variant="outline" onClick={handleClose}>
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={createMutation.isPending || updateMutation.isPending}
              >
                <Shield className="mr-2 h-4 w-4" />
                {isEdit ? 'Save Changes' : 'Create Profile'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
