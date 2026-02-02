#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib.sh"

require_root
load_env "${ROOT_DIR}"

log "Disable swap..."
swapoff -a || true
sed -ri 's/^([^#].*\sswap\s.*)$/#\1/g' /etc/fstab || true

log "Sysctl for k8s..."
modprobe br_netfilter || true
cat >/etc/sysctl.d/99-k8s.conf <<'SYSCTL'
net.ipv4.ip_forward=1
net.bridge.bridge-nf-call-iptables=1
net.bridge.bridge-nf-call-ip6tables=1
SYSCTL
sysctl --system >/dev/null

log "Install OS deps (includes envsubst via gettext-base)..."
apt-get update -y
apt-get install -y curl ca-certificates jq tar iptables open-iscsi nfs-common gettext-base
systemctl enable --now iscsid

log "Install Helm (if missing)..."
if ! command -v helm >/dev/null 2>&1; then
  tmp="$(mktemp -d)"
  curl -fsSL -o "${tmp}/get-helm-3" https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
  chmod 700 "${tmp}/get-helm-3"
  "${tmp}/get-helm-3"
  rm -rf "${tmp}"
fi

log "Prereqs OK"
