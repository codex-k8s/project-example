#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Использование:
  gen-asyncapi-ts.sh --app <path> --spec <path> [--out <dir>]

По умолчанию:
  --out = src/shared/ws/generated

Требования:
  - установлен `asyncapi` (AsyncAPI CLI). Для установки: make install-codegen-tools
USAGE
}

app=""
spec=""
out="src/shared/ws/generated"

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
  echo "AsyncAPI spec не найден: $spec"
  exit 2
fi
if ! command -v asyncapi >/dev/null 2>&1; then
  echo "Не найден asyncapi в PATH. Запусти: make install-codegen-tools"
  exit 2
fi

(
  cd "$app_abs"
  mkdir -p "$out"
  asyncapi generate models typescript "$spec_abs" \
    -o "$out" \
    --tsModelType interface \
    --tsEnumType union \
    --tsModuleSystem ESM \
    --tsExportType named \
    --tsIncludeComments \
    --no-interactive
)

echo "Готово: сгенерировано (AsyncAPI -> TS): $app/$out"

