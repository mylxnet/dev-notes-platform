<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchSettings, updateSettings } from '../../admin'
import { ApiError, errorMessageOf } from '../../api'
import { loadSiteInfo } from '../../site'

const icpNumber = ref('')
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const success = ref('')

/** 与服务端一致：可留空清空，填写时 5–100 位汉字/字母/数字/连字符。 */
const valid = computed(
  () => icpNumber.value === '' || /^[\u4e00-\u9fa5A-Za-z0-9-]{5,100}$/.test(icpNumber.value)
)

onMounted(async () => {
  try {
    const settings = await fetchSettings()
    icpNumber.value = settings.icp_number
  } catch (err) {
    error.value = errorMessageOf(err)
  } finally {
    loading.value = false
  }
})

async function submit() {
  error.value = ''
  success.value = ''
  if (!valid.value) {
    error.value = '备案号需为 5–100 个汉字、字母、数字或连字符，例如 甘ICP备12345678号-1；留空可恢复占位文案。'
    return
  }

  saving.value = true
  try {
    const saved = await updateSettings({ icp_number: icpNumber.value.trim() })
    icpNumber.value = saved.icp_number
    await loadSiteInfo(true)
    success.value = saved.icp_number ? '备案号已保存，前台页脚已更新。' : '已清空备案号，前台页脚恢复占位文案。'
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : '保存失败，请稍后重试。'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <div class="admin-heading">
      <div>
        <h1>站点设置</h1>
        <p>维护前台页脚展示的 ICP 备案号；备案通过后填写即可，留空时继续显示占位文案。</p>
      </div>
    </div>

    <div class="callout">备案号只在页脚展示，不做外部跳转；本站不开放注册、评论与投稿。</div>

    <div class="table-wrap" style="padding: 24px; max-width: 560px">
      <form class="form-grid" novalidate @submit.prevent="submit">
        <label>
          ICP 备案号
          <input
            v-model="icpNumber"
            type="text"
            placeholder="例如 甘ICP备12345678号-1"
            aria-label="ICP 备案号"
            :disabled="loading || saving"
          />
        </label>
        <span class="section-subtitle">可留空；填写时限 5–100 位汉字、字母、数字或连字符。</span>
        <div v-if="error" class="error-banner" role="alert" style="margin-bottom: 0">{{ error }}</div>
        <div v-if="success" class="callout" role="status">{{ success }}</div>
        <button class="btn btn-primary" type="submit" :disabled="loading || saving">
          {{ saving ? '保存中…' : '保存备案号' }}
        </button>
      </form>
    </div>
  </div>
</template>
