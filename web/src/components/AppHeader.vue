<template>
  <header class="app-header glass-panel">
    <button class="mobile-toggle" @click="$emit('toggleSidebar')">
      <Menu class="w-6 h-6" />
    </button>

    <div class="header-title">
      <h2>{{ pageTitle }}</h2>
    </div>

    <div class="header-actions">
      <!-- Realtime SSE Connection Status Badge -->
      <div
        :class="['badge', realtimeStore.isConnected ? 'badge-online' : 'badge-offline']"
        title="Status Koneksi Event Stream Realtime"
      >
        <span :class="['pulse-dot', realtimeStore.isConnected ? 'pulse-dot-online' : '']"></span>
        <span>{{ realtimeStore.isConnected ? 'Live Sync' : 'Connecting...' }}</span>
      </div>

      <!-- Active Provider Badge -->
      <div v-if="activeLLM" class="badge badge-primary" title="Provider LLM Aktif">
        <Cpu class="w-3.5 h-3.5" />
        <span>{{ activeLLM.provider.toUpperCase() }} ({{ activeLLM.model }})</span>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Menu, Cpu } from 'lucide-vue-next'
import { useRealtimeStore } from '../stores/realtime'
import { useLLMStore } from '../stores/llm'

defineEmits(['toggleSidebar'])

const route = useRoute()
const realtimeStore = useRealtimeStore()
const llmStore = useLLMStore()

const activeLLM = computed(() => llmStore.activeConfig)

const pageTitle = computed(() => {
  switch (route.path) {
    case '/sessions':
      return 'Koneksi WhatsApp'
    case '/llm-config':
      return 'Konfigurasi Provider LLM'
    case '/knowledge':
      return 'Basis Pengetahuan (FAQ)'
    case '/conversations':
      return 'Percakapan & Agent Take-over'
    default:
      return 'Overview Dashboard'
  }
})
</script>

<style scoped>
.app-header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  border-radius: 0 0 var(--radius-lg) var(--radius-lg);
  margin-bottom: 24px;
}

.mobile-toggle {
  display: none;
  background: transparent;
  border: none;
  color: var(--text-main);
  cursor: pointer;
}

.header-title h2 {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-main);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

@media (max-width: 768px) {
  .mobile-toggle {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 6px;
    margin-right: 8px;
    border-radius: var(--radius-sm);
  }
  .mobile-toggle:hover {
    background: rgba(255, 255, 255, 0.1);
  }
  .app-header {
    padding: 0 14px;
    height: auto;
    min-height: 60px;
    padding-top: 10px;
    padding-bottom: 10px;
    gap: 8px;
  }
  .header-title h2 {
    font-size: 15px;
    line-height: 1.3;
  }
  .header-actions {
    gap: 6px;
  }
  .badge-primary {
    display: none; /* Hide secondary detailed badge on small screens to prevent header overcrowding */
  }
}

@media (min-width: 481px) and (max-width: 768px) {
  .badge-primary {
    display: inline-flex;
  }
}
</style>
