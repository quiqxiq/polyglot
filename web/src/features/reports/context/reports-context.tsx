import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import type { HotspotReport } from '@/gen/v1/hotspot_pb'

export type ReportsDialogType = 'report-delete'

type ReportsContextType = {
  open: ReportsDialogType | null
  setOpen: (dialog: ReportsDialogType | null) => void
  currentReport: HotspotReport | null
  setCurrentReport: React.Dispatch<React.SetStateAction<HotspotReport | null>>
}

const ReportsContext = React.createContext<ReportsContextType | null>(null)

export function ReportsProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<ReportsDialogType>(null)
  const [currentReport, setCurrentReport] = useState<HotspotReport | null>(null)

  return (
    <ReportsContext.Provider
      value={{
        open,
        setOpen,
        currentReport,
        setCurrentReport,
      }}
    >
      {children}
    </ReportsContext.Provider>
  )
}

export const useReports = () => {
  const context = React.useContext(ReportsContext)
  if (!context) {
    throw new Error('useReports must be used within <ReportsProvider>')
  }
  return context
}
