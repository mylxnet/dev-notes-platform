# 项目状态 · 我的开发笔记

> v1.4.11 · by Mr lin · 2026-11-27

## 当前版本

v1.4.11 · Section 3 已完成，Section 4 进行中：8 个项目共 31 篇，全部为真实技术复盘。数据库结构版本 `schemaVersion = 4`（`note` 含 `is_featured` 推荐字段）。Section 5 部署：Docker 构建链就绪，`1.4.8` / `1.4.9` / `1.4.10` / `1.4.11` 镜像均已推送阿里云；1Panel 上机待 ICP 备案通过。

**当前 `app/build/dev-notes` 是 `bash scripts/build.sh` 产出的正式包**（v1.4.11，明文凭据计数 0，可对外发）：需要一键登录的本机验收包用 `scripts/build-test.sh` 单独产出（计数 1），用完覆盖回正式包，测试包不得外发。本机验收数据用 `deploy/data` 的副本（WSL 重启会清空 `/tmp`，原来的 `/tmp/dn-dev` 已不可靠）。

## 本轮改动记录（v1.4.11，2026-11-27）

- 后台接入「数据备份」页面：新增 `web/src/views/admin/AdminBackupView.vue`、`/admin/backup` 路由与侧栏「数据备份」入口；`admin.ts` 增加概览、二进制下载、预检、恢复与快照回滚封装，`api.ts` 导出 `baseUrl` 供二进制下载复用。
- 页面能力：统计卡（库大小、内容数、schema、图片占用、快照份数）、生成并下载 `.tar.gz`、拖拽上传自动预检、输入 `RESTORE` 才可恢复、恢复前快照列表与一键回滚、恢复后自检结果。
- 交互保护：打包/预检/恢复期间显示不确定进度动画并禁用冲突按钮；回滚前二次确认；恢复完成后刷新概览与快照列表。
- 修复上一版只做接口没做界面、后台点不到功能的问题（原 `prototype/backup.html` 仅原型）。
- 修复窄屏（≤700px）快照表被压成逐字换行：隐藏时间列（时间已在文件名中），「回滚」按钮恢复 64×46，横向溢出 0。
- 新增回归脚本 `test/backup_page_check.py`（完整链路）与 `test/backup_narrow_check.py`（390px 溢出与按钮尺寸）。

### 验收证据（v1.4.11，2026-11-27）

- `test/backup_page_check.py`：侧栏进入 `/admin/backup`，统计卡 3 张，标题「数据备份」；点「生成备份包」真实下载 `dev-notes-backup-20261126-1652.tar.gz`（221,446 字节）；重传该包显示「预检通过，可恢复」；输入 `RESTORE` 后显示「恢复完成 · 31 篇笔记 · 8 个项目 · 3 张图片 · schema v4」与快照 `pre-restore-20261126-165236.tar.gz`；控制台 0 报错。
- `test/backup_narrow_check.py`（390×900）：`scrollWidth=390 / clientWidth=390`，统计卡单列 354px，「回滚」按钮 64×46，控制台 0 报错。
- 打包：`bash scripts/build.sh` 通过五处版本校验，`vite build` 54.24s（110 模块），后端 `-tags embedui` 静态编译；正式包实跑 `/healthz` = `{"db":"ok","status":"ok","version":"1.4.11"}`，`/admin/backup` 200。
- 测试：`go test ./...` 全绿（api / auth / backup / config / render / store / validate），`gofmt -l` 无输出，`go vet ./...` 通过。
- 发包（2026-11-27）：`bash scripts/push-image.sh` RC=0，`1.4.11` 与 `latest` 推送阿里云，digest 同为 `sha256:9fa9f71c…`，manifest `application/vnd.oci.image.manifest.v1+json`（无 OCI 空 manifest）。
- 镜像端到端：用推送的 `1.4.11` 起容器（`-p 18092:8080`，挂 `deploy/data` 副本并 `chown -R 100:101`），`Up (healthy)`，`/healthz` = `1.4.11`，`/admin/backup` 200。
- 产物复测：在 v1.4.11 正式包上再跑 `backup_page_check.py`，下载 221,443 字节、预检通过、恢复后 31 篇 / 8 项目 / 3 图片、快照 `pre-restore-20261126-170541.tar.gz`。

## 上一轮改动记录（v1.4.10，2026-11-27）

- 新增备份/恢复后端能力：`internal/backup` 负责打包（数据库 + 媒体 + manifest + sha256）、解包预检与落盘替换；`internal/api/backup.go` 暴露概览、下载、预检、恢复、快照列表与快照回滚 6 个接口；`internal/store/backup.go` 提供内容统计、schema 版本与媒体引用查询。
- 恢复流程内建安全网：`confirm=RESTORE` 二次确认、恢复前自动快照、任一步失败自动回滚、`maintenanceGuard` 在恢复期间拦截其余 API（503 + `Retry-After`），并发恢复用 `restoreMu` 拒绝第二个请求。
- `Server` 的库句柄从固定指针改为可替换的 `storeRef`（`s.store` → `s.store()`），恢复后换连接、不重启进程。
- 新增 `app/internal/api/backup_test.go`，为上述接口补齐 6 个端到端 HTTP 用例：未登录 401、缺 CSRF 403、概览与下载包预检、完整恢复链路（含恢复前快照与回滚）、坏包拒绝、非 `.tar.gz` 拒绝。
- 修复新测试文件与 `api_admin_test.go` 重复声明 `itoa` 的编译冲突，改为复用既有函数。
- 后台备份管理界面未接入正式前端，本版只留 `prototype/backup.html` 原型，接口先行（v1.4.11 已接入正式页面）。

### 上一轮验收证据（v1.4.10，2026-11-27）

- `go test ./internal/api/ -run 'Backup|Restore' -v`：6 个用例全部 PASS，包耗时 5.337s。
- `go test ./...`：`internal/api` 20.292s、`internal/backup` 4.228s、`internal/store` 9.391s，其余包全绿。
- `gofmt -l internal/api/` 无输出；`go vet ./internal/api/` 通过。
- 覆盖的接口：`GET /api/admin/backup/summary`、`POST /api/admin/backup`、`POST /api/admin/backup/inspect`、`POST /api/admin/backup/restore`、`GET /api/admin/backup/snapshots`、`POST /api/admin/backup/snapshots/:name/restore`。
- 打包：`bash scripts/build.sh` 通过五处版本校验，`vite build` 2.92s（108 模块），后端 `-tags embedui` 静态编译完成；`app/build/dev-notes` 实跑 `/healthz` = `{"db":"ok","status":"ok","version":"1.4.10"}`，首页 200。
- 发包（2026-11-27）：`bash scripts/push-image.sh` RC=0，`1.4.10` 与 `latest` 推送阿里云，digest 同为 `sha256:1e7ae6c2…`，manifest `application/vnd.oci.image.manifest.v1+json`（无 OCI 空 manifest）。
- 镜像端到端：用推送的 `1.4.10` 起容器（`-p 18091:8080`，挂 `deploy/data` 副本并 `chown -R 100:101`），状态 `Up (healthy)`，`/healthz` 返回 `1.4.10`，`/api/notes` 正常返回已有数据。

## 历史改动记录（v1.4.9，2026-11-26）

- 推荐改为笔记自身属性：`note` 加 `is_featured`（`schemaVersion` 3 → 4），后台新建/编辑页勾选「设为首页推荐」，后台列表状态后加「推荐」徽标。
- 首页面板改为取「已发布 + 已推荐」按发布时间倒序的前 3 条；「全部」进入 `/notes?featured=1` 全量推荐列表，标题变「推荐笔记」，可用「仅看推荐 ✕」取消，筛选状态写进 URL。
- 站点设置精简为只维护 ICP 备案号，附「去笔记管理勾选推荐」的引导链接。
- 迁移回填：旧 `settings.featured_notes` 里的推荐位自动写成 `is_featured = 1`，此前配的 3 篇不丢；旧值缺失或 JSON 损坏按未配置跳过。
- 决策记录：上一版把推荐做成「站点设置三个下拉选 slug」，复盘认为推荐本质是笔记属性，改成字段后不必跨页记忆标题，也能在列表里一眼看出谁被推荐（用户明确要求）。
- 已执行 `bash scripts/build.sh` 打包（`app/build/dev-notes` = v1.4.9），并执行 `bash scripts/push-image.sh` 把 `1.4.9` 与 `latest` 推到阿里云。

### 上一轮验收证据（v1.4.9，2026-11-26）

- `go test ./...` 全绿（`api` 含改写后的 `TestSiteFeaturedNotes`：勾选后 `featured=1` 的 `total` 4 篇、首页仅 3 条且按时间倒序、草稿不出现、取消推荐后掉回 3 篇、`featured_slugs` 被忽略）。
- `go test ./internal/store` 新增 `TestMigrateFeaturedBackfill`：造 `user_version=3` 库并写入 `settings.featured_notes`，`Open` 后对应笔记 `is_featured=1`、另一篇仍为 0，重复打开不报错。
- 浏览器端到端（无头 Chromium，1280×900）：库 `user_version=4`、迁移回填 3 条；后台列表徽标 3 个；勾选一条已发布笔记后推荐数 3 → 4、徽标 4 个；首页面板 3 条且都在推荐集合内；点「全部」跳 `/notes?featured=1`、标题「推荐笔记」、卡片 4 张「共 4 篇」；点「仅看推荐 ✕」回 `/notes`、标题回「笔记列表」；站点设置 `select` 数 0；取消推荐后回到 3；控制台 0 报错。
- 排版自检：推荐开关在 1280px 为「☐ 设为首页推荐 + 说明」单行，390px 下勾选框在左、说明换行不溢出；390px 推荐列表「共 N 篇」单行不拆（`.list-count { flex: none; white-space: nowrap }`）。
- 正式包（`:8080`，内嵌构建）：首页面板 3 条 → 「全部」`/notes?featured=1`「推荐笔记」共 3 篇 → 取消筛选回 `/notes` 6 张卡片；390px 横向溢出 0；后台徽标 3、设置页 `select` 0、编辑页推荐开关 1；控制台 0 报错。
- 发包（2026-11-26）：`bash scripts/push-image.sh` 推送 `1.4.9` + `latest` 到阿里云，`1.4.9` manifest digest `sha256:5fabada4…`、`latest` digest `sha256:16e2ecb3…`，media type `application/vnd.oci.image.manifest.v1+json`，无 OCI 空 manifest。
- 镜像端到端：用推送的 `1.4.9` 起容器（`-p 18086:8080`，挂 `deploy/data` 副本并 `chown -R 100:101`），状态 `Up (healthy)`，`/healthz` = `1.4.9`；容器内库自动迁到 `user_version=4`、`note` 有 `is_featured` 列；浏览器勾选「踩的循环的坑」后首页面板出现 1 条、「全部」进 `/notes?featured=1` 共 1 篇；控制台 0 报错。
- Git：提交 `aa022aa`，已推送 `origin/main`（`d71dbc8..aa022aa`）。

