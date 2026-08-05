<template>
  <div class="conversations-view">
    <!-- Filter Tabs -->
    <div class="filter-tabs glass-panel">
      <button
        v-for="tab in filterTabs"
        :key="tab.value"
        :class="['filter-tab', { active: convStore.statusFilter === tab.value }]"
        @click="handleFilterChange(tab.value)"
      >
        <component :is="tab.icon" class="w-4 h-4" />
        <span>{{ tab.label }}</span>
        <span class="tab-count">{{ getCountByStatus(tab.value) }}</span>
      </button>
    </div>

    <!-- Main Dual-Pane Chat Layout -->
    <div :class="['chat-dual-pane', `mobile-view-${mobileActiveView}`]">
      <!-- Left Sidebar: Conversations List -->
      <div class="chat-list-panel glass-panel">
        <div class="list-header">
          <h4>Daftar Percakapan</h4>
          <button class="btn btn-secondary btn-sm" @click="convStore.fetchConversations()">
            <RefreshCw class="w-3.5 h-3.5" />
          </button>
        </div>

        <div v-if="convStore.loading && convStore.conversations.length === 0" class="panel-loading">
          <Loader2 class="w-6 h-6 spin text-indigo-400" />
          <span>Memuat percakapan...</span>
        </div>

        <div v-else-if="convStore.filteredConversations.length === 0" class="panel-empty">
          <MessageSquare class="w-10 h-10 text-slate-600 mb-2" />
          <p>Belum ada percakapan dengan status ini.</p>
        </div>

        <div v-else class="conv-items-scroll">
          <div
            v-for="conv in convStore.filteredConversations"
            :key="conv.id"
            :class="['conv-item', { active: convStore.activeConversation?.id === conv.id }]"
            @click="selectConvMobile(conv.id)"
          >
            <div class="conv-avatar">
              {{ conv.customer_wa_number.slice(-4) }}
            </div>
            <div class="conv-info">
              <div class="conv-top">
                <strong class="customer-num">+{{ conv.customer_wa_number }}</strong>
                <span class="time-text">{{ formatDateShort(conv.updated_at) }}</span>
              </div>
              <div class="conv-bottom">
                <span :class="['badge', getStatusBadgeClass(conv.status)]">
                  {{ getStatusText(conv.status) }}
                </span>
                <span v-if="conv.assigned_agent_id" class="agent-tag">
                  <UserCheck class="w-3 h-3" /> Agent
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Main: Chat Workspace -->
      <div class="chat-main-panel glass-panel">
        <template v-if="convStore.activeConversation">
          <!-- Chat Header Bar -->
          <div class="chat-header-bar">
            <div class="chat-target-info">
              <button class="btn btn-secondary btn-sm mobile-back-btn" @click="mobileActiveView = 'list'">
                <ArrowLeft class="w-4 h-4" />
                <span>Daftar</span>
              </button>
              <div class="target-avatar">
                {{ convStore.activeConversation.customer_wa_number.slice(-4) }}
              </div>
              <div class="target-details">
                <h4>+{{ convStore.activeConversation.customer_wa_number }}</h4>
                <div class="target-status">
                  <span :class="['badge', getStatusBadgeClass(convStore.activeConversation.status)]">
                    {{ getStatusText(convStore.activeConversation.status) }}
                  </span>
                  <span class="text-xs text-muted session-id-text">ID: #{{ convStore.activeConversation.session_id }}</span>
                </div>
              </div>
            </div>

            <div class="chat-header-actions">
              <!-- Take Over Button -->
              <button
                v-if="convStore.activeConversation.status === 'bot'"
                class="btn btn-warning btn-sm"
                title="Hentikan bot dan ambil alih obrolan secara manual"
                @click="handleTakeOver(convStore.activeConversation.id)"
              >
                <UserCheck class="w-4 h-4" />
                <span class="action-btn-text">Ambil Alih</span>
              </button>

              <!-- Reset Bot Button -->
              <button
                v-if="convStore.activeConversation.status === 'escalation'"
                class="btn btn-primary btn-sm"
                title="Aktifkan AI Bot kembali untuk membalas otomatis"
                @click="handleResetBot(convStore.activeConversation.id)"
              >
                <Bot class="w-4 h-4" />
                <span class="action-btn-text">Aktifkan Bot</span>
              </button>

              <!-- Close Button -->
              <button
                v-if="convStore.activeConversation.status !== 'done'"
                class="btn btn-secondary btn-sm"
                @click="handleClose(convStore.activeConversation.id)"
              >
                <CheckCircle2 class="w-4 h-4 text-emerald-400" />
                <span class="action-btn-text">Selesai</span>
              </button>
            </div>
          </div>

          <!-- Messages Stream Container -->
          <div class="messages-container" ref="messagesContainerRef">
            <div v-if="!convStore.activeConversation.messages || convStore.activeConversation.messages.length === 0" class="no-msgs">
              <span>Belum ada riwayat pesan di percakapan ini.</span>
            </div>

            <div
              v-for="msg in convStore.activeConversation.messages"
              :key="msg.id"
              :class="['msg-bubble-row', `sender-${msg.sender_type}`]"
            >
              <div class="msg-bubble">
                <div class="msg-sender-label">
                  <span>{{ getSenderLabel(msg.sender_type) }}</span>
                  <span class="msg-time">{{ formatTime(msg.created_at) }}</span>
                </div>
                <div class="msg-text">{{ msg.content }}</div>

                <!-- Token Usage Metrics Pill -->
                <div v-if="msg.sender_type === 'bot' && (msg.token_in > 0 || msg.token_out > 0)" class="token-pill">
                  <Zap class="w-3 h-3 text-amber-400" />
                  <span>Tokens: In {{ msg.token_in }} | Out {{ msg.token_out }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Agent Message Input Composer -->
          <div class="chat-composer">
            <div v-if="convStore.activeConversation.status === 'bot'" class="bot-active-warning">
              <AlertTriangle class="w-4 h-4 text-amber-400 flex-shrink-0" />
              <span>Bot sedang aktif membalas otomatis. Klik <strong>"Ambil Alih"</strong> untuk membalas manual.</span>
            </div>

            <div class="composer-box">
              <textarea
                v-model="replyText"
                rows="2"
                class="form-textarea composer-input"
                placeholder="Ketik balasan manual ke pelanggan..."
                @keydown.enter.exact.prevent="handleSendMessage"
              ></textarea>
              <button
                class="btn btn-primary send-btn"
                :disabled="!replyText.trim() || sending"
                @click="handleSendMessage"
              >
                <Loader2 v-if="sending" class="w-4 h-4 spin" />
                <Send v-else class="w-4 h-4" />
                <span>Kirim</span>
              </button>
            </div>
          </div>
        </template>

        <template v-else>
          <div class="no-selection-placeholder">
            <MessageSquare class="w-16 h-16 text-slate-700 mb-3" />
            <h4>Pilih Percakapan dari Daftar</h4>
            <p>Pilih salah satu nomor pelanggan di sebelah kiri untuk melihat pesan dan mengambil alih obrolan.</p>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue'
import {
  MessageSquare,
  Bot,
  UserCheck,
  CheckCircle2,
  RefreshCw,
  Loader2,
  Send,
  Zap,
  AlertTriangle,
  ArrowLeft,
} from 'lucide-vue-next'

import { useConversationsStore } from '../stores/conversations'

const convStore = useConversationsStore()

const replyText = ref('')
const sending = ref(false)
const messagesContainerRef = ref<HTMLDivElement | null>(null)
const mobileActiveView = ref<'list' | 'chat'>('list')

const filterTabs = [
  { value: '', label: 'Semua Percakapan', icon: MessageSquare },
  { value: 'bot', label: 'Bot Aktif', icon: Bot },
  { value: 'escalation', label: 'Eskalasi Agent', icon: UserCheck },
  { value: 'done', label: 'Selesai', icon: CheckCircle2 },
]

onMounted(() => {
  convStore.fetchConversations()
})

function selectConvMobile(id: number) {
  convStore.selectConversation(id)
  mobileActiveView.value = 'chat'
}

watch(
  () => convStore.activeConversation?.messages?.length,
  () => {
    nextTick(scrollToBottom)
  }
)

function handleFilterChange(val: string) {
  convStore.statusFilter = val
  convStore.fetchConversations()
}

function getCountByStatus(status: string) {
  if (!status) return convStore.conversations.length
  return convStore.conversations.filter((c) => c.status === status).length
}

async function handleTakeOver(id: number) {
  await convStore.takeOver(id)
}

async function handleResetBot(id: number) {
  await convStore.resetBot(id)
}

async function handleClose(id: number) {
  if (confirm('Apakah Anda yakin ingin menyelesaikan percakapan ini?')) {
    await convStore.closeConversation(id)
  }
}

async function handleSendMessage() {
  if (!replyText.value.trim() || !convStore.activeConversation) return
  sending.value = true
  try {
    await convStore.sendMessage(convStore.activeConversation.id, replyText.value.trim())
    replyText.value = ''
    scrollToBottom()
  } catch (err: any) {
    alert(err.message || 'Gagal mengirim pesan')
  } finally {
    sending.value = false
  }
}

function scrollToBottom() {
  if (messagesContainerRef.value) {
    messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight
  }
}

function getSenderLabel(type: string) {
  switch (type) {
    case 'customer': return 'Pelanggan'
    case 'bot': return 'GNET Bot AI'
    case 'agent': return 'Agent Admin'
    default: return type
  }
}

function getStatusBadgeClass(status: string) {
  switch (status) {
    case 'bot': return 'badge-primary'
    case 'escalation': return 'badge-warning'
    case 'done': return 'badge-online'
    default: return 'badge-offline'
  }
}

function getStatusText(status: string) {
  switch (status) {
    case 'bot': return 'Bot Aktif'
    case 'escalation': return 'Eskalasi Agent'
    case 'done': return 'Selesai'
    default: return status
  }
}

function formatDateShort(d: string) {
  if (!d) return ''
  return new Date(d).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function formatTime(d: string) {
  if (!d) return ''
  return new Date(d).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>

<style scoped>
.conversations-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: calc(100vh - 120px);
}

.filter-tabs {
  display: flex;
  gap: 8px;
  padding: 8px;

  button {
    border-radius: var(--radius-md);
  }
}

.filter-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  background: transparent;
  border: 1px solid transparent;
  cursor: pointer;

  &:hover {
    color: var(--text-main);
    background: rgba(255, 255, 255, 0.05);
  }

  &.active {
    background: rgba(99, 102, 241, 0.2);
    color: var(--primary-light);
    border-color: rgba(99, 102, 241, 0.4);
  }
}

.tab-count {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 9999px;
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-main);
}

