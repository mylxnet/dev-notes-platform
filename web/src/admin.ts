// 管理端接口封装：统一注入 CSRF 头，并在 401 时通知路由跳转登录。
import {
  ApiError,
  requestJSON,
  type ContentStatus,
  type Note,
  type PagedNotes,
  type PagedProjects,
  type Project,
  type SiteInfo
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

export function fetchSettings(): Promise<SiteInfo> {
  return adminRequest<SiteInfo>('/api/admin/settings')
}

export function updateSettings(input: SiteInfo): Promise<SiteInfo> {
  return adminRequest<SiteInfo>('/api/admin/settings', { method: 'PUT', body: input })
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
