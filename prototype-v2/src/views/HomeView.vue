<script setup lang="ts">
import { publishedNotes, publishedProjects } from '../data'

const recentNotes = publishedNotes().slice(0, 4)
const panelNotes = publishedNotes().slice(0, 3)
const indexProjects = publishedProjects().slice(0, 2)
</script>

<template>
  <div>
    <section class="container masthead">
      <div class="masthead-top">
        <span><strong>Field Notes</strong> / 个人开发记录</span>
        <span>ISSUE 01 — 2026</span>
      </div>
      <div class="hero">
        <div>
          <div class="eyebrow">把写过的代码，整理成能回看的笔记</div>
          <h1 class="hero-title">开发<span class="thin">笔记<span class="accent-bar"></span></span></h1>
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
            <strong>最近更新 / Recent</strong>
            <RouterLink to="/notes">全部 →</RouterLink>
          </div>
          <div class="panel-list">
            <RouterLink v-for="(note, index) in panelNotes" :key="note.id" :to="`/notes/${note.slug}`" class="panel-item">
              <span class="idx">0{{ index + 1 }}</span>
              <span>{{ note.title }}</span>
              <span>{{ note.publishedAt.slice(5) }}</span>
            </RouterLink>
          </div>
        </aside>
      </div>
    </section>

    <section class="container section">
      <div class="section-head">
        <div>
          <div class="section-kicker">01 / Notes</div>
          <h2 class="section-title">最近笔记</h2>
          <div class="section-subtitle">按时间倒序，记录真实的排查过程与取舍</div>
        </div>
        <RouterLink to="/notes" class="link-more">全部笔记 →</RouterLink>
      </div>
      <div class="note-list">
        <RouterLink v-for="(note, index) in recentNotes" :key="note.id" :to="`/notes/${note.slug}`" class="note-row">
          <span class="idx">0{{ index + 1 }}</span>
          <div>
            <div class="note-title">{{ note.title }}</div>
            <p class="note-summary">{{ note.summary }}</p>
            <div class="note-meta">
              <span class="tag">{{ note.publishedAt }}</span>
            </div>
          </div>
          <span class="link-more">阅读全文 →</span>
        </RouterLink>
      </div>
    </section>

    <section class="container section">
      <div class="section-head">
        <div>
          <div class="section-kicker">02 / Projects</div>
          <h2 class="section-title">项目索引</h2>
          <div class="section-subtitle">个人练习作品，仅作记录与源码外链</div>
        </div>
        <RouterLink to="/projects" class="link-more">全部项目 →</RouterLink>
      </div>
      <div class="project-grid">
        <RouterLink v-for="(project, index) in indexProjects" :key="project.id" :to="`/projects/${project.slug}`" class="project-card">
          <div class="project-top">
            <span class="project-name"><span class="idx">0{{ index + 1 }}</span>{{ project.name }}</span>
            <span class="status status-published">已发布</span>
          </div>
          <p class="project-summary">{{ project.summary }}</p>
          <div class="tech-list">
            <span v-for="tech in project.tech" :key="tech">{{ tech }}</span>
          </div>
          <div class="card-foot">
            <span>查看项目记录</span>
            <span class="link-more">→</span>
          </div>
        </RouterLink>
      </div>
    </section>
  </div>
</template>
