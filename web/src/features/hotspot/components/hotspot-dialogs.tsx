import { UserMutateDialog } from './users/user-mutate-dialog'
import { UserResetDialog } from './users/user-reset-dialog'
import { UserDeleteDialog } from './users/user-delete-dialog'
import { ProfileMutateDialog } from './profiles/profile-mutate-dialog'
import { ProfileDeleteDialog } from './profiles/profile-delete-dialog'
import { ActiveKickDialog } from './active/active-kick-dialog'
import { HostDeleteDialog } from './hosts/host-delete-dialog'
import { VoucherGenerateDialog } from './vouchers/voucher-generate-dialog'
import { VoucherPrintDialog } from './vouchers/voucher-print-dialog'
import { ExpireMonitorDialog } from './expire/expire-monitor-dialog'

export function HotspotDialogs() {
  return (
    <>
      <UserMutateDialog />
      <UserResetDialog />
      <UserDeleteDialog />
      <ProfileMutateDialog />
      <ProfileDeleteDialog />
      <ActiveKickDialog />
      <HostDeleteDialog />
      <VoucherGenerateDialog />
      <VoucherPrintDialog />
      <ExpireMonitorDialog />
    </>
  )
}
