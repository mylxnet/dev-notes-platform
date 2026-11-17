# 项目状态 · 我的开发笔记

> v1.2.0 · by Mr lin · 2026-11-17

## 当前版本

v1.2.0 · Section 3 已完成：后台管理与认证、内容前置校验、Markdown 渲染净化、前台正文改为服务端渲染；下一阶段 Section 4（真实内容录入）。

## Section 3 实施记录（2026-11-17）

- 新增 `internal/auth`：签名 Cookie 会话（`dn_session`，12 小时，payload `v1|adminID|expires|admin.updated_at`）、CSRF 双提交（`dn_csrf` + `X-CSRF-Token`）、登录失败限流（同 IP 15 分钟 5 次）、`RequireAdmin` 中间件。
- 新增 `internal/validate`：字段长度/slug 正则/GitHub 仓库地址/技术栈数量/密码强度，以及「经营性词汇」与「下载类表述」两级拦截，命中返回 `CONTENT_BLOCKED` + `hits` + `suggestion`。
- 新增 `internal/render`：goldmark（GFM）+ bluemonday 白名单净化；不启用 `html.WithUnsafe()`，原始 HTML 丢弃；链接限 `http/https/站内/锚点`；图片仅允许 `/assets/`。
- 新增管理 API：`POST /api/admin/login|logout`、`GET /api/admin/me`、`PUT /api/admin/password`、`POST /api/admin/preview`、笔记与项目 CRUD。写操作统一 CSRF 校验，请求体上限 1 MB。
- 新增后台界面：登录页、`AdminLayout` 侧栏布局、笔记管理（状态筛选 + 搜索 + 行内删除确认）、笔记编辑（Markdown 预览 + 字段级错误定位 + 未保存离开提示）、项目管理、项目编辑、账号设置。
- 前台笔记详情改用 `content_html`（服务端净化后的 HTML），并补充 `.article-markdown` 排版样式；站点头部恢复「后台管理」入口。
- `store` 补充 `GetAdminByID`、`UpdateAdminPassword`（`updated_at` 单调递增保证旧会话失效）、`CountAllNotes`。
- 路由 `api.New` 增加会话管理器参数；`main.go` 创建 `auth.Manager`。

### 验收证据（2026-11-17）

- `go test ./... -count=1` 全绿：`api`（17 个接口用例，含认证/CSRF/CRUD/拦截/限流）、`auth`、`validate`、`render`、`store`、`config`。
- `bash scripts/build.sh` 通过版本一致性校验并产出 `app/build/dev-notes`（v1.2.0，前端 151 KB / gzip 54 KB）。
- 浏览器交互验收（Playwright + 本机 Chrome，WSL 实例 IP 直连）：
  - 未登录访问 `/admin/notes` → 跳 `/admin/login?redirect=/admin/notes`；
  - 登录成功进入列表，3 行数据；新建笔记保存成功并出现在列表；
  - 前台访问新笔记：`h2=1`、`pre>code=1`、`li=2`，Markdown 渲染正常；
  - 提交含「网盘地址 / 获取安装包」的正文 → 422，页面显示「命中表述：获取安装包、网盘地址；修正建议：请改为源码仓库或删除下载相关表述」；
  - 项目管理 3 行、7 列表头正常；删除笔记后列表不再出现；
  - 修改密码成功提示「密码已更新，当前会话已自动续期」；
  - 控制台无异常（仅未登录探测的 401 属预期）。

## Section 2 实施记录（2026-11-17）

- 前端 `web/` 由原型工程转为正式实现，移除原型专用模块：`auth.ts`（模拟登录）、`toast.ts` 与 `ToastMessage.vue`（全局提示）、`data.ts`（mock 内容）。
- 新增 `web/src/api.ts`：公开内容类型、`fetchNotes` / `fetchNote` / `fetchProjects` / `fetchProject`、`ApiError`、`formatDate`、`techStackOf`、`errorMessageOf`。
- 五个前台页面接入真实接口：首页（最近笔记 + 前两个项目）、笔记列表（项目筛选 + 关键词搜索 + 计数）、笔记详情、项目索引、项目详情（关联笔记）。
- 页面三态：骨架屏加载、错误横幅（带重试）、空状态；列表请求带 300ms 防抖与 `AbortController` 取消。
- 路由改为 `createWebHistory`，Go `NoRoute` 处理：`/api/` 前缀返回 JSON 404，静态文件命中直接返回，其余回退 `index.html`。
- 新增 `app/internal/webui`：`embed_on.go`（`//go:build embedui`，`//go:embed all:dist`）与 `embed_off.go`（默认空实现），默认构建不嵌入资源。
- 新增 `scripts/build.sh`：校验 `app/main.go` 版本与 `app/VERSION` 一致 → `npm ci` → `npm run build` → `CGO_ENABLED=0 go build -tags embedui -ldflags -X main.version`。
- 后端 `model.Project` 增加 `note_count`，由 `projectColumns` 子查询统计已发布笔记数，供项目卡片直接展示。
- 版本升至 `v1.1.0`：`app/VERSION`、`app/main.go`、`web/package.json`、`web/package-lock.json`、页脚与文档同步。