### 上一轮踩坑（v1.4.9）

- **`npm ci` 会打断正在运行的 Vite**：打包脚本第一步删 `node_modules` 重装，正在 dev 的页面随后点「用标题生成」报「生成失败，请手动填写链接标识」。根因是 `slug.ts` 用 `import('pinyin-pro')` 动态加载，依赖被换掉后取不到模块，异常落到 catch 提示。现象上极像代码 bug，实际是环境被换。教训：给正在验收的 dev server 打包前先停 dev server；或联调期改用 `npm run build`（不重装依赖）。排查手段：看 Vite 日志 + 复现脚本抓控制台报错，控制台为空即说明不是前端代码问题。
- **残留进程占端口 → 数据看起来「全丢了」**：`8080` 被 53 分钟前启动的 `/tmp/dn-cc-bin`（挂空库 `/tmp/dn-cc`）占着，新后端 `bind: address already in use` 静默退出，前台查到 0 篇笔记、首页推荐面板按设计隐藏、后台推荐位无候选。教训：启动服务后必须核对 `/healthz` 的版本号与端口占用者（`ss -lptn 'sport = :8080'`、`tr '\0' '\n' < /proc/<pid>/environ | grep DATA_DIR`），不能只看 `curl` 有响应就认为服务是自己起的。

## 历史改动记录（v1.4.8，2026-11-26）

- 主题按钮重做：桌面端保留「浅色主题」长文案并补 `title` 提示；≤640px 收成 44×44 圆形按钮、只留状态圆点，桌面/窄屏都不再压住右下角表单控件。
- 主题改为三态：跟随系统（默认，监听 `prefers-color-scheme`）/ 浅色 / 深色，点击浮动按钮循环切换；旧键 `dev-notes-theme` 迁移到 `dev-notes-theme-mode`。
- 首屏主题初始化脚本从 `index.html` 内联迁到 `web/public/theme-init.js`（生产 CSP `script-src 'self'` 会拦内联脚本）。
- 后台内容区窄屏增加 88px 底部留白，页脚为按钮预留的右内边距从 132px 收到 76px；`vite.config.ts` 开启 `server.watch.usePolling`。
- 后台新建笔记/项目支持「用标题（名称）生成」链接标识，编辑页不自动改写；新增 `web/src/slug.ts`，依赖 `pinyin-pro`（动态 import，只进后台编辑页 chunk）。
- 首页右上角面板改为「推荐笔记」：后台站点设置三个推荐位（存 `settings.featured_notes`），`/api/site` 只下发存在且已发布的笔记，最多 3 条；全部为空时首页隐藏面板、hero 转单列。（v1.4.9 已改成笔记自身 `is_featured` 字段，此段仅作历史记录）
- 修复既有的窄屏问题：≤640px 时后台页头按钮被压成竖排文字（`.admin-heading` 改纵向堆叠、按钮 `white-space: nowrap`、`.row-actions` 允许换行）。
- 本轮已执行 `bash scripts/build.sh` 打包，`app/build/dev-notes` 为 v1.4.8 正式包；`bash scripts/push-image.sh` 已把 `1.4.8` 与 `latest` 推到阿里云。

### 历史验收证据（v1.4.8，2026-11-26）

- 无头 Chromium（390×900 / 1280×900）跑 `test/theme_button_check.py`：390px 按钮 44×44、`border-radius: 50%`、文本长度 0、圆点居中偏差 0px、`title` =「当前为跟随系统，点击切换」；点一次后 `title` 变「当前为浅色，点击切换」、`data-theme="light"`。
- 遮挡检查（滚到底、真实矩形相交）：公开 `/`、`/projects`、`/about` 与后台 `/admin/notes`、`/admin/notes/new`、`/admin/projects`、`/admin/settings`、`/admin/account` 全部 `[]`，横向溢出 0px。
- 1280px：按钮 100×40，长文案「浅色主题」可见，`title` 正常。
- 生产产物（`app/build/dev-notes` 起在 `:18080`，挂 `deploy/data` 副本）：`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.4.8"}`；控制台报错 0（修复前有 5 条内联脚本 CSP 违规）；拦截业务 JS 后 `data-theme` 仍为 `dark`，证明 `theme-init.js` 在 Vue 挂载前生效；下发 CSS 中 `theme-short` 出现 0 次、含 `width:44px;height:44px;…border-radius:50%` 与 `padding:18px 18px 88px`。
- 构建：`bash scripts/build.sh` 版本一致性校验通过（`1.4.8` 五处一致），`vite build` 3.59s，后端 `-tags embedui` 静态编译，产物 27.8 MB。
- 发包（2026-11-26）：`bash scripts/push-image.sh` **RC=0**，`1.4.8` 与 `latest` 已推送阿里云；`docker manifest inspect` 显示 `application/vnd.oci.image.manifest.v1+json`（与 `1.4.7` 一致，无 OCI 空 manifest）；`latest` digest `sha256:4a8ef02957cc…`。
- 云端镜像端到端：用推送后的 `1.4.8` 起容器（`-p 18085:8080`，挂 `deploy/data` 副本并 `chown -R 100:101`），状态 `Up (healthy)`，`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.4.8"}`，首页 `200`、`/theme-init.js` `200`；对容器跑 `test/theme_button_check.py` 全部通过（44×44、无遮挡、CSP 报错 0）。
- Git：提交 `b60b34e`，已推送 `origin/main`（`585d998..b60b34e`）。

### 历史踩坑（v1.4.8）

- **Vite 在 WSL 下不热更新**：现象是改完 `ThemeSwitch.vue` 模板后，浏览器仍拿到旧组件（`title` 取不到、仍是 `theme-short`），但文件内容已是新版。根因：源码在 `/mnt/e`（Windows drvfs 挂载），inotify 事件不会传进 WSL，Vite 的 watcher 收不到变更，继续返回缓存的 transform 结果。定位方式：直接 `curl http://127.0.0.1:5173/src/components/ThemeSwitch.vue` 看服务端返回的模块，发现仍是旧代码。修复：`vite.config.ts` 加 `server.watch.usePolling: true, interval: 300`。教训：在 `/mnt/` 下开发必须开轮询，否则「改了没生效」会误判成代码问题；同时避免同时留多个 vite 进程抢同一端口。
- **生产 CSP 拦掉内联主题脚本**：现象是正式包里控制台报 `Executing inline script violates the following Content Security Policy directive 'script-src 'self''`，首屏防闪烁失效。根因：`app/internal/api/router.go` 的 `securityHeaders()` 只放行 `script-src 'self'`，而 `index.html` 用的是内联 `<script>`。修复：脚本迁到 `web/public/theme-init.js`，用 `<script src="/theme-init.js">` 同步加载（`public/` 下的文件不带哈希、不享长缓存，改完即时生效）。教训：Vite 的 `index.html` 内联脚本在 dev 下正常、生产被 CSP 拦，验收必须对**构建产物**再跑一遍，不能只测 dev server。
- **遮挡指标要用真实矩形相交**：最初只比较水平区间，得出「按钮覆盖下拉 327px」的假结论（实际垂直方向根本不相交）。改为按矩形交面积计算后，才暴露出真实问题是滚到底时压住「保存」主按钮（444px²），据此补了底部留白。

## 本轮改动记录（v1.4.7，2026-11-25）

- 项目索引与首页项目卡片改为「左缩略图 + 右信息」横向布局：缩略图固定 208×156（4:3），`object-fit: contain` 以高为准、水平居中、两边留白；无封面用首字母占位块，卡片高度统一。
- 项目详情页改为「左说明 + 右截图」并排，截图限高 520px、按比例居中、不放大不裁切；截图可点击打开全屏灯箱（模糊遮罩、Esc/点遮罩关闭、打开时锁滚动）。
- 卡片摘要限两行、底部「关联笔记数 / 查看详情」沉底；≤640px 自动改为上下排列；清掉不再使用的 `.project-cover` / `.project-hero` / `.about-grid` 样式。
- 部署与文档同步：`deploy/docker-compose.yml` 镜像 tag 跟到 `1.4.7`（含「必须先推送镜像」注释）、`doc/DEPLOY.md` 升到 v1.4.7、`doc/PROJECT_STATE.md` 验收命令由 `/tmp/dn-dev` 改为 `deploy/data` 副本。
- 版本升至 `v1.4.7`（`app/VERSION` / `app/main.go` / `web/package.json` / `web/package-lock.json` / `web/src/version.ts` 五处 + 文档）。

### 验收证据（v1.4.7，2026-11-25）

- `go test ./... -count=1`（WSL `lxsyzd`）：`internal/api` 13.9s、`store` 2.9s 等全部 ok。
- `bash scripts/build.sh`：版本一致性校验通过，前端 `vite build` 产出 `dist/index.html` 0.64 kB、CSS 22.39 kB、JS 172.40 kB，后端以 `embedui` 静态编译；最终产物为正式包（明文凭据计数 **0**）。
- 本机冒烟（挂 `deploy/data` 的临时副本）：`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.4.7"}`；`/api/projects` 返回 SiteGuard 等真实项目；`/api/notes` 返回已发布笔记；首页与 `/admin/login` 均 `200`；下发的 `/assets/index-*.js` 内含 `1.4.7`。
- 发包（2026-11-25）：`bash scripts/push-image.sh` **RC=0**，`1.4.7` 与 `latest` 已推送阿里云；`docker buildx imagetools inspect` 显示 media type `application/vnd.oci.image.manifest.v1+json`、digest `sha256:ba93f61c…`（无 OCI 空 manifest）；镜像大小 88.1 MB。
- 云端端到端：`docker pull` 后挂 `deploy/data` 副本（`chown -R 100:101`）起容器，状态 `Up (healthy)`，日志 `版本 1.4.7 env=production`；`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.4.7"}`，`/api/projects` 返回真实项目，首页 `200`。

## 历史改动记录（v1.4.6，2026-11-17）

- 新增 `deploy/Dockerfile`：Node 22 编前端 → Go 1.27 按 `TARGETOS/TARGETARCH` 编后端 → alpine 3.20 运行；非 root、带 healthcheck，依赖走 npmmirror / goproxy.cn。
- 新增 `deploy/docker-compose.yml`（1Panel 可直接粘贴）、`deploy/.env.example`、根 `.dockerignore`、`scripts/push-image.sh`（只推 amd64，传统 docker build + push，复用已登录会话，不写凭据）。
- 重写 `doc/DEPLOY.md` 到 v1.4.6：二进制 + systemd 与 Docker + 1Panel Compose 两条路线、环境变量、备份/升级/回滚、备案步骤与故障排查。
- 版本升至 `v1.4.6`（5 处版本字段 + 文档）。

