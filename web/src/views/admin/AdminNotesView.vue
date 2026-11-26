<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import EmptyState from '../../components/EmptyState.vue'
import Pager from '../../components/Pager.vue'
import { deleteNote, listNotes } from '../../admin'
import { errorMessageOf, formatDate, statusText, type ContentStatus, type Note } from '../../api'

/** 与前台笔记列表一致：每页 6 条，翻页由后端分页承担。 */
const PAGE_SIZE = 6

const notes = ref<Note[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(true)
const error = ref('')
const statusFilter = ref<'all' | ContentStatus>('all')
const keyword = ref('')
const pendingDelete = ref<number | null>(null)
const deleting = ref(false)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)))

let searchTimer: number | undefined

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await listNotes({
      status: statusFilter.value === 'all' ? undefined : statusFilter.value,
      keyword: keyword.value.trim() || undefined,
      page: page.value,
      pageSize: PAGE_SIZE
    })
    notes.value = result.items
    total.value = result.total

    // 筛选、搜索或删除把总页数压小后，当前页可能越界，退到最后一页重新取。
    const lastPage = Math.max(1, Math.ceil(result.total / PAGE_SIZE))
    if (page.value > lastPage) {
      page.value = lastPage
      void load()
    }
  } catch (err) {
    error.value = errorMessageOf(err)
    notes.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function goTo(target: number) {
  const clamped = Math.min(Math.max(1, target), totalPages.value)
  if (clamped === page.value) return
  page.value = clamped
  void load()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function setStatus(value: 'all' | ContentStatus) {
  if (statusFilter.value === value) return
  statusFilter.value = value
  page.value = 1
  void load()
}

watch(keyword, () => {
  window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(() => {
    page.value = 1
    void load()
  }, 300)
})

async function confirmDelete(id: number) {
  deleting.value = true
  try {
    await deleteNote(id)
    pendingDelete.value = null
    await load()
  } catch (err) {
    error.value = errorMessageOf(err)
  } finally {
    deleting.value = false
  }
}

onMounted(load)
onUnmounted(() => window.clearTimeout(searchTimer))
</script>

<template>
  <div>
    <div class="admin-heading">
      <div>
        <h1>笔记管理</h1>
        <p>维护笔记正文、关联项目与发布状态。草稿与隐藏条目不会出现在前台。</p>
      </div>
      <RouterLink to="/admin/notes/new" class="btn btn-primary">新建笔记</RouterLink>
    </div>

    <div class="callout">保存前服务端会拦截经营性词汇与下载类表述，命中后拒绝保存并给出具体提示。</div>

    <div class="filters">
      <button class="chip" :class="{ active: statusFilter === 'all' }" type="button" @click="setStatus('all')">全部</button>
      <button class="chip" :class="{ active: statusFilter === 'published' }" type="button" @click="setStatus('published')">
        已发布
      </button>
      <button class="chip" :class="{ active: statusFilter === 'draft' }" type="button" @click="setStatus('draft')">草稿</button>
      <button class="chip" :class="{ active: statusFilter === 'hidden' }" type="button" @click="setStatus('hidden')">隐藏</button>
      <input v-model="keyword" class="search" type="search" placeholder="搜索标题或摘要" aria-label="搜索笔记" />
    </div>

    <div v-if="error" class="error-banner" role="alert">
      {{ error }}
      <button class="link-more" type="button" style="margin-left: 12px" @click="load">重试</button>
    </div>

    <div class="table-wrap" v-if="loading" aria-busy="true">
      <div style="padding: 22px">
        <div v-for="index in 5" :key="index" class="skeleton-line" :style="{ width: `${92 - index * 9}%` }" />
      </div>
    </div>
    <div class="table-wrap" v-else-if="notes.length">
      <table>
        <thead>
          <tr>
            <th>标题</th>
            <th>关联项目</th>
            <th>状态</th>
            <th>发布时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="note in notes" :key="note.id">
            <td>{{ note.title }}</td>
            <td>{{ note.project?.name ?? '站点记录' }}</td>
            <td>
              <span class="status" :class="`status-${note.status}`">{{ statusText[note.status] }}</span>
              <span v-if="note.is_featured" class="badge-featured">推荐</span>
            </td>
            <td>{{ formatDate(note.published_at) || '—' }}</td>
            <td>
              <div class="row-actions">
                <RouterLink class="icon-btn" :to="`/admin/notes/${note.id}/edit`">编辑</RouterLink>
                <template v-if="pendingDelete === note.id">
                  <button class="icon-btn" type="button" :disabled="deleting" @click="confirmDelete(note.id)">
                    {{ deleting ? '删除中…' : '确认删除' }}
                  </button>
                  <button class="icon-btn" type="button" :disabled="deleting" @click="pendingDelete = null">取消</button>
                </template>
                <button v-else class="icon-btn" type="button" @click="pendingDelete = note.id">删除</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <EmptyState v-else title="没有匹配的笔记" description="调整状态筛选或搜索关键词，或新建一篇笔记。" />

    <!-- 只有一页时没有分页条，用这行补上总数；多页时总数由分页条承担，避免重复。 -->
    <p v-if="!loading && total && totalPages === 1" class="admin-total">共 {{ total }} 篇</p>

    <div v-if="!loading" class="admin-pager">
      <Pager
        :page="page"
        :total-pages="totalPages"
        :total="total"
        unit="篇"
        label="后台笔记分页"
        @change="goTo"
      />
    </div>
  </div>
</template>
