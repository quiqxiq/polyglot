import React, { createContext, useContext, useState } from 'react'
import type { PPPActiveSession, PPPProfile, PPPSecret } from '@/gen/v1/ppp_pb'
import useDialogState from '@/hooks/use-dialog-state'

type DialogType =
  | 'secret-create'
  | 'secret-update'
  | 'secret-delete'
  | 'secrets-multi-delete'
  | 'profile-create'
  | 'profile-update'
  | 'profile-delete'
  | 'active-kick'
  | 'active-multi-kick'

interface PPPContextType {
  open: DialogType | null
  setOpen: (type: DialogType | null) => void
  currentSecret: PPPSecret | null
  setCurrentSecret: (secret: PPPSecret | null) => void
  currentProfile: PPPProfile | null
  setCurrentProfile: (profile: PPPProfile | null) => void
  currentActiveSession: PPPActiveSession | null
  setCurrentActiveSession: (session: PPPActiveSession | null) => void
  selectedSecretRows: PPPSecret[]
  setSelectedSecretRows: (rows: PPPSecret[]) => void
  selectedActiveRows: PPPActiveSession[]
  setSelectedActiveRows: (rows: PPPActiveSession[]) => void
}

const PPPContext = createContext<PPPContextType | null>(null)

export function PPPProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<DialogType>(null)
  const [currentSecret, setCurrentSecret] = useState<PPPSecret | null>(null)
  const [currentProfile, setCurrentProfile] = useState<PPPProfile | null>(null)
  const [currentActiveSession, setCurrentActiveSession] = useState<PPPActiveSession | null>(null)
  const [selectedSecretRows, setSelectedSecretRows] = useState<PPPSecret[]>([])
  const [selectedActiveRows, setSelectedActiveRows] = useState<PPPActiveSession[]>([])

  return (
    <PPPContext.Provider
      value={{
        open,
        setOpen,
        currentSecret,
        setCurrentSecret,
        currentProfile,
        setCurrentProfile,
        currentActiveSession,
        setCurrentActiveSession,
        selectedSecretRows,
        setSelectedSecretRows,
        selectedActiveRows,
        setSelectedActiveRows,
      }}
    >
      {children}
    </PPPContext.Provider>
  )
}

export function usePPP() {
  const context = useContext(PPPContext)
  if (!context) {
    throw new Error('usePPP must be used within a PPPProvider')
  }
  return context
}
