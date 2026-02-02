#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib.sh"

require_root
load_env "${ROOT_DIR}"
kube_env

if command -v velero >/dev/null 2>&1; then
  log "velero cli already installed"
else
  log "Install Velero CLI..."
  tmp="$(mktemp -d)"
  curl -fsSL -o "${tmp}/velero.tgz" "https://github.com/vmware-tanzu/velero/releases/download/${VELERO_VERSION}/velero-${VELERO_VERSION}-linux-amd64.tar.gz"
  tar -xzf "${tmp}/velero.tgz" -C "${tmp}"
  install -m 0755 "${tmp}/velero-${VELERO_VERSION}-linux-amd64/velero" /usr/local/bin/velero
  rm -rf "${tmp}"
fi

kubectl get ns velero >/dev/null 2>&1 || kubectl create ns velero

cat >/tmp/velero-credentials <<EOF2
[default]
aws_access_key_id=${VELERO_ACCESS_KEY}
aws_secret_access_key=${VELERO_SECRET_KEY}
EOF2

log "Install Velero server..."
velero install \
  --provider aws \
  --plugins "${VELERO_AWS_PLUGIN_IMAGE}" \
  --bucket "${VELERO_BUCKET}" \
  --prefix "${VELERO_PREFIX}" \
  --secret-file /tmp/velero-credentials \
  --backup-location-config "region=${VELERO_REGION},s3Url=${VELERO_S3_URL},s3ForcePathStyle=${VELERO_S3_FORCE_PATH_STYLE}" || true

log "Velero done"
