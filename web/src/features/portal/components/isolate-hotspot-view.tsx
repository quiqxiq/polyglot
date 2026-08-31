import { useState } from 'react'
import {
  WifiOff,
  ShieldAlert,
  Phone,
  QrCode,
  CreditCard,
  Ticket,
  Sparkles,
  HelpCircle,
  LogIn,
} from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { toast } from 'sonner'

export function IsolateHotspotView() {
  const [username, setUsername] = useState('')
  const [voucherCode, setVoucherCode] = useState('')

  const handleRenew = (e: React.FormEvent) => {
    e.preventDefault()
    if (!username.trim()) {
      toast.error('Masukkan Username Hotspot Anda')
      return
    }
    toast.success('Mengecek status akun hotspot...')
  }

  const handleVoucherLogin = (e: React.FormEvent) => {
    e.preventDefault()
    if (!voucherCode.trim()) {
      toast.error('Masukkan Kode Voucher')
      return
    }
    toast.success('Memverifikasi kode voucher...')
  }

  return (
    <div className='min-h-screen bg-gradient-to-b from-background via-muted/30 to-background flex flex-col items-center justify-center p-4 sm:p-6 lg:p-8'>
      <div className='w-full max-w-xl space-y-6'>
        {/* Header Alert */}
        <div className='text-center space-y-3'>
          <div className='inline-flex items-center justify-center p-3 rounded-2xl bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20 shadow-sm animate-pulse'>
            <WifiOff className='h-8 w-8' />
          </div>
          <div className='space-y-1'>
            <Badge variant='outline' className='bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30 text-xs px-3 py-1 font-semibold gap-1.5'>
              <ShieldAlert className='h-3.5 w-3.5' /> STATUS: HOTSPOT DITANGGUHKAN / MASA AKTIF HABIS
            </Badge>
            <h1 className='text-2xl sm:text-3xl font-bold tracking-tight text-foreground'>
              Pemberitahuan Akses Hotspot
            </h1>
            <p className='text-sm text-muted-foreground max-w-md mx-auto'>
              Masa aktif paket hotspot Anda telah berakhir atau terdapat perpanjangan langganan bulanan yang belum terbayar.
            </p>
          </div>
        </div>

        {/* Card Pilihan Solusi */}
        <Card className='border-border/80 shadow-lg bg-card/95 backdrop-blur-sm'>
          <CardHeader className='pb-3 border-b'>
            <CardTitle className='text-base sm:text-lg'>
              Pilih Opsi Perpanjangan atau Akses
            </CardTitle>
            <CardDescription className='text-xs'>
              Silakan perpanjang akun hotspot langganan Anda atau gunakan kode voucher baru.
            </CardDescription>
          </CardHeader>

          <CardContent className='pt-5 space-y-4'>
            <Tabs defaultValue='renew' className='w-full'>
              <TabsList className='grid grid-cols-2 w-full h-9'>
                <TabsTrigger value='renew' className='text-xs gap-1.5'>
                  <CreditCard className='h-3.5 w-3.5' /> Perpanjang Langganan
                </TabsTrigger>
                <TabsTrigger value='voucher' className='text-xs gap-1.5'>
                  <Ticket className='h-3.5 w-3.5' /> Beli / Pakai Voucher
                </TabsTrigger>
              </TabsList>

              {/* Tab Perpanjang Akun Hotspot */}
              <TabsContent value='renew' className='space-y-4 pt-3'>
                <form onSubmit={handleRenew} className='space-y-3'>
                  <div className='space-y-1.5'>
                    <Label className='text-xs font-semibold'>Username Akun Hotspot</Label>
                    <Input
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      placeholder='Masukkan username hotspot'
                      className='h-9 text-xs font-mono'
                    />
                  </div>
                  <Button type='submit' className='w-full h-9 gap-1.5 bg-primary text-primary-foreground text-xs'>
                    <QrCode className='h-3.5 w-3.5' /> Bayar Perpanjangan via QRIS / VA
                  </Button>
                </form>

                <Alert className='bg-emerald-500/10 border-emerald-500/20 text-xs py-2.5'>
                  <Sparkles className='h-4 w-4 text-emerald-600 dark:text-emerald-400' />
                  <AlertTitle className='font-semibold text-emerald-900 dark:text-emerald-300'>
                    Aktivasi Instan
                  </AlertTitle>
                  <AlertDescription className='text-emerald-700 dark:text-emerald-400 mt-0.5'>
                    Sesi hotspot Anda akan langsung aktif kembali tanpa perlu login ulang begitu pembayaran terkonfirmasi.
                  </AlertDescription>
                </Alert>
              </TabsContent>

              {/* Tab Pakai / Beli Voucher */}
              <TabsContent value='voucher' className='space-y-4 pt-3'>
                <form onSubmit={handleVoucherLogin} className='space-y-3'>
                  <div className='space-y-1.5'>
                    <Label className='text-xs font-semibold'>Masukkan Kode Voucher Baru</Label>
                    <div className='flex gap-2'>
                      <Input
                        value={voucherCode}
                        onChange={(e) => setVoucherCode(e.target.value)}
                        placeholder='Contoh: VCH-98124'
                        className='h-9 text-xs font-mono uppercase'
                      />
                      <Button type='submit' size='sm' className='h-9 gap-1 text-xs'>
                        <LogIn className='h-3.5 w-3.5' /> Login
                      </Button>
                    </div>
                  </div>
                </form>

                {/* Paket Voucher Tersedia */}
                <div className='space-y-2 pt-2 border-t'>
                  <Label className='text-xs font-semibold text-muted-foreground'>Beli Voucher Online:</Label>
                  <div className='grid grid-cols-2 gap-2 text-xs'>
                    <div className='rounded-lg border p-2.5 bg-background hover:border-primary/50 transition-colors cursor-pointer flex flex-col justify-between'>
                      <div>
                        <p className='font-semibold text-foreground'>Voucher 12 Jam</p>
                        <p className='text-[11px] text-muted-foreground'>Unlimited Speed</p>
                      </div>
                      <p className='font-mono font-bold text-primary mt-2'>Rp 3.000</p>
                    </div>
                    <div className='rounded-lg border p-2.5 bg-background hover:border-primary/50 transition-colors cursor-pointer flex flex-col justify-between'>
                      <div>
                        <p className='font-semibold text-foreground'>Voucher 24 Jam</p>
                        <p className='text-[11px] text-muted-foreground'>Unlimited Speed</p>
                      </div>
                      <p className='font-mono font-bold text-primary mt-2'>Rp 5.000</p>
                    </div>
                    <div className='rounded-lg border p-2.5 bg-background hover:border-primary/50 transition-colors cursor-pointer flex flex-col justify-between'>
                      <div>
                        <p className='font-semibold text-foreground'>Voucher 7 Hari</p>
                        <p className='text-[11px] text-muted-foreground'>Unlimited Speed</p>
                      </div>
                      <p className='font-mono font-bold text-primary mt-2'>Rp 20.000</p>
                    </div>
                    <div className='rounded-lg border p-2.5 bg-background hover:border-primary/50 transition-colors cursor-pointer flex flex-col justify-between'>
                      <div>
                        <p className='font-semibold text-foreground'>Voucher 30 Hari</p>
                        <p className='text-[11px] text-muted-foreground'>Unlimited Speed</p>
                      </div>
                      <p className='font-mono font-bold text-primary mt-2'>Rp 50.000</p>
                    </div>
                  </div>
                </div>
              </TabsContent>
            </Tabs>
          </CardContent>

          <CardFooter className='border-t pt-4 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-muted-foreground'>
            <div className='flex items-center gap-1.5'>
              <HelpCircle className='h-4 w-4 text-blue-500' />
              <span>Butuh bantuan admin hotspot?</span>
            </div>
            <Button
              variant='outline'
              size='sm'
              className='gap-1.5 text-emerald-600 hover:text-emerald-700 border-emerald-500/30 hover:bg-emerald-500/10 w-full sm:w-auto'
              onClick={() => window.open('https://wa.me/', '_blank')}
            >
              <Phone className='h-3.5 w-3.5' /> Chat WhatsApp
            </Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  )
}
