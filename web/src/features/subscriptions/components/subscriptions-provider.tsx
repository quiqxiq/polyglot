import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog-state'
import { type Subscription } from '@/gen/v1/billing_pb'

type SubscriptionsDialogType =
  | 'change-plan'
  | 'suspend'
  | 'resume'
  | 'terminate'
  | 'activate'

type SubscriptionsContextType = {
  open: SubscriptionsDialogType | null
  setOpen: (str: SubscriptionsDialogType | null) => void
  currentRow: Subscription | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Subscription | null>>
}

const SubscriptionsContext = React.createContext<SubscriptionsContextType | null>(
  null
)

export function SubscriptionsProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<SubscriptionsDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Subscription | null>(null)

  return (
    <SubscriptionsContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </SubscriptionsContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useSubscriptions = () => {
  const subscriptionsContext = React.useContext(SubscriptionsContext)

  if (!subscriptionsContext) {
    throw new Error('useSubscriptions has to be used within <SubscriptionsProvider>')
  }

  return subscriptionsContext
}
