import { CustomersMutateDrawer } from './customers-mutate-drawer'
import { CustomersDeleteDialog } from './customers-delete-dialog'
import { CustomersImportDialog } from './customers-import-dialog'
import { SubscriptionsCreateDialog } from '@/features/subscriptions/components/subscriptions-create-dialog'
import { useCustomers } from './customers-provider'

export function CustomersDialogs() {
  const { open, setOpen, currentRow } = useCustomers()
  return (
    <>
      <CustomersMutateDrawer
        key='customer-create'
        open={open === 'create'}
        onOpenChange={() => setOpen('create')}
        currentRow={null}
      />
      {currentRow && (
        <CustomersMutateDrawer
          key={`customer-update-${currentRow.id}`}
          open={open === 'update'}
          onOpenChange={() => setOpen('update')}
          currentRow={currentRow}
        />
      )}
      {currentRow && (
        <CustomersDeleteDialog
          key='customer-delete'
          open={open === 'delete'}
          onOpenChange={() => setOpen('delete')}
          currentRow={currentRow}
        />
      )}
      {open === 'import' && (
        <CustomersImportDialog
          key='customer-import'
          open={open === 'import'}
          onOpenChange={() => setOpen('import')}
        />
      )}
      {open === 'create-subscription' && currentRow && (
        <SubscriptionsCreateDialog
          key={`cust-create-sub-${currentRow.id}`}
          open={open === 'create-subscription'}
          onOpenChange={() => setOpen('create-subscription')}
          initialCustomerId={currentRow.id}
          lockCustomer
        />
      )}
    </>
  )
}
