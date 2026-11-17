<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import SiteHeader from './components/SiteHeader.vue'
import SiteFooter from './components/SiteFooter.vue'
import ThemeSwitch from './components/ThemeSwitch.vue'
import { loadSiteInfo } from './site'

const route = useRoute()
const isAdmin = computed(() => route.path.startsWith('/admin'))

// 页脚备案号由公开接口提供；后台页面不渲染页脚，但保存后会主动刷新这份状态。
onMounted(() => void loadSiteInfo())
</script>

<template>
  <div class="app-shell">
    <SiteHeader v-if="!isAdmin" />
    <main style="flex: 1">
      <RouterView v-slot="{ Component, route: currentRoute }">
        <Transition name="fade" mode="out-in">
          <component :is="Component" :key="currentRoute.fullPath" />
        </Transition>
      </RouterView>
    </main>
    <SiteFooter v-if="!isAdmin" />
    <ThemeSwitch />
  </div>
</template>
