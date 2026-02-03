#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Использование:
  gen-openapi-ts.sh --app <path> --spec <path> [--out <dir>]

Пример:
  tools/codegen/openapi/gen-openapi-ts.sh \
    --app services/external/web-frontend \
    --spec services/external/chat-backend/api/server/api.yaml \
    --out src/shared/api/generated

Требования:
  - OpenAPI spec в YAML/JSON
USAGE
}

app=""
spec=""
out="src/shared/api/generated"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --app) app="$2"; shift 2 ;;
    --spec) spec="$2"; shift 2 ;;
    --out) out="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Неизвестный аргумент: $1"; usage; exit 2 ;;
  esac
done

if [[ -z "$app" || -z "$spec" ]]; then
  echo "--app и --spec обязательны"
  usage
  exit 2
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
app_abs="$root/$app"
spec_abs="$root/$spec"

if [[ ! -d "$app_abs" ]]; then
  echo "Frontend app не найден: $app"
  exit 2
fi
if [[ ! -f "$spec_abs" ]]; then
  echo "OpenAPI spec не найден: $spec"
  exit 2
fi
if ! command -v npx >/dev/null 2>&1; then
  echo "Не найден npx. Установи Node.js и повтори."
  exit 2
fi

(
  cd "$app_abs"
  mkdir -p "$out"
  # Генерация TS клиента/типов поверх axios.
  npx @hey-api/openapi-ts -i "$spec_abs" -o "$out" -c @hey-api/client-axios
)

echo "Готово: сгенерировано: $app/$out"
