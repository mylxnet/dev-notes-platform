<script setup lang="ts">
import NoteCard from '../components/NoteCard.vue'
import { publishedNotes, publishedProjects } from '../data'

const recentNotes = publishedNotes().slice(0, 4)
const panelNotes = publishedNotes().slice(0, 3)
const indexProjects = publishedProjects().slice(0, 2)
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
          <RouterLink v-for="note in panelNotes" :key="note.id" :to="`/notes/${note.slug}`" class="panel-item">
            <span>{{ note.title }}</span>
            <span>{{ note.publishedAt.slice(5) }}</span>
          </RouterLink>
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
      <div class="note-list">
        <NoteCard v-for="note in recentNotes" :key="note.id" :note="note" />
      </div>
    </section>

    <section class="container section">
      <div class="section-head">
        <div>
          <h2 class="section-title">项目索引</h2>
          <div class="section-subtitle">个人练习作品，仅作记录与源码外链</div>
        </div>
        <RouterLink to="/projects" class="link-more">全部项目 →</RouterLink>
      </div>
      <div class="project-grid">
        <RouterLink v-for="project in indexProjects" :key="project.id" :to="`/projects/${project.slug}`" class="project-card">
          <div class="project-top">
            <span class="project-name">{{ project.name }}</span>
            <span class="status status-published">已发布</span>
          </div>
          <p class="project-summary">{{ project.summary }}</p>
          <div class="tech-list">
            <span v-for="tech in project.tech" :key="tech">{{ tech }}</span>
          </div>
          <div class="card-foot">
            <span>查看项目记录</span>
            <span>→</span>
          </div>
        </RouterLink>
      </div>
    </section>
  </div>
</template>
