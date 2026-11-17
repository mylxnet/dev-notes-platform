// 公开内容类型与 API 客户端：字段与后端 model 的 JSON 标签保持一致。

export type ContentStatus = 'published' | 'draft' | 'hidden'

export interface ProjectSummary {
  name: string
  slug: string
}

export interface Project {
  id: number
  name: string
  slug: string
  summary: string
  /** 后端可能返回 null，展示前需兜底为空数组。 */
  tech_stack: string[] | null
  repo_url: string
  sort_order: number
  status: ContentStatus
  /** 已发布关联笔记数，由后端统计，避免前端逐个请求。 */
  note_count: number
  created_at: string
  updated_at: string
}

export interface Note {
  id: number
  title: string
  slug: string
  summary: string
  /** 列表接口不返回正文，仅详情接口返回。 */
  content?: string
  project_id?: number
  project?: ProjectSummary
  status: ContentStatus
  published_at?: string
  created_at: string
  updated_at: string
}

export interface PagedNotes {
  items: Note[]
  total: number
  page: number
  page_size: number
}

export interface ProjectDetail {
  project: Project
  notes: Note[]
  note_count: number
}

interface ErrorBody {
  error?: { code?: string; message?: string }
}

interface DataBody<T> {
  data?: T
}

/** 接口错误：带上 HTTP 状态码与后端错误码，便于页面区分「未找到」与「服务异常」。 */
export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(message: string, status: number, code = '') {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

const baseUrl = (import.meta.env.VITE_API_BASE ?? '').replace(/\/+$/, '')

async function request<T>(path: string, signal?: AbortSignal): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${baseUrl}${path}`, {
      headers: { Accept: 'application/json' },
      signal
    })
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw error
    }
    throw new ApiError('网络连接失败，请稍后重试', 0, 'NETWORK_ERROR')
  }

  const body = (await response.json().catch(() => null)) as (DataBody<T> & ErrorBody) | null

  if (!response.ok) {
    throw new ApiError(
      body?.error?.message ?? `请求失败（HTTP ${response.status}）`,
      response.status,
      body?.error?.code ?? 'HTTP_ERROR'
    )
  }
  if (!body || !('data' in body)) {
    throw new ApiError('接口返回格式异常', response.status, 'INVALID_PAYLOAD')
  }
  return body.data as T
}

/** 查询已发布笔记，支持项目筛选、关键词与分页。 */
export function fetchNotes(
  params: { project?: string; keyword?: string; page?: number; pageSize?: number } = {},
  signal?: AbortSignal
): Promise<PagedNotes> {
  const query = new URLSearchParams()
  if (params.project) query.set('project', params.project)
  if (params.keyword) query.set('q', params.keyword)
  if (params.page && params.page > 1) query.set('page', String(params.page))
  if (params.pageSize) query.set('page_size', String(params.pageSize))
  const suffix = query.toString()
  return request<PagedNotes>(`/api/notes${suffix ? `?${suffix}` : ''}`, signal)
}

/** 查询单篇已发布笔记。 */
export function fetchNote(slug: string, signal?: AbortSignal): Promise<Note> {
  return request<Note>(`/api/notes/${encodeURIComponent(slug)}`, signal)
}

/** 查询已发布项目列表。 */
export function fetchProjects(signal?: AbortSignal): Promise<{ items: Project[]; total: number }> {
  return request<{ items: Project[]; total: number }>('/api/projects', signal)
}

/** 查询项目详情及其已发布笔记。 */
export function fetchProject(slug: string, signal?: AbortSignal): Promise<ProjectDetail> {
  return request<ProjectDetail>(`/api/projects/${encodeURIComponent(slug)}`, signal)
}

/** 把后端返回的 RFC3339 时间格式化为 YYYY-MM-DD，空值返回空串。 */
export function formatDate(value?: string | null): string {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

/** 项目技术栈兜底，避免后端空数组或 null 导致渲染异常。 */
export function techStackOf(project: Project): string[] {
  return Array.isArray(project.tech_stack) ? project.tech_stack : []
}

/** 内容状态的展示文案。 */
export const statusText: Record<ContentStatus, string> = {
  published: '已发布',
  draft: '草稿',
  hidden: '隐藏'
}

/** 统一把未知异常转成可展示的中文提示。 */
export function errorMessageOf(error: unknown): string {
  if (error instanceof ApiError) return error.message
  return '内容加载失败，请稍后重试'
}
