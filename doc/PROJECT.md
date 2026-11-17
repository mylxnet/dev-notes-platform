# 技术方案 · 我的开发笔记

> v1.0.2 · by Mr lin · 2026-11-17
>
> 依据：`doc/DESIGN.md` v1.0.2、已确认的 `prototype/` 第一版界面。本文档是后续编码、测试、部署的实施依据，代码实现与本文档不一致时，以本文档评审结论为准。

## 一、方案目标与范围

### 1.1 目标

1. 交付一个可长期运行的个人开发笔记站点，前台公开、后台仅本人维护。
2. 满足甘肃个人 ICP 备案对“个人非经营性信息分享”的内容边界要求。
3. 单容器部署，SQLite 单文件持久化，备份即复制文件。
4. 首页以笔记为主体，项目只作为索引，弱化产品目录观感。
5. 站内不托管任何安装包、可执行文件或下载入口，只外链 GitHub 源码仓库。

### 1.2 范围内

- 前台 5 个页面：首页、笔记列表、笔记详情、项目索引与详情、关于。
- 后台 4 个模块：登录、笔记管理、项目管理、账号设置。
- 单管理员认证、内容状态管理、服务端前置校验。
- SQLite 数据持久化、Go embed 静态资源、Docker 单容器部署。
- 浅色/暗色主题、桌面/移动端响应式、基础可访问性。

### 1.3 范围外

- 注册、评论、留言、投稿、点赞、访问统计。
- 多管理员、角色权限、操作审计流水。
- 站内搜索服务（仅保留标题/摘要的本地筛选）。
- 图片上传、附件托管、对象存储。
- 付费、广告、赞助、捐赠等任何经营性能力。

### 1.4 成功标准

| 编号 | 标准 | 验证方式 |
|---|---|---|
| A1 | 前台仅展示 `published` 内容 | 接口测试 + 浏览器检查 |
| A2 | 未登录访问后台接口返回 401 | 接口测试 |
| A3 | 保存含经营性/下载类表述的内容被拒绝 | 单元测试 + 接口测试 |
| A4 | Markdown 原始 HTML 不执行 | 安全测试 |
| A5 | 容器重启后数据不丢 | 部署测试 |
| A6 | 桌面 1440px、移动 390px 无横向溢出 | 浏览器测试 |
| A7 | 控制台 0 错误、0 警告 | 浏览器测试 |
| A8 | 版本号在项目文件、界面、文档三处一致 | 发布前检查 |

## 二、总体架构

```text
Browser
  │  HTTPS（备案通过后）
  ▼
1Panel / Nginx 反向代理
  │  http://127.0.0.1:8080
  ▼
Go 1.27 + Gin 单进程
  ├── /api/*            公开与管理 API
  ├── /admin/*          管理界面（SPA 回退）
  ├── /*                前台页面（SPA 回退）
  ├── embed FS          web/dist 静态资源
  └── SQLite            /app/data/notes.db
```

### 2.1 关键架构决策

| 编号 | 决策 | 理由 |
|---|---|---|
| D1 | 笔记为主、项目为索引 | 弱化产品目录观感，符合个人备案内容边界 |
| D2 | 站内不托管 exe，仅外链 GitHub | SuperClip-C 无 LICENSE、LanFileShare 禁商业分发 |
| D3 | 单容器 + Go embed 静态资源 | 1Panel 部署运维对象最少 |
| D4 | 后台放 `/admin`，不分子域名 | 避免子域名额外备案 |
| D5 | SQLite 数据卷持久化 | 备份即复制单文件 |
| D6 | 界面采用第一版原型 | 用户对比两版后确认第一版更耐看 |
| D7 | 单 SPA、history 路由、Go 回退 | `/admin` 路径必须真实存在，不使用 hash 路由 |
| D8 | 服务端渲染 HTML + 前端 `v-html` | Markdown 净化在服务端完成，前端不信任原始输入 |
| D9 | 签名 Cookie 会话 | 单管理员场景无需会话表，密码变更即可失效旧会话 |

### 2.2 请求分区

| 前缀 | 处理方 | 鉴权 |
|---|---|---|
| `/api/notes`、`/api/projects` | 公开 API | 无 |
| `/api/admin/*` | 管理 API | 会话 Cookie + CSRF 校验 |
| `/admin`、`/admin/*` | 管理界面 SPA | 前端路由守卫 + 后端接口鉴权 |
| `/`、`/notes/*`、`/projects/*`、`/about` | 前台 SPA | 无 |
| `/healthz` | 健康检查 | 无 |

## 三、技术栈与版本基线

