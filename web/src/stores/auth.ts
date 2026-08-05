import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '../types'
import { loginApi, registerApi, getMeApi } from '../api/client'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(
    localStorage.getItem('gnet_user')
      ? JSON.parse(localStorage.getItem('gnet_user')!)
      : null
  )

  const token = ref<string | null>(localStorage.getItem('gnet_token') || null)

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const userEmail = computed(() => user.value?.email || '')

  async function login(email: string, pass: string) {
    const res = await loginApi(email, pass)
    token.value = res.token
    user.value = res.user
    localStorage.setItem('gnet_token', res.token)
    localStorage.setItem('gnet_user', JSON.stringify(res.user))
  }

  async function register(email: string, pass: string, role?: string) {
    const res = await registerApi(email, pass, role)
    token.value = res.token
    user.value = res.user
    localStorage.setItem('gnet_token', res.token)
    localStorage.setItem('gnet_user', JSON.stringify(res.user))
  }

  async function fetchMe() {
    if (!token.value) return
    try {
      const res = await getMeApi()
      user.value = res.user
      localStorage.setItem('gnet_user', JSON.stringify(res.user))
    } catch {
      logout()
    }
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('gnet_token')
    localStorage.removeItem('gnet_user')
  }

  return {
    user,
    token,
    isAuthenticated,
    isAdmin,
    userEmail,
    login,
    register,
    fetchMe,
    logout,
  }
})
