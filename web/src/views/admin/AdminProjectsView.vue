<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import EmptyState from '../../components/EmptyState.vue'
import Pager from '../../components/Pager.vue'
import { deleteProject, listProjects } from '../../admin'
import { errorMessageOf, statusText, type Project } from '../../api'

/** 与笔记列表、前台保持一致：每页 6 条，翻页由后端分页承担。 */
const PAGE_SIZE = 6

const projects = ref<Project[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(true)
const error = ref('')
const pendingDelete = ref<number | null>(null)
const deleting = ref(false)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await listProjects({ page: page.value, pageSize: PAGE_SIZE })
    projects.value = result.items
    total.value = result.total

    // 删除把总页数压小后，当前页可能越界，退到最后一页重新取。
    const lastPage = Math.max(1, Math.ceil(result.total / PAGE_SIZE))
    if (page.value > lastPage) {
      page.value = lastPage
      void load()
    }
  } catch (err) {
    error.value = errorMessageOf(err)
    projects.value = []
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

    <!-- 只有一页时没有分页条，用这行补上总数；多页时总数由分页条承担，避免重复。 -->
    <p v-if="!loading && total && totalPages === 1" class="admin-total">共 {{ total }} 个</p>

    <div v-if="!loading" class="admin-pager">
      <Pager
        :page="page"
        :total-pages="totalPages"
        :total="total"
        unit="个"
        label="后台项目分页"
        @change="goTo"
      />
    </div>
  </div>
</template>
