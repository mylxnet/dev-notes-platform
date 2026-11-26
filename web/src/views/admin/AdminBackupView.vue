<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  downloadBackupPackage,
  fetchBackupSummary,
  inspectBackupPackage,
  listBackupSnapshots,
  restoreBackupPackage,
  restoreBackupSnapshot,
  type BackupInspection,
  type BackupSnapshot,
  type BackupSummary,
  type RestoreOutcome
} from '../../admin'
import { ApiError, errorMessageOf } from '../../api'

const summary = ref<BackupSummary | null>(null)
const snapshots = ref<BackupSnapshot[]>([])
const loading = ref(true)
const actionError = ref('')
const notice = ref('')

const backingUp = ref(false)
const lastBackup = ref<{ name: string; bytes: number } | null>(null)
const rollingBack = ref('')

const picked = ref<File | null>(null)
const inspecting = ref(false)
const inspection = ref<BackupInspection | null>(null)
const confirmText = ref('')
const restoring = ref(false)
const restoreResult = ref<RestoreOutcome | null>(null)
const restoreError = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const dragging = ref(false)

const canRestore = computed(
  () =>
    !!picked.value &&
    !!inspection.value?.restorable &&
    confirmText.value.trim().toUpperCase() === 'RESTORE' &&
    !restoring.value &&
    !inspecting.value
)

const busy = computed(() => backingUp.value || restoring.value || inspecting.value || rollingBack.value !== '')

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = bytes
  let index = 0
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024
    index += 1
  }
  return `${index === 0 || value >= 10 ? Math.round(value) : value.toFixed(1)} ${units[index]}`
}

