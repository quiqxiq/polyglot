import { PlansMutateDialog } from './plans-mutate-dialog'
import { PlansDeleteDialog } from './plans-delete-dialog'
import { usePlans } from './plans-provider'

export function PlansDialogs() {
  const { currentRow } = usePlans()

  return (
    <>
      <PlansMutateDialog key='plan-create' />

      {currentRow && (
        <>
          <PlansMutateDialog key={`plan-update-${currentRow.id}`} />
          <PlansDeleteDialog key='plan-delete' />
        </>
      )}
    </>
  )
}
