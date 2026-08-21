import { useEffect } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2, ShieldCheck, UserCheck } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
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
import { useGetMe, useUpdateMe } from '../api/use-profile'

const profileFormSchema = z.object({
  fullName: z.string().min(2, 'Nama lengkap minimal 2 karakter.'),
  phoneNumber: z.string().min(8, 'Nomor HP/WhatsApp minimal 8 digit.'),
  email: z.string().email('Format email tidak valid.').or(z.literal('')),
  specialization: z.string().optional(),
})

type ProfileFormValues = z.infer<typeof profileFormSchema>

export function ProfileForm() {
  const { data: user, isLoading } = useGetMe()
  const updateMeMutation = useUpdateMe()

  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(profileFormSchema),
    defaultValues: {
      fullName: '',
      phoneNumber: '',
      email: '',
      specialization: '',
    },
    mode: 'onChange',
  })

  useEffect(() => {
    if (user) {
      form.reset({
        fullName: user.fullName || '',
        phoneNumber: user.phoneNumber || '',
        email: user.email || '',
        specialization: user.specialization || '',
      })
    }
  }, [user, form])

  function onSubmit(data: ProfileFormValues) {
    updateMeMutation.mutate({
      fullName: data.fullName,
      phoneNumber: data.phoneNumber,
      email: data.email || '',
      specialization: data.specialization || '',
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
        {/* Identitas Akun (Read-only) */}
        <div className='grid grid-cols-1 gap-4 rounded-lg border bg-muted/30 p-4 md:grid-cols-2'>
          <div>
            <span className='text-xs font-semibold text-muted-foreground uppercase'>
              Username
            </span>
            <div className='mt-1 flex items-center gap-2 font-mono text-sm font-medium'>
              <UserCheck className='h-4 w-4 text-primary' />
              {user?.username || '-'}
            </div>
          </div>
          <div>
            <span className='text-xs font-semibold text-muted-foreground uppercase'>
              Role & Hak Akses
            </span>
            <div className='mt-1 flex flex-wrap gap-1.5'>
              <Badge variant='secondary' className='capitalize'>
                <ShieldCheck className='mr-1 h-3 w-3 text-primary' />
                {user?.role || 'user'}
              </Badge>
              {user?.roles?.map(
                (r) =>
                  r !== user.role && (
                    <Badge key={r} variant='outline' className='capitalize'>
                      {r}
                    </Badge>
                  )
              )}
            </div>
          </div>
        </div>

        {/* Form Input Editable */}
        <FormField
          control={form.control}
          name='fullName'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Nama Lengkap</FormLabel>
              <FormControl>
                <Input placeholder='Contoh: Budi Santoso' {...field} />
              </FormControl>
              <FormDescription>
                Nama lengkap Anda yang akan tampil di sistem dan laporan.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className='grid grid-cols-1 gap-6 md:grid-cols-2'>
          <FormField
            control={form.control}
            name='phoneNumber'
            render={({ field }) => (
              <FormItem>
                <FormLabel>No. WhatsApp / HP</FormLabel>
                <FormControl>
                  <Input placeholder='Contoh: 081234567890' {...field} />
                </FormControl>
                <FormDescription>
                  Nomor aktif untuk notifikasi darurat dan operasional.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='email'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Email</FormLabel>
                <FormControl>
                  <Input placeholder='user@example.com' type='email' {...field} />
                </FormControl>
                <FormDescription>
                  Alamat email untuk pemulihan dan komunikasi resmi.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <FormField
          control={form.control}
          name='specialization'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Spesialisasi / Posisi</FormLabel>
              <FormControl>
                <Input placeholder='Contoh: Fiber Optic, Mikrotik NOC, CS Tier-1' {...field} />
              </FormControl>
              <FormDescription>
                Bidang keahlian atau divisi tempat Anda bertugas (opsional).
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button
          type='submit'
          disabled={updateMeMutation.isPending}
          className='min-w-32'
        >
          {updateMeMutation.isPending && (
            <Loader2 className='mr-2 h-4 w-4 animate-spin' />
          )}
          Simpan Profil
        </Button>
      </form>
    </Form>
  )
}
