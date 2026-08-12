'use client'

import { DevicesActionDialog } from './devices-action-dialog'
import { DevicesDeleteDialog } from './devices-delete-dialog'
import { DevicesTestDialog } from './devices-test-dialog'
import { DeviceTerminalDialog } from './device-terminal-dialog'

export function DevicesDialogs() {
  return (
    <>
      <DevicesActionDialog />
      <DevicesDeleteDialog />
      <DevicesTestDialog />
      <DeviceTerminalDialog />
    </>
  )
}
