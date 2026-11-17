---
title: 后台显示同步成功，收件箱一封都没进
slug: imap-inbox-flag-not-standard
summary: 用户说新邮件收不到，第一反应是前端没刷新。查下来后端根本没打开过收件箱：挑文件夹的代码只认 `\Inbox` 这个标记，而它并不在 IMAP 的标准属性里。
project: mailhub
status: published
published_at: 2026-11-18T02:00:00Z
---

## 用户看到的是这样

后台定时收信上线两天后，反馈来了：新邮件收不到。

手边的证据全是「正常」的：

| 证据 | 值 |
|---|---|
| 账号状态 | `ok` |
| `lastSyncAt` | 每轮都在前进 |
| 手动点刷新 | 立刻能收到 |
| 浏览器 F5 | 没有任何变化 |

`lastSyncAt` 在动，说明连上服务器了、也跑完了一轮。可收件箱里就是没有新信。

默认的怀疑方向是前端：列表没重新拉、缓存没失效、SPA 忘了刷新。所以先看接口——`GET /api/messages` 返回的就是空数组。前端拿到的数据本身是空的，这条线断了。

## 后台轮询和手动刷新不是一条路

同步只有一个入口 `SyncService.run()`，但调用方传的 `scope` 不一样：

- 收件页的刷新按钮 → `scope='all'`
- 后台定时轮询 → `scope='primary'`，走另一个分支 `pickPrimaryFolders()`

手动能收到、后台收不到，差异只可能出在这个函数的筛选上。原来的实现：

```ts
const byFlag = folders.filter(f => f.flags.has('\\Inbox') || f.flags.has('\\Sent'))
if (byFlag.length) return byFlag
// 取不到再按名称回落（实际上没有机会执行）
```

`\Inbox` 看着天经地义。问题就出在这个假设上。

## `\Inbox` 不是 IMAP 属性

IMAP 的 `LIST` 返回里，文件夹可以带一串特殊用途标记，由 RFC 6154 定义：`\All`、`\Flagged`、`\Important`、`\Sent`、`\Drafts`、`\Spam`、`\Trash`。**没有 `\Inbox`**。

IANA 的 IMAP Mailbox Name Attributes 注册表里，`Inbox` 那一行的说明是 "JMAP only"——它是 JMAP 的邮箱角色概念，IMAP 的 `LIST` 不会返回。Gmail 的 `LIST` 结果里 `INBOX` 一个特殊标记都不带，而 `[Gmail]/Sent Mail` 老老实实带着 `\Sent`。

于是上面那段代码在 Gmail 上的实际结果：

```
byFlag = [ \Sent 命中：[Gmail]/Sent Mail ]
       → 只剩「已发送」
       → 收件箱从头到尾没被打开过
```

`if (byFlag.length) return byFlag` 这一句把名称回落彻底短路了：只要有任意一个 flag 命中，后面的兜底逻辑永远不会跑。

这也解释了为什么状态是绿的——轮询确实连上了服务器、确实扫了已发送、也确实写了 `ok`。它只是从来没扫收件箱。

## 定案靠一次对比

代码看着像根因，但还得排除「Gmail 是不是另有原因」。做法是同时制造一封新信，分别触发两条路径：

| 触发方式 | `scope` | 结果 | 耗时 |
|---|---|---|---|
| 后台轮询（等节拍） | `primary` | 0 封 | ~9s |
| 手动刷新 | `all` | 2 封 | ~44s |

同样的新信，全量路径立刻入库，后台路径完全看不见。范围是唯一的变量，可以定案。

那个 44s 当时还被误读过一次：后台只花 9s、全量花 44s，看上去像「后台同步更高效」，真相是它只打开了 1 个文件夹。

## 修复：标记和名字取并集

```ts
const byFlag = folders.filter(f => f.flags.has('\\Inbox') || f.flags.has('\\Sent'))
const INBOX_OR_SENT = /^(inbox|sent|sent messages|sent mail|已发送|已发送邮件|已发送的邮件)$/i
const byName = folders.filter(f => INBOX_OR_SENT.test(f.name) || INBOX_OR_SENT.test(lastSegmentOf(f.path)))
const merged = dedupe([...byFlag, ...byName])
if (merged.length) return merged
return folders   // 全取不到，回落全量
```

两个细节：

- 名称匹配同时看 `name` 与 `path` 末段。有些服务商在 `name` 上带命名空间前缀，只看 `name` 会漏。
- 收件箱在 IMAP 里固定叫 `INBOX`、大小写不敏感，这是规范唯一给的名字保证，靠名字比靠属性可靠。

## 三个当时没想到的点

1. **「命中即 return」的筛选最危险**。它把「多策略兜底」写成了「先命中者通吃」，一条命中就扼杀其余判断。这版代码后来整个删掉了，改成按排除清单判定收件类文件夹，反而更稳。
2. **同步成功不代表扫到了该扫的东西**。状态和时间戳只能证明连上过服务器；判断收件箱有没有被扫，唯一硬指标是有没有新邮件入库。
3. **两条路径范围不同会互相掩盖 bug**。手动全量一直正常，本地怎么测都测不出来——除非专门去盯后台轮询那一条。

## 没解决的部分

- 只验证了 Gmail。Exchange、自建 Dovecot 的 `LIST` 属性没逐家对比，靠名称匹配兜住，属于「大概率对」而不是「证明对」。
- `scope` 参数没删。它现在的语义收窄成「单飞合并要不要补跑一次」，不再代表文件夹范围，名字留着有点误导。

## 一句话

`\Inbox` 这种「听起来必须是标准」的东西，恰恰最该去规范里核一遍。IMAP 里收件箱靠名字，不靠标记。
