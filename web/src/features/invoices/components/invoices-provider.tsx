import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { type Invoice } from '@/gen/v1/billing_pb'

export type InvoicesDialogType =
  | 'cashier'
  | 'generate'
  | 'detail'
  | 'print'

export interface InvoicesFilterState {
  customerId: string
  status: string
  period: string
}

type InvoicesContextType = {
  open: InvoicesDialogType | null
  setOpen: (str: InvoicesDialogType | null) => void
  currentInvoice: Invoice | null
  setCurrentInvoice: React.Dispatch<React.SetStateAction<Invoice | null>>
  filters: InvoicesFilterState
  setFilters: React.Dispatch<React.SetStateAction<InvoicesFilterState>>
  resetFilters: () => void
}

const initialFilters: InvoicesFilterState = {
  customerId: '',
  status: '',
  period: '',
}

const InvoicesContext = React.createContext<InvoicesContextType | null>(null)

export function InvoicesProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<InvoicesDialogType>(null)
  const [currentInvoice, setCurrentInvoice] = useState<Invoice | null>(null)
  const [filters, setFilters] = useState<InvoicesFilterState>(initialFilters)

  const resetFilters = () => setFilters(initialFilters)

  return (
    <InvoicesContext
      value={{
        open,
        setOpen,
        currentInvoice,
        setCurrentInvoice,
        filters,
        setFilters,
        resetFilters,
      }}
    >
      {children}
    </InvoicesContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useInvoices = () => {
  const context = React.useContext(InvoicesContext)
  if (!context) {
    throw new Error('useInvoices must be used within <InvoicesProvider>')
  }
  return context
}
