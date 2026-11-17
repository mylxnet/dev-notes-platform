<script setup lang="ts">
// 后台管理入口在 Section 3 接入认证后恢复，避免当前出现死链。
</script>

<template>
  <header class="site-header">
    <div class="container header-inner">
      <RouterLink to="/" class="brand">
        <span class="brand-mark">N</span>
        <span>我的开发笔记</span>
      </RouterLink>
      <nav class="nav" aria-label="主导航">
        <RouterLink to="/">首页</RouterLink>
        <RouterLink to="/notes">笔记</RouterLink>
        <RouterLink to="/projects">项目</RouterLink>
        <RouterLink to="/about">关于</RouterLink>
      </nav>
    </div>
  </header>
</template>