### 验收证据（2026-11-17）

- `go vet ./...` 无输出；`go test ./... -count=1`：`internal/api`、`internal/config`、`internal/store` 全部通过（含新增 `note_count`、静态资源与 SPA 回退用例）。
- `bash scripts/build.sh`：前端 49 个模块构建通过，`app/build/dev-notes` 产出成功。
- 接口冒烟：`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.1.0"}`；`/`、`/notes/hello` 均 200（SPA 回退生效）；`/api/missing` 返回 404。
- 浏览器交互验收（Playwright + 本机 Chrome，临时库 3 项目 / 3 笔记）：
  - 首页：3 条笔记、2 个项目；
  - 笔记列表：初始 3 条（共 3 篇），筛选 SiteGuard 后 1 条（共 1 篇）；
  - 搜索 `CGo` 命中「用纯 Go 驱动摆脱 CGo 依赖」，搜索无结果返回空状态；
  - 笔记详情标题正确；`/notes/not-exist` 显示「笔记不存在或未公开」且过渡后不透明度为 1；
  - 项目索引 3 张卡片，均显示「1 篇关联笔记」；
  - 页面无 JS 异常（仅 `not-exist` 详情接口预期的 404）。

## Section 1 实施记录（2026-11-17）

- 建立 Go 模块 `app/go.mod`，模块路径 `dev-notes-platform`，Go 1.27.0。
- 完成配置加载与启动校验：`SESSION_SECRET`、`ADMIN_INIT_USERNAME`、`ADMIN_INIT_PASSWORD`、`DATA_DIR`、`HTTP_ADDR` 等。
- 完成 SQLite 连接与迁移：WAL、`busy_timeout`、外键、`user_version`，建立 project / note / admin 表与索引。
- 完成模型与 Store：项目、笔记、管理员的查询与增删改；已发布内容过滤；唯一约束与 `ErrNotFound` 处理。
- 完成公开 API：`/healthz`、`/api/projects`、`/api/projects/:slug`、`/api/notes`、`/api/notes/:slug`。
- 完成管理员种子：首次启动按环境变量写入 bcrypt 哈希，已存在则跳过。
- 未完成（已顺延）：后台登录/管理 API、内容校验、Markdown 渲染、Docker 与 1Panel 部署。

## 设计收尾记录（2026-11-17）

- 完成两版界面原型并做浏览器验收：第一版 `prototype/` 采用柔和圆角与笔记驱动布局；第二版 `prototype-v2/` 为编辑部极简方向。
- 用户对比后确认采用第一版；第二版归档，不参与正式构建。
- 完成 `doc/PROJECT.md` 详细技术方案，覆盖架构、数据模型、API、认证、校验、测试、构建部署与开发计划。
- 截图更新：`prototype/screenshots/home.png`、`admin-notes.png`、`dark.png`、`mobile.png`。

## 已完成

- 6 个仓库的公开信息调研（定位、技术栈、许可证、发行状态）
- 后台边界裁决：方案 A，纯内容管理，无注册、无评论
- 技术栈裁决：Go 1.27 + Gin + SQLite，单容器 1Panel 部署
- 测试环境确认：lxsyzd WSL（Ubuntu 22.04，Go 1.27.0，Node 22.23.2，Docker 29.8.0）
- 第一版界面原型 `prototype/`：前台 6 页 + 后台 4 模块，已通过浏览器验收
- 第二版界面原型 `prototype-v2/`：编辑部极简方向，对比后归档，不参与构建
- 界面方案确认：采用第一版，页面、主题、交互以 `prototype/` 为基线
- 设计方案、部署手册、项目状态文档、详细技术方案
- Git 仓库初始化（`main`），首个基线提交 `3f41179`
- Section 1：Go 模块、配置、SQLite 迁移、模型、Store、公开 API、健康检查、单元与接口测试
- Section 2：正式前台、真实 API 接入、history 路由、静态资源 embed、构建脚本与浏览器验收
- Section 3：后台管理与认证、内容前置校验、Markdown 渲染净化、前台正文服务端渲染

