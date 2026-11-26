# 部署手册 · 我的开发笔记

> v1.4.10 · by Mr lin · 2026-11-27

## 一、部署前置条件

| 条件 | 说明 |
|---|---|
| 阿里云 ECS | 中国内地节点，包年包月 ≥3 个月，已购公网带宽 |
| 域名 | 已完成实名认证，持有者与备案主体一致，A 记录指向该 ECS |
| 1Panel | 已在 ECS 安装，可访问管理面板 |
| ICP 备案 | **通过前不得开放 80/443，不得解析域名到 ECS** |
| 镜像仓库（Docker 路线） | `registry.cn-hangzhou.aliyuncs.com/mylxnet/dev-notes`；先在 WSL 执行 `docker login --username=<阿里云账号> registry.cn-hangzhou.aliyuncs.com`，脚本不保存密码 |

两条路线都可用，数据目录布局一致（宿主机 `data/` 对应容器内 `/app/data`），可互相迁移：

| 路线 | 适用场景 | 升级方式 |
|---|---|---|
| 单文件二进制 + systemd | 先把站跑起来、最小化依赖 | 打新包 → scp → 改软链 → restart |
| Docker 镜像 + 1Panel Compose | 长期维护、频繁升级 | 改 compose 里的镜像 tag → 重新创建容器 |

## 二、构建（在 lxsyzd WSL 中）

```bash
cd /mnt/e/app/网站备案/dev-notes-platform
bash scripts/build.sh          # 正式包：app/build/dev-notes
```

`build.sh` 校验五处版本一致性、用国内源安装前端依赖、构建 Vue 产物到 `app/internal/webui/dist`，再以 `-tags embedui` 静态编译单文件二进制。产物已内嵌前端，无需另装 Node。

**禁止用 `scripts/build-test.sh` 的产物部署**：测试包登录页预填管理员密码。

## 三、数据持久化与备份

容器内 SQLite 路径：`/app/data/notes.db`；后台上传的封面/正文图片在 `/app/data/media/`。备份必须同时带上两者，图片不在数据库里。

```bash
# 二进制路线：宿主机目录 /opt/dev-notes/data
sudo systemctl stop dev-notes
sudo tar czf /opt/backup/dev-notes-$(date +%F).tar.gz -C /opt/dev-notes data
sudo systemctl start dev-notes

# Compose 路线：compose 文件同级的 ./data
tar czf backup-$(date +%F).tar.gz -C <compose目录> data
```

应用内备份（v1.4.10 起接口可用，后台界面待接入）：

- `GET /api/admin/backup/summary` 看规模与已有快照；`POST /api/admin/backup` 下载 `.tar.gz`（数据库 + `media/` + `manifest.json` + 逐文件 sha256）。
- 恢复前先 `POST /api/admin/backup/inspect` 预检（不改数据），再 `POST /api/admin/backup/restore` 并带 `confirm=RESTORE`；服务器会先把当前数据打成 `data/backups/pre-restore-*.tar.gz`，失败自动回滚。
- 恢复期间其余 API 返回 503（`Retry-After: 3`），请避开业务高峰；单包上限 500 MB。

## 四、路线 A：二进制 + systemd

1. WSL 打正式包后上传：
   ```bash
   scp app/build/dev-notes <用户名>@<ECS_IP>:/tmp/dev-notes
   ```
2. ECS 上建目录与用户（用版本号文件 + 软链便于回滚）：
   ```bash
   sudo useradd -r -s /usr/sbin/nologin devnotes
   sudo mkdir -p /opt/dev-notes/bin /opt/dev-notes/data
   sudo install -m 0755 /tmp/dev-notes /opt/dev-notes/bin/dev-notes-1.4.10
   sudo ln -sfn /opt/dev-notes/bin/dev-notes-1.4.10 /opt/dev-notes/bin/dev-notes
   sudo chown -R devnotes:devnotes /opt/dev-notes
   ```
3. 写 `/etc/dev-notes.env`（权限 600），字段见第六节；`SESSION_SECRET` 用 `openssl rand -hex 32` 生成。
4. 建 `/etc/systemd/system/dev-notes.service`：
   ```ini
   [Unit]
   Description=Dev Notes Platform
   After=network-online.target
   Wants=network-online.target
   [Service]
   Type=simple
   User=devnotes
   Group=devnotes
   WorkingDirectory=/opt/dev-notes
   EnvironmentFile=/etc/dev-notes.env
   ExecStart=/opt/dev-notes/bin/dev-notes
   Restart=on-failure
   RestartSec=5
   NoNewPrivileges=true
   PrivateTmp=true
   ProtectSystem=strict
   ReadWritePaths=/opt/dev-notes/data
   [Install]
   WantedBy=multi-user.target
   ```
