# 多阶段构建：前端在 node 中编译，Go 二进制固定按 linux/amd64 编译，镜像内带 embedui 静态前端。
# 为兼容阿里云 ACR（拒收 BuildKit 的 OCI 空 manifest），用 legacy builder 构建；
# legacy builder 不识别 BUILDPLATFORM/TARGETARCH 自动变量，故不使用 --platform 与平台 ARG。
# 依赖源可通过 build-arg 覆盖，默认国内镜像源。

# ---- 前端 ----
FROM node:22-alpine AS web
ARG NPM_REGISTRY=https://registry.npmmirror.com
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci --registry="$NPM_REGISTRY"
COPY web/ ./
RUN npm run build

# ---- 后端 ----
FROM golang:1.27-alpine AS build
ARG VERSION=0.0.0
ARG GOPROXY=https://goproxy.cn,direct
WORKDIR /src
ENV CGO_ENABLED=0 GOPROXY="$GOPROXY"
# go.mod 位于 app/，所以 app/ 的内容就是模块根，不能再套一层目录。
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY app/ ./
COPY --from=web /app/internal/webui/dist/ ./internal/webui/dist/
RUN GOOS=linux GOARCH=amd64 go build -tags embedui -trimpath \
    -ldflags "-s -w -X main.version=$VERSION" \
    -o /out/dev-notes .

# ---- 运行镜像 ----
FROM alpine:3.20
ARG VERSION=0.0.0
LABEL org.opencontainers.image.title="dev-notes-platform" \
      org.opencontainers.image.description="个人开发笔记与项目索引站" \
      org.opencontainers.image.authors="by Mr lin" \
      org.opencontainers.image.version="${VERSION}"

RUN addgroup -S app && adduser -S -G app app

COPY --from=build /out/dev-notes /app/dev-notes
RUN mkdir -p /app/data && chown -R app:app /app /app/data

WORKDIR /app
USER app

ENV APP_ENV=production \
    HTTP_ADDR=:8080 \
    DATA_DIR=/app/data \
    COOKIE_SECURE=auto \
    LOG_LEVEL=info

VOLUME ["/app/data"]
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/app/dev-notes"]
