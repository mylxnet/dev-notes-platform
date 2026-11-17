<script setup lang="ts">
import { computed, ref } from 'vue'
import { changePassword } from '../../admin'
import { ApiError } from '../../api'

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const error = ref('')
const success = ref('')
const submitting = ref(false)

const newPasswordValid = computed(
  () => newPassword.value.length >= 12 && /[A-Za-z]/.test(newPassword.value) && /[0-9]/.test(newPassword.value)
)

const passwordHint = computed(() => {
  if (!newPassword.value) return '至少 12 位，且同时包含字母和数字'
  return newPasswordValid.value ? '格式符合要求' : '至少 12 位，且同时包含字母和数字'
})

async function submit() {
  error.value = ''
  success.value = ''

  if (!currentPassword.value) {
    error.value = '请输入当前密码。'
    return
  }
  if (!newPasswordValid.value) {
    error.value = '新密码至少 12 位，且同时包含字母和数字。'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    error.value = '两次输入的新密码不一致。'
    return
  }

  submitting.value = true
  try {
    await changePassword(currentPassword.value, newPassword.value)
    success.value = '密码已更新，当前会话已自动续期。'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (err) {
    if (err instanceof ApiError) {
      error.value = err.hits?.length ? `${err.message}（${err.hits.join('、')}）` : err.message
    } else {
      error.value = '保存失败，请稍后重试。'
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div>
    <div class="admin-heading">
      <div>
        <h1>账号设置</h1>
        <p>修改后旧会话立即失效，当前页面会自动续期，不需要重新登录。</p>
      </div>
    </div>

    <div class="table-wrap" style="padding: 24px; max-width: 520px">
      <form class="form-grid" novalidate @submit.prevent="submit">
        <label>
          当前密码
          <input
            v-model="currentPassword"
            type="password"
            autocomplete="current-password"
            placeholder="请输入当前密码"
            :disabled="submitting"
          />
        </label>
        <label>
          新密码
          <input
            v-model="newPassword"
            type="password"
            autocomplete="new-password"
            placeholder="至少 12 位，含字母和数字"
            :disabled="submitting"
          />
        </label>
        <span class="section-subtitle" :style="{ color: newPasswordValid ? 'var(--green)' : undefined }">
          {{ passwordHint }}
        </span>
        <label>
          确认新密码
          <input
            v-model="confirmPassword"
            type="password"
            autocomplete="new-password"
            placeholder="再次输入新密码"
            :disabled="submitting"
          />
        </label>
        <div v-if="error" class="error-banner" role="alert" style="margin-bottom: 0">{{ error }}</div>
        <div v-if="success" class="callout" role="status">{{ success }}</div>
        <button class="btn btn-primary" type="submit" :disabled="submitting">
          {{ submitting ? '保存中…' : '保存新密码' }}
        </button>
      </form>
    </div>
  </div>
</template>
