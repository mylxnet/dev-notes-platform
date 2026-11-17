<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { login } from '../../admin'
import { ApiError } from '../../api'

const router = useRouter()
const route = useRoute()

const username = ref('')
const password = ref('')
const error = ref('')
const submitting = ref(false)

async function submit() {
  error.value = ''
  if (!username.value.trim() || !password.value) {
    error.value = '请输入账号与密码。'
    return
  }

  submitting.value = true
  try {
    await login(username.value.trim(), password.value)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    router.replace(redirect.startsWith('/admin') ? redirect : '/admin/notes')
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : '登录失败，请稍后重试。'
    password.value = ''
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="login-shell">
    <section class="login-art">
      <div>
        <div class="eyebrow">SINGLE ADMIN</div>
        <h1>只属于一个人的<br />内容管理入口</h1>
        <p>没有注册、没有评论、没有投稿。这里只维护笔记与项目两类内容，保存前会拦截不符合备案边界的表述。</p>
      </div>
      <div>v1.3.4 by Mr lin</div>
    </section>
    <section class="login-panel">
      <div class="login-card">
        <h2>后台登录</h2>
        <p class="hint">单管理员账号，密码错误连续 5 次将锁定 15 分钟。</p>
        <form class="form-grid" novalidate @submit.prevent="submit">
          <label>
            账号
            <input v-model="username" type="text" autocomplete="username" placeholder="请输入管理员账号" :disabled="submitting" />
          </label>
          <label>
            密码
            <input
              v-model="password"
              type="password"
              autocomplete="current-password"
              placeholder="请输入密码"
              :disabled="submitting"
            />
          </label>
          <span class="field-error" aria-live="polite">{{ error }}</span>
          <button class="btn btn-primary" type="submit" :disabled="submitting">
            {{ submitting ? '登录中…' : '登录管理后台' }}
          </button>
        </form>
        <RouterLink to="/" class="back-link">返回站点首页</RouterLink>
      </div>
    </section>
  </div>
</template>
