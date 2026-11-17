---
title: 密码输错，提示「请先登录」
slug: 401-on-public-auth-paths
summary: 登录接口返回 401 表示「账号或密码错误」，客户端却统一当成会话失效，把服务端文案换掉还顺手跳转。顺带修了两个尺寸问题：提示条把按钮顶下去 96px、页面多出 48px 可滚。
project: jitijingjiguanli
status: published
published_at: 2026-11-18T00:00:00Z
---

## 三个问题，一次修完

| # | 现象 | 根因 |
|---|---|---|
| 1 | 密码输错，页面提示「请先登录」，看不到真实原因 | 客户端见到 401 一律走会话失效分支 |
| 2 | 错误提示出现/消失时，提交按钮上下跳 96px | 错误条是独立块，插在字段与按钮之间，把卡片撑高 |
| 3 | 认证页整页可滚动，点提交时页面跳动 | 容器重复声明 `min-height:100vh` + `padding:24px`，外层已做过一次 |

## 问题 1：401 有两种含义

`401` 在不同接口上含义不同：

- 业务接口的 401 = **会话失效**，该跳登录页；
- `/auth/login`、`/auth/register`、`/auth/reset-password` 的 401 = **这次操作失败**，属于正常业务结果。

而 HTTP 封装里没有区分：

```ts
if (res.status === 401) {
  onUnauthorized()               // 触发跳转
  throw new Error('请先登录')     // 覆盖了服务端文案
}
```

修法是加白名单，只有白名单之外的 401 才当会话失效：

```ts
const PUBLIC_AUTH_PATHS = ['/auth/login', '/auth/register', '/auth/reset-password']

if (res.status === 401 && !PUBLIC_AUTH_PATHS.includes(path)) {
  if (onUnauthorized) onUnauthorized()
  throw new Error('请先登录')
}
// 白名单路径落到通用错误分支，透传服务端 { error: { code, message } }
```

两个细节：白名单写的是**去掉 `/api` 前缀后的 path**，与内部拼接口径一致；`/auth/registration-status` 不必列入，它用的是裸 `fetch`，不走这层封装。

## 问题 2：让提示与输入框同行

Vant 默认把校验错误追加在字段下方（`.van-field__error-message`），这是撑高的元凶。做法是**把提示塞进字段行的右侧插槽**：

| 手段 | 作用 |
|---|---|
| `:show-error-message="false"` | 关掉默认的下方错误消息 |
| `@failed="onValidateFailed"` | 校验失败时取第一个出错字段名，决定提示挂哪一行 |
| 每个字段的 `#extra` 插槽 | 渲染在 `.van-cell` 最右侧，与输入框**同一行** |
| `errorField` + `showError(field, msg)` | 一个 ref 决定落点，另一个存文案 |
| `.field-error { line-height: inherit; white-space: nowrap; max-width: 140px }` | 行高继承 24px，出现/消失不改变行盒高度；超长省略 |

还有一个必须写 `flush: 'sync'` 的坑：登录失败时会先清空密码再写提示，用默认的 `pre` 队列，清空回调会晚于赋值执行，把刚写好的提示抹掉。

```ts
watch([username, password], () => { error.value = '' }, { flush: 'sync' })
```

提示落点约定：必填/格式错误 → 第一个出错字段行；账号密码错（401）→ 密码行；`USERNAME_TAKEN` → 账号行；`REGISTRATION_CLOSED` 等整单失败 → 组织名称行。

## 问题 3：删掉重复的容器声明

`.login-page` / `.register-page` 只保留 `width:100%` 与 flex 居中，删掉 `min-height:100vh`、`padding:24px`、`background`——居中与留白统一交给外层 `.auth-wrap`。重复一次就是 100vh + 48px。

## 实测

| 指标 | 改前 | 改后 |
|---|---|---|
| 空表提交前后按钮 top 差 | 96px | **0** |
| `scrollHeight - clientHeight` | 96px | **0px** |
| `.van-field__error-message` 节点数 | 1 | **0** |
| 提示位置 | 独立块 | 与输入框同行，距右 16px |
| 真实 HTTP | — | 错误口令返回 `401 INVALID_CREDENTIALS`，页面显示「账号或密码错误」 |

`GET /api/me` 未登录仍走「请先登录」并跳转，说明会话失效分支没有被削弱。

## 一条经验

**用状态码表达语义的接口，客户端必须知道"这个 401 是业务失败还是会话过期"。** 统一处理看着省事，代价是把服务端辛苦写的错误文案全丢掉，用户永远只看到一句话。