| 层 | 选型 | 基线版本 | 说明 |
|---|---|---|---|
| 前端框架 | Vue 3 + TypeScript | 3.5.x | Composition API，`<script setup>` |
| 前端路由 | Vue Router | 4.5.x | history 模式 |
| 构建 | Vite | 6.4.x | 相对路径 `base: './'` 不用于生产，生产使用根路径 |
| 后端语言 | Go | 1.27.0 | WSL `lxsyzd` 实测版本 |
| Web 框架 | Gin | v1.10.x | 轻量、中间件生态成熟 |
| 数据库 | SQLite | modernc.org/sqlite | 纯 Go 驱动，无 CGo |
| 密码哈希 | bcrypt | golang.org/x/crypto | cost 12 |
| Markdown | goldmark | 最新稳定版 | 禁用原始 HTML |
| HTML 净化 | bluemonday | 最新稳定版 | 白名单策略 |
| 运行时 | Node.js / npm | 22.23.2 / 10.9.8 | WSL `lxsyzd` 实测版本 |
| 容器 | Docker | 29.8.0 | WSL `lxsyzd` 实测版本 |
| 部署 | Docker Compose + 1Panel | Compose v2 | 单容器 |

**许可证约束**：仅引入 MIT / BSD / Apache 2.0 等宽松许可依赖；不引入付费授权组件；不打包 SuperClip-C、LanFileShare 的任何二进制。

## 四、目标工程目录

```text
dev-notes-platform/
├── README.md                  面向使用者
├── CHANGELOG.md               逐版本变更记录
├── CODEBUDDY.md               工程约定
├── doc/
│   ├── DESIGN.md              设计方案
│   ├── PROJECT.md             本技术方案
│   ├── PROJECT_STATE.md       状态与交接
│   └── DEPLOY.md              部署手册
├── prototype/                 已确认的第一版界面原型（冻结参考）
├── prototype-v2/              第二版候选（归档，不参与构建）
├── app/                       Go 主程序
│   ├── main.go
│   ├── embed.go
│   ├── go.mod
│   └── internal/
│       ├── api/               路由、处理器、中间件
│       ├── auth/              密码、会话、CSRF
│       ├── config/            环境变量与启动配置
│       ├── model/             领域模型与 DTO
│       ├── render/            Markdown 渲染与净化
│       ├── store/             SQLite 访问与迁移
│       └── validate/          内容校验与词汇拦截
├── web/                       Vue 前台 + 管理界面
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── src/
│   │   ├── main.ts
│   │   ├── App.vue
│   │   ├── router.ts
│   │   ├── api/               请求封装
│   │   ├── components/        通用组件
│   │   ├── views/             前台页面
│   │   ├── views/admin/       后台页面
│   │   └── styles/            主题与全局样式
│   └── dist/                  构建产物，Go embed 来源
├── deploy/
│   ├── docker-compose.yml
│   └── .env.example
├── scripts/
│   ├── build.sh               构建二进制与镜像
│   └── package.sh             生成发布包
└── test/
    ├── test_api.sh
    ├── test_security.sh
    └── test_deploy.sh
```

`prototype/` 与 `prototype-v2/` 不参与生产构建；`web/` 落地时以 `prototype/` 的视觉与交互为基准重写为正式组件。

## 五、后端设计

### 5.1 启动流程

1. 读取环境变量并校验必填项，缺失时输出明确错误并退出。
2. 创建 `DATA_DIR`，打开 `notes.db`。
3. 设置 SQLite PRAGMA：`journal_mode=WAL`、`busy_timeout=5000`、`foreign_keys=ON`、`synchronous=NORMAL`。
4. 读取 `PRAGMA user_version`，按版本顺序执行内嵌迁移脚本，全程事务。
5. 若 `admin` 表为空，使用 `ADMIN_INIT_USERNAME` / `ADMIN_INIT_PASSWORD` 创建首个管理员；密码仅用于当次 bcrypt，不写日志。
6. 注册中间件、路由与静态资源回退。
7. 监听 `HTTP_ADDR`，支持 `SIGTERM` 优雅关闭，最长等待 10 秒。

### 5.2 配置项

