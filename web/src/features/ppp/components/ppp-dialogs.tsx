import { SecretMutateDialog } from './secrets/secret-mutate-dialog'
import { SecretDeleteDialog } from './secrets/secret-delete-dialog'
import { ProfileMutateDialog } from './profiles/profile-mutate-dialog'
import { ProfileDeleteDialog } from './profiles/profile-delete-dialog'
import { ActiveKickDialog } from './active/active-kick-dialog'
import { ActivePingDialog } from './active/active-ping-dialog'

export function PPPDialogs() {
  return (
    <>
      <SecretMutateDialog />
      <SecretDeleteDialog />
      <ProfileMutateDialog />
      <ProfileDeleteDialog />
      <ActiveKickDialog />
      <ActivePingDialog />
    </>
  )
}
