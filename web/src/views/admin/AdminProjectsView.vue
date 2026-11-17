<script setup lang="ts">
import { onMounted, ref } from 'vue'
import EmptyState from '../../components/EmptyState.vue'
import { deleteProject, listProjects } from '../../admin'
import { errorMessageOf, statusText, type Project } from '../../api'

const projects = ref<Project[]>([])
const loading = ref(true)
const error = ref('')
const pendingDelete = ref<number | null>(null)
const deleting = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    projects.value = (await listProjects()).items
  } catch (err) {
    error.value = errorMessageOf(err)
    projects.value = []
  } finally {
    loading.value = false
  }
}

async function confirmDelete(id: number) {
  deleting.value = true
  try {
    await deleteProject(id)
    pendingDelete.value = null
    await load()
  } catch (err) {
    error.value = errorMessageOf(err)
  } finally {
    deleting.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="admin-heading">
      <div>
        <h1>项目管理</h1>
        <p>项目只提供源码外链，不托管安装包。删除项目不会删除关联笔记，笔记会变为「站点记录」。</p>
      </div>
      <RouterLink to="/admin/projects/new" class="btn btn-primary">新建项目</RouterLink>
    </div>

    <div class="callout">源码地址仅允许 GitHub 仓库地址；技术栈最多 10 项。</div>

    <div v-if="error" class="error-banner" role="alert">
      {{ error }}
      <button class="link-more" type="button" style="margin-left: 12px" @click="load">重试</button>
    </div>

    <div class="table-wrap" v-if="loading" aria-busy="true">
      <div style="padding: 22px">
        <div v-for="index in 4" :key="index" class="skeleton-line" :style="{ width: `${90 - index * 8}%` }" />
      </div>
    </div>
    <div class="table-wrap" v-else-if="projects.length">
      <table>
        <thead>
          <tr>
            <th>名称</th>
            <th>标识</th>
            <th>技术栈</th>
            <th>状态</th>
            <th>笔记</th>
            <th>排序</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="project in projects" :key="project.id">
            <td>{{ project.name }}</td>
            <td>{{ project.slug }}</td>
            <td>{{ project.tech_stack?.join(' / ') || '—' }}</td>
            <td><span class="status" :class="`status-${project.status}`">{{ statusText[project.status] }}</span></td>
            <td>{{ project.note_count }}</td>
            <td>{{ project.sort_order }}</td>
            <td>
              <div class="row-actions">
                <RouterLink class="icon-btn" :to="`/admin/projects/${project.id}/edit`">编辑</RouterLink>
                <template v-if="pendingDelete === project.id">
                  <button class="icon-btn" type="button" :disabled="deleting" @click="confirmDelete(project.id)">
                    {{ deleting ? '删除中…' : '确认删除' }}
                  </button>
                  <button class="icon-btn" type="button" :disabled="deleting" @click="pendingDelete = null">取消</button>
                </template>
                <button v-else class="icon-btn" type="button" @click="pendingDelete = project.id">删除</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <EmptyState v-else title="还没有项目" description="新建第一个项目后，可在笔记里关联它。" />
  </div>
</template>
