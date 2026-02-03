#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Использование:
  dupl-go.sh [--service <path>] [--threshold <n>]

Поведение:
  - без --service: ищет дубли во всех Go-модулях (go.mod) в репозитории
  - с --service: ищет дубли только в указанной директории (или модулях внутри неё)

По умолчанию:
  threshold берётся из tools/lint/dupl.yaml (fallback: 120)

Зависимости:
  - dupl в PATH
USAGE
}

service=""
threshold=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --service) service="$2"; shift 2 ;;
    --threshold) threshold="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Неизвестный аргумент: $1"; usage; exit 2 ;;
  esac
done

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cfg="$root/tools/lint/dupl.yaml"

if ! command -v dupl >/dev/null 2>&1; then
  echo "Не найден dupl в PATH. Запусти: make install-lint-tools"
  exit 2
fi

if [[ -z "$threshold" ]]; then
  threshold="$(sed -n 's/^threshold:[[:space:]]*//p' "$cfg" 2>/dev/null | head -n 1 | tr -d '\r')"
  threshold="${threshold:-120}"
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
  while IFS= read -r -d '' m; do mods+=("$(dirname "$m")"); done < <(find "$root" -name go.mod -print0 | sort -z)
fi

if [[ ${#mods[@]} -eq 0 ]]; then
  echo "go.mod не найден(ы), dupl пропущен."
  exit 0
fi

failed=0
for mod in "${mods[@]}"; do
  rel="${mod#$root/}"
  echo "dupl: $rel (threshold=$threshold)"

  tmp_list="$(mktemp)"
  # Исключаем generated/vendor/node_modules и прочие сборочные каталоги.
  # dupl не умеет exclude, поэтому кормим список файлов через stdin.
  find "$mod" \
    -type d \( \
      -name vendor -o -name node_modules -o -name dist -o -name build -o -name .git -o -name .cache -o -name generated \
    \) -prune -false \
    -o -type f -name '*.go' -print \
    | sort > "$tmp_list"

  if [[ ! -s "$tmp_list" ]]; then
    rm -f "$tmp_list"
    echo "dupl: нет go-файлов (после исключений)"
    continue
  fi

  # dupl возвращает exit code 0 даже при найденных дублях, поэтому проверяем вывод.
  out="$(dupl -plumbing -files -t "$threshold" < "$tmp_list")"
  if [[ -n "$out" ]]; then
    echo "$out"
    failed=1
  else
    echo "dupl: дублей не найдено."
  fi
  rm -f "$tmp_list"
done

if [[ "$failed" -ne 0 ]]; then
  echo "Ошибка: найдены дубли (dupl)."
  exit 1
fi

echo "Готово: дубли не найдены."
