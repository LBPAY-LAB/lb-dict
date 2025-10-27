# Temporal Setup Guide

This guide explains how to set up and use Temporal in the DICT LBPay project.

## Overview

Temporal is a durable workflow orchestration platform that provides reliability, scalability, and visibility for business-critical applications. This project uses Temporal for managing DICT workflows.

## Architecture

The Temporal setup includes:

- **Temporal Server** (v1.36.0): Core workflow engine
- **Temporal UI** (v2.35.1): Web-based UI for monitoring workflows
- **PostgreSQL 16**: Persistence layer for Temporal state
- **Elasticsearch 7.17**: Advanced visibility and search capabilities

## Services

### Temporal Server
- **Port**: 7233
- **Image**: `temporalio/auto-setup:1.36.0`
- **Container**: `conn-dict-temporal`

### Temporal UI
- **Port**: 8088
- **URL**: http://localhost:8088
- **Image**: `temporalio/ui:2.35.1`
- **Container**: `conn-dict-temporal-ui`

### PostgreSQL for Temporal
- **Port**: 5433 (external) / 5432 (internal)
- **Image**: `postgres:16-alpine`
- **Container**: `conn-dict-temporal-postgres`
- **Database**: `temporal`
- **User**: `temporal`
- **Password**: `temporal`

### Elasticsearch
- **Port**: 9200
- **Image**: `elasticsearch:7.17.10`
- **Container**: `conn-dict-temporal-elasticsearch`

## Quick Start

### 1. Start Temporal Services

```bash
# Start all Temporal services
docker-compose up temporal temporal-ui temporal-postgres temporal-elasticsearch -d

# Or start in foreground to see logs
docker-compose up temporal temporal-ui temporal-postgres temporal-elasticsearch
```

### 2. Verify Services

```bash
# Check running containers
docker ps | grep temporal

# Check Temporal Server health
curl http://localhost:7233/

# Access Temporal UI
open http://localhost:8088
```

### 3. Run Example Workflow

```bash
# Build and run the example
go run cmd/temporal-example/main.go
```

## Configuration

### Environment Variables

Configure Temporal connection in your `.env` file:

```bash
# Temporal Configuration
TEMPORAL_HOST=localhost
TEMPORAL_PORT=7233
TEMPORAL_NAMESPACE=default
TEMPORAL_TASK_QUEUE=conn-dict-task-queue
```

### Programmatic Configuration

```go
import "github.com/lbpay-lab/conn-dict/internal/infrastructure/temporal"

// Load default configuration
config := temporal.DefaultConfig()

// Or create custom configuration
config := temporal.Config{
    Host:      "localhost",
    Port:      7233,
    Namespace: "default",
    TaskQueue: "my-task-queue",
}
```

## Using the Temporal Client

### Create a Client

```go
import (
    "github.com/lbpay-lab/conn-dict/internal/infrastructure/temporal"
    "github.com/sirupsen/logrus"
)

logger := logrus.New()

client, err := temporal.NewClient(temporal.ClientConfig{
    HostPort:  "localhost:7233",
    Namespace: "default",
    Logger:    logger,
})
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Perform health check
err = client.HealthCheck(context.Background())
if err != nil {
    log.Fatal("Health check failed:", err)
}
```

### Create a Worker

```go
worker, err := temporal.NewWorker(client, temporal.WorkerConfig{
    TaskQueue:               "my-task-queue",
    MaxConcurrentWorkflows:  100,
    MaxConcurrentActivities: 200,
    Logger:                  logger,
})
if err != nil {
    log.Fatal(err)
}

// Register workflows and activities
worker.RegisterWorkflow(MyWorkflow)
worker.RegisterActivity(MyActivity)

// Start the worker
err = worker.Start()
if err != nil {
    log.Fatal(err)
}
defer worker.Stop()
```

### Execute a Workflow

```go
import (
    "go.temporal.io/sdk/client"
)

workflowOptions := client.StartWorkflowOptions{
    ID:        "my-workflow-id",
    TaskQueue: "my-task-queue",
    WorkflowExecutionTimeout: 5 * time.Minute,
}

run, err := client.ExecuteWorkflow(
    context.Background(),
    workflowOptions,
    MyWorkflow,
    workflowInput,
)
if err != nil {
    log.Fatal(err)
}

// Get workflow result
var result MyWorkflowResult
err = run.Get(context.Background(), &result)
```

## Workflow Development

### Creating a Workflow

```go
package workflows

import (
    "go.temporal.io/sdk/workflow"
)

type MyWorkflowInput struct {
    Data string
}

type MyWorkflowResult struct {
    Status  string
    Message string
}

func MyWorkflow(ctx workflow.Context, input MyWorkflowInput) (*MyWorkflowResult, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("Workflow started", "input", input.Data)

    // Your workflow logic here

    return &MyWorkflowResult{
        Status:  "completed",
        Message: "Success",
    }, nil
}
```

### Creating an Activity

