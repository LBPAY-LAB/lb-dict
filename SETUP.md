# DICT LBPay - Setup Guide

**Last Updated**: 2025-10-26
**Version**: 1.0.0
**Environment**: Local Development

---

## 📋 Prerequisites

### Required Tools
- **Docker** 24.0+ & **Docker Compose** 2.20+
- **Go** 1.24.5+
- **Java** 17+ (for XML Signer)
- **Maven** 3.9+ or **Gradle** 8.5+
- **Node.js** 18+ (optional, for frontend)
- **Make** (optional, for Makefiles)

### System Requirements
- **RAM**: Minimum 8GB (16GB recommended)
- **Disk**: 10GB free space
- **OS**: macOS, Linux, or Windows with WSL2

---

## 🚀 Quick Start (5 minutes)

### 1. Clone and Configure
```bash
# Already in the directory
cd /Users/jose.silva.lb/LBPay/IA_Dict

# Copy environment variables
cp .env.example .env

# Review and adjust .env if needed
vim .env
```

### 2. Start Infrastructure
```bash
# Start all infrastructure services
docker-compose up -d

# Wait for services to be healthy (~30 seconds)
docker-compose ps

# Check logs
docker-compose logs -f
```

### 3. Run Database Migrations
```bash
# conn-dict migrations
cd conn-dict
goose -dir migrations postgres "postgres://conn_dict_user:conn_dict_password_dev@localhost:5432/conn_dict?sslmode=disable" up

# conn-bridge migrations (when created)
cd ../conn-bridge
# goose -dir migrations postgres "postgres://conn_bridge_user:conn_bridge_password_dev@localhost:5432/conn_bridge?sslmode=disable" up

cd ..
```

### 4. Verify Infrastructure
```bash
# PostgreSQL
psql -h localhost -U dict_admin -d dict -c "SELECT 1;"

# Redis
redis-cli ping

# Temporal UI
open http://localhost:8088

# Grafana
open http://localhost:3000
# Login: admin / admin

# Prometheus
open http://localhost:9090

# Jaeger
open http://localhost:16686

# Pulsar
curl http://localhost:8080/admin/v2/clusters
```

---

## 🏗️ Service Ports

| Service | Port(s) | Description |
|---------|---------|-------------|
| **PostgreSQL** | 5432 | Main database |
| **Redis** | 6379 | Cache and sessions |
| **Temporal Server** | 7233 (gRPC), 7234 (HTTP) | Workflow engine |
| **Temporal UI** | 8088 | Workflow monitoring |
| **Pulsar** | 6650 (binary), 8080 (admin) | Message broker |
| **Vault** | 8200 | Secrets management |
| **Prometheus** | 9090 | Metrics collection |
| **Grafana** | 3000 | Dashboards |
| **Jaeger** | 16686 (UI), 14268 (collector) | Distributed tracing |
| **conn-dict** | 9092 (gRPC), 8081 (HTTP), 9093 (metrics) | RSFN Connect service |
| **conn-bridge** | 9094 (gRPC), 8082 (HTTP), 9095 (metrics) | RSFN Bridge service |
| **core-dict** | 8080 (REST), 9090 (gRPC), 9091 (metrics) | Core DICT service |

---

## 🗄️ Database Setup

### Create Databases (Already done by init.sql)
```sql
-- Databases created:
-- - dict (main)
-- - conn_dict
-- - conn_bridge
-- - core_dict
```

### Run Migrations
```bash
# Install goose if not installed
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run conn-dict migrations
cd conn-dict
goose -dir migrations postgres "postgres://conn_dict_user:conn_dict_password_dev@localhost:5432/conn_dict?sslmode=disable" up

# Check migration status
goose -dir migrations postgres "postgres://conn_dict_user:conn_dict_password_dev@localhost:5432/conn_dict?sslmode=disable" status

# Rollback last migration (if needed)
goose -dir migrations postgres "postgres://conn_dict_user:conn_dict_password_dev@localhost:5432/conn_dict?sslmode=disable" down
```

### Verify Tables
```bash
psql -h localhost -U conn_dict_user -d conn_dict -c "\dt"

# Expected tables:
# - claims
# - entries
# - infractions
# - audit_logs (partitioned)
# - event_logs (partitioned)
# - goose_db_version
```

---

## 🎯 Running Services

### conn-dict (RSFN Connect)
```bash
cd conn-dict

# Install dependencies
go mod download

# Run tests
go test ./... -v -short

# Run service
go run cmd/server/main.go
```