.chat-dual-pane {
  display: flex;
  gap: 16px;
  flex: 1;
  min-height: 0;
}

.chat-list-panel {
  width: 320px;
  display: flex;
  flex-direction: column;
  padding: 0;
  overflow: hidden;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);

  h4 {
    font-size: 14px;
    font-weight: 700;
  }
}

.conv-items-scroll {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.conv-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: background 0.2s ease;

  &:hover {
    background: rgba(255, 255, 255, 0.04);
  }

  &.active {
    background: rgba(99, 102, 241, 0.15);
    border-left: 3px solid var(--primary);
  }
}

.conv-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary) 0%, var(--accent-cyan) 100%);
  color: #ffffff;
  font-size: 13px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.conv-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.conv-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.customer-num {
  font-size: 13px;
}

.time-text {
  font-size: 11px;
  color: var(--text-muted);
}

.conv-bottom {
  display: flex;
  align-items: center;
  gap: 8px;
}

.agent-tag {
  font-size: 10px;
  color: var(--color-warning);
  display: flex;
  align-items: center;
  gap: 2px;
}

.chat-main-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 0;
  overflow: hidden;
}

.chat-header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  background: rgba(15, 23, 42, 0.5);
}

.chat-target-info {
  display: flex;
  align-items: center;
  gap: 12px;

  h4 {
    font-size: 16px;
    font-weight: 700;
  }
}

