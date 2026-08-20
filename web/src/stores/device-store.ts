import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface DeviceState {
  selectedDeviceId: string
  setSelectedDeviceId: (id: string) => void
}

export const useDeviceStore = create<DeviceState>()(
  persist(
    (set) => ({
      selectedDeviceId: '',
      setSelectedDeviceId: (selectedDeviceId) => set({ selectedDeviceId }),
    }),
    {
      name: 'polyglot_selected_device',
    }
  )
)
