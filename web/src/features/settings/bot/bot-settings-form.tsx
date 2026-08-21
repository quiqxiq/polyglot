import { useEffect } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Clock, Cpu, Loader2, Save, ShieldAlert, Users } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
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
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { useGetBotSettings, useUpdateBotSettings } from '../api/use-settings'

const botSettingsSchema = z.object({
  burstLimit: z.coerce.number().min(1, 'Minimal 1 pesan.').max(50),
  burstWindowSecs: z.coerce.number().min(1, 'Minimal 1 detik.').max(300),
  mute1HSecs: z.coerce.number().min(60, 'Minimal 60 detik.').max(86400),
  ban24HSecs: z.coerce.number().min(300, 'Minimal 300 detik.').max(604800),
  dailyChatLimit: z.coerce.number().min(1, 'Minimal 1 chat.').max(1000),
  sessionTimeoutMinutes: z.coerce.number().min(1, 'Minimal 1 menit.').max(1440),
  slidingWindowSize: z.coerce.number().min(2, 'Minimal 2 pesan.').max(50),
  llmMaxOutputTokens: z.coerce.number().min(128, 'Minimal 128 token.').max(8192),
  whitelistAllStaff: z.boolean(),
  customWhitelistPhones: z.string().optional(),
})

type BotSettingsValues = z.infer<typeof botSettingsSchema>

