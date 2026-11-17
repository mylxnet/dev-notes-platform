---
title: 哪些邮件算收件：同一份判定被写了两遍
slug: visibility-scope-single-source
summary: 通知说收到 6 封，列表只多 4 封；联系人页还留着服务器上早就删掉的邮件。根因一样——判断「什么算收件」的那套条件，在三个地方各写了一份。
project: mailhub
status: published
published_at: 2026-11-18T04:00:00Z
---

## 两个现象，一个根

第一个：同步完成后的 toast 和通知铃铛显示「收到 N 封新邮件」，但聚合收件箱只多出 M 封，N > M，Gmail 一信多标签时差得更多。

第二个隔了一版才冒出来：服务器上已经进了回收站或垃圾箱的邮件，收件箱里没有了，却还挂在联系人页的「往来邮件」和「往来封数」里。

表面是两个 bug，实际是同一件事：**「哪些邮件算收件」这套判定，代码里有三份实现，各自慢慢跑偏。**

## 第一份和第二份

收件列表的可见性由 `inboxScope()` 决定，条件是：

```
账号 enabled
AND archivedLocal = false
AND trashedLocal = false
AND 不关联任何被排除的文件夹（\Sent / \Drafts / \Trash / \Junk / \Spam）
```

同步收尾的计数却是另一套：

```ts
for (const folder of targets) {
  totalNew += added     // 不分文件夹，全加
}
```

`added` 是「新建的 Message 条数」，它把被聚合口径排除掉的文件夹也算进去了——回收站里的信同样 +1。两边口径分裂，N 就永远大于 M。

修法不是调数字，是把判定收敛成一个函数：

```ts
/** 唯一事实来源；必须与 inboxScope() 的 flags:{ contains } 严格等价 */
export function isInboxFolder(flags: string): boolean
```

同步侧改成 `if (isInboxFolder(f.flags)) inboxNew += added`。通知条数与返回给前端的列表同源之后，前端一行没改，toast 自己就一致了。

## 收敛时踩的第一个坑：Prisma 的 `contains` 是 LIKE

`inboxScope()` 里写的是 `folder: { flags: { contains: '\\Sent' } }`。直觉上这是「精确匹配某个标记」，写 JS 复刻版时自然按精确匹配写：

```ts
flags.split(' ').includes('\\Sent')    // 直觉版，错的
```

真库一跑就翻了。Prisma 在 SQLite 下把 `contains` 编译成 `LIKE '%\Sent%'`，语义是**子串匹配 + ASCII 大小写不敏感**：

| `flags` 实际值 | 直觉版（精确） | 真库 `contains` |
|---|---|---|
| `\Sent` | 排除 | 排除 |
| `\sent` | 不排除 | **排除** |
| `\SentItems` | 不排除 | **排除** |
| `\TrashBin` | 不排除 | **排除** |
| `\All` | 不排除 | 不排除 |

反斜杠在 LIKE 里还要转义，等于又叠了一层容易写错的东西。结论：JS 侧不能凭语义直觉复刻，必须拿同一份 ORM 谓词去对照。

对照的做法是临时库 + `prisma db push`，造 13 组带不同 flags 的样本，逐例断言：

```
断言：!isInboxFolder(flags) === (该行能否被 inboxScope() 查出来)
```

13 例全过才算等价。这套断言后来一直留着——凡是「JS 侧要和 SQL 谓词对齐」的地方，读文档没用，只能真库跑。

## 第二个坑：`every` 对空关系恒为真

第三份实现出现在清理逻辑里：删掉「有关联、但所有关联都落在非收件文件夹」的幽灵行。初版写的是：

```ts
folderLinks: { every: { folder: { flags: { not: { contains: EXCLUDED } } } } }
```

`every` 在关系为空时取值是 `true`——空集上任何全称命题都成立。也就是说**一条关联都没有的邮件也会被选中**，那些刚同步一半、关联还没建起来的正常邮件，正在删除范围里。

正确的写法要用 `some` 保证非空、`every` 保证全落：

```ts
folderLinks: {
  some: {},                                              // 至少有一条关联
  every: { folder: { flags: { in: EXCLUDED_FLAGS } } },   // 且全都落在排除清单
}
```

同一段代码初版还把条件写反过一次：`some: {}` 配 `NOT some 非收件`，实际命中的是「全落在收件夹」的那批正常邮件，跑起来会把正常收信删掉。清理类 where 只靠肉眼看逻辑没用，这次是拿真库做谓词对照当场抓出来的。

## 第三份：联系人模块自己抄了一套

联系人页的往来邮件只判了 `trashedLocal`，漏了「非收件文件夹」和「账号是否启用」两层。于是回收站里的邮件在收件箱消失了，却还挂在联系人名下。

这里没有新逻辑，只是把自写的那套换成 `{ ...inboxScope(), fromAddr: ... }`。顺带把 `inboxScope()` 里的排除清单抽成 `nonInboxFolderWhere()`，给清理逻辑共用——一份清单，三个调用方。

## 还剩一个误差没消

Gmail 一信多标签：同一封邮件同时挂在 `INBOX` 和另一个自定义标签下。

- 列表按 **message 级**判定，只要有一个关联落在收件类文件夹就看得见；
- 通知按 **文件夹级**计数，每个命中的文件夹都算一次。

极端情况下列表 1 封、通知说 2 封。彻底对齐要把计数下推到 message 级做去重统计，属于额外一轮改造，这版先记着。

## 一句

同一个概念写两遍，第二遍总会慢半步。凡是「什么算可见」这种横跨列表、计数、清理三处的判定，第一件事是把它变成一个函数，第二件事是拿真库证明这个函数和 SQL 谓词真的等价。
