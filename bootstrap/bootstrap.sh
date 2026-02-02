#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

chmod +x "${ROOT_DIR}/scripts/"*.sh

"${ROOT_DIR}/scripts/00_prereqs.sh"
"${ROOT_DIR}/scripts/10_generate_k3s_registries.sh"
"${ROOT_DIR}/scripts/20_install_k3s.sh"
"${ROOT_DIR}/scripts/30_install_calico_and_networking.sh"
"${ROOT_DIR}/scripts/40_install_addons_certmanager_metallb.sh"
"${ROOT_DIR}/scripts/50_install_longhorn_and_backups_A.sh"
"${ROOT_DIR}/scripts/60_deploy_namespace_registries.sh"
"${ROOT_DIR}/scripts/61_update_registry_hosts.sh"
"${ROOT_DIR}/scripts/70_install_velero.sh"
"${ROOT_DIR}/scripts/80_host_firewall.sh"

echo "DONE"
