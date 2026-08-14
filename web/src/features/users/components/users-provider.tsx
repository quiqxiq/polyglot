import React, { useState } from 'react'
import { type User } from '@/gen/v1/users_pb'

export type UsersDialogType = 'create' | 'edit' | 'reset' | 'toggle' | 'delete'

type UsersContextType = {
  open: UsersDialogType | null
  setOpen: (type: UsersDialogType | null) => void
  currentRow: User | null
  setCurrentRow: (user: User | null) => void
}

const UsersContext = React.createContext<UsersContextType | null>(null)

export function UsersProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useState<UsersDialogType | null>(null)
  const [currentRow, setCurrentRow] = useState<User | null>(null)

  return (
    <UsersContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </UsersContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useUsers() {
  const usersContext = React.useContext(UsersContext)

  if (!usersContext) {
    throw new Error('useUsers has to be used within <UsersProvider>')
  }

  return usersContext
}
