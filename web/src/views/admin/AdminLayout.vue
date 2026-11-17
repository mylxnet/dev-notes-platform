<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchSession, logout, onUnauthorized, type AdminInfo } from '../../admin'

const router = useRouter()
const route = useRoute()

const admin = ref<AdminInfo | null>(null)
const ready = ref(false)
const loggingOut = ref(false)

function gotoLogin() {
  const redirect = route.fullPath
  router.replace({
    path: '/admin/login',
    query: redirect.startsWith('/admin') ? { redirect } : undefined
  })
}

onMounted(async () => {
  onUnauthorized(gotoLogin)
  try {
    admin.value = await fetchSession()
  } catch {
    gotoLogin()
    return
  }
  ready.value = true
})

async function onLogout() {
  loggingOut.value = true
  try {
    await logout()
  } catch {
    // 退出失败也回到登录页，避免停在半失效状态。
  } finally {
    loggingOut.value = false
  }
  router.replace('/admin/login')
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
        v1.2.0 by Mr lin<br />
        单管理员账号 · 不开放注册
      </div>
    </aside>
    <div class="admin-main">
      <header class="admin-topbar">
        <strong>我的开发笔记 · 管理后台</strong>
        <div class="admin-user">
          <span>{{ admin?.username ?? '管理员' }}</span>
          <button class="btn btn-ghost" type="button" :disabled="loggingOut" @click="onLogout">
            {{ loggingOut ? '退出中…' : '退出登录' }}
          </button>
        </div>
      </header>
      <div class="admin-content">
        <RouterView v-if="ready" />
        <div v-else class="skeleton-block" aria-busy="true">
          <div class="skeleton-line" style="width: 32%" />
          <div class="skeleton-line" style="width: 64%" />
          <div class="skeleton-line" style="width: 48%" />
        </div>
      </div>
    </div>
  </div>
</template>