### 验收证据（v1.4.6，2026-11-17）

- `go test ./...` 全部 ok；`bash scripts/build.sh` 版本一致性通过。
- 镜像构建：`docker build --platform linux/amd64 -f deploy/Dockerfile -t dev-notes:1.4.6-selftest .` **RC=0**；构建日志显示多阶段前端/后端全部成功。
- 容器冒烟（`docker run -p 18094:8080`）：`/healthz` 返回 `1.4.6`；`GET /api/notes`、`/api/projects`、`/api/site` 正常；登录拿到 CSRF 后 `POST /api/admin/media/upload` 返回 200 与 `/media/uploads/<32hex>.png`；`ls /app/data` 显示 `notes.db` 与 `media/` 均属 `app` 用户；镜像内 `id app` 实测为 uid=100(app) / gid=101(app)（alpine 中 gid 100 已占用，`addgroup -S app` 分到 101），故宿主机数据卷应 `chown -R 100:101`。
- 修复记录：首版 Dockerfile 把 `app/` 复制到 `/src/app` 导致 `go build ./app` 报「主模块不包含该包」（app/go.mod 就是模块根，不应再套一层），改为 `COPY app/ ./` + `go build .` 后通过。用户要求取消 arm64，推送脚本改为只构建 amd64。
- 镜像推送：`bash scripts/push-image.sh` **RC=0**，`1.4.6` 与 `latest` 均已上传阿里云；`docker buildx imagetools inspect` 显示 media type `application/vnd.oci.image.manifest.v1+json`（不含 OCI 空 manifest）。
- 云端镜像端到端：从阿里云 `docker pull` 后带 `SESSION_SECRET` / `ADMIN_INIT_USERNAME` / `ADMIN_INIT_PASSWORD` 启容器，状态 `Up (healthy)`；`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.4.6"}`，`/api/site` 正常。未传必需变量时容器按设计拒绝启动并给出明确错误。

## 历史改动记录（v1.4.5，2026-11-17）

- 新增 `web/src/components/MarkdownEditor.vue`，12 个工具栏按钮、Markdown 原文输入、实时分栏预览、三种视图模式与图片上传；接入 `AdminNoteEditView.vue` 正文与 `AdminProjectEditView.vue` 项目说明。
- 预览走服务端 `POST /api/admin/preview`，400ms 防抖 + 请求序号只应用最新响应；上传走 `POST /api/admin/media/upload`，成功后插入 `![alt](/media/uploads/…)` 并选中 alt。
- 移除两个编辑页原有的「预览 / 预览说明」按钮与独立预览区，避免重复入口。
- 新增界面原型 `prototype/md-editor.html`（可直接在浏览器打开），先确认工具栏、上传与分栏交互再写真实代码。
- 窄屏（≤640px）`.editor-row` 改单列、表单控件 `min-width: 0`，修掉日期控件造成的 34px 横向溢出。
- 版本升至 `v1.4.5`（5 处版本字段 + 文档）。

### 验收证据（v1.4.5，2026-11-17）

- 原型浏览器验收：图片按钮选文件 → 上传中状态 → 插入语法并预览，`alt` 文字自动选中，控制台 0 错误。
- `gofmt -l .` 无输出；`go vet ./...` 无输出；`go test ./... -count=1` 全部 ok；`bash scripts/build-test.sh` 版本校验与构建通过，`/healthz` 返回 `1.4.5`。
- 浏览器验收（Playwright，真实库 `/tmp/dn-dev`，只上传测试图、未保存任何笔记/项目）：
  - 工具栏 12 个按钮；1280px 分栏为 `439px + 439px`；点「粗体」把 `这是正文` 变成 `**这是正文**`；
  - 输入 `## 验收标题` 后右侧 400ms 内出现 `<h2>验收标题</h2>` 与 `<strong>这是正文</strong>`；
  - 上传测试 PNG → 插入 `/media/uploads/d685…png`、预览图片 1 张、按钮从「上传中…」恢复为「图片」；
  - 「仅编辑」隐藏预览、「仅预览」隐藏输入框；项目说明页编辑器 1 个、原文长度 265 正常回填；
  - 390px 分栏变为单列 `352px`、横向溢出 **0**（修复前为 34）；控制台错误 0 条。

## 历史改动记录（v1.4.4，2026-11-17）

- 后台项目编辑页封面框新增「选择图片上传」按钮：前端先校验扩展名（png/jpg/jpeg/webp）与大小（≤2 MB），上传成功自动填入 `/media/uploads/…`，封面缩略图即时刷新。
- 新增 `api/internal/api/media.go`：`POST /api/admin/media/upload`（鉴权 + CSRF）做扩展名白名单、`http.DetectContentType` 内容嗅探、2 MB 上限、crypto/rand 随机文件名，写 `.part` 后原子改名到 `DATA_DIR/media`；`GET /media/uploads/:name` 严格名校验后下发，长缓存。
- `api.New` 增加 `mediaDir` 参数（`cfg.DataDir + "/media"`）；上传路由单独 `bodyLimit(3<<20)`，管理端其他接口仍为 1 MB。
- `requestJSON` 识别 `FormData` 并交给浏览器设置 multipart 边界；`admin.ts` 新增 `uploadMedia`。
- 版本升至 `v1.4.4`（5 处版本字段 + 文档）。

### 验收证据（v1.4.4，2026-11-17）

- `gofmt -l .` 无输出；`go vet ./...` 无输出；`go test ./... -count=1` 全部 ok。
- 新增 `TestMediaUpload`：未登录 401、缺 CSRF 403、SVG 扩展名 422、改名伪装的文本 422、超 2 MB 422；合法 1×1 PNG 返回 `/media/uploads/<32位hex>.png`、文件落盘、直接 GET 200 且 `Content-Type: image/png`、非法文件名 404。
- 浏览器（Playwright，独立临时库 + 测试包）：选伪造 PNG → 服务端嗅探拒绝且封面路径保持空；选真实 PNG → 路径自动填入且格式匹配、缩略图 1 张；保存项目后前台详情页与卡片封面 `naturalWidth=1`（测试图为 1×1）、直接访问图片 200 `image/png`。控制台仅出现那次**故意触发**的 422，无其他错误。
- 真实库未改动：服务切回 `/tmp/dn-dev` 后 `/healthz` 返回 `1.4.4`，项目数仍为 8。

## 历史改动记录（v1.4.3，2026-11-17）

- 新增 `settings(key TEXT PRIMARY KEY, value TEXT, updated_at TEXT)` 键值表，`schemaVersion = 3`；`Store.GetSetting/SetSetting` 处理未配置返回空串与同键覆盖（UPSERT）。
- 接口：`GET /api/site`（公开，页脚用）、`GET/PUT /api/admin/settings`（鉴权 + CSRF）；`validate.ICPNumber` 校验备案号，空串合法用于恢复占位。
- 前端：新增 `AdminSettingsView.vue` 与 `/admin/settings` 路由、后台侧栏入口；`SiteFooter.vue` 改为从共享状态 `web/src/site.ts` 读备案号，`App.vue` 挂载时拉一次，后台保存后 `loadSiteInfo(true)` 立即刷新。
- 备案号规则：留空或 5–100 位汉字/字母/数字/连字符；非法值前后端都拦截，提示里给出示例。
- 版本升至 `v1.4.3`（5 处版本字段 + 文档）。

### 验收证据（v1.4.3，2026-11-17）

- `gofmt -l .` 无输出；`go vet ./...` 无输出；`go test ./... -count=1` 全部 ok。
- 新增用例：`TestSettingsStore`（未配置空串、覆盖更新、清空）、`TestSiteSettings`（公开初值为空、未登录 401、缺 CSRF 403、`<script>` 值 422 且指明字段、保存后公开接口可读、清空后恢复空串）；`TestMigrateV1Database` 补断言：v1 老库升到 v3 后 `settings` 表可读。
- 浏览器验收（Playwright，临时空库 + 测试包）：初始页脚为「甘ICP备xxxxxxxx号（备案通过后显示）」；设置页初始值为空、版本行 `v1.4.3 by Mr lin`；填 `<script>x</script>` 被前端拦截并提示示例；保存 `甘ICP备12345678号-1` 后首页页脚显示该备案号；清空后项目页页脚恢复占位；390px 页脚无横向溢出；控制台错误 0 条。
- 真实库升级核验（`/tmp/dn-dev/notes.db`，升级前备份为 `notes.db.bak-before-v3`）：`user_version=3`、`settings` 表存在且 0 行、项目 8 / 笔记 31 未变、`/api/site` 返回空备案号、前台接口 30 篇已发布 + 8 个项目；`/healthz` 返回 `1.4.3`。

## 历史改动记录（v1.4.2，2026-11-17）

- 后台登录页 `AdminLoginView.vue` 仅在 `import.meta.env.VITE_TEST_LOGIN === '1'` 时，用 `VITE_TEST_USERNAME` / `VITE_TEST_PASSWORD` 预填两个输入框；正式构建无凭据，输入框留空。
- 新增 `scripts/build-test.sh`：注入 `DN_TEST_BUILD=1`、`VITE_TEST_LOGIN=1` 与测试凭据（默认测试账号密码，可由同名环境变量覆盖），再调用 `scripts/build.sh`；输出目录参数透传。
- `scripts/build.sh` 增加防误用：未标记 `DN_TEST_BUILD=1` 却带 `VITE_TEST_LOGIN=1` 时直接退出。
- 新增 `web/src/vite-env.d.ts` 声明三个构建变量类型；版本升至 `v1.4.2`（5 处版本字段 + 文档）。

### 验收证据（v1.4.2，2026-11-17）

- 防误用：直接 `VITE_TEST_LOGIN=1 bash scripts/build.sh /tmp/dn-guard-check` → 输出「错误：正式构建不允许 VITE_TEST_LOGIN=1；测试包请用 scripts/build-test.sh」，`guard-exit=1`。
- 隔离性：`bash scripts/build.sh /tmp/dn-release-check` 后正式二进制中明文密码 `grep` 计数 **0**；`bash scripts/build-test.sh` 后 `app/build/dev-notes` 中计数 **1**。
- 构建：`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.4.2"}`；`gofmt -l .` 无输出；`go vet ./...` 无输出；`go test ./... -count=1` 全部 ok。
- 浏览器（Playwright）：`/admin/login` 预填账号 `admin`、密码长度 13；不输入任何内容直接点「登录管理后台」→ 落到 `/admin/notes`；后台侧栏版本行 `v1.4.2 by Mr lin`；控制台错误 0 条。

## 历史改动记录（v1.4.1，2026-11-17）

