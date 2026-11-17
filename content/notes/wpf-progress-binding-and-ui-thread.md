---
title: 界面不动的三种原因，一条条对着改
slug: wpf-progress-binding-and-ui-thread
summary: 批量生成时界面无响应、进度条不涨、日志不出——三个症状容易被当成一个问题，实际分别是线程、属性通知、集合通知。这里按症状拆成三条独立检查项。
project: excelmailmerge
status: published
published_at: 2026-11-17T16:00:00Z
---

「界面卡住」是用户的原话，但落到代码里是三个不同的问题，可以分别验证、分别修。

## ① 窗口完全拖不动 → 重活干在 UI 线程

**怎么确认**：任务管理器里进程 CPU 打满、窗口无响应（标题栏出现"未响应"）。

**原因**：`await Task.Yield()` 在 WPF 里是"让出后回到 Dispatcher"，不是切后台。它之后的同步代码照样占着 UI 线程。

```csharp
await Task.Yield();
RunModeA();                       // 成百上千次单元格操作，全在 UI 线程
```

**改**：

```csharp
await Task.Run(() => RunModeA());
```

判断标准：只要一段逻辑的耗时随数据量增长，就不该待在 UI 线程上。

## ② 窗口能点，但进度条停在 0 → 计算属性没被通知

**怎么确认**：日志在涨、进度条不动。

**原因**：进度模型是普通类，`Done` 变化不触发 `PropertyChanged`；而界面绑的是计算属性：

```csharp
public double ProgressPercent => Total == 0 ? 0 : Done * 100.0 / Total;
```

计算属性不会因为 `Done` 改变而自动刷新，必须显式通知它。

**改**：实现 `INotifyPropertyChanged`，并且在 `Done` 的 setter 里**同时通知两个属性名**：

```csharp
_propertyChanged(nameof(Done));
_propertyChanged(nameof(ProgressPercent));   // 计算属性单独通知
```

另外 `ProgressBar.Value` 默认 `TwoWay`，绑只读属性会直接抛异常，XAML 里要写 `Mode=OneWay`。

## ③ 进度条动了，日志区还是空 → 集合类型不对

**怎么确认**：`Add` 之后集合 Count 在涨，界面零行。

**原因**：`Logs` 是 `List<string>`，WPF 只对 `INotifyCollectionChanged` 的集合做增量渲染。

**改**：换成 `ObservableCollection<string>`，其余代码不用动。

## ④ 测试代码里 `.Result` 卡死

同一个病根在测试里也出现过：

```csharp
Task.Run(() => engine.GenerateAsync(...)).Result;   // 死锁
```

`GenerateAsync` 内部的 `await` 捕获了 WPF 的 `SynchronizationContext`，主线程阻塞等结果、续体又要回主线程——互等。

**改**：`Task.Run(() => engine.GenerateAsync(...)).GetAwaiter().GetResult();`，让整个调用跑在没有 Dispatcher 上下文的线程池里。

## 一条判据

先分症状再动手：**完全无响应 = 线程问题；有响应但数字不动 = 属性通知；数字动了文字不出 = 集合通知**。三者修法完全不同，混在一起改容易改坏能用的那部分。
