# Kubernetes Manifests - Core DICT

This directory contains Kubernetes manifests for deploying the Core DICT service to production.

## Files Overview

| File | Description |
|------|-------------|
| `namespace.yaml` | Creates the `dict-system` namespace |
| `configmap.yaml` | Non-sensitive configuration (ports, timeouts, feature flags) |
| `secret.yaml.example` | Template for sensitive data (DO NOT COMMIT REAL SECRETS) |
| `deployment.yaml` | Deployment, ServiceAccount, Role, RoleBinding |
| `service.yaml` | ClusterIP service + headless service |
| `hpa.yaml` | HorizontalPodAutoscaler (3-10 replicas) |
| `networkpolicy.yaml` | Network isolation rules |
| `pdb.yaml` | PodDisruptionBudget (min 2 available during updates) |

## Prerequisites

1. **Kubernetes Cluster**: v1.28+ with gRPC health probe support
2. **Secret Management**: Vault, Sealed Secrets, or External Secrets Operator
3. **Infrastructure**:
   - PostgreSQL 16+ (accessible from cluster)
   - Redis 7+ (accessible from cluster)
   - Apache Pulsar (accessible from cluster)
   - Connect Service (conn-dict) deployed

## Deployment Steps

### 1. Create Namespace
```bash
kubectl apply -f namespace.yaml
```

### 2. Create ConfigMap
```bash
kubectl apply -f configmap.yaml
```

### 3. Create Secrets

**Option A: Manual (NOT RECOMMENDED for production)**
```bash
# Copy example and edit with real values
cp secret.yaml.example secret.yaml
nano secret.yaml  # Edit with real credentials

# Apply (DO NOT COMMIT secret.yaml)
kubectl apply -f secret.yaml

# Delete local file
rm secret.yaml
```

**Option B: Using Vault (RECOMMENDED)**
```bash
# Assuming you have External Secrets Operator installed
kubectl apply -f secret-external.yaml
```

**Option C: Using kubectl create secret**
```bash
kubectl create secret generic core-dict-secrets \
  --namespace=dict-system \
  --from-literal=DB_HOST=postgres.production.internal \
  --from-literal=DB_USER=core_dict_user \
  --from-literal=DB_PASSWORD=<from-vault> \
  --from-literal=REDIS_HOST=redis.production.internal \
  --from-literal=REDIS_PASSWORD=<from-vault> \
  --from-literal=CONNECT_URL=conn-dict.dict-system.svc.cluster.local:9092 \
  --from-literal=PULSAR_URL=pulsar://pulsar.dict-system.svc.cluster.local:6650 \
  --from-literal=JWT_SECRET_KEY=<from-vault>
```

### 4. Deploy Application
```bash
# Apply all remaining manifests
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f hpa.yaml
kubectl apply -f pdb.yaml
kubectl apply -f networkpolicy.yaml
```

### 5. Verify Deployment
```bash
# Check pods
kubectl get pods -n dict-system -l app=core-dict

# Check service
kubectl get svc -n dict-system core-dict

# Check HPA
kubectl get hpa -n dict-system core-dict-hpa

# View logs
kubectl logs -n dict-system -l app=core-dict --tail=50 -f

# Check events
kubectl get events -n dict-system --sort-by='.lastTimestamp'
```

## Testing

### Port-Forward for Local Testing
```bash
# Forward gRPC port
kubectl port-forward -n dict-system svc/core-dict 9090:9090

# Test health check
grpcurl -plaintext localhost:9090 dict.core.v1.CoreDictService/HealthCheck
```

### Exec into Pod
```bash
# Get pod name
POD_NAME=$(kubectl get pods -n dict-system -l app=core-dict -o jsonpath='{.items[0].metadata.name}')

# Exec into pod
kubectl exec -it -n dict-system $POD_NAME -- sh

# Inside pod, check environment
env | grep -E "DB_|REDIS_|PULSAR_"
```

## Monitoring

### View Metrics
```bash
# Port-forward metrics port
kubectl port-forward -n dict-system svc/core-dict 9091:9091

# Curl metrics
curl http://localhost:9091/metrics
```

### Check Prometheus Scraping
```bash
# View service monitor (if using Prometheus Operator)
kubectl get servicemonitor -n dict-system core-dict
```

## Scaling