- 后台「笔记管理」由一次性拉 50 条改为每页 6 条 + 分页条，翻页参数走后端 `/api/admin/notes?page=&page_size=`（接口早已支持，前端没接）。
- 后台「项目管理」同样每页 6 条：`store.ListAllProjects` 改为接收 `AdminProjectFilter{Limit, Offset}` 并新增 `CountAllProjects`；`GET /api/admin/projects` 增 `page` / `page_size`（默认 20、上限 100），用 `respondList` 返回 `total` / `page` / `page_size`。
- 抽出 `web/src/components/Pager.vue`（页码折叠规则 + 按钮结构 + 页码文案），前台笔记列表、后台笔记/项目三处共用；`NotesView.vue` 里原来的内联分页模板与 `pageNumbers` 计算删除。
- 交互对齐前台：翻页回顶、切筛选/改搜索词回第 1 页、总页数变少时退到最后一页（含删掉最后一页唯一一条的场景）。
- 后台只有一页时不渲染翻页按钮，用一行「共 N 篇 / 个」补总数；多页时该行隐藏，总数由分页条承担。
- 新增样式 `.admin-total`、`.admin-pager .pager`（靠左、间距比前台紧）。
- 版本升至 `v1.4.1`（5 处版本字段 + 文档）。

### 验收证据（v1.4.1，2026-11-17）

- `gofmt -l .` 无输出；`go vet ./...` 无输出；`go test ./... -count=1` 全部 ok。
- 新增用例：`TestAdminProjectPaging`（7 个项目按 `sort_order` 分页：第 1 页 6 条为 1..6、第 2 页 1 条为 7、越界页返回空列表、`CountAllProjects=7`）；`TestAdminProjectListPaging`（接口层：`page=1&page_size=6` 返回 6 条且 `total=7`、`page=2` 返回 1 条、不传参数时按默认 `page_size=20` 返回 7 条）。
- 接口自查（临时库 8 个项目、15 篇笔记）：`projects page2: 2 total 8 page 2 size 6`；`notes page2: 6 total 15 page 2 size 6`。
- 浏览器验收（Playwright + WSL Chromium，1280px / 390px）：
  - 后台笔记：第 1 页 6 行、`第 1 / 3 页 · 共 15 篇`、页码按钮 `上一页 1 2 3 下一页`；翻下一页 6 行且 `scrollY=0`；点第 3 页 3 行、高亮为 3；切「已发布」回到 `第 1 / 2 页 · 共 7 篇`；搜索命中 1 条时分页条消失、改为一行「共 1 篇」；多页时独立总条数行为 0（不重复）。
  - 后台项目：第 1 页 6 行、`第 1 / 2 页 · 共 8 个`；第 2 页 2 行、首行为 `paging-demo-7`。
  - 390px 后台笔记：可见按钮仅 `上一页 / 下一页`，横向溢出 0。
  - 前台 `/notes` 回归：第 1 页 6 张卡片、`第 1 / 2 页 · 共 7 篇`；翻第 2 页 1 张、`scrollY=0`。
  - 控制台错误 0 条。
- 构建：`bash scripts/build.sh` 通过，`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.4.1"}`。

## 历史改动记录（v1.4.0，2026-11-17）

- 项目表新增 `description`（Markdown 说明，≤ 20 KB）与 `cover_image`（封面图站内路径）两列，`schemaVersion` 升到 2。
- 迁移用 `PRAGMA table_info(project)` 探测列存在性，缺列才 `ALTER TABLE ... ADD COLUMN`；新建库跑 v1 建表后再走 v2，与老库同一路径。
- Store 四个触点同步改：`projectColumns` 增两列、`scanProject` 的 `Scan` 实参按同序补齐、`CreateProject` 的 INSERT 扩到 11 个占位符、`UpdateProject` 增两列赋值。
- 接口层：`projectInput` 增 `Description`/`CoverImage`（去首尾空格后入库）；前台详情把 `description` 渲染成 `description_html` 并把原文置空；前台列表把 `description` 置空但保留 `cover_image`。
- 校验层：说明 ≤ 20 KB 且纳入备案词扫描（顺序追加 `description`）；封面路径必须匹配 `^/(assets|media)/…\.(png|jpe?g|webp|svg|gif)$` 且不含 `..`、`//`。
- 前端：项目卡片显示封面（16:9 容器 + `object-fit: contain`，不裁字）；详情页封面按原比例展示、说明用 `.project-description` 渲染（内嵌 `h2/h3` 字号下调）；后台项目编辑页加封面输入（带缩略图）、说明编辑区（字节计数）与「预览说明」。
- 版本升至 `v1.4.0`：`app/VERSION`、`app/main.go`、`web/package.json`、`web/package-lock.json`、`web/src/version.ts`、全部文档同步。

### 验收证据（v1.4.0，2026-11-17）

- `gofmt -l .` 无输出；`go vet ./...` 无输出；`go test ./... -count=1`：`api`、`auth`、`config`、`render`、`store`、`validate` 全部 ok。
- 新增用例：`TestAdminProjectCRUD` 扩展说明/封面的创建与更新读回；`TestProjectDetailIncludesPublishedNotes` 校验 `description_html` 含 `<h2>定位</h2>` 且 `description` 为空；`TestProjectListOmitsDescriptionSource`；`TestProjectCoverImageRules`（7 组路径正反例，含 `..` 穿越）；`TestProjectDescriptionRules`（超限、命中禁用词、正常文本）；`TestMigrateV1Database`（v1 建表 + `user_version=1` → `Open` 自动补列、老数据保留、新列可写可读、重复打开不报错）。
- `bash scripts/build.sh`：版本一致性校验通过（1.4.0），前端 68 模块构建成功，`app/build/dev-notes` 产出。
- 接口冒烟（临时库 `DATA_DIR=/tmp/dn-e2e`，`:18091`）：
  - `POST /api/admin/projects` 带说明与封面 → 201，返回 `id=1 cover=/media/siteguard-origin.svg desc_bytes=133`；
  - `/api/projects` 列表 → `slug=cover-desc-demo cover='/media/siteguard-origin.svg' description=''`；
  - `/api/projects/cover-desc-demo` → `description=''`、`description_html` 含 `<h2>这是什么</h2>` / `<strong>说明渲染</strong>` / `<ul>`；
  - 非法封面 `https://evil.example/a.png` → 422 `VALIDATION_FAILED field=cover_image`；
  - 说明含「支付」→ 422 `CONTENT_BLOCKED field=description hits=["支付"]`。
- 浏览器验收（Playwright + WSL 内 Chromium，1280px / 390px，`http://localhost:18091`）：
  - `/projects`：1 张卡片，封面元素 531×298.7（16:9），原图 860×330 未被裁切；390px 下封面 308×173.2，横向溢出 0；
  - `/projects/cover-desc-demo`：详情封面 1180×454（比例 2.599，原图 2.606，未变形）；说明 `h2=['这是什么']`、`li` 2 条、`strong=['说明渲染']`，回退静态文案节点数为 0；
  - `/admin/projects/1/edit`：说明字段与封面路径回填正确、缩略图 1 个、字节计数「133 / 20480 字节」、预览 `h2/li` 正常；填外链封面后本地即拦截并提示「封面图需填站内路径，如 /media/projects/example.png」；
  - 控制台错误 0 条。
- 复验：改 `object-fit: contain` 与说明标题字号后重新打包、重启（同库），上述指标复测一致。

## 历史改动记录（v1.3.19，2026-11-17）

- 首页「项目索引」区块副标题由「个人练习作品，仅作记录与源码外链」改为「个人练习作品，共 N 个」，与 `/projects` 一致。
- 首页只渲染前 2 个项目，数字改用 `fetchProjects()` 的 `total`（新增 `projectTotal`），不能取 `indexProjects.length`。
- 版本升至 `v1.3.19`。

### 验收证据（v1.3.19，2026-11-17）

- 1280px 首页：项目索引区块副标题为「个人练习作品，共 8 个」，下方仍只渲染 2 张卡片，两张卡片状态均为「已完成」（与 `/projects` 一致）。
- 全站搜索「仅作记录与源码外链」→ 0 处残留；项目卡片不再出现「已发布」。
- 构建通过，`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.3.19"}`。

## 历史改动记录（v1.3.18，2026-11-17）

- 项目状态文案改为「已完成」：`api.ts` 新增 `projectStatusText`，项目索引卡片与项目详情页统一引用；笔记侧仍用原 `statusText`。
- 项目卡片状态标签改为按 `project.status` 动态取类名与文案，不再写死 `status-published` / 「已发布」。
- 删除项目索引页顶部提示条「本站不托管安装包或可执行文件。受许可证约束…」；`.callout` 样式保留（后台仍在用）。
- 验收：`/projects` 8 张卡片全为「已完成」，`.callout` 节点数 0；`/projects/siteguard` 同步；首页笔记仍「已发布」。

## 历史改动记录（v1.3.17，2026-11-17）

- 项目索引页副标题改为「个人练习作品，共 N 个」（N 取 `projects.length`），删除右侧重复的「已公开 N 个」。
- 验收：`.section-subtitle` = 「个人练习作品，共 8 个」，卡片数 8，右侧无重复计数。

## 历史改动记录（v1.3.16，2026-11-17）

- 首页导语（`HomeView.vue` 的 `.lead`）删除「内容仅作个人非经营性分享，项目只提供源码入口，不提供任何下载。」，只保留「这里记录我在学习软件开发过程中的技术笔记、问题排查与项目复盘。」。
- 验收：1280px 首页 `.hero .lead` 与全页文本均不含「下载」「非经营性」。

## 历史改动记录（v1.3.15，2026-11-17）

- 前台页头去掉「后台管理」，访客侧不再暴露后台入口；后台改为直接访问 `/admin`。`README.md` 的「进入后台」同步改写。
- 笔记列表改为每页 6 篇 + 页码分页（上一页 / 页码 / 下一页 + 「第 X / Y 页 · 共 N 篇」）；`PAGE_SIZE` 由 50 改为 6。
- 翻页滚回顶部；切项目筛选、改搜索词自动回第 1 页；筛选后总页数变少时自动退到最后一页。
- 窄屏（≤600px）隐藏页码按钮，只留上一页 / 下一页与页码文字。
- 删除无用的 `.header-actions` 样式与窄屏隐藏规则。

### 验收证据（v1.3.15，2026-11-17）

- 1280px：`/notes` 首屏 6 篇；页码条按钮为 `上一页 1 2 3 4 5 6 下一页`，信息「第 1 / 6 页 · 共 31 篇」。
- 点第 3 页后：信息变「第 3 / 6 页 · 共 31 篇」，高亮按钮为 `3`，列表换成第 13~18 篇，`window.scrollY == 0`（已回顶部）。
- 页头链接实测为 `我的开发笔记 / 首页 / 笔记 / 项目 / 关于`，已无「后台管理」。
- 390px：页码按钮隐藏，仅 `上一页 / 下一页` + 页码文字，`scrollWidth == 390`。
- 构建通过，`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.3.15"}`。

