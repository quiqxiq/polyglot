import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface DeviceState {
  selectedDeviceId: string
  hasHydrated: boolean
  setSelectedDeviceId: (id: string) => void
  setHasHydrated: (hydrated: boolean) => void
}

export const useDeviceStore = create<DeviceState>()(
  persist(
    (set) => ({
      selectedDeviceId: '',
      hasHydrated: false,
      setSelectedDeviceId: (selectedDeviceId) => set({ selectedDeviceId }),
      setHasHydrated: (hasHydrated) => set({ hasHydrated }),
    }),
    {
      name: 'polyglot_selected_device',
      onRehydrateStorage: () => (state) => {
        state?.setHasHydrated(true)
      },
    }
  )
)
