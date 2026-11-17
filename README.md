# 我的开发笔记

> 个人开发记录站点 · Vue 3 + TypeScript + Go + Gin + SQLite · v1.0.3 by Mr lin

记录个人学习软件开发过程中的技术笔记、问题排查与项目复盘。前台展示练习项目与笔记，后台供本人维护内容。站点用于甘肃个人 ICP 备案，内容严格限定为个人非经营性信息分享。

## 核心特性

| 特性 | 说明 |
|---|---|
| 笔记驱动 | 首页以笔记为主体，项目以索引形式呈现 |
| 单管理员后台 | 无注册、无评论，仅本人维护内容 |
| 源码外链 | 站内不托管安装包，仅链接 GitHub 仓库 |
| 备案友好 | 页脚预留备案号位置，经营性表述前置拦截 |
| 单容器部署 | Go embed 静态资源，1Panel Compose 一键部署 |

## 快速开始

当前进度：第一版界面已确认，Section 1（Go 后端骨架、SQLite、公开 API）已完成；正式前台与后台将在 Section 2–3 接入。

- 设计方案：`doc/DESIGN.md`
- 技术方案：`doc/PROJECT.md`
- 部署手册：`doc/DEPLOY.md`
- 项目状态与待办：`doc/PROJECT_STATE.md`

```bash
# 后端测试与启动（在 lxsyzd WSL 中执行）
cd app
go test ./... -count=1
CGO_ENABLED=0 go build -trimpath -o build/dev-notes .

SESSION_SECRET=0123456789abcdef0123456789abcdef \
ADMIN_INIT_USERNAME=admin \
ADMIN_INIT_PASSWORD=admin12345678 \
DATA_DIR=./data HTTP_ADDR=127.0.0.1:8080 ./build/dev-notes
```

访问 `http://127.0.0.1:8080/healthz` 应返回 `{"status":"ok","db":"ok","version":"1.0.3"}`。

## 界面原型

第一版原型已确认，源码与可直接打开的构建产物在 `prototype/`；第二版候选归档在 `prototype-v2/`，不参与正式构建。

```bash
cd prototype
npm install
npm run dev
```

## 项目结构

```
dev-notes-platform/
├── README.md             本文件
├── CHANGELOG.md          逐版本变更记录
├── CODEBUDDY.md          工程约定
├── doc/
│   ├── DESIGN.md         设计方案
│   ├── PROJECT.md        详细技术方案
│   ├── DEPLOY.md         部署手册
│   └── PROJECT_STATE.md  项目状态与交接
├── prototype/            已确认的第一版界面原型
├── prototype-v2/         第二版候选（归档）
├── deploy/               容器编排
├── app/                  Go 后端（Section 1 已完成）
├── web/                  Vue 前台与管理界面（待 Section 2–3）
├── scripts/              构建脚本
└── test/                 测试
```

## 文件说明

- 设计方案：`doc/DESIGN.md`
- 详细技术方案：`doc/PROJECT.md`
- 部署手册：`doc/DEPLOY.md`
- 项目状态与待办：`doc/PROJECT_STATE.md`
- 版本变更：`CHANGELOG.md`
