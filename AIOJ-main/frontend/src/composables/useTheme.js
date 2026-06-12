import { ref, watch } from 'vue'

const THEME_KEY = 'oj-theme'
const theme = ref(localStorage.getItem(THEME_KEY) || 'light')

function applyTheme(t) {
  document.documentElement.setAttribute('data-theme', t)
  // Element Plus dark mode class
  if (t === 'dark') {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

// Apply on load
applyTheme(theme.value)

watch(theme, (val) => {
  localStorage.setItem(THEME_KEY, val)
  applyTheme(val)
})

export function useTheme() {
  function toggleTheme() {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
  }

  function setTheme(t) {
    theme.value = t
  }

  return {
    theme,
    toggleTheme,
    setTheme,
    isDark: () => theme.value === 'dark'
  }
}
