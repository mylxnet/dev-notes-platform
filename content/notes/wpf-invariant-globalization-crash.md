---
title: 同一份代码，两份打包参数
slug: wpf-invariant-globalization-crash
summary: Debug 里跑得好好的 WPF 程序，打成自包含单文件后启动即崩。两次崩溃都出在同一处：发布参数和 XAML 资源假设不一致，而 Debug 永远碰不到。
project: excelmailmerge
status: published
published_at: 2026-11-17T14:00:00Z
---

同一份代码、同一个提交，只在"怎么打包"上不同，结果一个是正常窗口，一个是双击没反应。

## 两种打包的差异

| | Debug（F5） | 自包含单文件（publish） |
|---|---|---|
| 资源裁剪 | 不裁剪 | 按依赖裁剪 |
| 运行时 | 机器上的共享运行时 | 内嵌，随程序分发 |
| 文化资源 | 全在 | 视 `InvariantGlobalization` 而定 |
| XAML 里的 `Icon` 资源解析 | 从项目文件系统读 | 从单文件包内嵌区读 |

前三行解释了第一次崩溃，第四行解释了第二次。

## 崩溃一：被裁掉的文化资源

启动即退出，事件日志里一句：

```
System.InvalidOperationException: Cannot find non-neutral culture related to 'en-us'
```

堆栈指向 `XmlLanguage.GetSpecificCulture()`。csproj 里有一行为了"精简多语言"加进去的配置：

```xml
<InvariantGlobalization>true</InvariantGlobalization>
```

它把运行时切成无区域性模式。WPF **不支持**这个模式：发布裁剪时 `PresentationCore.resources.dll` 被一起裁掉，`XmlLanguage` 找不到 `en-us` 对应的区域性，直接抛异常。

Debug 之所以正常，是因为它不走裁剪。

结论很简单：删掉这行。多语言资源一共约 1 MB，不值得为它牺牲启动。

## 崩溃二：被单文件打包弄丢的图标

修好上一条、加上应用图标后，同一个打包模式又崩了一次。这次是 XAML 里写了：

```xml
<Window Icon="app.ico" ...>
```

单文件发布时，XAML 的 `Icon` 属性找不到对应的嵌入资源。改法是删掉 XAML 里的 `Icon`，只保留 csproj：

```xml
<ApplicationIcon>app.ico</ApplicationIcon>
```

窗口标题栏图标会自动继承，效果一样，但不再依赖运行时资源查找。

## 这类问题的排查顺序

1. 先确认"只有哪种打包模式会崩"，把差异缩小到一个维度；
2. 打开 GUI 程序的崩溃日志落盘（`%TEMP%\*.txt`），否则连堆栈都没有；
3. 按打包模式集中回归一次——同一模式下往往不止一个坑，修完一个会撞上下一个。
