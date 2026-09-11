import { useNavigate } from '@tanstack/react-router'
import { Wifi } from 'lucide-react'

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { PortalOtpLogin } from './portal-otp-login'

/** Halaman login portal mandiri (F5-2). */
export function PortalLoginPage() {
  const navigate = useNavigate()

  return (
    <div className='flex min-h-screen items-center justify-center bg-gradient-to-b from-background via-muted/30 to-background p-4'>
      <Card className='w-full max-w-md border-border/80 shadow-lg'>
        <CardHeader className='space-y-2 text-center'>
          <div className='mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 text-primary'>
            <Wifi className='h-6 w-6' />
          </div>
          <CardTitle className='text-lg'>Portal Pelanggan</CardTitle>
          <CardDescription>
            Masuk dengan kode OTP WhatsApp untuk melihat tagihan, membayar online, dan memantau
            status layanan Anda.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <PortalOtpLogin
            onSuccess={() => navigate({ to: '/portal' })}
            title='Login via OTP WhatsApp'
          />
        </CardContent>
      </Card>
    </div>
  )
}
