---
title: 手机端点「往来」没反应，服务端返回 200
slug: stale-chunk-html-as-js
summary: 页面懒加载的 JS 请求拿到了 200，内容却是 index.html。严格 MIME 校验把 import() 拒了，路由静默中止——一个"200 但全是错的"响应，比 404 难查得多。
project: jitijingjiguanli
status: published
published_at: 2026-11-17T22:00:00Z
---

## 现象

手机端升级后点「往来」没反应，其他页面偶尔也"永远打不开"。桌面端正常，本机正常。

先按路由问题查：清缓存、换浏览器、本机 390×844 手机壳模拟——**都能正常出数**。接口也正常。

## 关键取证

直接请求一个不存在的 chunk：

```
GET /assets/ContactsPage-abc123.js

HTTP/1.1 200 OK
Content-Type: text/html
Content-Length: 742
```

`200 + text/html`。也就是说，请求 JS 拿到了 `index.html`。

根因链路是这样的：

1. 内嵌前端对未命中的非 `/api/` 路径一律**回退 `index.html`**（为了 hash 路由的深链接）；
2. 客户端还留着升级前的旧 `index.html`，它引用的旧哈希 chunk 在新镜像里已经不存在；
3. 旧页面去请求旧 chunk → 命中回退 → 拿到 `200 text/html`；
4. 浏览器对 `import()` 做严格 MIME 校验，`text/html` 不是 JS → 拒绝执行；
5. 懒加载抛错，Vue Router 静默中止跳转 → 用户看到"点了没反应"。

**这是一个 200 的错误响应。** 没有 404、没有红色错误，控制台只有一条容易被忽略的 MIME 提示。

## 改法：四类响应分开对待

| 请求 | 改前 | 改后 |
|---|---|---|
| `/assets/<缺失>.js` | `200 + text/html`（index.html） | **`404 asset not found`**（`text/plain`） |
| `/`、`/index.html` | 200，无缓存头 | 200 + `Cache-Control: no-cache` |
| `/assets/<哈希>.js\|.css\|.mjs` | 200，无缓存头 | 200 + `public, max-age=31536000, immutable` |
| 其余非 `/api/` 路径 | 回退 index.html | 不变 |

三条规则各有理由：

- **assets 目录下的未命中必须 404**。这个目录是构建产物，路径带内容哈希，命中不了就是真的不存在，回退 HTML 只会制造假 200；
- **`index.html` 必须 `no-cache`**。它是入口，缓存住就等于把用户钉在旧版本上；
- **哈希资源可以 `immutable` 一年**。内容变了文件名就变，缓存不会错。

## 回归测试要做负向验证

```go
// server/cmd/server/main_test.go
func TestStaticAsset404AndCacheHeaders(t *testing.T) {
    // 用 httptest 起真服务，断言四类响应的状态码、Content-Type 与缓存头
}
```

写完还要做一次**反向取证**：把 404 分支临时关掉再跑，测试必须变红（实测 `FAIL main_test.go:57`）。否则分不清"断言真的在守"和"断言恒真"。

## 一个必须承认的限制

这次只堵住了「HTML 被当 JS 加载」这一条。已经打开、且用户不刷新的旧页面，仍然会看到旧界面——区别是现在会在网络面板里显式 404，不再静默失败。

真正彻底的做法是「版本漂移自愈」（检测到 chunk 404 就强制刷新），本次没做。

## 运维注意

- 换镜像后**首次访问**才会拿到新 `index.html`，已打开的标签页需要刷新一次；
- 如果将来把前端改成非哈希产物名，`immutable` 会把错误版本钉死一年——建议保持 Vite 默认的 `assets/[name]-[hash]`。

## 一句话

**引入 SPA 回退时，一定要给"构建产物目录"留一个例外。** 回退是为了深链接，不是为了把不存在的脚本伪装成成功响应。
