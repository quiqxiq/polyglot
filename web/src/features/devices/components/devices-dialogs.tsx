'use client'

import { useDevicesContext } from './devices-provider'
import { DevicesActionDialog } from './devices-action-dialog'
import { DevicesDeleteDialog } from './devices-delete-dialog'
import { DevicesTestDialog } from './devices-test-dialog'
import { DeviceTerminalDialog } from './device-terminal-dialog'
import { DevicePingSettingsDialog } from './device-ping-settings-dialog'
import { DevicePingAnalyticsDialog } from './device-ping-analytics-dialog'

export function DevicesDialogs() {
  const { open, setOpen, currentRow } = useDevicesContext()

  return (
    <>
      <DevicesActionDialog />
      <DevicesDeleteDialog />
      <DevicesTestDialog />
      <DeviceTerminalDialog />
      <DevicePingSettingsDialog
        open={open === 'ping-settings'}
        onOpenChange={(isOpen) => {
          if (!isOpen) setOpen(null)
        }}
        device={currentRow}
      />
      <DevicePingAnalyticsDialog
        open={open === 'ping-analytics'}
        onOpenChange={(isOpen) => {
          if (!isOpen) setOpen(null)
        }}
        device={currentRow}
      />
    </>
  )
}
