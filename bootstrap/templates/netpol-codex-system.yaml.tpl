apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-from-env-namespaces
  namespace: ${NS_NAME}
spec:
  podSelector: {}
  policyTypes: [Ingress]
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          scope: env
    - namespaceSelector:
        matchLabels:
          scope: platform
