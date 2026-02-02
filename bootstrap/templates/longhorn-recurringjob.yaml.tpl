apiVersion: longhorn.io/v1beta2
kind: RecurringJob
metadata:
  name: ${JOB_NAME}
  namespace: longhorn-system
spec:
  cron: "${JOB_CRON}"
  task: "${JOB_TASK}"
  groups:
  - "${JOB_GROUP}"
  retain: ${JOB_RETAIN}
  concurrency: ${JOB_CONCURRENCY}
