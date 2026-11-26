// 管理端接口封装：统一注入 CSRF 头，并在 401 时通知路由跳转登录。
import {
  ApiError,
  baseUrl,
  requestJSON,
  type AdminSettings,
  type ContentStatus,
  type Note,
  type PagedNotes,
  type PagedProjects,
  type Project
} from './api'

export interface AdminInfo {
  id: number
  username: string
  updated_at: string
}

export interface NoteInput {
  title: string
  slug: string
  summary: string
  content: string
  project_id: number | null
  /** 是否进首页推荐面板。 */
  is_featured: boolean
  status: ContentStatus
  /** 空串表示沿用原发布时间或由后端补齐。 */
  published_at: string
}

export interface ProjectInput {
  name: string
  slug: string
  summary: string
  tech_stack: string[]
  repo_url: string
  /** 项目说明的 Markdown 原文，可留空。 */
  description: string
  /** 封面图站内路径，可留空表示没有封面。 */
  cover_image: string
  sort_order: number
  status: ContentStatus
}

/** CSRF token 保存在内存里，页面刷新后由 /api/admin/me 重新取回。 */
let csrfToken = ''
let unauthorizedHandler: (() => void) | null = null

export function setCsrfToken(token: string) {
  csrfToken = token
}

/** 注册 401 回调，由路由层跳转登录页。 */
export function onUnauthorized(handler: () => void) {
  unauthorizedHandler = handler
}

interface AdminRequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  body?: unknown
}

async function adminRequest<T>(path: string, options: AdminRequestOptions = {}): Promise<T> {
  const method = options.method ?? 'GET'
  const headers: Record<string, string> = {}
  if (method !== 'GET' && csrfToken) {
    headers['X-CSRF-Token'] = csrfToken
  }
  try {
    return await requestJSON<T>(path, { ...options, method, headers })
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      csrfToken = ''
      unauthorizedHandler?.()
    }
    throw error
  }
}

interface SessionPayload {
  admin: AdminInfo
  csrf_token: string
}

export async function login(username: string, password: string): Promise<AdminInfo> {
  const data = await requestJSON<SessionPayload>('/api/admin/login', {
    method: 'POST',
    body: { username, password }
  })
  csrfToken = data.csrf_token
  return data.admin
}

export async function fetchSession(): Promise<AdminInfo> {
  const data = await requestJSON<SessionPayload>('/api/admin/me')
  csrfToken = data.csrf_token
  return data.admin
}

