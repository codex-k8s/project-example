#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib.sh"

require_root
load_env "${ROOT_DIR}"

mkdir -p /etc/rancher/k3s

# Build mirrors YAML for every namespace registry:
# "registry.<ns>.svc.cluster.local:5000" -> endpoint http://registry.<ns>.svc.cluster.local:5000
REGISTRY_MIRRORS_YAML=""
while read -r ns; do
  host="registry.${ns}.svc.cluster.local:${REGISTRY_PORT}"
  REGISTRY_MIRRORS_YAML+=$'  "'"${host}"$'":\n'
  REGISTRY_MIRRORS_YAML+=$'    endpoint:\n'
  REGISTRY_MIRRORS_YAML+=$'      - "http://'"${host}"$'"\n'
done < <(env_namespaces)

export REGISTRY_MIRRORS_YAML

log "Render /etc/rancher/k3s/registries.yaml ..."
envsubst < "${ROOT_DIR}/templates/k3s-registries.yaml.tpl" > /etc/rancher/k3s/registries.yaml

log "registries.yaml generated."
