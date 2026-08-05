<template>
  <div class="dashboard-overview">
    <!-- Active LLM Banner -->
    <div class="active-llm-banner glass-panel">
      <div class="banner-content">
        <div class="banner-badge">
          <Cpu class="w-6 h-6 text-cyan-400" />
        </div>
        <div class="banner-info">
          <span class="banner-tag">LLM Provider Aktif</span>
          <h3 v-if="activeLLM">
            {{ activeLLM.provider.toUpperCase() }} — <span class="text-indigo-400">{{ activeLLM.model }}</span>
          </h3>
          <h3 v-else class="text-warning">Belum ada Provider LLM yang diaktifkan</h3>
          <p>Sistem chatbot merespons menggunakan provider ini. API Key tersimpan aman (AES-256-GCM).</p>
        </div>
      </div>
      <router-link to="/llm-config" class="btn btn-secondary btn-sm">
        <Settings class="w-4 h-4" />
        <span>Kelola LLM</span>
      </router-link>
    </div>

    <!-- Stat Grid -->
    <div class="stat-grid">
      <StatCard
        title="Koneksi WhatsApp"
        :value="totalSessions"
        :subtext="`${onlineSessions} nomor aktif/online`"
        :icon="Smartphone"
        iconBgClass="icon-bg-primary"
      />
      <StatCard
        title="Chatbot Aktif"
        :value="activeBots"
        :subtext="`${activeBots} dari ${totalSessions} nomor otomatis membalas`"
        :icon="Bot"
        iconBgClass="icon-bg-emerald"
      />
      <StatCard
        title="Basis Pengetahuan (FAQ)"
        :value="totalKnowledge"
        subtext="Grounding acuan jawaban bot"
        :icon="BookOpen"
        iconBgClass="icon-bg-cyan"
      />
      <StatCard
        title="Percakapan Berjalan"
        :value="totalConversations"
        :subtext="`${escalatedConvs} obrolan diekskalasi ke agent`"
        :icon="MessageSquare"
        iconBgClass="icon-bg-amber"
      />
    </div>

    <!-- Main Content Layout (Quick Actions + Activity Feed) -->
    <div class="overview-content">
      <!-- Quick Action Shortcuts -->
      <div class="section-card glass-panel">
        <div class="section-header">
          <Zap class="w-5 h-5 text-amber-400" />
          <h3>Aksi Cepat</h3>
        </div>
        <div class="shortcuts-grid">
          <router-link to="/sessions" class="shortcut-card">
            <Smartphone class="w-6 h-6 text-indigo-400" />
            <span>Tambah Nomor WA</span>
            <ArrowRight class="w-4 h-4 arrow" />
          </router-link>
          <router-link to="/knowledge" class="shortcut-card">
            <PlusCircle class="w-6 h-6 text-cyan-400" />
            <span>Tambah FAQ Baru</span>
            <ArrowRight class="w-4 h-4 arrow" />
          </router-link>
          <router-link to="/conversations" class="shortcut-card">
            <MessageSquare class="w-6 h-6 text-emerald-400" />
            <span>Live Monitoring Chat</span>
            <ArrowRight class="w-4 h-4 arrow" />
          </router-link>
        </div>
      </div>

      <!-- Realtime Activity Log -->
      <div class="section-card glass-panel flex-1">
        <div class="section-header">
          <Activity class="w-5 h-5 text-emerald-400" />
          <h3>Aktivitas System Realtime</h3>
          <span class="live-dot pulse-dot pulse-dot-online ml-auto"></span>
        </div>
        <div class="activity-list">
          <div v-if="recentConversations.length === 0" class="empty-state">
            <p>Belum ada percakapan terbaru.</p>
          </div>
          <div
            v-for="conv in recentConversations"
            :key="conv.id"
            class="activity-item"
            @click="router.push('/conversations')"
          >
            <div class="activity-avatar">
              {{ conv.customer_wa_number.slice(-4) }}
            </div>
            <div class="activity-details">
              <div class="activity-title">
                <strong>+{{ conv.customer_wa_number }}</strong>
                <span :class="['badge', getStatusBadgeClass(conv.status)]">
                  {{ conv.status }}
                </span>
              </div>
              <span class="activity-time">{{ formatDate(conv.updated_at) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  Cpu,
  Settings,
  Smartphone,
  Bot,
  BookOpen,
  MessageSquare,
  Zap,
  ArrowRight,
  PlusCircle,
  Activity,
} from 'lucide-vue-next'

import StatCard from '../components/StatCard.vue'
import { useSessionStore } from '../stores/sessions'
import { useLLMStore } from '../stores/llm'
import { useKnowledgeStore } from '../stores/knowledge'
import { useConversationsStore } from '../stores/conversations'

const router = useRouter()
const sessionStore = useSessionStore()
const llmStore = useLLMStore()
const knowledgeStore = useKnowledgeStore()
const convStore = useConversationsStore()

onMounted(async () => {
  await Promise.all([
    sessionStore.fetchSessions(),
    llmStore.fetchConfigs(),
    knowledgeStore.fetchKnowledge(),
    convStore.fetchConversations(),
  ])
})

const activeLLM = computed(() => llmStore.activeConfig)
const totalSessions = computed(() => sessionStore.sessions.length)
const onlineSessions = computed(() => sessionStore.sessions.filter((s) => s.status === 'online').length)
const activeBots = computed(() => sessionStore.sessions.filter((s) => s.is_bot_enabled).length)
const totalKnowledge = computed(() => knowledgeStore.entries.length)
const totalConversations = computed(() => convStore.conversations.length)
const escalatedConvs = computed(() => convStore.conversations.filter((c) => c.status === 'escalation').length)
const recentConversations = computed(() => convStore.conversations.slice(0, 5))

function getStatusBadgeClass(status: string) {
  switch (status) {
    case 'bot': return 'badge-primary'
    case 'escalation': return 'badge-warning'
    case 'done': return 'badge-online'
    default: return 'badge-offline'
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.dashboard-overview {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.active-llm-banner {
  padding: 20px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: linear-gradient(135deg, rgba(19, 27, 46, 0.9) 0%, rgba(15, 23, 42, 0.9) 100%);
  border-left: 4px solid var(--accent-cyan);
}

.banner-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.banner-badge {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  background: rgba(6, 182, 212, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
}

.banner-info {
  display: flex;
  flex-direction: column;

  h3 {
    font-size: 16px;
    font-weight: 700;
  }

  p {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 2px;
  }
}

.banner-tag {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--accent-cyan-light);
  font-weight: 700;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 20px;
}

.overview-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.section-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;

  h3 {
    font-size: 16px;
    font-weight: 700;
  }
}

.shortcuts-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.shortcut-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-main);
  text-decoration: none;
  font-weight: 600;
  font-size: 14px;
  transition: all 0.2s ease;

  .arrow {
    margin-left: auto;
    color: var(--text-muted);
    transition: transform 0.2s;
  }

  &:hover {
    background: rgba(255, 255, 255, 0.07);
    border-color: rgba(99, 102, 241, 0.4);

    .arrow {
      transform: translateX(4px);
      color: var(--primary-light);
    }
  }
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--border-color);
  cursor: pointer;
  transition: background 0.2s;

  &:hover {
    background: rgba(255, 255, 255, 0.06);
  }
}

.activity-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(99, 102, 241, 0.2);
  color: var(--primary-light);
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.activity-details {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.activity-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
}

.activity-time {
  font-size: 11px;
  color: var(--text-muted);
}

.empty-state {
  text-align: center;
  padding: 30px;
  color: var(--text-muted);
  font-size: 13px;
}

@media (max-width: 900px) {
  .overview-content {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 600px) {
  .active-llm-banner {
    flex-direction: column;
    align-items: flex-start;
    gap: 14px;
  }
  .stat-grid {
    grid-template-columns: 1fr;
    gap: 14px;
  }
}
</style>
