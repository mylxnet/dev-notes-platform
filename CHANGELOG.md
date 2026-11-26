# 更新日志 · 我的开发笔记

> 版本记录 · by Mr lin

## 版本历史速览

| 版本 | 日期 | 摘要 |
|---|---|---|
| v1.4.9 | 2026-11-26 | 推荐改为笔记字段：编辑页勾选「设为首页推荐」、列表加推荐徽标，首页「全部」进推荐全量列表，站点设置只留备案号 |
| v1.4.8 | 2026-11-26 | 主题按钮窄屏收成圆形不再压表单、补 title 提示；修复生产 CSP 拦掉首屏主题脚本、WSL 下 Vite 不热更新 |
| v1.4.7 | 2026-11-25 | 项目页改版：横向缩略图卡片（contain 以高为准、居中留空）、详情页图文并排，图片不再放大 |
| v1.4.6 | 2026-11-17 | 补齐 Docker 构建链：多阶段 Dockerfile、1Panel Compose、阿里云推送脚本，重写部署手册 |
| v1.4.5 | 2026-11-17 | 笔记正文与项目说明接入带工具栏、实时分栏预览和图片上传的 Markdown 编辑器 |
| v1.4.4 | 2026-11-17 | 后台项目封面支持选择图片上传，服务端安全校验后写入 DATA_DIR/media，无需重新打包 |
| v1.4.3 | 2026-11-17 | 后台「站点设置」可维护 ICP 备案号，前台页脚从公开接口读取，未填写显示占位 |
| v1.4.2 | 2026-11-17 | 测试构建可预填后台登录账号密码，正式构建不带任何测试凭据 |
| v1.4.1 | 2026-11-17 | 后台笔记/项目列表改为每页 6 条并带分页条，抽出通用分页组件供三处复用 |
| v1.4.0 | 2026-11-17 | 项目新增详情说明（Markdown）与封面图，卡片显示封面、详情页渲染说明，老库自动补列 |
| v1.3.19 | 2026-11-17 | 首页「项目索引」副标题同步改为「个人练习作品，共 N 个」 |
| v1.3.18 | 2026-11-17 | 项目状态文案改为「已完成」，删除项目页顶部免责提示条 |
| v1.3.17 | 2026-11-17 | 项目索引页副标题改为「个人练习作品，共 N 个」，去掉右侧重复计数 |
| v1.3.16 | 2026-11-17 | 首页导语删除「内容仅作个人非经营性分享…不提供任何下载」一句 |
| v1.3.15 | 2026-11-17 | 前台去掉「后台管理」入口；笔记列表改为每页 6 篇并加页码分页 |
| v1.3.14 | 2026-11-17 | 关于页宽度改为跟随 container，与项目页左右边界一致（1180px） |
| v1.3.13 | 2026-11-17 | 关于页单栏改为居中 780px，标题、导语、段落统一右边缘 |
| v1.3.12 | 2026-11-17 | 关于页删除右侧信息栏，改单栏限宽排版 |
| v1.3.11 | 2026-11-17 | 关于页重写为「我是谁 / 为什么建这个站 / 在哪里可以找到我」，版本署名移入关于页，页脚只留备案号 |
| v1.3.10 | 2026-11-17 | 站点头部标识由纯文字「N」换成折角笔记纸 + 字母 N 的矢量 logo，同步 favicon |
| v1.3.9 | 2026-11-17 | 新增台账汇总器项目与 3 篇复盘（单元格选择模式、索引器绑定排序、交付前脱敏） |
| v1.3.8 | 2026-11-17 | 新增 QuickOcr 项目与 1 篇「免费开源 OCR」成本账文章 |
| v1.3.7 | 2026-11-17 | 新增 MailHub 项目与 3 篇复盘（收件箱标记、UID 区间陷阱、可见性口径） |
| v1.3.6 | 2026-11-17 | 新增记账管理平台项目与 8 篇复盘（不可变账本、联动校验、收益口径、备份时区、静态缓存、移动端 PDF、401 分流、坏账核销） |
| v1.3.5 | 2026-11-17 | 新增 excelMailmerge 项目与 4 篇复盘；确立「同批文章不套同一骨架」的写作约定 |
| v1.3.4 | 2026-11-17 | 录入 SiteGuard 的 4 篇技术复盘，其中一篇启用 `/media/` 插图 |
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

### v1.4.9 · 2026-11-26

**新增**

- 笔记新增「设为首页推荐」字段（`note.is_featured`，`schemaVersion` 3 → 4）：后台新建/编辑页勾选即推荐，勾选后显示「已推荐（首页展示）」说明。
- 后台笔记列表在状态徽标后加「推荐」小徽标，一眼看出哪些笔记进了首页推荐。
- 首页「推荐笔记」面板的「全部」进入 `/notes?featured=1`：标题变「推荐笔记」，可继续按项目、关键词筛选，也能点「仅看推荐 ✕」退回全部笔记；筛选状态写进 URL，刷新或分享后仍是同一视图。