## 历史改动记录（v1.3.14，2026-11-17）

- 关于页宽度去掉单独限宽（原居中 780px），改为跟随 `.container`，与项目页同宽。
- 段落字号提到 `1.02rem`、行高 1.95、段间距 16px，缓解 1180px 宽栏的长行阅读压力。
- 验收：1440px 视口下 `.project-grid` 与 `.about-single` 实测均为 `left=130 / right=1310 / 宽=1180`。

## 历史改动记录（v1.3.13，2026-11-17）

- 关于页单栏由「左对齐 720px」改为「居中 780px」（`margin-inline: auto`），修掉宽屏右侧大块空白。
- 去掉 `.about-prose p` 的 `66ch` 与 `.about-prose .lead` 的 `62ch`，标题 / 导语 / 段落统一到同一条右边缘。
- 验收：1440px 下宽 780、左右各 330，`h1` 与 `p.lead` 均为 780；390px 下容器 358、左右各 16，无横向滚动。

## 历史改动记录（v1.3.12，2026-11-17）

- 关于页删除右侧信息栏（内容范围 / 项目 / 更新节奏 / 后台入口 / 站点性质 / 备案信息），只留三节正文 + 末尾版本行。
- 版式改单栏：`AboutView.vue` 去掉 `.about-grid` 包裹，改 `.about-single`（`max-width: 720px`）；`.about-grid` / `.definition` 样式保留，`ProjectDetailView` 仍在用。
- 版本升到 `v1.3.12`，`web/src/version.ts` 同步（关于页、后台侧栏、登录页的版本行都从这里取）。

### 验收证据（v1.3.12，2026-11-17）

- 构建通过，`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.3.12"}`。
- 关于页 DOM：`.about-grid` 与 `.definition` 均不存在；`.about-single` 计算宽度上限 `720px`；`.about-version` 文本 `v1.3.12 by Mr lin`。
- 1280px：`scrollWidth == 1280`，单栏正文左侧与页头 logo 对齐，无横向滚动。

## 历史改动记录（v1.3.11，2026-11-17）

- 页脚（`SiteFooter.vue`）去掉「个人非经营性技术记录 · 源码仅外链 GitHub」，只留备案号一行；版本与署名按要求移到关于页。
- 关于页（`AboutView.vue`）重写：结构参照常见个人技术博客的关于页，正文三节「我是谁？」「为什么建这个站？」「在哪里可以找到我？」，右侧信息栏六项（内容范围 / 项目 / 更新节奏 / 后台入口 / 站点性质 / 备案信息），正文末尾一行 `v1.3.11 by Mr lin`。
- 新增 `web/src/version.ts` 作为前端唯一版本来源（`APP_VERSION` / `APP_AUTHOR` / `VERSION_LINE`），关于页、后台侧栏、登录页统一引用，后续升版本只改这一处 + 三处构建字段。
- 修 `.definition` 的 `align-content`：正文变长后右栏网格行被拉伸，条目间出现过百像素空白。

### 验收证据（v1.3.11，2026-11-17）

- 构建：`vite build` + Go embed 通过，`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.3.11"}`。
- 桌面 1280px：关于页 `h1=关于`，`h2` 三节齐全，正文 9 段，`.about-version` 文本 `v1.3.11 by Mr lin`，`scrollWidth == clientWidth == 1280`。
- 页脚：`innerText` 仅 `甘ICP备xxxxxxxx号（备案通过后显示）`，版本行已从全站页脚消失。
- 窄屏 390px：`scrollWidth == 390`，无横向滚动条，正文与信息栏正常堆叠。
- 右侧信息栏：修正后条目间距回到 18px，不再被拉伸。

## 历史改动记录（v1.3.10，2026-11-17）

- 品牌标识改版：`<span class="brand-mark">N</span>` → 组件 `BrandLogo.vue`（内联 SVG，32×32 设计稿）。
- 图形为「折角笔记纸 + 字母 N」：纸面暖白、右上折角压深、N 用 accent 描边；外框仍是 `--accent` 圆角方块，暗色主题跟随 `#e08a6c`。
- 方案取舍：同时比过「折角纸 + 两道笔记线」（A）与「加粗版」（B），两者在 32px 上像通用文档图标、24px 以下线条糊成一块；最终选 C「纸上的 N」，字母在 16px 仍可辨，且延续原来「N」的记忆点。
- `favicon.svg` 同步为同一几何的 64×64 版本。
- `.brand-mark` 从「文字容器」改为「SVG 容器」：去掉字号/字重/文字色，加 `flex: none`、`overflow: hidden`，SVG 100% 填充；尺寸 32×32、圆角 10px 不变，头部与侧栏布局零位移。

### 验收证据（v1.3.10，2026-11-17）

- 首页 DOM：`.brand-mark` 内含 `<svg>`，`getBoundingClientRect` 为 `32 × 32`，`path` 节点 3 个，底色 `var(--accent)`、纸面 `#fffaf3`。
- 暗色主题：`document.documentElement.dataset.theme='dark'` 后 `.brand-mark` 计算背景为 `rgb(224, 138, 108)`（即 `--accent` 暗色值），纸面与 N 对比正常。
- 浏览器截图（Edge 1280px，明暗两套）确认头部 logo 与「我的开发笔记」文字基线对齐，导航未位移。
- 后台侧栏 `AdminLayout.vue` 同一组件、同一尺寸，随侧栏 32px 占位渲染。

## 历史改动记录（v1.3.9，2026-11-17）

- 新增项目 `exceltohub`（台账汇总器 / LedgerHub，id=8）：`C# / .NET 8 / WPF / ClosedXML`，源码外链 `https://github.com/mylxnet/ExcelToHub`；该仓库未声明许可证，站点只放外链，不做分发说明。
- 新增 3 篇，素材取自 `README.md` 与 `doc/PROJECT_STATE.md`（踩坑 1~5、v1.1.6 脱敏记录、38 项冒烟 / 13 项 UI 回归）：
  - `wpf-cell-selection-kills-row-apis` —— `SelectionUnit="Cell"` 视作一次 API 变更：`SelectedItem` 恒空、`IsSelected` 赋值抛异常、`SelectedCellsChanged` 同一格不重抛。
  - `datatable-indexer-binding-sort-crash` —— 索引器绑定 `{Binding [列名]}` 被排序当成属性名，异常文本反推到 `SortDescription`。
  - `desensitize-proof-by-unzip` —— 脱敏验收靠解包 `xl/sharedStrings.xml` 双向核对，并明确「移出仓库 ≠ 历史清除」这一遗留。
- 骨架分别为「一个前提三处塌方」「异常文本反推」「证据链式验收」，延续 §12.3 约定。

### 验收证据（v1.3.9，2026-11-17）

- 项目创建：`POST /api/admin/projects` 返回 `201 id=8`。
- 导入：`创建 3，更新 0，跳过 28，失败 0`。
- 站点统计：笔记 31 篇；`/api/projects/exceltohub` 返回技术栈 `C#/.NET 8/WPF/ClosedXML`、`note_count=3`。
- 正文渲染结构（`content_html` 统计）：`wpf-cell-selection-kills-row-apis` h2=7/table=2/li=4/pre=7；`datatable-indexer-binding-sort-crash` 8/2/0/4；`desensitize-proof-by-unzip` 6/3/0/2。
- 备案词自检：全文扫描 `hardWords` 与 `downloadPhrases`，命中 0。

## 历史改动记录（v1.3.8，2026-11-17）

- 新增项目 `quickocr`（QuickOcr，id=7）：`C# / .NET 8 / WPF / Tesseract 5`，MIT，源码外链 `https://github.com/mylxnet/QuickOcr`。
- 新增 1 篇 `free-ocr-ledger`（免费的 OCR 能用：钱省下来了，代价挪到了别处），素材取自该仓库 `README.md` 与文件树（无 `doc/` 目录，文档全在 README）。
- 骨架为「记账」：收入栏记四项许可（代码 MIT / Tesseract Apache 2.0 / 语言包 tessdata_fast / self-contained 运行时），支出栏记体积 226 MB、精度天花板、首次联网拉语言包、无测试无 doc 的维护面，末尾给三条判断标准。

### 验收证据（v1.3.8，2026-11-17）

- 项目创建：`POST /api/admin/projects` 返回 `201 id=7`。
- 导入：`创建 1，更新 0，跳过 27，失败 0`。
- 站点统计：笔记 28 篇；`quickocr` 关联 1 篇（`free-ocr-ledger`），`/api/projects/quickocr` 技术栈返回 `C#/.NET 8/WPF/Tesseract 5`、仓库为真实地址。
- 备案词自检：全文扫描 `hardWords` 与 `downloadPhrases`，命中 0（文中「语言包」「发行包」等表述均未触表）。

## 历史改动记录（v1.3.7，2026-11-17）

- 新增项目 `mailhub`（MailHub，id=6）：`NestJS / React / TypeScript / SQLite / Prisma`，自托管多账号邮件客户端。源码仓库只在本地（`E:\traework\mailhub`），GitHub 未公开，因此 `repo_url` 留空，项目页显示「源码入口暂未提供」。
- 新增 3 篇，素材取自该仓库 `doc/SYNC_POLLING.md`、`doc/SPAM_RULES.md` 与 `PROJECT_STATE.md` §3.23~§3.30：
  - `imap-inbox-flag-not-standard` —— `\Inbox` 不是 IMAP 标准属性（IANA 标注 JMAP only），后台轮询只筛到「已发送」，收件箱从未被打开。
  - `imap-uid-range-star-trap` —— `uidNext:*` 的 `*` 是邮箱最大 UID 且区间无序，守卫 `>` 差一格，每轮把最新那封旧信计成新增。
  - `visibility-scope-single-source` —— 「什么算收件」在列表、计数、清理三处各写一份，附 Prisma `contains` 走 LIKE 的真库对照与 `every` 空关系恒真两个坑。
- 3 篇分别采用「排查时间线」「单行代码解剖」「真库谓词对照」三种骨架，延续 §12.3 的写作约定。

### 验收证据（v1.3.7，2026-11-17）

