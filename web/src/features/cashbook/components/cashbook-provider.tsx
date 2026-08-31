import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { type CashAccount, type CashCategory, type CashTransaction } from '@/gen/v1/cashbook_pb'

export type CashbookDialogType =
  | 'create-transaction'
  | 'create-account'
  | 'edit-account'
  | 'create-category'
  | 'edit-category'
  | 'detail-transaction'

export interface CashbookFilterState {
  accountId: string
  categoryId: string
  direction: string
  fromUnix?: number
  toUnix?: number
}

type CashbookContextType = {
  open: CashbookDialogType | null
  setOpen: (str: CashbookDialogType | null) => void
  currentAccount: CashAccount | null
  setCurrentAccount: React.Dispatch<React.SetStateAction<CashAccount | null>>
  currentCategory: CashCategory | null
  setCurrentCategory: React.Dispatch<React.SetStateAction<CashCategory | null>>
  currentTransaction: CashTransaction | null
  setCurrentTransaction: React.Dispatch<React.SetStateAction<CashTransaction | null>>
  filters: CashbookFilterState
  setFilters: React.Dispatch<React.SetStateAction<CashbookFilterState>>
  resetFilters: () => void
}

const initialFilters: CashbookFilterState = {
  accountId: '',
  categoryId: '',
  direction: '',
}

const CashbookContext = React.createContext<CashbookContextType | null>(null)

export function CashbookProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<CashbookDialogType>(null)
  const [currentAccount, setCurrentAccount] = useState<CashAccount | null>(null)
  const [currentCategory, setCurrentCategory] = useState<CashCategory | null>(null)
  const [currentTransaction, setCurrentTransaction] = useState<CashTransaction | null>(null)
  const [filters, setFilters] = useState<CashbookFilterState>(initialFilters)

  const resetFilters = () => setFilters(initialFilters)

  return (
    <CashbookContext
      value={{
        open,
        setOpen,
        currentAccount,
        setCurrentAccount,
        currentCategory,
        setCurrentCategory,
        currentTransaction,
        setCurrentTransaction,
        filters,
        setFilters,
        resetFilters,
      }}
    >
      {children}
    </CashbookContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useCashbook = () => {
  const context = React.useContext(CashbookContext)
  if (!context) {
    throw new Error('useCashbook must be used within <CashbookProvider>')
  }
  return context
}
