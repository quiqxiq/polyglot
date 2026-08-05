import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Conversation, Message } from '../types'
import {
  listConversationsApi,
  getConversationApi,
  takeOverApi,
  resetBotApi,
  sendAgentMessageApi,
  closeConversationApi,
} from '../api/client'

export const useConversationsStore = defineStore('conversations', () => {
  const conversations = ref<Conversation[]>([])
  const activeConversation = ref<Conversation | null>(null)
  const loading = ref(false)
  const statusFilter = ref<string>('') // '', 'bot', 'escalation', 'done'

  const filteredConversations = computed(() => {
    if (!statusFilter.value) return conversations.value
    return conversations.value.filter((c) => c.status === statusFilter.value)
  })

  async function fetchConversations(status?: string) {
    loading.value = true
    try {
      const res = await listConversationsApi(status || statusFilter.value)
      conversations.value = res.conversations || []
    } finally {
      loading.value = false
    }
  }

  async function selectConversation(id: number) {
    loading.value = true
    try {
      const res = await getConversationApi(id)
      activeConversation.value = res.conversation
    } finally {
      loading.value = false
    }
  }

  async function takeOver(id: number) {
    const res = await takeOverApi(id)
    if (activeConversation.value && activeConversation.value.id === id) {
      activeConversation.value.status = 'escalation'
    }
    await fetchConversations()
    return res
  }

  async function resetBot(id: number) {
    const res = await resetBotApi(id)
    if (activeConversation.value && activeConversation.value.id === id) {
      activeConversation.value.status = 'bot'
    }
    await fetchConversations()
    return res
  }

  async function sendMessage(id: number, content: string) {
    const res = await sendAgentMessageApi(id, content)
    if (activeConversation.value && activeConversation.value.id === id) {
      if (!activeConversation.value.messages) {
        activeConversation.value.messages = []
      }
      activeConversation.value.messages.push(res.data)
    }
    return res
  }

  async function closeConversation(id: number) {
    const res = await closeConversationApi(id)
    if (activeConversation.value && activeConversation.value.id === id) {
      activeConversation.value.status = 'done'
    }
    await fetchConversations()
    return res
  }

  function handleNewMessage(msg: Message) {
    // If active conversation matches, append message
    if (activeConversation.value && activeConversation.value.id === msg.conversation_id) {
      if (!activeConversation.value.messages) {
        activeConversation.value.messages = []
      }
      // Avoid duplicate
      if (!activeConversation.value.messages.some((m) => m.id === msg.id)) {
        activeConversation.value.messages.push(msg)
      }
    }
    // Refresh conversation list order
    fetchConversations()
  }

  function handleConversationUpdate(conv: Conversation) {
    const idx = conversations.value.findIndex((c) => c.id === conv.id)
    if (idx !== -1) {
      conversations.value[idx] = conv
    }
    if (activeConversation.value && activeConversation.value.id === conv.id) {
      activeConversation.value.status = conv.status
      activeConversation.value.assigned_agent_id = conv.assigned_agent_id
    }
  }

  return {
    conversations,
    activeConversation,
    loading,
    statusFilter,
    filteredConversations,
    fetchConversations,
    selectConversation,
    takeOver,
    resetBot,
    sendMessage,
    closeConversation,
    handleNewMessage,
    handleConversationUpdate,
  }
})
