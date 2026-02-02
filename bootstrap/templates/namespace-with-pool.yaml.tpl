apiVersion: v1
kind: Namespace
metadata:
  name: ${NS_NAME}
  annotations:
    cni.projectcalico.org/ipv4pools: '${CALICO_IPV4POOLS_JSON}'
