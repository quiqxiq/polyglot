import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { Ticket } from 'lucide-react'
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
import { voucherGenerateSchema, type VoucherGenerateValues } from '../../data/schema'
import { USER_MODES, CHAR_SETS } from '../../data/constants'
import { useGenerateVouchersMutation } from '../../api/use-hotspot-vouchers'
import { useHotspotProfilesQuery } from '../../api/use-hotspot-profiles'
import { useHotspotServersQuery } from '../../api/use-hotspot-sessions'
import { useDeviceStore } from '@/stores/device-store'
import { useHotspot } from '../../context/hotspot-context'

export function VoucherGenerateDialog() {
  const { open, setOpen, currentProfile, setCurrentProfile, setPrintBatchComment, setPrintSingleUserId } =
    useHotspot()
  const { selectedDeviceId } = useDeviceStore()

  const isOpen = open === 'voucher-generate'

  const { data: profiles = [] } = useHotspotProfilesQuery(selectedDeviceId, isOpen)
  const { data: servers = [] } = useHotspotServersQuery(selectedDeviceId, isOpen)

  const generateMutation = useGenerateVouchersMutation()

  const form = useForm<VoucherGenerateValues>({
    resolver: zodResolver(voucherGenerateSchema),
    defaultValues: {
      count: 10,
      server: 'all',
      userType: 'vc',
      userLength: 4,
      prefix: '',
      characterSet: 'mix',
      profile: 'default',
      timeLimit: '3h',
      dataLimit: '',
      comment: '',
    },
  })

  useEffect(() => {
    if (isOpen) {
      const activeProf = currentProfile?.name || profiles[0]?.name || 'default'
      form.reset({
        count: 10,
        server: 'all',
        userType: 'vc',
        userLength: 4,
        prefix: '',
        characterSet: 'mix',
        profile: activeProf,
        timeLimit: currentProfile?.validity || '3h',
        dataLimit: '',
        comment: '',
      })
    }
  }, [isOpen, currentProfile, profiles, form])

  const onSubmit = async (values: VoucherGenerateValues) => {
    try {
      const res = await generateMutation.mutateAsync({
        deviceId: selectedDeviceId,
        profile: values.profile,
        count: values.count,
        userType: values.userType,
        userLength: values.userLength,
        prefix: values.prefix?.trim() || '',
        characterSet: values.characterSet,
        server: values.server === 'all' ? '' : values.server,
        timeLimit: values.timeLimit?.trim() || '',
        dataLimit: values.dataLimit?.trim() || '',
        comment: values.comment?.trim() || '',
      })

      const generatedCount = res.vouchers?.length || values.count
      toast.success(`Successfully generated ${generatedCount} vouchers!`)

      // Auto-prompt to print batch if vouchers have a batch comment tag
      const batchTag = res.vouchers?.[0]?.comment || values.comment
      if (batchTag) {
        setPrintBatchComment(batchTag)
        setPrintSingleUserId('')
        setOpen('voucher-print')
      } else {
        setOpen(null)
      }
      setCurrentProfile(null)
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : 'Failed to generate vouchers')
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={(val) => !val && setOpen(null)}>
      <DialogContent className='sm:max-w-[540px]'>
        <DialogHeader>
          <div className='flex items-center gap-2'>
            <Ticket className='size-5 text-primary' />
            <DialogTitle>Generate Voucher Batch</DialogTitle>
          </div>
          <DialogDescription>
            Generate multiple random voucher credentials on MikroTik with custom prefixes and validity.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4'>
            <Tabs defaultValue='general' className='w-full'>
              <TabsList className='grid w-full grid-cols-2'>
                <TabsTrigger value='general'>Batch & Credentials</TabsTrigger>
                <TabsTrigger value='limits'>Limits & Prefix</TabsTrigger>
              </TabsList>

              <TabsContent value='general' className='space-y-3 pt-2'>
                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='count'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Quantity (Vouchers) *</FormLabel>
                        <FormControl>
                          <Input
                            type='number'
                            min='1'
                            max='500'
                            {...field}
                            onChange={(e) => field.onChange(parseInt(e.target.value, 10) || 1)}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name='profile'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Profile Package *</FormLabel>
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

                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='userType'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>User Mode</FormLabel>
                        <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder='Select mode' />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {USER_MODES.map((m) => (
                              <SelectItem key={m.value} value={m.value}>
                                {m.label}
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
                    name='server'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Hotspot Server</FormLabel>
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
                </div>

                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='userLength'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Code Length</FormLabel>
                        <Select
                          onValueChange={(val) => field.onChange(parseInt(val, 10))}
                          defaultValue={String(field.value)}
                          value={String(field.value)}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder='4' />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {[3, 4, 5, 6, 7, 8, 10, 12].map((len) => (
                              <SelectItem key={len} value={String(len)}>
                                {len} Characters
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
                    name='characterSet'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Character Set</FormLabel>
                        <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder='mix' />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {CHAR_SETS.map((c) => (
                              <SelectItem key={c.value} value={c.value}>
                                {c.label}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </TabsContent>

              <TabsContent value='limits' className='space-y-3 pt-2'>
                <FormField
                  control={form.control}
                  name='prefix'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Voucher Prefix</FormLabel>
                      <FormControl>
                        <Input placeholder='e.g. VIP or NET' {...field} />
                      </FormControl>
                      <FormDescription className='text-[11px]'>
                        Optional prefix added before generated username (e.g. VIP-xxxx).
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className='grid grid-cols-2 gap-3'>
                  <FormField
                    control={form.control}
                    name='timeLimit'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Time Limit (Uptime)</FormLabel>
                        <FormControl>
                          <Input placeholder='e.g. 3h, 1d' {...field} />
                        </FormControl>
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
                          <Input placeholder='e.g. 1000M' {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <FormField
                  control={form.control}
                  name='comment'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Batch Comment Tag</FormLabel>
                      <FormControl>
                        <Input placeholder='Leave empty for auto-tag (e.g. vc-xxx)' {...field} />
                      </FormControl>
                      <FormDescription className='text-[11px]'>
                        Used to group vouchers for batch printing and tracking.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </TabsContent>
            </Tabs>

            <DialogFooter className='pt-2'>
              <Button type='button' variant='outline' onClick={() => setOpen(null)}>
                Cancel
              </Button>
              <Button type='submit' disabled={generateMutation.isPending}>
                {generateMutation.isPending ? 'Generating...' : 'Generate & Print'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
