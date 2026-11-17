<script setup lang="ts">
import { computed } from 'vue'
import { projectById, type Note } from '../data'

const props = defineProps<{ note: Note }>()
const project = computed(() => projectById(props.note.projectId))
</script>

<template>
  <RouterLink :to="`/notes/${note.slug}`" class="note-row">
    <div>
      <div class="note-title">{{ note.title }}</div>
      <p class="note-summary">{{ note.summary }}</p>
      <div class="note-meta">
        <span class="tag" v-if="project">{{ project.name }}</span>
        <span>{{ note.publishedAt }}</span>
      </div>
    </div>
    <span class="link-more">阅读全文 →</span>
  </RouterLink>
</template>
