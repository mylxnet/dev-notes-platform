# 设计方案 · 我的开发笔记

> v1.0.3 · by Mr lin · 2026-11-17

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
├── web/                       # Vue 前台 + 管理界面源码
│   ├── src/
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
6. 页脚：备案号位置（备案通过后显示）、`v1.0.3 by Mr lin`

### 后台（单管理员）

1. 登录：`/admin` 路径，账号密码校验
2. 项目：新增、编辑、排序、状态（已发布 / 草稿 / 隐藏）
3. 笔记：新增、编辑、Markdown 正文、关联项目、状态
4. 账号：修改管理员密码
5. 前置校验：保存时拦截经营性词汇与敏感词，命中即拒绝并给出具体提示
6. 发布规则：草稿与隐藏状态不出现在前台

### 界面方案确认（v1.0.2）

| 方案 | 视觉方向 | 结论 |
|---|---|---|
| 第一版 `prototype/` | 柔和大留白、圆角卡片、轻量现代、橙色点缀 | **确认采用**，作为正式界面基线 |
| 第二版 `prototype-v2/` | 编辑部极简、超大标题、细线网格、编号索引 | 归档留档，不参与构建 |

两版原型均完成桌面与移动端浏览器验收，控制台 0 错误。第一版更耐看、与个人笔记站调性更一致。

## 六、数据模型

```
project
  id, name, slug, summary, tech_stack, repo_url, sort_order,
  status(published|draft|hidden), created_at, updated_at

note
  id, title, slug, summary, content(markdown), project_id,
  status(published|draft|hidden), published_at, created_at, updated_at

admin
  id, username, password_hash, updated_at
```

字段约束、索引与迁移策略见 `doc/PROJECT.md` 第五节。

## 七、API 设计

| 方法 | 路径 | 用途 | 鉴权 |
|---|---|---|---|
| GET | `/api/notes` | 笔记列表（仅已发布） | 无 |
| GET | `/api/notes/:slug` | 笔记详情 | 无 |
| GET | `/api/projects` | 项目列表（仅已发布） | 无 |
| GET | `/api/projects/:slug` | 项目详情 | 无 |
| POST | `/api/admin/login` | 登录 | 无 |
| POST | `/api/admin/logout` | 退出 | 是 |
| GET/POST/PUT/DELETE | `/api/admin/projects` | 项目管理 | 是 |
| GET/POST/PUT/DELETE | `/api/admin/notes` | 笔记管理 | 是 |
| PUT | `/api/admin/password` | 修改密码 | 是 |

请求/响应结构、错误码与校验规则见 `doc/PROJECT.md` 第五节。

## 八、安全与容错

- 密码用 bcrypt 哈希，明文不落库、不进日志
- 会话 Cookie 设置 HttpOnly、SameSite=Lax；后台接口统一鉴权中间件
- SQLite 使用参数化查询，禁止字符串拼接
- Markdown 渲染禁用原始 HTML，防止存储型 XSS
- 前置校验在服务端执行，不依赖前端
- 数据库文件放 `/app/data/`，容器重启不丢数据
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

1. Section 1：Go 骨架 + SQLite + 数据模型 + 基础 API
2. Section 2：前台页面 + 构建产物 embed
3. Section 3：后台管理界面 + 认证 + 前置校验
4. Section 4：真实内容录入（6 个项目 + 5 篇笔记）
5. Section 5：Docker 封装 + 1Panel 部署 + 浏览器验收

任务拆分、产出与验收标准见 `doc/PROJECT.md` 第九节。

## 十二、变更记录

| 版本 | 日期 | 变更 |
|---|---|---|
| v1.0.3 | 2026-11-17 | 完成 Section 1：Go 骨架、SQLite 迁移、数据模型、Store、公开 API、健康检查与测试 |
| v1.0.2 | 2026-11-17 | 确认第一版界面方案；新增详细技术方案 `doc/PROJECT.md`；同步版本与文档 |
| v1.0.1 | 2026-11-17 | 新增 CODEBUDDY.md 工程约定文件；完成两版界面原型 |
| v1.0.0 | 2026-11-17 | 初始设计方案 |
