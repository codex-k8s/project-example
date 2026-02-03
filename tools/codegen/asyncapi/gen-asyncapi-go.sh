#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Использование:
  gen-asyncapi-go.sh --service <path> [--spec <path>] [--out <dir>] [--package <name>]

По умолчанию:
  --spec    = api/server/asyncapi.yaml
  --out     = internal/transport/async/generated
  --package = generated

Требования:
  - установлен `asyncapi` (AsyncAPI CLI). Для установки: make install-codegen-tools
USAGE
}

service=""
spec=""
out="internal/transport/async/generated"
pkg="generated"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --service) service="$2"; shift 2 ;;
    --spec) spec="$2"; shift 2 ;;
    --out) out="$2"; shift 2 ;;
    --package) pkg="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Неизвестный аргумент: $1"; usage; exit 2 ;;
  esac
done

if [[ -z "$service" ]]; then
  echo "--service обязателен"
  usage
  exit 2
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
svc_abs="$root/$service"
if [[ ! -d "$svc_abs" ]]; then
  echo "Сервис не найден: $service"
  exit 2
fi

if ! command -v asyncapi >/dev/null 2>&1; then
  echo "Не найден asyncapi в PATH. Запусти: make install-codegen-tools"
  exit 2
fi

spec="${spec:-api/server/asyncapi.yaml}"
spec_abs="$svc_abs/$spec"
if [[ ! -f "$spec_abs" ]]; then
  echo "AsyncAPI spec не найден: $spec (ожидали файл: $spec_abs)"
  exit 2
fi

mkdir -p "$svc_abs/$out"

(
  cd "$svc_abs"
  asyncapi generate models go "$spec" \
    -o "$out" \
    --packageName "$pkg" \
    --goIncludeTags \
    --goIncludeComments \
    --no-interactive
)

echo "Готово: сгенерировано (AsyncAPI -> Go): $service/$out"

