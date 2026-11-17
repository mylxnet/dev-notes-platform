<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import EmptyState from '../components/EmptyState.vue'
import { notes, projectBySlug, statusText } from '../data'

const route = useRoute()
const project = computed(() => projectBySlug(String(route.params.slug)))
const relatedNotes = computed(() =>
  project.value ? notes.filter((item) => item.projectId === project.value!.id && item.status === 'published') : []
)
</script>

<template>
  <div class="container section" v-if="project">
    <div class="section-head">
      <div>
        <div class="section-kicker">Project File / {{ project.slug }}</div>
        <h1 class="section-title">{{ project.name }}</h1>
        <div class="section-subtitle">{{ project.summary }}</div>
      </div>
      <span class="status" :class="`status-${project.status}`">{{ statusText[project.status] }}</span>
    </div>

    <div class="callout" v-if="project.status !== 'published'">
      该项目当前未公开，仅后台可见。正式站点中草稿与隐藏条目不会出现在前台。
    </div>

    <div class="about-grid">
      <div>
        <h2>项目说明</h2>
        <p>
          这是个人练习过程中的阶段性记录。本站只描述项目定位与技术取舍，源码入口外链到 GitHub，不提供任何安装包或可执行文件。
        </p>
        <div class="tech-list">
          <span v-for="tech in project.tech" :key="tech">{{ tech }}</span>
        </div>
        <p><a href="#" class="link-more" @click.prevent>{{ project.repoLabel }}（原型示意） →</a></p>
      </div>
      <div>
        <h2>关联笔记</h2>
        <div v-if="relatedNotes.length" class="panel-list">
          <RouterLink v-for="(note, index) in relatedNotes" :key="note.id" :to="`/notes/${note.slug}`" class="panel-item">
            <span class="idx">0{{ index + 1 }}</span>
            <span>{{ note.title }}</span>
            <span>{{ note.publishedAt.slice(5) }}</span>
          </RouterLink>
        </div>
        <EmptyState v-else title="暂无已公开笔记" description="相关记录仍在整理中。" />
      </div>
    </div>
  </div>
  <div class="container section" v-else>
    <EmptyState title="项目不存在" description="返回项目索引查看已公开条目。" />
    <RouterLink to="/projects" class="back-link">返回项目索引</RouterLink>
  </div>
</template>
