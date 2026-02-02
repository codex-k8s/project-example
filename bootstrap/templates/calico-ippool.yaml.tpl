apiVersion: projectcalico.org/v3
kind: IPPool
metadata:
  name: ${POOL_NAME}
spec:
  cidr: ${POOL_CIDR}
  vxlanMode: Always
  natOutgoing: true
  disabled: false
  assignmentMode: ${ASSIGNMENT_MODE}
