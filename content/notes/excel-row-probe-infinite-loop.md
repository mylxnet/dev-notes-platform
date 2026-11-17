---
title: 那一百秒里，程序在数第 1048576 行
slug: excel-row-probe-infinite-loop
summary: 生成第一条数据就卡一百多秒。把时间线摊开看，程序在空转一个永远不结束的探查循环：循环体改了循环条件依赖的变量，一路数到 Excel 的行数上限。
project: excelmailmerge
status: published
published_at: 2026-11-17T15:00:00Z
---

用户点了「开始生成」，然后等了一百多秒。这段时间里程序没死、没报错，只是在原地数行。

## 这一百秒里发生的事

单文件模式要复用模板格式，开头有一段"向下探查是否有带格式的空行"的代码：

```csharp
int lastRow = src.LastRowUsed().RowNumber();

for (int r = lastRow + 1; r <= lastRow + 30; r++)
{
    if (src.Row(r).Height > 0)
        lastRow = r;        // 循环体修改了循环条件里的变量
}
```

把它跑一遍：

| 轮次 | `r` | `lastRow` | 条件 `r <= lastRow + 30` |
|---|---|---|---|
| 1 | 30 | 30 | 成立 |
| 2 | 31 | 31 | 成立 |
| 3 | 32 | 32 | 成立 |
| … | … | … | 永成立，直到 `r` 撞上 1048576 |

`r` 每前进一格，`lastRow` 就跟一格，上界永远比 `r` 大 30。循环不会停，只会撞到 Excel 的硬上限。

而且每一行都会"命中"——因为判断条件是 `Height > 0`，而没有自定义行高的行返回的是**默认行高 15**，`15 > 0` 恒为真。等于每轮都在推高边界。

一百多秒，就是这么来的：约一百万次无意义的行对象访问。

## 改法

边界只算一次，行高判断改成"与默认值有差异"：

```csharp
int lastRow = src.LastRowUsed().RowNumber();
int maxRowProbe = lastRow + 30;     // 一次算好，不再变

for (int r = lastRow + 1; r <= maxRowProbe; r++)
{
    if (Math.Abs(src.Row(r).Height - defaultRowHeight) > 0.01)
        lastRow = r;
}
```

## 顺手记一笔

同一个方法里还有列方向的同类问题：`LastColumn()` 不包含"有边框但没值"的单元格，所以要显式往后探查若干列，并把边界各 +1 扩展。写法与上面一致——先固定边界，再在循环里累加结果。
