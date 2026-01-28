# TaskForge

[![Go Reference](https://pkg.go.dev/badge/github.com/agincgit/taskforge.svg)](https://pkg.go.dev/github.com/agincgit/taskforge)
[![Go Report Card](https://goreportcard.com/badge/github.com/agincgit/taskforge)](https://goreportcard.com/report/github.com/agincgit/taskforge)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A lightweight task orchestration and queue framework for Go applications.

## Features

- **Task Queue Management** — Enqueue, reserve, complete, cancel, and retry tasks with built-in state machine
- **Template System** — Define reusable task templates with default inputs and scheduling
- **Cron Scheduling** — Recurring task execution via cron expressions
- **Worker Registration** — Track workers with heartbeat monitoring
- **PostgreSQL Storage** — Production-ready persistence with GORM

## Installation

```bash
go get github.com/agincgit/taskforge
```

Requires Go 1.21+

## Quick Start

```go
package main

import (
    "context"

    "github.com/agincgit/taskforge/pkg/taskforge"
    "github.com/agincgit/taskforge/pkg/model"
    "github.com/agincgit/taskforge/internal/persistence"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    // Connect to database
    db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

    // Run migrations
    persistence.Migrate(db)

    // Create manager
    mgr, _ := taskforge.NewManager(taskforge.Config{
        DB:      db,
        Context: context.Background(),
    })

    // Enqueue a task
    task := &model.Task{
        Type:    "send_email",
        Payload: `{"to": "user@example.com"}`,
    }
    mgr.Enqueue(context.Background(), task)

    // Reserve and process
    reserved, _ := mgr.Reserve(context.Background())
    // ... process task ...
    mgr.Complete(context.Background(), reserved.ID, true)
}
```

## Architecture

```
taskforge/
├── cmd/taskforge/       # Default server implementation
├── pkg/
│   ├── taskforge/       # Public API (Manager, Status, Config)
│   ├── model/           # Domain models (Task, Template, Worker)
│   └── scheduler/       # Cron-based task scheduling
└── internal/            # HTTP handlers, persistence, config
```

### Package Overview

| Package | Import Path | Description |
|---------|-------------|-------------|
| `taskforge` | `github.com/agincgit/taskforge/pkg/taskforge` | Core Manager API and configuration |
| `model` | `github.com/agincgit/taskforge/pkg/model` | Task, Template, and Worker models |
| `scheduler` | `github.com/agincgit/taskforge/pkg/scheduler` | Cron-based recurring task scheduler |

## Task Lifecycle

```
Pending → InProgress → Succeeded
       ↘            ↘ Failed → (retry) → Pending
         → PendingCancellation → Cancelled
```

## API Endpoints

The default server exposes a REST API under `/taskforge/api/v1`:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/tasks` | Create task |
| `GET` | `/tasks` | List tasks |
| `GET` | `/tasks/:id` | Get task |
| `PUT` | `/tasks/:id` | Update task |
| `DELETE` | `/tasks/:id` | Delete task |
| `POST` | `/tasktemplate` | Create template |
| `GET` | `/tasktemplate` | List templates |
| `POST` | `/workers` | Register worker |
| `PUT` | `/workers/:id/heartbeat` | Worker heartbeat |
| `POST` | `/workerqueue` | Enqueue job |

## Configuration

Configuration is loaded from environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `TASKFORGE_DB_HOST` | PostgreSQL host | `localhost` |
| `TASKFORGE_DB_PORT` | PostgreSQL port | `5432` |
| `TASKFORGE_DB_USER` | Database user | `postgres` |
| `TASKFORGE_DB_PASSWORD` | Database password | **required** |
| `TASKFORGE_DB_NAME` | Database name | `taskforge_db` |
| `TASKFORGE_DB_SSLMODE` | SSL mode | `disable` |
| `TASKFORGE_PORT` | HTTP server port | `8080` |
| `TASKFORGE_LOG_LEVEL` | Log level | `info` |
| `TASKFORGE_HOSTNAME` | Worker hostname | auto-detected |

### Example

```bash
export TASKFORGE_DB_PASSWORD=secret
export TASKFORGE_DB_HOST=postgres.example.com
./taskforge
```

### Docker

```bash
docker run -e TASKFORGE_DB_PASSWORD=secret \
           -e TASKFORGE_DB_HOST=postgres \
           -p 8080:8080 \
           taskforge
```

## License

MIT License — see [LICENSE](LICENSE) for details.
