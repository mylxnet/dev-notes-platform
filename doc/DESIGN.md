# 设计方案 · 我的开发笔记

> v1.4.10 · by Mr lin · 2026-11-27

## 一、开发目的

搭建一个真实可用的个人站点，用于展示已完成的软件项目与配套开发笔记，并提供一个仅本人使用的后台管理入口。站点同时作为甘肃个人 ICP 备案的申报主体，内容边界必须严格符合“个人非经营性信息分享”的定性。

## 二、定位与合规边界

| 维度 | 界定 |
|---|---|
| 主体 | 个人，非企业、非组织 |
| 性质 | 非经营性信息分享，不提供商品售卖、付费服务、广告、赞助、捐赠 |
| 内容 | 开发笔记、问题排查、设计取舍、项目复盘 |
| 软件呈现 | 仅作个人练习作品记录，源码入口外链 GitHub；站内不托管任何安装包或可执行文件 |
| 社交功能 | 无注册、无登录（前台）、无评论、无投稿、无留言板 |
| 后台 | 单管理员账号，`/admin` 路径，不开放注册 |
| 名称 | 我的开发笔记（全中文 5 字，无禁用词，名称与页面标题一致） |

**许可证约束**：SuperClip-C 无 LICENSE（仅公开可读），LanFileShare 明确禁止商业分发。因此站点不得出现“下载”“获取安装包”等表述，只写“源码仓库”外链。

## 三、技术栈

| 层 | 选型 | 理由 |
|---|---|---|
| 前台 | Vue 3 + TypeScript + Vite | 用户熟练，静态构建 |
| 后台 API | Go 1.27 + Gin | 单二进制、依赖少、并发稳定 |
| 数据库 | SQLite（modernc 纯 Go 驱动） | 零运维、文件型、避免 CGo |
| 管理界面 | Vue 3 + TypeScript（与前台同工程） | 与前台同工具链，`embed` 一并嵌入 |
| 认证 | 单管理员账号 + 密码哈希（bcrypt），会话 Cookie | 无注册入口，最小实现 |
| 部署 | Docker 单容器 + 1Panel Compose | 运维对象最少，1Panel 原生支持 |

**不使用云服务**：站点域名须指向境内 ECS 以满足备案，数据留在自有服务器。

## 四、项目结构

```
dev-notes-platform/
├── README.md
├── CHANGELOG.md
├── CODEBUDDY.md
├── doc/
│   ├── DESIGN.md
│   ├── PROJECT.md
│   ├── DEPLOY.md
│   └── PROJECT_STATE.md
├── prototype/                 # 已确认的第一版界面原型（冻结参考）
├── prototype-v2/              # 第二版候选（归档，不参与构建）
├── deploy/
│   └── docker-compose.yml
├── app/                       # Go 主程序
│   ├── main.go
│   ├── embed.go
│   ├── internal/
│   │   ├── api/
│   │   ├── auth/
│   │   ├── config/
│   │   ├── model/
│   │   ├── render/
│   │   ├── store/
│   │   └── validate/
│   └── go.mod
├── content/notes/             # 文章源（Markdown + 头部字段），由导入工具写入站点
├── tools/                     # 可直接使用的运维工具（如批量导入文章）
├── web/                       # Vue 前台 + 管理界面源码
│   ├── src/
│   ├── public/media/          # 文章内插图（SVG），构建时原样进入 dist/
│   └── dist/                  # 构建产物，由 Go embed 嵌入
├── scripts/
│   └── build.sh
└── test/
    └── smoke_test.sh
```

## 五、功能清单

### 前台（公开）

1. 首页：站点定位、最近笔记列表（标题 + 摘要 + 日期）、项目索引标签
2. 笔记列表：按日期倒序，支持按项目筛选
3. 笔记详情：Markdown 正文渲染
4. 项目列表与详情：定位、技术栈、源码仓库外链、关联笔记
5. 关于：站点说明、非经营性声明、备案号占位
6. 页脚：备案号位置（备案通过后显示）、版本与署名在关于页展示

### 后台（单管理员）

1. 登录：`/admin` 路径，账号密码校验
2. 项目：新增、编辑、排序、状态（已发布 / 草稿 / 隐藏）、封面图上传、详情说明
3. 笔记：新增、编辑、Markdown 正文、关联项目、状态、设为首页推荐（`is_featured`）
4. 站点设置：维护 ICP 备案号
5. 数据备份与恢复（v1.4.10 起后端接口可用，后台界面待接入）：下载 `.tar.gz` 备份、上传预检、恢复前自动快照、快照一键回滚
6. 账号：修改管理员密码
7. 前置校验：保存时拦截经营性词汇与敏感词，命中即拒绝并给出具体提示
8. 发布规则：草稿与隐藏状态不出现在前台

### 界面方案确认（v1.0.2）

