import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link } from '@tanstack/react-router'
import { CheckCircle2, Loader2, MapPin, Package } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { SelectDropdown } from '@/components/select-dropdown'
import { SubmitRegistrationRequest } from '@/gen/v1/registration_pb'
import {
  usePublicPlansQuery,
  useSubmitRegistrationMutation,
} from '../api/use-registration'
import {
  submitRegistrationSchema,
  type SubmitRegistrationValues,
} from '../data/schema'

/**
 * Halaman publik pendaftaran pasang baru (/register) — diakses calon
 * pelanggan tanpa login. Setelah submit, tampilkan token untuk lacak status.
 */
export function RegisterPublicPage() {
  const { data: plans = [], isLoading: plansLoading } = usePublicPlansQuery()
  const submitMutation = useSubmitRegistrationMutation()
  const [submittedToken, setSubmittedToken] = useState<string>('')
  const [submittedNo, setSubmittedNo] = useState<string>('')
  const [coords, setCoords] = useState<{ lat: number; lng: number } | null>(null)
  const [locating, setLocating] = useState(false)

  const form = useForm<SubmitRegistrationValues>({
    resolver: zodResolver(submitRegistrationSchema) as never,
    defaultValues: {
      fullName: '',
      phone: '',
      email: '',
      address: '',
      planId: '',
      notes: '',
      nik: '',
    },
  })

  const planOptions = plans.map((p) => ({
    label: `${p.name} — Rp${Number(p.price).toLocaleString('id-ID')}/bln`,
    value: p.id,
  }))

  const handleUseMyLocation = () => {
    if (!navigator.geolocation) {
      toast.error('Browser tidak mendukung geolokasi')
      return
    }
    setLocating(true)
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setCoords({ lat: pos.coords.latitude, lng: pos.coords.longitude })
        setLocating(false)
        toast.success('Lokasi berhasil diambil')
      },
      () => {
        setLocating(false)
        toast.error('Gagal mengambil lokasi. Izinkan akses lokasi.')
      },
      { enableHighAccuracy: true, timeout: 10000 },
    )
  }

  const handleSubmit = async (values: SubmitRegistrationValues) => {
    try {
      const res = await submitMutation.mutateAsync(
        new SubmitRegistrationRequest({
          fullName: values.fullName,
          phone: values.phone,
          email: values.email || '',
          address: values.address,
          notes: values.notes || '',
          planId: values.planId,
          nik: values.nik || '',
          latitude: coords?.lat ?? 0,
          longitude: coords?.lng ?? 0,
          hasCoordinates: coords !== null,
        }),
      )
      setSubmittedToken(res.registration?.publicToken || '')
      setSubmittedNo(res.registration?.registrationNo || '')
      toast.success('Pendaftaran berhasil dikirim!')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal mengirim pendaftaran')
    }
  }

  if (submittedToken) {
    return (
      <div className='flex min-h-svh items-center justify-center bg-muted/30 p-4'>
        <Card className='w-full max-w-lg text-center'>
          <CardHeader>
            <div className='mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-600'>
              <CheckCircle2 className='h-8 w-8' />
            </div>
            <CardTitle className='text-xl'>Pendaftaran Terkirim</CardTitle>
            <CardDescription>
              Nomor registrasi{' '}
              <span className='font-mono font-semibold'>{submittedNo}</span>.
              Simpan tautan di bawah untuk memantau status pemasangan.
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-3'>
            <div className='rounded-md bg-muted p-3 text-left text-xs'>
              <p className='text-muted-foreground'>Token lacak status:</p>
              <p className='mt-1 break-all font-mono font-semibold'>
                {submittedToken}
              </p>
            </div>
            <Button asChild className='w-full'>
              <Link to='/register/status' search={{ token: submittedToken }}>
                Lacak Status Pendaftaran
              </Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className='flex min-h-svh items-center justify-center bg-muted/30 p-4'>
      <Card className='w-full max-w-xl'>
        <CardHeader>
          <div className='flex items-center gap-2'>
            <Package className='h-5 w-5 text-primary' />
            <CardTitle>Pendaftaran Pasang Baru</CardTitle>
          </div>
          <CardDescription>
            Isi data di bawah untuk berlangganan internet. Tim kami akan
            menghubungi Anda untuk survei lokasi.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(handleSubmit)}
              className='flex flex-col gap-4'
            >
              <FormField
                control={form.control}
                name='fullName'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Nama Lengkap</FormLabel>
                    <FormControl>
                      <Input placeholder='Nama sesuai KTP' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
                <FormField
                  control={form.control}
                  name='phone'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Nomor WhatsApp</FormLabel>
                      <FormControl>
                        <Input placeholder='08xxxxxxxxxx' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='nik'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>NIK (opsional)</FormLabel>
                      <FormControl>
                        <Input placeholder='16 digit NIK' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <FormField
                control={form.control}
                name='email'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email (opsional)</FormLabel>
                    <FormControl>
                      <Input type='email' placeholder='nama@email.com' {...field} />
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
                    <FormLabel>Alamat Pemasangan</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder='Alamat lengkap + patokan lokasi'
                        rows={3}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <div className='flex flex-wrap items-center gap-2 rounded-md border border-dashed p-3'>
                <MapPin className='h-4 w-4 text-muted-foreground' />
                {coords ? (
                  <span className='font-mono text-xs'>
                    {coords.lat.toFixed(6)}, {coords.lng.toFixed(6)}
                  </span>
                ) : (
                  <span className='text-xs text-muted-foreground'>
                    Tandai titik lokasi rumah (opsional, membantu tim survei)
                  </span>
                )}
                <Button
                  type='button'
                  size='sm'
                  variant='outline'
                  className='ml-auto'
                  onClick={handleUseMyLocation}
                  disabled={locating}
                >
                  {locating ? 'Mengambil...' : 'Gunakan Lokasi Saya'}
                </Button>
              </div>

              <FormField
                control={form.control}
                name='planId'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Paket Layanan</FormLabel>
                    <FormControl>
                      <SelectDropdown
                        isControlled
                        defaultValue={field.value || undefined}
                        onValueChange={field.onChange}
                        placeholder={
                          plansLoading ? 'Memuat paket...' : 'Pilih paket'
                        }
                        items={planOptions}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='notes'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Catatan (opsional)</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder='Contoh: rumah pagar hijau sebelah warung'
                        rows={2}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <Button
                type='submit'
                disabled={submitMutation.isPending}
                className='w-full'
              >
                {submitMutation.isPending ? (
                  <>
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' /> Mengirim...
                  </>
                ) : (
                  'Kirim Pendaftaran'
                )}
              </Button>
              <p className='text-center text-xs text-muted-foreground'>
                Sudah mendaftar?{' '}
                <Link to='/register/status' search={{ token: '' }} className='text-primary underline'>
                  Lacak status pendaftaran
                </Link>
              </p>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  )
}
