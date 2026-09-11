import { useState } from 'react'
import { Loader2, LogIn, MessageCircle, ShieldCheck } from 'lucide-react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  PortalLoginRequest,
  RequestOTPRequest,
} from '@/gen/v1/portal_pb'
import {
  usePortalLoginMutation,
  useRequestOTPMutation,
} from '../api/use-portal'
import { setPortalToken } from '../api/portal-http'

interface PortalOtpLoginProps {
  /** Identifier awal (no. HP / kode portal / kode pelanggan). */
  identifier?: string
  /** Dipanggil setelah login sukses dengan token sesi portal. */
  onSuccess: (token: string) => void
  /** Judul kecil di atas form (opsional). */
  title?: string
  /** Mode ringkas untuk disematkan di halaman isolir. */
  compact?: boolean
}

/**
 * Form login portal dua langkah: minta OTP via WhatsApp lalu verifikasi.
 * OTP dikirim oleh worker antrean WhatsApp (F5-7) sehingga bisa lambat;
 * UI memberi tahu pelanggan menunggu pesan masuk.
 */
export function PortalOtpLogin({ identifier: preset, onSuccess, title, compact }: PortalOtpLoginProps) {
  const [identifier, setIdentifier] = useState(preset ?? '')
  const [otp, setOtp] = useState('')
  const [step, setStep] = useState<'identifier' | 'otp'>('identifier')

  const requestOTP = useRequestOTPMutation()
  const login = usePortalLoginMutation()

  const handleRequestOTP = (e: React.FormEvent) => {
    e.preventDefault()
    const value = identifier.trim()
    if (!value) {
      toast.error('Masukkan nomor HP atau kode portal Anda')
      return
    }
    requestOTP.mutate(new RequestOTPRequest({ identifier: value }), {
      onSuccess: (res) => {
        setStep('otp')
        toast.success(res.message || 'OTP dikirim via WhatsApp', {
          description: 'Kode berlaku beberapa menit. Tunggu pesan WhatsApp masuk.',
        })
      },
      onError: (err) => {
        toast.error('Gagal mengirim OTP', { description: err.message })
      },
    })
  }

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault()
    if (!otp.trim()) {
      toast.error('Masukkan kode OTP dari WhatsApp')
      return
    }
    login.mutate(new PortalLoginRequest({ identifier: identifier.trim(), otp: otp.trim() }), {
      onSuccess: (res) => {
        if (!res.token) {
          toast.error('Login gagal: token tidak diterima')
          return
        }
        setPortalToken(res.token)
        toast.success('Login portal berhasil')
        onSuccess(res.token)
      },
      onError: (err) => {
        toast.error('Login gagal', { description: err.message })
      },
    })
  }

  return (
    <form
      onSubmit={step === 'identifier' ? handleRequestOTP : handleLogin}
      className={compact ? 'space-y-3' : 'space-y-4'}
    >
      <div className='flex items-center gap-2 text-sm font-semibold'>
        <ShieldCheck className='h-4 w-4 text-emerald-600' />
        {title ?? 'Verifikasi Kepemilikan Akun'}
      </div>

      <div className='space-y-1.5'>
        <Label htmlFor='portal-identifier' className='text-xs'>
          Nomor HP / Kode Portal
        </Label>
        <Input
          id='portal-identifier'
          value={identifier}
          onChange={(e) => setIdentifier(e.target.value)}
          placeholder='Contoh: 081234567890'
          className='h-10 text-sm'
          disabled={step === 'otp'}
          autoComplete='tel'
        />
      </div>

      {step === 'otp' && (
        <div className='space-y-1.5 animate-in fade-in-50'>
          <Label htmlFor='portal-otp' className='text-xs'>
            Kode OTP WhatsApp
          </Label>
          <Input
            id='portal-otp'
            value={otp}
            onChange={(e) => setOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
            placeholder='6 digit kode'
            className='h-10 text-sm font-mono tracking-widest'
            inputMode='numeric'
            autoComplete='one-time-code'
          />
          <p className='text-[11px] text-muted-foreground flex items-center gap-1'>
            <MessageCircle className='h-3 w-3' /> Kode dikirim ke WhatsApp terdaftar.
          </p>
        </div>
      )}

      <div className='flex gap-2'>
        {step === 'identifier' ? (
          <Button type='submit' className='h-10 gap-1.5' disabled={requestOTP.isPending}>
            {requestOTP.isPending ? <Loader2 className='h-4 w-4 animate-spin' /> : <MessageCircle className='h-4 w-4' />}
            Kirim OTP
          </Button>
        ) : (
          <>
            <Button type='submit' className='h-10 gap-1.5' disabled={login.isPending}>
              {login.isPending ? <Loader2 className='h-4 w-4 animate-spin' /> : <LogIn className='h-4 w-4' />}
              Masuk
            </Button>
            <Button
              type='button'
              variant='outline'
              className='h-10'
              onClick={() => {
                setStep('identifier')
                setOtp('')
              }}
            >
              Ubah Nomor
            </Button>
          </>
        )}
      </div>
    </form>
  )
}
