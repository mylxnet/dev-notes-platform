---
title: 日志全绿，表格里一个字都没进去
slug: sendinput-batch-alt-keyup
summary: 自动化走查两次全绿、日志每一行都对，真人用 Alt+左键粘贴却完全无效。根因是两个按键动作不在同一批 SendInput 里，目标程序收到的是 Alt+Ctrl+V。
project: superclip
status: published
published_at: 2026-11-17T03:00:00Z
---

## 现象

SuperClip-C 有一个「填表接力」功能：按住 `Alt` 点一下输入框，就把列表第一行贴进去。设计目标里最典型的场景是往 Excel/WPS 表格里填数据。

真人在 WPS 表格上试，反馈是三种说法：`Alt+左键` 没效果、没反应、能贴出东西但每次都是一样的条目。

而在同一时间段里，自动化走查连着两轮全绿，用户自己的使用日志也全绿：

```
[relay] 贴第 1 条 → XLMAIN|工作簿1
[paste] … 焦点控件=EXCEL7，事件数=4
```

日志证明"取了第 1 行、按键投给了那个焦点控件"，但表格毫无变化。

## 排查

先看注入那一批的事件数：`事件数=4`。一次完整的粘贴需要 4 个事件（`Ctrl↓ V↓ V↑ Ctrl↑`），那 `Alt↑` 去哪了？

接力手势是"**按住** `Alt` 点"，点击那一刻单独发过一次 `Alt↑`，而 `Ctrl+V` 是 60 ms 后由焦点等待定时器发的。于是真人的手指还压着 `Alt` 时，注入发生了——目标程序实收的是 `Alt+Ctrl+V`。在 Excel 与 WPS 里，这个组合是**选择性粘贴**，表格当然一个字都不动。

走查脚本压不出这个问题，因为脚本自己合成 `Alt` 和点击，`Alt` 是按脚本的安排按时抬起的。

## 修复

关键是把"抬 `Alt`"和"注入 `V`"排进**同一个 `SendInput` 批次**，让物理按键插不进中间：

```cpp
// 5 或 6 个事件：Alt↑ 恒发，Ctrl↑ 按需
void PasteService::SendCtrlV(bool releaseCtrl) {
    std::vector<INPUT> batch;
    batch.push_back(KeyUp(VK_MENU));      // 恒发：点那一刻抬过，这里再抬一次是幂等的
    if (releaseCtrl) batch.push_back(KeyUp(VK_CONTROL));
    batch.push_back(KeyDown('V'));
    batch.push_back(KeyUp('V'));
    if (releaseCtrl) batch.push_back(KeyDown(VK_CONTROL));
    SendInput(static_cast<UINT>(batch.size()), batch.data(), sizeof(INPUT));
}
```

同时在注入那一刻复查 `Ctrl` 是否仍被按住——但要选对 API：

```cpp
// WM_TIMER 不带键盘状态快照，GetKeyState 读到的是上一条已处理按键的状态，不可信
bool ctrlDown = (GetAsyncKeyState(VK_CONTROL) & 0x8000) != 0;
```

而 `Alt↑` 恰恰**不能**按键态判断：点击那一刻已经单独抬过一次，此时读 `GetAsyncKeyState(VK_MENU)` 读到的是自己抬起的结果（手指其实还压着），一判断就会漏发，批次退回 4 事件等于没修。

## 反转：改两个变量，退了一步

下一个版本为了验证两个猜想——键盘提示菜单态、以及补充扫描码——同时改了批次顺序、补了 `MapVirtualKeyW` 的扫描码、删掉点击时的单独 `Alt↑`。结果更糟：连上一版唯一能成功的编辑态也失效了，两个猜想双双证伪，而且因为一次改三处，无法归因。

真正的答案是**对照组**给出的。为了区分猜想，做了一次完全不涉及 `Alt` 的对照：在 WPS 网格里单击选中空单元格，用快速模式选中一条，按空格做普通粘贴。用户回答：**没成功过**。

这说明注入链在 WPS 网格（`EXCEL7`，单元格仅选中）上从来就没成功过，和接力、和 `Alt`、和那轮改动都无关。两种实现只剩一个差别：旧版是 `keybd_event` 四次独立调用，新版是一次 `SendInput` 投整批。

## 教训

1. **脚本走查全绿 ≠ 真人手势可用**。涉及修饰键的注入功能，必须真人手验。
2. **"按键投给了那个焦点控件"不等于"内容落进了目标"**。落点的唯一判据是目标程序里的内容真的变了。
3. **一条现象能同时解释两个假设时，别一次改两个变量**——改完退一步却无法归因，比不改更糟。
4. **对照组比修复组更有信息量**。那个不涉及 `Alt` 的空格粘贴对照，一句话就把从旧版本起所有"真表格落点已实证"的结论全推翻了。
5. 日志会替代码"报喜"：先写"贴第 N 条"再调注入，注入被静默丢弃时那行日志就是在说谎。
