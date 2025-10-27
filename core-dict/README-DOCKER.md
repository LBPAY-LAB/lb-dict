# Core DICT - Docker Infrastructure Guide

Complete guide for running Core DICT infrastructure with Docker Compose.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Services](#services)
- [Usage Commands](#usage-commands)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)
- [Production Considerations](#production-considerations)

---

## Overview

The Core DICT infrastructure stack includes:

| Service | Purpose | Port(s) | Required |
|---------|---------|---------|----------|
| **PostgreSQL 16** | Primary database | 5432 | Yes |
| **Redis 7** | Cache layer | 6379 | Yes |
| **Apache Pulsar 3.2** | Event streaming | 6650, 8080 | Yes |
| **Temporal 1.22** | Workflow orchestration | 7233, 7234 | Yes |
| **Temporal UI** | Temporal web interface | 8088 | Optional (tools profile) |
| **PGAdmin** | Database management UI | 5050 | Optional (tools profile) |
| **Redis Commander** | Redis management UI | 8081 | Optional (tools profile) |
| **Prometheus** | Metrics collection | 9090 | Optional (monitoring profile) |
| **Grafana** | Metrics visualization | 3000 | Optional (monitoring profile) |

---

## Prerequisites

### Required

- **Docker**: version 20.10+ ([Install Docker](https://docs.docker.com/get-docker/))
- **Docker Compose**: version 2.0+ (included with Docker Desktop)

### Optional

- **Goose**: for database migrations ([Install Goose](https://github.com/pressly/goose))
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

### System Requirements

- **RAM**: Minimum 4GB available (8GB recommended)
- **Disk**: At least 10GB free space
- **CPU**: 2+ cores recommended

---

## Quick Start

### 1. Clone and Navigate

```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/core-dict
```

### 2. Configure Environment

```bash
# Use the provided .env file (already configured for local development)
# Or customize it:
cp .env .env.local
nano .env.local
```

### 3. Start Infrastructure

```bash
# Start core services (PostgreSQL, Redis, Pulsar, Temporal)
docker-compose up -d

# Or start with all tools
docker-compose --profile tools up -d

# Or start everything (including monitoring)
docker-compose --profile tools --profile monitoring up -d
```

### 4. Initialize Database

```bash
# Wait for PostgreSQL to be ready (30-60 seconds)
docker-compose logs -f postgres

# Initialize database and run migrations
./scripts/init-db.sh
```

### 5. Verify Services

```bash
# Check all services are healthy
docker-compose ps

# Expected output: All services should show "healthy" status
```

---

## Services

### PostgreSQL 16

**Purpose**: Primary relational database for Core DICT

**Configuration**:
- Database: `core_dict`
- User: `postgres`
- Password: `postgres` (change in production!)
- Port: `5432`

**Connection String**:
```
postgres://postgres:postgres@localhost:5432/core_dict?sslmode=disable
```

**Connect with psql**:
```bash
docker exec -it core-dict-postgres psql -U postgres -d core_dict
```

**Features**:
- UTF-8 encoding
- Extensions: uuid-ossp, pg_trgm, pgcrypto, pg_stat_statements
- Performance tuning for development
- Persistent volumes
- Health checks

---

### Redis 7

**Purpose**: Cache layer for entries, accounts, and session data

**Configuration**:
- Port: `6379`
- No password (for local dev)
- Max memory: 512MB (LRU eviction)

**Connection**:
```bash
docker exec -it core-dict-redis redis-cli
```

**Features**:
- AOF persistence (every second)
- RDB snapshots (configurable intervals)
- Persistent volumes
- Health checks

**Common Commands**:
```bash
# Ping
redis-cli ping

# Check keys
redis-cli KEYS '*'

# Get cache stats
redis-cli INFO stats

# Monitor operations
redis-cli MONITOR
```

---

### Apache Pulsar 3.2

**Purpose**: Event streaming platform for async communication

**Configuration**:
- Broker Port: `6650`
- Admin API: `8080`
- Mode: Standalone (for development)

**Admin UI**: http://localhost:8080

**Topics** (auto-created):
- `persistent://dict/events/key-events`
- `persistent://dict/events/claim-events`
- `persistent://lb-conn/dict/rsfn-dict-req-out`
- `persistent://lb-conn/dict/rsfn-dict-res-in`

**Common Commands**:
```bash
# Access Pulsar admin CLI
docker exec -it core-dict-pulsar bash

# List topics
bin/pulsar-admin topics list dict/events

# Create topic
bin/pulsar-admin topics create persistent://dict/events/test

# Publish test message
bin/pulsar-client produce persistent://dict/events/test --messages "Hello World"

# Consume messages
bin/pulsar-client consume persistent://dict/events/test -s "test-sub" -n 0
```

**Health Check**:
```bash
curl http://localhost:8080/admin/v2/clusters
```

---

### Temporal Server 1.22

**Purpose**: Durable workflow orchestration for long-running processes

**Configuration**:
- gRPC Port: `7233`
- Database: PostgreSQL (shared with Core DICT)
- Namespace: `default`

**Features**:
- Automatic schema setup
- PostgreSQL backend
- Persistent workflows
- Durable timers

**Connect with tctl**:
```bash
# Install tctl
brew install tctl

# List workflows
tctl workflow list

# Describe workflow
tctl workflow describe -w <workflow_id>

# Query workflow
tctl workflow query -w <workflow_id> -qt <query_type>
```

---

### Temporal UI (Optional)

**Purpose**: Web interface for Temporal workflows

**Configuration**:
- Port: `8088`
- Profile: `tools`

**Access**: http://localhost:8088

**Start**:
```bash
docker-compose --profile tools up -d
```

**Features**:
- View workflows and activities
- Query workflow state
- Retry failed workflows
- View workflow history
- Search and filter

---

### PGAdmin (Optional)

**Purpose**: Web-based PostgreSQL management

**Configuration**:
- Port: `5050`
- Email: `admin@lbpay.local`
- Password: `admin`
- Profile: `tools`

**Access**: http://localhost:5050

**First Time Setup**:
1. Open http://localhost:5050
2. Login with credentials above
3. Add server:
   - Name: Core DICT
   - Host: `postgres` (container name)
   - Port: `5432`
   - Username: `postgres`
   - Password: `postgres`

---

### Redis Commander (Optional)

**Purpose**: Web-based Redis management

**Configuration**:
- Port: `8081`
- Profile: `tools`

**Access**: http://localhost:8081

---

### Prometheus (Optional)

**Purpose**: Metrics collection and storage

**Configuration**:
- Port: `9090`
- Profile: `monitoring`

**Access**: http://localhost:9090

**Requires**: Create `monitoring/prometheus.yml` (see configuration section)

---

### Grafana (Optional)

**Purpose**: Metrics visualization and dashboards

**Configuration**:
- Port: `3000`
- Username: `admin`
- Password: `admin`
- Profile: `monitoring`

**Access**: http://localhost:3000

**Requires**: Configure datasources and dashboards in `monitoring/grafana/`

---

## Usage Commands

### Start Services

```bash
# Core services only (PostgreSQL, Redis, Pulsar, Temporal)
docker-compose up -d

# With management tools (PGAdmin, Redis Commander, Temporal UI)
docker-compose --profile tools up -d

# With monitoring (Prometheus, Grafana)
docker-compose --profile monitoring up -d

# Everything
docker-compose --profile tools --profile monitoring up -d

# Start in foreground (see logs)
docker-compose up
```

### Stop Services

```bash
# Stop all services (preserve volumes)
docker-compose down

# Stop and remove volumes (WARNING: deletes data!)
docker-compose down -v

# Stop specific service
docker-compose stop postgres
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f postgres
docker-compose logs -f temporal
docker-compose logs -f pulsar

# Last 100 lines
docker-compose logs --tail=100 postgres

# Since timestamp
docker-compose logs --since 2024-01-01T10:00:00 postgres
```

### Check Status

```bash
# All services
docker-compose ps

# Health status
docker-compose ps --format json | jq '.[].Health'

# Detailed inspect
docker inspect core-dict-postgres
```

### Restart Services

```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart postgres
docker-compose restart temporal
```

### Execute Commands

```bash
# PostgreSQL
docker exec -it core-dict-postgres psql -U postgres -d core_dict

# Redis
docker exec -it core-dict-redis redis-cli

# Pulsar
docker exec -it core-dict-pulsar bash

# Temporal
docker exec -it core-dict-temporal tctl workflow list
```

### Resource Usage

```bash
# Check CPU/Memory usage
docker stats

# Check specific service
docker stats core-dict-postgres
```

---

## Configuration

### Environment Variables

All configuration is managed via `.env` file. Key variables:

```bash
# Database
POSTGRES_DB=core_dict
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_PORT=5432

# Redis
REDIS_PORT=6379

# Pulsar
PULSAR_PORT=6650
PULSAR_HTTP_PORT=8080

# Temporal
TEMPORAL_PORT=7233
TEMPORAL_WEB_PORT=8088
```

### Custom Configuration

For production, create `.env.production`:

```bash
cp .env .env.production
nano .env.production
```

Load custom environment:

```bash
docker-compose --env-file .env.production up -d
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose logs <service_name>

# Check disk space
df -h

# Check memory
free -m

# Check ports
netstat -tuln | grep <port>
```

### Database Connection Refused

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Verify health
docker exec core-dict-postgres pg_isready -U postgres

# Test connection
docker exec core-dict-postgres psql -U postgres -d core_dict -c "SELECT 1;"
```

### Temporal Not Starting

```bash
# Temporal requires PostgreSQL to be healthy first
# Check PostgreSQL is ready
docker-compose ps postgres

# Check Temporal logs
docker-compose logs temporal

# Recreate Temporal
docker-compose stop temporal
docker-compose rm -f temporal
docker-compose up -d temporal
```

### Pulsar Not Healthy

```bash
# Pulsar takes 30-60 seconds to become healthy
# Check logs
docker-compose logs pulsar

# Test admin API
curl http://localhost:8080/admin/v2/clusters

# Restart if needed
docker-compose restart pulsar
```

### Port Conflicts

```bash
# Check what's using the port
lsof -i :5432
lsof -i :6379

# Change ports in .env
POSTGRES_PORT=5433
REDIS_PORT=6380
```

### Volume Issues

```bash
# List volumes
docker volume ls | grep core-dict

# Remove specific volume
docker volume rm core-dict-postgres-data

# Remove all volumes (WARNING: deletes data!)
docker-compose down -v
```

### Out of Memory

```bash
# Check memory usage
docker stats

# Reduce service memory limits in docker-compose.yml
# Or increase Docker Desktop memory allocation
```

### Slow Performance

```bash
# Check resource limits
docker stats

# Increase Docker Desktop resources:
# - CPU: 4+ cores
# - RAM: 8GB+
# - Disk: 20GB+

# Check disk I/O
docker exec core-dict-postgres iostat
```

---

## Production Considerations

### Security

**DO NOT use default passwords in production!**

```bash
# Generate secure passwords
openssl rand -base64 32

# Update .env.production
POSTGRES_PASSWORD=<secure_password>
PGADMIN_PASSWORD=<secure_password>
GRAFANA_PASSWORD=<secure_password>
```

### Networking

- Use custom networks with firewall rules
- Enable SSL/TLS for all connections
- Use VPN or private networks
- Restrict external access

### Volumes

- Use named volumes with backup strategy
- Consider external storage (NFS, EBS, etc.)
- Implement automated backups
- Test restore procedures

### Monitoring

- Enable Prometheus + Grafana
- Set up alerting rules
- Monitor disk space, memory, CPU
- Track database query performance

### High Availability

For production, use:
- PostgreSQL: Patroni cluster or RDS
- Redis: Redis Sentinel or ElastiCache
- Pulsar: Multi-broker cluster
- Temporal: Multi-node cluster

### Backup Strategy

```bash
# PostgreSQL backup
docker exec core-dict-postgres pg_dump -U postgres core_dict > backup.sql

# Restore
docker exec -i core-dict-postgres psql -U postgres core_dict < backup.sql

# Redis backup
docker exec core-dict-redis redis-cli SAVE
docker cp core-dict-redis:/data/dump.rdb ./redis-backup.rdb
```

---

## Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)
- [Apache Pulsar Documentation](https://pulsar.apache.org/docs/)
- [Temporal Documentation](https://docs.temporal.io/)

---

## Quick Reference

### Service URLs

```
PostgreSQL:         localhost:5432
Redis:              localhost:6379
Pulsar Broker:      localhost:6650
Pulsar Admin:       http://localhost:8080
Temporal gRPC:      localhost:7233
Temporal UI:        http://localhost:8088
PGAdmin:            http://localhost:5050
Redis Commander:    http://localhost:8081
Prometheus:         http://localhost:9090
Grafana:            http://localhost:3000
```

### Default Credentials

```
PostgreSQL:         postgres / postgres
PGAdmin:            admin@lbpay.local / admin
Grafana:            admin / admin
Redis:              (no password)
```

### Common Workflows

```bash
# Fresh start
docker-compose down -v
docker-compose up -d
./scripts/init-db.sh

# Update services
docker-compose pull
docker-compose up -d

# View all logs
docker-compose logs -f

# Check health
docker-compose ps
```

---

**Last Updated**: 2025-10-27
**Version**: 1.0