### Manual Scaling
```bash
# Scale to 5 replicas
kubectl scale deployment core-dict -n dict-system --replicas=5

# Check scaling
kubectl get pods -n dict-system -l app=core-dict
```

### Auto-Scaling (HPA)
```bash
# Check HPA status
kubectl get hpa -n dict-system core-dict-hpa

# Describe HPA
kubectl describe hpa -n dict-system core-dict-hpa

# Trigger load test to see auto-scaling in action
# (use k6 or similar)
```

## Updates & Rollbacks

### Rolling Update
```bash
# Update image version
kubectl set image deployment/core-dict \
  -n dict-system \
  core-dict=lbpay/core-dict:1.1.0

# Watch rollout
kubectl rollout status deployment/core-dict -n dict-system

# Check rollout history
kubectl rollout history deployment/core-dict -n dict-system
```

### Rollback
```bash
# Rollback to previous version
kubectl rollout undo deployment/core-dict -n dict-system

# Rollback to specific revision
kubectl rollout undo deployment/core-dict -n dict-system --to-revision=2
```

## Troubleshooting

### Pods Not Starting
```bash
# Check pod status
kubectl describe pod -n dict-system <pod-name>

# Common issues:
# 1. Image pull error → Check image tag and registry credentials
# 2. CrashLoopBackOff → Check logs: kubectl logs -n dict-system <pod-name>
# 3. Pending → Check resources: kubectl describe node
```

### Database Connection Issues
```bash
# Check if DB_HOST is resolvable from pod
kubectl exec -it -n dict-system <pod-name> -- nslookup postgres.production.internal

# Check if DB port is accessible
kubectl exec -it -n dict-system <pod-name> -- nc -zv postgres.production.internal 5432

# Verify DB credentials in secret
kubectl get secret core-dict-secrets -n dict-system -o jsonpath='{.data.DB_USER}' | base64 -d
```

### gRPC Health Probe Failures
```bash
# Check if gRPC server is listening
kubectl exec -it -n dict-system <pod-name> -- netstat -tuln | grep 9090

# Test health check manually
kubectl exec -it -n dict-system <pod-name> -- grpcurl -plaintext localhost:9090 dict.core.v1.CoreDictService/HealthCheck
```

### Network Policy Issues
```bash
# Check if network policy is blocking traffic
kubectl describe networkpolicy -n dict-system core-dict-netpol

# Test connectivity from another pod
kubectl run -it --rm debug --image=nicolaka/netshoot -n lb-connect -- sh
# Inside debug pod:
grpcurl -plaintext core-dict.dict-system.svc.cluster.local:9090 list
```

## Security

### RBAC
The deployment creates:
- ServiceAccount: `core-dict`
- Role: Read-only access to ConfigMaps and Secrets
- RoleBinding: Binds Role to ServiceAccount

### Network Policy
- **Ingress**: Only from `lb-connect` namespace and `monitoring` namespace
- **Egress**: Only to PostgreSQL, Redis, Pulsar, Connect, and DNS

### Pod Security
- Runs as non-root user (UID 1000)
- Read-only root filesystem (optional)
- No privilege escalation
- All capabilities dropped

## Clean Up

```bash
# Delete all resources
kubectl delete -f networkpolicy.yaml
kubectl delete -f pdb.yaml
kubectl delete -f hpa.yaml
kubectl delete -f service.yaml
kubectl delete -f deployment.yaml
kubectl delete -f configmap.yaml
kubectl delete -f secret.yaml  # If you created it manually
kubectl delete -f namespace.yaml  # ⚠️ This will delete everything in the namespace
```

## Production Checklist

- [ ] Secrets stored in Vault (not in Git)
- [ ] PostgreSQL migrations applied
- [ ] PostgreSQL user created with proper permissions
- [ ] Redis accessible from cluster
- [ ] Pulsar topics created
- [ ] Connect service deployed and accessible
- [ ] Network policies tested
- [ ] Resource limits tuned based on load tests
- [ ] HPA thresholds validated
- [ ] Prometheus alerts configured
- [ ] PagerDuty integration enabled
- [ ] Runbook created and shared with on-call team
- [ ] Backup and disaster recovery plan in place

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [gRPC Health Checking](https://github.com/grpc-ecosystem/grpc-health-probe)
- [Horizontal Pod Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)

---

**Last Updated**: 2025-10-27
**Version**: 1.0.0
