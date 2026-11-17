# 我的开发笔记

> 个人开发记录站点 · Vue 3 + TypeScript + Go + Gin + SQLite · v1.0.2 by Mr lin

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

当前为设计收尾阶段：第一版界面已确认，详细技术方案已完成，实现尚未开始。

- 设计方案：`doc/DESIGN.md`
- 技术方案：`doc/PROJECT.md`
- 部署手册：`doc/DEPLOY.md`
- 项目状态与待办：`doc/PROJECT_STATE.md`

```bash
# 实现完成后
bash scripts/build.sh     # 在 lxsyzd WSL 中构建
```

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
├── app/                  Go 主程序
├── web/                  Vue 前台与管理界面
├── scripts/              构建脚本
└── test/                 测试
```

## 文件说明

- 设计方案：`doc/DESIGN.md`
- 详细技术方案：`doc/PROJECT.md`
- 部署手册：`doc/DEPLOY.md`
- 项目状态与待办：`doc/PROJECT_STATE.md`
- 版本变更：`CHANGELOG.md`
