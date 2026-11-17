#!/usr/bin/env bash

# 测试构建：登录页预填固定测试账号与密码，仅用于本机验收。
# 用法：bash scripts/build-test.sh [输出目录]（默认 app/build）
# 注意：本脚本打出的产物含明文凭据，禁止部署到公网；对外发布必须用 scripts/build.sh。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

export DN_TEST_BUILD=1
export VITE_TEST_LOGIN=1
export VITE_TEST_USERNAME="${VITE_TEST_USERNAME:-admin}"
export VITE_TEST_PASSWORD="${VITE_TEST_PASSWORD:-admin12345678}"

log() { printf '[build-test] %s\n' "$*"; }

log '测试构建：登录页将预填测试凭据，产物禁止对外发布'
log "账号：${VITE_TEST_USERNAME}（密码由构建变量注入，不在脚本回显）"

bash "$REPO_ROOT/scripts/build.sh" "$@"

log '完成：测试包已生成（禁止对外发布）'
