import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { type Customer } from '@/gen/v1/customer_pb'

type CustomersDialogType = 'create' | 'update' | 'delete'

type CustomersContextType = {
  open: CustomersDialogType | null
  setOpen: (str: CustomersDialogType | null) => void
  currentRow: Customer | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Customer | null>>
}

const CustomersContext = React.createContext<CustomersContextType | null>(null)

export function CustomersProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<CustomersDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Customer | null>(null)

  return (
    <CustomersContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </CustomersContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useCustomers = () => {
  const customersContext = React.useContext(CustomersContext)

  if (!customersContext) {
    throw new Error('useCustomers has to be used within <CustomersProvider>')
  }

  return customersContext
}
