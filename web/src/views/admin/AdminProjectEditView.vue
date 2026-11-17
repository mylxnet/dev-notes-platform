<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { createProject, listProjects, updateProject, type ProjectInput } from '../../admin'
import { ApiError, errorMessageOf, statusText, type ContentStatus } from '../../api'

const route = useRoute()
const router = useRouter()

const projectId = computed(() => Number(route.params.id ?? 0))
const isEdit = computed(() => projectId.value > 0)

const loading = ref(isEdit.value)
const saving = ref(false)
const loaded = ref(!isEdit.value)
const dirty = ref(false)

const error = ref('')
const fieldError = ref<{ field?: string; message: string; hits?: string[]; suggestion?: string } | null>(null)

const form = ref({
  name: '',
  slug: '',
  summary: '',
  techText: '',
  repo_url: '',
  sort_order: 0,
  status: 'draft' as ContentStatus
})

const statuses: ContentStatus[] = ['published', 'draft', 'hidden']

function toInput(): ProjectInput {
  return {
    name: form.value.name.trim(),
    slug: form.value.slug.trim(),
    summary: form.value.summary.trim(),
    tech_stack: form.value.techText
      .split(/[,，\n]/)
      .map((item) => item.trim())
      .filter(Boolean),
    repo_url: form.value.repo_url.trim(),
    sort_order: Number(form.value.sort_order) || 0,
    status: form.value.status
  }
}

function validateLocally(): boolean {
  if (!form.value.name.trim() || form.value.name.trim().length > 120) {
    fieldError.value = { field: 'name', message: '项目名称需为 1–120 个字符' }
    return false
  }
  if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(form.value.slug.trim())) {
    fieldError.value = { field: 'slug', message: '链接标识只能使用小写字母、数字和连字符，长度 2–80' }
    return false
  }
  if (!form.value.summary.trim() || form.value.summary.trim().length > 300) {
    fieldError.value = { field: 'summary', message: '摘要需为 1–300 个字符' }
    return false
  }
  const stack = toInput().tech_stack
  if (stack.length > 10 || stack.some((item) => item.length > 30)) {
    fieldError.value = { field: 'tech_stack', message: '技术栈最多 10 项，单项不超过 30 个字符' }
    return false
  }
  if (form.value.repo_url.trim() && !/^https:\/\/github\.com\/[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+(?:\.git)?\/?$/.test(form.value.repo_url.trim())) {
    fieldError.value = { field: 'repo_url', message: '源码地址仅允许 GitHub 仓库地址（https://github.com/owner/repo）' }
    return false
  }
  return true
}

async function save() {
  error.value = ''
  fieldError.value = null
  if (!validateLocally()) return

  saving.value = true
  try {
    const input = toInput()
    if (isEdit.value) {
      await updateProject(projectId.value, input)
    } else {
      await createProject(input)
    }
    dirty.value = false
    router.push('/admin/projects')
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 422) {
        fieldError.value = { field: err.field, message: err.message, hits: err.hits, suggestion: err.suggestion }
      } else {
        error.value = err.message
      }
    } else {
      error.value = '保存失败，请稍后重试。'
    }
  } finally {
    saving.value = false
  }
}

watch(
  form,
  () => {
    if (loaded.value) dirty.value = true
  },
  { deep: true }
)

onMounted(async () => {
  if (!isEdit.value) return
  try {
    const result = await listProjects()
    const project = result.items.find((item) => item.id === projectId.value)
    if (!project) {
      error.value = '项目不存在'
      return
    }
    form.value = {
      name: project.name,
      slug: project.slug,
      summary: project.summary,
      techText: (project.tech_stack ?? []).join('、'),
      repo_url: project.repo_url,
      sort_order: project.sort_order,
      status: project.status
    }
    loaded.value = true
    dirty.value = false
  } catch (err) {
    error.value = errorMessageOf(err)
  } finally {
    loading.value = false
  }
})

onBeforeRouteLeave(() => {
  if (!dirty.value || saving.value) return true
  return window.confirm('有未保存的修改，确定离开吗？')
})
</script>

<template>
  <div>
    <div class="admin-heading">
      <div>
        <h1>{{ isEdit ? '编辑项目' : '新建项目' }}</h1>
        <p>项目信息用于前台索引展示，源码地址只允许 GitHub 仓库。</p>
      </div>
      <div class="row-actions">
        <RouterLink to="/admin/projects" class="btn">返回列表</RouterLink>
        <button class="btn btn-primary" type="button" :disabled="saving || loading" @click="save">
          {{ saving ? '保存中…' : '保存' }}
        </button>
      </div>
    </div>

    <div v-if="error" class="error-banner" role="alert">{{ error }}</div>
    <div v-if="fieldError" class="error-banner" role="alert">
      <strong>{{ fieldError.message }}</strong>
      <div v-if="fieldError.hits?.length" style="margin-top: 6px">命中表述：{{ fieldError.hits.join('、') }}</div>
      <div v-if="fieldError.suggestion" style="margin-top: 6px">修正建议：{{ fieldError.suggestion }}</div>
    </div>

    <div v-if="loading" class="skeleton-block" aria-busy="true">
      <div class="skeleton-line" style="width: 36%" />
      <div class="skeleton-line" style="width: 68%" />
    </div>

    <form v-else class="form-grid form-grid-wide" novalidate @submit.prevent="save">
      <label>
        项目名称
        <input v-model="form.name" type="text" maxlength="120" placeholder="1–120 个字符" />
      </label>
      <label>
        链接标识（slug）
        <input v-model="form.slug" type="text" placeholder="小写字母、数字与连字符，如 superclip" />
      </label>
      <label>
        摘要
        <textarea v-model="form.summary" rows="2" maxlength="300" placeholder="1–300 个字符，用于卡片展示" />
      </label>
      <label>
        技术栈
        <input v-model="form.techText" type="text" placeholder="用逗号或顿号分隔，最多 10 项，如 Go、Vue、SQLite" />
      </label>
      <label>
        源码仓库
        <input v-model="form.repo_url" type="url" placeholder="https://github.com/owner/repo" />
      </label>

      <div class="editor-row">
        <label>
          状态
          <select v-model="form.status">
            <option v-for="value in statuses" :key="value" :value="value">{{ statusText[value] }}</option>
          </select>
        </label>
        <label>
          排序
          <input v-model.number="form.sort_order" type="number" min="0" max="9999" />
        </label>
      </div>

      <div class="row-actions">
        <button class="btn btn-primary" type="submit" :disabled="saving">
          {{ saving ? '保存中…' : '保存' }}
        </button>
        <RouterLink to="/admin/projects" class="btn">取消</RouterLink>
      </div>
    </form>
  </div>
</template>
