<template>
  <div v-if="show" class="modal-overlay" @click.self="$emit('cancel')">
    <div class="modal-card glass-panel">
      <div class="modal-header">
        <div class="header-icon-box" :class="variantClass">
          <AlertTriangle class="w-6 h-6" />
        </div>
        <div class="header-text">
          <h3>{{ title || 'Konfirmasi Tindakan' }}</h3>
          <p class="subtitle">{{ subtitle || 'Tindakan ini memerlukan konfirmasi Anda.' }}</p>
        </div>
        <button class="close-btn" @click="$emit('cancel')">
          <X class="w-5 h-5" />
        </button>
      </div>

      <div class="modal-body">
        <p class="confirm-message">{{ message }}</p>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" @click="$emit('cancel')">
          Batal
        </button>
        <button
          :class="['btn', confirmButtonClass]"
          :disabled="loading"
          @click="$emit('confirm')"
        >
          <Loader2 v-if="loading" class="w-4 h-4 spin mr-2" />
          <span>{{ confirmText || 'Ya, Lanjutkan' }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { AlertTriangle, X, Loader2 } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    show: boolean
    title?: string
    subtitle?: string
    message: string
    confirmText?: string
    variant?: 'danger' | 'warning' | 'info'
    loading?: boolean
  }>(),
  {
    variant: 'danger',
    loading: false,
  }
)

defineEmits(['confirm', 'cancel'])

const variantClass = computed(() => {
  switch (props.variant) {
    case 'warning': return 'icon-warning'
    case 'info': return 'icon-info'
    default: return 'icon-danger'
  }
})

const confirmButtonClass = computed(() => {
  switch (props.variant) {
    case 'warning': return 'btn-warning'
    case 'info': return 'btn-primary'
    default: return 'btn-danger'
  }
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(9, 13, 22, 0.85);
  backdrop-filter: blur(10px);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.modal-card {
  width: 100%;
  max-width: 420px;
  padding: 24px;
}

.modal-header {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 16px;
  position: relative;
}

.header-icon-box {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.icon-danger {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.icon-warning {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.icon-info {
  background: rgba(99, 102, 241, 0.15);
  color: #6366f1;
  border: 1px solid rgba(99, 102, 241, 0.3);
}

.header-text {
  flex: 1;

  h3 {
    font-size: 17px;
    font-weight: 700;
    margin: 0 0 2px 0;
    color: var(--text-main);
  }
}

.subtitle {
  font-size: 12px;
  color: var(--text-muted);
  margin: 0;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;

  &:hover {
    color: var(--text-main);
  }
}

.confirm-message {
  font-size: 13.5px;
  color: var(--text-secondary);
  line-height: 1.5;
  margin: 0;
  background: var(--bg-input);
  padding: 12px 14px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
}

.modal-footer {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