| 方案 | 视觉方向 | 结论 |
|---|---|---|
| 第一版 `prototype/` | 柔和大留白、圆角卡片、轻量现代、橙色点缀 | **确认采用**，作为正式界面基线 |
| 第二版 `prototype-v2/` | 编辑部极简、超大标题、细线网格、编号索引 | 归档留档，不参与构建 |

两版原型均完成桌面与移动端浏览器验收，控制台 0 错误。第一版更耐看、与个人笔记站调性更一致。

## 六、数据模型

```
project
  id, name, slug, summary, tech_stack, repo_url,
  description(markdown 说明), cover_image(封面站内路径),
  sort_order, status(published|draft|hidden), created_at, updated_at

note
  id, title, slug, summary, content(markdown), project_id,
  status(published|draft|hidden), published_at, created_at, updated_at

admin
  id, username, password_hash, updated_at

settings
  key, value, updated_at          -- 站点配置，当前存 icp_number
```

字段约束、索引与迁移策略见 `doc/PROJECT.md` 第五节。

## 七、API 设计

| 方法 | 路径 | 用途 | 鉴权 |
|---|---|---|---|
| GET | `/api/notes` | 笔记列表（仅已发布） | 无 |
| GET | `/api/notes/:slug` | 笔记详情 | 无 |
| GET | `/api/projects` | 项目列表（仅已发布） | 无 |
| GET | `/api/projects/:slug` | 项目详情 | 无 |
| GET | `/api/site` | 站点公开信息（页脚备案号） | 无 |
| GET/PUT | `/api/admin/settings` | 站点配置（备案号） | 是 |
| POST | `/api/admin/media/upload` | 项目封面图上传 | 是 |
| GET | `/media/uploads/:name` | 运行时上传图片下发 | 无 |
| POST | `/api/admin/login` | 登录 | 无 |
| POST | `/api/admin/logout` | 退出 | 是 |
| GET/POST/PUT/DELETE | `/api/admin/projects` | 项目管理 | 是 |
| GET/POST/PUT/DELETE | `/api/admin/notes` | 笔记管理 | 是 |
| PUT | `/api/admin/password` | 修改密码 | 是 |
| GET | `/api/admin/backup/summary` | 备份概览：库/媒体规模、内容数、schema 版本、已有快照 | 是 |
| POST | `/api/admin/backup` | 生成备份包并以 `.tar.gz` 附件下发，服务器不留临时包 | 是 + CSRF |
| POST | `/api/admin/backup/inspect` | 上传包预检：解包、校验 sha256、比对 schema 版本、检查图片，不改数据 | 是 + CSRF |
| POST | `/api/admin/backup/restore` | 用上传包覆盖当前数据，需 `confirm=RESTORE`，覆盖前自动快照 | 是 + CSRF |
| GET | `/api/admin/backup/snapshots` | 列出恢复前快照，供一键回滚 | 是 |
| POST | `/api/admin/backup/snapshots/:name/restore` | 用指定快照回滚，文件名走白名单校验 | 是 + CSRF |

请求/响应结构、错误码与校验规则见 `doc/PROJECT.md` 第五节。

## 八、安全与容错

- 密码用 bcrypt 哈希，明文不落库、不进日志
- 会话 Cookie 设置 HttpOnly、SameSite=Lax；后台接口统一鉴权中间件
- SQLite 使用参数化查询，禁止字符串拼接
- Markdown 渲染禁用原始 HTML，防止存储型 XSS
- 前置校验在服务端执行，不依赖前端
- 数据库文件放 `/app/data/`，容器重启不丢数据
- 封面上传只收 PNG/JPG/WEBP（内容嗅探 + 2 MB 上限 + 随机文件名），禁 SVG；文件存 `DATA_DIR/media/`，与备份一并保留
- 后台入口 `/admin`，不单独使用子域名，避免额外备案

## 九、部署方案

见 `doc/DEPLOY.md`。要点：单容器 Compose，SQLite 数据卷持久化，1Panel 反向代理到域名，备案通过后再开放 80/443。

## 十、测试策略

1. 单元测试：校验器、仓库链接白名单、Markdown 净化
2. 接口冒烟：登录、增删改查、草稿不出现在前台
3. 浏览器测试：前台 5 个页面 + 后台 3 个模块，浅色/暗色、桌面/移动端
4. 安全测试：未登录访问后台接口返回 401，敏感词保存被拒绝
5. 部署测试：1Panel 部署后重启容器，数据不丢

详细测试用例见 `doc/PROJECT.md` 第八节。

## 十一、开发计划（按 section 推进）

1. Section 1：Go 骨架 + SQLite + 数据模型 + 基础 API（已完成）
2. Section 2：前台页面 + 构建产物 embed（已完成）
3. Section 3：后台管理界面 + 认证 + 前置校验（已完成）
4. Section 4：真实内容录入（6 个项目 + 5 篇笔记）
5. Section 5：Docker 封装 + 1Panel 部署 + 浏览器验收

