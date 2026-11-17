# 项目状态 · 我的开发笔记

> v1.0.3 · by Mr lin · 2026-11-17

## 当前版本

v1.0.3 · Section 1 已完成：Go 骨架、SQLite、数据模型、Store、公开 API 与测试已跑通；下一阶段 Section 2。

## Section 1 实施记录（2026-11-17）

- 建立 Go 模块 `app/go.mod`，模块路径 `dev-notes-platform`，Go 1.27.0。
- 完成配置加载与启动校验：`SESSION_SECRET`、`ADMIN_INIT_USERNAME`、`ADMIN_INIT_PASSWORD`、`DATA_DIR`、`HTTP_ADDR` 等。
- 完成 SQLite 连接与迁移：WAL、`busy_timeout`、外键、`user_version`，建立 project / note / admin 表与索引。
- 完成模型与 Store：项目、笔记、管理员的查询与增删改；已发布内容过滤；唯一约束与 `ErrNotFound` 处理。
- 完成公开 API：`/healthz`、`/api/projects`、`/api/projects/:slug`、`/api/notes`、`/api/notes/:slug`。
- 完成管理员种子：首次启动按环境变量写入 bcrypt 哈希，已存在则跳过。
- 测试结果：
  - `go test ./... -count=1`：`internal/api`、`internal/config`、`internal/store` 全部通过。
  - 启动冒烟：`/healthz` 返回 `{"db":"ok","status":"ok","version":"1.0.3"}`；空库下 `/api/notes` 与 `/api/projects` 返回空列表。
- 构建结果：`CGO_ENABLED=0` 静态二进制约 27 MB，可正常启动与优雅退出。
- 未完成：正式前台 `web/`、embed 静态资源、后台登录/管理 API、内容校验、Docker 与 1Panel 部署。

## 设计收尾记录（2026-11-17）

- 完成两版界面原型并做浏览器验收：第一版 `prototype/` 采用柔和圆角与笔记驱动布局；第二版 `prototype-v2/` 为编辑部极简方向。
- 用户对比后确认采用第一版；第二版归档，不参与正式构建。
- 完成 `doc/PROJECT.md` 详细技术方案，覆盖架构、数据模型、API、认证、校验、测试、构建部署与开发计划。
- 版本统一升至 `v1.0.2`：文档、`README.md`、`CHANGELOG.md`、原型工程与界面页脚/侧栏均已同步，并重新构建 `prototype/dist`。
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

## 已知限制

- mailhub 仓库地址返回 404，GitHub 搜索无匹配，暂不纳入展示
- SuperClip-C、LanFileShare 的许可证限制分发，站点只能放源码外链
- 前台无评论、无访问统计（方案 A 范围外）
- 第二版原型仅归档，不参与正式构建；是否删除待确认
- 正式前台 `web/` 尚未开始，当前只有 `prototype/` 原型
- 后台登录、管理 API、内容校验、Markdown 渲染尚未实现
- 静态资源 embed、Dockerfile、Compose 与 1Panel 部署尚未实现
- 6 个项目的真实 GitHub 地址、5 篇笔记的真实内容尚未最终确认
- 原型与构建复用了 `lxsyzd` WSL，未建项目专属隔离环境；正式编码前需留意依赖冲突

## 踩坑记录

| 现象 | 根因 | 修复 | 教训 |
|---|---|---|---|
| 通过 `wsl bash -lc '... &'` 启动的静态服务随即退出 | WSL 会话结束时后台进程被回收 | 在同一命令内 `sleep 2` 并用 `curl` 验证，或使用长期运行的服务管理方式 | 后台服务启动后必须做存活校验，不能只看命令退出码 |
| PowerShell 双引号包裹 WSL/Bash 命令时变量被提前展开 | PowerShell 先解析 `$HOME`、`$P` 等变量 | 使用完整路径、避免嵌套变量，必要时转义 | 跨 PowerShell 与 Bash 的命令要按层校验引号与变量 |
| `go build -ldflags "-s -w -X main.version=..."` 被拆成多个参数 | PowerShell、Bash、Go 三层引号嵌套 | 冒烟构建先省略 ldflags，使用 `main.version` 默认值；正式脚本用单引号包裹完整 ldflags | 跨层传参时先用最小命令验证，再拼复杂参数 |
| 后台启动命令把 `cd && go build && 服务` 整段放入后台，导致 `curl` 在构建完成前执行 | `&` 作用于整个 `&&` 链 | 构建与服务分两条命令执行 | 启动冒烟前先确认二进制已生成 |
| 路由切换后立即截图，拿到上一页内容 | Vue 过渡 `mode="out-in"` 需要约 200ms | 截图前等待 1 秒 | 浏览器验收要等待渲染稳定，不能只看 URL |

## 待办清单

1. 确认 mailhub 正确仓库地址，或决定移除
2. 确认 6 个项目的真实 GitHub 地址与许可证状态
3. 确认 5 篇笔记的真实内容来源
4. Section 2：正式前台页面 + Go embed 静态资源
5. Section 3：后台管理 + 认证 + 前置校验
6. Section 4：真实内容录入
7. Section 5：Docker + 1Panel 部署 + 浏览器验收
8. 每个 Section 完成后升小版本、同步文档并提交

## 版本一致性检查

当前需要保持一致的四处：

| 位置 | 当前值 |
|---|---|
| `app/VERSION` / `app/main.go` | `1.0.3` |
| `prototype/package.json` | `1.0.3` |
| `prototype/` 界面页脚与后台侧栏 | `v1.0.3 by Mr lin` |
| `doc/*.md` / `README.md` / `CHANGELOG.md` | `v1.0.3` |

正式实现后，构建脚本需通过 ldflags 注入版本号并与 `app/VERSION` 校验。

## 构建与发布命令

```bash
# 单元与接口测试
cd app
go test ./... -count=1

# 静态二进制
CGO_ENABLED=0 go build -trimpath -o build/dev-notes .

# 本地启动（示例，密码仅用于本地）
SESSION_SECRET=0123456789abcdef0123456789abcdef \
ADMIN_INIT_USERNAME=admin \
ADMIN_INIT_PASSWORD=admin12345678 \
DATA_DIR=./data \
HTTP_ADDR=127.0.0.1:8080 \
./build/dev-notes
```

详细说明见 `doc/DEPLOY.md` 第二节与 `doc/PROJECT.md` 第七节。
