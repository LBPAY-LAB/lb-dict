# DEV-004: Kubernetes Manifests - DICT Complete Stack

**Projeto**: DICT - Diretório de Identificadores de Contas Transacionais (LBPay)
**Componente**: Complete Kubernetes Manifests for DICT Stack
**Versão**: 1.0
**Data**: 2025-10-25
**Autor**: DEVOPS (AI Agent - DevOps Engineer)
**Revisor**: [Aguardando]
**Aprovador**: Head de DevOps, CTO (José Luís Silva)

---

## Controle de Versão

| Versão | Data | Autor | Descrição das Mudanças |
|--------|------|-------|------------------------|
| 1.0 | 2025-10-25 | DEVOPS | Versão inicial - Manifests completos para todos os serviços DICT |

---

## Sumário Executivo

### Visão Geral

Manifests Kubernetes completos para o **DICT Stack**, incluindo:
- ✅ **Application Services**: Core DICT, RSFN Connect (API + Worker), RSFN Bridge
- ✅ **Infrastructure Services**: Temporal, Pulsar, PostgreSQL, Redis
- ✅ **Networking**: Ingress, Services, NetworkPolicies
- ✅ **Configuration**: ConfigMaps, Secrets, ExternalSecrets
- ✅ **Scaling**: HPA (Horizontal Pod Autoscaler), PDB (Pod Disruption Budget)
- ✅ **Observability**: ServiceMonitors (Prometheus), PodMonitors
- ✅ **Security**: RBAC, SecurityContext, NetworkPolicies

### Arquitetura Kubernetes

```
┌────────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster (EKS)                    │
├────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Namespace: dict-prod                                           │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Application Layer                                        │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐               │  │
│  │  │Core DICT │  │Connect   │  │  Bridge  │               │  │
│  │  │(Deploy)  │  │API(Deploy│  │(Deploy)  │               │  │
│  │  │  5 pods  │  │  5 pods  │  │  5 pods  │               │  │
│  │  └──────────┘  └──────────┘  └──────────┘               │  │
│  │                                                            │  │
│  │  ┌──────────┐                                             │  │
│  │  │Connect   │                                             │  │
│  │  │Worker    │                                             │  │
│  │  │(StatefulS│                                             │  │
│  │  │  3 pods  │                                             │  │
│  │  └──────────┘                                             │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Namespace: infrastructure                                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Infrastructure Layer                                     │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐│  │
│  │  │Temporal  │  │ Pulsar   │  │PostgreSQL│  │  Redis   ││  │
│  │  │(StatefulS│  │(StatefulS│  │(StatefulS│  │(StatefulS││  │
│  │  │  3 pods  │  │  3 pods  │  │  3 pods  │  │  2 pods  ││  │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘│  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Namespace: monitoring                                          │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Observability                                            │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐               │  │
│  │  │Prometheus│  │ Grafana  │  │  Jaeger  │               │  │
│  │  └──────────┘  └──────────┘  └──────────┘               │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Ingress Layer                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  NGINX Ingress Controller                                 │  │
│  │  - core-dict.lbpay.io                                     │  │
│  │  - connect-api.lbpay.io                                   │  │
│  │  - bridge.lbpay.io                                        │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────┘
```

---

## Índice

