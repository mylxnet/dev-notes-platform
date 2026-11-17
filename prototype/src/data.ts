export type ContentStatus = 'published' | 'draft' | 'hidden'

export interface Project {
  id: number
  slug: string
  name: string
  summary: string
  tech: string[]
  repoLabel: string
  status: ContentStatus
  sortOrder: number
}

export interface Note {
  id: number
  slug: string
  title: string
  summary: string
  content: string[]
  projectId: number
  status: ContentStatus
  publishedAt: string
}

export const statusText: Record<ContentStatus, string> = {
  published: '已发布',
  draft: '草稿',
  hidden: '隐藏'
}

export const projects: Project[] = [
  {
    id: 1,
    slug: 'superclip',
    name: 'SuperClip-C',
    summary: '剪贴板历史与片段管理工具，支持格式化文本与快捷粘贴。',
    tech: ['Go', 'WPF', 'SQLite'],
    repoLabel: '源码仓库',
    status: 'published',
    sortOrder: 1
  },
  {
    id: 2,
    slug: 'lanfileshare',
    name: 'LanFileShare',
    summary: '局域网内的轻量文件互传工具，零配置发现设备。',
    tech: ['Go', 'HTTP', 'Web'],
    repoLabel: '源码仓库',
    status: 'published',
    sortOrder: 2
  },
  {
    id: 3,
    slug: 'keynest',
    name: 'KeyNest',
    summary: '本地密钥与配置片段管理器，按项目分组检索。',
    tech: ['TypeScript', 'Electron'],
    repoLabel: '源码仓库',
    status: 'draft',
    sortOrder: 3
  },
  {
    id: 4,
    slug: 'siteguard',
    name: 'SiteGuard',
    summary: '站点可用性巡检与日志摘要小工具，适合单台服务器自托管。',
    tech: ['Python', 'FastAPI'],
    repoLabel: '源码仓库',
    status: 'published',
    sortOrder: 4
  },
  {
    id: 5,
    slug: 'notes-cli',
    name: 'Notes CLI',
    summary: '命令行快速记录与检索开发笔记的小工具。',
    tech: ['Go', 'Cobra'],
    repoLabel: '源码仓库',
    status: 'draft',
    sortOrder: 5
  },
  {
    id: 6,
    slug: 'mailhub',
    name: 'Mail Hub',
    summary: '邮件归档实验项目，仓库地址待确认，暂不对外展示。',
    tech: ['Node.js'],
    repoLabel: '源码仓库',
    status: 'hidden',
    sortOrder: 6
  }
]

export const notes: Note[] = [
  {
    id: 1,
    slug: 'sqlite-pure-go-driver',
    title: '用纯 Go 驱动摆脱 CGo 依赖',
    summary: '记录 SQLite 选型过程：为什么选择 modernc 驱动，以及交叉编译时踩过的坑。',
    content: [
      '项目需要一个零运维的本地存储方案，SQLite 是自然的选项。问题出在驱动上：常见的 mattn/go-sqlite3 依赖 CGo，一旦开启交叉编译，构建链就变复杂。',
      '最终选择 modernc.org/sqlite。它是纯 Go 实现，不需要 C 编译器，单条 go build 就能产出静态二进制。代价是构建时间稍长，但对于个人项目完全可以接受。',
      '这次取舍的结论很直接：当部署环境不可控时，少一个系统依赖比少几秒构建时间更重要。'
    ],
    projectId: 4,
    status: 'published',
    publishedAt: '2026-11-12'
  },
  {
    id: 2,
    slug: 'clipboard-history-design',
    title: '剪贴板历史的数据结构设计',
    summary: '把剪贴板内容拆成文本、文件路径与代码片段三类后，检索和清理都变简单了。',
    content: [
      '最初把所有剪贴板内容都存成字符串，结果文件路径和长代码混在一起，检索体验很差。',
      '后来按内容形态拆成三类：纯文本、文件路径、代码片段。每一类有独立的摘要生成规则，列表页只显示摘要，详情页再展示完整内容。',
      '这个改动没有引入新依赖，只是把数据结构理顺，却让后续的功能扩展轻松很多。'
    ],
    projectId: 1,
    status: 'published',
    publishedAt: '2026-11-08'
  },
  {
    id: 3,
    slug: 'lan-discovery',
    title: '局域网设备发现的两种思路',
    summary: '对比 UDP 广播与轻量 HTTP 服务发现，记录在小规模家庭网络里的实际表现。',
    content: [
      '第一种思路是 UDP 广播：客户端发一个探测包，设备收到后回应自身信息。优点是零配置，缺点是部分路由器会拦截广播包。',
      '第二种思路是在已知网段内并发探测一个固定 HTTP 端口。实现更简单，但需要用户确认网段。',
      '最终采用两种方式的组合：优先广播，失败后回退到网段探测，并在界面上明确提示当前使用的方式。'
    ],
    projectId: 2,
    status: 'published',
    publishedAt: '2026-10-29'
  },
  {
    id: 4,
    slug: 'markdown-sanitize',
    title: 'Markdown 渲染与存储型 XSS 防护',
    summary: '关闭原始 HTML 只是第一步，还要处理链接协议与图片来源。',
    content: [
      '笔记系统允许作者写 Markdown，如果不加限制，渲染出来的原始 HTML 会成为存储型 XSS 的入口。',
      '处理分三层：渲染器关闭原始 HTML 支持；链接只放行 http 与 https 协议；图片只允许站内路径。',
      '服务端保存时再做一次校验，不依赖前端拦截。这样即使后台被绕过，前台也不会执行危险内容。'
    ],
    projectId: 4,
    status: 'published',
    publishedAt: '2026-10-20'
  },
  {
    id: 5,
    slug: 'config-encryption',
    title: '本地配置加密的取舍',
    summary: '密钥存在本机时，加密到底防住了什么，又带来了什么新麻烦。',
    content: [
      '把密钥明文放在配置文件里总让人不安，但引入加密也会带来密钥管理的新问题。',
      '如果解密密钥和加密配置放在同一台机器上，那么加密能防的只是误提交到公开仓库，而不是针对本机的攻击。',
      '因此这一版只做最低目标：把配置模板与真实密钥分离，并在 README 里写清楚不要把真实配置提交到版本库。'
    ],
    projectId: 3,
    status: 'draft',
    publishedAt: '2026-10-05'
  }
]

export function projectBySlug(slug: string): Project | undefined {
  return projects.find((item) => item.slug === slug)
}

export function noteBySlug(slug: string): Note | undefined {
  return notes.find((item) => item.slug === slug)
}

export function projectById(id: number): Project | undefined {
  return projects.find((item) => item.id === id)
}

export function publishedNotes(): Note[] {
  return notes.filter((item) => item.status === 'published')
}

export function publishedProjects(): Project[] {
  return projects.filter((item) => item.status === 'published').sort((a, b) => a.sortOrder - b.sortOrder)
}