| 变量 | 必填 | 默认 | 说明 |
|---|---|---|---|
| `APP_ENV` | 否 | `production` | `development` / `production` |
| `HTTP_ADDR` | 否 | `:8080` | 监听地址 |
| `DATA_DIR` | 否 | `/app/data` | 数据目录 |
| `DB_PATH` | 否 | `${DATA_DIR}/notes.db` | SQLite 文件路径 |
| `SESSION_SECRET` | 是 | 无 | 至少 32 字节随机串，用于签名 Cookie |
| `ADMIN_INIT_USERNAME` | 首次启动是 | 无 | 初始管理员账号 |
| `ADMIN_INIT_PASSWORD` | 首次启动是 | 无 | 初始密码，至少 12 位 |
| `COOKIE_SECURE` | 否 | `true` | HTTPS 环境为 true；本地调试可 false |
| `TRUSTED_PROXY` | 否 | 空 | 反向代理 IP/CIDR，用于取真实客户端 IP |
| `LOG_LEVEL` | 否 | `info` | `debug` / `info` / `warn` / `error` |
| `TZ` | 否 | `Asia/Shanghai` | 时间显示时区 |

启动校验规则：`SESSION_SECRET` 缺失或过短直接拒绝启动；首次启动缺少管理员初始账号密码直接拒绝启动，并提示“请设置 ADMIN_INIT_USERNAME / ADMIN_INIT_PASSWORD”。

### 5.3 包职责

| 包 | 职责 | 关键类型 |
|---|---|---|
| `internal/config` | 读取与校验环境变量 | `Config` |
| `internal/model` | 领域模型、请求/响应 DTO | `Project`、`Note`、`Admin`、`APIError` |
| `internal/store` | SQLite 连接、迁移、CRUD | `Store`、`ProjectRepo`、`NoteRepo`、`AdminRepo` |
| `internal/api` | Gin 路由、处理器、中间件 | `Server`、`AuthMiddleware` |
| `internal/auth` | bcrypt、会话签名、CSRF | `SessionCodec`、`CSRFCodec` |
| `internal/validate` | 字段校验、slug、URL 白名单、词汇拦截 | `Validator`、`ValidationError` |
| `internal/render` | Markdown 渲染与 HTML 净化 | `MarkdownRenderer` |

### 5.4 数据模型

```sql
CREATE TABLE IF NOT EXISTS project (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT NOT NULL,
  slug        TEXT NOT NULL UNIQUE,
  summary     TEXT NOT NULL,
  tech_stack  TEXT NOT NULL DEFAULT '[]',
  repo_url    TEXT NOT NULL,
  sort_order  INTEGER NOT NULL DEFAULT 0,
  status      TEXT NOT NULL DEFAULT 'draft'
              CHECK (status IN ('published', 'draft', 'hidden')),
  created_at  TEXT NOT NULL,
  updated_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_project_status_sort
  ON project (status, sort_order, id);

CREATE TABLE IF NOT EXISTS note (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  title         TEXT NOT NULL,
  slug          TEXT NOT NULL UNIQUE,
  summary       TEXT NOT NULL,
  content       TEXT NOT NULL,
  project_id    INTEGER,
  status        TEXT NOT NULL DEFAULT 'draft'
                CHECK (status IN ('published', 'draft', 'hidden')),
  published_at  TEXT,
  created_at    TEXT NOT NULL,
  updated_at    TEXT NOT NULL,
  FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_note_status_published
  ON note (status, published_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_note_project
  ON note (project_id, status);

CREATE TABLE IF NOT EXISTS admin (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  username      TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  updated_at    TEXT NOT NULL
);
```

约定：

- 所有时间存 UTC RFC3339 字符串，展示时转 `Asia/Shanghai`。
- `tech_stack` 存 JSON 数组字符串，读取时解析为 `[]string`，最多 10 项。
- `published_at` 仅在 `status=published` 时必填；转为草稿时保留原值但不展示。
- 删除项目时笔记不删除，`project_id` 置空，前台显示为“站点记录”。

### 5.5 Store 接口

```go
type NoteRepo interface {
    ListPublished(ctx context.Context, f NoteFilter) ([]Note, error)
    GetPublishedBySlug(ctx context.Context, slug string) (*Note, error)
    ListAdmin(ctx context.Context, f AdminNoteFilter) ([]Note, error)
    GetByID(ctx context.Context, id int64) (*Note, error)
    Create(ctx context.Context, n *Note) (int64, error)
    Update(ctx context.Context, n *Note) error
    Delete(ctx context.Context, id int64) error
}
```

所有 SQL 使用参数化查询；禁止拼接用户输入；写操作使用事务；错误统一包装为 `store.ErrNotFound`、`store.ErrConflict`。

### 5.6 认证与会话

