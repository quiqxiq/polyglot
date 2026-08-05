<template>
  <div class="stat-card glass-panel">
    <div class="stat-header">
      <span class="stat-title">{{ title }}</span>
      <div :class="['stat-icon-wrapper', iconBgClass]">
        <component :is="icon" class="stat-icon" />
      </div>
    </div>
    <div class="stat-body">
      <div class="stat-value">{{ value }}</div>
      <div v-if="subtext" class="stat-subtext">
        <span>{{ subtext }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  title: string
  value: string | number
  subtext?: string
  icon: Component
  iconBgClass?: string
}>()
</script>

<style scoped>
.stat-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 2px;
  background: linear-gradient(90deg, var(--primary) 0%, var(--accent-cyan) 100%);
  opacity: 0.6;
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.stat-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.stat-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(99, 102, 241, 0.15);
  color: var(--primary-light);
}

.stat-icon {
  width: 20px;
  height: 20px;
}

.stat-value {
  font-size: 28px;
  font-weight: 800;
  color: var(--text-main);
  line-height: 1.2;
}

.stat-subtext {
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-muted);
}

.icon-bg-primary {
  background: rgba(99, 102, 241, 0.15);
  color: var(--primary-light);
}

.icon-bg-emerald {
  background: rgba(16, 185, 129, 0.15);
  color: var(--color-success);
}

.icon-bg-cyan {
  background: rgba(6, 182, 212, 0.15);
  color: var(--accent-cyan-light);
}

.icon-bg-amber {
  background: rgba(245, 158, 11, 0.15);
  color: var(--color-warning);
}
</style>