export function BotSettingsForm() {
  const { data: settings, isLoading } = useGetBotSettings()
  const updateMutation = useUpdateBotSettings()

  const form = useForm<BotSettingsValues>({
    resolver: zodResolver(botSettingsSchema) as any,
    defaultValues: {
      burstLimit: 3,
      burstWindowSecs: 5,
      mute1HSecs: 3600,
      ban24HSecs: 86400,
      dailyChatLimit: 10,
      sessionTimeoutMinutes: 30,
      slidingWindowSize: 10,
      llmMaxOutputTokens: 1024,
      whitelistAllStaff: true,
      customWhitelistPhones: '',
    },
  })

  useEffect(() => {
    if (settings) {
      form.reset({
        burstLimit: settings.burstLimit || 3,
        burstWindowSecs: settings.burstWindowSecs || 5,
        mute1HSecs: settings.mute1hSecs || 3600,
        ban24HSecs: settings.ban24hSecs || 86400,
        dailyChatLimit: settings.dailyChatLimit || 10,
        sessionTimeoutMinutes: settings.sessionTimeoutMinutes || 30,
        slidingWindowSize: settings.slidingWindowSize || 10,
        llmMaxOutputTokens: settings.llmMaxOutputTokens || 1024,
        whitelistAllStaff: settings.whitelistAllStaff ?? true,
        customWhitelistPhones: settings.customWhitelistPhones || '',
      })
    }
  }, [settings, form])

  function onSubmit(data: BotSettingsValues) {
    updateMutation.mutate({
      burstLimit: data.burstLimit,
      burstWindowSecs: data.burstWindowSecs,
      mute1hSecs: data.mute1HSecs,
      ban24hSecs: data.ban24HSecs,
      dailyChatLimit: data.dailyChatLimit,
      sessionTimeoutMinutes: data.sessionTimeoutMinutes,
      slidingWindowSize: data.slidingWindowSize,
      llmMaxOutputTokens: data.llmMaxOutputTokens,
      whitelistAllStaff: data.whitelistAllStaff,
      customWhitelistPhones: data.customWhitelistPhones || '',
    })
  }

  if (isLoading) {
    return (
      <div className='flex h-48 items-center justify-center'>
        <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
      </div>
    )
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-6'>
        {/* Anti-Spam & Burst Section */}
        <Card>
          <CardHeader className='pb-3'>
            <CardTitle className='flex items-center gap-2 text-base'>
              <ShieldAlert className='h-5 w-5 text-destructive' />
              Anti-Spam & Burst Protection
            </CardTitle>
            <CardDescription>
              Mencegah bot WhatsApp kebanjiran pesan beruntun dari pengguna atau bot liar.
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='burstLimit'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Batas Pesan Burst</FormLabel>
                    <FormControl>
                      <Input type='number' min={1} max={50} {...field} />
                    </FormControl>
                    <FormDescription>
                      Jumlah pesan maksimal dalam rentang waktu singkat (default: 3).
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='burstWindowSecs'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Rentang Deteksi Burst (Detik)</FormLabel>
                    <FormControl>
                      <Input type='number' min={1} max={300} {...field} />
                    </FormControl>
                    <FormDescription>
                      Rentang waktu deteksi burst spam (default: 5 detik).
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='mute1HSecs'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Durasi Mute Level 1 (Detik)</FormLabel>
                    <FormControl>
                      <Input type='number' min={60} {...field} />
                    </FormControl>
                    <FormDescription>
                      Mute sementara saat terdeteksi spam pertama kali (3600 = 1 jam).
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='ban24HSecs'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Durasi Ban Level 2 (Detik)</FormLabel>
                    <FormControl>
                      <Input type='number' min={300} {...field} />
                    </FormControl>
                    <FormDescription>
                      Blokir saat spam berulang &gt;= 3 kali (86400 = 24 jam).
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </CardContent>
        </Card>

        {/* Kuota & Sesi Interaksi */}
        <Card>
          <CardHeader className='pb-3'>
            <CardTitle className='flex items-center gap-2 text-base'>
              <Clock className='h-5 w-5 text-primary' />
              Kuota Chat & Sesi Interaksi
            </CardTitle>
            <CardDescription>
              Mengatur batasan percakapan gratis harian dan durasi memori sesi pelanggan.
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='dailyChatLimit'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Batas Chat Harian / Nomor</FormLabel>
                    <FormControl>
                      <Input type='number' min={1} max={1000} {...field} />
                    </FormControl>
                    <FormDescription>
                      Batas respon AI otomatis per pelanggan per hari (default: 10).
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='sessionTimeoutMinutes'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Timeout Sesi (Menit)</FormLabel>
                    <FormControl>
                      <Input type='number' min={1} max={1440} {...field} />
                    </FormControl>
                    <FormDescription>
                      Masa berlaku riwayat obrolan sebelum dianggap percakapan baru (default: 30).
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </CardContent>
        </Card>

        {/* Konteks LLM */}
        <Card>
          <CardHeader className='pb-3'>
            <CardTitle className='flex items-center gap-2 text-base'>
              <Cpu className='h-5 w-5 text-amber-500' />
              Parameter Konteks LLM
            </CardTitle>
            <CardDescription>
              Menentukan seberapa banyak riwayat percakapan dan token yang dikirim ke model AI.
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-4'>
            <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='slidingWindowSize'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Sliding Window Context (Pesan)</FormLabel>
                    <FormControl>
                      <Input type='number' min={2} max={50} {...field} />
                    </FormControl>
                    <FormDescription>
                      Jumlah pesan terakhir yang disertakan ke dalam prompt konteks (default: 10).
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='llmMaxOutputTokens'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Maksimal Output Tokens</FormLabel>
                    <FormControl>
                      <Input type='number' min={128} max={8192} {...field} />
                    </FormControl>
                    <FormDescription>
                      Batas maksimal token jawaban model AI (default: 1024).
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </CardContent>
        </Card>

        {/* Kebijakan Whitelist */}
        <Card>
          <CardHeader className='pb-3'>
            <CardTitle className='flex items-center gap-2 text-base'>
              <Users className='h-5 w-5 text-emerald-500' />
              Kebijakan Whitelist (Bebas Limit)
            </CardTitle>
            <CardDescription>
              Nomor yang terdaftar di whitelist tidak akan terkena batasan anti-spam maupun kuota harian.
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-4'>
            <FormField
              control={form.control}
              name='whitelistAllStaff'
              render={({ field }) => (
                <FormItem className='flex flex-row items-center justify-between rounded-lg border p-3 shadow-xs'>
                  <div className='space-y-0.5'>
                    <FormLabel className='text-sm font-semibold'>
                      Otomatis Whitelist Semua Staf (Tabel Users)
                    </FormLabel>
                    <FormDescription>
                      Semua akun user/teknisi/admin yang memiliki nomor HP di database otomatis bebas limit.
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

            <FormField
              control={form.control}
              name='customWhitelistPhones'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Nomor Whitelist Khusus (Tambahan)</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder='Contoh: 628123456789, 081987654321'
                      rows={3}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    Masukkan nomor WhatsApp tambahan yang ingin dibebaskan dari limit (pisahkan dengan tanda koma).
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </CardContent>
        </Card>

        <Button
          type='submit'
          disabled={updateMutation.isPending}
          className='min-w-44'
        >
          {updateMutation.isPending ? (
            <Loader2 className='mr-2 h-4 w-4 animate-spin' />
          ) : (
            <Save className='mr-2 h-4 w-4' />
          )}
          Simpan Konfigurasi
        </Button>
      </form>
    </Form>
  )
}
