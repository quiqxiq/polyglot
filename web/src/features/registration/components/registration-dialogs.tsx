import { RegistrationCreateDialog } from './registration-create-dialog'
import { RegistrationScheduleDialog } from './registration-schedule-dialog'
import { RegistrationInstallDialog } from './registration-install-dialog'
import { RegistrationConvertDialog } from './registration-convert-dialog'
import { RegistrationRejectDialog } from './registration-reject-dialog'
import { RegistrationCancelDialog } from './registration-cancel-dialog'
import { RegistrationDetailSheet } from './registration-detail-sheet'
import { useRegistration } from './registration-provider'

export function RegistrationDialogs() {
  const { open, setOpen, currentRow } = useRegistration()

  return (
    <>
      <RegistrationCreateDialog
        key='registration-submit'
        open={open === 'submit'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'submit' : null)}
      />
      {currentRow && (
        <RegistrationScheduleDialog
          key={`registration-schedule-${currentRow.id}`}
          open={open === 'schedule'}
          onOpenChange={(isOpen) => setOpen(isOpen ? 'schedule' : null)}
          registration={currentRow}
        />
      )}
      {currentRow && (
        <RegistrationInstallDialog
          key={`registration-install-${currentRow.id}`}
          open={open === 'install'}
          onOpenChange={(isOpen) => setOpen(isOpen ? 'install' : null)}
          registration={currentRow}
        />
      )}
      {currentRow && (
        <RegistrationConvertDialog
          key={`registration-convert-${currentRow.id}`}
          open={open === 'convert'}
          onOpenChange={(isOpen) => setOpen(isOpen ? 'convert' : null)}
          registration={currentRow}
        />
      )}
      {currentRow && (
        <RegistrationRejectDialog
          key={`registration-reject-${currentRow.id}`}
          open={open === 'reject'}
          onOpenChange={(isOpen) => setOpen(isOpen ? 'reject' : null)}
          registration={currentRow}
        />
      )}
      {currentRow && (
        <RegistrationCancelDialog
          key={`registration-cancel-${currentRow.id}`}
          open={open === 'cancel'}
          onOpenChange={(isOpen) => setOpen(isOpen ? 'cancel' : null)}
          registration={currentRow}
        />
      )}
      {currentRow && (
        <RegistrationDetailSheet
          key={`registration-detail-${currentRow.id}`}
          open={open === 'detail'}
          onOpenChange={(isOpen) => setOpen(isOpen ? 'detail' : null)}
          registration={currentRow}
        />
      )}
    </>
  )
}
