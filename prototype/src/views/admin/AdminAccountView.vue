<script setup lang="ts">
import { ref } from 'vue'
import { showToast } from '../../toast'

const username = ref('admin')
const newPassword = ref('')
const confirmPassword = ref('')
const error = ref('')

function submit() {
  if (username.value.trim().length < 3) {
    error.value = '账号至少 3 个字符。'
    return
  }
  if (newPassword.value.length < 8) {
    error.value = '新密码至少 8 位，请重新输入。'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致。'
    return
  }
  error.value = ''
  showToast('原型演示：密码修改未接入')
}
</script>

<template>
  <div>
    <div class="admin-heading">
      <div>
        <h1>账号设置</h1>
        <p>单管理员账号，无注册入口。修改密码后立即生效。</p>
      </div>
    </div>
    <div class="message-card">
      <form class="form-grid" novalidate @submit.prevent="submit">
        <label>
          管理员账号
          <input v-model="username" type="text" autocomplete="username" />
        </label>
        <label>
          新密码
          <input v-model="newPassword" type="password" autocomplete="new-password" placeholder="至少 8 位" />
        </label>
        <label>
          确认新密码
          <input v-model="confirmPassword" type="password" autocomplete="new-password" placeholder="再次输入新密码" />
        </label>
        <span class="field-error" aria-live="polite">{{ error }}</span>
        <div>
          <button class="btn btn-primary" type="submit">保存修改</button>
        </div>
      </form>
    </div>
  </div>
</template>