- 开发库重建：上一轮 `/tmp/dn-dev` 已随 WSL 重启丢失，本轮新建 `DATA_DIR=/tmp/dn-dev` 并全量重导，`/healthz` 返回 `1.3.6`（重建时二进制尚为旧版，升版后重新构建并复测为 `1.3.7`）。
- 项目创建：`POST /api/admin/projects` 依次返回 `id=1..6`，`mailhub` 为 `id=6`；`repo_url` 传空串通过校验（校验仅要求非空时必须是 GitHub 仓库地址）。
- 导入：`创建 27，更新 0，跳过 0，失败 0`。
- 站点统计：笔记 27 篇；`siteguard 4 / superclip 4 / lanfileshare 4 / excelmailmerge 4 / jitijingjiguanli 8 / mailhub 3`（早期 3 篇占位笔记未重建，故前三项各少 1 篇）。
- 浏览器（Edge/Chromium 1280px，`http://172.25.209.179:18090`）：
  - `imap-inbox-flag-not-standard`：`h1` 正确，`h2=8 / table=2 / li=9 / pre=3`；
  - `imap-uid-range-star-trap`：`h2=7 / table=1 / li=9 / pre=4`；
  - `visibility-scope-single-source`：`h2=7 / table=1 / li=2 / pre=7`；
  - 三页 `scrollWidth == clientWidth == 1280`，无横向滚动条。
  - `/projects/mailhub`：技术栈渲染为 `NestJS / React / TypeScript / SQLite / Prisma`，外链数为 0，显示「源码入口暂未提供」。
- 备案词自检：三篇全文扫描 `hardWords` 与 `downloadPhrases`，命中 0；正文里「营销订阅文件夹」等表述均避开了词表。

## 历史改动记录（v1.3.6，2026-11-17）

- 新增项目 `jitijingjiguanli`（记账管理平台，id=5）：`Go/SQLite/Vue 3/Vant`，源码外链 `mylxnet/jitijingjiguanli`。
- 新增 8 篇，素材来自该仓库 `docs/` 下的联动规则、校验摘要、坏账核销、备份、认证修复、静态资源、投资收益口径等方案文档。
- 8 篇分别采用决策对比、核对清单、取证推演、SQL 方案、链路取证、能力矩阵、问题清单、约束推导八种结构（延续 §12.3 的约定）。

### 验收证据（v1.3.6，2026-11-17）

- 项目创建：`id=5 slug=jitijingjiguanli`。
- 导入首轮：`创建 7，更新 0，跳过 16，失败 1` —— 失败项为 `mobile-pdf-preview-pdfjs`，`HTTP 422 内容包含不符合备案边界的表述`，定位到正文含「下载链接」被拦截。
- 改写为「下载按钮的 href」后复跑：`创建 1，更新 0，跳过 23，失败 0`。
- 站点统计：笔记 27 篇；`siteguard 5 / superclip 5 / lanfileshare 5 / excelmailmerge 4 / jitijingjiguanli 8`。
- 浏览器（Chromium 1280px）：项目页技术栈 `['Go','SQLite','Vue 3','Vant']`、外链为真实仓库；8 篇详情页结构各异——`h2` 数量 6/6/8/7/7/7/6/5，表格 1/2/3/2/1/2/3/3，列表 2/1/1/2/2/3/1/1。

## 历史改动记录（v1.3.5，2026-11-17）

- 新增项目 `excelMailmerge`（id=4）：`C#/.NET 8/WPF/ClosedXML`，源码外链 `mylxnet/excelMailmerge`。
- 新增 4 篇：`wpf-invariant-globalization-crash`、`excel-row-probe-infinite-loop`、`wpf-progress-binding-and-ui-thread`、`single-instance-mutex-after-kill`。
- **按用户反馈调整写作方式**：这 4 篇不用同一骨架，分别写成「打包参数对照」「时间线叙事」「症状清单」「时序推演」；约定写入 `doc/PROJECT.md` §12.3。

### 验收证据（v1.3.5，2026-11-17）

- 项目创建：`POST /api/admin/projects` 返回 `id=4 slug=excelmailmerge`。
- 导入：`创建 4，更新 0，跳过 12，失败 0`。
- 站点统计：笔记 19 篇；`siteguard / superclip / lanfileshare` 各 5 篇，`excelmailmerge` 4 篇。
- 浏览器（Chromium 1280px）：项目页技术栈 `['C#', '.NET 8', 'WPF', 'ClosedXML']`、外链为真实仓库；4 篇详情页结构确实不同——`h2` 数量 4/3/5/4，表格出现 1/1/0/0，列表出现 0/0/0/1。

## 历史改动记录（v1.3.4，2026-11-17）

- 录入 SiteGuard 4 篇：`fnos-fpk-install-rejected`、`gateway-base-href-missing-assets`、`reverse-proxy-origin-403`、`baseline-filter-asymmetry`，素材来自该仓库 `doc/DEBUG_FPK.md` 与 `doc/PROJECT_STATE.md` 的踩坑 4/5/6。
- `reverse-proxy-origin-403` 正文首次使用 `/media/siteguard-origin.svg`，把 v1.3.0 建立的配图通道跑通端到端（此前两张图一直没被引用）。

### 验收证据（v1.3.4，2026-11-17）

- 导入：`创建 4，更新 0，跳过 8，失败 0`。
- 用 v1.3.3→v1.3.4 二进制 + 数据库副本起临时实例（`:18084`）：`/healthz` 返回 `version: "1.3.4"`；`/api/notes/reverse-proxy-origin-403` 的 `content_html` 保留 `<img src="/media/siteguard-origin.svg" alt="网关反代下 Origin 与 Host 对不上">`；`/media/siteguard-origin.svg` 返回 `200 image/svg+xml`（2946 字节）。
- 浏览器（Chromium 1280px）：该文详情页插图 `naturalWidth=860 naturalHeight=330`，渲染 `719×276`，比例与原始一致，未被压扁。
- 站点统计：笔记 15 篇；siteguard / superclip / lanfileshare 各 5 篇。

## 历史改动记录（v1.3.3，2026-11-17）

- 录入 LanFileShare 4 篇：`http-listener-to-tcplistener`、`streaming-multipart-1mb-window`、`device-name-to-safe-path`、`ua-model-detection`，素材来自该仓库 `README.md` / `DEVELOPMENT.md` / `RELEASES.md`。
- 项目卡片纠错（走管理 API `PUT /api/admin/projects/:id`）：仓库地址从 `github.com/example/*` 占位改为 `mylxnet/site-guard`、`mylxnet/SuperClip-C`、`mylxnet/LanFileShare`；技术栈分别改为 `Node.js/ESM/单容器自托管`、`C++/Win32 API/CMake`、`C#/WPF/.NET 8`。
- 仓库事实核对来自 GitHub API（`users/mylxnet/repos` 与各仓库 `contents`），不靠猜：`site-guard` 的 `package.json` 显示 `engines.node >=22.13`、唯一依赖 `nodemailer`、MIT。

### 验收证据（v1.3.3，2026-11-17）

- 导入：`创建 4，更新 0，跳过 4，失败 0`（幂等，已存在的 4 篇 SuperClip-C 文章被跳过）。
- `/api/notes` 总数 7 → 11；`/api/projects` 中 `lanfileshare` 的 `note_count` 1 → 5。
- 浏览器（Chromium 1280px）：`/projects/siteguard|superclip|lanfileshare` 详情页分别渲染出 `源码仓库 →` 外链，指向三个真实仓库；技术栈标签与上表一致。
- 4 篇新文章详情页：`h1` 正确，`h2` 分别为 6/5/4/6，代码块 2/2/2/3 正常高亮，表格正常渲染。

## 历史改动记录（v1.3.2，2026-11-17）

- 新增 `content/notes/`：文章为 Markdown + 头部字段（`title`/`slug`/`summary`/`project`/`status`/`published_at`），随代码版本化。
- 新增 `tools/import-notes.sh`：登录管理 API 后逐篇导入，`project` 按项目 slug 解析为 `project_id`；按 slug 幂等，默认跳过已存在，`MODE=update` 覆盖。
- 录入 4 篇文章（均关联 SuperClip-C）：`win32-topmost-silently-dropped`、`sendinput-batch-alt-keyup`、`rc-codepage-resource-garbled`、`favorite-partition-invariant`，素材取自该仓库 `doc/PROJECT_STATE.md` 的真实踩坑记录。
- 选题口径：只写有明确现象、可复现根因、有事后判据的复盘，不写泛泛的"最佳实践"。

### 验收证据（v1.3.2，2026-11-17）

- 导入（对运行中的 `http://localhost:18080`）：`创建 4，更新 0，跳过 0，失败 0`。
- 幂等复跑：`创建 0，更新 0，跳过 4，失败 0`。
- 接口核对：`/api/notes` 总数 3 → 7；`/api/projects` 中 `superclip` 的 `note_count` 1 → 5，4 篇均落在 SuperClip-C 下。
- 浏览器验收（Playwright + WSL 内 Chromium，1280px）：笔记列表 7 条；4 篇详情页均 `h1` 正确、`h2` 各 5 个、代码块与表格正常渲染（如 `win32-topmost-silently-dropped`：`pre=2`、`table=1`、`p=9`）。

## 历史改动记录（v1.3.1，2026-11-17）

- `COOKIE_SECURE` 由布尔改为三档：`auto`（默认）/ `true`(always) / `false`(never)，非法值启动即失败。
- `auth.NewManager` 第二个参数改为 `CookieSecureMode`；新增 `useSecure(c)`：`auto` 下直连 TLS 或请求头 `X-Forwarded-Proto: https` 才加 `Secure`。
- 现场问题定位：用户以 `http://<内网IP>:端口` 访问后台，登录接口 200 但 `/api/admin/me` 立即 401、Cookie 列表为空 —— 根因是 `Secure` Cookie 在 http 下被浏览器丢弃。
- 另一条现场问题：`http://127.0.0.1:18080` 打不开，`http://localhost:18080` 正常 —— WSL2 的 localhost 转发只监听 `::1`，IPv4 回环无监听（已记入踩坑表）。

### 验收证据（v1.3.1，2026-11-17）

- `gofmt -l .` 无输出；`go vet ./...` 无输出；`go test ./... -count=1`：`api`、`auth`、`config`、`render`、`store`、`validate` 全部 ok。
- 新增用例：`TestSessionCookieSecureFollowsScheme`（auto+http 不加 Secure、auto+TLS 加、auto+`X-Forwarded-Proto: https` 加、always 加、never 不加）、`TestLoadCookieSecureMode`（空值/大小写/非法值）。
- 修复前复现（v1.3.0 二进制）：`http://127.0.0.1:18081` 登录后新建笔记 201 成功；`http://172.25.209.179:18082` 登录后 `/api/admin/me` 401、Cookie 列表为空、弹回登录页。
- 修复后验证（v1.3.1 二进制，不设 `COOKIE_SECURE` 走默认 auto，`http://172.25.209.179:18083`）：`/healthz` 返回 `version: "1.3.1"`；登录后 URL 落在 `/admin/notes`；Cookie 为 `dn_session`/`dn_csrf` 且均无 Secure；`POST /api/admin/notes` 返回 201；列表中出现新笔记「HTTP 访问验收」。

## 本轮改动记录（v1.3.0，2026-11-17）

