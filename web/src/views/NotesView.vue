<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import NoteCard from '../components/NoteCard.vue'
import EmptyState from '../components/EmptyState.vue'
import Pager from '../components/Pager.vue'
import { errorMessageOf, fetchNotes, fetchProjects, type Note, type Project } from '../api'

/** 每页固定 6 篇，翻页由后端分页承担。 */
const PAGE_SIZE = 6

const projects = ref<Project[]>([])
const notes = ref<Note[]>([])
const total = ref(0)
const page = ref(1)
const activeProject = ref<string | null>(null)
const keyword = ref('')
const loading = ref(true)
const error = ref('')

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)))

let activeController: AbortController | null = null
let searchTimer: number | undefined

async function load() {
  activeController?.abort()
  const controller = new AbortController()
  activeController = controller
  loading.value = true
  error.value = ''
  try {
    const result = await fetchNotes(
      {
        project: activeProject.value ?? undefined,
        keyword: keyword.value.trim() || undefined,
        page: page.value,
        pageSize: PAGE_SIZE
      },
      controller.signal
    )
    if (controller !== activeController) return
    notes.value = result.items
    total.value = result.total

    // 筛选或搜索把总页数压小后，当前页可能越界，退到最后一页重新取。
    const lastPage = Math.max(1, Math.ceil(result.total / PAGE_SIZE))
    if (page.value > lastPage) {
      page.value = lastPage
      void load()
    }
  } catch (err) {
    if (controller !== activeController) return
    if (err instanceof DOMException && err.name === 'AbortError') return
    error.value = errorMessageOf(err)
    notes.value = []
    total.value = 0
  } finally {
    if (controller === activeController) loading.value = false
  }
}

function goTo(target: number) {
  const clamped = Math.min(Math.max(1, target), totalPages.value)
  if (clamped === page.value) return
  page.value = clamped
  void load()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function loadProjects() {
  try {
    const result = await fetchProjects()
    projects.value = result.items
  } catch {
    // 筛选栏依赖项目列表，失败时退化为「全部」筛选，不阻塞笔记列表。
    projects.value = []
  }
}

function selectProject(slug: string | null) {
  if (activeProject.value === slug) return
  activeProject.value = slug
  page.value = 1
  load()
}

watch(keyword, () => {
  window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(() => {
    page.value = 1
    load()
  }, 300)
})

onMounted(() => {
  loadProjects()
  load()
})

onUnmounted(() => {
  window.clearTimeout(searchTimer)
  activeController?.abort()
})
</script>

<template>
  <div class="container section">
    <div class="section-head">
      <div>
        <h1 class="section-title">笔记列表</h1>
        <div class="section-subtitle">按日期倒序 · 支持按项目筛选</div>
      </div>
      <span class="section-subtitle">共 {{ total }} 篇</span>
    </div>

    <div class="filters">
      <button class="chip" :class="{ active: activeProject === null }" type="button" @click="selectProject(null)">全部</button>
      <button
        v-for="project in projects"
        :key="project.id"
        class="chip"
        :class="{ active: activeProject === project.slug }"
        type="button"
        @click="selectProject(project.slug)"
      >
        {{ project.name }}
      </button>
      <input v-model="keyword" class="search" type="search" placeholder="搜索笔记标题或摘要" aria-label="搜索笔记" />
    </div>

    <div v-if="error" class="error-banner" role="alert">
      {{ error }}
      <button class="link-more" type="button" style="margin-left: 12px" @click="load">重试</button>
    </div>

    <div v-if="loading" class="skeleton-block" aria-busy="true">
      <div v-for="index in 4" :key="index">
        <div class="skeleton-line" style="width: 42%" />
        <div class="skeleton-line" style="width: 76%" />
      </div>
    </div>
    <div v-else-if="notes.length" class="note-list">
      <NoteCard v-for="note in notes" :key="note.id" :note="note" />
    </div>
    <EmptyState v-else title="没有匹配的笔记" description="换个关键词，或选择「全部」项目筛选。" />

    <Pager
      v-if="!loading"
      :page="page"
      :total-pages="totalPages"
      :total="total"
      unit="篇"
      label="笔记分页"
      @change="goTo"
    />
  </div>
</template>
