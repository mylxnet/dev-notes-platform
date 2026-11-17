---
title: 把两次启动的时序画出来，问题就自己出现了
slug: single-instance-mutex-after-kill
summary: 连续点两次图标，旧实例被杀掉、新实例判定"已有实例"后也退出，最后什么都不剩。不用猜内存或杀毒软件，把六步时序写下来，错在哪一步一目了然。
project: excelmailmerge
status: published
published_at: 2026-11-17T17:00:00Z
---

## 故障时的时序

```
T0  用户双击图标，实例 B 启动
T1  B 执行 KillRunningInstances() → 实例 A 收到 Kill，进入退出流程
T2  B 立即执行 new Mutex(true, name, out createdNew)
T3  系统尚未释放 A 持有的命名内核对象 → createdNew == false
T4  B 认为"已有实例在跑" → NotifyExistingInstance() → Shutdown()
T5  A 已退出，B 也退出 → 进程数 0
```

关键在 T2 与 T3 之间：`Kill()` 返回只代表"信号已经发出"，A 的进程对象退出、它持有的 Mutex 被系统回收，都在之后发生。这段时间里 B 想拿的那把锁确实还被占着——只是占着它的进程已经不打算再工作了。

于是出现一个自相矛盾的中间态：**锁被占用，但没有活着的实例**。

## 修复后的时序

```
T1  B 依次 Kill 并 WaitForExit(5000)，等 A 真正退出
T2  B 进入 TryAcquireMutex(20 次 × 200ms)
T3  前几次仍拿不到（对象还在释放中）→ 继续重试
T4  系统释放内核对象 → createdNew == true → B 拿到锁
T5  B 正常显示窗口
```

重试窗口 4 秒，远大于内核对象释放的实际耗时；真有另一个实例在正常运行时，20 次都会失败，仍然走"唤起已有窗口并退出"的老逻辑，不会变成多开。

代码只有两处改动：

```csharp
// 杀进程后等它退出
foreach (var p in processes) { p.Kill(entireProcessTree: true); p.WaitForExit(5000); }

// 抢锁改成轮询
for (int i = 0; i < 20; i++)
{
    var m = new Mutex(true, MutexName, out bool createdNew);
    if (createdNew) return m;
    m.Dispose();
    Thread.Sleep(200);
}
return null;
```

## 为什么不用"先杀后等"换个思路

也考虑过"不杀旧实例，只唤起它"。这个方案更干净，但当时程序已经支持 `Kill` 残留进程（应对异常退出留下的僵尸窗口），两者混用才出现这个时序漏洞。修复保留了两条路径，代价是必须接受"杀完要等"这个前提。

## 验收方式

单实例类问题不能只看日志里 `createdNew` 的值，要看结果：

- 连点两次 → 最终恰好一个窗口；
- 正常单开 → 立即启动，不因重试而变慢；
- 手动保留一个实例再启动 → 只唤起旧窗口。

第一条才是用户真正在意的。
