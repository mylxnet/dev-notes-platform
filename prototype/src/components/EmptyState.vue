<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <div class="empty">
    <strong>{{ title }}</strong>
    <p v-if="description" style="margin: 8px 0 0">{{ description }}</p>
  </div>
</template>
