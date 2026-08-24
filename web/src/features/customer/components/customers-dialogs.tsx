import { CustomersMutateDrawer } from './customers-mutate-drawer'
import { CustomersDeleteDialog } from './customers-delete-dialog'
import { useCustomers } from './customers-provider'

export function CustomersDialogs() {
  const { currentRow } = useCustomers()

  return (
    <>
      <CustomersMutateDrawer key='customer-create' />

      {currentRow && (
        <>
          <CustomersMutateDrawer key={`customer-update-${currentRow.id}`} />
          <CustomersDeleteDialog key='customer-delete' />
        </>
      )}
    </>
  )
}
