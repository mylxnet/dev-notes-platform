<script setup lang="ts">
import { onMounted, ref } from 'vue'
import NoteCard from '../components/NoteCard.vue'
import EmptyState from '../components/EmptyState.vue'
import {
  errorMessageOf,
  fetchNotes,
  fetchProjects,
  formatDate,
  techStackOf,
  type Note,
  type Project
} from '../api'

const panelNotes = ref<Note[]>([])
const recentNotes = ref<Note[]>([])
const indexProjects = ref<Project[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [notesResult, projectsResult] = await Promise.all([
      fetchNotes({ pageSize: 4 }),
      fetchProjects()
    ])
    recentNotes.value = notesResult.items
    panelNotes.value = notesResult.items.slice(0, 3)
    indexProjects.value = projectsResult.items.slice(0, 2)
  } catch (err) {
    error.value = errorMessageOf(err)
    recentNotes.value = []
    panelNotes.value = []
    indexProjects.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <section class="container hero">
      <div>
        <div class="eyebrow">个人开发记录</div>
        <h1>把写过的代码，<br />整理成能回看的笔记</h1>
        <p class="lead">
          这里记录我在学习软件开发过程中的技术笔记、问题排查与项目复盘。内容仅作个人非经营性分享，项目只提供源码入口，不提供任何下载。
        </p>
        <div class="hero-actions">
          <RouterLink to="/notes" class="btn btn-primary">浏览全部笔记</RouterLink>
          <RouterLink to="/projects" class="btn">查看项目索引</RouterLink>
        </div>
      </div>
      <aside class="hero-panel" aria-label="最近更新">
        <div class="panel-head">
          <strong>最近更新</strong>
          <RouterLink to="/notes" class="link-more">全部</RouterLink>
        </div>
        <div class="panel-list">
          <template v-if="loading">
            <div v-for="index in 3" :key="index" class="panel-item">
              <span class="skeleton-line" style="width: 70%; margin: 0" />
              <span class="skeleton-line" style="width: 28%; margin: 0" />
            </div>
          </template>
          <template v-else-if="panelNotes.length">
            <RouterLink v-for="note in panelNotes" :key="note.id" :to="`/notes/${note.slug}`" class="panel-item">
              <span>{{ note.title }}</span>
              <span>{{ formatDate(note.published_at) }}</span>
            </RouterLink>
          </template>
          <div v-else class="panel-item">
            <span>暂无已发布笔记</span>
          </div>
        </div>
      </aside>
    </section>

    <section class="container section">
      <div class="section-head">
        <div>
          <h2 class="section-title">最近笔记</h2>
          <div class="section-subtitle">按时间倒序，记录真实的排查过程与取舍</div>
        </div>
        <RouterLink to="/notes" class="link-more">全部笔记 →</RouterLink>
      </div>

      <div v-if="error" class="error-banner" role="alert">
        {{ error }}
        <button class="link-more" type="button" style="margin-left: 12px" @click="load">重试</button>
      </div>

      <div v-if="loading" class="skeleton-block" aria-busy="true">
        <div v-for="index in 3" :key="index">
          <div class="skeleton-line" style="width: 38%" />
          <div class="skeleton-line" style="width: 78%" />
        </div>
      </div>
      <div v-else-if="recentNotes.length" class="note-list">
        <NoteCard v-for="note in recentNotes" :key="note.id" :note="note" />
      </div>
      <EmptyState v-else title="还没有已发布的笔记" description="笔记整理完成后会显示在这里。" />
    </section>

    <section class="container section">
      <div class="section-head">
        <div>
          <h2 class="section-title">项目索引</h2>
          <div class="section-subtitle">个人练习作品，仅作记录与源码外链</div>
        </div>
        <RouterLink to="/projects" class="link-more">全部项目 →</RouterLink>
      </div>

      <div v-if="loading" class="skeleton-block" aria-busy="true">
        <div v-for="index in 2" :key="index">
          <div class="skeleton-line" style="width: 46%" />
          <div class="skeleton-line" style="width: 86%" />
          <div class="skeleton-line" style="width: 62%" />
        </div>
      </div>
      <div v-else-if="indexProjects.length" class="project-grid">
        <RouterLink v-for="project in indexProjects" :key="project.id" :to="`/projects/${project.slug}`" class="project-card">
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
            <span>→</span>
          </div>
        </RouterLink>
      </div>
      <EmptyState v-else title="还没有已公开的项目" description="项目记录整理完成后会显示在这里。" />
    </section>
  </div>
</template>
