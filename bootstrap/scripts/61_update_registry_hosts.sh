#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib.sh"

require_root
load_env "${ROOT_DIR}"
kube_env

BEGIN_MARK="# BEGIN K3S-REGISTRY-SVC"
END_MARK="# END K3S-REGISTRY-SVC"

tmp="$(mktemp)"

# Strip old block if exists
awk -v b="${BEGIN_MARK}" -v e="${END_MARK}" '
  $0==b {skip=1; next}
  $0==e {skip=0; next}
  !skip {print}
' /etc/hosts > "${tmp}"

{
  echo "${BEGIN_MARK}"
  while read -r ns; do
    ip="$(kubectl -n "${ns}" get svc registry -o jsonpath='{.spec.clusterIP}' 2>/dev/null || true)"
    [[ -n "${ip}" && "${ip}" != "None" ]] || continue
    echo "${ip} registry.${ns}.svc.cluster.local"
  done < <(env_namespaces)
  echo "${END_MARK}"
} >> "${tmp}"

cp "${tmp}" /etc/hosts
rm -f "${tmp}"

log "/etc/hosts updated for registry.<ns>.svc.cluster.local"