.target-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary) 0%, var(--accent-cyan) 100%);
  color: #ffffff;
  font-size: 14px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.target-status {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 2px;
}

.chat-header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.msg-bubble-row {
  display: flex;

  &.sender-customer {
    justify-content: flex-start;

    .msg-bubble {
      background: rgba(30, 41, 59, 0.9);
      border: 1px solid var(--border-color);
      border-radius: 4px 16px 16px 16px;
    }
  }

  &.sender-bot {
    justify-content: flex-end;

    .msg-bubble {
      background: linear-gradient(135deg, rgba(99, 102, 241, 0.25) 0%, rgba(79, 70, 229, 0.25) 100%);
      border: 1px solid rgba(99, 102, 241, 0.4);
      border-radius: 16px 4px 16px 16px;
    }
  }

  &.sender-agent {
    justify-content: flex-end;

    .msg-bubble {
      background: linear-gradient(135deg, rgba(16, 185, 129, 0.25) 0%, rgba(5, 150, 105, 0.25) 100%);
      border: 1px solid rgba(16, 185, 129, 0.4);
      border-radius: 16px 4px 16px 16px;
    }
  }
}

.msg-bubble {
  max-width: 75%;
  padding: 12px 16px;
  box-shadow: var(--shadow-sm);
}

