# 更新日志 · 我的开发笔记

> 版本记录 · by Mr lin

## 版本历史速览

| 版本 | 日期 | 摘要 |
|---|---|---|
| v1.3.3 | 2026-11-17 | 录入 LanFileShare 的 4 篇技术复盘，三个项目卡片的仓库地址与技术栈改为真实值 |
| v1.3.2 | 2026-11-17 | 新增 `content/notes/` 文章源与 `tools/import-notes.sh` 导入工具，录入 SuperClip-C 的 4 篇技术复盘 |
| v1.3.1 | 2026-11-17 | 修复 HTTP 访问下 Secure Cookie 被丢弃导致后台登录后立即失效 |
| v1.3.0 | 2026-11-17 | 新增站内配图目录 `/media/` 与两张技术插图，正文图片按容器等比缩放 |
| v1.2.0 | 2026-11-17 | 完成 Section 3：后台管理与认证、内容校验、Markdown 渲染净化、前台正文改为服务端渲染 |
| v1.1.0 | 2026-11-17 | 完成 Section 2：正式前台接入真实 API、history 路由、静态资源 embed 与构建脚本 |
| v1.0.3 | 2026-11-17 | 完成 Section 1：Go 骨架、SQLite、数据模型、Store、公开 API 与测试 |
| v1.0.2 | 2026-11-17 | 确认第一版界面方案，新增详细技术方案，同步文档与版本 |
| v1.0.1 | 2026-11-17 | 新增工程约定文件，完成两版界面原型 |
| v1.0.0 | 2026-11-17 | 初始设计方案 |

## 详细更新日志（按版本倒序）

### v1.3.3 · 2026-11-17

**新增**

- 录入 LanFileShare 的 4 篇技术复盘：`http-listener-to-tcplistener`（http.sys 抢端口导致端口在听但收不到请求）、`streaming-multipart-1mb-window`（1MB 滑动窗口把内存与文件大小解耦，附终止符分支死锁）、`device-name-to-safe-path`（客户端设备名的四层路径防御）、`ua-model-detection`（UA 精简后用 Client Hints 取型号）。

**修复**

- 三个项目卡片的仓库地址由占位符（`github.com/example/*`）改为真实地址：`mylxnet/site-guard`、`mylxnet/SuperClip-C`、`mylxnet/LanFileShare`。
- 技术栈纠错：SiteGuard 由 `Python/FastAPI` 改为 `Node.js/ESM/单容器自托管`；SuperClip-C 由 `Go/WPF/SQLite` 改为 `C++/Win32 API/CMake`；LanFileShare 由 `Go/HTTP/Web` 改为 `C#/WPF/.NET 8`。
- 三处项目简介改为与仓库实际定位一致（不再出现「局域网零配置发现」这类与实际实现不符的描述）。

**调整**

- 版本统一升至 `v1.3.3`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。

### v1.3.2 · 2026-11-17

**新增**

- 新增 `content/notes/`：文章以 Markdown + 简单头部字段（`title` / `slug` / `summary` / `project` / `status` / `published_at`）落仓库，正文与代码一起版本化。
- 新增 `tools/import-notes.sh`：登录后台后逐篇导入，`project` 字段按项目 slug 自动解析成 `project_id`；按 slug 幂等，已存在默认跳过，`MODE=update` 时覆盖更新。
- 录入 SuperClip-C 的 4 篇技术复盘：`win32-topmost-silently-dropped`、`sendinput-batch-alt-keyup`、`rc-codepage-resource-garbled`、`favorite-partition-invariant`，素材来自该项目 `doc/PROJECT_STATE.md` 的真实踩坑记录。

**优化**

- 文章内容不再只存在数据库里：源文件进仓库后，换机器、重建库都能一键复现，补上「Section 4 真实内容录入」的可重复路径。

**调整**

- 版本统一升至 `v1.3.2`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。

### v1.3.1 · 2026-11-17

**新增**

- `COOKIE_SECURE` 改为三档：`auto`（默认）、`true`/`always`、`false`/`never`。`auto` 按请求是否 HTTPS 决定 Secure 标记，同时识别直连 TLS 与反向代理的 `X-Forwarded-Proto: https`。
- 非法取值（如 `maybe`）在启动阶段直接失败，并提示「COOKIE_SECURE 只能是 auto / true / false」，不进入运行期才暴露。

**优化**

- 反向代理终结 TLS 的部署可以继续给会话 Cookie 保留 `Secure`，不必再用 `COOKIE_SECURE=false` 换取可用性。

**修复**

- 修复用内网 IP 或域名以 HTTP 访问后台时，`Secure` Cookie 被浏览器丢弃、登录后 `/api/admin/me` 立刻返回 401 并被弹回登录页、后台无法新增或修改数据的问题。
- 同一缺陷也解释了「本地 localhost 测试一律正常」：浏览器把 `localhost` 与 `127.0.0.1` 视为安全上下文，Secure Cookie 照样保存，问题只在换地址后暴露。

