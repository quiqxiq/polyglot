import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import type {
  HotspotUser,
  HotspotProfile,
  HotspotActiveSession,
  HotspotHost,
} from '@/gen/v1/hotspot_pb'

export type HotspotDialogType =
  | 'user-create'
  | 'user-update'
  | 'user-delete'
  | 'user-reset'
  | 'profile-create'
  | 'profile-update'
  | 'profile-delete'
  | 'session-kick'
  | 'host-delete'
  | 'voucher-generate'
  | 'voucher-print'
  | 'expire-monitor'

type HotspotContextType = {
  open: HotspotDialogType | null
  setOpen: (dialog: HotspotDialogType | null) => void
  currentUser: HotspotUser | null
  setCurrentUser: React.Dispatch<React.SetStateAction<HotspotUser | null>>
  currentProfile: HotspotProfile | null
  setCurrentProfile: React.Dispatch<React.SetStateAction<HotspotProfile | null>>
  currentSession: HotspotActiveSession | null
  setCurrentSession: React.Dispatch<React.SetStateAction<HotspotActiveSession | null>>
  currentHost: HotspotHost | null
  setCurrentHost: React.Dispatch<React.SetStateAction<HotspotHost | null>>
  printBatchComment: string
  setPrintBatchComment: React.Dispatch<React.SetStateAction<string>>
  printSingleUserId: string
  setPrintSingleUserId: React.Dispatch<React.SetStateAction<string>>
}

const HotspotContext = React.createContext<HotspotContextType | null>(null)

export function HotspotProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<HotspotDialogType>(null)
  const [currentUser, setCurrentUser] = useState<HotspotUser | null>(null)
  const [currentProfile, setCurrentProfile] = useState<HotspotProfile | null>(null)
  const [currentSession, setCurrentSession] = useState<HotspotActiveSession | null>(null)
  const [currentHost, setCurrentHost] = useState<HotspotHost | null>(null)
  const [printBatchComment, setPrintBatchComment] = useState<string>('')
  const [printSingleUserId, setPrintSingleUserId] = useState<string>('')

  return (
    <HotspotContext.Provider
      value={{
        open,
        setOpen,
        currentUser,
        setCurrentUser,
        currentProfile,
        setCurrentProfile,
        currentSession,
        setCurrentSession,
        currentHost,
        setCurrentHost,
        printBatchComment,
        setPrintBatchComment,
        printSingleUserId,
        setPrintSingleUserId,
      }}
    >
      {children}
    </HotspotContext.Provider>
  )
}

export const useHotspot = () => {
  const context = React.useContext(HotspotContext)
  if (!context) {
    throw new Error('useHotspot must be used within <HotspotProvider>')
  }
  return context
}
