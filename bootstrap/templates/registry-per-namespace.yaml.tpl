apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: registry-data
  namespace: ${NS_NAME}
  labels:
    app: registry
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: ${REGISTRY_STORAGECLASS}
  resources:
    requests:
      storage: ${REGISTRY_PVC_SIZE}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: registry
  namespace: ${NS_NAME}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: registry
  template:
    metadata:
      labels:
        app: registry
    spec:
      containers:
      - name: registry
        image: registry:2
        ports:
        - containerPort: ${REGISTRY_PORT}
        env:
        - name: REGISTRY_HTTP_ADDR
          value: "0.0.0.0:${REGISTRY_PORT}"
        resources:
          requests:
            cpu: "50m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
        volumeMounts:
        - name: data
          mountPath: /var/lib/registry
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: registry-data
---
apiVersion: v1
kind: Service
metadata:
  name: registry
  namespace: ${NS_NAME}
  labels:
    app: registry
spec:
  type: ClusterIP
  selector:
    app: registry
  ports:
  - name: registry
    port: ${REGISTRY_PORT}
    targetPort: ${REGISTRY_PORT}