**调整**

- 新增测试：`TestSessionCookieSecureFollowsScheme`（auto 下 http/https/反代三种场景 + always/never 强制）、`TestLoadCookieSecureMode`（含非法值启动失败）。
- 版本统一升至 `v1.3.1`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。

### v1.3.0 · 2026-11-17

**新增**

- 新增站内配图目录 `web/public/media/`：构建时随 `public/` 一并进入 `dist/`，由 Go 静态文件处理直接下发，无需改路由。
- Markdown 图片白名单由 `^/assets/` 放宽为 `^/(?:assets|media)/`，笔记正文可用 `![说明](/media/xxx.svg)` 引用自绘插图；外链、协议相对地址（`//cdn.example/a.png`）与未授权目录仍被剔除。
- 新增两张技术插图：`siteguard-origin.svg`（网关反代下 Origin 与 Host 不一致导致必然 403，以及自定义请求头的解法）、`mailhub-imap-cursor.svg`（IMAP 星号范围与 UID 游标错位导致误报新邮件）。

**优化**

- 正文图片样式补 `height: auto` 与 `display: block`：SVG 自带 `width`/`height` 属性，只限宽会压扁比例，现在按容器宽度等比缩放并居中留白。
- 净化变量改名 `assetSRC` → `localImageSRC`，注释区分「`/assets/` 是前端构建产物、`/media/` 是笔记配图」两个来源。

**修复**

- 修正 mailhub 插图的差一错误：101 已存在时 `uidNext` 应为 102，原图误标为 101，会让读者把「下一个可用 UID」与「当前最大 UID」混为一谈。
- 修正该插图里守卫条件写反的表述：正确写法是 `uid > cursor`，误写成 `uid >= cursor` 才会每轮多算 1 封。
- 修复页脚版本号与署名被右下角固定的主题按钮遮挡的问题：桌面为页脚预留按钮宽度，窄屏改为底部留白。
- 修复 `TestParseRejectsTamperedSession` 偶发失败：原用例改写 Cookie 末位字符，而 base64url 末位包含未使用位，解码后字节可能不变，导致「篡改」实际无效。

**调整**

- 图片白名单测试改为表驱动：正例覆盖 `/assets/` 与 `/media/*.svg`，反例覆盖外链、协议相对地址与未授权目录。
- 版本统一升至 `v1.3.0`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。

### v1.2.0 · 2026-11-17

**新增**

- 新增 `internal/auth`：HMAC-SHA256 签名会话 Cookie（12 小时）、CSRF 双提交 token、登录失败限流（同一 IP 15 分钟内 5 次）、鉴权中间件。
- 新增 `internal/validate`：标题/摘要/正文/slug/技术栈/仓库地址/状态/密码规则，以及经营性词汇与下载类表述两级拦截。
- 新增 `internal/render`：goldmark 渲染 + bluemonday 白名单净化，禁用原始 HTML，链接限 http/https/站内，图片仅允许 `/assets/`。
- 新增管理 API：登录、退出、当前会话、修改密码、Markdown 预览、笔记与项目增删改查。
- 新增后台界面：登录页、侧栏布局、笔记管理（筛选/搜索/删除确认）、笔记编辑（含预览与字段级错误定位）、项目管理、项目编辑、账号设置。
- 前台笔记详情改为渲染服务端返回的 `content_html`，新增正文排版样式。
- 站点头部恢复「后台管理」入口；移动端自动隐藏。
- 管理端请求体上限 1 MB，写操作强制 CSRF 校验。

**优化**

- 前台详情接口不再返回 Markdown 原文，只返回净化后的 HTML。
- 项目列表与后台列表共用查询，后台笔记列表不返回正文，减少响应体积。
- 编辑已有笔记时未改发布时间会沿用原值；改为发布且留空时前端自动补当天日期。
- 修改密码后旧会话立即失效，同时自动为当前页面续期，避免被踢出。

**修复**

- 修复 bluemonday 默认不允许相对 URL 导致站内图片与相对链接被清空的问题（`AllowRelativeURLs(true)`）。
- 修复后台骨架屏复用表格类名造成的样式串扰。

**调整**

- 后台编辑器采用独立页面（`/admin/notes/:id/edit`、`/admin/projects/:id/edit`），便于长正文编辑与刷新。
- 版本统一升至 `v1.2.0`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。

### v1.1.0 · 2026-11-17

**新增**