**优化**

- 推荐不再依赖站点设置：首页取「已发布 + 已推荐」按发布时间倒序的前 3 条，推荐篇数不设上限，多的去全量列表看。
- 「站点设置」精简为只维护 ICP 备案号，并给出「去笔记管理勾选推荐」的引导链接。

**修复**

- 迁移旧数据：`settings.featured_notes` 里已配置的推荐位会自动回填为 `is_featured = 1`，此前选的 3 篇不会丢；旧值缺失或损坏时按未配置处理，不阻塞启动。

**调整**

- `GET /api/notes` 新增 `featured=1` 筛选；`/api/site` 的 `featured_notes` 改为查笔记字段；`PUT /api/admin/settings` 只接受 `icp_number`，传 `featured_slugs` 会被忽略。
- 版本升至 `v1.4.9`（`app/VERSION`、`app/main.go`、`web/package.json`、`web/package-lock.json`、`web/src/version.ts` 五处 + 文档）。

**验收证据（2026-11-26，无头 Chromium）**

| 用例 | 实际结果 |
|---|---|
| v3 库迁移 | `user_version` = 4，旧推荐位回填 3 条（库内与接口一致） |
| 后台勾选推荐 | 开关初始未勾 → 勾选保存 → 推荐数 3 → 4，列表徽标 3 → 4 |
| 首页面板 | 3 条，均为推荐集合内笔记 |
| 点「全部」 | 跳到 `/notes?featured=1`，标题「推荐笔记」，卡片 4 张、文案「共 4 篇」 |
| 取消筛选 | 地址回 `/notes`，标题回「笔记列表」 |
| 站点设置页 | `select` 数量 0，只剩「ICP 备案号」 |
| 取消推荐 | 推荐数回到 3，首页与推荐列表同步移除 |
| 控制台 | 0 报错 |

### v1.4.8 · 2026-11-26

**新增**

- 后台「新建笔记 / 新建项目」支持按标题（名称）生成链接标识：中文转无声调拼音（`pinyin-pro`，动态加载，不进前台首屏包）、全角字母数字转半角、超 80 字符回退到最后一个连字符处截断。
- 生成的标识与已有内容重名时自动追加 `-2`、`-3`（已用标识由后台列表接口一次取 100 条，取不到时由数据库唯一约束返回 409 兜底）。
- 「用标题生成 / 用名称生成」按钮只在新建时出现；编辑已有内容不自动改写标识，避免已发布链接失效。离开标题（名称）框时补一次生成，仅在标识仍为空时触发。
- 首页右上角面板由「最近更新」改为「推荐笔记」：内容由后台「站点设置」的三个推荐位维护，存 `settings` 表的 `featured_notes` 键（slug 数组）；`/api/site` 校验「存在且已发布」后按顺序下发，最多 3 条，全空时首页隐藏该面板、hero 转单列，不再与下方「最近笔记」重复。
- 主题支持「跟随系统 + 手动」双功能：浮动按钮点击循环「跟随系统 → 浅色 → 深色」，默认跟随系统并监听 `prefers-color-scheme` 实时切换；手动选择后记住并停止跟随，可再切回跟随系统。旧键 `dev-notes-theme` 自动迁移到新键 `dev-notes-theme-mode`。
- 浮动主题按钮补 `title` 提示（「当前为浅色，点击切换」），悬停可看出当前模式与点击结果。

**优化**

- 窄屏（≤640px）主题按钮收成 44×44 圆形按钮，只保留状态圆点，不再显示文字，避免压住右下角的表单控件；桌面端仍显示「浅色主题」长文案。
- 按钮内部去掉 flex `gap`，改为圆点单独留 8px 边距，桌面端文字连写为「浅色主题」，不再显示成「浅色 主题」。
- 后台内容区在窄屏增加 88px 底部留白，滚到底时「保存」等末行按钮不会被固定主题按钮覆盖；页脚为按钮预留的右内边距同步从 132px 收到 76px。

**修复**

- 生产构建首屏主题初始化失效：`index.html` 的内联主题脚本被 CSP `script-src 'self'` 拦掉（控制台报「Executing inline script violates…」），首屏只能等 Vue 挂载后才切深色，出现浅色闪一下。改为独立文件 `web/public/theme-init.js`，用 `<script src>` 同步加载，既符合 CSP 又保持首屏前生效。
- WSL 下改代码不生效：源码挂在 `/mnt/e`（Windows drvfs），inotify 事件传不到 WSL，Vite 仍返回旧编译结果（表现为改了模板但页面无变化、`title` 取不到）。在 `vite.config.ts` 开 `server.watch.usePolling`（间隔 300ms）解决。
- 窄屏（≤640px）后台页头与按钮文字被压成竖排：`.admin-heading` 在 640px 以下改为纵向堆叠，`.btn` 加 `white-space: nowrap`，`.row-actions` 允许换行。

**调整**

