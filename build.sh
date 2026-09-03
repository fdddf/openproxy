#!/bin/bash
set -e

COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
NAME="ghcr.io/fdddf/openproxy"

# execute the docker build with name and push
function docker_build_and_push() {
  local name=$1
  local tag=$2
  local dockerfile=$3
  local context=$4

  docker build --build-arg BUILD_COMMIT=${tag} --build-arg BUILD_DATE=${BUILD_DATE} -t ${name}:${tag} -f ${dockerfile} ${context}
  docker push ${name}:${tag}
}

docker_build_and_push $NAME $COMMIT "Dockerfile" .

# config.yaml holds local secrets and is gitignored; fall back to the sample so
# a clean checkout can still render the manifest.
CONFIG_SRC="src/config.yaml"
if [ ! -f "$CONFIG_SRC" ]; then
  CONFIG_SRC="src/config.example.yaml"
fi

cat <<EOF > deploy_k8s.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: openproxy-config
data:
  config.yaml: |
$(sed 's/^/    /' "$CONFIG_SRC")

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openproxy
  labels:
    app: openproxy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: openproxy
  template:
    metadata:
      labels:
        app: openproxy
    spec:
      volumes:
      - name: config
        configMap:
          name: openproxy-config
      containers:
        - name: openproxy
          image: ${NAME}:${COMMIT}
          imagePullPolicy: Always
          command:
            - /app/openproxy
            - --config=/etc/openproxy/config.yaml
          ports:
            - containerPort: 8081
              name: http
          volumeMounts:
          - name: config
            mountPath: /etc/openproxy
            readOnly: true

---
apiVersion: v1
kind: Service
metadata:
  name: openproxy
  labels:
    app: openproxy
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 8081
    targetPort: 8081
  selector:
    app: openproxy

EOF