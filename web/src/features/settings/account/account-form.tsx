import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { KeyRound, Loader2, Lock } from 'lucide-react'
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
import { useChangePassword } from '../api/use-profile'

const accountFormSchema = z
  .object({
    oldPassword: z.string().min(1, 'Password lama harus diisi.'),
    newPassword: z
      .string()
      .min(8, 'Password baru minimal harus 8 karakter.')
      .max(64, 'Password baru maksimal 64 karakter.'),
    confirmPassword: z.string().min(1, 'Konfirmasi password baru harus diisi.'),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: 'Konfirmasi password tidak cocok dengan password baru.',
    path: ['confirmPassword'],
  })

type AccountFormValues = z.infer<typeof accountFormSchema>

export function AccountForm() {
  const changePasswordMutation = useChangePassword()

  const form = useForm<AccountFormValues>({
    resolver: zodResolver(accountFormSchema),
    defaultValues: {
      oldPassword: '',
      newPassword: '',
      confirmPassword: '',
    },
  })

  function onSubmit(data: AccountFormValues) {
    changePasswordMutation.mutate(
      {
        oldPassword: data.oldPassword,
        newPassword: data.newPassword,
      },
      {
        onSuccess: () => {
          form.reset()
        },
      }
    )
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className='max-w-xl space-y-6'>
        <div className='flex items-center gap-2 text-sm text-muted-foreground'>
          <Lock className='h-4 w-4 text-primary' />
          Pastikan password baru Anda kuat dan tidak mudah ditebak.
        </div>

        <FormField
          control={form.control}
          name='oldPassword'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Password Saat Ini</FormLabel>
              <FormControl>
                <Input
                  type='password'
                  placeholder='Masukkan password saat ini'
                  {...field}
                />
              </FormControl>
              <FormDescription>
                Dibutuhkan untuk memverifikasi kepemilikan akun Anda.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='newPassword'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Password Baru</FormLabel>
              <FormControl>
                <Input
                  type='password'
                  placeholder='Minimal 8 karakter'
                  {...field}
                />
              </FormControl>
              <FormDescription>
                Gunakan kombinasi huruf, angka, dan simbol.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='confirmPassword'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Konfirmasi Password Baru</FormLabel>
              <FormControl>
                <Input
                  type='password'
                  placeholder='Ulangi password baru'
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button
          type='submit'
          disabled={changePasswordMutation.isPending}
          className='min-w-36'
        >
          {changePasswordMutation.isPending ? (
            <Loader2 className='mr-2 h-4 w-4 animate-spin' />
          ) : (
            <KeyRound className='mr-2 h-4 w-4' />
          )}
          Ubah Password
        </Button>
      </form>
    </Form>
  )
}
