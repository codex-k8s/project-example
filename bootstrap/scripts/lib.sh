#!/usr/bin/env bash
set -euo pipefail

log() { echo "[$(date -Is)] $*"; }
die() { echo "ERROR: $*" >&2; exit 1; }

require_root() { [[ "${EUID}" -eq 0 ]] || die "Run as root"; }

# Load config.env and export for envsubst
load_env() {
  local root_dir="$1"
  [[ -f "${root_dir}/config.env" ]] || die "Missing ${root_dir}/config.env"
  set -a
  # shellcheck disable=SC1090
  source "${root_dir}/config.env"
  set +a
}

templates_dir() {
  local root_dir="$1"
  echo "${root_dir}/templates"
}

apply_tpl() {
  local root_dir="$1"
  local tpl="$2"
  envsubst < "$(templates_dir "$root_dir")/${tpl}" | kubectl apply -f -
}

env_namespaces() {
  local project
  for project in ${PROJECTS}; do
    echo "${project}-prod"
    echo "${project}-staging"
    echo "${project}-ai-staging"
    local n
    for n in $(seq 1 "${AI_DEV_SLOTS}"); do
      echo "${project}-ai-dev-${n}"
    done
  done
}

registry_namespaces() {
  env_namespaces
  echo "actions-runner-system"
}

project_index() {
  local target="$1"
  local idx=0
  local p
  for p in ${PROJECTS}; do
    if [[ "$p" == "$target" ]]; then
      echo "$idx"
      return 0
    fi
    idx=$((idx+1))
  done
  return 1
}

project_base_octet() {
  # idx=0 -> 64; idx=1 -> 72; idx=2 -> 80; idx=3 -> 88
  local project="$1"
  local idx
  idx="$(project_index "$project")" || die "Unknown project: $project"
  echo $((64 + idx*8))
}

kube_env() {
  export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
}

wait_node_ready() {
  kube_env
  log "Waiting node Ready..."
  kubectl wait --for=condition=Ready node --all --timeout=300s
}

get_node_ips() {
  kube_env
  local internal external
  internal="$(kubectl get node -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' | awk '{print $1}' || true)"
  external="$(kubectl get node -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}' | awk '{print $1}' || true)"
  echo "${internal} ${external}"
}

build_ipblock_yaml_4sp() {
  # Prints list items for NetworkPolicy "from:" with 4-space indent
  local out=""
  local ip
  for ip in "$@"; do
    [[ -n "${ip}" ]] || continue
    out+=$'    - ipBlock:\n'
    out+=$'        cidr: '"${ip}"$'/32\n'
  done
  printf "%b" "${out}"
}

build_yaml_list_4sp() {
  # Prints list items with 4-space indent ("    - item")
  local out=""
  local item
  for item in "$@"; do
    [[ -n "${item}" ]] || continue
    out+="    - ${item}"$'\n'
  done
  printf "%b" "${out}"
}
