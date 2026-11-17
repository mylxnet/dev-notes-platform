<script setup lang="ts">
import { useRouter } from 'vue-router'
import { logout } from '../../auth'

const router = useRouter()

function onLogout() {
  logout()
  router.push({ name: 'admin-login' })
}
</script>

<template>
  <div class="admin-shell">
    <aside class="admin-sidebar">
      <div class="admin-brand">
        <span class="brand-mark">N</span>
        <span>内容管理</span>
      </div>
      <nav class="admin-nav" aria-label="后台导航">
        <RouterLink to="/admin/notes">笔记管理</RouterLink>
        <RouterLink to="/admin/projects">项目管理</RouterLink>
        <RouterLink to="/admin/account">账号设置</RouterLink>
      </nav>
      <div class="admin-side-note">
        v1.0.2 by Mr lin<br />
        单管理员账号 · 不开放注册
      </div>
    </aside>
    <div class="admin-main">
      <header class="admin-topbar">
        <strong>我的开发笔记 · 管理后台</strong>
        <div class="admin-user">
          <span>管理员</span>
          <button class="btn btn-ghost" type="button" @click="onLogout">退出登录</button>
        </div>
      </header>
      <div class="admin-content">
        <RouterView />
      </div>
    </div>
  </div>
</template>
