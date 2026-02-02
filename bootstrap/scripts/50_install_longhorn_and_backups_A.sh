#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib.sh"

require_root
load_env "${ROOT_DIR}"
kube_env

log "Install Longhorn..."
kubectl apply -f "https://raw.githubusercontent.com/longhorn/longhorn/${LONGHORN_VERSION}/deploy/longhorn.yaml"
kubectl -n longhorn-system rollout status deploy/longhorn-ui --timeout=900s || true
kubectl -n longhorn-system rollout status deploy/longhorn-manager --timeout=900s || true

log "Configure Longhorn S3 backup target (default) ..."
apply_tpl "${ROOT_DIR}" "longhorn-s3-secret.yaml.tpl"
apply_tpl "${ROOT_DIR}" "longhorn-default-resource.yaml.tpl"

log "Create Longhorn RecurringJobs (method A baseline)..."
# PROD
export JOB_NAME="prod-snap-hourly";  export JOB_CRON="0 * * * *";   export JOB_TASK="snapshot"; export JOB_GROUP="prod";   export JOB_RETAIN="48"; export JOB_CONCURRENCY="2"; apply_tpl "${ROOT_DIR}" "longhorn-recurringjob.yaml.tpl"
export JOB_NAME="prod-backup-daily"; export JOB_CRON="15 2 * * *";  export JOB_TASK="backup";   export JOB_GROUP="prod";   export JOB_RETAIN="14"; export JOB_CONCURRENCY="1"; apply_tpl "${ROOT_DIR}" "longhorn-recurringjob.yaml.tpl"

# STAGING (and ai-staging)
export JOB_NAME="staging-snap-daily";  export JOB_CRON="0 3 * * *";  export JOB_TASK="snapshot"; export JOB_GROUP="staging"; export JOB_RETAIN="14"; export JOB_CONCURRENCY="2"; apply_tpl "${ROOT_DIR}" "longhorn-recurringjob.yaml.tpl"
export JOB_NAME="staging-backup-weekly"; export JOB_CRON="30 3 * * 0"; export JOB_TASK="backup"; export JOB_GROUP="staging"; export JOB_RETAIN="4"; export JOB_CONCURRENCY="1"; apply_tpl "${ROOT_DIR}" "longhorn-recurringjob.yaml.tpl"

# DEV (ai-dev-N)
export JOB_NAME="dev-snap-daily";   export JOB_CRON="0 4 * * *";   export JOB_TASK="snapshot"; export JOB_GROUP="dev";    export JOB_RETAIN="7"; export JOB_CONCURRENCY="2"; apply_tpl "${ROOT_DIR}" "longhorn-recurringjob.yaml.tpl"
export JOB_NAME="dev-backup-weekly"; export JOB_CRON="30 4 * * 0"; export JOB_TASK="backup";   export JOB_GROUP="dev";    export JOB_RETAIN="2"; export JOB_CONCURRENCY="1"; apply_tpl "${ROOT_DIR}" "longhorn-recurringjob.yaml.tpl"

log "Create Longhorn StorageClasses with recurringJobSelector (method A)..."
# Per Longhorn docs: recurringJobSelector JSON with name/isGroup
export SC_REPLICAS="${LONGHORN_REPLICAS}"

# longhorn-prod
export SC_NAME="longhorn-prod"
export SC_DEFAULT="false"
export SC_EXTRA_PARAMS_YAML="  recurringJobSelector: '[{\"name\":\"prod\",\"isGroup\":true}]'"
apply_tpl "${ROOT_DIR}" "longhorn-storageclass.yaml.tpl"

# longhorn-staging
export SC_NAME="longhorn-staging"
export SC_DEFAULT="false"
export SC_EXTRA_PARAMS_YAML="  recurringJobSelector: '[{\"name\":\"staging\",\"isGroup\":true}]'"
apply_tpl "${ROOT_DIR}" "longhorn-storageclass.yaml.tpl"

# longhorn-dev
export SC_NAME="longhorn-dev"
export SC_DEFAULT="false"
export SC_EXTRA_PARAMS_YAML="  recurringJobSelector: '[{\"name\":\"dev\",\"isGroup\":true}]'"
apply_tpl "${ROOT_DIR}" "longhorn-storageclass.yaml.tpl"

# longhorn-registry (no backups)
export SC_NAME="longhorn-registry"
export SC_DEFAULT="false"
export SC_EXTRA_PARAMS_YAML=""
apply_tpl "${ROOT_DIR}" "longhorn-storageclass.yaml.tpl"

log "Set default StorageClass..."
# Remove default from original "longhorn" (created by Longhorn)
kubectl patch storageclass longhorn -p '{"metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}' >/dev/null 2>&1 || true

case "${DEFAULT_SC}" in
  prod)    kubectl patch storageclass longhorn-prod -p '{"metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}' ;;
  staging) kubectl patch storageclass longhorn-staging -p '{"metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}' ;;
  dev)     kubectl patch storageclass longhorn-dev -p '{"metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}' ;;
  *) die "DEFAULT_SC must be prod|staging|dev" ;;
esac

log "Longhorn + backups A done"
