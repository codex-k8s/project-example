apiVersion: v1
kind: ConfigMap
metadata:
  name: longhorn-default-resource
  namespace: longhorn-system
data:
  default-resource.yaml: |
    "backup-target": "s3://${LH_S3_BUCKET}@${LH_S3_REGION}/${LH_S3_PREFIX}/"
    "backup-target-credential-secret": "longhorn-s3-cred"
    "backupstore-poll-interval": "300"
