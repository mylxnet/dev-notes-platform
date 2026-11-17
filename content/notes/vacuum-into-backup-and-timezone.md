---
title: 备份文件字节一致，却在凌晨三点没跑起来
slug: vacuum-into-backup-and-timezone
summary: 用一条 VACUUM INTO 拿到一致性快照，合同文件也在同一个 .db 里，下载下来 SHA1 完全一致。真正出问题的不是备份本身，而是「每天 03:00」这个时间。
project: jitijingjiguanli
status: published
published_at: 2026-11-17T21:00:00Z
---

## 一条 SQL 解决一致性

SQLite 运行期有 WAL，直接 `cp` 数据库文件拿到的是不一致的快照。备份用的是：

```sql
VACUUM INTO '<备份文件路径>';
```

它生成**静态一致快照**，天然规避 WAL 不一致；产物是一个独立的 `.db` 文件，与数据同库同版本，可以直接覆盖恢复。

另一个省事的地方：合同原件以 BLOB 存在 `contract.file_data` 字段里，不落文件系统。所以**一致性备份 = 业务数据 + 全部合同文件**，一份 `.db` 就是全量，不用单独备合同。

## 生命周期

| 动作 | 接口 | 说明 |
|---|---|---|
| 列表 | `GET /api/backups` | 时间倒序 |
| 创建 | `POST /api/backups` | 生成快照并触发保留清理 |
| 下载 | `GET /api/backups/:id/download` | 流式下发 |
| 恢复 | `POST /api/backups/:id/restore` | 覆盖当前库，恢复后强制重新登录 |
| 删除 | `DELETE /api/backups/:id` | 最新一份受保护不可删 |

下载实现有两个细节值得记：

- `id` 先过 `filepath.Base` 消毒，防路径穿越；再 `os.Stat` 校验存在且非目录，否则 404；
- 用 `c.FileAttachment(path, id)` 流式下发，自动带 `Content-Disposition: attachment`，大文件不占内存。

前端用 `credentials:'include'` 拉 blob，从 `Content-Disposition` 里解析文件名，再 `URL.createObjectURL` + 隐性 `<a download>` 触发保存。

## 真正的问题：三点没跑

自动备份配置成「每日 03:00」，代码里是这样算下次触发时间的：

```go
next := now.Location()      // ← 用进程本地时区
```

而出问题的环境是容器，**默认时区是 UTC**。于是「凌晨 3 点」被算成了北京时间 **11:00**——备份确实每天都在跑，只是晚到了 8 小时，且日志时间戳同样是 UTC，比北京时间少 8 小时，排查时第一眼完全对不上。

修复是把时区固定下来，不依赖进程环境：

```go
// 复用平台层的固定时区，不读 now.Location()
loc := platform.TZ()   // Asia/Shanghai
```

同时在运行侧统一 `TZ=Asia/Shanghai`：三个 Dockerfile、三个 compose、部署脚本全部设置，日志时间也回到东八区。

## 验证

| 检查项 | 结果 |
|---|---|
| 下载接口 | `200` |
| 响应头 | `Content-Disposition: attachment; filename="jz-backup-….db"`、`application/octet-stream` |
| 二进制一致性 | 下载文件 SHA1 == 服务器备份文件 SHA1 |
| 前端下载按钮 | 设置页备份行出现，点击无失败提示 |

## 保留策略与安全边界

- 每次创建后 `prune()`：超出 `APP_KEEP_BACKUP`（默认 30 份）删除最旧；
- 最新一份不可删（`ErrLatestProtected`），但**下载不受此限制**；
- 所有备份接口都要求登录；
- 恢复成功后清空会话 Cookie——旧会话属于恢复前的库，留着就是错位的身份。

## 两条留给运维的话

1. 备份目录要与数据目录**分开挂载**（`/data` 与 `/backups` 两个卷），否则单点故障会同时毁掉数据和备份。
2. 恢复是破坏性操作，当前只靠前端确认弹窗兜底，没有二次输入校验——用之前先下载一份当前库。

## 一句话

`VACUUM INTO` 让"一致性"这件事变得便宜，而**时区**这种"看起来与功能无关"的配置，才是让功能静默跑偏的常客。