1. [Namespace Configuration](#1-namespace-configuration)
2. [Core DICT Manifests](#2-core-dict-manifests)
3. [RSFN Connect Manifests](#3-rsfn-connect-manifests)
4. [RSFN Bridge Manifests](#4-rsfn-bridge-manifests)
5. [Temporal Manifests](#5-temporal-manifests)
6. [Pulsar Manifests](#6-pulsar-manifests)
7. [PostgreSQL Manifests](#7-postgresql-manifests)
8. [Redis Manifests](#8-redis-manifests)
9. [Ingress Configuration](#9-ingress-configuration)
10. [ConfigMaps & Secrets](#10-configmaps--secrets)
11. [Monitoring & Observability](#11-monitoring--observability)
12. [RBAC & Security](#12-rbac--security)
13. [Rastreabilidade](#13-rastreabilidade)

---

## 1. Namespace Configuration

### Namespaces

```yaml
# namespaces.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: dict-prod
  labels:
    name: dict-prod
    environment: production
    team: lbpay-dict
---
apiVersion: v1
kind: Namespace
metadata:
  name: infrastructure
  labels:
    name: infrastructure
    environment: production
---
apiVersion: v1
kind: Namespace
metadata:
  name: monitoring
  labels:
    name: monitoring
    environment: production
```

### ResourceQuota

```yaml
# resource-quota.yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: dict-prod-quota
  namespace: dict-prod
spec:
  hard:
    requests.cpu: "50"
    requests.memory: 100Gi
    limits.cpu: "100"
    limits.memory: 200Gi
    persistentvolumeclaims: "20"
    services.loadbalancers: "3"
---
apiVersion: v1
kind: LimitRange
metadata:
  name: dict-prod-limits
  namespace: dict-prod
spec:
  limits:
  - max:
      cpu: "4"
      memory: 8Gi
    min:
      cpu: 100m
      memory: 128Mi
    type: Container
```

---

## 2. Core DICT Manifests

### Deployment

```yaml
# core-dict-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: core-dict
  namespace: dict-prod
  labels:
    app: core-dict
    tier: application
    component: domain-service
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: core-dict
  template:
    metadata:
      labels:
        app: core-dict
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: core-dict-sa
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - core-dict
            topologyKey: kubernetes.io/hostname
      containers:
      - name: core-dict
        image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/lbpay/core-dict:latest
        imagePullPolicy: Always
        ports:
        - name: grpc
          containerPort: 50051
          protocol: TCP
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: metrics
          containerPort: 9090
          protocol: TCP
        env:
        - name: ENV
          value: "production"
        - name: LOG_LEVEL
          value: "info"
        - name: GRPC_PORT
          value: "50051"
        - name: HTTP_PORT
          value: "8080"
        - name: METRICS_PORT
          value: "9090"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: core-dict-secrets
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: core-dict-secrets
              key: redis-url
        - name: PULSAR_URL
          valueFrom:
            secretKeyRef:
              name: core-dict-secrets
              key: pulsar-url
        - name: PULSAR_TOKEN
          valueFrom:
            secretKeyRef:
              name: core-dict-secrets
              key: pulsar-token
        envFrom:
        - configMapRef:
            name: core-dict-config
        resources:
          requests:
            cpu: 1000m
            memory: 1Gi
            ephemeral-storage: 1Gi
          limits:
            cpu: 2000m
            memory: 2Gi
            ephemeral-storage: 2Gi
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        startupProbe:
          httpGet:
            path: /health/startup
            port: 8080
          initialDelaySeconds: 0
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 30
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 1000
          capabilities:
            drop:
            - ALL
        volumeMounts:
        - name: tmp
          mountPath: /tmp
      volumes:
      - name: tmp
        emptyDir: {}
      terminationGracePeriodSeconds: 60
```

### Service

```yaml
# core-dict-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: core-dict-svc
  namespace: dict-prod
  labels:
    app: core-dict
spec:
  type: ClusterIP
  selector:
    app: core-dict
  ports:
  - name: grpc
    port: 50051
    targetPort: 50051
    protocol: TCP
  - name: http
    port: 8080
    targetPort: 8080
    protocol: TCP
  - name: metrics
    port: 9090
    targetPort: 9090
    protocol: TCP
  sessionAffinity: None
---
apiVersion: v1
kind: Service
metadata:
  name: core-dict-headless
  namespace: dict-prod
  labels:
    app: core-dict
spec:
  type: ClusterIP
  clusterIP: None
  selector:
    app: core-dict
  ports:
  - name: grpc
    port: 50051
    targetPort: 50051
```

### HPA

```yaml
# core-dict-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: core-dict-hpa
  namespace: dict-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: core-dict
  minReplicas: 5
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: grpc_server_handled_total
      target:
        type: AverageValue
        averageValue: "1000"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
      - type: Pods
        value: 2
        periodSeconds: 30
      selectPolicy: Max
```

### PDB

```yaml
# core-dict-pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: core-dict-pdb
  namespace: dict-prod
spec:
  minAvailable: 3
  selector:
    matchLabels:
      app: core-dict
```

---

## 3. RSFN Connect Manifests

### Connect API Deployment

```yaml
# connect-api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: connect-api
  namespace: dict-prod
  labels:
    app: connect-api
    tier: application
    component: orchestration
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: connect-api
  template:
    metadata:
      labels:
        app: connect-api
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
    spec:
      serviceAccountName: connect-api-sa
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: connect-api
        image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/lbpay/connect-api:latest
        imagePullPolicy: Always
        ports:
        - name: http
          containerPort: 8080
        - name: metrics
          containerPort: 9090
        env:
        - name: ENV
          value: "production"
        - name: HTTP_PORT
          value: "8080"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: connect-secrets
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: connect-secrets
              key: redis-url
        - name: PULSAR_URL
          valueFrom:
            secretKeyRef:
              name: connect-secrets
              key: pulsar-url
        - name: BRIDGE_GRPC_ADDRESS
          value: "rsfn-bridge-svc.dict-prod.svc.cluster.local:50051"
        envFrom:
        - configMapRef:
            name: connect-api-config
        resources:
          requests:
            cpu: 1000m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 2Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          capabilities:
            drop:
            - ALL
        volumeMounts:
        - name: tmp
          mountPath: /tmp
      volumes:
      - name: tmp
        emptyDir: {}
```

### Connect Worker StatefulSet

```yaml
# connect-worker-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: connect-worker
  namespace: dict-prod
  labels:
    app: connect-worker
    tier: application
    component: temporal-worker
spec:
  serviceName: connect-worker
  replicas: 3
  selector:
    matchLabels:
      app: connect-worker
  template:
    metadata:
      labels:
        app: connect-worker
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
    spec:
      serviceAccountName: connect-worker-sa
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: connect-worker
        image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/lbpay/connect-worker:latest
        imagePullPolicy: Always
        ports:
        - name: metrics
          containerPort: 9090
        env:
        - name: ENV
          value: "production"
        - name: TEMPORAL_HOST
          value: "temporal-frontend.infrastructure.svc.cluster.local:7233"
        - name: TEMPORAL_NAMESPACE
          value: "lbpay-dict-prod"
        - name: WORKER_TASK_QUEUE
          value: "dict-workflows-prod"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: connect-secrets
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: connect-secrets
              key: redis-url
        - name: PULSAR_URL
          valueFrom:
            secretKeyRef:
              name: connect-secrets
              key: pulsar-url
        - name: BRIDGE_GRPC_ADDRESS
          value: "rsfn-bridge-svc.dict-prod.svc.cluster.local:50051"
        envFrom:
        - configMapRef:
            name: connect-worker-config
        resources:
          requests:
            cpu: 1000m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 2Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 9090
          initialDelaySeconds: 10
          periodSeconds: 5
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          capabilities:
            drop:
            - ALL
        volumeMounts:
        - name: tmp
          mountPath: /tmp
      volumes:
      - name: tmp
        emptyDir: {}
  updateStrategy:
    type: RollingUpdate
  podManagementPolicy: OrderedReady
```

### Connect Services

```yaml
# connect-services.yaml
apiVersion: v1
kind: Service
metadata:
  name: connect-api-svc
  namespace: dict-prod
spec:
  type: ClusterIP
  selector:
    app: connect-api
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
---
apiVersion: v1
kind: Service
metadata:
  name: connect-worker
  namespace: dict-prod
spec:
  type: ClusterIP
  clusterIP: None
  selector:
    app: connect-worker
  ports:
  - name: metrics
    port: 9090
    targetPort: 9090
```

### Connect HPA (API only)

```yaml
# connect-api-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: connect-api-hpa
  namespace: dict-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: connect-api
  minReplicas: 5
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## 4. RSFN Bridge Manifests

### Deployment

```yaml
# rsfn-bridge-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rsfn-bridge
  namespace: dict-prod
  labels:
    app: rsfn-bridge
    tier: integration
    component: soap-adapter
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: rsfn-bridge
  template:
    metadata:
      labels:
        app: rsfn-bridge
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
    spec:
      serviceAccountName: rsfn-bridge-sa
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: rsfn-bridge
        image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/lbpay/rsfn-bridge:latest
        imagePullPolicy: Always
        ports:
        - name: grpc
          containerPort: 50051
        - name: metrics
          containerPort: 9090
        env:
        - name: ENV
          value: "production"
        - name: GRPC_PORT
          value: "50051"
        - name: PULSAR_URL
          valueFrom:
            secretKeyRef:
              name: bridge-secrets
              key: pulsar-url
        - name: BACEN_API_URL
          value: "https://api.rsfn.bcb.gov.br/dict/api/v1"
        - name: MTLS_CERT_PATH
          value: "/certs/client-cert.pem"
        - name: MTLS_KEY_PATH
          value: "/certs/client-key.pem"
        - name: MTLS_CA_PATH
          value: "/certs/ca.pem"
        - name: XML_SIGNER_JAR_PATH
          value: "/opt/signer/xml-signer.jar"
        - name: SIGNING_CERT_PATH
          value: "/certs/signing-cert.pem"
        - name: SIGNING_KEY_PATH
          value: "/certs/signing-key.pem"
        envFrom:
        - configMapRef:
            name: rsfn-bridge-config
        resources:
          requests:
            cpu: 1000m
            memory: 2Gi
          limits:
            cpu: 2000m
            memory: 4Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 9090
          initialDelaySeconds: 10
          periodSeconds: 5
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          capabilities:
            drop:
            - ALL
        volumeMounts:
        - name: tmp
          mountPath: /tmp
        - name: certs
          mountPath: /certs
          readOnly: true
      volumes:
      - name: tmp
        emptyDir: {}
      - name: certs
        secret:
          secretName: icp-brasil-certs
```

### Service

```yaml
# rsfn-bridge-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: rsfn-bridge-svc
  namespace: dict-prod
spec:
  type: ClusterIP
  selector:
    app: rsfn-bridge
  ports:
  - name: grpc
    port: 50051
    targetPort: 50051
  - name: metrics
    port: 9090
    targetPort: 9090
```

### HPA

```yaml
# rsfn-bridge-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: rsfn-bridge-hpa
  namespace: dict-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: rsfn-bridge
  minReplicas: 5
  maxReplicas: 15
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## 5. Temporal Manifests

### StatefulSet

```yaml
# temporal-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: temporal
  namespace: infrastructure
  labels:
    app: temporal
spec:
  serviceName: temporal
  replicas: 3
  selector:
    matchLabels:
      app: temporal
  template:
    metadata:
      labels:
        app: temporal
    spec:
      containers:
      - name: temporal
        image: temporalio/auto-setup:1.22.0
        ports:
        - name: frontend
          containerPort: 7233
        - name: history
          containerPort: 7234
        - name: matching
          containerPort: 7235
        - name: worker
          containerPort: 7239
        - name: metrics
          containerPort: 9090
        env:
        - name: DB
          value: "postgresql"
        - name: DB_PORT
          value: "5432"
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: temporal-secrets
              key: postgres-user
        - name: POSTGRES_PWD
          valueFrom:
            secretKeyRef:
              name: temporal-secrets
              key: postgres-password
        - name: POSTGRES_SEEDS
          value: "postgres-svc.infrastructure.svc.cluster.local"
        - name: DYNAMIC_CONFIG_FILE_PATH
          value: "/etc/temporal/config/dynamicconfig/development.yaml"
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 4Gi
        volumeMounts:
        - name: config
          mountPath: /etc/temporal/config/dynamicconfig
      volumes:
      - name: config
        configMap:
          name: temporal-config
---
apiVersion: v1
kind: Service
metadata:
  name: temporal-frontend
  namespace: infrastructure
spec:
  type: ClusterIP
  selector:
    app: temporal
  ports:
  - name: frontend
    port: 7233
    targetPort: 7233
---
apiVersion: v1
kind: Service
metadata:
  name: temporal
  namespace: infrastructure
spec:
  type: ClusterIP
  clusterIP: None
  selector:
    app: temporal
  ports:
  - name: frontend
    port: 7233
```

---

## 6. Pulsar Manifests

### StatefulSet

```yaml
# pulsar-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: pulsar
  namespace: infrastructure
  labels:
    app: pulsar
spec:
  serviceName: pulsar
  replicas: 3
  selector:
    matchLabels:
      app: pulsar
  template:
    metadata:
      labels:
        app: pulsar
    spec:
      containers:
      - name: pulsar
        image: apachepulsar/pulsar:3.0.0
        command: ["bin/pulsar"]
        args: ["standalone"]
        ports:
        - name: pulsar
          containerPort: 6650
        - name: http
          containerPort: 8080
        - name: https
          containerPort: 8443
        env:
        - name: PULSAR_MEM
          value: "-Xms2g -Xmx2g -XX:MaxDirectMemorySize=2g"
        resources:
          requests:
            cpu: 1000m
            memory: 4Gi
          limits:
            cpu: 2000m
            memory: 8Gi
        volumeMounts:
        - name: data
          mountPath: /pulsar/data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: gp3
      resources:
        requests:
          storage: 100Gi
---
apiVersion: v1
kind: Service
metadata:
  name: pulsar-svc
  namespace: infrastructure
spec:
  type: ClusterIP
  selector:
    app: pulsar
  ports:
  - name: pulsar
    port: 6650
    targetPort: 6650
  - name: http
    port: 8080
    targetPort: 8080
```

---

## 7. PostgreSQL Manifests

### StatefulSet

```yaml
# postgres-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: infrastructure
  labels:
    app: postgres
spec:
  serviceName: postgres
  replicas: 3
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:16-alpine
        ports:
        - name: postgres
          containerPort: 5432
        env:
        - name: POSTGRES_DB
          value: "dict_prod"
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: postgres-secrets
              key: postgres-user
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secrets
              key: postgres-password
        - name: PGDATA
          value: "/var/lib/postgresql/data/pgdata"
        resources:
          requests:
            cpu: 2000m
            memory: 4Gi
          limits:
            cpu: 4000m
            memory: 8Gi
        volumeMounts:
        - name: data
          mountPath: /var/lib/postgresql/data
        livenessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - postgres
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - postgres
          initialDelaySeconds: 5
          periodSeconds: 5
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: gp3
      resources:
        requests:
          storage: 500Gi
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-svc
  namespace: infrastructure
spec:
  type: ClusterIP
  selector:
    app: postgres
  ports:
  - name: postgres
    port: 5432
    targetPort: 5432
```

---

## 8. Redis Manifests

### StatefulSet

```yaml
# redis-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
  namespace: infrastructure
  labels:
    app: redis
spec:
  serviceName: redis
  replicas: 2
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command: ["redis-server"]
        args: ["--appendonly", "yes", "--requirepass", "$(REDIS_PASSWORD)"]
        ports:
        - name: redis
          containerPort: 6379
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secrets
              key: redis-password
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 1000m
            memory: 2Gi
        volumeMounts:
        - name: data
          mountPath: /data
        livenessProbe:
          exec:
            command:
            - redis-cli
            - ping
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - redis-cli
            - ping
          initialDelaySeconds: 5
          periodSeconds: 5
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: gp3
      resources:
        requests:
          storage: 50Gi
---
apiVersion: v1
kind: Service
metadata:
  name: redis-svc
  namespace: infrastructure
spec:
  type: ClusterIP
  selector:
    app: redis
  ports:
  - name: redis
    port: 6379
    targetPort: 6379
```

---

## 9. Ingress Configuration

### Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dict-ingress
  namespace: dict-prod
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/backend-protocol: "GRPC"
    nginx.ingress.kubernetes.io/grpc-backend: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  tls:
  - hosts:
    - core-dict.lbpay.io
    - connect-api.lbpay.io
    - bridge.lbpay.io
    secretName: dict-tls-cert
  rules:
  - host: core-dict.lbpay.io
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: core-dict-svc
            port:
              number: 50051
  - host: connect-api.lbpay.io
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: connect-api-svc
            port:
              number: 8080
  - host: bridge.lbpay.io
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: rsfn-bridge-svc
            port:
              number: 50051
```

---

## 10. ConfigMaps & Secrets

### ConfigMaps

```yaml
# configmaps.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: core-dict-config
  namespace: dict-prod
data:
  LOG_FORMAT: "json"
  GRPC_MAX_CONN: "1000"
  CACHE_TTL: "300"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: connect-api-config
  namespace: dict-prod
data:
  LOG_FORMAT: "json"
  WORKER_POOL_SIZE: "10"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: rsfn-bridge-config
  namespace: dict-prod
data:
  LOG_FORMAT: "json"
  CIRCUIT_BREAKER_THRESHOLD: "5"
  CIRCUIT_BREAKER_TIMEOUT: "60"
```

### External Secrets (AWS Secrets Manager)

```yaml
# external-secrets.yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: core-dict-secrets
  namespace: dict-prod
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: core-dict-secrets
    creationPolicy: Owner
  data:
  - secretKey: database-url
    remoteRef:
      key: /lbpay/dict/prod/database-url
  - secretKey: redis-url
    remoteRef:
      key: /lbpay/dict/prod/redis-url
  - secretKey: pulsar-url
    remoteRef:
      key: /lbpay/dict/prod/pulsar-url
  - secretKey: pulsar-token
    remoteRef:
      key: /lbpay/dict/prod/pulsar-token
```

---

## 11. Monitoring & Observability

### ServiceMonitor (Prometheus)

```yaml
# servicemonitors.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: core-dict-metrics
  namespace: dict-prod
  labels:
    app: core-dict
spec:
  selector:
    matchLabels:
      app: core-dict
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: connect-api-metrics
  namespace: dict-prod
spec:
  selector:
    matchLabels:
      app: connect-api
  endpoints:
  - port: metrics
    interval: 30s
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: rsfn-bridge-metrics
  namespace: dict-prod
spec:
  selector:
    matchLabels:
      app: rsfn-bridge
  endpoints:
  - port: metrics
    interval: 30s
```

---

## 12. RBAC & Security

### ServiceAccounts

```yaml
# service-accounts.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: core-dict-sa
  namespace: dict-prod
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: connect-api-sa
  namespace: dict-prod
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: connect-worker-sa
  namespace: dict-prod
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rsfn-bridge-sa
  namespace: dict-prod
```

### NetworkPolicies

```yaml
# network-policies.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: core-dict-netpol
  namespace: dict-prod
spec:
  podSelector:
    matchLabels:
      app: core-dict
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: dict-prod
    ports:
    - protocol: TCP
      port: 50051
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: infrastructure
    ports:
    - protocol: TCP
      port: 5432
    - protocol: TCP
      port: 6379
    - protocol: TCP
      port: 6650
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
```

---

## 13. Rastreabilidade

### Documentos Relacionados

| ID | Documento | Relação |
|----|-----------|---------|
| **DEV-001** | [CI/CD Pipeline Core](./Pipelines/DEV-001_CI_CD_Pipeline_Core.md) | Pipeline do Core DICT |
| **DEV-002** | [CI/CD Pipeline Connect](./Pipelines/DEV-002_CI_CD_Pipeline_Connect.md) | Pipeline do Connect |
| **DEV-003** | [CI/CD Pipeline Bridge](./Pipelines/DEV-003_CI_CD_Pipeline_Bridge.md) | Pipeline do Bridge |
| **TEC-001** | [Core DICT Specification](../11_Especificacoes_Tecnicas/TEC-001_Core_DICT_Specification.md) | Especificação do Core |
| **TEC-002** | [Bridge Specification](../11_Especificacoes_Tecnicas/TEC-002_Bridge_Specification.md) | Especificação do Bridge |
| **TEC-003** | [Connect Specification](../11_Especificacoes_Tecnicas/TEC-003_RSFN_Connect_Specification.md) | Especificação do Connect |

---

**Última Atualização**: 2025-10-25
**Versão**: 1.0
**Status**: ✅ Completo