- 版本升至 `v1.4.8`（`app/VERSION`、`app/main.go`、`web/package.json`、`web/package-lock.json`、`web/src/version.ts` 五处 + 文档同步）。
- `deploy/docker-compose.yml` 镜像 tag 跟到 `1.4.8`；`1.4.8` 与 `latest` 已于 2026-11-26 推送阿里云（`scripts/push-image.sh`），容器冒烟 `Up (healthy)`。

**相关文件**

- 新增 `web/src/slug.ts`、`web/public/theme-init.js`；改动 `AdminNoteEditView.vue`、`AdminProjectEditView.vue`、`ThemeSwitch.vue`、`styles.css`、`index.html`、`vite.config.ts`、`web/package.json`（新增 `pinyin-pro@^3.29.5` 依赖）。
- 推荐笔记：后端 `app/internal/api/site.go`；前端 `HomeView.vue`、`AdminSettingsView.vue`、`api.ts`（`FeaturedNote` / `AdminSettings` 类型）、`admin.ts`、`site.ts`。
- 交互原型 `prototype/slug-auto.html`、`prototype/featured-theme.html`（均离线可用），截图在 `prototype/screenshots/`。

**验收证据（2026-11-26，无头 Chromium）**

| 用例 | 实际结果 |
|---|---|
| 390px 按钮尺寸 / 圆角 | 44×44 / `50%` |
| 390px 按钮文本 | 空（只有圆点），圆点水平居中偏差 0px |
| 390px `title` | 「当前为跟随系统，点击切换」，点一次变「当前为浅色，点击切换」 |
| 390px 点击后主题 | `data-theme="light"` |
| 390px 七个页面滚到底遮挡（公开 3 页 + 后台 4 页） | 全部 `[]`，无控件被压；横向溢出 0px |
| 1280px 按钮 | 100×40，长文案「浅色主题」可见，`title` 正常 |
| 生产产物 CSP | 控制台报错 0（修复前 5 条内联脚本违规） |
| 拦截业务 JS 后首屏主题 | `data-theme="dark"`，证明 `theme-init.js` 在 Vue 之前已生效 |
| 二进制冒烟 | `/healthz` 返回 `{"db":"ok","status":"ok","version":"1.4.8"}` |
| 镜像发包 | `1.4.8` / `latest` 已推送阿里云；容器 `Up (healthy)`，`/healthz` 返回 `1.4.8`，首页与 `/theme-init.js` 均 `200` |

### v1.4.7 · 2026-11-25

**新增**

- 项目索引与首页项目卡片改为「左缩略图 + 右信息」横向布局：缩略图固定 208×156（4:3），`object-fit: contain` 以高为准、水平居中、两边留白；没有封面的项目显示首字母占位块，尺寸与有图卡片一致，整行卡片高度统一。
- 项目详情页截图支持点击放大：全屏灯箱 + 模糊遮罩、右上角关闭按钮，支持 Esc 与点击遮罩关闭，打开时锁定页面滚动；放大视图按原图尺寸上限显示，不放大原图。

**优化**

- 项目详情页改为「左说明 + 右截图」并排：截图限高 520px、按比例居中、不放大不裁切，外套窗口栏；关联笔记移到下方独立区块。
- 卡片摘要限两行、底部「关联笔记数 / 查看详情」沉底对齐；窄屏（≤640px）卡片自动改为上下排列。
- 清理不再使用的 `.project-cover`、`.project-hero`、`.about-grid` 样式。

**修复**

- 详情页封面原为 `width: 100%`，1030px 宽的截图在 1180px 容器里被放大到 1180px，又糊又占满一屏；改为限高 520px、宽度按比例自适应、居中显示。

**调整**

- `deploy/docker-compose.yml` 镜像 tag 跟到 `1.4.7`，并注明「tag 必须先用 `scripts/push-image.sh` 推送」；`doc/DEPLOY.md` 升到 v1.4.7，二进制路径与健康检查示例版本号同步。
- `doc/PROJECT_STATE.md` 验收命令由 `/tmp/dn-dev` 改为 `deploy/data` 副本（WSL 重启清空 `/tmp`，旧路径不可靠）。
- 发包：`1.4.7` 与 `latest` 镜像已推送阿里云（`application/vnd.oci.image.manifest.v1+json`，88.1 MB），云端拉取后容器 `Up (healthy)`、`/healthz` 返回 `1.4.7`。
- 版本升至 `v1.4.7`（版本字段与文档同步）。

### v1.4.6 · 2026-11-17

**新增**

- `deploy/Dockerfile`：多阶段构建（Node 22 编前端、Go 1.27 编后端、alpine 3.20 运行），按 `TARGETOS/TARGETARCH` 在目标架构内编译，非 root 运行、带 healthcheck；依赖源默认走 npmmirror / goproxy.cn。
- `deploy/docker-compose.yml`：1Panel 可直接粘贴，镜像 `registry.cn-hangzhou.aliyuncs.com/mylxnet/dev-notes:1.4.6`，端口只绑 `127.0.0.1:8080`，数据卷 `./data:/app/data`，必填变量用 `${VAR:?}` 防弱密码上线。
- `scripts/push-image.sh`：构建并推送 amd64 镜像，版本取自 `app/VERSION`，支持 `--tag` 覆盖，复用已登录的 Docker 会话，不保存凭据。
- `deploy/.env.example` 与根 `.dockerignore`：提供变量模板并缩小构建上下文。

