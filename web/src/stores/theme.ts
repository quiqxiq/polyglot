import { ref } from 'vue'

export type Theme = 'dark' | 'light'

const savedTheme = (localStorage.getItem('theme') as Theme) || 'dark'
export const currentTheme = ref<Theme>(savedTheme)

export function applyTheme(theme: Theme) {
  currentTheme.value = theme
  localStorage.setItem('theme', theme)
  document.documentElement.setAttribute('data-theme', theme)
}

// Initial theme application
applyTheme(currentTheme.value)

export function useTheme() {
  function toggleTheme() {
    const nextTheme = currentTheme.value === 'dark' ? 'light' : 'dark'
    applyTheme(nextTheme)
  }

  return {
    currentTheme,
    toggleTheme,
    setTheme: applyTheme,
  }
}
