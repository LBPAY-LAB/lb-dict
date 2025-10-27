# Core DICT - Troubleshooting Guide

Quick reference for common issues and solutions.

---

## 🔥 Common Errors

### 1. "Not Implemented" Error in Real Mode

**Symptom**:
```
ERROR: Real mode not yet implemented. Use CORE_DICT_USE_MOCK_MODE=true
```

**Cause**: Handlers are set to `nil` due to interface incompatibilities

**Solution**:
```bash
# Option 1: Use Mock Mode (temporary)
export CORE_DICT_USE_MOCK_MODE=true
go run cmd/grpc/*.go

# Option 2: Fix interfaces (permanent)
# See REAL_MODE_STATUS.md section "Soluções Propostas"
```

**Status**: ⚠️ Known Issue - Interface refactoring needed

---

### 2. "Failed to connect to PostgreSQL"

**Symptom**:
```
ERROR: failed to connect to PostgreSQL: failed to ping database
```

**Cause**: PostgreSQL not running or wrong credentials

**Solution**:
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# If not running, start it
docker-compose up -d postgres

# Test connection
psql -h localhost -p 5432 -U dict_app -d lbpay_core_dict

# Check credentials in .env
cat .env | grep DB_
```

**Environment Variables**:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_NAME=lbpay_core_dict
DB_USER=dict_app
DB_PASSWORD=dict_password
```

---

### 3. "Failed to connect to Redis"

**Symptom**:
```
ERROR: failed to connect to Redis: dial tcp [::1]:6379: connect: connection refused
```

**Cause**: Redis not running

**Solution**:
```bash
# Check if Redis is running
docker ps | grep redis

# If not running, start it
docker-compose up -d redis

# Test connection
redis-cli -h localhost -p 6379 ping
# Expected output: PONG

# Check credentials in .env
cat .env | grep REDIS_
```

**Environment Variables**:
```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

---

### 4. Port Already in Use

**Symptom**:
```
ERROR: Failed to listen: listen tcp :9090: bind: address already in use
```

**Cause**: Another process using port 9090

**Solution**:
```bash
# Find process using port 9090
lsof -i :9090

# Kill process (if safe)
kill -9 <PID>

