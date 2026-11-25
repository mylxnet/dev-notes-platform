<script setup lang="ts">
import { onMounted, ref } from 'vue'
import EmptyState from '../components/EmptyState.vue'
import { errorMessageOf, fetchProjects, projectStatusText, techStackOf, type Project } from '../api'

const projects = ref<Project[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await fetchProjects()
    projects.value = result.items
  } catch (err) {
    error.value = errorMessageOf(err)
    projects.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)

function initialOf(name: string): string {
  const head = name.trim().slice(0, 1)
  return head ? head.toUpperCase() : '·'
}
</script>

<template>
  <div class="container section">
    <div class="section-head">
      <div>
        <h1 class="section-title">项目索引</h1>
        <div class="section-subtitle">个人练习作品，共 {{ projects.length }} 个</div>
      </div>
    </div>

    <div v-if="error" class="error-banner" role="alert">
      {{ error }}
      <button class="link-more" type="button" style="margin-left: 12px" @click="load">重试</button>
    </div>

    <div v-if="loading" class="skeleton-block" aria-busy="true">
      <div v-for="index in 4" :key="index">
        <div class="skeleton-line" style="width: 44%" />
        <div class="skeleton-line" style="width: 84%" />
        <div class="skeleton-line" style="width: 60%" />
      </div>
    </div>
    <div v-else-if="projects.length" class="project-grid">
      <RouterLink v-for="project in projects" :key="project.id" :to="`/projects/${project.slug}`" class="project-card">
        <div class="project-thumb" :class="{ 'is-empty': !project.cover_image }">
          <img
            v-if="project.cover_image"
            :src="project.cover_image"
            :alt="`${project.name} 封面`"
            loading="lazy"
          />
          <span v-else class="project-thumb-initial">{{ initialOf(project.name) }}</span>
        </div>
        <div class="project-body">
          <div class="project-top">
            <span class="project-name">{{ project.name }}</span>
            <span class="status" :class="`status-${project.status}`">
              {{ projectStatusText[project.status] }}
            </span>
          </div>
          <p class="project-summary">{{ project.summary }}</p>
          <div class="tech-list">
            <span v-for="tech in techStackOf(project)" :key="tech">{{ tech }}</span>
          </div>
          <div class="card-foot">
            <span>{{ project.note_count }} 篇关联笔记</span>
            <span class="card-go">查看详情 →</span>
          </div>
        </div>
      </RouterLink>
    </div>
    <EmptyState v-else title="还没有已公开的项目" description="项目记录整理完成后会显示在这里。" />
  </div>
</template>
