import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useCustomers } from './customers-provider'

function UnderConstructionNotice({ title }: { title: string }) {
  return (
    <>
      <DialogHeader>
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription>
          Form sedang dalam pengembangan — akan tersedia pada task berikutnya.
        </DialogDescription>
      </DialogHeader>
      <p className='text-sm text-muted-foreground'>
        Under construction.
      </p>
    </>
  )
}

export function CustomersDialogs() {
  const { open, setOpen, currentRow, setCurrentRow } = useCustomers()

  return (
    <>
      <Dialog
        key='customer-create'
        open={open === 'create'}
        onOpenChange={() => setOpen('create')}
      >
        <DialogContent className='sm:max-w-md'>
          <UnderConstructionNotice title='Tambah Pelanggan' />
        </DialogContent>
      </Dialog>

      {currentRow && (
        <>
          <Dialog
            key={`customer-update-${currentRow.id}`}
            open={open === 'update'}
            onOpenChange={() => {
              setOpen('update')
              setTimeout(() => {
                setCurrentRow(null)
              }, 500)
            }}
          >
            <DialogContent className='sm:max-w-md'>
              <UnderConstructionNotice title='Edit Pelanggan' />
            </DialogContent>
          </Dialog>

          <Dialog
            key='customer-delete'
            open={open === 'delete'}
            onOpenChange={() => {
              setOpen('delete')
              setTimeout(() => {
                setCurrentRow(null)
              }, 500)
            }}
          >
            <DialogContent className='sm:max-w-md'>
              <UnderConstructionNotice
                title={
                  currentRow
                    ? `Hapus Pelanggan: ${currentRow.name}`
                    : 'Hapus Pelanggan'
                }
              />
            </DialogContent>
          </Dialog>
        </>
      )}
    </>
  )
}
