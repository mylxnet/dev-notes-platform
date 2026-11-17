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
      <div class="header-actions">
        <RouterLink to="/admin/notes" class="btn btn-ghost">后台管理</RouterLink>
      </div>
    </div>
  </header>
</template>
