apiVersion: v1
kind: Secret
metadata:
  name: longhorn-s3-cred
  namespace: longhorn-system
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: "${LH_S3_ACCESS_KEY}"
  AWS_SECRET_ACCESS_KEY: "${LH_S3_SECRET_KEY}"
  AWS_ENDPOINTS: "${LH_S3_ENDPOINT}"
  VIRTUAL_HOSTED_STYLE: "${LH_S3_VIRTUAL_HOSTED_STYLE}"
