#!/usr/bin/env bash

# 构建并推送 Docker 镜像到阿里云仓库；复用本机 docker login 会话，脚本不保存凭据。
# 用法：bash scripts/push-image.sh [--tag 额外标签]
# 只构建 linux/amd64，用传统 legacy builder + push 产出 Docker v2s2 manifest，
# 兼容老工具与第三方更新检测（HEAD manifests/<tag> 不会因 Accept 白名单误判 404）。
# Docker 29 默认 docker build 走 BuildKit，会附带 OCI 空 manifest（阿里云 ACR 拒收），故设 DOCKER_BUILDKIT=0。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="$(tr -d '\r\n' < "$REPO_ROOT/app/VERSION")"
REGISTRY="${REGISTRY:-registry.cn-hangzhou.aliyuncs.com}"
NAMESPACE="${NAMESPACE:-mylxnet}"
IMAGE_NAME="${IMAGE_NAME:-dev-notes}"
NPM_REGISTRY="${NPM_REGISTRY:-https://registry.npmmirror.com}"
GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
IMAGE="$REGISTRY/$NAMESPACE/$IMAGE_NAME"
EXTRA_TAG=""

while [ $# -gt 0 ]; do
	case "$1" in
		--tag)
			EXTRA_TAG="${2:?--tag 需要一个标签值}"
			shift 2
			;;
		*)
			printf '未知参数：%s\n用法：bash scripts/push-image.sh [--tag 额外标签]\n' "$1" >&2
			exit 1
			;;
	esac
done

log() { printf '[push-image] %s\n' "$*"; }

if ! docker info >/dev/null 2>&1; then
	printf '错误：Docker 守护进程不可用\n' >&2
	exit 1
fi

log "构建正式包（$VERSION）"
bash "$REPO_ROOT/scripts/build.sh"

log "构建镜像：$IMAGE:$VERSION（linux/amd64，legacy builder → Docker v2s2 manifest）"
DOCKER_BUILDKIT=0 docker build \
	--build-arg "VERSION=$VERSION" \
	--build-arg "NPM_REGISTRY=$NPM_REGISTRY" \
	--build-arg "GOPROXY=$GOPROXY" \
	-f "$REPO_ROOT/deploy/Dockerfile" \
	-t "$IMAGE:$VERSION" \
	"$REPO_ROOT"

docker tag "$IMAGE:$VERSION" "$IMAGE:latest"
if [ -n "$EXTRA_TAG" ]; then
	docker tag "$IMAGE:$VERSION" "$IMAGE:$EXTRA_TAG"
fi

log "推送镜像：$IMAGE:$VERSION / $IMAGE:latest"
docker push "$IMAGE:$VERSION"
docker push "$IMAGE:latest"
if [ -n "$EXTRA_TAG" ]; then
	docker push "$IMAGE:$EXTRA_TAG"
fi

log "完成：$IMAGE:$VERSION / $IMAGE:latest"
log "1Panel：容器 → Compose → 创建，镜像填 $IMAGE:$VERSION"
