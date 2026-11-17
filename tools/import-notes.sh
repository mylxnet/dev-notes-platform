#!/usr/bin/env bash
# 把 content/notes/*.md 通过管理 API 导入站点。
#
# 用法：
#   ADMIN_PASSWORD=你的密码 bash tools/import-notes.sh
#   ADMIN_PASSWORD=... API=http://localhost:8080 NOTES_DIR=content/notes bash tools/import-notes.sh
#   ADMIN_PASSWORD=... MODE=update bash tools/import-notes.sh      # 已存在的 slug 覆盖正文
#
# 依赖：bash、curl、python3。文章格式见 content/notes/README 说明（首尾 --- 包住的头部字段）。
set -euo pipefail

API="${API:-http://127.0.0.1:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-}"
NOTES_DIR="${NOTES_DIR:-content/notes}"
MODE="${MODE:-create}"

if [[ -z "$ADMIN_PASSWORD" ]]; then
  echo "缺少 ADMIN_PASSWORD 环境变量（管理员密码）" >&2
  exit 1
fi
if ! command -v curl >/dev/null || ! command -v python3 >/dev/null; then
  echo "需要 curl 与 python3" >&2
  exit 1
fi
if [[ ! -d "$NOTES_DIR" ]]; then
  echo "找不到文章目录：$NOTES_DIR（请在仓库根目录执行）" >&2
  exit 1
fi

WORK="$(mktemp -d)"
JAR="$WORK/cookies.txt"
trap 'rm -rf "$WORK"' EXIT

json_field() {
  # $1=文件 $2=python 表达式
  python3 -c "import json,sys;d=json.load(open(sys.argv[1]));print($2)" "$1"
}

api_post() {
  # $1=path $2=body文件 $3=csrf
  curl -sS -b "$JAR" -H "Content-Type: application/json" -H "X-CSRF-Token: $3" \
    -d @"$2" -o "$WORK/resp.json" -w '%{http_code}' "$API$1"
}

echo "[import] API：$API"
echo "[import] 文章目录：$NOTES_DIR（模式：$MODE）"

# 1. 登录拿会话与 CSRF token
if ! curl -sS -c "$JAR" -H 'Content-Type: application/json' \
  -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" \
  -o "$WORK/login.json" "$API/api/admin/login"; then
  echo "登录请求失败，确认 API 地址与服务状态" >&2
  exit 1
fi
CSRF="$(json_field "$WORK/login.json" "d['data']['csrf_token']")"
echo "[import] 已登录：$ADMIN_USERNAME"

# 2. 项目 slug -> id
curl -sS -b "$JAR" -o "$WORK/projects.json" "$API/api/admin/projects"
python3 - "$WORK/projects.json" "$WORK/project_map.tsv" <<'PY'
import json, sys
data = json.load(open(sys.argv[1]))['data']['items']
with open(sys.argv[2], 'w', encoding='utf-8') as out:
    for item in data:
        out.write(f"{item['slug']}\t{item['id']}\n")
PY

# 3. 现有笔记 slug -> id
curl -sS -b "$JAR" -o "$WORK/notes.json" "$API/api/admin/notes?page_size=100"
python3 - "$WORK/notes.json" "$WORK/note_map.tsv" <<'PY'
import json, sys
data = json.load(open(sys.argv[1]))['data']['items']
with open(sys.argv[2], 'w', encoding='utf-8') as out:
    for item in data:
        out.write(f"{item['slug']}\t{item['id']}\n")
PY

created=0
updated=0
skipped=0
failed=0

shopt -s nullglob
for md in "$NOTES_DIR"/*.md; do
  # 4. 解析头部与正文，生成请求体
  if ! python3 - "$md" "$WORK/project_map.tsv" "$WORK/body.json" <<'PY'
import json, sys

path, project_map, out_path = sys.argv[1], sys.argv[2], sys.argv[3]
raw = open(path, encoding='utf-8').read()
if not raw.startswith('---'):
    sys.exit('缺少头部（首行应为 ---）')
_, head, body = raw.split('---', 2)

meta = {}
for line in head.strip().splitlines():
    if not line.strip():
        continue
    key, _, value = line.partition(':')
    meta[key.strip()] = value.strip()

projects = {}
for line in open(project_map, encoding='utf-8'):
    slug, _, pid = line.partition('\t')
    projects[slug.strip()] = int(pid) if pid.strip() else None

for required in ('title', 'slug', 'summary'):
    if not meta.get(required):
        sys.exit(f'头部缺少字段：{required}')

payload = {
    'title': meta['title'],
    'slug': meta['slug'],
    'summary': meta['summary'],
    'content': body.lstrip('\n'),
    'project_id': projects.get(meta.get('project', '')),
    'status': meta.get('status', 'draft'),
    'published_at': meta.get('published_at', ''),
}
with open(out_path, 'w', encoding='utf-8') as out:
    json.dump(payload, out, ensure_ascii=False)
PY
  then
    echo "  ✗ 解析失败：$md"
    failed=$((failed + 1))
    continue
  fi

  slug="$(json_field "$WORK/body.json" "d['slug']")"
  title="$(json_field "$WORK/body.json" "d['title']")"

  # 5. 已存在则按模式处理
  existing_id="$(awk -F'\t' -v s="$slug" '$1==s {print $2}' "$WORK/note_map.tsv" || true)"
  if [[ -n "$existing_id" ]]; then
    if [[ "$MODE" != "update" ]]; then
      echo "  · 跳过（已存在）：$slug"
      skipped=$((skipped + 1))
      continue
    fi
    code="$(api_post "/api/admin/notes/$existing_id" "$WORK/body.json" "$CSRF")"
    if [[ "$code" == "200" ]]; then
      echo "  ↻ 已更新：$title"
      updated=$((updated + 1))
    else
      echo "  ✗ 更新失败（HTTP $code）：$slug → $(python3 -c "import json;d=json.load(open('$WORK/resp.json'));print(d.get('error',{}).get('message',''))" 2>/dev/null || true)"
      failed=$((failed + 1))
    fi
    continue
  fi

  code="$(api_post "/api/admin/notes" "$WORK/body.json" "$CSRF")"
  if [[ "$code" == "201" ]]; then
    echo "  + 已创建：$title"
    created=$((created + 1))
  else
    echo "  ✗ 创建失败（HTTP $code）：$slug → $(python3 -c "import json;d=json.load(open('$WORK/resp.json'));print(d.get('error',{}).get('message',''))" 2>/dev/null || true)"
    failed=$((failed + 1))
  fi
done

echo "[import] 完成：创建 $created，更新 $updated，跳过 $skipped，失败 $failed"
[[ "$failed" -eq 0 ]]
