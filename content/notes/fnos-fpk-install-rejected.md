---
title: 报错只有一句话，怎么找出真正的原因
slug: fnos-fpk-install-rejected
summary: 应用中心装包失败，只回一句「应用包格式不符合系统版本要求」。用官方裸模板加单变量二分，逐条排除十一个假设，最后锁定两个 wizard 文件——它们和"格式"没有任何字面关系。
project: siteguard
status: published
published_at: 2026-11-17T10:00:00Z
---

## 现象

把 Node.js 服务打成飞牛 fnOS 的原生应用包（`.fpk`），应用中心安装时只回一句：

```
应用包格式不符合系统版本要求
```

没有行号、没有字段名、没有校验项。同一套 `fnpack` 打出来的官方模板包却能装——说明问题在包里，而不是工具链。

## 排查：裸模板 + 单变量

这类"泛化报错"最忌讳整包盲改。做法是拿官方裸模板 `fnpack create testapp` 当基线，一次只加一个变量：

| # | 变量 | 结果 |
|---|---|---|
| 1 | 原始包（`os_min_version=1.1.3100`、`checkport=false`、含 wizard、ui 为网关 iframe） | 报错 |
| 2 | 删掉 `os_min_version` 与 `checkport` | 同样报错 |
| 3 | 改到 WSL 内打包（LF、保留可执行位） | 同样报错 |
| 4 | 官方裸模板 | **能装**（空壳，打开即停用属预期） |
| 5 | 裸模板 + 中文 `desc` / `display_name` | 能装 |
| 6 | 裸模板 + `version=0.7.1` | 能装 |
| 7 | 裸模板 + 项目 manifest + 项目 ui/config | 打包时报 entry name 不一致（构造 bug，作废） |
| 8 | 裸模板 + 单 share（`siteguard/data`） | 对照文档格式合法 |
| 9 | 裸模板 + `username` / `groupname` | 对照文档格式合法 |
| 10 | 抽包核对 manifest / resource / privilege / framework | 均符合规范，`checksum` 为工具自动添加 |
| 11 | 逐文件 diff 裸模板 vs 项目包 | **唯一结构性差异：裸模板的 `wizard/` 是空目录，项目包多了 `wizard/install`、`wizard/uninstall`** |
| 12 | 裸模板 + 这两个 wizard 文件 | 报错复现 |

第 11 步是关键：前面十次都在改"看起来像格式"的字段，而真正的差异是两个**内容型文件**——`wizard/install`（内容为空数组 `[]`）和 `wizard/uninstall`（一个卸载开关）。

## 根因

fnOS 1.x 的校验器对 `wizard/` 目录内容敏感：只要这两个文件存在，无论 manifest 怎么写，都会走到"包格式不符合"这条泛化分支。这也解释了为什么前一版把 `os_min_version`、`checkport` 都删了仍然报同样的错。

## 修复

删除这两个文件，让 `wizard/` 回归空目录。

顺手确认了删除不会伤到卸载逻辑：

```bash
# 卸载脚本本来就用默认值兜底，wizard 表单不存在时变量为空 → 默认保留数据
wizard_purge="${wizard_purge:-}"
```

也就是说 `cmd/uninstall_init` 早已用默认值判断，删掉 wizard 表单后该变量恒为空，默认行为变成"保留数据"——这是安全方向；`cmd/uninstall_callback` 只做幂等兜底。

改完版本号在五处同步，重新出包，实机安装通过，桌面图标打开能正常进入网关入口。

## 教训

1. **打包工具的校验严于它自己的产出**。`fnpack` 能打出包，不代表应用中心会接受；差异只能用"装包"验证。
2. 报错信息越泛化，越要用**裸模板 + 单变量二分**，而不是在完整包里猜。第 11 步的 diff 比前 10 次猜测加起来有效。
3. 别被字段名误导：根因是两个 wizard 文件，和"格式""版本"这些词毫无关系。
4. 能 SSH 到设备时，`appcenter-cli install-fpk` 的错误输出比应用中心详细得多，比反复试装高效。