**优化**

- `doc/DEPLOY.md` 从 v1.3.5 重写到 v1.4.6：补齐二进制 + systemd 与 Docker + 1Panel Compose 两条路线、环境变量表、备份/升级/回滚、备案合规步骤与故障排查；修掉「引用不存在的 Compose 文件」的过期内容。

**修复**

- 推送阿里云报 `unknown manifest class for application/vnd.oci.empty.v1+json`：Docker 29 默认 BuildKit 附带 OCI 空 manifest，ACR 拒收；改用 `DOCKER_BUILDKIT=0` 走 legacy builder，Dockerfile 同步固定 amd64 编译（legacy builder 不识别 `BUILDPLATFORM`/`TARGETARCH` 自动变量）。

**调整**

- 只推 amd64：推送脚本用 legacy builder + `docker push`，产物为不含 attestation 的 OCI image manifest；已验证云端拉取、容器启动与 `/healthz` 返回 `1.4.6` 全通过。
- 版本升至 `v1.4.6`（5 处版本字段 + 文档）。

### v1.4.5 · 2026-11-17

**新增**

- 新增通用组件 `web/src/components/MarkdownEditor.vue`：12 个工具栏按钮（H2/H3、粗体、斜体、引用、列表、链接、图片、行内代码、代码块、分隔线）+ 分栏实时预览 + 图片上传，已接入笔记正文与项目说明两处编辑页。
- 「图片」按钮复用 `POST /api/admin/media/upload`：前端校验扩展名与 2 MB 上限，上传后自动插入 `![图注](/media/uploads/…)` 并选中 alt 文字便于修改。
- 预览复用服务端 `POST /api/admin/preview`，400ms 防抖 + 请求序号防竞态；支持「分栏 / 仅编辑 / 仅预览」三种模式，窄屏自动上下排列。
- 新增界面原型 `prototype/md-editor.html`，先在浏览器确认交互再落地代码。

**优化**

- 两个编辑页移除原来的「预览 / 预览说明」按钮与独立预览区，避免重复入口；正文仍以 Markdown 原文保存，已有文章与服务端渲染规则不变。
- 窄屏（≤640px）编辑页三列表单改为单列并允许输入控件收缩，消除日期控件的横向溢出。

**修复**

- 无。

**调整**

- 版本升至 `v1.4.5`（5 处版本字段 + 文档）。

### v1.4.4 · 2026-11-17

**新增**

- 后台项目编辑页封面框新增「选择图片上传」：选择文件后自动上传并填入站内路径，带上传中状态、缩略图与错误提示；上传后无需重新打包。
- 新增 `POST /api/admin/media/upload`（鉴权 + CSRF）：只收 PNG/JPG/WEBP，单张 ≤ 2 MB，随机 32 位十六进制文件名，写入 `DATA_DIR/media`。
- 新增 `GET /media/uploads/:name`：运行时下发上传的图片，文件名严格白名单校验、长缓存；构建产物里的 `/media/` 仍由静态回退处理。
- `requestJSON` 支持 `FormData`，由浏览器生成 multipart 边界，不手写 `Content-Type`。

**优化**

- 上传文件先做扩展名与内容嗅探双重校验（`http.DetectContentType`），改名伪装的非图片会被拒绝；先写 `.part` 再原子改名，避免半成品被访问。

**修复**

- 无。

**调整**

- `api.New` 增加 `mediaDir` 参数，由 `cfg.DataDir + "/media"` 传入；上传路由单独放行 3 MB 请求体（鉴权与 CSRF 不变），其余管理接口仍是 1 MB。
- 版本升至 `v1.4.4`（5 处版本字段 + 文档）。

### v1.4.3 · 2026-11-17

**新增**

- 数据库新增 `settings` 键值表（键 `icp_number`），`schemaVersion` 升到 3；新库和 v1/v2 老库都走同一条迁移路径。
- 公开接口 `GET /api/site` 返回页脚备案号；后台接口 `GET/PUT /api/admin/settings` 读取与保存，PUT 走会话鉴权 + CSRF。
- 后台侧栏新增「站点设置」页：单输入框保存备案号，支持清空恢复占位；前端即时校验与服务端同规则。
- 新增 `web/src/site.ts` 共享站点信息：整站只拉一次公开接口，后台保存后主动刷新，页脚即时更新。

**优化**

- `SiteFooter.vue` 不再写死「甘ICP备xxxxxxxx号（备案通过后显示）」；有备案号时显示备案号，留空时继续显示占位文案。

**修复**

- 无。

