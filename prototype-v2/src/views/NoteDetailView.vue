<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import EmptyState from '../components/EmptyState.vue'
import { noteBySlug, projectById } from '../data'

const route = useRoute()
const note = computed(() => noteBySlug(String(route.params.slug)))
const project = computed(() => (note.value ? projectById(note.value.projectId) : undefined))
</script>

<template>
  <div class="container" v-if="note">
    <article class="article">
      <div class="article-body">
        <div class="eyebrow">{{ note.publishedAt }} · 开发笔记</div>
        <h1>{{ note.title }}</h1>
        <p>{{ note.summary }}</p>
        <p v-for="(paragraph, index) in note.content" :key="index">{{ paragraph }}</p>
      </div>
      <aside class="article-aside">
        <dl>
          <dt>关联项目</dt>
          <dd v-if="project"><RouterLink :to="`/projects/${project.slug}`" class="link-more">{{ project.name }}</RouterLink></dd>
          <dd v-else>站点记录</dd>
          <dt>发布时间</dt>
          <dd>{{ note.publishedAt }}</dd>
          <dt>内容状态</dt>
          <dd>{{ note.status === 'published' ? '已发布' : '未公开' }}</dd>
        </dl>
      </aside>
    </article>
  </div>
  <div class="container section" v-else>
    <EmptyState title="笔记不存在或未公开" description="返回笔记列表查看已发布内容。" />
    <RouterLink to="/notes" class="back-link">返回笔记列表</RouterLink>
  </div>
</template>
