# 部署手册 · 我的开发笔记

> v1.2.0 · by Mr lin · 2026-11-17

## 一、部署前置条件

| 条件 | 说明 |
|---|---|
| 阿里云 ECS | 中国内地节点，包年包月 ≥3 个月，已购公网带宽 |
| 域名 | 已完成实名认证，持有者与备案主体一致，A 记录指向该 ECS |
| 1Panel | 已在 ECS 安装，可访问管理面板 |
| ICP 备案 | **通过前不得开放 80/443，不得解析域名到 ECS** |

## 二、构建（在 lxsyzd WSL 中）

```bash
cd /mnt/e/app/网站备案/dev-notes-platform
bash scripts/build.sh
```

构建脚本做四件事：校验版本一致性（`app/VERSION` 与 `app/main.go`）、安装前端依赖、构建 Vue 产物到 `app/internal/webui/dist`、以 `-tags embedui` 静态编译为 `app/build/dev-notes`。Docker 镜像在 Section 5 补入。

产物可直接运行（已内嵌前端，无需另装 Node）：

```bash
cd app
DATA_DIR=/opt/dev-notes/data \
SESSION_SECRET=<32 位以上随机串> \
ADMIN_INIT_USERNAME=admin \
ADMIN_INIT_PASSWORD=<12 位以上> \
HTTP_ADDR=127.0.0.1:8080 \
./build/dev-notes
```

访问 `http://127.0.0.1:8080/` 应看到前台首页；`/admin` 为后台入口，用上面注入的初始账号登录（仅首次启动且 `admin` 表为空时写入）。深路径（如 `/notes/<slug>`、`/admin/notes`）依赖 Go 回退 `index.html`，反向代理需放行全路径。生产环境应设置 `COOKIE_SECURE=true` 并启用 HTTPS，否则会话 Cookie 不会下发。详细构建参数、环境变量与版本注入方式见 `doc/PROJECT.md` 第七节。

## 三、数据持久化

容器内 SQLite 路径：`/app/data/notes.db`

Compose 中绑定挂载到宿主机 1Panel 管理的目录，例如：

```yaml
volumes:
  - /opt/1panel/apps/dev-notes/data:/app/data
```

备份时直接复制该目录下的 `.db` 文件即可。

## 四、1Panel 部署步骤

1. 「容器」→「Compose」→「创建」，粘贴 `deploy/docker-compose.yml` 内容
2. 确认数据卷路径，填好管理员初始账号与密码（首次启动写入）
3. 启动容器，确认状态为运行中，日志无报错
4. 「网站」→「创建」→「反向代理」，目标填 `127.0.0.1:8080`
5. 绑定域名，先不启用 HTTPS
6. 本地用 `IP:8080` 验证前台与 `/admin` 后台
7. **备案通过后**：解析域名、启用 80/443、申请 Let's Encrypt 证书

## 五、备案后合规检查

- 页脚显示真实备案号，并链接至 `https://beian.miit.gov.cn`
- 网站开通 30 日内完成公安联网备案
- 前台内容与备案备注一致：个人记录、非经营性、无下载入口
- 草稿与隐藏条目确认未出现在前台

## 六、日常运维

| 操作 | 命令 / 位置 |
|---|---|
| 查看日志 | 1Panel 容器日志，或 `docker logs dev-notes` |
| 备份数据 | 复制数据卷目录下 `notes.db` |
| 升级 | 替换镜像后重新创建容器，数据卷不变 |
| 修改管理员密码 | 登录后进后台「账号」页 |

## 七、回滚

镜像升级异常时，1Panel Compose 中切回上一版本镜像标签，重新创建容器。数据库结构变更需提前备份 `.db` 文件。

## 八、故障排查

| 症状 | 原因 / 处理 |
|---|---|
| 容器启动即退出 | 数据卷目录权限不足，确认容器用户可写 |
| 后台登录失败 | 确认初始账号已写入；忘记密码时删除 admin 表记录后重启重建 |
| 登录后立即跳回登录页 | `COOKIE_SECURE=true` 但当前是 HTTP 访问，浏览器拒收 Cookie；本地调试设 `false` |
| 登录提示失败次数过多 | 同一 IP 15 分钟内失败 5 次；等待或重启容器清零（计数在内存） |
| 保存内容被拒绝并列出命中词 | 内容含经营性词汇或下载类表述，按提示改写为「源码仓库」等合规表述 |
| 前台显示草稿 | 检查条目状态字段，确认接口只查 published |
| 反向代理 502 | 容器未监听 8080，或 Compose 端口未映射 |