- 密码：`bcrypt.GenerateFromPassword(password, 12)`；校验使用 `bcrypt.CompareHashAndPassword`。
- 会话 Cookie 名：`dn_session`。
- Cookie 值：`base64url(payload) + "." + base64url(HMAC-SHA256(payload))`。
- payload：`v1|adminID|expiresUnix|adminUpdatedAtUnix`。
- 有效期 12 小时；每次请求校验签名、过期时间与 `admin.updated_at`。
- 修改密码会更新 `admin.updated_at`，旧 Cookie 自动失效。
- Cookie 属性：`HttpOnly`、`SameSite=Lax`、`Path=/`、`Secure=COOKIE_SECURE`、`Max-Age=43200`。
- 登录失败限流：同一 IP 15 分钟内最多 5 次失败，超出返回 429；成功登录清零。
- 登录响应不区分“账号不存在”和“密码错误”，统一提示“账号或密码不正确”。

### 5.7 中间件

| 中间件 | 作用 |
|---|---|
| `RequestID` | 生成请求 ID，写入响应头与日志 |
| `Recovery` | 捕获 panic，返回 500，不泄露堆栈 |
| `AccessLog` | 方法、路径、状态码、耗时、请求 ID；不记录密码与 Cookie |
| `SecurityHeaders` | CSP、`X-Content-Type-Options`、`X-Frame-Options`、`Referrer-Policy` |
| `Auth` | 校验会话 Cookie，失败返回 401 |
| `CSRF` | 管理端写操作校验 `Origin` 同源 + `X-CSRF-Token` 双提交 |
| `BodyLimit` | 管理端请求体上限 1 MB |

### 5.8 校验与内容安全

| 字段 | 规则 | 失败提示 |
|---|---|---|
| `name` / `title` | 1–120 字符，去首尾空格 | 标题需为 1–120 个字符 |
| `slug` | `^[a-z0-9]+(?:-[a-z0-9]+)*$`，2–80 字符，唯一 | 链接标识只能使用小写字母、数字和连字符 |
| `summary` | 1–300 字符 | 摘要需为 1–300 个字符 |
| `content` | 1–100 KB | 正文不能为空且不超过 100 KB |
| `tech_stack` | 最多 10 项，每项 1–30 字符 | 技术栈最多 10 项，单项不超过 30 个字符 |
| `repo_url` | 必须匹配 `https://github.com/{owner}/{repo}` | 源码地址仅允许 GitHub 仓库地址 |
| `status` | `published` / `draft` / `hidden` | 内容状态不合法 |
| `published_at` | `published` 时必填，RFC3339 | 发布状态必须填写发布时间 |
| 管理员密码 | 至少 12 位，含字母与数字 | 新密码至少 12 位，且同时包含字母和数字 |

词汇拦截分两级：

1. **硬拦截**：售卖、下单、支付、收费、会员、广告、赞助、捐赠、破解、激活码、代购、推广等经营性词汇。
2. **下载类拦截**：下载安装包、获取安装包、exe 下载、网盘地址、提取码等表述；允许“源码仓库”“查看源码”等合规表述。

命中时返回 422，响应中包含字段名、命中词与修正建议，例如：

```json
{
  "error": {
    "code": "CONTENT_BLOCKED",
    "message": "内容包含不符合备案边界的表述",
    "field": "content",
    "hits": ["下载安装包"],
    "suggestion": "请改为“源码仓库”或删除下载相关表述"
  }
}
```

Markdown 渲染管线：

1. `goldmark` 解析，禁用原始 HTML。
2. 链接协议白名单：`http`、`https`。
3. 图片仅允许站内相对路径 `/assets/`，不允许外链图片。
4. `bluemonday` 按白名单净化，移除 `script`、`iframe`、`style`、事件属性。
5. 输出 HTML 存入响应字段 `content_html`，前端仅对净化后的 HTML 使用 `v-html`。

### 5.9 API 契约

统一响应：

```json
{ "data": {}, "meta": {} }
```

统一错误：

```json
{ "error": { "code": "VALIDATION_FAILED", "message": "参数校验失败", "field": "slug" } }
```

公开接口：

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/notes?project={slug}&page=1&page_size=10&q={kw}` | 已发布笔记列表 |
| GET | `/api/notes/{slug}` | 已发布笔记详情，含 `content_html` |
| GET | `/api/projects` | 已发布项目列表 |
| GET | `/api/projects/{slug}` | 已发布项目详情 + 关联笔记 |

管理接口：

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/admin/login` | 登录，设置会话 Cookie，返回 CSRF Token |
| POST | `/api/admin/logout` | 退出，清除 Cookie |
| GET | `/api/admin/me` | 当前管理员信息 |
| GET | `/api/admin/notes?status=&q=&page=` | 全部状态笔记 |
| POST | `/api/admin/notes` | 新建笔记 |
| PUT | `/api/admin/notes/{id}` | 更新笔记 |
| DELETE | `/api/admin/notes/{id}` | 删除笔记 |
| GET | `/api/admin/projects` | 全部状态项目 |
| POST | `/api/admin/projects` | 新建项目 |
| PUT | `/api/admin/projects/{id}` | 更新项目 |
| DELETE | `/api/admin/projects/{id}` | 删除项目 |
| PUT | `/api/admin/password` | 修改密码 |