- 图片白名单放宽：`render.go` 中 `assetSRC`（`^/assets/`）改为 `localImageSRC`（`^/(?:assets|media)/`），`AllowRelativeURLs(true)` 保持不变，外链与协议相对地址仍被剔除。
- 新增 `web/public/media/siteguard-origin.svg`、`web/public/media/mailhub-imap-cursor.svg`，随 `public/` 进入 `dist/`，由 `noRouteHandler` 的静态分支直接下发（`http.FileServer` 按扩展名给出 `image/svg+xml`）。
- `.article-markdown img` 增加 `display: block`、`height: auto`、`margin: 24px auto`，解决 SVG 自带宽高属性被压扁的问题。
- `render_test.go` 图片用例改为表驱动：正例 `/assets/demo.png`、`/media/siteguard-arch.svg`；反例 `https://evil.example/track.gif`、`//cdn.example/a.png`、`/static/a.png`。

### 验收证据（v1.3.0，2026-11-17）

- `go test ./internal/render/... -count=1`：7 个用例全绿（含表驱动的内部图片用例）。
- `go vet ./...` 无输出；`go test ./... -count=1` 全仓通过；`go test ./internal/auth/... -count=30` 稳定通过（排除篡改用例的偶发失败）。
- `bash scripts/build.sh`：版本一致性校验通过，前端 63 个模块构建成功，`app/build/dev-notes` 产出（v1.3.0），`dist/media/` 内含两张 SVG。
- 接口冒烟（临时库 + `:18081`）：`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.3.0"}`；`/media/mailhub-imap-cursor.svg` 返回 `200 image/svg+xml`（3280 字节）；正文净化结果保留 `<img src="/media/mailhub-imap-cursor.svg">`，外链图片 `src` 被剔除。
- 浏览器验收（Playwright + WSL 内 Chromium）：1280px 下插图渲染 719.2×299.4（比例 2.402，与原始 860×358 一致）；390px 下渲染 358.0×149.0（比例仍为 2.402），未被压扁。
- 页脚遮挡复检：桌面页脚「v1.3.0 by Mr lin」矩形 `[991,847,1098,870]` 与主题按钮 `[1152,840,1260,880]` 不重叠；390px 下版本行 `y=801–824` 与按钮 `y=840–880` 不重叠。
- 两张 SVG 已按同一视觉规范重绘：860 宽、白底圆角卡片、`#d97757` 强调色、`hot`/`ok` 区分陷阱与解法。

**待办**：两张插图尚未被文章正文引用（正文在数据库里，仓库无 Markdown 源文件），需在录入文章时补 `![说明](/media/xxx.svg)`。

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
| D20 | `COOKIE_SECURE` 默认 `auto`，按请求协议判断 | 兼顾两类部署：HTTP 直连能登录，反代终结 TLS 时仍保留 Secure；代价是判断依赖 `X-Forwarded-Proto`，代理未透传时按 HTTP 处理（宁可可用，不牺牲登录） |
| D21 | 加字段走 `user_version` + 探列式 `ALTER TABLE`，不重建表 | 存量库有真实内容，重建要搬数据且易丢；探列让迁移可重跑（中断后再次启动不会因「列已存在」失败）。代价是迁移代码里保留 v1 历史建表语句，读代码要按版本顺序看 |
| D22 | 项目说明存 Markdown 原文、前台只给 `description_html` | 与笔记正文同一口径：前端不接触原始输入，渲染与净化只在服务端一次；列表接口置空说明原文，避免长文随列表重复下发 |
| D23 | 封面图只收站内路径、不做上传接口 | 站点规模小、配图走仓库 `web/public/media/`，无上传即无文件写入与鉴权面；代价是运营时要手工放文件再填路径，故后台加缩略图预览兜底填错路径 |
| D24 | 三处列表统一 6 条/页，分页条抽成 `Pager.vue` | 后台与前台同一套翻页手感，用户不用在两处适应不同节奏；页码折叠规则只维护一份，避免前后台分页行为漂移。代价是后台表格一屏放得下 10 行，6 条会多翻一页，属可接受 |
| D25 | 后台项目列表也在服务端分页，而非前端切片 | 与笔记列表同一套 `LIMIT/OFFSET + COUNT` 写法，`total` 由 SQL 统计，删除后页码回退逻辑通用；代价是项目量少时多一次 COUNT 查询，可忽略 |
| D26 | 测试凭据走构建变量 + 独立测试构建脚本，不硬编码进组件 | 登录页是公开页面，硬编码凭据会随正式产物泄露；用 `VITE_TEST_LOGIN=1` 开关把测试便利限制在验收包内，`build.sh` 再加防误用校验。代价是测试包与正式包是两份产物，必须明确区分用途 |
| D27 | 站点配置用 `settings` 键值表，不往 `admin` 表加列 | 备案号是站点属性而非管理员属性；键值表后续加配置项无需迁移表结构。代价是 `schemaVersion` 升到 3，前台多一次极轻量的公开查询（整站只拉一次） |
| D28 | 封面图支持上传，但只收 png/jpg/webp、禁 SVG，文件存 `DATA_DIR/media` | 解决「换图要重打包」痛点，同时把文件写入面压到最小：扩展名 + 内容嗅探双校验、随机名防穿越、2 MB 上限；SVG 可内嵌脚本故不支持。代价是备份要带上 media 目录 |
| D29 | Markdown 编辑器自建组件，预览仍走服务端接口 | 依赖从简（不引第三方编辑器），正文原文与既有渲染/净化管线完全一致；用防抖与请求序号解决实时预览的接口压力与竞态。代价是工具栏交互代码要自己维护 |
| D30 | Docker 镜像在容器内编译，只推 amd64，用 legacy builder 推送 | 不把 WSL 本地二进制硬塞进镜像；`DOCKER_BUILDKIT=0` 避开 Docker 29 BuildKit 附带的 OCI 空 manifest（ACR 拒收），产物为不含 attestation 的 OCI image manifest，云端拉取运行已验证。代价是构建比直接 COPY 二进制慢、legacy builder 已废弃、只覆盖 amd64 |

## 已知限制

