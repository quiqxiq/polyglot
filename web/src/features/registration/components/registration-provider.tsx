import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import type { Registration } from '@/gen/v1/registration_pb'

export type RegistrationDialogType =
  | 'submit'
  | 'schedule'
  | 'install'
  | 'convert'
  | 'reject'
  | 'cancel'
  | 'detail'

type RegistrationContextType = {
  open: RegistrationDialogType | null
  setOpen: (str: RegistrationDialogType | null) => void
  currentRow: Registration | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Registration | null>>
}

const RegistrationContext = React.createContext<RegistrationContextType | null>(null)

export function RegistrationProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<RegistrationDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Registration | null>(null)

  return (
    <RegistrationContext.Provider value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </RegistrationContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useRegistration = () => {
  const context = React.useContext(RegistrationContext)
  if (!context) {
    throw new Error('useRegistration must be used within <RegistrationProvider>')
  }
  return context
}
