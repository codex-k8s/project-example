#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib.sh"

require_root
load_env "${ROOT_DIR}"
kube_env

log "Install Calico operator..."
kubectl apply -f "https://raw.githubusercontent.com/projectcalico/calico/${CALICO_VERSION}/manifests/operator-crds.yaml"
kubectl apply -f "https://raw.githubusercontent.com/projectcalico/calico/${CALICO_VERSION}/manifests/tigera-operator.yaml"

log "Apply Calico Installation CR..."
kubectl apply -f <(envsubst < "${ROOT_DIR}/templates/calico-installation.yaml.tpl")

log "Wait IPPool CRD..."
kubectl wait --for=condition=Established crd/ippools.crd.projectcalico.org --timeout=180s

log "Create platform IPPool (Automatic) ${PLATFORM_POOL_CIDR} ..."
export POOL_NAME="platform"
export POOL_CIDR="${PLATFORM_POOL_CIDR}"
export ASSIGNMENT_MODE="Automatic"
apply_tpl "${ROOT_DIR}" "calico-ippool.yaml.tpl"

log "Create namespaces (ingress + platform namespaces with platform pool annotation)..."
export CALICO_IPV4POOLS_JSON='["platform"]'
apply_tpl "${ROOT_DIR}" "namespace-ingress.yaml.tpl"

for ns in cert-manager metallb-system longhorn-system velero; do
  export NS_NAME="${ns}"
  export CALICO_IPV4POOLS_JSON='["platform"]'
  apply_tpl "${ROOT_DIR}" "namespace-with-pool.yaml.tpl"
done

log "Create per-project pools + namespaces..."
for project in ${PROJECTS}; do
  base="$(project_base_octet "${project}")"

  # prod / staging / ai-staging as /16 Manual
  for env in prod staging ai-staging; do
    case "${env}" in
      prod)      oct2="${base}" ;;
      staging)   oct2=$((base+1)) ;;
      ai-staging) oct2=$((base+2)) ;;
    esac

    export POOL_NAME="${project}-${env}-pool"
    export POOL_CIDR="10.${oct2}.0.0/16"
    export ASSIGNMENT_MODE="Manual"
    apply_tpl "${ROOT_DIR}" "calico-ippool.yaml.tpl"

    export NS_NAME="${project}-${env}"
    export CALICO_IPV4POOLS_JSON="[\"${POOL_NAME}\"]"
    apply_tpl "${ROOT_DIR}" "namespace-with-pool.yaml.tpl"
  done

  # ai-dev-N as /24 Manual in 10.(base+3).X.0/24
  for n in $(seq 1 "${AI_DEV_SLOTS}"); do
    oct3=$((n-1))
    export POOL_NAME="${project}-ai-dev-${n}-pool"
    export POOL_CIDR="10.$((base+3)).${oct3}.0/24"
    export ASSIGNMENT_MODE="Manual"
    apply_tpl "${ROOT_DIR}" "calico-ippool.yaml.tpl"

    export NS_NAME="${project}-ai-dev-${n}"
    export CALICO_IPV4POOLS_JSON="[\"${POOL_NAME}\"]"
    apply_tpl "${ROOT_DIR}" "namespace-with-pool.yaml.tpl"
  done
done

log "Apply env NetworkPolicies (deny cross-env ingress; allow same-ns + ingress ns + host)..."
read -r node_internal node_external < <(get_node_ips)
HOST_IPBLOCKS_YAML="$(build_ipblock_yaml_4sp "${node_internal}" "${node_external}")"
export HOST_IPBLOCKS_YAML

while read -r ns; do
  export NS_NAME="${ns}"
  apply_tpl "${ROOT_DIR}" "netpol-env.yaml.tpl"
done < <(env_namespaces)

log "Calico + networking done"