# Or use different port
export GRPC_PORT=9091
go run cmd/grpc/*.go
```

---

### 5. Compilation Errors

**Symptom**:
```
cmd/grpc/real_handler_init.go:123:45: cannot use entryRepo...
```

**Cause**: Interface incompatibility or missing dependencies

**Solution**:
```bash
# Clean and rebuild
go clean
go mod tidy
go mod download
go build -o /tmp/test ./cmd/grpc/

# If still fails, check go.mod
cat go.mod

# Ensure dict-contracts is available
ls -la ../dict-contracts/
```

---

### 6. "Handlers will return 'Not Implemented' errors"

**Symptom**:
```
WARN: Command handlers set to nil due to interface incompatibilities
```

**Cause**: This is expected - handlers are intentionally disabled

**Solution**: This is a known limitation. See `REAL_MODE_STATUS.md` for fix plan.

**Workaround**:
```bash
# Use Mock Mode for testing
export CORE_DICT_USE_MOCK_MODE=true
```

---

## 🔍 Diagnostic Commands

### Check Server Status

```bash
# Is server running?
ps aux | grep core-dict

# Check listening ports
lsof -i :9090

# Test health endpoint
grpcurl -plaintext localhost:9090 dict.core.v1.CoreDictService/HealthCheck
```

### Check Infrastructure

```bash
# PostgreSQL
docker exec -it postgres psql -U dict_app -d lbpay_core_dict -c "SELECT version();"

# Redis
docker exec -it redis redis-cli ping

# All containers
docker-compose ps
```

### Check Logs

```bash
# Server logs (if running with systemd)
journalctl -u core-dict -f

# Docker logs
docker-compose logs -f core-dict

# Application logs (if using file logging)
tail -f /var/log/core-dict/app.log
```

### Check Configuration

```bash
# Show all env vars
env | grep -E "(DB_|REDIS_|GRPC_|CORE_DICT_)"

# Validate .env file
cat .env

# Test config loading
grep "loadConfig" cmd/grpc/real_handler_init.go
```

---

## 🧪 Testing Commands

### Quick Health Check

```bash
# Mock Mode (should always work)
export CORE_DICT_USE_MOCK_MODE=true
go run cmd/grpc/*.go &
sleep 2
grpcurl -plaintext localhost:9090 dict.core.v1.CoreDictService/HealthCheck
kill %1
```

### Run Test Suite

```bash
# Automated tests
./test_real_mode_init.sh

# Manual compilation test
go build -o /tmp/test ./cmd/grpc/
/tmp/test &
sleep 2
kill %1
```

### Test with Infrastructure

```bash
# Start all services
docker-compose up -d

# Wait for services to be ready
sleep 5

# Check connectivity
docker exec postgres pg_isready
docker exec redis redis-cli ping

# Run server
export CORE_DICT_USE_MOCK_MODE=false
go run cmd/grpc/*.go
```

---

## 🐛 Debug Mode

### Enable Debug Logging

```bash
# Set log level to debug
export LOG_LEVEL=debug

# Run server
go run cmd/grpc/*.go

# You should see detailed logs:
# - Connection attempts
# - SQL queries (if available)
# - Cache operations
# - gRPC requests/responses
```

### Enable Go Race Detector

```bash
# Build with race detector
go build -race -o /tmp/core-dict-race ./cmd/grpc/

# Run
/tmp/core-dict-race

# This will detect data races at runtime
```

### Profile Performance

```bash
# CPU profiling
go build -o /tmp/core-dict ./cmd/grpc/
/tmp/core-dict &
PID=$!

# After some time
kill -USR1 $PID  # Generate profile
go tool pprof /tmp/cpu.prof
```

---

## 📋 Checklist: "Server Won't Start"

- [ ] Is Docker running? (`docker ps`)
- [ ] Is PostgreSQL running? (`docker ps | grep postgres`)
- [ ] Is Redis running? (`docker ps | grep redis`)
- [ ] Are ports available? (`lsof -i :9090`)
- [ ] Is .env file present? (`ls -la .env`)
- [ ] Are credentials correct? (`cat .env`)
- [ ] Does code compile? (`go build ./cmd/grpc/`)
- [ ] Are dependencies up to date? (`go mod tidy`)
- [ ] Is LOG_LEVEL set? (`echo $LOG_LEVEL`)

---

## 📋 Checklist: "RPCs Return Errors"

- [ ] Is Mock Mode enabled? (`echo $CORE_DICT_USE_MOCK_MODE`)
- [ ] Are handlers initialized? (Check server logs for "handlers set to nil")
- [ ] Is database schema created? (Run migrations)
- [ ] Are tables created? (`psql -h localhost -U dict_app -d lbpay_core_dict -c "\dt"`)
- [ ] Is cache working? (`redis-cli ping`)
- [ ] Are there interface errors? (Check compilation warnings)

---

## 🆘 Emergency Recovery

### Complete Reset

```bash
# 1. Stop everything
docker-compose down -v
killall core-dict 2>/dev/null

# 2. Clean build artifacts
go clean -cache -modcache -testcache
rm -f /tmp/core-dict*

# 3. Restart infrastructure
docker-compose up -d

# 4. Wait for services
sleep 10

# 5. Rebuild
go mod tidy
go build -o /tmp/core-dict ./cmd/grpc/

# 6. Test with Mock Mode
export CORE_DICT_USE_MOCK_MODE=true
/tmp/core-dict &
sleep 2
grpcurl -plaintext localhost:9090 dict.core.v1.CoreDictService/HealthCheck
kill %1
```

### Reset Database

```bash
# Drop and recreate
docker-compose down -v
docker-compose up -d postgres
sleep 5

# Run migrations
make migrate
# or
goose -dir migrations postgres "host=localhost port=5432 user=dict_app password=dict_password dbname=lbpay_core_dict sslmode=disable" up
```

---

## 📞 Getting Help

1. **Check Documentation**:
   - `REAL_MODE_STATUS.md` - Implementation status
   - `IMPLEMENTATION_SUMMARY.md` - High-level overview
   - This file (`TROUBLESHOOTING.md`) - Common issues

2. **Run Diagnostics**:
   ```bash
   ./test_real_mode_init.sh
   ```

3. **Check Logs**:
   - Server logs (stdout/stderr)
   - PostgreSQL logs (`docker logs postgres`)
   - Redis logs (`docker logs redis`)

4. **Test Components Individually**:
   - PostgreSQL: `psql -h localhost -U dict_app -d lbpay_core_dict`
   - Redis: `redis-cli -h localhost ping`
   - gRPC: `grpcurl -plaintext localhost:9090 list`

5. **Create Issue**:
   - Include error message
   - Include relevant logs
   - Include environment (OS, Go version, Docker version)
   - Include output of `./test_real_mode_init.sh`

---

## 🔗 Useful Links

- Go pgx docs: https://pkg.go.dev/github.com/jackc/pgx/v5
- Go Redis docs: https://redis.uptrace.dev/
- gRPC Go docs: https://grpc.io/docs/languages/go/
- grpcurl: https://github.com/fullstorydev/grpcurl

---

**Last Updated**: 2025-10-27
**Version**: 1.0