.msg-sender-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 11px;
  font-weight: 700;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.msg-time {
  font-size: 10px;
  color: var(--text-muted);
  font-weight: 400;
}

.msg-text {
  font-size: 14px;
  line-height: 1.5;
  white-space: pre-line;
}

.token-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  font-family: 'JetBrains Mono', monospace;
  color: var(--text-muted);
  background: rgba(0, 0, 0, 0.25);
  padding: 2px 6px;
  border-radius: 4px;
  margin-top: 6px;
}

.chat-composer {
  padding: 16px;
  border-top: 1px solid var(--border-color);
  background: rgba(15, 23, 42, 0.6);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.bot-active-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--color-warning);
  background: var(--color-warning-bg);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
}

.composer-box {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.composer-input {
  flex: 1;
  resize: none;
}

.send-btn {
  height: 44px;
  padding: 0 20px;
}

.no-selection-placeholder {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  text-align: center;
  padding: 40px;

  h4 {
    font-size: 18px;
    font-weight: 700;
  }

  p {
    font-size: 13px;
    color: var(--text-muted);
    max-width: 360px;
    margin-top: 4px;
  }
}

.panel-loading, .panel-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: var(--text-secondary);
  font-size: 13px;
}

.btn-warning {
  background: var(--color-warning-bg);
  color: var(--color-warning);
  border-color: rgba(245, 158, 11, 0.4);

  &:hover {
    background: var(--color-warning);
    color: #000000;
  }
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.mobile-back-btn {
  display: none;
}

@media (max-width: 768px) {
  .conversations-view {
    height: auto;
    min-height: calc(100vh - 110px);
  }
  .filter-tabs {
    overflow-x: auto;
    white-space: nowrap;
    padding-bottom: 6px;
  }
  .filter-tab {
    flex-shrink: 0;
  }
  .mobile-back-btn {
    display: inline-flex;
    margin-right: 6px;
  }
  .chat-dual-pane {
    position: relative;
    height: calc(100vh - 170px);
    min-height: 450px;
  }
  .chat-dual-pane.mobile-view-list .chat-list-panel {
    display: flex;
    width: 100%;
    height: 100%;
  }
  .chat-dual-pane.mobile-view-list .chat-main-panel {
    display: none;
  }
  .chat-dual-pane.mobile-view-chat .chat-list-panel {
    display: none;
  }
  .chat-dual-pane.mobile-view-chat .chat-main-panel {
    display: flex;
    width: 100%;
    height: 100%;
  }
  .action-btn-text {
    display: none;
  }
  .session-id-text {
    display: none;
  }
}
</style>