**调整**

- 备案号允许留空，填写时限 5–100 位汉字、字母、数字或连字符（如 `甘ICP备12345678号-1`），禁止标签与脚本字符，非法值返回 422 并指明 `icp_number` 字段。
- 版本升至 `v1.4.3`（5 处版本字段 + 文档）。

### v1.4.2 · 2026-11-17

**新增**

- 构建开关 `VITE_TEST_LOGIN=1`：仅测试包在后台登录页预填账号与密码，测试时打开 `/admin/login` 直接点登录即可，不必每次输入。
- 新增 `scripts/build-test.sh`：固化测试构建变量（账号、密码可用同名环境变量覆盖）并调用正式构建脚本；生成的产物明确标注「禁止对外发布」。
- 新增 `web/src/vite-env.d.ts`：补齐 `VITE_TEST_LOGIN` / `VITE_TEST_USERNAME` / `VITE_TEST_PASSWORD` 的类型声明。

**优化**

- `scripts/build.sh` 增加防误用校验：未带 `DN_TEST_BUILD=1` 却设置了 `VITE_TEST_LOGIN=1` 时直接退出，避免测试凭据被误打进正式包。

**修复**

- 无。

**调整**

- 测试凭据通过构建变量注入，不硬编码在组件里；正式构建中这些变量未定义，前端产物不含明文密码。
- 版本升至 `v1.4.2`（5 处版本字段 + 文档）。

### v1.4.1 · 2026-11-17

**新增**

- 后台「笔记管理」列表改为每页 6 条并带分页条，此前是一次性拉 50 条、无翻页控件。
- 后台「项目管理」列表同样每页 6 条并带分页条；`GET /api/admin/projects` 新增 `page` / `page_size` 参数（默认 20、上限 100），响应带上 `total` / `page` / `page_size`。
- 新增通用组件 `web/src/components/Pager.vue`：页码折叠规则（7 页以内全列，更多时首尾 + 当前页左右各一页，其余省略号）与按钮结构集中在一处，前台笔记列表、后台笔记/项目列表共用。

**优化**

- 后台翻页与前台一致：翻页回到页面顶部；切换状态筛选、修改搜索词自动回第 1 页；筛选/搜索/删除让总页数变少时自动退到最后一页。
- 后台只有一页时不显示翻页按钮，改用一行「共 N 篇 / 个」补上总数；多页时总数由分页条承担，不再重复显示两处。
- 窄屏（≤600px）后台分页条与前台一致：隐藏页码按钮，只留上一页 / 下一页与页码文字。

**修复**

- 无。

**调整**

- `store.ListAllProjects` 签名改为接收 `AdminProjectFilter{Limit, Offset}`，新增 `CountAllProjects`，与笔记列表的 `ListAllNotes` / `CountAllNotes` 保持同一套分页写法。
- 版本升至 `v1.4.1`（`app/VERSION`、`app/main.go`、`web/package.json`、`package-lock.json`、`web/src/version.ts`、文档）。

### v1.4.0 · 2026-11-17

**新增**

- 项目新增两个字段：`description`（详情说明，Markdown 原文，上限 20 KB）与 `cover_image`（封面图站内路径，空串表示无封面）。
- 数据库升级到 `schemaVersion = 2`：迁移用 `PRAGMA table_info(project)` 探测列，缺列才执行 `ALTER TABLE project ADD COLUMN`，老库自动补列、中断重跑不报错，已有数据原样保留。
- 后台项目编辑页新增「封面图（站内路径）」输入（带缩略图预览）与「项目说明（Markdown）」编辑区，附字节计数与「预览说明」按钮，预览走与服务端一致的渲染净化管线。
- 项目详情接口返回渲染净化后的 `description_html`；项目索引卡片在填了封面时显示封面图。

**优化**

- 公开列表接口把 `description` 置空，长文说明不随列表下发，减少响应体积；`cover_image` 保留供卡片使用。
- 卡片封面用 `aspect-ratio: 16/9` + `object-fit: contain`，同一行卡片高度齐平且不裁掉示意图上的文字；详情页封面按原始比例展示。
- 详情页说明内的 `h2/h3` 字号下调到 1.15rem / 1rem，避免与「项目说明」这一节标题层级倒挂。

**修复**

- 无。

**调整**

- 封面图路径只允许站内图片：必须匹配 `^/(assets|media)/…\.(png|jpe?g|webp|svg|gif)$` 且不含 `..`、`//`，前后端同时拦截，避免外链图与目录穿越。
- 项目说明纳入备案边界词扫描（顺序为 `title → name → summary → content → description`），命中返回 422 并指明 `description` 字段。
- 版本升至 `v1.4.0`（`app/VERSION`、`app/main.go`、`web/package.json`、`package-lock.json`、`web/src/version.ts`、文档）。

### v1.3.19 · 2026-11-17

**调整**