export async function logout(): Promise<void> {
  try {
    await adminRequest<{ ok: boolean }>('/api/admin/logout', { method: 'POST' })
  } finally {
    csrfToken = ''
  }
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<AdminInfo> {
  const data = await adminRequest<SessionPayload>('/api/admin/password', {
    method: 'PUT',
    body: { current_password: currentPassword, new_password: newPassword }
  })
  csrfToken = data.csrf_token
  return data.admin
}

export async function previewMarkdown(content: string): Promise<string> {
  const data = await adminRequest<{ html: string }>('/api/admin/preview', {
    method: 'POST',
    body: { content }
  })
  return data.html
}

export function listNotes(params: { status?: string; keyword?: string; page?: number; pageSize?: number } = {}): Promise<PagedNotes> {
  const query = new URLSearchParams()
  if (params.status) query.set('status', params.status)
  if (params.keyword) query.set('q', params.keyword)
  if (params.page && params.page > 1) query.set('page', String(params.page))
  if (params.pageSize) query.set('page_size', String(params.pageSize))
  const suffix = query.toString()
  return adminRequest<PagedNotes>(`/api/admin/notes${suffix ? `?${suffix}` : ''}`)
}

export function getNote(id: number): Promise<Note> {
  return adminRequest<Note>(`/api/admin/notes/${id}`)
}

export function createNote(input: NoteInput): Promise<Note> {
  return adminRequest<Note>('/api/admin/notes', { method: 'POST', body: input })
}

export function updateNote(id: number, input: NoteInput): Promise<Note> {
  return adminRequest<Note>(`/api/admin/notes/${id}`, { method: 'PUT', body: input })
}

export function deleteNote(id: number): Promise<{ ok: boolean }> {
  return adminRequest<{ ok: boolean }>(`/api/admin/notes/${id}`, { method: 'DELETE' })
}

export function listProjects(params: { page?: number; pageSize?: number } = {}): Promise<PagedProjects> {
  const query = new URLSearchParams()
  if (params.page && params.page > 1) query.set('page', String(params.page))
  if (params.pageSize) query.set('page_size', String(params.pageSize))
  const suffix = query.toString()
  return adminRequest<PagedProjects>(`/api/admin/projects${suffix ? `?${suffix}` : ''}`)
}

/** 上传项目封面图，返回可直接填进封面字段的站内路径。 */
export function uploadMedia(file: File): Promise<{ path: string }> {
  const form = new FormData()
  form.append('file', file)
  return adminRequest<{ path: string }>('/api/admin/media/upload', { method: 'POST', body: form })
}

export function fetchSettings(): Promise<AdminSettings> {
  return adminRequest<AdminSettings>('/api/admin/settings')
}

export function updateSettings(input: AdminSettings): Promise<AdminSettings> {
  return adminRequest<AdminSettings>('/api/admin/settings', { method: 'PUT', body: input })
}

export function createProject(input: ProjectInput): Promise<Project> {
  return adminRequest<Project>('/api/admin/projects', { method: 'POST', body: input })
}

export function updateProject(id: number, input: ProjectInput): Promise<Project> {
  return adminRequest<Project>(`/api/admin/projects/${id}`, { method: 'PUT', body: input })
}

export function deleteProject(id: number): Promise<{ ok: boolean }> {
  return adminRequest<{ ok: boolean }>(`/api/admin/projects/${id}`, { method: 'DELETE' })
}

/** 备份概览：库与媒体规模、内容数、schema 版本与已有快照。 */
export interface BackupSummary {
  db_bytes: number
  media_files: number
  media_bytes: number
  notes: number
  projects: number
  schema_version: number
  app_version: string
  snapshots: BackupSnapshot[]
  max_archive_bytes: number
}

/** 服务器上的恢复前快照。 */
export interface BackupSnapshot {
  name: string
  bytes: number
  mod_time: string
}

/** 备份包清单与预检结论。 */
export interface BackupInspection {
  restorable: boolean
  manifest: {
    format: string
    app_version: string
    schema_version: number
    created_at: string
    entries: { path: string; bytes: number; sha256: string }[]
    counts: { notes: number; projects: number; media_files: number; media_bytes: number }
  }
  missing: string[]
  problems: string[]
  warnings: string[]
}

/** 恢复成功后的自检结果。 */
export interface RestoreOutcome {
  notes: number
  projects: number
  media_files: number
  missing: string[]
  snapshot: string
  schema_version: number
}

export function fetchBackupSummary(): Promise<BackupSummary> {
  return adminRequest<BackupSummary>('/api/admin/backup/summary')
}

/** 生成备份包并取回二进制；响应不是 JSON，不能走 requestJSON。 */
export async function downloadBackupPackage(): Promise<{ blob: Blob; filename: string }> {
  const headers: Record<string, string> = {}
  if (csrfToken) headers['X-CSRF-Token'] = csrfToken

  let response: Response
  try {
    response = await fetch(`${baseUrl}/api/admin/backup`, {
      method: 'POST',
      headers,
      credentials: 'same-origin'
    })
  } catch {
    throw new ApiError('网络连接失败，请稍后重试', 0, 'NETWORK_ERROR')
  }

  if (response.status === 401) {
    csrfToken = ''
    unauthorizedHandler?.()
  }
  if (!response.ok) {
    const detail = (await response.json().catch(() => null)) as
      | { error?: { message?: string; code?: string } }
      | null
    throw new ApiError(
      detail?.error?.message ?? `生成备份失败（HTTP ${response.status}）`,
      response.status,
      detail?.error?.code ?? 'BACKUP_FAILED'
    )
  }

  const disposition = response.headers.get('Content-Disposition') ?? ''
  const matched = /filename="?([^";]+)"?/i.exec(disposition)
  return { blob: await response.blob(), filename: matched?.[1]?.trim() ?? 'dev-notes-backup.tar.gz' }
}

function backupForm(file: File, confirm?: string): FormData {
  const form = new FormData()
  form.append('file', file)
  if (confirm) form.append('confirm', confirm)
  return form
}

/** 只预检，不改数据。 */
export function inspectBackupPackage(file: File): Promise<BackupInspection> {
  return adminRequest<BackupInspection>('/api/admin/backup/inspect', { method: 'POST', body: backupForm(file) })
}

export function restoreBackupPackage(file: File, confirm: string): Promise<RestoreOutcome> {
  return adminRequest<RestoreOutcome>('/api/admin/backup/restore', {
    method: 'POST',
    body: backupForm(file, confirm)
  })
}

export function listBackupSnapshots(): Promise<{ items: BackupSnapshot[] }> {
  return adminRequest<{ items: BackupSnapshot[] }>('/api/admin/backup/snapshots')
}

export function restoreBackupSnapshot(name: string): Promise<RestoreOutcome> {
  return adminRequest<RestoreOutcome>(`/api/admin/backup/snapshots/${encodeURIComponent(name)}/restore`, {
    method: 'POST'
  })
}
