<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import SiteHeader from './components/SiteHeader.vue'
import SiteFooter from './components/SiteFooter.vue'
import ThemeSwitch from './components/ThemeSwitch.vue'

const route = useRoute()
const isAdmin = computed(() => route.path.startsWith('/admin'))
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
