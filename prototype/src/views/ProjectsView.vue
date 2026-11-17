<script setup lang="ts">
import { notes, publishedProjects } from '../data'

const projects = publishedProjects()
function noteCount(projectId: number) {
  return notes.filter((item) => item.projectId === projectId && item.status === 'published').length
}
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
    <div class="project-grid">
      <RouterLink v-for="project in projects" :key="project.id" :to="`/projects/${project.slug}`" class="project-card">
        <div class="project-top">
          <span class="project-name">{{ project.name }}</span>
          <span class="status status-published">已发布</span>
        </div>
        <p class="project-summary">{{ project.summary }}</p>
        <div class="tech-list">
          <span v-for="tech in project.tech" :key="tech">{{ tech }}</span>
        </div>
        <div class="card-foot">
          <span>{{ noteCount(project.id) }} 篇关联笔记</span>
          <span>查看详情 →</span>
        </div>
      </RouterLink>
    </div>
  </div>
</template>