- 首页「项目索引」区块副标题由「个人练习作品，仅作记录与源码外链」改为「个人练习作品，共 N 个」，与 `/projects` 页保持一致。
- 首页只展示前 2 个项目，因此数字取 `fetchProjects()` 返回的 `total`（新增 `projectTotal`），不用 `indexProjects.length`，否则会显示成「共 2 个」。
- 首页项目卡片的状态标签也改用 `projectStatusText`（显示「已完成」），此前仍写死「已发布」。
- 版本升至 `v1.3.19`。

**修复**

- 补上 v1.3.17 遗漏的位置：上一轮只改了 `/projects` 页，首页同一处文案没跟着改。

### v1.3.18 · 2026-11-17

**调整**

- 项目状态文案改为「已完成」：新增 `projectStatusText`（published → 已完成 / draft → 草稿 / hidden → 隐藏），项目索引卡片与项目详情页统一使用；笔记侧仍用 `statusText`（已发布 / 草稿 / 隐藏），不受影响。
- 项目索引卡片的状态标签不再写死 `status-published` + 文本，改为按 `project.status` 动态取类名与文案。
- 删除项目索引页顶部的提示条「本站不托管安装包或可执行文件。受许可证约束，项目入口仅链接到 GitHub 源码仓库。」；`.callout` 样式保留，后台多处仍在用。
- 版本升至 `v1.3.18`。

**修复**

- 无。

### v1.3.17 · 2026-11-17

**调整**

- 项目索引页副标题由「个人练习作品记录 · 仅提供源码外链」改为「个人练习作品，共 {{ projects.length }} 个」，数量随已公开项目实时变化。
- 删除标题右侧原有的「已公开 N 个」——数量已并入副标题，保留会出现两处同样的数字。
- 版本升至 `v1.3.17`（`app/VERSION`、`app/main.go`、`web/package.json`/lock、`web/src/version.ts`、文档）。

**修复**

- 无。

### v1.3.16 · 2026-11-17

**调整**

- 首页导语删除「内容仅作个人非经营性分享，项目只提供源码入口，不提供任何下载。」一句，只保留「这里记录我在学习软件开发过程中的技术笔记、问题排查与项目复盘。」。
- 版本升至 `v1.3.16`（`app/VERSION`、`app/main.go`、`web/package.json`/lock、`web/src/version.ts`、文档）。

**修复**

- 无。

### v1.3.15 · 2026-11-17

**新增**

- 笔记列表分页：每页固定 6 篇，底部新增页码条（上一页 / 页码 / 下一页 + 「第 X / Y 页 · 共 N 篇」）。7 页以内列全部页码，超过时首尾 + 当前页左右各一页，其余折成省略号。
- 切页后自动回到页面顶部；切换项目筛选或改搜索词自动回到第 1 页；若筛选后总页数变少导致当前页越界，自动退到最后一页重新取数。

**调整**

- 前台页头去掉「后台管理」按钮，访客侧不再暴露后台入口，仅保留 首页 / 笔记 / 项目 / 关于 四个导航项；后台改为直接访问 `/admin`。
- `NotesView` 的 `PAGE_SIZE` 由 50 改为 6，分页交给后端 `page` 参数。
- 删除随之无用的 `.header-actions` 样式与窄屏隐藏规则；`README.md` 的「进入后台」一行同步改为「直接访问 `/admin`」。
- 窄屏（≤600px）页码按钮隐藏，只留上一页 / 下一页与页码文字，避免 8 个按钮挤成两行。
- 版本升至 `v1.3.15`。

**修复**

- 无。

### v1.3.14 · 2026-11-17

**调整**

- 关于页宽度去掉单独限宽，改为跟随 `.container`，与项目页同宽：1440px 视口下两边都是 `left 130 / right 1310 / 宽 1180`。
- 单栏正文随之变宽（原来 780px 居中），段落 `font-size` 提到 `1.02rem`、`line-height` 提到 `1.95`、段间距 16px，缓解长行阅读压力。
- 版本升至 `v1.3.14`（`app/VERSION`、`app/main.go`、`web/package.json`/lock、`web/src/version.ts`、文档）。

**修复**

- 无。

### v1.3.13 · 2026-11-17

**优化**

- 关于页单栏由「左对齐 720px」改为「居中 780px」（`margin-inline: auto`）：宽屏下原来右侧空出 400px 以上，视觉失衡；居中后左右留白相等（1440px 视口实测各 330px）。
- 正文宽度统一：去掉 `.about-prose p` 的 `max-width: 66ch` 与 `.about-prose .lead` 的 `62ch`，标题、导语、段落右边缘对齐到同一条线（实测三者均为 780px）。
- 窄屏自适应不受影响：390px 视口下容器为 358px，左右各 16px，无横向滚动。

**修复**

- 无。

### v1.3.12 · 2026-11-17

**调整**