function formatDateTime(value?: string | null): string {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (item: number) => String(item).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(
    date.getMinutes()
  )}`
}

function messageOf(error: unknown, fallback: string): string {
  if (error instanceof ApiError) return error.message
  return fallback
}

async function loadSummary() {
  try {
    summary.value = await fetchBackupSummary()
    snapshots.value = (await listBackupSnapshots()).items
  } catch (error) {
    actionError.value = errorMessageOf(error)
  } finally {
    loading.value = false
  }
}

onMounted(loadSummary)

function triggerDownload(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

async function onBackup() {
  if (busy.value) return
  actionError.value = ''
  notice.value = ''
  backingUp.value = true
  try {
    const { blob, filename } = await downloadBackupPackage()
    triggerDownload(blob, filename)
    lastBackup.value = { name: filename, bytes: blob.size }
    notice.value = `已生成 ${filename}（${formatBytes(blob.size)}），浏览器正在下载。`
    await loadSummary()
  } catch (error) {
    actionError.value = messageOf(error, '生成备份失败，请稍后重试')
  } finally {
    backingUp.value = false
  }
}

async function pickFile(file: File) {
  restoreResult.value = null
  restoreError.value = ''
  confirmText.value = ''
  inspection.value = null
  picked.value = file

  if (!file.name.toLowerCase().endsWith('.tar.gz')) {
    restoreError.value = '只接受 .tar.gz 备份包，请重新选择。'
    return
  }
  if (summary.value && file.size > summary.value.max_archive_bytes) {
    restoreError.value = `备份包不能超过 ${formatBytes(summary.value.max_archive_bytes)}。`
    return
  }

  inspecting.value = true
  try {
    inspection.value = await inspectBackupPackage(file)
    if (!inspection.value.restorable) {
      restoreError.value = '预检未通过，这份备份不能恢复到当前版本。'
    }
  } catch (error) {
    restoreError.value = messageOf(error, '预检失败，请确认文件是否完整')
  } finally {
    inspecting.value = false
  }
}

function onFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  target.value = ''
  if (file) void pickFile(file)
}

function onDrop(event: DragEvent) {
  event.preventDefault()
  dragging.value = false
  if (busy.value) return
  const file = event.dataTransfer?.files?.[0]
  if (file) void pickFile(file)
}

function resetPicked() {
  picked.value = null
  inspection.value = null
  restoreError.value = ''
  confirmText.value = ''
}

async function onRestore() {
  if (!picked.value || !canRestore.value) return
  restoreError.value = ''
  notice.value = ''
  restoring.value = true
  try {
    restoreResult.value = await restoreBackupPackage(picked.value, confirmText.value.trim())
    notice.value = '恢复完成：数据已替换为备份包内容。'
    await loadSummary()
  } catch (error) {
    restoreError.value = messageOf(error, '恢复失败，数据应已回滚到恢复前状态')
  } finally {
    restoring.value = false
  }
}

async function onRollback(snapshot: BackupSnapshot) {
  if (busy.value) return
  const agreed = window.confirm(
    `确定回滚到 ${snapshot.name}？\n回滚会覆盖当前内容，执行前会先把当前数据另存一份快照。`
  )
  if (!agreed) return

  actionError.value = ''
  notice.value = ''
  rollingBack.value = snapshot.name
  try {
    restoreResult.value = await restoreBackupSnapshot(snapshot.name)
    notice.value = `已回滚到 ${snapshot.name}。`
    await loadSummary()
  } catch (error) {
    actionError.value = messageOf(error, '回滚失败，请稍后重试')
  } finally {
    rollingBack.value = ''
  }
}
</script>

<template>
  <div>
    <div class="admin-heading">
      <div>
        <h1>数据备份</h1>
        <p>把数据库和图片打成一份自包含的包，随时恢复到本机或另一台服务器。</p>
      </div>
      <button class="btn btn-primary" type="button" :disabled="loading || busy" @click="onBackup">
        {{ backingUp ? '正在生成…' : '生成备份包' }}
      </button>
    </div>

    <div class="callout">
      备份包固定包含 <code>notes.db</code>（SQLite 一致性快照，含 WAL 未落盘数据）、<code>media/</code>（后台上传的图片）与
      <code>manifest.json</code>（版本与逐文件 sha256）。程序内置插图随版本发布，不需要备份。
    </div>

    <div class="backup-stats">
      <div class="backup-stat">
        <span class="backup-stat-k">数据库</span>
        <strong class="backup-stat-v">{{ loading ? '—' : formatBytes(summary?.db_bytes ?? 0) }}</strong>
        <span class="backup-stat-n">
          {{ summary?.notes ?? 0 }} 篇笔记 · {{ summary?.projects ?? 0 }} 个项目 · schema v{{ summary?.schema_version ?? '—' }}
        </span>
      </div>
      <div class="backup-stat">
        <span class="backup-stat-k">上传图片</span>
        <strong class="backup-stat-v">{{ summary?.media_files ?? 0 }} 张</strong>
        <span class="backup-stat-n">{{ formatBytes(summary?.media_bytes ?? 0) }} · 位于 data/media/</span>
      </div>
      <div class="backup-stat">
        <span class="backup-stat-k">恢复前快照</span>
        <strong class="backup-stat-v">{{ snapshots.length }} 份</strong>
        <span class="backup-stat-n">服务器 data/backups/，可一键回滚</span>
      </div>
    </div>

    <div v-if="actionError" class="error-banner" role="alert">{{ actionError }}</div>
    <div v-if="notice" class="callout" role="status">{{ notice }}</div>

    <section class="backup-panel">
      <div class="panel-head">
        <h2>立即备份</h2>
        <span class="tag">手动下载 · 不占服务器空间</span>
      </div>
      <p class="section-subtitle">
        生成后直接以 <code>.tar.gz</code> 下载到你本机；服务器只保留恢复前的自动快照，下载完临时文件即删。
      </p>

      <div v-if="backingUp" class="backup-progress" role="status">
        <div class="backup-progress-line"><span>正在打包数据库与图片…</span></div>
        <div class="backup-bar"><i /></div>
      </div>

      <div v-if="lastBackup" class="backup-result">
        <strong>已生成：</strong>
        <code>{{ lastBackup.name }}</code>
        <span class="muted"> · {{ formatBytes(lastBackup.bytes) }}</span>
      </div>
    </section>

    <section class="backup-panel">
      <div class="panel-head">
        <h2>恢复前快照</h2>
        <span class="tag">服务器上仅保留这一类备份</span>
      </div>
      <p class="section-subtitle">
        每次执行恢复前，系统会先把当时的数据打成快照。恢复后发现内容不对，可在这里一键回滚。
      </p>

      <div v-if="snapshots.length === 0" class="empty">还没有快照。执行过一次恢复后，这里会出现可回滚的记录。</div>
      <div v-else class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>文件</th>
              <th>大小</th>
              <th>时间</th>
              <th style="text-align: right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="snapshot in snapshots" :key="snapshot.name">
              <td><code>{{ snapshot.name }}</code></td>
              <td>{{ formatBytes(snapshot.bytes) }}</td>
              <td class="muted">{{ formatDateTime(snapshot.mod_time) }}</td>
              <td style="text-align: right">
                <button
                  class="btn btn-sm btn-ghost"
                  type="button"
                  :disabled="busy"
                  @click="onRollback(snapshot)"
                >
                  {{ rollingBack === snapshot.name ? '回滚中…' : '回滚' }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="backup-panel">
      <div class="panel-head">
        <h2>从备份恢复</h2>
        <span class="tag tag-warn">会覆盖当前全部内容</span>
      </div>
      <p class="section-subtitle">
        先选包预检，确认内容与版本无误后再输入 <code>RESTORE</code> 执行。恢复前会自动把当前数据存成快照。
      </p>

      <div
        class="backup-drop"
        :class="{ 'backup-drop-active': dragging, 'backup-drop-disabled': busy }"
        role="button"
        tabindex="0"
        @click="!busy && fileInput?.click()"
        @keydown.enter.prevent="!busy && fileInput?.click()"
        @dragover.prevent="dragging = !busy"
        @dragleave.prevent="dragging = false"
        @drop="onDrop"
      >
        <strong>{{ inspecting ? '正在预检…' : '拖入备份包，或点击选择文件' }}</strong>
        <span>只接受 .tar.gz，单个文件不超过 {{ formatBytes(summary?.max_archive_bytes ?? 0) }}</span>
        <input ref="fileInput" class="backup-file" type="file" accept=".gz,.tar.gz" @change="onFileChange" />
      </div>

      <div v-if="picked" class="backup-picked">
        <div class="backup-picked-head">
          <code>{{ picked.name }}</code>
          <span class="muted"> · {{ formatBytes(picked.size) }}</span>
          <button class="btn btn-sm btn-ghost" type="button" :disabled="busy" @click="resetPicked">取消</button>
        </div>

        <div v-if="inspecting" class="backup-progress" role="status">
          <div class="backup-progress-line"><span>正在解包并校验 sha256…</span></div>
          <div class="backup-bar"><i /></div>
        </div>

        <div v-if="inspection">
          <div class="backup-picked-head">
            <span class="tag" :class="inspection.restorable ? 'tag-ok' : 'tag-warn'">
              {{ inspection.restorable ? '预检通过，可恢复' : '预检未通过' }}
            </span>
          </div>
          <ul class="backup-checks">
            <li>格式：{{ inspection.manifest.format }}（程序 {{ inspection.manifest.app_version }}，schema v{{ inspection.manifest.schema_version }}）</li>
            <li>创建时间：{{ formatDateTime(inspection.manifest.created_at) }}</li>
            <li>
              包内内容：{{ inspection.manifest.counts.notes }} 篇笔记 · {{ inspection.manifest.counts.projects }} 个项目 ·
              {{ inspection.manifest.counts.media_files }} 张图片
            </li>
            <li v-for="problem in inspection.problems" :key="problem" class="backup-check-bad">问题：{{ problem }}</li>
            <li v-for="warning in inspection.warnings" :key="warning" class="backup-check-warn">提醒：{{ warning }}</li>
            <li v-for="missing in inspection.missing" :key="missing" class="backup-check-warn">
              数据库引用了这张图片但包里没有：{{ missing }}
            </li>
          </ul>

          <div class="backup-confirm">
            <label>
              确认覆盖当前数据，请输入 <code>RESTORE</code>
              <input
                v-model="confirmText"
                type="text"
                placeholder="RESTORE"
                autocomplete="off"
                :disabled="!inspection.restorable || restoring"
              />
            </label>
            <div class="backup-confirm-actions">
              <button class="btn btn-primary" type="button" :disabled="!canRestore" @click="onRestore">
                {{ restoring ? '正在恢复…' : '执行恢复' }}
              </button>
              <span class="section-subtitle">将先把当前数据另存为 data/backups/pre-restore-*.tar.gz</span>
            </div>
          </div>

          <div v-if="restoring" class="backup-progress" role="status">
            <div class="backup-progress-line"><span>正在替换数据库与图片，请勿关闭页面…</span></div>
            <div class="backup-bar"><i /></div>
          </div>
        </div>

        <div v-if="restoreError" class="error-banner" role="alert">{{ restoreError }}</div>
      </div>

      <div v-if="restoreResult" class="backup-result">
        <strong>恢复完成 · 自检：</strong>
        {{ restoreResult.notes }} 篇笔记 · {{ restoreResult.projects }} 个项目 · {{ restoreResult.media_files }} 张图片 ·
        schema v{{ restoreResult.schema_version }}
        <div class="section-subtitle">
          恢复前快照：<code>{{ restoreResult.snapshot }}</code>
          <template v-if="restoreResult.missing.length">；缺失图片 {{ restoreResult.missing.length }} 张</template>
        </div>
      </div>
    </section>

    <div class="callout">
      恢复只替换 <code>notes.db</code> 与 <code>media/</code>，不改环境变量。管理员账号密码会回到备份时的状态；若恢复后登不进后台，请用服务器上的
      <code>ADMIN_INIT_USERNAME / ADMIN_INIT_PASSWORD</code> 重置。恢复期间其余接口会返回 503，属正常现象。
    </div>
  </div>
</template>
