import { defineStore } from 'pinia'
import { ref } from 'vue'
import { createSSEConnection } from '../api/client'
import { useSessionStore } from './sessions'
import { useConversationsStore } from './conversations'

export const useRealtimeStore = defineStore('realtime', () => {
  const isConnected = ref(false)
  let eventSource: EventSource | null = null

  function connect() {
    if (eventSource) return

    eventSource = createSSEConnection(
      (event, data) => {
        isConnected.value = true
        const sessionStore = useSessionStore()
        const convStore = useConversationsStore()

        switch (event) {
          case 'session_status':
            sessionStore.handleStatusUpdate(data)
            break
          case 'new_message':
            convStore.handleNewMessage(data)
            break
          case 'conversation_update':
            convStore.handleConversationUpdate(data)
            break
        }
      },
      () => {
        // SSE Connection Handshake Open Success
        isConnected.value = true
      }
    )

    eventSource.onerror = () => {
      isConnected.value = false
    }
  }

  function disconnect() {
    if (eventSource) {
      eventSource.close()
      eventSource = null
      isConnected.value = false
    }
  }

  return {
    isConnected,
    connect,
    disconnect,
  }
})