状态码：

| 状态码 | 场景 |
|---|---|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求体格式错误 |
| 401 | 未登录或会话失效 |
| 404 | 资源不存在或未公开 |
| 409 | slug 冲突 |
| 422 | 字段校验或内容拦截 |
| 429 | 登录失败次数过多 |
| 500 | 服务端错误 |

示例：笔记列表响应

```json
{
  "data": {
    "items": [
      {
        "id": 1,
        "title": "用纯 Go 驱动摆脱 CGo 依赖",
        "slug": "sqlite-pure-go-driver",
        "summary": "记录 SQLite 选型过程与交叉编译踩坑。",
        "project": { "name": "SiteGuard", "slug": "siteguard" },
        "published_at": "2026-11-12T00:00:00Z"
      }
    ],
    "total": 4,
    "page": 1,
    "page_size": 10
  }
}
```

### 5.10 静态资源与路由回退

- `//go:embed all:web/dist` 嵌入构建产物。
- 命中真实文件直接返回；未命中且路径以 `/api/` 开头返回 404 JSON。
- 其余路径回退到 `index.html`，由 Vue Router 接管。
- `index.html` 设置 `Cache-Control: no-cache`；带 hash 的静态资源设置长缓存。
- 生产环境启用 gzip；如反向代理已压缩则不重复压缩。

### 5.11 日志与健康检查

- 结构化 JSON 日志：`time`、`level`、`request_id`、`method`、`path`、`status`、`latency_ms`。
- 不记录密码、Cookie、CSRF Token、完整请求体。
- `GET /healthz` 返回 `{"status":"ok","db":"ok","version":"1.0.2"}`，数据库不可用时返回 503。

### 5.12 错误处理原则

- 对外返回稳定错误码与中文提示，不返回 SQL、堆栈、文件路径。
- 对内使用 `errors.Is` / `errors.As` 判断 `ErrNotFound`、`ErrConflict`。
- 所有写操作失败时保持事务回滚，前端保留用户输入并提示修正建议。

## 六、前端设计

### 6.1 工程组织

- Vue 3 + TypeScript + Vite，单 SPA，history 路由。
- 以 `prototype/` 第一版为视觉基线，拆分为正式组件。
- 不引入 UI 组件库；使用原生 CSS 变量与语义化类名，减少依赖。
- `web/src/api/` 统一封装 `fetch`：JSON、超时、401 跳转、错误归一化。
- 不引入 Pinia；会话状态用 `useAuth` composable，列表状态留在页面内。

### 6.2 路由

| 路径 | 页面 | 鉴权 |
|---|---|---|
| `/` | 首页 | 无 |
| `/notes` | 笔记列表 | 无 |
| `/notes/:slug` | 笔记详情 | 无 |
| `/projects` | 项目索引 | 无 |
| `/projects/:slug` | 项目详情 | 无 |
| `/about` | 关于 | 无 |
| `/admin/login` | 后台登录 | 无 |
| `/admin/notes` | 笔记管理 | 是 |
| `/admin/projects` | 项目管理 | 是 |
| `/admin/account` | 账号设置 | 是 |

路由守卫：访问 `/admin/*`（登录页除外）时调用 `/api/admin/me`，401 跳转登录并记录 `redirect` 参数。

### 6.3 组件清单

| 组件 | 职责 |
|---|---|
| `SiteHeader` | 主导航、后台入口、移动端折叠 |
| `SiteFooter` | 备案号占位、非经营性声明、版本与署名 |
| `ThemeSwitch` | 浅色/暗色切换，`localStorage` 持久化 |
| `NoteCard` | 笔记列表项，标题、摘要、项目、日期 |
| `ProjectCard` | 项目索引项，技术栈、关联笔记数、源码外链 |
| `EmptyState` | 空列表提示 |
| `ErrorBanner` | 接口错误与字段错误 |
| `SkeletonList` | 列表加载骨架 |
| `ConfirmDialog` | 删除确认，禁止误操作 |
| `AdminLayout` | 后台侧栏、顶栏、退出登录 |
| `NoteEditor` | 笔记表单、Markdown 预览、保存前校验 |
| `ProjectEditor` | 项目表单、排序与状态 |
| `PasswordForm` | 修改密码与二次确认 |

### 6.4 视觉与交互规范

