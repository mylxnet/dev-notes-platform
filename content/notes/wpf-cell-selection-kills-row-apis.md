---
title: 给表格换个选择模式，三个功能一起坏了
slug: wpf-cell-selection-kills-row-apis
summary: 预览表从整行选择改成单元格选择之后，「删这一行」删不掉、「点同一格取源」没反应、右键没选中过的行直接崩。三个症状、一个前提：单元格选择模式下行级 API 全部作废。
project: exceltohub
status: published
published_at: 2026-11-18T08:00:00Z
---

## 一次改版留下的三颗雷

台账预览表的初衷是「整行选中」：点哪一行选哪一行，右键删掉这一行。

后来改版，要让人能精确点某一个格子取地址，选择模式改成了：

```xml
<DataGrid SelectionUnit="Cell" ... />
```

改完界面确实好看了，点格子取源也顺手了。代价是三处功能在接下来几天里陆续炸掉，而且表现完全不同，看不出是同一个原因。

## 症状一：右键「删除此行」，一行都删不掉

菜单点了有反应，日志却每次都打同一句：

```
[提示] 请先在预览表的某一行上点右键再选「删除此行」
```

明明就是在那一行上点的右键。

菜单项的代码是这么取行的：

```csharp
var row = dataGrid.SelectedItem as DataRowView;
if (row == null) {
    log("[提示] 请先在预览表的某一行上点右键再选「删除此行」");
    return;
}
```

`SelectionUnit="Cell"` 之后，**WPF 不再维护 `SelectedItem`**。选中一个单元格时它是 `null`，选中一整行也是 `null`——因为「行选中」这个概念在这套模式里不存在。

所以那句提示文案不是「你没点对」，而是「我取不到行」。修法是把右键命中的行自己记下来：

```csharp
private DataRowView? _rightClickRow;

void OnPreviewRightDown(object sender, MouseButtonEventArgs e) {
    var row = FindAncestor<DataGridRow>((DependencyObject)e.OriginalSource);
    _rightClickRow = row?.Item as DataRowView;   // 自己记
}

// 菜单项优先用记下来的行；键盘/程序触发菜单时退回单元格集合
var target = _rightClickRow ?? (grid.SelectedCells.Count > 0
    ? grid.SelectedCells[0].Item as DataRowView
    : null);
```

## 症状二：右键「还没选中过」的行，程序直接没了

这个更凶。表现是弹一个未处理异常框，程序退出，**正在配置的规则全丢**。

诡异之处在于：行本来就选中过的时候，右键不崩；刚进第②步直接右键，几乎必崩。

异常内容指向的很直接：

```
InvalidOperationException: 父级 DataGrid 上 SelectionUnit 属性的当前值不允许选择行
```

右键处理里有一句「顺手把这一行标记为选中」：

```csharp
row.IsSelected = true;   // 在 Cell 模式下这句会抛
```

为什么「有时候不崩」？因为 `IsSelected` 已经是 `true` 时，赋值不产生属性变更，WPF 直接跳过 setter 的校验——**没变化就不报错**。一旦这行的 `IsSelected` 是 `false`，赋值就是一次真实的变更请求，于是抛。

「时好时坏」的 bug 大多是这个结构：**错误路径只在状态发生变化时才走到**。

修法是彻底不再碰行级状态，改成给被点中的单元格上视觉标记；要删哪一行，由 `_rightClickRow` 说了算。

## 症状三：删掉一行之后，点同一个位置的格子不取源

第三次是功能层面的：右键隐藏掉第 2 行，然后去点现在显示在第 2 行位置的格子（原本的第 3 行），规则表毫无反应。

取源逻辑挂在 `SelectedCellsChanged` 上，而 WPF 只在「选中的单元格**变了**」时才抛这个事件。刚才右键时已经把这个格子选过了，再点一次——选择没变，事件不抛，整个取源流程被跳过。

```csharp
// 左键按下：清掉「本次已取源」标记
void OnPreviewLeftDown(...) { _sourcePicked = false; }

// 左键抬起：本格还没被处理过，就补一次
void OnPreviewLeftUp(...) {
    if (!_sourcePicked) PickSource(currentCell);   // 一次点击仍只取一次
}
```

这条修完之后有个必须保住的语义：**一次点击还是只取一次源**。补跑和重复触发是两件事，别为了修「没反应」变成「点一下取两次」。

## 三个症状，一个前提

把三件事并排看，原因就一条：

| 症状 | 直接原因 | 共同前提 |
|---|---|---|
| 删行删不掉 | `SelectedItem` 恒为 `null` | `SelectionUnit="Cell"` |
| 右键崩溃 | 给 `IsSelected` 赋值抛异常 | 同上 |
| 点同一格不取源 | `SelectedCellsChanged` 不重复抛 | 同上 |

`SelectionUnit` 不只是「视觉上选中什么」，它决定了**哪些 API 还有意义**。改成 `Cell` 之后，行级的 `SelectedItem`、`SelectedItems`、`IsSelected` 全部作废；单元格级的 `SelectedCells`、`CurrentCell` 才是可用面。

## 修好了要能证明

这三个问题都在鼠标路径上，靠手点很容易「这次没崩就算好了」。所以补了一个 UI 自动化回归（`test_hide.ps1`，13 项），逐条断言：

| # | 断言 | 结果 |
|---|---|---|
| 1 | 右键预览第 2 行弹出上下文菜单（弹窗 1 个） | PASS |
| 5 | 点「删除此行」后行数 9 → 8，只掉一行 | PASS |
| 6 | 原第 3 行顶上来成为显示第 2 行 | PASS |
| 9 | 隐藏后点该行，取到 `B3` 而不是 `B2` | PASS |
| 11 | 「恢复所有隐藏行」后行数 8 → 9 | PASS |
| 12 | 内容与初始一致 | PASS |

第 9 项是这轮里最关键的一条：它同时验证了「隐藏只影响预览、不影响真实行号」和「同一位置再点能取源」。

写这个脚本本身又踩了四个坑，值得单独记：

1. **WPF 的右键菜单不是 `#32768`**。那是 Win32 菜单的类名，WPF 的 `ContextMenu` 是进程内 `HwndWrapper[...]` 弹窗。筛窗口要按「本进程 + 可见 + 非主窗口」，不能按类名。
2. **同名菜单项会出现两次**。菜单弹窗和主窗口都暴露 `MenuItem`，按名字找要先去重。
3. **`SetForegroundWindow` 会失败**。前台锁不让抢，得先 `AttachThreadInput(当前线程, 前台线程)`，而且**每次重试都要重新抢**——进第②步后的第一次点击常常落到别的窗口上。
4. **最小化时所有 `BoundingRectangle` 是 -32000 级别的负值**。先 `ShowWindow(SW_RESTORE)`，否则点哪儿都是点在屏幕外。

还有一条：鼠标要先挪开再挪到目标格。原地不动直接右键，WPF 收不到移动/进入事件，菜单根本不弹。

## 一条留给下次的规矩

改 `SelectionUnit`、`SelectionMode`、`CanUserSortColumns` 这类「模式型」属性时，把它当成一次 API 变更来处理：先把该控件上所有依赖旧模式的代码列出来（行级操作、事件订阅、右键路径），一条条过。

它们的共同特征是——**编译期不会报错，运行期也不一定报错**。上面三个问题，一个静默失效、一个偶发崩溃、一个依赖操作顺序，没有一个会在改完那一刻暴露。
