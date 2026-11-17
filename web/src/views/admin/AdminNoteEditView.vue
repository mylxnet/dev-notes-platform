<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { createNote, getNote, listProjects, previewMarkdown, updateNote, type NoteInput } from '../../admin'
import {
  ApiError,
  errorMessageOf,
  statusText,
  type ContentStatus,
  type Project
} from '../../api'

const route = useRoute()
const router = useRouter()

const noteId = computed(() => Number(route.params.id ?? 0))
const isEdit = computed(() => noteId.value > 0)

const projects = ref<Project[]>([])
const loading = ref(isEdit.value)
const saving = ref(false)
const loaded = ref(!isEdit.value)
const dirty = ref(false)

const error = ref('')
const success = ref('')
const fieldError = ref<{ field?: string; message: string; hits?: string[]; suggestion?: string } | null>(null)

const form = ref({
  title: '',
  slug: '',
  summary: '',
  content: '',
  project_id: null as number | null,
  status: 'draft' as ContentStatus,
  published_at: ''
})

const showPreview = ref(false)
const previewHtml = ref('')
const previewLoading = ref(false)

const statuses: ContentStatus[] = ['published', 'draft', 'hidden']

function toInput(): NoteInput {
  return {
    title: form.value.title.trim(),
    slug: form.value.slug.trim(),
    summary: form.value.summary.trim(),
    content: form.value.content,
    project_id: form.value.project_id,
    status: form.value.status,
    published_at: form.value.published_at ? `${form.value.published_at}T00:00:00Z` : ''
  }
}

function validateLocally(): boolean {
  if (!form.value.title.trim() || form.value.title.trim().length > 120) {
    fieldError.value = { field: 'title', message: '标题需为 1–120 个字符' }
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
  if (!form.value.content.trim()) {
    fieldError.value = { field: 'content', message: '正文不能为空且不超过 100 KB' }
    return false
  }
  if (form.value.status === 'published' && !form.value.published_at) {
    form.value.published_at = new Date().toISOString().slice(0, 10)
  }
  return true
}

async function save() {
  error.value = ''
  success.value = ''
  fieldError.value = null
  if (!validateLocally()) return

  saving.value = true
  try {
    const input = toInput()
    if (isEdit.value) {
      await updateNote(noteId.value, input)
    } else {
      await createNote(input)
    }
    dirty.value = false
    success.value = '已保存'
    router.push('/admin/notes')
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

async function togglePreview() {
  if (showPreview.value) {
    showPreview.value = false
    return
  }
  previewLoading.value = true
  error.value = ''
  try {
    previewHtml.value = await previewMarkdown(form.value.content)
    showPreview.value = true
  } catch (err) {
    error.value = errorMessageOf(err)
  } finally {
    previewLoading.value = false
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
  try {
    projects.value = (await listProjects()).items
  } catch {
    // 项目列表失败不阻塞编辑，关联项目可留空。
  }

  if (!isEdit.value) return

  try {
    const note = await getNote(noteId.value)
    form.value = {
      title: note.title,
      slug: note.slug,
      summary: note.summary,
      content: note.content ?? '',
      project_id: note.project_id ?? null,
      status: note.status,
      published_at: note.published_at ? note.published_at.slice(0, 10) : ''
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
        <h1>{{ isEdit ? '编辑笔记' : '新建笔记' }}</h1>
        <p>正文使用 Markdown。保存前会做一次服务端校验，命中备案边界词会拒绝保存。</p>
      </div>
      <div class="row-actions">
        <RouterLink to="/admin/notes" class="btn">返回列表</RouterLink>
        <button class="btn" type="button" :disabled="previewLoading || saving" @click="togglePreview">
          {{ previewLoading ? '渲染中…' : showPreview ? '关闭预览' : '预览' }}
        </button>
        <button class="btn btn-primary" type="button" :disabled="saving || loading" @click="save">
          {{ saving ? '保存中…' : '保存' }}
        </button>
      </div>
    </div>

    <div v-if="error" class="error-banner" role="alert">{{ error }}</div>
    <div v-if="success" class="callout" role="status">{{ success }}</div>
    <div v-if="fieldError" class="error-banner" role="alert">
      <strong>{{ fieldError.message }}</strong>
      <div v-if="fieldError.hits?.length" style="margin-top: 6px">命中表述：{{ fieldError.hits.join('、') }}</div>
      <div v-if="fieldError.suggestion" style="margin-top: 6px">修正建议：{{ fieldError.suggestion }}</div>
    </div>

    <div v-if="loading" class="skeleton-block" aria-busy="true">
      <div class="skeleton-line" style="width: 40%" />
      <div class="skeleton-line" style="width: 72%" />
      <div class="skeleton-line" style="width: 88%" />
    </div>

    <form v-else class="form-grid form-grid-wide" novalidate @submit.prevent="save">
      <label>
        标题
        <input v-model="form.title" type="text" maxlength="120" placeholder="1–120 个字符" />
      </label>
      <label>
        链接标识（slug）
        <input v-model="form.slug" type="text" placeholder="小写字母、数字与连字符，如 sqlite-pure-go-driver" />
      </label>
      <label>
        摘要
        <textarea v-model="form.summary" rows="2" maxlength="300" placeholder="1–300 个字符，用于列表展示" />
      </label>

      <div class="editor-row">
        <label>
          关联项目
          <select v-model="form.project_id">
            <option :value="null">站点记录（不关联项目）</option>
            <option v-for="project in projects" :key="project.id" :value="project.id">{{ project.name }}</option>
          </select>
        </label>
        <label>
          状态
          <select v-model="form.status">
            <option v-for="value in statuses" :key="value" :value="value">{{ statusText[value] }}</option>
          </select>
        </label>
        <label>
          发布时间
          <input v-model="form.published_at" type="date" :disabled="form.status !== 'published'" />
        </label>
      </div>

      <label>
        正文（Markdown）
        <textarea v-model="form.content" class="editor-textarea" placeholder="支持标题、列表、代码块、表格；不允许原始 HTML 与外链图片" />
      </label>

      <div v-if="showPreview">
        <div class="section-subtitle">预览（服务端渲染并净化）</div>
        <div class="article-markdown preview-panel" v-html="previewHtml" />
      </div>

      <div class="row-actions">
        <button class="btn btn-primary" type="submit" :disabled="saving">
          {{ saving ? '保存中…' : '保存' }}
        </button>
        <RouterLink to="/admin/notes" class="btn">取消</RouterLink>
      </div>
    </form>
  </div>
</template>