| 项 | 规范 |
|---|---|
| 背景 | 浅色 `#faf9f5`，暗色 `#141413` |
| 主色 | 橙色 `#d97757`，仅用于强调与主按钮 |
| 文字 | 浅色 `#141413`，暗色 `#f5f4ef`；弱化文字 `#6b6b63` |
| 圆角 | 卡片 20–24px，按钮胶囊形，输入框 10–12px |
| 字体 | 系统中文字体栈；不使用 Inter；标题字重 650–800，字距收紧 |
| 间距 | 页面最大宽度 1180px；区块间距 48–88px |
| 动效 | 仅 `transform` / `opacity`；150–300ms；尊重 `prefers-reduced-motion` |
| 状态 | 每个列表必须提供加载、空、错误三态 |
| 反馈 | 保存按钮提交中禁用并显示“保存中”；删除必须二次确认 |
| 版本 | 页脚与后台侧栏显示 `v1.0.2 by Mr lin`，版本号在署名之前 |

### 6.5 后台交互

1. 登录：账号/密码前端先做长度校验，提交中禁用按钮；错误提示显示在表单下方。
2. 列表：状态筛选、关键词搜索、分页；切换筛选时保留搜索词。
3. 编辑：标题、slug、摘要、正文、关联项目、状态；slug 可手动修改，保存时服务端查重。
4. 保存：前端先做必填与长度校验，服务端再做完整校验；422 时定位到字段并展示修正建议。
5. 删除：弹出确认框，确认后执行；失败时保留列表并提示。
6. 离开未保存表单时提示“有未保存的修改”。
7. 修改密码成功后，当前会话失效并跳转登录页。

### 6.6 可访问性与性能

- 语义化标签：`header`、`nav`、`main`、`article`、`footer`。
- 所有输入框有 `label`，错误信息使用 `aria-live="polite"`。
- 键盘可达：焦点可见，弹窗焦点锁定，Esc 关闭。
- 图片设置宽高与 `loading="lazy"`；本项目无大图依赖。
- 首屏 JS 预算：gzip 后不超过 180 KB；超出时按路由拆分。
- 不引入外部字体、CDN 或统计脚本。

### 6.7 合规文案与备案位

- 页脚固定展示：`个人非经营性技术记录 · 源码仅外链 GitHub`。
- 备案号占位：`甘ICP备xxxxxxxx号（备案通过后显示）`；通过后替换为真实备案号并链接 `https://beian.miit.gov.cn`。
- 前台不出现“下载”“安装包”“购买”“会员”“赞助”等入口或按钮。
- 关于页展示非经营性声明与内容范围。

## 七、构建与发布

### 7.1 本地与 WSL 分工

- 生成动作在 `lxsyzd` WSL 内完成：依赖安装、构建、测试、镜像打包。
- 本地只保留源码、构建产物与发布包；`node_modules`、缓存、临时脚本留在 WSL。
- 依赖安装使用国内镜像：`npm config set registry https://registry.npmmirror.com`。
- 任务结束后删除 WSL 临时工作目录。

### 7.2 构建脚本

`scripts/build.sh` 步骤：

1. 读取版本号：`VERSION=$(git describe --tags --always --dirty)`，无 Git 时读取 `app/VERSION`。
2. `cd web && npm ci && npm run build`。
3. `cd app && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o ../build/dev-notes`。
4. 可选：`docker build -t dev-notes:${VERSION} .`。
5. 输出构建摘要：版本、二进制大小、前端产物大小。

要求：脚本可重复执行；失败立即退出（`set -euo pipefail`）；不把版本号写死在脚本里。

### 7.3 Dockerfile

```dockerfile
# 构建前端
FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci --registry=https://registry.npmmirror.com
COPY web/ ./
RUN npm run build

# 构建后端
FROM golang:1.27-alpine AS api
WORKDIR /src
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY app/ ./
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/dev-notes .

# 运行
FROM alpine:3.20
RUN adduser -D -u 10001 app && mkdir -p /app/data && chown -R app:app /app
WORKDIR /app
COPY --from=api /out/dev-notes /app/dev-notes
USER app
EXPOSE 8080
VOLUME ["/app/data"]
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/app/dev-notes"]
```

### 7.4 Compose

```yaml
services:
  dev-notes:
    image: dev-notes:1.0.2
    container_name: dev-notes
    restart: unless-stopped
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      APP_ENV: production
      DATA_DIR: /app/data
      SESSION_SECRET: ${SESSION_SECRET}
      ADMIN_INIT_USERNAME: ${ADMIN_INIT_USERNAME}
      ADMIN_INIT_PASSWORD: ${ADMIN_INIT_PASSWORD}
      COOKIE_SECURE: "false"
      TZ: Asia/Shanghai
    volumes:
      - /opt/1panel/apps/dev-notes/data:/app/data
```

