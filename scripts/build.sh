#!/usr/bin/env bash
# 构建前端与后端：前端产物输出到 app/internal/webui/dist，随后以 embedui 标签嵌入静态二进制。
# 用法：scripts/build.sh [输出目录]，输出目录默认 app/build
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$REPO_ROOT/app"
WEB_DIR="$REPO_ROOT/web"
VERSION="$(tr -d '\r\n' < "$APP_DIR/VERSION")"
OUTPUT_DIR="${1:-$APP_DIR/build}"
NPM_REGISTRY="${NPM_REGISTRY:-https://registry.npmmirror.com}"

log() { printf '[build] %s\n' "$*"; }

log "版本：$VERSION"

# 防误用：正式构建不允许携带测试预填开关，凭据只能由 build-test.sh 显式注入。
if [ "${VITE_TEST_LOGIN:-}" = "1" ] && [ "${DN_TEST_BUILD:-}" != "1" ]; then
	printf '错误：正式构建不允许 VITE_TEST_LOGIN=1；测试包请用 scripts/build-test.sh\n' >&2
	exit 1
fi

# 版本一致性：main.go 的默认值必须与 app/VERSION 相同，避免交付物版本错位。
if ! grep -q "var version = \"$VERSION\"" "$APP_DIR/main.go"; then
	printf '错误：app/main.go 的 version 与 app/VERSION（%s）不一致\n' "$VERSION" >&2
	exit 1
fi

# 前端版本来源（界面版本行）与 package.json 也要同步，否则页面写的行为会和产物对不上。
if ! grep -q "APP_VERSION = '$VERSION'" "$WEB_DIR/src/version.ts"; then
	printf '错误：web/src/version.ts 的 APP_VERSION 与 app/VERSION（%s）不一致\n' "$VERSION" >&2
	exit 1
fi

if ! grep -q "\"version\": \"$VERSION\"" "$WEB_DIR/package.json"; then
	printf '错误：web/package.json 的 version 与 app/VERSION（%s）不一致\n' "$VERSION" >&2
	exit 1
fi

log "安装前端依赖（$NPM_REGISTRY）"
(cd "$WEB_DIR" && npm ci --registry="$NPM_REGISTRY")

log "构建前端"
(cd "$WEB_DIR" && npm run build)

log "编译后端（CGO_ENABLED=0，embedui）"
mkdir -p "$OUTPUT_DIR"
(cd "$APP_DIR" && CGO_ENABLED=0 go build -tags embedui -trimpath \
	-ldflags "-s -w -X main.version=$VERSION" \
	-o "$OUTPUT_DIR/dev-notes" .)

log "完成：$OUTPUT_DIR/dev-notes"
