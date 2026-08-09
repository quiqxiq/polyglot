import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { Device } from '@/gen/v1/device_pb'

type DevicesDialogType = 'add' | 'edit' | 'delete' | 'test'

type DevicesContextType = {
  open: DevicesDialogType | null
  setOpen: (str: DevicesDialogType | null) => void
  currentRow: Device | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Device | null>>
}

const DevicesContext = React.createContext<DevicesContextType | null>(null)

export function DevicesProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<DevicesDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Device | null>(null)

  return (
    <DevicesContext.Provider value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </DevicesContext.Provider>
  )
}

export const useDevicesContext = () => {
  const context = React.useContext(DevicesContext)
  if (!context) {
    throw new Error('useDevicesContext must be used within <DevicesProvider>')
  }
  return context
}