- 新增 `web/src/api.ts`：公开内容类型、四个公开接口封装、`ApiError` 与日期格式化。
- 正式前台接入真实 API：首页、笔记列表、笔记详情、项目索引、项目详情全部改为请求后端，不再依赖原型 mock 数据。
- 笔记列表支持按项目筛选与关键词搜索，请求带 300ms 防抖与 `AbortController` 取消旧请求。
- 新增加载骨架屏、错误横幅（含重试）、空状态三态展示。
- 新增 `scripts/build.sh`：版本一致性校验 → 前端构建 → `-tags embedui` 静态编译，产物输出 `app/build/dev-notes`。
- 新增 `app/internal/webui` 包：以 `embedui` 构建标签把前端产物嵌入二进制；默认构建不嵌入，便于本地用 Vite 开发。
- 后端项目列表补充 `note_count`（已发布关联笔记数），由 SQL 子查询统计，避免前端逐个请求详情。

**优化**

- 路由由 hash 模式改为 history 模式，Go 侧 `NoRoute` 对非 `/api/` 路径回退 `index.html`（落地 D7）。
- 静态资源按 `/assets/` 前缀设置一年期不可变缓存，`index.html` 使用 `no-cache`。
- 笔记详情正文按空行拆分为段落展示，`white-space: pre-wrap` 保留原始换行。
- 项目详情源码入口改为真实外链 `repo_url`，新窗口打开并带 `noopener noreferrer`。

**修复**

- 移除原型遗留的模拟认证（`auth.ts`）、全局提示（`toast.ts`、`ToastMessage.vue`）与 mock 数据引用，消除构建期模块解析失败。
- 修正 `base: './'` 在 history 路由下导致 `/notes/:slug` 静态资源 404 的问题，改为 `base: '/'`。
- 骨架屏容器不再复用 `.note-row` / `.project-card` 类，避免与真实内容混用。

**调整**

- `web/` 由原型工程转为正式前端，包名 `dev-notes-prototype` → `dev-notes-web`。
- 后台入口暂从站点头部移除，待 Section 3 接入认证后恢复，避免当前出现死链。
- 版本统一升至 `v1.1.0`：`app/VERSION`、`app/main.go`、`web/package.json`、界面页脚与文档同步。

### v1.0.3 · 2026-11-17

**新增**

- 新增 `app/` Go 模块：配置、SQLite 迁移、项目/笔记/管理员模型、Store、公开 API、健康检查。
- 新增单元与接口测试：配置校验、已发布内容过滤、管理员种子、唯一约束、健康检查、公开 API、404。
- 新增 `app/VERSION`，后端版本号纳入版本一致性检查。

**优化**

- SQLite 使用 WAL、`busy_timeout`、外键与单连接池，降低单管理员场景的写锁风险。
- 公开 API 在 Store 层强制过滤 `published`，草稿与隐藏内容不会进入前台。
- 管理员首次启动按环境变量写入 bcrypt 哈希，已存在时跳过。

**修复**

- 修正可空项目外键的参数传递，避免驱动差异导致写入异常。
- 更新项目时避免技术栈写入 `null`。

**调整**

- 项目状态调整为“Section 1 已完成，待 Section 2”。
- 版本统一升至 `v1.0.3`：文档、原型工程与界面页脚/侧栏同步，并重新构建 `prototype/dist`。
- Git 仓库已初始化，基线提交为 `3f41179`。

### v1.0.2 · 2026-11-17

**新增**

- 新增 `doc/PROJECT.md`：总体架构、数据模型、API 契约、认证会话、内容校验、前端规范、测试方案、开发计划与风险对策。
- 新增 `CHANGELOG.md`：建立逐版本变更记录。

**优化**

- 确认第一版界面原型 `prototype/` 为正式界面基线；第二版 `prototype-v2/` 归档为候选方案，不参与构建。
- 设计方案补充界面方案确认结论与技术方案链接。
- README 补充技术方案、原型与版本变更入口。

**修复**

- 统一项目文档、原型工程与界面页脚版本号为 `v1.0.2`。

**调整**

- 项目状态由“设计阶段”调整为“设计收尾，待进入 Section 1 编码”。
- 待办清单补充真实仓库地址、真实笔记内容与 Git 仓库初始化确认。
- 项目状态文档补充今日收尾记录、验证证据与下次开工入口。

### v1.0.1 · 2026-11-17

**新增**

- 新增 `CODEBUDDY.md` 工程约定文件。
- 完成第一版与第二版界面原型，覆盖前台 6 页与后台 4 模块。

**优化**

- 前台调整为笔记驱动，弱化产品目录观感。
- 两版原型完成桌面与移动端浏览器验收。

### v1.0.0 · 2026-11-17

**新增**

- 初始设计方案 `doc/DESIGN.md`。
- 部署手册 `doc/DEPLOY.md`。
- 项目状态与交接文档 `doc/PROJECT_STATE.md`。

## 下次开发计划

1. 确认 mailhub 仓库地址与 6 个项目的真实 GitHub 地址。
2. 确认 5 篇笔记的真实内容来源。
3. Section 4：真实内容录入（6 个项目 + 5 篇笔记），并为已有的 `/media/` 插图补正文引用。
4. Section 5：Docker 镜像、1Panel 部署与线上浏览器验收。
5. 每个 Section 完成后升小版本、同步文档并提交。
