import { UserMutateDialog } from './users/user-mutate-dialog'
import { UserResetDialog } from './users/user-reset-dialog'
import { UserDeleteDialog } from './users/user-delete-dialog'
import { UserBulkCleanerDialog } from './users/user-bulk-cleaner-dialog'
import { ProfileMutateDialog } from './profiles/profile-mutate-dialog'
import { ProfileDeleteDialog } from './profiles/profile-delete-dialog'
import { ActiveKickDialog } from './active/active-kick-dialog'
import { HostDeleteDialog } from './hosts/host-delete-dialog'
import { VoucherGenerateDialog } from './vouchers/voucher-generate-dialog'
import { VoucherPrintDialog } from './vouchers/voucher-print-dialog'
import { QuickVoucherCheckerDialog } from './vouchers/quick-voucher-checker-dialog'
import { BindingMutateDialog } from './bindings/binding-mutate-dialog'
import { BindingDeleteDialog } from './bindings/binding-delete-dialog'
import { CookieDeleteDialog } from './cookies/cookie-delete-dialog'
import { CookieClearAllDialog } from './cookies/cookie-clear-all-dialog'
import { ExpireMonitorDialog } from './expire/expire-monitor-dialog'
import { useHotspot } from '../context/hotspot-context'

export function HotspotDialogs() {
  const { open, setOpen } = useHotspot()

  return (
    <>
      <UserMutateDialog />
      <UserResetDialog />
      <UserDeleteDialog />
      <UserBulkCleanerDialog
        open={open === 'user-bulk-cleaner'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'user-bulk-cleaner' : null)}
      />
      <ProfileMutateDialog />
      <ProfileDeleteDialog />
      <ActiveKickDialog />
      <HostDeleteDialog />
      <VoucherGenerateDialog />
      <VoucherPrintDialog />
      <QuickVoucherCheckerDialog
        open={open === 'quick-voucher-checker'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'quick-voucher-checker' : null)}
      />
      <BindingMutateDialog
        open={open === 'binding-create'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'binding-create' : null)}
        isEdit={false}
      />
      <BindingMutateDialog
        open={open === 'binding-update'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'binding-update' : null)}
        isEdit={true}
      />
      <BindingDeleteDialog
        open={open === 'binding-delete'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'binding-delete' : null)}
      />
      <CookieDeleteDialog
        open={open === 'cookie-delete'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'cookie-delete' : null)}
      />
      <CookieClearAllDialog
        open={open === 'cookie-clear-all'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'cookie-clear-all' : null)}
      />
      <ExpireMonitorDialog />
    </>
  )
}
