<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '../../auth'

const router = useRouter()
const username = ref('')
const password = ref('')
const error = ref('')

function submit() {
  if (username.value.trim().length < 3) {
    error.value = '账号至少 3 个字符，请检查后重试。'
    return
  }
  if (password.value.length < 4) {
    error.value = '密码至少 4 个字符，请检查后重试。'
    return
  }
  if (!login(username.value, password.value)) {
    error.value = '账号或密码校验未通过。'
    return
  }
  error.value = ''
  router.push({ name: 'admin-notes' })
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
      <div>v1.0.3 by Mr lin</div>
    </section>
    <section class="login-panel">
      <div class="login-card">
        <h2>后台登录</h2>
        <p class="hint">原型演示：任意账号（≥3 字符）与密码（≥4 字符）即可进入。</p>
        <form class="form-grid" novalidate @submit.prevent="submit">
          <label>
            账号
            <input v-model="username" type="text" autocomplete="username" placeholder="请输入管理员账号" />
          </label>
          <label>
            密码
            <input v-model="password" type="password" autocomplete="current-password" placeholder="请输入密码" />
          </label>
          <span class="field-error" aria-live="polite">{{ error }}</span>
          <button class="btn btn-primary" type="submit">登录管理后台</button>
        </form>
        <RouterLink to="/" class="back-link">返回站点首页</RouterLink>
      </div>
    </section>
  </div>
</template>