```go
package activities

import (
    "context"
    "go.temporal.io/sdk/activity"
)

type MyActivityInput struct {
    Data string
}

type MyActivityResult struct {
    Result string
}

func MyActivity(ctx context.Context, input MyActivityInput) (*MyActivityResult, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Activity started", "input", input.Data)

    // Your activity logic here

    return &MyActivityResult{
        Result: "processed",
    }, nil
}
```

## Monitoring and Debugging

### Temporal UI

Access the Temporal UI at http://localhost:8088 to:
- View workflow executions
- Monitor workflow history
- Debug failed workflows
- Search workflows
- View workflow details and stack traces

### Docker Logs

```bash
# View Temporal Server logs
docker logs -f conn-dict-temporal

# View Temporal UI logs
docker logs -f conn-dict-temporal-ui

# View PostgreSQL logs
docker logs -f conn-dict-temporal-postgres
```

### Health Checks

```bash
# Check Temporal Server health
curl http://localhost:7233/

# Check Elasticsearch health
curl http://localhost:9200/_cluster/health

# Check PostgreSQL health
docker exec conn-dict-temporal-postgres pg_isready -U temporal
```

## Namespaces

Temporal uses namespaces to isolate workflows. The default namespace is `default`.

### Creating a Custom Namespace

```bash
# Using temporal CLI (if installed)
temporal operator namespace create my-namespace

# Or using docker exec
docker exec conn-dict-temporal tctl --ns my-namespace namespace register
```

## Troubleshooting

### Temporal Server Not Starting

1. Check if PostgreSQL is running:
   ```bash
   docker ps | grep temporal-postgres
   ```

2. Check PostgreSQL logs:
   ```bash
   docker logs conn-dict-temporal-postgres
   ```

3. Verify PostgreSQL health:
   ```bash
   docker exec conn-dict-temporal-postgres pg_isready -U temporal
   ```

### Cannot Connect to Temporal

1. Verify Temporal is running:
   ```bash
   docker ps | grep temporal
   ```

2. Check if port 7233 is accessible:
   ```bash
   nc -zv localhost 7233
   ```

3. Review Temporal logs:
   ```bash
   docker logs conn-dict-temporal
   ```

### Temporal UI Not Accessible

1. Verify UI is running:
   ```bash
   docker ps | grep temporal-ui
   ```

2. Check if port 8088 is accessible:
   ```bash
   nc -zv localhost 8088
   ```

3. Review UI logs:
   ```bash
   docker logs conn-dict-temporal-ui
   ```

### Workflow Execution Issues

1. Check worker is running and registered:
   - Verify worker logs
   - Ensure workflows/activities are registered
   - Check task queue name matches

2. Review workflow execution in Temporal UI:
   - Check workflow history
   - Review error messages
   - Examine stack traces

3. Verify namespace:
   - Ensure client and server use same namespace
   - Default namespace is `default`

## Performance Tuning

### Worker Configuration

Adjust worker concurrency based on your needs:

```go
worker, err := temporal.NewWorker(client, temporal.WorkerConfig{
    TaskQueue:               "my-task-queue",
    MaxConcurrentWorkflows:  100,  // Increase for more concurrent workflows
    MaxConcurrentActivities: 200,  // Increase for more concurrent activities
    Logger:                  logger,
})
```

### PostgreSQL Tuning

For production, adjust PostgreSQL configuration in docker-compose.yml:

```yaml
temporal-postgres:
  environment:
    POSTGRES_SHARED_BUFFERS: 256MB
    POSTGRES_EFFECTIVE_CACHE_SIZE: 1GB
    POSTGRES_MAX_CONNECTIONS: 100
```

### Elasticsearch Memory

Adjust Elasticsearch memory based on visibility needs:

```yaml
temporal-elasticsearch:
  environment:
    - ES_JAVA_OPTS=-Xms512m -Xmx512m  # Increase for better performance
```

## Best Practices

1. **Use Unique Workflow IDs**: Ensure each workflow execution has a unique ID
2. **Set Timeouts**: Always set workflow and activity timeouts
3. **Handle Errors**: Implement proper error handling and retries
4. **Use Activities for Side Effects**: Keep workflows deterministic
5. **Monitor Workflows**: Regularly check Temporal UI for failed workflows
6. **Version Workflows**: Use versioning when updating workflow code
7. **Test Workflows**: Write unit tests for workflows and activities
8. **Log Appropriately**: Use structured logging for better debugging

## Additional Resources

- [Temporal Documentation](https://docs.temporal.io/)
- [Temporal Go SDK](https://github.com/temporalio/sdk-go)
- [Temporal Samples](https://github.com/temporalio/samples-go)
- [Temporal Best Practices](https://docs.temporal.io/dev-guide/go/best-practices)

## Support

For issues or questions:
1. Check the Temporal UI at http://localhost:8088
2. Review Docker logs
3. Consult Temporal documentation
4. Contact the development team
