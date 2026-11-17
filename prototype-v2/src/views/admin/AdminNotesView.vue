<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import EmptyState from '../../components/EmptyState.vue'
import { showToast } from '../../toast'
import { notes, projectById, statusText, type ContentStatus } from '../../data'

const loading = ref(true)
const statusFilter = ref<'all' | ContentStatus>('all')
const keyword = ref('')

onMounted(() => {
  window.setTimeout(() => {
    loading.value = false
  }, 360)
})

const rows = computed(() => {
  const text = keyword.value.trim().toLowerCase()
  return notes.filter((note) => {
    const matchStatus = statusFilter.value === 'all' || note.status === statusFilter.value
    const matchText = !text || note.title.toLowerCase().includes(text)
    return matchStatus && matchText
  })
})

const hasError = computed(() => keyword.value.trim().toLowerCase() === 'error')

function projectName(id: number) {
  return projectById(id)?.name ?? '站点记录'
}
</script>

<template>
  <div>
    <div class="admin-heading">
      <div>
        <div class="section-kicker">Backend / 01</div>
        <h1>笔记管理</h1>
        <p>维护笔记正文、关联项目与发布状态。草稿与隐藏条目不会出现在前台。</p>
      </div>
      <button class="btn btn-primary" type="button" @click="showToast('原型演示：新建笔记表单未接入')">新建笔记</button>
    </div>

    <div class="callout">保存前服务端会拦截经营性词汇与下载类表述，命中后拒绝保存并给出具体提示。</div>

    <div class="filters">
      <button class="chip" :class="{ active: statusFilter === 'all' }" type="button" @click="statusFilter = 'all'">全部</button>
      <button class="chip" :class="{ active: statusFilter === 'published' }" type="button" @click="statusFilter = 'published'">已发布</button>
      <button class="chip" :class="{ active: statusFilter === 'draft' }" type="button" @click="statusFilter = 'draft'">草稿</button>
      <button class="chip" :class="{ active: statusFilter === 'hidden' }" type="button" @click="statusFilter = 'hidden'">隐藏</button>
      <input v-model="keyword" class="search" type="search" placeholder="搜索标题" aria-label="搜索笔记标题" />
    </div>

    <div v-if="hasError" class="error-banner" role="alert">加载失败：原型演示错误状态，请修改关键词后重试。</div>

    <div class="table-wrap" v-if="loading" aria-label="加载中">
      <div style="padding: 22px">
        <div class="skeleton-line" v-for="n in 5" :key="n" :style="{ width: `${90 - n * 8}%` }"></div>
      </div>
    </div>
    <div class="table-wrap" v-else-if="rows.length">
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
          <tr v-for="note in rows" :key="note.id">
            <td>{{ note.title }}</td>
            <td>{{ projectName(note.projectId) }}</td>
            <td><span class="status" :class="`status-${note.status}`">{{ statusText[note.status] }}</span></td>
            <td>{{ note.publishedAt }}</td>
            <td>
              <div class="row-actions">
                <button class="icon-btn" type="button" @click="showToast('原型演示：编辑功能未接入')">编辑</button>
                <button class="icon-btn" type="button" @click="showToast('原型演示：删除功能未接入')">删除</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <EmptyState v-else title="没有匹配的笔记" description="调整状态筛选或搜索关键词。" />
  </div>
</template>
