<template>
  <Teleport to="body">
    <div
      v-if="show"
      class="modal-overlay"
      @click.self="close"
    >
      <div class="terminal-modal-card glass-panel animate-fade-in">
        <!-- Modal Header -->
        <div class="terminal-modal-header">
          <div class="flex items-center gap-3">
            <div class="terminal-icon-box">
              <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                />
              </svg>
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="font-bold text-lg text-main">{{ deviceName || 'Device Terminal' }}</h3>
                <span class="badge badge-info uppercase">
                  {{ vendor || 'CLI' }}
                </span>
                <span class="badge badge-success">
                  ● PTY Session Active
                </span>
              </div>
              <p class="text-xs text-muted mt-1">Device ID: {{ deviceId }}</p>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex items-center gap-2">
            <button
              @click="clearTerminal"
              class="btn btn-sm btn-secondary"
              title="Clear terminal screen"
            >
              Clear
            </button>
            <button
              @click="reconnect"
              class="btn btn-sm btn-primary"
              title="Reconnect PTY session"
            >
              Reconnect
            </button>
            <button
              @click="close"
              class="btn btn-sm btn-secondary"
              title="Tutup Modal Terminal"
            >
              ✕
            </button>
          </div>
        </div>

        <!-- Terminal Viewport -->
        <div class="terminal-viewport" ref="terminalContainer"></div>

        <!-- Status Bar -->
        <div class="terminal-modal-footer">
          <div class="flex items-center gap-4">
            <span>Tekan <kbd class="kbd">TAB</kbd> untuk Auto-Completion</span>
            <span><kbd class="kbd">Ctrl+C</kbd> Interrupt</span>
          </div>
          <div class="text-muted font-mono">Xterm.js Interactive PTY Session</div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted, nextTick } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

const props = defineProps<{
  show: boolean
  deviceId: string
  deviceName?: string
  vendor?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const terminalContainer = ref<HTMLElement | null>(null)
let term: Terminal | null = null
let fitAddon: FitAddon | null = null
let socket: WebSocket | null = null

function close() {
  cleanup()
  emit('close')
}

function clearTerminal() {
  if (term) {
    term.clear()
  }
}

function reconnect() {
  cleanup()
  nextTick(() => {
    initTerminal()
  })
}

function cleanup() {
  if (socket) {
    socket.close()
    socket = null
  }
  if (term) {
    term.dispose()
    term = null
  }
  fitAddon = null
}

async function initTerminal() {
  if (!props.show || !props.deviceId || !terminalContainer.value) return

  // 1. Create Xterm Instance with NetOps Engine Dark Theme
  term = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    theme: {
      background: '#090d16',
      foreground: '#e2e8f0',
      cursor: '#38bdf8',
      selectionBackground: '#334155',
      black: '#090d16',
      red: '#ef4444',
      green: '#10b981',
      yellow: '#f59e0b',
      blue: '#3b82f6',
      magenta: '#ec4899',
      cyan: '#06b6d4',
      white: '#f8fafc',
    },
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalContainer.value)
  fitAddon.fit()

  term.writeln('\x1b[38;5;39mConnecting to ' + (props.deviceName || props.deviceId) + ' Interactive PTY Session...\x1b[0m\r\n')

  const wsUrl = `ws://${window.location.hostname}:8080/ws/devices/${encodeURIComponent(props.deviceId)}/terminal`
  socket = new WebSocket(wsUrl)

  socket.onopen = () => {
    socket?.send(JSON.stringify({
      device_id: props.deviceId,
      cols: term?.cols,
      rows: term?.rows,
    }))
  }

  socket.onmessage = (event) => {
    if (term) {
      term.write(event.data)
    }
  }

  socket.onerror = () => {
    if (term) {
      term.writeln('\r\n\x1b[31mTerminal PTY Connection Error\x1b[0m')
    }
  }

  socket.onclose = () => {
    if (term) {
      term.writeln('\r\n\x1b[33mSession Closed\x1b[0m')
    }
  }

  term.onData((data) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        device_id: props.deviceId,
        input_data: data,
      }))
    }
  })

  term.onResize(({ cols, rows }) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        device_id: props.deviceId,
        cols,
        rows,
      }))
    }
  })
}

watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      nextTick(() => {
        initTerminal()
      })
    } else {
      cleanup()
    }
  },
  { immediate: true }
)

onUnmounted(() => {
  cleanup()
})
</script>

<style scoped>
.terminal-modal-card {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 1050px;
  height: 82vh;
  background: var(--bg-card-solid, #131b2e);
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.12));
  border-radius: var(--radius-lg, 16px);
  overflow: hidden;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.7);
}

.terminal-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  background: rgba(19, 27, 46, 0.95);
  border-bottom: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
}

.terminal-icon-box {
  padding: 8px;
  border-radius: 8px;
  background: rgba(56, 189, 248, 0.1);
  color: #38bdf8;
  border: 1px solid rgba(56, 189, 248, 0.2);
}

.terminal-viewport {
  flex: 1;
  background: #090d16;
  padding: 12px;
  overflow: hidden;
  position: relative;
}

.terminal-modal-footer {
  padding: 10px 24px;
  background: rgba(15, 23, 42, 0.95);
  border-top: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
  font-size: 12px;
  color: var(--text-muted, #94a3b8);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.kbd {
  padding: 2px 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #f8fafc;
}

:deep(.xterm) {
  height: 100%;
  padding: 4px;
}
</style>
