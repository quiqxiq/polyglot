import React, { useState } from 'react'
import { type KnowledgeItem } from '@/gen/v1/knowledge_pb'
import useDialogState from '@/hooks/use-dialog-state'

export type KnowledgeDialogType = 'delete'

type KnowledgeContextType = {
  open: KnowledgeDialogType | null
  setOpen: (open: KnowledgeDialogType | null) => void
  currentRow: KnowledgeItem | null
  setCurrentRow: React.Dispatch<React.SetStateAction<KnowledgeItem | null>>
}

const KnowledgeContext = React.createContext<KnowledgeContextType | null>(null)

export function KnowledgeProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<KnowledgeDialogType>(null)
  const [currentRow, setCurrentRow] = useState<KnowledgeItem | null>(null)

  return (
    <KnowledgeContext.Provider
      value={{
        open,
        setOpen,
        currentRow,
        setCurrentRow,
      }}
    >
      {children}
    </KnowledgeContext.Provider>
  )
}

export const useKnowledgeContext = () => {
  const context = React.useContext(KnowledgeContext)
  if (!context) {
    throw new Error(
      'useKnowledgeContext must be used within <KnowledgeProvider>'
    )
  }
  return context
}
