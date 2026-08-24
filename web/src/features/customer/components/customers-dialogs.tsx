import { CustomersMutateDrawer } from './customers-mutate-drawer'
import { CustomersDeleteDialog } from './customers-delete-dialog'
import { CustomersImportDialog } from './customers-import-dialog'
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
    </>
  )
}
