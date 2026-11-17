---
title: 返回值是 TRUE，窗口却没置顶
slug: win32-topmost-silently-dropped
summary: SetWindowPos 返回 TRUE、GetLastError 为 0，窗口的置顶位却没落上。排查下来是窗口不在前台时系统会静默丢弃 z-order 变更——返回值只能说明调用被接受，不能说明状态已生效。
project: superclip
status: published
published_at: 2026-11-17T02:00:00Z
---

## 现象

SuperClip-C 有个「启动即置顶」的需求。配置里写的是 `Topmost=true`，冷启动后去读窗口扩展样式，`GWL_EXSTYLE` 里的 `WS_EX_TOPMOST` 位却是 0。

诡异的地方在于调用本身完全"成功"：

```cpp
// 返回值 TRUE，GetLastError() = 0，但事后读 exstyle 置顶位仍是 0
SetWindowPos(hwnd_, HWND_TOPMOST, 0, 0, 0, 0,
             SWP_NOMOVE | SWP_NOSIZE | SWP_NOACTIVATE);
```

连续 6 次冷启动，6 次都是这个结果。

## 排查

按"哪一步没生效"逐个排除，很快发现换成别的入口也一样：

| 尝试 | 结果 |
|---|---|
| `BringWindowToTop` | 返回 TRUE，无效果 |
| `SwitchToThisWindow` | 无效 |
| 由外部进程跨进程调用 `SetWindowPos(HWND_TOPMOST)` | 同样返回 TRUE，位置不变 |
| 先用鼠标真实点一下该窗口，再执行同一条调用 | **立刻生效** |

重试、加大延时都救不了——不是时序问题。真正有区分度的信号是最后一行：**窗口成为前台之后，同一条调用才生效**。

由此定位到后台带（z-order band）的前台限制：目标窗口不在前台时，系统会**静默丢弃**这条请求，既不报错也不返回失败。

## 修复

修复分两条路，一条正道、一条兜底：

```cpp
// 1) 显示链：ShowWindow + UpdateWindow 之后再断言一次
ShowWindow(hwnd_, SW_SHOW);
UpdateWindow(hwnd_);
SetWindowPos(hwnd_, HWND_TOPMOST, 0, 0, 0, 0,
             SWP_NOMOVE | SWP_NOSIZE | SWP_NOACTIVATE);

// 2) 激活兜底：不在 WM_ACTIVATE 里直接改窗口状态，延后一条消息再补发
case WM_ACTIVATE:
    if (LOWORD(wParam) == WA_ACTIVE) {
        PostMessageW(hwnd_, WM_APP_RAISE_TOPMOST, 0, 0);
    }
    break;
```

第 2 条里的"延后"是必须的。系统在处理 `WM_ACTIVATE` 的过程中还会再动一次窗口状态，此时直接改会被紧接着的动作覆盖掉。

## 验证

- 修复后连续两次冷启动（包含把前台让给靶窗的那次），读回的 `topmost_bit` 都是 `True`。
- 兜底分支后来也拿到了实机证据：重启由后台拉起的实例正好处于丢带态（`exstyle=0x40000`，置顶位 0），外部进程的 `BringWindowToTop` 与跨进程 `SetWindowPos(HWND_TOPMOST)` 都返回 TRUE 而位不落；注入一次 `Alt` + `SetForegroundWindow` 让本窗成为前台后，位仍为 0，**800 ms 后变成 `0x40008` 并保持**——说明延后补发那条链路真的被走过了。

顺带确认了一件事：**外部脚本修不好它**。所有跨进程的置顶尝试都无效，只能靠应用自身的激活兜底。

## 教训

1. **Win32 的 TRUE 只表示"调用被接受"，不表示"状态已变更"**。凡是系统可能不答应的调用——z-order、前台、剪贴板——都要用**事后断言**，而不是信任返回码。
2. 自动化脚本的判据要用 `GWL_EXSTYLE` 的位，不要用 API 返回值。判据错了，测试会一直"通过"。
3. 遇到"返回成功但没效果"，先找**能让它生效的那个差异条件**（这里是窗口是否在前台），比反复重试有用得多。
