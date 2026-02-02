#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib.sh"

require_root
load_env "${ROOT_DIR}"
kube_env

log "Install cert-manager..."
kubectl apply -f "https://github.com/cert-manager/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml"
kubectl -n cert-manager rollout status deploy/cert-manager --timeout=600s || true

log "Install MetalLB..."
kubectl apply -f "https://raw.githubusercontent.com/metallb/metallb/${METALLB_VERSION}/config/manifests/metallb-native.yaml"

log "Configure MetalLB L2 pool..."
# METALLB_L2_ADDRESSES is space-separated
# shellcheck disable=SC2206
addr_arr=(${METALLB_L2_ADDRESSES})
[[ "${#addr_arr[@]}" -gt 0 ]] || die "METALLB_L2_ADDRESSES is empty in config.env"
METALLB_ADDRESSES_YAML="$(build_yaml_list_4sp "${addr_arr[@]}")"
export METALLB_ADDRESSES_YAML

apply_tpl "${ROOT_DIR}" "metallb-l2.yaml.tpl"

log "cert-manager + MetalLB done"