- mailhub 源码只在本地 `E:\traework\mailhub`，GitHub 未公开（`api.github.com/repos/mylxnet/mailhub` 返回 404），项目卡外链留空，页面显示「源码入口暂未提供」
- 开发库放在 `/tmp/dn-dev`，重启即丢；重建靠 `content/notes/` + 管理 API 全量重导，项目行（名称、摘要、技术栈）只存在于库中，重建时需手工补录
- SuperClip-C、LanFileShare 的许可证限制分发，站点只能放源码外链
- 前台无评论、无访问统计（方案 A 范围外）
- 第二版原型仅归档，不参与正式构建；是否删除待确认
- 后台编辑器为纯文本框 + 预览，暂无工具栏、快捷键与图片上传（图片只能用手写站内路径 `/assets/` 或 `/media/`）
- 登录限流计数存活于内存，多实例部署时各实例独立计数
- 文章 slug 需手填，未做中文标题自动转拼音
- Dockerfile、Compose 与 1Panel 部署尚未实现
- 6 个项目的真实 GitHub 地址、5 篇笔记的真实内容尚未最终确认
- 原型与构建复用了 `lxsyzd` WSL，未建项目专属隔离环境
- 插图固定 860 宽、白底深字，暗色主题下呈白色卡片；390px 下缩到约 358px，图内小字偏小，暂未做暗色适配与移动端简化版

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
| 正文里 `/media/xxx.svg` 的图片被净化后 `src` 清空 | 图片白名单仍是 `^/assets/`，`/media/` 不在允许目录内 | 白名单放宽为 `^/(?:assets|media)/`，测试补 `/media/` 正例 | 新增资源目录时要同步检查净化白名单与静态下发两处，只加文件不改白名单等于没生效 |
| 正文 SVG 插图在窄容器下被压扁变形 | SVG 自带 `width`/`height` 属性，CSS 只写 `max-width: 100%` 不会让高度自适应 | 补 `display: block` 与 `height: auto` | 替换元素（img/svg）缩放宽度的同时必须显式放开高度，否则比例丢失 |
| 页脚「v1.3.0 by Mr lin」被右下角主题按钮压住 | 主题按钮 `position: fixed`，页脚没有为它预留空间 | 页脚右内边距留 132px，≤600px 改为底部留白 76px | 固定定位元素会覆盖正常流内容，页脚类区域必须预留避让空间 |
| 文章导入失败但看不出原因（HTTP 422） | 正文含备案词表中的短语（本次是「下载链接」），校验在服务端拦截 | 改写为「下载按钮的 href」；导入工具已回显服务端 message，先看它再改正文 | 写作前先扫一遍禁词：售卖/下单/支付/收费/会员/广告/赞助/捐赠/破解/激活码/代购/推广，以及「下载安装包/下载地址/下载链接/网盘地址/提取码」等短语 |
| 主仓库 `app/`、`doc/`、`tools/`、`scripts/` 突然"消失"，但 `git status` 不报删除 | 误在本仓库执行了 `git sparse-checkout set --no-cone '/*.md'`（本想对 `.tmp_lfs` 克隆执行，`cd` 失败后落到了父仓库），149 个文件被标 `skip-worktree`，工作区只剩根目录 md | `git sparse-checkout disable` 恢复；核验 `Test-Path` 与 `git ls-tree -r HEAD` 文件数（152）一致 | 在仓库内做临时克隆/实验前先确认当前目录；`skip-worktree` 会让缺文件看起来"没改过"，排查时先查 `git ls-files -v` 的 S 标记 |
| 用内网 IP 以 http 访问后台，登录后立刻被踢回登录页、无法新增或修改 | `COOKIE_SECURE` 默认 `true`，Cookie 带 `Secure`，HTTP 下被浏览器丢弃 → `/api/admin/me` 401 | `COOKIE_SECURE` 改为三档，默认 `auto` 按请求协议判断（含 `X-Forwarded-Proto`） | Secure Cookie 只在 HTTPS 与 localhost 生效，本机测试通过不代表换地址可用；会话类问题先看 Cookie 是否落盘 |
| `TestParseRejectsTamperedSession` 时好时坏 | 用例改写 base64url 末位字符，末位含未使用位，解码后字节可能完全相同，篡改等于没改 | 改为改写签名段首字符，保证解码字节必变 | 校验签名的用例不能靠「改一个字符」想当然，要确认改动真的改变了字节 |
| 给 `project` 加列后若只改建表语句，老库查询报 `no such column: description` | `CREATE TABLE IF NOT EXISTS` 对已存在的表不生效，老库结构不会跟着变 | 加 v2 迁移：`PRAGMA table_info` 探列 + `ALTER TABLE ... ADD COLUMN`，并补 `TestMigrateV1Database` 覆盖 v1 老库升级 | SQLite 改表结构只有 `ALTER TABLE` 真生效，建表语句只服务于新库 |
| 加列后 `sql: expected 13 destination arguments in Scan, not 11` | 列常量 `projectColumns` 加了列，`scanProject` 的 `Scan` 实参没同步 | 列常量、`Scan` 实参、INSERT 占位符、UPDATE 赋值四处同批改，顺序保持一致 | 项目字段的 4 个 SQL 触点必须一起改，漏一处只会在运行时炸 |
| 卡片封面用 `object-fit: cover` 把示意图上下文字裁掉 | `cover` 按容器比例裁剪，而技术示意图原图比例 2.6、卡片容器 16:9 | 卡片改 `object-fit: contain`（保整图、留白用卡片底色），详情页封面按原比例自然高度 | 技术插图优先保证内容完整，定比容器配 `contain`；照片类素材才用 `cover` |
| 详情页说明里的 `h2` 比「项目说明」这一节标题还大 | `.article-markdown h2` 是 1.5rem，而节标题只有默认 h2 尺寸 | 说明容器加 `.project-description`，把内嵌 `h2/h3` 降到 1.15rem / 1rem，并让首个子元素不留上边距 | 嵌进区块的 Markdown 不能直接沿用文章正文的字号阶梯，要和所在区块的标题层级对齐 |
| 后台分页条下面又冒出一行「共 15 篇」，与分页条里的「共 15 篇」重复 | 先加了独立总条数行，后来才加分页条，两处都在显示总数 | 独立总条数行改为仅在 `totalPages === 1` 时渲染（此时没有分页条）；多页时交给分页条 | 同一信息只允许一个权威展示位；加新组件时要回头查旧文案是否重复 |
| 后台笔记首页断言拿到 0 行，但分页信息已经显示 15 篇 | 表格 `rows.count()` 在接口返回前执行，随后分页条才渲染，属于各断言之间有时序差 | 断言前先 `wait_for_selector("tbody tr")` 等真实行出现 | 浏览器断言要等具体元素，不能靠 `networkidle` 之后立刻数节点 |
| 想让登录框预填密码，但直接把密码写进 Vue 组件会随正式包一起发布 | 登录页是公开静态资源，任何访客都能从页面/源码读到管理员密码 | 凭据改为构建变量注入，只有 `scripts/build-test.sh` 打出的包预填；正式包 `grep` 计数为 0，并在 `build.sh` 加开关防误用 | 涉及凭据的「测试便利」必须按产物隔离，不能只靠记得改回来 |
| 切换测试包后用户反馈「笔记和项目找不到了」 | 启动登录验收服务时给了新空库 `DATA_DIR=/tmp/dn-login` 并 `pkill` 掉原服务，用户看到的是空站；真实开发库是 `/tmp/dn-dev/notes.db`（8 项目 / 31 笔记），数据从未丢失 | 先用只读 sqlite 核对两个库的行数确认根因；备份真实库后用 `DATA_DIR=/tmp/dn-dev` 重启，前台恢复 30 篇已发布 + 8 个项目 | 启动服务前必须确认 `DATA_DIR` 指向哪份数据；杀旧服务前先记下它挂载的库，不能随便换新目录充测试 |
| 切回真实库后，测试包预填的 `admin12345678` 登录失败（401） | 真实库的管理员密码此前在后台改过，与测试包预填的默认测试密码不同；bcrypt 哈希不可反推 | 前台内容不受影响；后台改用用户自己的密码登录，或经用户确认后重建测试包时通过 `VITE_TEST_PASSWORD` 注入正确密码 | 预填凭据只对“用同一套初始密码创建的库”有效；换库后要重新核对凭据，不要误判为登录功能坏了 |
| 重置密码后登录一直跳回登录页 | 一次性哈希脚本命名 `_tmp_hash.go`，Go 构建会忽略下划线开头的文件，`go run` 报 `no Go files in ...`；脚本没校验退出码，把空哈希写进了 `admin.password_hash` | 改名 `tmp_hash_gen.go`，写入前断言哈希长度 60、`UPDATE` 影响行数为 1；重启后用浏览器验证预填登录成功（31 篇 / 8 个项目正常） | 一次性 Go 脚本别用 `_` 开头；写库脚本必须校验生成物非空且影响行数，失败即中止 |
| 390px 下编辑页横向溢出 34px | `.editor-row` 固定三列，日期控件固有宽度约 159px，挤出了窄屏容器 | ≤640px 改为单列并给表单控件加 `min-width: 0`，复测溢出为 0 | 响应式自检不能只看主内容区，要逐个元素找 `getBoundingClientRect().right > innerWidth` 的真凶 |
| 镜像构建报「主模块不包含 dev-notes-platform/app 包」 | `go.mod` 就在 `app/` 里，Dockerfile 却把 `app/` 复制成 `/src/app`，形成嵌套模块被 Go 排除 | 改为 `COPY app/ ./` + `go build .`，模块根直接是 `/src` | 复制源码目录前先确认 `go.mod` 在哪一层；嵌套 go.mod 会让子目录被当成独立模块 |
| 推送阿里云报 `unknown manifest class for application/vnd.oci.empty.v1+json` | Docker 29 默认 `docker build` 走 BuildKit，层已全部 Pushed 但附带 OCI 空 manifest（attestation），阿里云 ACR 不支持该 manifest class，最后一次推送被拒 | 推送脚本设 `DOCKER_BUILDKIT=0` 走 legacy builder；Dockerfile 去掉 `--platform=$BUILDPLATFORM/$TARGETPLATFORM` 与 `TARGETOS/TARGETARCH`，固定 `GOOS=linux GOARCH=amd64`（legacy builder 不识别这些自动变量）；重推 RC=0 | registry 报错先看具体 media type，别把「层已 Pushed」当成成功；单架构推送要确认 builder 会不会附带 attestation，再选兼容 registry 的方式 |
| 凭印象给出数据卷属主 `chown 100:100`，与容器实际用户不符 | alpine 中 gid 100 已被系统占用，`addgroup -S app` 实际分到 gid 101，`adduser -S -G app app` 分到 uid 100 | 实测 `docker run --rm --entrypoint sh <镜像> -c "id app"` 得到 uid=100 / gid=101；部署文档与验收记录统一改为 `chown -R 100:101 ./data` | 非 root 容器的 uid/gid 必须实测，不能按「第一个普通用户是 100:100」推断；属主设错会导致数据卷只读、容器启动失败 |

### 密码重置记录（2026-11-17，经用户同意）

- 操作对象：真实开发库 `/tmp/dn-dev/notes.db` 的 `admin` 表；操作前备份为 `notes.db.bak-before-pwd-reset`。
- 结果：管理员 `admin` 的密码重置为默认测试密码，测试包预填可一键登录；`bcrypt.DefaultCost` 生成 60 字符哈希，更新行数 1。
- 验证：`/admin/login` 不输入直接点登录 → `/admin/notes`，后台 31 篇笔记 / 8 个项目，控制台错误 0。

## 待办清单

1. ~~确认 mailhub 正确仓库地址，或决定移除~~ —— 已确认：源码在本地 `E:\traework\mailhub`，GitHub 未公开，项目卡外链留空
2. 确认 8 个项目的真实 GitHub 地址与许可证状态（7 个已有真实地址；mailhub 无公开仓库；ExcelToHub 无 LICENSE 文件，QuickOcr 的 README 声明 MIT 但仓库无 LICENSE 文件——两者都只放外链，站点不做分发）
3. ~~三篇早期占位笔记~~ —— 开发库重建后已不存在（`lan-discovery` / `sqlite-pure-go-driver` / `clipboard-history-design` 不再入库）
4. `mailhub-imap-cursor.svg` 仍未在任何文章中被引用（`siteguard-origin.svg` 已随 `reverse-proxy-origin-403` 上线）；本轮 3 篇纯文字，没有插图
5. `tools/import-notes.sh` 目前只处理新建与更新，未做正文配图上传（插图仍走 `web/public/media/`）
6. 项目元数据（名称/摘要/技术栈/说明/封面路径）只存在于数据库，重建开发库要手工补录；考虑加一份 `content/projects.json` + 入库脚本，避免下次重来
7. Section 5：Docker 构建链已就绪，`1.4.6` / `1.4.7` / `1.4.8` / `1.4.9` / `1.4.10` / `1.4.11` 镜像均在阿里云，`1.4.11` 已通过本地容器冒烟。待办是 1Panel 上机部署与浏览器验收（需 ICP 备案通过后再开放）
8. 每个 Section 完成后升小版本、同步文档并提交

## 版本一致性检查

当前需要保持一致的五处：

| 位置 | 当前值 |
|---|---|
| `app/VERSION` / `app/main.go` | `1.4.11` |
| `web/package.json` / `web/package-lock.json` | `1.4.11` |
| `web/src/version.ts` | `APP_VERSION = '1.4.11'` |
| `web/src/views/AboutView.vue` / 后台侧栏 / 登录页 | `{{ VERSION_LINE }}` → `v1.4.11 by Mr lin` |
| `doc/*.md` / `README.md` / `CHANGELOG.md` / `deploy/docker-compose.yml` 镜像 tag | `v1.4.11` |

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
cd app && DATA_DIR=/mnt/e/app/网站备案/dev-notes-platform/deploy/data \
  SESSION_SECRET=0123456789abcdef0123456789abcdef \
  HTTP_ADDR=127.0.0.1:8080 go run .
cd web && npm run dev          # Vite 已把 /api 代理到 127.0.0.1:8080

# 启动已打包的二进制做验收：挂 deploy/data 的副本
DATA_DIR=/mnt/e/app/网站备案/dev-notes-platform/deploy/data HTTP_ADDR=:18091 \
  SESSION_SECRET=0123456789abcdef0123456789abcdef ./app/build/dev-notes

# 主题按钮断点回归（WSL 内 playwright + chromium，输出 JSON）
BASE=http://127.0.0.1:5173 ADMIN_USER=admin ADMIN_PASS=<当前密码> \
  python3 test/theme_button_check.py
```

`DATA_DIR` 决定看到哪份内容：验收用的真实开发库副本在 `deploy/data`（`notes.db` + `media/`，8 个项目 / 31 篇笔记）；临时库（如 `deploy/data-login`）只有造数数据。**WSL 重启会清空 `/tmp`，原先的 `/tmp/dn-dev` 已不再使用。**重启前先确认旧服务挂载的目录，不要直接换新目录，否则会表现为「笔记和项目找不到」。管理员密码若在后台改过，测试包预填的默认密码会失效，需用当前密码登录或重建测试包时通过 `VITE_TEST_PASSWORD` 注入。**2026-11-17 已按用户同意把真实库密码重置为默认测试密码，测试包预填可直接登录。** 运行时上传的图片写入 `DATA_DIR/media/`，备份时要一并带上；构建产物 `web/public/media/` 里的预置图仍随二进制发布。

详细说明见 `doc/DEPLOY.md` 第二节与 `doc/PROJECT.md` 第七节。