5. 启动并自检：
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now dev-notes
   sudo systemctl status dev-notes
   curl -s http://127.0.0.1:8080/healthz
   # 期望 {"db":"ok","status":"ok","version":"1.4.10"}
   ```

**1Panel 界面配合**：「文件」上传二进制、编辑 env；「终端」粘贴上面第 4、5 步命令；「网站」建反向代理；「计划任务」跑备份脚本。systemd 服务注册目前仍需命令，1Panel 没有纯点击入口。

## 五、路线 B：Docker 镜像 + 1Panel Compose

1. WSL 内先登录阿里云仓库（只登录一次，凭据由 Docker 管理）：
   ```bash
   docker login --username=<阿里云账号> registry.cn-hangzhou.aliyuncs.com
   ```
2. 构建并推送 amd64 镜像：
   ```bash
   bash scripts/push-image.sh
   # 版本取 app/VERSION（当前 1.4.10）；可附加 --tag 覆盖，如 --tag v1.4.10
   ```
   脚本复用 `build.sh` 的版本校验，由 `deploy/Dockerfile` 在镜像内编译前后端；**只构建 `linux/amd64`**，用 `DOCKER_BUILDKIT=0 docker build` + `docker push` 推送。Docker 29 默认走 BuildKit，会附带 OCI 空 manifest（`application/vnd.oci.empty.v1+json`），阿里云 ACR 拒收；legacy builder 不附带该类型，产物为不含 attestation 的 OCI image manifest，已验证可正常拉取与运行。
3. 准备变量文件：把 `deploy/.env.example` 复制为 `deploy/.env`，填 `SESSION_SECRET`（≥32 位）与 `ADMIN_INIT_PASSWORD`（≥12 位）。`.env` 已在 `.gitignore` 中，不会入库。
4. 把数据放进 compose 项目目录的 `./data/`（`notes.db` 与 `media/`），并设属主为容器运行用户 `uid=100:gid=101`：`chown -R 100:101 ./data`。实测容器内 `app` 用户是 uid=100、gid=101（alpine 中 gid 100 已被占用，`addgroup -S app` 分到 101）；属主不对会导致启动时数据库只读或直接失败。空库首次启动也建议先建好 `./data/` 并 chown。
5. 1Panel「容器 → Compose → 创建」，粘贴 `deploy/docker-compose.yml` 内容；确认数据卷 `./data:/app/data`、端口映射 `127.0.0.1:8080:8080`。
6. 容器启动后看日志与健康检查，`curl -s http://127.0.0.1:8080/healthz` 应返回与 compose 中 tag 相同的版本号（当前 `1.4.10`）。
7. 升级：改 compose 中镜像 tag（如 `:1.5.0`），1Panel 点「重新创建容器」，数据卷不变。回滚则改回旧 tag。

> ℹ️ `1.4.10` 与 `latest` 已于 2026-11-27 推送阿里云（`application/vnd.oci.image.manifest.v1+json`，digest `sha256:1e7ae6c2…`），可直接上机；后续升版本时记得先执行第 2 步的 `scripts/push-image.sh`，否则 1Panel 拉到新 tag 会失败。**首次启动会自动把库迁到 `user_version=4`**（补 `note.is_featured` 并回填旧推荐位），迁移前建议先备份 `notes.db`。

`deploy/.env.example` 不可删除；compose 用 `${VAR:?}` 强制变量存在，缺变量时容器拒绝启动，避免弱密码上线。

## 六、环境变量

| 变量 | 必填 | 说明 |
|---|---|---|
| `SESSION_SECRET` | 是 | ≥32 位随机串，`openssl rand -hex 32` 生成 |
| `ADMIN_INIT_USERNAME` / `ADMIN_INIT_PASSWORD` | 首次 | 仅 `admin` 表为空时写入；密码 ≥12 位，改密走后台「账号设置」 |
| `DATA_DIR` | 否 | 默认 `/app/data`，内含 `notes.db` + `media/` |
| `HTTP_ADDR` | 否 | 默认 `:8080` |
| `COOKIE_SECURE` | 否 | 默认 `auto`：按请求协议判断 Secure；全 HTTPS 可设 `true`，HTTP 调试设 `false` |
| `APP_ENV` / `LOG_LEVEL` | 否 | 默认 `production` / `info` |

## 七、1Panel 建站与备案

1. 「网站 → 创建 → 反向代理」，目标 `127.0.0.1:8080`。
2. 确认代理**放行全部路径**：`/notes/<slug>`、`/admin/notes` 等深路径靠 Go 回退 `index.html`。
3. 备案通过前：不解析域名、不开 80/443；本地用 SSH 隧道 `ssh -L 8080:127.0.0.1:8080 <用户>@<ECS_IP>` 验收（Windows 浏览器用 `localhost:8080`）。
4. 备案通过后：域名 A 记录指向 ECS，安全组放行 80/443，1Panel 申请 Let's Encrypt 并开启 HTTPS/强制跳转；如全程 HTTPS，把 `COOKIE_SECURE` 改为 `true` 并重启服务/容器。
5. 后台「站点设置」填真实 ICP 备案号，页脚即时更新；确认前台无草稿、无下载入口。

## 八、故障排查

| 症状 | 原因 / 处理 |
|---|---|
| 容器启动即退出 | 数据卷目录权限不足；或 compose 的必填变量未填（缺 `SESSION_SECRET` / `ADMIN_INIT_PASSWORD`） |
| 后台登录失败 | 首次初始账号未写入；忘记密码时进后台「账号设置」或删 admin 表记录后重启重建 |
| 登录后立即跳回登录页 | HTTP 下 `COOKIE_SECURE=true` 被浏览器拒收；设回 `auto` |
| 反向代理 502 | 服务未监听 8080，或端口只绑了 `127.0.0.1` 而容器/代理网络不通 |
| 登录提示失败次数过多 | 同 IP 15 分钟内失败 5 次，等待或重启清零（计数在内存） |
| 保存内容被拒绝并列出命中词 | 内容含经营性词汇或下载类表述，按提示改为「源码仓库」等合规表述 |
| 前台显示草稿 | 检查条目状态，确认接口只查 `published` |
| 推送报 `unknown manifest class for application/vnd.oci.empty.v1+json` | Docker 29 默认 BuildKit 附带 OCI 空 manifest，阿里云 ACR 拒收；用 `DOCKER_BUILDKIT=0`（脚本已设），勿改回默认 |
| 镜像推送后更新检测报 404 | 单架构 OCI 单 manifest 可能被只认 docker v2s2 / OCI index 的检测误判；本项目走 1Panel 手动指定 tag，不受影响 |
