<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import EmptyState from '../components/EmptyState.vue'
import {
  ApiError,
  errorMessageOf,
  fetchProject,
  formatDate,
  projectStatusText,
  techStackOf,
  type Note,
  type Project
} from '../api'

const route = useRoute()

const project = ref<Project | null>(null)
const relatedNotes = ref<Note[]>([])
const loading = ref(true)
const error = ref('')
const notFound = ref(false)

async function load(slug: string) {
  loading.value = true
  error.value = ''
  notFound.value = false
  project.value = null
  relatedNotes.value = []
  try {
    const detail = await fetchProject(slug)
    project.value = detail.project
    relatedNotes.value = detail.notes
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      notFound.value = true
    } else {
      error.value = errorMessageOf(err)
    }
  } finally {
    loading.value = false
  }
}

watch(
  () => String(route.params.slug ?? ''),
  (slug) => {
    if (slug) load(slug)
  },
  { immediate: true }
)
</script>

<template>
  <div class="container section" v-if="loading">
    <div class="section-head" aria-busy="true">
      <div style="width: 100%">
        <div class="skeleton-line" style="width: 12%" />
        <div class="skeleton-line" style="width: 34%; height: 26px" />
        <div class="skeleton-line" style="width: 58%" />
      </div>
    </div>
    <div class="about-grid">
      <div>
        <div class="skeleton-line" style="width: 90%" />
        <div class="skeleton-line" style="width: 76%" />
      </div>
      <div>
        <div class="skeleton-line" style="width: 84%" />
        <div class="skeleton-line" style="width: 64%" />
      </div>
    </div>
  </div>

  <div class="container section" v-else-if="project">
    <div class="section-head">
      <div>
        <div class="eyebrow">项目记录</div>
        <h1 class="section-title">{{ project.name }}</h1>
        <div class="section-subtitle">{{ project.summary }}</div>
      </div>
      <span class="status" :class="`status-${project.status}`">
        {{ projectStatusText[project.status] }}
      </span>
    </div>

    <figure v-if="project.cover_image" class="project-hero">
      <img :src="project.cover_image" :alt="`${project.name} 封面`" />
    </figure>

    <div class="about-grid">
      <div>
        <h2>项目说明</h2>
        <div
          v-if="project.description_html"
          class="article-markdown project-description"
          v-html="project.description_html"
        />
        <p v-else>
          这是个人练习过程中的阶段性记录。本站只描述项目定位与技术取舍，源码入口外链到 GitHub，不提供任何安装包或可执行文件。
        </p>
        <div class="tech-list">
          <span v-for="tech in techStackOf(project)" :key="tech">{{ tech }}</span>
        </div>
        <p>
          <a
            v-if="project.repo_url"
            :href="project.repo_url"
            class="link-more"
            target="_blank"
            rel="noopener noreferrer"
          >
            源码仓库 →
          </a>
          <span v-else class="section-subtitle">源码入口暂未提供</span>
        </p>
      </div>
      <div>
        <h2>关联笔记</h2>
        <div v-if="relatedNotes.length" class="panel-list">
          <RouterLink v-for="note in relatedNotes" :key="note.id" :to="`/notes/${note.slug}`" class="panel-item">
            <span>{{ note.title }}</span>
            <span>{{ formatDate(note.published_at) }}</span>
          </RouterLink>
        </div>
        <EmptyState v-else title="暂无已公开笔记" description="相关记录仍在整理中。" />
      </div>
    </div>
  </div>

  <div class="container section" v-else-if="error">
    <EmptyState title="项目加载失败" :description="error" />
    <RouterLink to="/projects" class="back-link">返回项目索引</RouterLink>
  </div>

  <div class="container section" v-else>
    <EmptyState
      :title="notFound ? '项目不存在或未公开' : '项目暂时无法查看'"
      description="返回项目索引查看已公开条目。"
    />
    <RouterLink to="/projects" class="back-link">返回项目索引</RouterLink>
  </div>
</template>