## 架构决策

| 编号 | 决策 | 理由 |
|---|---|---|
| D1 | 笔记为主、项目为索引 | 弱化产品目录观感，符合个人备案内容边界 |
| D2 | 站内不托管 exe，仅外链 GitHub | SuperClip-C 无许可证、LanFileShare 禁商业分发 |
| D3 | 单容器 + embed 静态资源 | 1Panel 部署运维对象最少 |
| D4 | 后台放 `/admin`，不分子域名 | 避免子域名额外备案 |
| D5 | SQLite 数据卷持久化 | 备份即复制单文件 |
| D6 | 界面采用第一版原型 | 用户对比两版后确认第一版更耐看、更符合站点调性 |
| D7 | 单 SPA + history 路由 + Go 回退 | `/admin` 必须是真实路径，不使用 hash 路由 |
| D8 | Markdown 服务端净化 | 前端不信任原始输入，避免存储型 XSS |
| D9 | 签名 Cookie 会话 | 单管理员场景无需会话表，改密即可失效旧会话 |
| D10 | 复用 lxsyzd WSL | 该环境已有 Go 1.27 / Node 22 / Docker 29，省去初始化；代价是与其他项目共用，存在依赖冲突风险 |
| D11 | SQLite 连接池上限 1 | 单管理员写入量极低，避免并发写锁；迁移与 PRAGMA 在单连接上稳定生效 |
| D12 | 公开 API 只返回 `published` | 草稿与隐藏内容不进入前台，接口层强制过滤 |
| D13 | 前端产物用 `embedui` 构建标签嵌入 | 默认构建不依赖前端产物，`go test ./...` 与本地 API 调试无需先构建前端 |
| D14 | 项目 `note_count` 由 SQL 子查询统计 | 项目卡片需要关联笔记数，避免前端 N 次请求详情 |
| D15 | 列表搜索由后端 `q` / `project` 过滤 | 与分页口径一致，前端只做展示与防抖 |
| D16 | 后台编辑器采用独立页面而非弹窗 | Markdown 正文长，独立页写作与刷新更稳；弹窗在移动端高度受限 |
| D17 | Markdown 在服务端渲染并净化后下发 | 前端不接触原始输入，避免存储型 XSS；预览复用同一管线 |
| D18 | 会话不落库，用签名 Cookie 携带 `admin.updated_at` | 单管理员无需会话表，改密即让旧会话失效 |
| D19 | 登录限流用进程内计数 | 单实例部署足够；重启后计数清零可接受 |

## 已知限制

- mailhub 仓库地址返回 404，GitHub 搜索无匹配，暂不纳入展示
- SuperClip-C、LanFileShare 的许可证限制分发，站点只能放源码外链
- 前台无评论、无访问统计（方案 A 范围外）
- 第二版原型仅归档，不参与正式构建；是否删除待确认
- 后台编辑器为纯文本框 + 预览，暂无工具栏、快捷键与图片上传（图片只能用 `/assets/` 站内路径）
- 登录限流计数存活于内存，多实例部署时各实例独立计数
- 文章 slug 需手填，未做中文标题自动转拼音
- Dockerfile、Compose 与 1Panel 部署尚未实现
- 6 个项目的真实 GitHub 地址、5 篇笔记的真实内容尚未最终确认
- 原型与构建复用了 `lxsyzd` WSL，未建项目专属隔离环境

## 踩坑记录

