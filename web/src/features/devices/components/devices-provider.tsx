import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { Device } from '@/gen/v1/device_pb'

export type DevicesDialogType = 'add' | 'edit' | 'delete' | 'test' | 'terminal'
export type ViewMode = 'card' | 'table'

type DevicesContextType = {
  open: DevicesDialogType | null
  setOpen: (str: DevicesDialogType | null) => void
  currentRow: Device | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Device | null>>
  viewMode: ViewMode
  setViewMode: React.Dispatch<React.SetStateAction<ViewMode>>
}

const DevicesContext = React.createContext<DevicesContextType | null>(null)

export function DevicesProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<DevicesDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Device | null>(null)
  const [viewMode, setViewMode] = useState<ViewMode>('card')

  return (
    <DevicesContext.Provider
      value={{
        open,
        setOpen,
        currentRow,
        setCurrentRow,
        viewMode,
        setViewMode,
      }}
    >
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

