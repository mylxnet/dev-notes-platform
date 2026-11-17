<script setup lang="ts">
import { computed, ref } from 'vue'
import NoteCard from '../components/NoteCard.vue'
import EmptyState from '../components/EmptyState.vue'
import { publishedNotes, publishedProjects } from '../data'

const activeProject = ref<number | null>(null)
const keyword = ref('')
const projects = publishedProjects()

const filteredNotes = computed(() => {
  const text = keyword.value.trim().toLowerCase()
  return publishedNotes().filter((note) => {
    const matchProject = activeProject.value === null || note.projectId === activeProject.value
    const matchText = !text || note.title.toLowerCase().includes(text) || note.summary.toLowerCase().includes(text)
    return matchProject && matchText
  })
})
</script>

<template>
  <div class="container section">
    <div class="section-head">
      <div>
        <h1 class="section-title">笔记列表</h1>
        <div class="section-subtitle">按日期倒序 · 支持按项目筛选</div>
      </div>
      <span class="section-subtitle">共 {{ filteredNotes.length }} 篇</span>
    </div>

    <div class="filters">
      <button class="chip" :class="{ active: activeProject === null }" type="button" @click="activeProject = null">全部</button>
      <button
        v-for="project in projects"
        :key="project.id"
        class="chip"
        :class="{ active: activeProject === project.id }"
        type="button"
        @click="activeProject = project.id"
      >
        {{ project.name }}
      </button>
      <input v-model="keyword" class="search" type="search" placeholder="搜索笔记标题或摘要" aria-label="搜索笔记" />
    </div>

    <div v-if="filteredNotes.length" class="note-list">
      <NoteCard v-for="note in filteredNotes" :key="note.id" :note="note" />
    </div>
    <EmptyState v-else title="没有匹配的笔记" description="换个关键词，或选择「全部」项目筛选。" />
  </div>
</template>