### 7.5 1Panel 部署

1. 上传镜像或从镜像仓库拉取。
2. 「容器」→「Compose」→ 创建，粘贴 `deploy/docker-compose.yml`。
3. 在 `.env` 中填写 `SESSION_SECRET`、初始管理员账号密码。
4. 启动容器，确认 `/healthz` 返回 200。
5. 「网站」→ 反向代理到 `127.0.0.1:8080`。
6. 备案通过前不解析域名、不开放 80/443。
7. 备案通过后绑定域名、启用 HTTPS、把 `COOKIE_SECURE` 改为 `true`。

### 7.6 备份、升级、回滚

- 备份：停止写入或使用 SQLite `.backup`，复制 `notes.db` 与 `-wal`、`-shm`。
- 升级：备份数据库 → 拉取新镜像 → 重新创建容器 → 检查 `/healthz` 与前台。
- 回滚：切回上一版本镜像标签，重新创建容器；不跨版本降级数据库结构。
- 迁移：启动时自动执行；迁移失败则拒绝启动，保留原数据库。

### 7.7 版本一致性

发布前必须核对：

1. `app/main.go` 或 `app/VERSION` 的版本号。
2. `web/package.json` 的 `version`。
3. 页面页脚与后台侧栏显示的版本。
4. `README.md`、`doc/*.md`、`CHANGELOG.md` 的版本。

四处必须一致；升版后重新构建并打包，不能只改文档。

## 八、测试方案

### 8.1 测试分层

| 层级 | 工具 | 目标 |
|---|---|---|
| 单元测试 | `go test ./...` | 校验器、slug、URL 白名单、词汇拦截、Markdown 净化、会话签名 |
| 接口测试 | `httptest` + `test/test_api.sh` | 登录、CRUD、状态过滤、401、422、429 |
| 浏览器测试 | Playwright CLI | 前台 6 页 + 后台 4 模块、主题、响应式、控制台 |
| 安全测试 | 手工 + 脚本 | XSS、SQL 注入、CSRF、Cookie 属性、越权 |
| 部署测试 | Docker Compose | 启动、健康检查、重启持久化、备份恢复 |

### 8.2 单元测试清单

| 用例 | 预期 |
|---|---|
| slug 合法/非法 | 合法通过；大写、空格、下划线、中文被拒绝 |
| repo_url 白名单 | GitHub URL 通过；其他域名、http、javascript 被拒绝 |
| 技术栈数量 | 11 项被拒绝 |
| 硬拦截词 | “购买”“赞助”等返回 422 |
| 下载类表述 | “下载安装包”被拒绝，“源码仓库”通过 |
| Markdown 原始 HTML | `<script>` 被移除 |
| Markdown 链接协议 | `javascript:` 被移除 |
| 会话签名 | 篡改 payload 后校验失败 |
| 过期会话 | 返回 401 |
| 密码变更 | 旧会话失效 |

### 8.3 接口测试清单

| 用例 | 预期 |
|---|---|
| 未登录 GET `/api/admin/notes` | 401 |
| 错误密码登录 | 401，统一提示 |
| 连续 6 次失败 | 第 6 次 429 |
| 创建笔记（合法） | 201，前台可见 |
| 创建草稿 | 201，前台不可见 |
| 重复 slug | 409 |
| 内容含禁用词 | 422，返回命中的字段与建议 |
| 未登录写操作 | 401 |
| CSRF Token 缺失 | 403 |
| 删除项目 | 笔记保留，`project_id` 置空 |
| 分页与筛选 | 总数、页码、项目筛选正确 |

### 8.4 浏览器测试清单

| 页面/场景 | 检查项 |
|---|---|
| 首页 | 标题、最近笔记、项目索引、页脚备案位与版本署名 |
| 笔记列表 | 筛选、搜索、空状态、移动端 |
| 笔记详情 | 正文、关联项目、错误 slug |
| 项目索引/详情 | 状态、技术栈、关联笔记、源码外链 |
| 关于 | 非经营性声明、备案说明 |
| 登录 | 空值、短密码、错误密码、成功跳转 |
| 笔记管理 | 加载骨架、状态筛选、编辑、删除确认、保存错误提示 |
| 项目管理 | 排序、状态、编辑 |
| 账号设置 | 密码长度、两次不一致、保存成功 |
| 主题 | 浅色/暗色切换、刷新后保持 |
| 响应式 | 1440px、1024px、390px 无横向溢出 |
| 控制台 | 0 错误、0 警告 |

### 8.5 安全测试清单