- 关于页删除右侧信息栏（内容范围 / 项目 / 更新节奏 / 后台入口 / 站点性质 / 备案信息六项），只保留三节正文与末尾版本行。
- 版式由两栏改为单栏：`AboutView` 不再套 `.about-grid`，改用 `.about-single`（`max-width: 720px`），正文行长保持在可读区间，长句不会拉满整屏。
- `.about-grid` 与 `.definition` 样式保留——`ProjectDetailView` 仍在用同一套两栏骨架。
- 版本统一升至 `v1.3.12`（`app/VERSION`、`app/main.go`、`web/package.json`/lock、`web/src/version.ts`、文档）。

**修复**

- 无。

### v1.3.11 · 2026-11-17

**新增**

- 新增 `web/src/version.ts`：`APP_VERSION` / `APP_AUTHOR` / `VERSION_LINE` 作为前端唯一版本来源，关于页、后台侧栏、登录页都从这里取，避免三处手写字符串各自跑偏。
- 关于页重写（`AboutView.vue`），结构参照常见个人技术博客的「关于」页：`我是谁？` / `为什么建这个站？` / `在哪里可以找到我？` 三节正文，右侧一栏站点信息（内容范围、项目、更新节奏、后台入口、站点性质、备案信息）。
- 关于页正文末尾新增版本署名行 `v1.3.11 by Mr lin`，带分割线，与正文区分。

**调整**

- 页脚去掉「个人非经营性技术记录 · 源码仅外链 GitHub」，只保留备案号一行；版本与署名按本轮要求移到关于页。
- 后台侧栏与后台登录页的版本行改为读 `VERSION_LINE`，不再写死字符串。

**修复**

- `.definition` 增加 `align-content: start`：关于页正文变长后，右侧信息栏的网格行被拉伸，条目之间出现大片空白。

### v1.3.10 · 2026-11-17

**新增**

- 新增组件 `web/src/components/BrandLogo.vue`：站点头部与后台侧栏的品牌标识，由原来的纯文字 `<span class="brand-mark">N</span>` 改为内联 SVG。
- 图形为「折角笔记纸 + 字母 N」：外框沿用 `--accent` 圆角方块（暗色主题自动跟随 `#e08a6c`），纸面暖白 `#fffaf3`，右上折角用 `color-mix(in srgb, var(--accent) 68%, #000)` 压深，字母 N 用 accent 描边。保留原「N」的识别度，同时补上「一页笔记」的语义。
- `web/public/favicon.svg` 换成同一套几何（64×64 版本），浏览器标签图标与站内标识一致。

**调整**

- `.brand-mark` 改为承载 SVG 的容器：去掉文字字号/字重/文字色，加 `flex: none` 与 `overflow: hidden`，SVG 按 100% 填充；尺寸仍为 32×32、圆角 10px，头部与后台侧栏布局不变。
- 版本统一升至 `v1.3.10`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。

**修复**

- 无。

### v1.3.9 · 2026-11-17

**新增**

- 新增第 8 个项目 `exceltohub`（台账汇总器 / LedgerHub，C# / .NET 8 / WPF / ClosedXML），源码外链 `https://github.com/mylxnet/ExcelToHub`。仓库未声明许可证，本站只放外链。
- 新增 3 篇文章，素材取自该仓库 `README.md`、`doc/PROJECT_STATE.md` 的踩坑 1~5 与 v1.1.6 脱敏记录：
  - `wpf-cell-selection-kills-row-apis`（`SelectionUnit="Cell"` 之后，删行删不掉、右键未选中行崩溃、点同一格不取源，三个症状一个前提；附 UI 自动化测 WPF 右键菜单的四个必要条件）
  - `datatable-indexer-binding-sort-crash`（`{Binding [列名]}` 索引器绑定 + 点表头 = `ArgumentException`，从异常文本反推到 `SortDescription` 机制）
  - `desensitize-proof-by-unzip`（脱敏验收物不是 diff 而是产物：解包 `xl/sharedStrings.xml` 双向核对，`test/` 目录移出仓库的遗留问题）

**调整**

- 3 篇分别采用「一个前提三处塌方」「异常文本反推」「证据链式验收」三种骨架，与既有 28 篇不重复。
- 版本统一升至 `v1.3.9`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。

**修复**

- 无代码改动。

### v1.3.8 · 2026-11-17

**新增**

- 新增第 7 个项目 `quickocr`（QuickOcr，C# / .NET 8 / WPF / Tesseract 5，MIT），源码外链 `https://github.com/mylxnet/QuickOcr`。
- 新增 1 篇文章 `free-ocr-ledger`（免费的 OCR 能用：钱省下来了，代价挪到了别处）：以「记账」骨架拆解免费开源的四个来源（代码 MIT / 引擎 Apache 2.0 / 语言包 tessdata_fast / self-contained 运行时）与四项支出（226 MB 体积、Tesseract 精度天花板、首次联网拉语言包、无测试无 doc 的维护面），并给出「三段许可是否干净 / 失败有没有出口 / 边界有没有写进文档」三条判断标准。

**调整**

- 版本统一升至 `v1.3.8`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。

**修复**

- 无代码改动。

