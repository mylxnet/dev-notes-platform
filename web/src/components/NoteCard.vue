<script setup lang="ts">
import { computed } from 'vue'
import { formatDate, type Note } from '../api'

const props = defineProps<{ note: Note }>()
const projectName = computed(() => props.note.project?.name ?? '')
const publishedDate = computed(() => formatDate(props.note.published_at))
</script>

<template>
  <RouterLink :to="`/notes/${note.slug}`" class="note-row">
    <div>
      <div class="note-title">{{ note.title }}</div>
      <p class="note-summary">{{ note.summary }}</p>
      <div class="note-meta">
        <span class="tag" v-if="projectName">{{ projectName }}</span>
        <span v-if="publishedDate">{{ publishedDate }}</span>
      </div>
    </div>
    <span class="link-more">阅读全文 →</span>
  </RouterLink>
</template>
