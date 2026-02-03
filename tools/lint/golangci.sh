#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Использование:
  golangci.sh [--service <path>]

Поведение:
  - без --service: найдёт все go.mod в репозитории и прогонит golangci-lint по каждому модулю
  - с --service: прогонит только в указанной директории (если там есть go.mod) или внутри неё

Зависимости:
  - golangci-lint в PATH
  - конфиг: .golangci.yml в корне репозитория
USAGE
}

service=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --service) service="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Неизвестный аргумент: $1"; usage; exit 2 ;;
  esac
done

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cfg="$root/.golangci.yml"

if [[ ! -f "$cfg" ]]; then
  echo "Не найден конфиг golangci-lint: $cfg"
  exit 2
fi
if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "Не найден golangci-lint в PATH. Запусти: make install-lint-tools"
  exit 2
fi

declare -a mods=()
if [[ -n "$service" ]]; then
  svc_abs="$root/$service"
  if [[ ! -d "$svc_abs" ]]; then
    echo "Директория не найдена: $service"
    exit 2
  fi
  if [[ -f "$svc_abs/go.mod" ]]; then
    mods+=("$svc_abs")
  else
    while IFS= read -r -d '' m; do mods+=("$(dirname "$m")"); done < <(find "$svc_abs" -name go.mod -print0 | sort -z)
  fi
else
  while IFS= read -r -d '' m; do mods+=("$(dirname "$m")"); done < <(find "$root" -name go.mod -not -path "$root/.local/old-svc/*" -print0 | sort -z)
fi

if [[ ${#mods[@]} -eq 0 ]]; then
  echo "go.mod не найден(ы), golangci-lint пропущен."
  exit 0
fi

failed=0
for mod in "${mods[@]}"; do
  rel="${mod#$root/}"
  echo "golangci-lint: $rel"
  (cd "$mod" && golangci-lint run -c "$cfg" ./...) || failed=1
done

if [[ "$failed" -ne 0 ]]; then
  echo "Ошибка: golangci-lint завершился с ошибками."
  exit 1
fi

echo "Готово: golangci-lint без ошибок."
