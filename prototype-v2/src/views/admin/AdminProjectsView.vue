<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import EmptyState from '../../components/EmptyState.vue'
import { showToast } from '../../toast'
import { projects, statusText, type ContentStatus } from '../../data'

const loading = ref(true)
const statusFilter = ref<'all' | ContentStatus>('all')

onMounted(() => {
  window.setTimeout(() => {
    loading.value = false
  }, 360)
})

const rows = computed(() =>
  [...projects]
    .filter((item) => statusFilter.value === 'all' || item.status === statusFilter.value)
    .sort((a, b) => a.sortOrder - b.sortOrder)
)
</script>

<template>
  <div>
    <div class="admin-heading">
      <div>
        <div class="section-kicker">Backend / 02</div>
        <h1>项目管理</h1>
        <p>维护项目定位、技术栈、源码外链与排序。隐藏条目不会出现在前台。</p>
      </div>
      <button class="btn btn-primary" type="button" @click="showToast('原型演示：新建项目表单未接入')">新建项目</button>
    </div>

    <div class="filters">
      <button class="chip" :class="{ active: statusFilter === 'all' }" type="button" @click="statusFilter = 'all'">全部</button>
      <button class="chip" :class="{ active: statusFilter === 'published' }" type="button" @click="statusFilter = 'published'">已发布</button>
      <button class="chip" :class="{ active: statusFilter === 'draft' }" type="button" @click="statusFilter = 'draft'">草稿</button>
      <button class="chip" :class="{ active: statusFilter === 'hidden' }" type="button" @click="statusFilter = 'hidden'">隐藏</button>
    </div>

    <div class="table-wrap" v-if="loading" aria-label="加载中">
      <div style="padding: 22px">
        <div class="skeleton-line" v-for="n in 5" :key="n" :style="{ width: `${88 - n * 7}%` }"></div>
      </div>
    </div>
    <div class="table-wrap" v-else-if="rows.length">
      <table>
        <thead>
          <tr>
            <th>排序</th>
            <th>项目名称</th>
            <th>技术栈</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="project in rows" :key="project.id">
            <td>{{ project.sortOrder }}</td>
            <td>{{ project.name }}</td>
            <td>{{ project.tech.join(' · ') }}</td>
            <td><span class="status" :class="`status-${project.status}`">{{ statusText[project.status] }}</span></td>
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
    <EmptyState v-else title="没有匹配的项目" description="换一个状态筛选试试。" />
  </div>
</template>
