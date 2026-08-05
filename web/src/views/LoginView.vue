<template>
  <div class="login-wrapper">
    <div class="login-card glass-panel">
      <div class="brand-header">
        <div class="brand-badge">
          <Bot class="w-8 h-8 text-indigo-400" />
        </div>
        <h1 class="gradient-text">GNET Bot Assistant</h1>
        <p class="subtitle">Dashboard Pengelola Chatbot WhatsApp</p>
      </div>

      <!-- Error Alert -->
      <div v-if="errorMsg" class="alert-box alert-danger">
        <AlertCircle class="w-5 h-5 flex-shrink-0" />
        <span>{{ errorMsg }}</span>
      </div>

      <!-- Mode Selector (Login vs Register First Admin) -->
      <div class="mode-tabs">
        <button
          :class="['mode-tab', { active: isLogin }]"
          @click="isLogin = true; errorMsg = ''"
        >
          Masuk
        </button>
        <button
          :class="['mode-tab', { active: !isLogin }]"
          @click="isLogin = false; errorMsg = ''"
        >
          Daftar Admin Baru
        </button>
      </div>

      <form @submit.prevent="handleSubmit" class="login-form">
        <div class="input-group">
          <label class="input-label">Email Admin / Agent</label>
          <div class="input-with-icon">
            <Mail class="input-icon" />
            <input
              v-model="email"
              type="email"
              class="form-input"
              placeholder="admin@gnet.co.id"
              required
            />
          </div>
        </div>

        <div class="input-group">
          <label class="input-label">Password</label>
          <div class="input-with-icon">
            <Lock class="input-icon" />
            <input
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              class="form-input"
              placeholder="••••••••"
              required
            />
            <button
              type="button"
              class="eye-btn"
              @click="showPassword = !showPassword"
            >
              <Eye v-if="!showPassword" class="w-4 h-4" />
              <EyeOff v-else class="w-4 h-4" />
            </button>
          </div>
        </div>

        <div v-if="!isLogin" class="input-group">
          <label class="input-label">Role Pengguna</label>
          <select v-model="role" class="form-select">
            <option value="admin">Admin (Akses Penuh)</option>
            <option value="agent">Agent / Teknisi (Monitoring & Takeover)</option>
          </select>
        </div>

        <button type="submit" class="btn btn-primary btn-block" :disabled="loading">
          <Loader2 v-if="loading" class="w-5 h-5 spin" />
          <span v-else>{{ isLogin ? 'Masuk ke Dashboard' : 'Daftar Akun Baru' }}</span>
        </button>
      </form>

      <div class="card-footer">
        <p>© 2026 PT. Ghaib Network (GNET). Standalone AI System.</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Bot, Mail, Lock, Eye, EyeOff, AlertCircle, Loader2 } from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const isLogin = ref(true)
const email = ref('')
const password = ref('')
const role = ref('admin')
const showPassword = ref(false)
const loading = ref(false)
const errorMsg = ref('')

async function handleSubmit() {
  errorMsg.value = ''
  loading.value = true
  try {
    if (isLogin.value) {
      await authStore.login(email.value, password.value)
    } else {
      await authStore.register(email.value, password.value, role.value)
    }
    router.push('/')
  } catch (err: any) {
    errorMsg.value = err.message || 'Terjadi kesalahan saat masuk'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 440px;
  padding: 36px 32px;
}

@media (max-width: 480px) {
  .login-card {
    padding: 24px 20px;
  }
}

.brand-header {
  text-align: center;
  margin-bottom: 28px;
}

.brand-badge {
  width: 60px;
  height: 60px;
  margin: 0 auto 16px;
  border-radius: 18px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.25) 0%, rgba(6, 182, 212, 0.25) 100%);
  border: 1px solid rgba(99, 102, 241, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 25px rgba(99, 102, 241, 0.2);
}

.brand-header h1 {
  font-size: 24px;
  font-weight: 800;
  letter-spacing: -0.5px;
}

.subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.alert-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-radius: var(--radius-md);
  font-size: 13px;
  margin-bottom: 20px;
}

.alert-danger {
  background: var(--color-danger-bg);
  border: 1px solid rgba(244, 63, 94, 0.3);
  color: #ff6b81;
}

.mode-tabs {
  display: flex;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 4px;
  margin-bottom: 24px;
}

.mode-tab {
  flex: 1;
  padding: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s ease;
}

.mode-tab.active {
  background: var(--primary);
  color: #ffffff;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.3);
}

.input-with-icon {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 12px;
  width: 18px;
  height: 18px;
  color: var(--text-muted);
}

.input-with-icon .form-input {
  padding-left: 38px;
  padding-right: 38px;
}

.eye-btn {
  position: absolute;
  right: 12px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
}

.btn-block {
  width: 100%;
  padding: 12px;
  font-size: 15px;
  margin-top: 8px;
}

.card-footer {
  margin-top: 28px;
  text-align: center;
  font-size: 11px;
  color: var(--text-muted);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