| 用例 | 预期 |
|---|---|
| 笔记正文写入 `<img onerror=alert(1)>` | 前台不执行，标签被净化 |
| 搜索参数写入 SQL 片段 | 参数化查询，不报错、不泄露 |
| 跨站表单提交管理接口 | Origin/CSRF 校验拒绝 |
| 会话 Cookie 检查 | HttpOnly、SameSite=Lax、生产 Secure |
| 未登录访问 `/admin/notes` | 前端跳登录，接口 401 |
| 登录暴力尝试 | 限流生效 |
| 错误页面 | 不泄露堆栈与路径 |

### 8.6 部署测试

1. `docker compose up -d` 后 `/healthz` 返回 200。
2. 登录后台创建一篇笔记，前台可见。
3. `docker compose restart` 后数据仍在。
4. 复制数据库文件到新容器，数据可读。
5. 备份文件恢复后条目数量一致。

### 8.7 验收标准

- 8.1–8.6 全部通过，保留命令输出或截图。
- 前台无草稿/隐藏内容。
- 页面无“下载”“购买”等违规入口。
- 版本号四处一致。
- 文档与代码同批提交。

## 九、开发计划

| 阶段 | 任务 | 产出 | 验收 | 预估 |
|---|---|---|---|---|
| Section 1 | Go 骨架、配置、SQLite、迁移、模型、Store、健康检查 | 可启动二进制 | `go test ./...` 通过，`/healthz` 200 | 2 人日 |
| Section 2 | 公开 API、Vue 前台 6 页、embed、主题 | 可浏览前台 | 浏览器 6 页通过，草稿不可见 | 2 人日 |
| Section 3 | 登录、会话、CSRF、管理 API、后台 4 模块、前置校验 | 可维护内容 | 登录/CRUD/校验/401 通过 | 3 人日 |
| Section 4 | 真实内容录入、SEO 标题、合规文案、备案位 | 6 项目 + 5 笔记 | 内容与备案边界一致 | 1 人日 |
| Section 5 | Docker、Compose、1Panel、端到端验收、文档 | 可部署镜像与发布包 | 重启数据不丢，浏览器验收通过 | 1 人日 |

合计约 9 人日；每阶段完成即升小版本、同步文档、提交一次中文 commit。

## 十、风险与对策

| 风险 | 概率 | 影响 | 对策 |
|---|---|---|---|
| 备案审核对“项目展示”敏感 | 中 | 高 | 弱化产品目录，仅保留笔记与源码外链；不出现下载/售卖入口 |
| 真实仓库地址失效或许可证变化 | 中 | 中 | 上线前逐个复核；失效条目改为草稿或移除 |
| SQLite 并发写入锁 | 低 | 中 | WAL + busy_timeout；单管理员场景写入量极低 |
| Markdown XSS | 中 | 高 | 服务端净化 + 前端只渲染净化后的 HTML + 安全测试 |
| 会话密钥泄露 | 低 | 高 | 环境变量注入，不写入仓库；定期轮换并重启 |
| 容器数据卷权限 | 中 | 中 | 非 root 用户 + 启动时检查目录可写 |
| 依赖供应链风险 | 低 | 中 | 锁定版本与 `go.sum` / `package-lock.json`，只用宽松许可依赖 |
| 主题/截图偏差 | 低 | 低 | 以 `prototype/` 为准，发布前截图核对 |

## 十一、待确认事项

| 编号 | 事项 | 影响 | 建议 |
|---|---|---|---|
| Q1 | mailhub 仓库地址 404 | 是否展示该条目 | 找到正确地址或移除 |
| Q2 | 6 个项目的真实 GitHub 地址 | 源码外链与许可证复核 | 上线前逐个确认 |
| Q3 | 5 篇笔记的真实内容 | Section 4 内容录入 | 由站长提供，或按原型示例重写 |
| Q4 | 域名与备案号 | 页脚展示与 HTTPS | 备案通过后替换占位 |
| Q5 | 初始管理员账号密码 | 首次启动 | 部署时通过 `.env` 注入，不写入文档 |
| Q6 | 第二版原型是否保留 | 仓库体积与混淆 | 归档在 `prototype-v2/`，不参与构建；确认后可删除 |

## 十二、变更记录

| 版本 | 日期 | 变更 |
|---|---|---|
| v1.0.2 | 2026-11-17 | 确认第一版界面方案；新增详细技术方案 `doc/PROJECT.md`；同步版本与文档 |
| v1.0.1 | 2026-11-17 | 新增 CODEBUDDY.md 工程约定文件；完成两版界面原型 |
| v1.0.0 | 2026-11-17 | 初始设计方案 |
