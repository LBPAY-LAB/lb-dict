# Core DICT - Quick Start Guide

Get up and running with Core DICT infrastructure in 5 minutes.

## Prerequisites

- Docker Desktop installed and running
- 4GB+ RAM available
- Ports 5434, 6380, 6651, 7235, 8083 available

## Step 1: Start Infrastructure (30 seconds)

```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/core-dict
docker-compose up -d
```

Expected output:
```
✓ Network core-dict-network   Created
✓ Container core-dict-postgres   Started
✓ Container core-dict-redis      Started
✓ Container core-dict-pulsar     Started
✓ Container core-dict-temporal   Started
```

## Step 2: Wait for Services (30 seconds)

```bash
# Wait for services to be healthy
sleep 30
docker-compose ps
```

Expected: 3-4 services showing "healthy" status.

## Step 3: Initialize Database (10 seconds)

```bash
./scripts/init-db.sh
```

Expected output:
```
[SUCCESS] PostgreSQL is ready!
[SUCCESS] Schemas created successfully!
[SUCCESS] Database verification completed!
```

## Step 4: Verify Everything Works

```bash
# Test PostgreSQL
docker exec core-dict-postgres psql -U postgres -c "SELECT 1;"

# Test Redis
docker exec core-dict-redis redis-cli ping

# Test Pulsar
curl http://localhost:8083/admin/v2/clusters
```

## You're Ready! 🎉

Your infrastructure is now running:

| Service | URL/Connection |
|---------|----------------|
| PostgreSQL | `localhost:5434` |
| Redis | `localhost:6380` |
| Pulsar Broker | `localhost:6651` |
| Pulsar Admin | `http://localhost:8083` |
| Temporal | `localhost:7235` |

## Common Commands

```bash
# View logs
docker-compose logs -f

# Stop everything
docker-compose down

# Restart a service
docker-compose restart postgres

# Clean everything (WARNING: deletes data!)
docker-compose down -v
```

## Connection Strings

```bash
# PostgreSQL
DATABASE_URL=postgres://postgres:postgres@localhost:5434/core_dict?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6380/0

# Pulsar
PULSAR_URL=pulsar://localhost:6651

# Temporal
TEMPORAL_HOST=localhost:7235
```

## Troubleshooting

### Ports already in use?

Edit `.env` file and change ports:
```bash
POSTGRES_PORT=5435
REDIS_PORT=6381
```

Then restart:
```bash
docker-compose down
docker-compose up -d
```

### Service not healthy?

Check logs:
```bash
docker-compose logs <service_name>
```

### Need to reset everything?

```bash
docker-compose down -v
docker-compose up -d
./scripts/init-db.sh
```

## Next Steps

1. ✅ Infrastructure running
2. 📚 Read [README-DOCKER.md](./README-DOCKER.md) for detailed docs
3. 🚀 Start developing your Core DICT application
4. 🧪 Run tests: `make test`
5. 📊 Add monitoring tools: `docker-compose --profile tools up -d`

## Help

- Full documentation: [README-DOCKER.md](./README-DOCKER.md)
- Infrastructure status: [INFRASTRUCTURE-STATUS.md](./INFRASTRUCTURE-STATUS.md)
- Main README: [README.md](./README.md)

---

**Total Setup Time**: ~1 minute
**Status**: Ready to develop! 🚀
