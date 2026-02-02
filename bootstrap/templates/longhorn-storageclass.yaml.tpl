apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ${SC_NAME}
  annotations:
    storageclass.kubernetes.io/is-default-class: "${SC_DEFAULT}"
provisioner: driver.longhorn.io
allowVolumeExpansion: true
reclaimPolicy: Delete
volumeBindingMode: Immediate
parameters:
  numberOfReplicas: "${SC_REPLICAS}"
  fsType: "ext4"
  staleReplicaTimeout: "30"
${SC_EXTRA_PARAMS_YAML}
