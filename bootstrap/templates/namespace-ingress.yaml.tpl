apiVersion: v1
kind: Namespace
metadata:
  name: ingress
  labels:
    role: ingress
  annotations:
    cni.projectcalico.org/ipv4pools: '${CALICO_IPV4POOLS_JSON}'
