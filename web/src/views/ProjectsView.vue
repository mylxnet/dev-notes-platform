<script setup lang="ts">
import { onMounted, ref } from 'vue'
import EmptyState from '../components/EmptyState.vue'
import { errorMessageOf, fetchProjects, techStackOf, type Project } from '../api'

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
</script>

<template>
  <div class="container section">
    <div class="section-head">
      <div>
        <h1 class="section-title">项目索引</h1>
        <div class="section-subtitle">个人练习作品记录 · 仅提供源码外链</div>
      </div>
      <span class="section-subtitle">已公开 {{ projects.length }} 个</span>
    </div>

    <div class="callout">
      本站不托管安装包或可执行文件。受许可证约束，项目入口仅链接到 GitHub 源码仓库。
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
        <div class="project-top">
          <span class="project-name">{{ project.name }}</span>
          <span class="status status-published">已发布</span>
        </div>
        <p class="project-summary">{{ project.summary }}</p>
        <div class="tech-list">
          <span v-for="tech in techStackOf(project)" :key="tech">{{ tech }}</span>
        </div>
        <div class="card-foot">
          <span>{{ project.note_count }} 篇关联笔记</span>
          <span>查看详情 →</span>
        </div>
      </RouterLink>
    </div>
    <EmptyState v-else title="还没有已公开的项目" description="项目记录整理完成后会显示在这里。" />
  </div>
</template>