任务拆分、产出与验收标准见 `doc/PROJECT.md` 第九节。

## 十二、变更记录

| 版本 | 日期 | 变更 |
|---|---|---|
| v1.4.10 | 2026-11-27 | 新增数据备份与恢复后端能力：`.tar.gz` 打包下载（库 + 媒体 + manifest/sha256）、上传预检、恢复前自动快照与失败回滚、快照一键回滚、恢复期维护拦截；补 6 个端到端接口测试，前端界面待接入 |
| v1.4.9 | 2026-11-26 | 推荐改为笔记自身属性（`note.is_featured`，schema 4）：后台编辑页勾选、列表显示徽标；首页面板取最新 3 条，「全部」进 `/notes?featured=1` 全量推荐列表；站点设置只留备案号 |
| v1.4.8 | 2026-11-26 | 主题按钮：桌面保留长文案 + `title` 提示，≤640px 收成 44px 圆形纯圆点不再压表单；首屏主题脚本改为 `public/theme-init.js`（生产 CSP 不放行内联） |
| v1.4.7 | 2026-11-25 | 项目页改版：横向缩略图卡片（contain 以高为准、居中留空）、详情页图文并排、截图点击灯箱放大 |
| v1.4.6 | 2026-11-17 | 补齐 Docker 构建链（镜像内编译、1Panel Compose、amd64 阿里云推送脚本），部署手册落地双路线 |
| v1.4.5 | 2026-11-17 | 笔记正文与项目说明接入带工具栏、实时分栏预览与图片上传的 Markdown 编辑器 |
| v1.4.4 | 2026-11-17 | 后台项目封面支持安全上传（png/jpg/webp、≤2 MB、随机名），存 DATA_DIR/media 并由独立路由下发 |
| v1.4.3 | 2026-11-17 | 后台「站点设置」维护 ICP 备案号，页脚改为读取公开接口；新增 settings 键值表（schemaVersion 3） |
| v1.4.2 | 2026-11-17 | 测试构建在登录框预填凭据以便本机验收，正式构建不含凭据并加开关防误用 |
| v1.4.1 | 2026-11-17 | 后台笔记/项目列表按 6 条/页分页，分页条抽成 `Pager.vue` 供前后台三个列表复用 |
| v1.4.0 | 2026-11-17 | 项目增加「详情说明（Markdown）」与「封面图（站内路径）」：老库按 v2 迁移自动补列；项目卡片显示封面，详情页渲染说明 HTML，后台表单可写可预览 |
| v1.3.5 | 2026-11-17 | 新增 excelMailmerge 项目与 4 篇复盘；确立「同批文章不套同一骨架」的写作约定 |
| v1.3.4 | 2026-11-17 | 录入 SiteGuard 的 4 篇技术复盘，其中 `reverse-proxy-origin-403` 启用 `/media/` 插图，配图通道端到端跑通 |
| v1.3.3 | 2026-11-17 | 录入 LanFileShare 的 4 篇技术复盘；三个项目卡片的仓库地址与技术栈改为真实值 |
| v1.3.2 | 2026-11-17 | 新增 `content/notes/` 文章源与 `tools/import-notes.sh` 导入工具（按 slug 幂等），录入 SuperClip-C 的 4 篇技术复盘 |
| v1.3.1 | 2026-11-17 | 修复 HTTP 访问下 Secure Cookie 被浏览器丢弃导致后台登录后立即失效；`COOKIE_SECURE` 改为 auto/true/false 三档 |
| v1.3.0 | 2026-11-17 | 新增站内配图目录 `/media/`：图片白名单放宽至 `/assets/` 与 `/media/`；正文图片等比自适应；新增两张技术插图（SiteGuard 同源判定、mailhub IMAP 游标）；修正页脚版本行被主题按钮遮挡 |
| v1.2.0 | 2026-11-17 | 完成 Section 3：后台管理与认证、CSRF、登录限流、内容前置校验、Markdown 渲染净化、前台正文服务端渲染 |
| v1.1.0 | 2026-11-17 | 完成 Section 2：正式前台接入真实 API、history 路由与 SPA 回退、静态资源 embed、构建脚本与浏览器验收 |
| v1.0.3 | 2026-11-17 | 完成 Section 1：Go 骨架、SQLite 迁移、数据模型、Store、公开 API、健康检查与测试 |
| v1.0.2 | 2026-11-17 | 确认第一版界面方案；新增详细技术方案 `doc/PROJECT.md`；同步版本与文档 |
| v1.0.1 | 2026-11-17 | 新增 CODEBUDDY.md 工程约定文件；完成两版界面原型 |
| v1.0.0 | 2026-11-17 | 初始设计方案 |
