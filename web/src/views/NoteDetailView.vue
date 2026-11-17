<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import EmptyState from '../components/EmptyState.vue'
import { ApiError, errorMessageOf, fetchNote, formatDate, statusText, type Note } from '../api'

const route = useRoute()

const note = ref<Note | null>(null)
const loading = ref(true)
const error = ref('')
const notFound = ref(false)

/** 服务端已渲染并净化，前端只负责展示；旧数据没有 HTML 时回退到纯文本分段。 */
const paragraphs = computed(() => {
  const raw = note.value?.content ?? ''
  return raw
    .split(/\n{2,}/)
    .map((item) => item.trim())
    .filter(Boolean)
})

async function load(slug: string) {
  loading.value = true
  error.value = ''
  notFound.value = false
  note.value = null
  try {
    note.value = await fetchNote(slug)
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
  <div class="container" v-if="loading">
    <article class="article">
      <div class="article-body" aria-busy="true">
        <div class="skeleton-line" style="width: 30%" />
        <div class="skeleton-line" style="width: 66%; height: 26px" />
        <div class="skeleton-line" style="width: 88%" />
        <div class="skeleton-line" style="width: 92%" />
        <div class="skeleton-line" style="width: 74%" />
      </div>
    </article>
  </div>

  <div class="container" v-else-if="note">
    <article class="article">
      <div class="article-body">
        <div class="eyebrow">{{ formatDate(note.published_at) }} · 开发笔记</div>
        <h1>{{ note.title }}</h1>
        <p class="lead">{{ note.summary }}</p>
        <div v-if="note.content_html" class="article-markdown" v-html="note.content_html" />
        <template v-else>
          <p v-for="(paragraph, index) in paragraphs" :key="index" style="white-space: pre-wrap">{{ paragraph }}</p>
        </template>
      </div>
      <aside class="article-aside">
        <dl>
          <dt>关联项目</dt>
          <dd v-if="note.project">
            <RouterLink :to="`/projects/${note.project.slug}`" class="link-more">{{ note.project.name }}</RouterLink>
          </dd>
          <dd v-else>站点记录</dd>
          <dt>发布时间</dt>
          <dd>{{ formatDate(note.published_at) || '未设置' }}</dd>
          <dt>内容状态</dt>
          <dd>{{ statusText[note.status] }}</dd>
        </dl>
      </aside>
    </article>
  </div>

  <div class="container section" v-else-if="error">
    <EmptyState title="笔记加载失败" :description="error" />
    <RouterLink to="/notes" class="back-link">返回笔记列表</RouterLink>
  </div>

  <div class="container section" v-else>
    <EmptyState
      :title="notFound ? '笔记不存在或未公开' : '笔记暂时无法查看'"
      description="返回笔记列表查看已发布内容。"
    />
    <RouterLink to="/notes" class="back-link">返回笔记列表</RouterLink>
  </div>
</template>
