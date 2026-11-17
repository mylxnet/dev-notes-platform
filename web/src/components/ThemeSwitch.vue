<script setup lang="ts">
import { onMounted, ref } from 'vue'

const theme = ref<'light' | 'dark'>('light')

function applyTheme() {
  document.documentElement.dataset.theme = theme.value
}

function toggleTheme() {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
  localStorage.setItem('dev-notes-theme', theme.value)
  applyTheme()
}

onMounted(() => {
  const saved = localStorage.getItem('dev-notes-theme')
  if (saved === 'dark' || saved === 'light') theme.value = saved
  applyTheme()
})
</script>

<template>
  <button
    class="theme-float"
    type="button"
    :aria-label="`当前为${theme === 'light' ? '浅色' : '深色'}主题，点击切换`"
    @click="toggleTheme"
  >
    <span class="theme-dot" aria-hidden="true"></span>
    <span>{{ theme === 'light' ? '浅色' : '深色' }}</span>
    <span>主题</span>
  </button>
</template>