| 现象 | 根因 | 修复 | 教训 |
|---|---|---|---|
| 通过 `wsl bash -lc '... &'` 启动的静态服务随即退出 | WSL 会话结束时后台进程被回收 | 在同一命令内 `sleep 2` 并用 `curl` 验证，或使用长期运行的服务管理方式 | 后台服务启动后必须做存活校验，不能只看命令退出码 |
| PowerShell 双引号包裹 WSL/Bash 命令时变量被提前展开 | PowerShell 先解析 `$HOME`、`$P` 等变量 | 使用完整路径、避免嵌套变量，必要时转义 | 跨 PowerShell 与 Bash 的命令要按层校验引号与变量 |
| `go build -ldflags "-s -w -X main.version=..."` 被拆成多个参数 | PowerShell、Bash、Go 三层引号嵌套 | 冒烟构建先省略 ldflags，使用 `main.version` 默认值；正式脚本用单引号包裹完整 ldflags | 跨层传参时先用最小命令验证，再拼复杂参数 |
| 后台启动命令把 `cd && go build && 服务` 整段放入后台，导致 `curl` 在构建完成前执行 | `&` 作用于整个 `&&` 链 | 构建与服务分两条命令执行 | 启动冒烟前先确认二进制已生成 |
| 路由切换后立即截图，拿到上一页内容 | Vue 过渡 `mode="out-in"` 需要约 200ms | 截图前等待 1 秒 | 浏览器验收要等待渲染稳定，不能只看 URL |
| 删除原型 mock 模块后 `vite build` 报 `Could not resolve "./components/ToastMessage.vue"` | `App.vue`、`router.ts` 仍保留原型 import | 同步清理 import 与组件注册，再检查 `router` 中的占位路由 | 删除模块前先全局搜索引用，删除后立刻构建验证 |
| 生产构建后直接访问 `/notes/xxx` 静态资源 404 | `base: './'` 生成相对路径，深路径下会解析到 `/notes/assets/...` | 改为 `base: '/'`，与 history 路由配合 | 使用 history 路由时 `base` 必须是绝对根路径 |
| 无头浏览器截取异步页面得到全白图片 | `--virtual-time-budget` 下 Vue 过渡停留在 `fade-enter-from`（`opacity: 0`） | 改用 Playwright 真机等待后再截图/断言 | 截图空白的先查 DOM 与 `opacity`，不要直接判定页面渲染失败 |
| 骨架屏与真实卡片复用同一 class，导致自动化断言计数错误 | 骨架容器沿用 `.note-row` / `.project-card` | 骨架改用独立 `.skeleton-block` | 加载态与内容态用不同 class，便于样式与测试区分 |
| bluemonday 净化后站内图片 `src` 被清空 | 该库默认不允许相对 URL | 调用 `p.AllowRelativeURLs(true)`，同时用 `Matching` 限定 `^/assets/` | 净化库的默认策略要逐条验证，不能只看元素白名单 |
| Windows 侧 `127.0.0.1:18080` 连接被拒，WSL 内 `curl` 正常 | WSL2 localhost 转发在重启实例后失效 | 改用 WSL 网卡 IP（`hostname -I`）访问 | 本机验收失败时先用 WSL 内 `curl` 排除服务问题，再怀疑端口转发 |
| 后台导航指向 `/admin/notes` 但该路由未定义，点击后被重定向首页 | 路由只加了登录页，缺子路由 | 一次性补齐后台全部子路由后再联调 | 导航项与路由必须同批交付，避免出现「点了没反应」 |

## 待办清单

1. 确认 mailhub 正确仓库地址，或决定移除
2. 确认 6 个项目的真实 GitHub 地址与许可证状态
3. 确认 5 篇笔记的真实内容来源
4. Section 4：真实内容录入
5. Section 5：Docker + 1Panel 部署 + 浏览器验收
6. 每个 Section 完成后升小版本、同步文档并提交

## 版本一致性检查

当前需要保持一致的四处：

| 位置 | 当前值 |
|---|---|
| `app/VERSION` / `app/main.go` | `1.2.0` |
| `web/package.json` / `web/package-lock.json` | `1.2.0` |
| `web/src/components/SiteFooter.vue` / 后台侧栏 / 登录页 | `v1.2.0 by Mr lin` |
| `doc/*.md` / `README.md` / `CHANGELOG.md` | `v1.2.0` |

`scripts/build.sh` 在编译前校验 `app/main.go` 的 `var version` 与 `app/VERSION` 一致，不一致直接失败。原型工程 `prototype/` 仍为 `v1.0.3`，作为冻结基线不再跟随版本。

## 构建与发布命令

```bash
# 单元与接口测试（默认构建，不嵌入前端）
cd app
go test ./... -count=1

# 完整构建：前端 + 内嵌静态资源二进制
cd /mnt/e/app/网站备案/dev-notes-platform
bash scripts/build.sh          # 产物：app/build/dev-notes

# 本地开发：后端（终端 1）+ 前端（终端 2）
cd app && DATA_DIR=/tmp/dn-dev SESSION_SECRET=0123456789abcdef0123456789abcdef \
  HTTP_ADDR=127.0.0.1:8080 go run .
cd web && npm run dev          # Vite 已把 /api 代理到 127.0.0.1:8080
```

详细说明见 `doc/DEPLOY.md` 第二节与 `doc/PROJECT.md` 第七节。
