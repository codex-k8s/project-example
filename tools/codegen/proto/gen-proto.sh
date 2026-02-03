#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Использование:
  gen-proto.sh --service <path> [--proto-root <dir>] [--proto <file.proto>] [--go-out <dir>] [--openapiv2-out <dir>] [--with-gateway] [--with-openapiv2]

По умолчанию:
  --proto-root = proto
  Go output      = internal/transport/grpc/generated
  OpenAPIv2 out  = api/server/generated/grpc/openapi

Примеры:
  make install-proto-tools
  make gen-proto-go SVC=services/external/my-svc WITH_GATEWAY=1 WITH_OPENAPIV2=1
  make gen-proto-go SVC=services/external/my-svc PROTO=proto/user/v1/user.proto WITH_OPENAPIV2=1

Требования:
  - установлен `protoc`
  - плагины в PATH: protoc-gen-go, protoc-gen-go-grpc, protoc-gen-grpc-gateway, protoc-gen-openapiv2
USAGE
}

service=""
proto_root=""
go_out_rel=""
openapiv2_out_rel=""
declare -a protos=()
with_gateway=0
with_openapiv2=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --service) service="$2"; shift 2 ;;
    --proto-root) proto_root="$2"; shift 2 ;;
    --proto) protos+=("$2"); shift 2 ;;
    --go-out) go_out_rel="$2"; shift 2 ;;
    --openapiv2-out) openapiv2_out_rel="$2"; shift 2 ;;
    --with-gateway) with_gateway=1; shift ;;
    --with-openapiv2) with_openapiv2=1; shift ;;
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

proto_root="${proto_root:-proto}"
proto_root_abs="$svc_abs/$proto_root"
if [[ ! -d "$proto_root_abs" ]]; then
  echo "proto-root не найден: $proto_root (ожидали директорию: $proto_root_abs)"
  exit 2
fi

if ! command -v protoc >/dev/null 2>&1; then
  echo "Не найден protoc. Установи protoc и повтори."
  exit 2
fi

required_plugins=(protoc-gen-go protoc-gen-go-grpc)
if [[ "$with_gateway" -eq 1 ]]; then required_plugins+=(protoc-gen-grpc-gateway); fi
if [[ "$with_openapiv2" -eq 1 ]]; then required_plugins+=(protoc-gen-openapiv2); fi
for p in "${required_plugins[@]}"; do
  if ! command -v "$p" >/dev/null 2>&1; then
    echo "Не найден $p в PATH. Запусти: make install-proto-tools"
    exit 2
  fi
done

go_out_rel="${go_out_rel:-internal/transport/grpc/generated}"
openapiv2_out_rel="${openapiv2_out_rel:-api/server/generated/grpc/openapi}"

mkdir -p "$svc_abs/$go_out_rel"
if [[ "$with_openapiv2" -eq 1 ]]; then
  mkdir -p "$svc_abs/$openapiv2_out_rel"
fi

declare -a files=()
if [[ ${#protos[@]} -gt 0 ]]; then
  for f in "${protos[@]}"; do
    if [[ -f "$svc_abs/$f" ]]; then
      files+=("$f")
      continue
    fi
    if [[ -f "$proto_root_abs/$f" ]]; then
      files+=("$proto_root/$f")
      continue
    fi
    echo "proto файл не найден: $f (искали: $svc_abs/$f и $proto_root_abs/$f)"
    exit 2
  done
else
  while IFS= read -r -d '' f; do
    # Сохраняем путь относительно service root, чтобы protoc корректно применил source_relative.
    rel="${f#$svc_abs/}"
    files+=("$rel")
  done < <(find "$proto_root_abs" -type f -name '*.proto' -print0 | sort -z)
fi

if [[ ${#files[@]} -eq 0 ]]; then
  echo "Не нашли .proto файлов в $proto_root_abs"
  exit 2
fi

cmd=(
  protoc
  -I "$proto_root_abs"
  --go_out "$svc_abs/$go_out_rel" --go_opt paths=source_relative
  --go-grpc_out "$svc_abs/$go_out_rel" --go-grpc_opt paths=source_relative
)

if [[ "$with_gateway" -eq 1 ]]; then
  cmd+=(--grpc-gateway_out "$svc_abs/$go_out_rel" --grpc-gateway_opt paths=source_relative)
fi

if [[ "$with_openapiv2" -eq 1 ]]; then
  cmd+=(--openapiv2_out "$svc_abs/$openapiv2_out_rel")
fi

(
  cd "$svc_abs"
  "${cmd[@]}" "${files[@]}"
)

echo "Готово: proto сгенерированы:"
echo "  - Go:       $service/$go_out_rel"
if [[ "$with_gateway" -eq 1 ]]; then
  echo "  - Gateway:  $service/$go_out_rel"
fi
if [[ "$with_openapiv2" -eq 1 ]]; then
  echo "  - OpenAPIv2:$service/$openapiv2_out_rel"
fi