### v1.3.7 · 2026-11-17

**新增**

- 新增第 6 个项目 `mailhub`（MailHub，NestJS / React / TypeScript / SQLite / Prisma）：自托管多账号邮件客户端，增量同步与后台定时收取。该仓库本地存在但 GitHub 未公开，项目卡不挂外链，页面按「源码入口暂未提供」渲染。
- 新增 3 篇文章，素材取自该仓库 `doc/` 的方案与踩坑记录：`imap-inbox-flag-not-standard`（`\Inbox` 不在 IMAP 标准属性里，后台轮询只扫到「已发送」）、`imap-uid-range-star-trap`（`uidNext:*` 的 `*` 是最大 UID，守卫差一格导致每轮误报新邮件）、`visibility-scope-single-source`（「什么算收件」的口径在列表、计数、清理三处各写一份，含 Prisma `contains` 走 LIKE 的真库对照）。

**调整**

- 3 篇沿用「同批文章不套同一骨架」的约定，分别写成「排查时间线」「单行代码解剖」「真库谓词对照」三种结构，均不含表格堆砌式的小结。
- 版本统一升至 `v1.3.7`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。
- 开发库重建：上一轮的临时库位于 WSL `/tmp/dn-dev`，随环境重启丢失；本轮按 `content/notes/` 全量重导（6 个项目 + 27 篇），早期 3 篇占位笔记不再重建。

**修复**

- 无代码改动。

### v1.3.6 · 2026-11-17

**新增**

- 新增第 5 个项目 `jitijingjiguanli`（记账管理平台，Go / SQLite / Vue 3 / Vant），源码外链 `https://github.com/mylxnet/jitijingjiguanli`。
- 新增 8 篇文章：`immutable-ledger-void-only`（流水只能作废）、`linkage-rules-verification`（联动规则逐条核对抓出 6 个 bug）、`reinvest-income-pollutes-principal`（再投资收益污染本金，三种口径划分）、`vacuum-into-backup-and-timezone`（VACUUM INTO 快照 + 时区导致备份晚 8 小时）、`stale-chunk-html-as-js`（旧 chunk 回退成 HTML，200 但 MIME 拒绝）、`mobile-pdf-preview-pdfjs`（iframe 塞 PDF 移动端白屏 → pdf.js + rAF 坑）、`401-on-public-auth-paths`（公开认证接口的 401 被当成会话失效）、`bad-debt-writeoff-design`（坏账核销：只清应收不动账）。

**修复**

- 第 8 篇初版因正文出现「下载链接」被服务端备案词校验拦截（`HTTP 422`），改写为「下载按钮的 href」后通过；导入工具如实回显了服务端错误信息，未静默失败。

**调整**

- 8 篇结构各不相同：决策对比、核对清单、取证推演、SQL 方案、链路取证、能力矩阵、问题清单、约束推导。
- 版本统一升至 `v1.3.6`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。

### v1.3.5 · 2026-11-17

**新增**

- 新增第 4 个项目 `excelMailmerge`（C# / .NET 8 / WPF / ClosedXML，MIT），源码外链 `https://github.com/mylxnet/excelMailmerge`。
- 新增 4 篇文章：`wpf-invariant-globalization-crash`、`excel-row-probe-infinite-loop`、`wpf-progress-binding-and-ui-thread`、`single-instance-mutex-after-kill`，素材来自该仓库 `PROJECT_STATE.md` 的 BUG-01/02/05/09/10/11/12/13。

**调整**

- **写作约定：同批文章不再共用同一个骨架**。这 4 篇分别采用「打包参数对照」「时间线叙事」「症状清单」与「时序推演」四种结构，避免整页读起来像同一份模板填空；约定已写入 `doc/PROJECT.md` 第十二节。
- 版本统一升至 `v1.3.5`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。

### v1.3.4 · 2026-11-17

**新增**

- 录入 SiteGuard 的 4 篇技术复盘：`fnos-fpk-install-rejected`（装包报错只有一句话，用裸模板单变量二分锁定两个 wizard 文件）、`gateway-base-href-missing-assets`（网关入口无尾斜杠导致相对路径 404，服务端注入 `<base>` 修复）、`reverse-proxy-origin-403`（反代改写 Host 使写操作全被判跨站，改用自定义请求头）、`baseline-filter-asymmetry`（忽略规则只过滤当前文本、基线未过滤导致规则失效）。
- `reverse-proxy-origin-403` 正文首次启用 `/media/` 插图：`![网关反代下 Origin 与 Host 对不上](/media/siteguard-origin.svg)`，验证 v1.3.0 放开的配图通道端到端可用。

**调整**

- 三个项目内容量趋于均衡：SiteGuard 5 篇、SuperClip-C 5 篇、LanFileShare 5 篇（含各自早期 1 篇占位）。
- 版本统一升至 `v1.3.4`：`app/VERSION`、`app/main.go`、`web/package.json`/lock、页脚、后台侧栏与登录页、文档同步。

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