### conn-bridge (RSFN Bridge)
```bash
cd conn-bridge

# Install dependencies
go mod download

# Build XML Signer (Java)
cd xml-signer
mvn clean package
cd ..

# Run tests
go test ./... -v -short

# Run service
go run cmd/server/main.go
```

### core-dict (Core DICT)
```bash
cd core-dict

# Install dependencies
go mod download

# Run tests
go test ./... -v -short

# Run service
go run cmd/server/main.go
```

---

## 🧪 Testing

### Unit Tests
```bash
# Run all unit tests
go test ./... -v -short

# Run with coverage
go test ./... -v -short -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests
```bash
# Requires infrastructure running
docker-compose up -d

# Run integration tests
go test ./... -v

# Run specific package
go test ./internal/infrastructure/pulsar/... -v
```

### E2E Tests
```bash
# Run end-to-end tests (when implemented)
cd dict-e2e-tests
go test ./... -v
```

---

## 📊 Monitoring and Observability

### Access Dashboards

**Grafana** (http://localhost:3000)
- Login: admin / admin
- Dashboards: Will be auto-provisioned
- Data sources: Prometheus, Jaeger, PostgreSQL

**Prometheus** (http://localhost:9090)
- Metrics explorer
- Targets status: Status → Targets
- Query examples:
  ```promql
  rate(http_requests_total[5m])
  up{job="conn-dict"}
  temporal_workflow_count
  ```

**Jaeger** (http://localhost:16686)
- Distributed tracing
- Search traces by service
- Dependency graph

**Temporal UI** (http://localhost:8088)
- Workflow execution history
- Task queue monitoring
- Worker status

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f postgres
docker-compose logs -f temporal

# Application logs (when running)
tail -f conn-dict/logs/app.log
```

---

## 🔒 Security Setup

### Development mTLS (Self-signed)
```bash
# Generate dev certificates
./scripts/generate-dev-certs.sh

# Certificates will be in:
# - certs/ca.pem
# - certs/server-cert.pem
# - certs/server-key.pem
# - certs/client-cert.pem
# - certs/client-key.pem
```

### Vault Setup
```bash
# Login to Vault
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dict-root-token

# Store secrets
vault kv put secret/dict/database \
  username=conn_dict_user \
  password=conn_dict_password_dev

vault kv put secret/dict/bacen \
  ispb=12345678 \
  cert_path=/path/to/cert.pem

# Read secrets
vault kv get secret/dict/database
```

---

## 🐛 Troubleshooting

### Infrastructure Won't Start
```bash
# Check Docker resources
docker system df
docker system prune

# Check logs
docker-compose logs postgres
docker-compose logs temporal

# Restart services
docker-compose restart
```

### Database Connection Issues
```bash
# Test connection
psql -h localhost -U dict_admin -d dict

# Check if PostgreSQL is accepting connections
docker-compose exec postgres pg_isready

# View PostgreSQL logs
docker-compose logs postgres
```

### Port Conflicts
```bash
# Check what's using ports
lsof -i :5432
lsof -i :6379
lsof -i :7233

# Kill process (if needed)
kill -9 <PID>
```

### Temporal Workflow Issues
```bash
# Check Temporal health
docker-compose exec temporal tctl cluster health

# List workflows
tctl workflow list

# Describe specific workflow
tctl workflow describe -w <workflow_id>
```

### Pulsar Connection Issues
```bash
# Check Pulsar health
curl http://localhost:8080/admin/v2/brokers/health

# List topics
docker-compose exec pulsar bin/pulsar-admin topics list public/default

# View topic stats
docker-compose exec pulsar bin/pulsar-admin topics stats persistent://public/default/dict-claim-events
```

---

## 🧹 Cleanup

### Stop Services
```bash
# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: deletes all data)
docker-compose down -v
```

### Clean Build Artifacts
```bash
# Go build cache
go clean -cache -testcache -modcache

# Docker
docker system prune -a --volumes
```

---

## 📚 Additional Resources

- [Temporal Documentation](https://docs.temporal.io/)
- [Pulsar Documentation](https://pulsar.apache.org/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Prometheus Documentation](https://prometheus.io/docs/)

---

## 🆘 Support

For issues or questions:
1. Check logs: `docker-compose logs -f`
2. Review troubleshooting section above
3. Contact: Jose Luis Silva (jose.silva.lb@lbpay.com)

---

**Happy Coding!** 🚀