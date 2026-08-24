import { useEffect, useMemo } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { SelectDropdown } from '@/components/select-dropdown'
import { Customer, CreateCustomerRequest, UpdateCustomerRequest } from '@/gen/v1/customer_pb'
import { customerFormSchema, type CustomerFormValues } from '../data/schema'
import { CUSTOMER_STATUS_META } from '../data/constants'
import {
  useCreateCustomerMutation,
  useUpdateCustomerMutation,
} from '../api/use-customer'

const STATUS_OPTIONS = Object.entries(CUSTOMER_STATUS_META).map(
  ([value, meta]) => ({ label: meta.label, value })
)

type CustomerStatus = CustomerFormValues['status']

function toFormStatus(status: string): CustomerStatus {
  return status in CUSTOMER_STATUS_META ? (status as CustomerStatus) : 'ACTIVE'
}

function defaultFormValues(currentRow: Customer | null): CustomerFormValues {
  if (!currentRow) {
    return {
      name: '',
      phone: '',
      email: '',
      address: '',
      latitude: 0,
      longitude: 0,
      hasCoordinates: false,
      status: 'ACTIVE',
      notes: '',
    }
  }
  return {
    name: currentRow.name,
    phone: currentRow.phone,
    email: currentRow.email ?? '',
    address: currentRow.address ?? '',
    latitude: currentRow.latitude ?? 0,
    longitude: currentRow.longitude ?? 0,
    hasCoordinates: currentRow.hasCoordinates ?? false,
    status: toFormStatus(currentRow.status),
    notes: currentRow.notes ?? '',
  }
}

type CustomersMutateDrawerProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow: Customer | null
}

export function CustomersMutateDrawer({
  open,
  onOpenChange,
  currentRow,
}: CustomersMutateDrawerProps) {
  const isOpen = open
  const isUpdate = !!currentRow
  const createMutation = useCreateCustomerMutation()
  const updateMutation = useUpdateCustomerMutation()

  // TFieldValues/TTransformedValues are inferred from the resolver: zod
  // `.default()` fields are optional on input but guaranteed on output.
  const form = useForm({
    resolver: zodResolver(customerFormSchema),
    defaultValues: defaultFormValues(currentRow),
  })

  // The drawer stays mounted while the provider switches create/update rows;
  // useForm only reads defaultValues once at mount, so resync explicitly.
  const initialValues = useMemo(() => defaultFormValues(currentRow), [currentRow])
  useEffect(() => {
    if (isOpen) form.reset(initialValues)
  }, [isOpen, initialValues, form])

  const hasCoordinates = useWatch({ control: form.control, name: 'hasCoordinates' })
  const showCoordinates = hasCoordinates ?? false

  const onSubmit = async (values: CustomerFormValues) => {
    try {
      if (isUpdate && currentRow) {
        await updateMutation.mutateAsync(
          new UpdateCustomerRequest({
            customer: new Customer({ ...values, id: currentRow.id }),
          })
        )
        toast.success('Pelanggan diperbarui')
      } else {
        await createMutation.mutateAsync(
          new CreateCustomerRequest({
            customer: new Customer({ ...values, email: values.email ?? '' }),
          })
        )
        toast.success('Pelanggan disimpan')
      }
      onOpenChange(false)
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : 'Gagal menyimpan pelanggan'
      toast.error(errorMessage)
    }
  }

  return (
    <Sheet
      open={open}
      onOpenChange={(v) => {
        onOpenChange(v)
        form.reset(defaultFormValues(currentRow))
      }}
    >
      <SheetContent className='flex flex-col'>
        <SheetHeader className='text-start'>
          <SheetTitle>{isUpdate ? 'Edit Pelanggan' : 'Tambah Pelanggan'}</SheetTitle>
          <SheetDescription>
            {isUpdate
              ? 'Perbarui data pelanggan yang sudah terdaftar.'
              : 'Daftarkan pelanggan baru dengan mengisi info berikut.'}{' '}
            Klik simpan setelah selesai.
          </SheetDescription>
        </SheetHeader>
        <Form {...form}>
          <form
            id='customers-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className='flex-1 space-y-6 overflow-y-auto px-4'
          >
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Nama</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder='Nama lengkap pelanggan' />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='phone'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Nomor HP</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder='081234567890' />
                  </FormControl>
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
                    <Input {...field} type='email' placeholder='nama@email.com' />
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
                  <FormLabel>Alamat</FormLabel>
                  <FormControl>
                    <Textarea
                      {...field}
                      value={field.value ?? ''}
                      placeholder='Alamat pemasangan'
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='status'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Status</FormLabel>
                  <SelectDropdown
                    defaultValue={field.value}
                    onValueChange={field.onChange}
                    placeholder='Pilih status'
                    items={STATUS_OPTIONS}
                  />
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='hasCoordinates'
              render={({ field }) => (
                <FormItem className='flex flex-row items-center justify-between rounded-lg border p-3 shadow-xs'>
                  <div className='space-y-0.5'>
                    <FormLabel>Pakai Koordinat</FormLabel>
                    <p className='text-muted-foreground text-xs'>
                      Simpan titik lokasi pemasangan untuk ditampilkan di peta.
                    </p>
                  </div>
                  <FormControl>
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                  </FormControl>
                </FormItem>
              )}
            />
            {showCoordinates && (
              <div className='grid grid-cols-2 gap-4'>
                <FormField
                  control={form.control}
                  name='latitude'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Latitude</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          step='any'
                          placeholder='-6.2088'
                          {...field}
                          onChange={(e) => field.onChange(e.target.valueAsNumber)}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name='longitude'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Longitude</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          step='any'
                          placeholder='106.8456'
                          {...field}
                          onChange={(e) => field.onChange(e.target.valueAsNumber)}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            )}
            <FormField
              control={form.control}
              name='notes'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Catatan</FormLabel>
                  <FormControl>
                    <Textarea
                      {...field}
                      value={field.value ?? ''}
                      placeholder='Catatan tambahan (opsional)'
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </form>
        </Form>
        <SheetFooter className='gap-2'>
          <SheetClose asChild>
            <Button variant='outline'>Cancel</Button>
          </SheetClose>
          <Button
            form='customers-form'
            type='submit'
            disabled={createMutation.isPending || updateMutation.isPending}
          >
            Save
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
