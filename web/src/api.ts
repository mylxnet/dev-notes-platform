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
  /** 项目说明的 Markdown 原文，只有后台接口下发。 */
  description?: string
  /** 封面图站内路径，空串或缺失表示没有封面。 */
  cover_image?: string
  /** 前台详情接口下发渲染并净化后的说明 HTML。 */
  description_html?: string
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
  /** 列表接口不返回正文，仅后台详情返回 Markdown 原文。 */
  content?: string
  /** 前台详情接口返回服务端渲染并净化后的 HTML。 */
  content_html?: string
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

/** 首页推荐面板需要的笔记摘要，由后台「站点设置」维护。 */
export interface FeaturedNote {
  slug: string
  title: string
  published_at: string
}

/** 站点公开信息：页脚备案号 + 首页推荐笔记。 */
export interface SiteInfo {
  icp_number: string
  featured_notes?: FeaturedNote[]
}

/** 后台站点设置：备案号与推荐位 slug（按展示顺序）。 */
export interface AdminSettings {
  icp_number: string
  featured_slugs: string[]
}

/** 后台项目列表与公开项目列表的分页响应。 */
export interface PagedProjects {
  items: Project[]
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
  error?: {
    code?: string
    message?: string
    field?: string
    hits?: string[]
    suggestion?: string
  }
}

interface DataBody<T> {
  data?: T
}

/** 校验失败时后端附带的定位信息。 */
export interface ApiErrorExtra {
  field?: string
  hits?: string[]
  suggestion?: string
}

/** 接口错误：带上 HTTP 状态码与后端错误码，便于页面区分「未找到」与「服务异常」。 */
export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly field?: string
  readonly hits?: string[]
  readonly suggestion?: string

  constructor(message: string, status: number, code = '', extra: ApiErrorExtra = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.field = extra.field
    this.hits = extra.hits
    this.suggestion = extra.suggestion
  }
}

const baseUrl = (import.meta.env.VITE_API_BASE ?? '').replace(/\/+$/, '')

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  body?: unknown
  signal?: AbortSignal
  headers?: Record<string, string>
}

/** 统一请求封装：JSON 编解码、错误归一化、可选取消与管理端自定义头。 */
export async function requestJSON<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json', ...options.headers }
  // FormData 由浏览器自动带 multipart 边界，不能手动写 Content-Type。
  const isForm = typeof FormData !== 'undefined' && options.body instanceof FormData
  if (options.body !== undefined && !isForm) {
    headers['Content-Type'] = 'application/json'
  }

  let response: Response
  try {
    response = await fetch(`${baseUrl}${path}`, {
      method: options.method ?? 'GET',
      headers,
      body: options.body === undefined ? undefined : isForm ? (options.body as FormData) : JSON.stringify(options.body),
      credentials: 'same-origin',
      signal: options.signal
    })
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw error
    }
    throw new ApiError('网络连接失败，请稍后重试', 0, 'NETWORK_ERROR')
  }

  const body = (await response.json().catch(() => null)) as (DataBody<T> & ErrorBody) | null

  if (!response.ok) {
    const detail = body?.error
    throw new ApiError(
      detail?.message ?? `请求失败（HTTP ${response.status}）`,
      response.status,
      detail?.code ?? 'HTTP_ERROR',
      { field: detail?.field, hits: detail?.hits, suggestion: detail?.suggestion }
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
  return requestJSON<PagedNotes>(`/api/notes${suffix ? `?${suffix}` : ''}`, { signal })
}

/** 查询站点公开信息（页脚备案号）。 */
export function fetchSite(signal?: AbortSignal): Promise<SiteInfo> {
  return requestJSON<SiteInfo>('/api/site', { signal })
}

/** 查询单篇已发布笔记。 */
export function fetchNote(slug: string, signal?: AbortSignal): Promise<Note> {
  return requestJSON<Note>(`/api/notes/${encodeURIComponent(slug)}`, { signal })
}

/** 查询已发布项目列表。 */
export function fetchProjects(signal?: AbortSignal): Promise<{ items: Project[]; total: number }> {
  return requestJSON<{ items: Project[]; total: number }>('/api/projects', { signal })
}

/** 查询项目详情及其已发布笔记。 */
export function fetchProject(slug: string, signal?: AbortSignal): Promise<ProjectDetail> {
  return requestJSON<ProjectDetail>(`/api/projects/${encodeURIComponent(slug)}`, { signal })
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

/** 项目状态的展示文案：项目讲的是「做完了没有」，不是「发没发布」。 */
export const projectStatusText: Record<ContentStatus, string> = {
  published: '已完成',
  draft: '草稿',
  hidden: '隐藏'
}

/** 统一把未知异常转成可展示的中文提示。 */
export function errorMessageOf(error: unknown): string {
  if (error instanceof ApiError) return error.message
  return '内容加载失败，请稍后重试'
}
