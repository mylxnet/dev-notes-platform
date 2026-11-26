<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

type ThemeMode = 'system' | 'light' | 'dark'

/** 点击顺序：跟随系统 → 浅色 → 深色 → 跟随系统。 */
const MODES: ThemeMode[] = ['system', 'light', 'dark']
const STORAGE_KEY = 'dev-notes-theme-mode'
/** 旧版本只存 light/dark，读取时迁移到新键，避免用户设置丢失。 */
const LEGACY_KEY = 'dev-notes-theme'
const LABELS: Record<ThemeMode, string> = { system: '跟随系统', light: '浅色', dark: '深色' }

const mode = ref<ThemeMode>('system')
const systemDark = ref(false)
let media: MediaQueryList | null = null

/** 手动模式优先；跟随系统时按系统偏好推导实际主题。 */
function resolved(): 'light' | 'dark' {
  if (mode.value === 'system') return systemDark.value ? 'dark' : 'light'
  return mode.value
}

function applyTheme() {
  document.documentElement.dataset.theme = resolved()
}

function readSavedMode(): ThemeMode {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved === 'system' || saved === 'light' || saved === 'dark') return saved
    const legacy = localStorage.getItem(LEGACY_KEY)
    if (legacy === 'light' || legacy === 'dark') {
      localStorage.setItem(STORAGE_KEY, legacy)
      return legacy
    }
  } catch {
    // 隐私模式下 localStorage 不可用，按跟随系统处理。
  }
  return 'system'
}

function cycleTheme() {
  mode.value = MODES[(MODES.indexOf(mode.value) + 1) % MODES.length]
  try {
    localStorage.setItem(STORAGE_KEY, mode.value)
  } catch {
    // 存不进去也不影响本次切换。
  }
  applyTheme()
}

/** 系统主题变化时只有「跟随系统」需要即时重绘。 */
function onSystemChange(event: MediaQueryListEvent) {
  systemDark.value = event.matches
  if (mode.value === 'system') applyTheme()
}

onMounted(() => {
  mode.value = readSavedMode()
  media = window.matchMedia('(prefers-color-scheme: dark)')
  systemDark.value = media.matches
  media.addEventListener('change', onSystemChange)
  applyTheme()
})

onUnmounted(() => {
  media?.removeEventListener('change', onSystemChange)
  media = null
})
</script>

<template>
  <button
    class="theme-float"
    type="button"
    :aria-label="`当前为${LABELS[mode]}，点击切换主题`"
    :title="`当前为${LABELS[mode]}，点击切换`"
    @click="cycleTheme"
  >
    <span class="theme-dot" :class="{ system: mode === 'system' }" aria-hidden="true"></span>
    <span class="theme-text">{{ LABELS[mode] }}</span>
    <span class="theme-suffix">主题</span>
  </button>
</template>
