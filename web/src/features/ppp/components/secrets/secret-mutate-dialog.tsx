import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Eye, EyeOff, KeyRound, RefreshCw, UserPlus } from 'lucide-react'
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
import { Switch } from '@/components/ui/switch'
import { useDeviceStore } from '@/stores/device-store'
import { usePPPProfilesQuery } from '../../api/use-ppp-profiles'
import {
  useCreatePPPSecretMutation,
  useUpdatePPPSecretMutation,
} from '../../api/use-ppp-secrets'
import { usePPP } from '../../context/ppp-context'
import { PPP_SERVICE_OPTIONS } from '../../data/constants'
import { pppSecretSchema, type PPPSecretFormValues } from '../../data/schema'

export function SecretMutateDialog() {
  const { open, setOpen, currentSecret, setCurrentSecret } = usePPP()
  const selectedDeviceId = useDeviceStore((state) => state.selectedDeviceId)
  const { data: profiles = [] } = usePPPProfilesQuery(selectedDeviceId)

  const [showPassword, setShowPassword] = useState(false)
  const isEdit = open === 'secret-update'
  const isOpen = open === 'secret-create' || isEdit

  const createMutation = useCreatePPPSecretMutation()
  const updateMutation = useUpdatePPPSecretMutation()

  const form = useForm<PPPSecretFormValues>({
    resolver: zodResolver(pppSecretSchema),
    defaultValues: {
      name: '',
      password: '',
      profile: 'default',
      service: 'pppoe',
      localAddress: '',
      remoteAddress: '',
      comment: '',
      callerId: '',
      disabled: false,
    },
  })

  useEffect(() => {
    if (isOpen) {
      if (isEdit && currentSecret) {
        form.reset({
          name: currentSecret.name,
          password: '',
          profile: currentSecret.profile || 'default',
          service: currentSecret.service || 'pppoe',
          localAddress: currentSecret.localAddress || '',
          remoteAddress: currentSecret.remoteAddress || '',
          comment: currentSecret.comment || '',
          callerId: currentSecret.callerId || '',
          disabled: currentSecret.disabled,
        })
      } else {
        form.reset({
          name: '',
          password: '',
          profile: 'default',
          service: 'pppoe',
          localAddress: '',
          remoteAddress: '',
          comment: '',
          callerId: '',
          disabled: false,
        })
      }
    }
  }, [isOpen, isEdit, currentSecret, form])

  const handleClose = () => {
    setOpen(null)
    setCurrentSecret(null)
    form.reset()
  }

  const generateRandomPassword = () => {
    const chars = 'abcdefghjkmnpqrstuvwxyz23456789'
    let pass = ''
    for (let i = 0; i < 8; i++) {
      pass += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    form.setValue('password', pass)
  }

  const onSubmit = async (values: PPPSecretFormValues) => {
    if (!selectedDeviceId) return

    if (isEdit && currentSecret) {
      await updateMutation.mutateAsync({
        deviceId: selectedDeviceId,
        rosId: currentSecret.id,
        name: values.name,
        password: values.password,
        profile: values.profile,
        service: values.service,
        localAddress: values.localAddress,
        remoteAddress: values.remoteAddress,
        comment: values.comment,
        callerId: values.callerId,
      })
    } else {
      await createMutation.mutateAsync({
        deviceId: selectedDeviceId,
        name: values.name,
        password: values.password,
        profile: values.profile,
        service: values.service,
        localAddress: values.localAddress,
        remoteAddress: values.remoteAddress,
        comment: values.comment,
        callerId: values.callerId,
        disabled: values.disabled,
      })
    }

    handleClose()
  }

  return (
    <Dialog open={isOpen} onOpenChange={(val) => !val && handleClose()}>
      <DialogContent className="max-w-xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <UserPlus className="h-5 w-5 text-primary" />
            {isEdit ? 'Edit PPPoE Secret' : 'Add New PPPoE Secret'}
          </DialogTitle>
          <DialogDescription>
            {isEdit
              ? `Update settings for secret "${currentSecret?.name}". Leave password blank to keep current.`
              : 'Create a new PPPoE subscriber account on the router.'}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Username / Name *</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g. user_01"
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
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex items-center justify-between">
                      <FormLabel>Password {isEdit ? '(Optional)' : '*'}</FormLabel>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="h-6 px-1.5 text-xs text-muted-foreground hover:text-foreground"
                        onClick={generateRandomPassword}
                      >
                        <RefreshCw className="mr-1 h-3 w-3" />
                        Generate
                      </Button>
                    </div>
                    <div className="relative">
                      <FormControl>
                        <Input
                          type={showPassword ? 'text' : 'password'}
                          placeholder={isEdit ? 'Leave blank to retain' : 'Password'}
                          {...field}
                          className="font-mono text-sm pr-9"
                        />
                      </FormControl>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="absolute right-0 top-0 h-full w-9 text-muted-foreground hover:text-foreground"
                        onClick={() => setShowPassword(!showPassword)}
                      >
                        {showPassword ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                        <span className="sr-only">Toggle password visibility</span>
                      </Button>
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="profile"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>PPP Profile *</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                      value={field.value}
                    >
                      <FormControl>
                        <SelectTrigger className="font-mono text-sm">
                          <SelectValue placeholder="Select profile" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {profiles.map((p) => (
                          <SelectItem key={p.id || p.name} value={p.name}>
                            {p.name} {p.rateLimit ? `(${p.rateLimit})` : ''}
                          </SelectItem>
                        ))}
                        {profiles.length === 0 && (
                          <SelectItem value="default">default</SelectItem>
                        )}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="service"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>PPP Service</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                      value={field.value}
                    >
                      <FormControl>
                        <SelectTrigger className="font-mono text-sm uppercase">
                          <SelectValue placeholder="Select service" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {PPP_SERVICE_OPTIONS.map((opt) => (
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
                name="remoteAddress"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Remote Address (Client IP)</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g. 192.168.10.2 (or pool name)"
                        {...field}
                        className="font-mono text-sm"
                      />
                    </FormControl>
                    <FormDescription className="text-[11px]">
                      Leave empty to inherit pool from profile.
                    </FormDescription>
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
                    <FormDescription className="text-[11px]">
                      Leave empty to inherit gateway from profile.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="callerId"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Caller ID (MAC Lock)</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g. 00:11:22:33:44:55"
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
                name="comment"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Comment / Note</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g. Rumah Bp. Ahmad / polyglot:12"
                        {...field}
                        className="text-sm"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            {!isEdit && (
              <FormField
                control={form.control}
                name="disabled"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                    <div className="space-y-0.5">
                      <FormLabel>Disable Account</FormLabel>
                      <FormDescription className="text-xs">
                        Create account in disabled state (subscriber cannot connect until enabled).
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            )}

            <DialogFooter className="pt-2">
              <Button type="button" variant="outline" onClick={handleClose}>
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={createMutation.isPending || updateMutation.isPending}
              >
                <KeyRound className="mr-2 h-4 w-4" />
                {isEdit ? 'Save Changes' : 'Create Secret'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
