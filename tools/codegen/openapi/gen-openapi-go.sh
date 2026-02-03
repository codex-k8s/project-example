#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Использование:
  gen-openapi-go.sh --service <path> [--spec <path>] [--out <file.go>]

Переменные (опционально):
  OPENAPI_CONFIG   путь до oapi-codegen config (по умолчанию: tools/codegen/openapi/configs/go_echo_server.yaml)
  OPENAPI_TEMPLATES путь до user templates dir (по умолчанию: tools/codegen/openapi/templates)

Пример:
  tools/codegen/openapi/gen-openapi-go.sh --service services/external/my-service
USAGE
}

service=""
spec=""
out_override=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --service) service="$2"; shift 2 ;;
    --spec) spec="$2"; shift 2 ;;
    --out) out_override="$2"; shift 2 ;;
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

spec="${spec:-api/server/api.yaml}"
spec_abs="$svc_abs/$spec"
if [[ ! -f "$spec_abs" ]]; then
  echo "OpenAPI spec не найден: $spec (ожидали файл: $spec_abs)"
  exit 2
fi

config="${OPENAPI_CONFIG:-tools/codegen/openapi/configs/go_echo_server.yaml}"
templates="${OPENAPI_TEMPLATES:-tools/codegen/openapi/templates}"

config_abs="$root/$config"
templates_abs="$root/$templates"

if [[ ! -f "$config_abs" ]]; then
  echo "Config не найден: $config_abs"
  exit 2
fi
if [[ ! -d "$templates_abs" ]]; then
  echo "Templates dir не найден: $templates_abs"
  exit 2
fi

# mkdir -p для output из config (сервис-локальный относительный путь)
# Парсим простой top-level `output: ...` без внешних зависимостей (yq/python).
out_rel="$(sed -n 's/^output:[[:space:]]*//p' "$config_abs" | head -n 1 | tr -d '\r')"
if [[ -z "$out_rel" ]]; then
  echo "Не смогли прочитать output из config: $config_abs"
  exit 2
fi

config_to_use="$config_abs"
if [[ -n "$out_override" ]]; then
  # Делаем временный config с переопределением output, чтобы агент мог указать путь.
  # Важно: output должен указывать на `**/generated/**` (см. дизайн-гайды).
  tmp_cfg="$(mktemp)"
  awk -v out="$out_override" '
    BEGIN { done=0 }
    /^output:[[:space:]]*/ && done==0 { print "output: " out; done=1; next }
    { print }
  ' "$config_abs" > "$tmp_cfg"
  config_to_use="$tmp_cfg"
  out_rel="$out_override"
fi

mkdir -p "$svc_abs/$(dirname "$out_rel")"

(
  cd "$svc_abs"
  # Важно: используем templates override, чтобы echo import был v5.
  go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest \
    -templates "$templates_abs" \
    -config "$config_to_use" \
    "$spec"
)

echo "Готово: сгенерировано: $service/$out_rel"

if [[ -n "${tmp_cfg:-}" ]]; then
  rm -f "$tmp_cfg"
fi
