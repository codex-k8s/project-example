#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib.sh"

require_root
load_env "${ROOT_DIR}"

[[ "${ENABLE_HOST_FIREWALL}" == "true" ]] || { log "Firewall disabled"; exit 0; }

if [[ -z "${EXT_IF}" ]]; then
  EXT_IF="$(ip route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="dev"){print $(i+1); exit}}' || true)"
fi
[[ -n "${EXT_IF}" ]] || die "Cannot determine EXT_IF. Set EXT_IF in config.env"

log "Install host firewall rules (outside only 22/80/443; pods -> host any port)..."

cat >/usr/local/sbin/host-firewall-k8s.sh <<EOF2
#!/usr/bin/env bash
set -euo pipefail
EXT_IF="${EXT_IF}"
POD_SUPER_CIDR="${POD_SUPER_CIDR}"
SSH_ALLOW_CIDR="${SSH_ALLOW_CIDR}"
HTTP_ALLOW_CIDR="${HTTP_ALLOW_CIDR}"
HTTPS_ALLOW_CIDR="${HTTPS_ALLOW_CIDR}"

iptables -N HOSTFW-IN >/dev/null 2>&1 || true
iptables -F HOSTFW-IN

iptables -A HOSTFW-IN -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
iptables -A HOSTFW-IN -i lo -j ACCEPT
iptables -A HOSTFW-IN -p icmp -j ACCEPT

# Pods -> host any port
iptables -A HOSTFW-IN -s "${POD_SUPER_CIDR}" -j ACCEPT

# Internet -> host only 22/80/443
iptables -A HOSTFW-IN -i "${EXT_IF}" -p tcp -s "${SSH_ALLOW_CIDR}" --dport 22 -j ACCEPT
iptables -A HOSTFW-IN -i "${EXT_IF}" -p tcp -s "${HTTP_ALLOW_CIDR}" --dport 80 -j ACCEPT
iptables -A HOSTFW-IN -i "${EXT_IF}" -p tcp -s "${HTTPS_ALLOW_CIDR}" --dport 443 -j ACCEPT

iptables -A HOSTFW-IN -j DROP
iptables -C INPUT -j HOSTFW-IN >/dev/null 2>&1 || iptables -I INPUT 1 -j HOSTFW-IN
EOF2

chmod +x /usr/local/sbin/host-firewall-k8s.sh

cat >/etc/systemd/system/host-firewall-k8s.service <<'EOF2'
[Unit]
Description=Host firewall rules for k8s
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/sbin/host-firewall-k8s.sh
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOF2

systemctl daemon-reload
systemctl enable --now host-firewall-k8s.service

log "Firewall installed"
