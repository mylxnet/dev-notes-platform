---
title: iframe 塞 PDF，桌面能看手机白屏
slug: mobile-pdf-preview-pdfjs
summary: 合同预览把 base64 塞进 iframe，桌面 Chromium 正常、安卓和微信白屏。换成 pdf.js 逐页画 canvas 后又撞上一个只在隐藏窗口里发生的「无声卡死」。
project: jitijingjiguanli
status: published
published_at: 2026-11-17T23:00:00Z
---

## 三端能力不一样

原来的实现：

```html
<iframe :src="`data:application/pdf;base64,${fileData}`" />
```

本地桌面 Chromium 里看是好的，所以一直以为没问题。实际各端行为：

| 环境 | iframe 内渲染 PDF | 说明 |
|---|---|---|
| 桌面 Chromium | ✅ | 内置阅读器接管 |
| 安卓 Chrome | ❌ | 没有内置 PDF 查看器 |
| iOS Safari | ❌ | QuickLook 只对**顶层导航**生效，iframe 内不触发 |
| 微信内置浏览器 | ❌ | 直接白屏 |

结论：**`data:` URI + iframe 这条路在移动端根本不成立**，与代码写得对不对无关。

## 换成 pdf.js 自渲染

```ts
const pdfjs = await import('pdfjs-dist')
pdfjs.GlobalWorkerOptions.workerSrc = workerUrl
```

几个实现要点：

- **worker 必须同源**。用 `pdfjs-dist/build/pdf.worker.min.mjs?url` 让 Vite 打包进产物，不走 CDN——内网部署根本连不上外网；
- 按容器宽度 × `min(devicePixelRatio, 2)` 逐页画 `<canvas>`，弹窗内滚动；
- 失败**不关弹窗**：顶部提示「无法在线渲染，请下载后查看」，下载按钮保留；
- blob URL 生命周期要管住：`src` 与下载按钮的 `href` 优先用 blob，关闭按钮 / Esc / 解析失败 / 组件卸载都要 `revokeObjectURL` 并清空 canvas。

## 一个反直觉的坑：别复用 loading 分支

PDF 分支**不能**走原来的 `previewLoading` 状态——那条分支会整块替换掉 `.preview-body`，canvas 容器直接不存在，渲染完也没地方画。

渲染中的状态改由一行文案承担（「正在渲染…」/「共 N 页」/ 失败提示），容器保持存在。

## 第二个坑：隐藏窗口里 rAF 不触发

自测时用 390×844 手机壳 iframe 打开 PDF，**不报错、控制台无输出、canvas 一直是空的**。

查下来：该窗口 `document.hidden === true` 时，`requestAnimationFrame` 完全不触发（`setTimeout` 正常），而 pdf.js 的 `render()` 依赖 rAF 推进。于是渲染永远卡在第一步，且**没有任何错误**——典型的"无声卡死"。

给 iframe 注入 rAF → `setTimeout` 垫片后跑通。但要清楚：这只证明**代码逻辑正确**，真机前台行为仍需实机复核。

## 代价与验证

体积：

| 产物 | 大小 | 加载时机 |
|---|---|---|
| `pdf-*.js` | 365 KB | 懒加载，只有打开 PDF 才拉 |
| `pdf.worker.min-*.mjs` | 1.37 MB | 同上 |
| `ContractsPage-*.js` | 861 KB | 进入合同页 |

验证（390×844）：

- 8 页 PDF → 弹窗全屏、显示「共 8 页」、8 张 canvas 均有实际像素（非白像素 2.7 万–6.9 万/页）、容器可滚动（`scrollHeight 4172 / clientHeight 697`）；
- 下载兜底 → 669,153 B，与库内 `file_size` 一致，`%PDF-1.7`，文件名正确；
- 关闭后取旧 blob → `TypeError`（已 revoke），canvas 归零。

## 还没验证的

- **真机未测**：只有桌面 Chromium，安卓/微信前台表现待实机确认；
- 非 PDF 类型（图片、Word、Excel）走 blob 后的表现未实测；
- 没有断点续传，大文件靠流式 + 懒渲染。

## 一句话

**「本地能看」不是判据。** 预览类功能要按目标端逐一列能力矩阵，桌面浏览器恰好是最宽容的那一端。
