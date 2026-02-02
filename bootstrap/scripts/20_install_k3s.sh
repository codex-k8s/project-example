#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib.sh"

require_root
load_env "${ROOT_DIR}"

if systemctl list-unit-files | grep -q '^k3s\.service'; then
  log "k3s already installed; skip install"
else
  log "Install k3s (flannel none, disable k3s netpol, disable local-storage, disable traefik & servicelb)..."
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server \
    --write-kubeconfig-mode 600 \
    --cluster-cidr ${POD_SUPER_CIDR} \
    --service-cidr ${SERVICE_CIDR} \
    --flannel-backend=none \
    --disable-network-policy \
    --disable traefik \
    --disable servicelb \
    --disable local-storage" sh -
fi

wait_node_ready
