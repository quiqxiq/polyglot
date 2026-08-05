<template>
  <!-- Full Screen Wrapper -->
  <div v-if="isLoginPage">
    <router-view />
  </div>

  <div v-else class="app-layout">
    <!-- Mobile Sidebar Backdrop Overlay -->
    <div
      v-if="sidebarOpen"
      class="sidebar-backdrop"
      @click="sidebarOpen = false"
    ></div>

    <!-- Sidebar -->
    <AppSidebar v-model:mobileOpen="sidebarOpen" />

    <!-- Main Content Area -->
    <main class="main-content">
      <AppHeader @toggleSidebar="sidebarOpen = !sidebarOpen" />
      <div class="content-body">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import AppSidebar from './components/AppSidebar.vue'
import AppHeader from './components/AppHeader.vue'
import { useAuthStore } from './stores/auth'
import { useRealtimeStore } from './stores/realtime'

const route = useRoute()
const authStore = useAuthStore()
const realtimeStore = useRealtimeStore()

const sidebarOpen = ref(false)

const isLoginPage = computed(() => route.path === '/login')

watch(
  () => route.path,
  () => {
    sidebarOpen.value = false
  }
)

onMounted(() => {
  if (authStore.isAuthenticated) {
    realtimeStore.connect()
  }
})
</script>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
}

.main-content {
  flex: 1;
  margin-left: 260px;
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding-bottom: 24px;
}

.content-body {
  padding: 0 24px;
  flex: 1;
}

@media (max-width: 768px) {
  .main-content {
    margin-left: 0;
  }
  .content-body {
    padding: 0 16px;
  }
}
</style>
