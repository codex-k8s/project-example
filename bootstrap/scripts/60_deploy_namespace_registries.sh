#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib.sh"

require_root
load_env "${ROOT_DIR}"
kube_env

log "Deploy registry in each env namespace..."

while read -r ns; do
  export NS_NAME="${ns}"
  export REGISTRY_PVC_SIZE
  export REGISTRY_PORT
  export REGISTRY_STORAGECLASS
  apply_tpl "${ROOT_DIR}" "registry-per-namespace.yaml.tpl"
done < <(registry_namespaces)

log "Registries deployed"
