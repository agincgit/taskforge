// Package taskforge provides a lightweight task orchestration and queue framework for Go.
//
// TaskForge enables applications to manage asynchronous task processing with support for
// enqueueing, scheduling, retries, and worker coordination. It is designed to be embedded
// as a library or run as a standalone server.
//
// # Core Concepts
//
//   - Task: A unit of work with type, payload, and lifecycle status
//   - Template: Reusable task definition with default inputs and optional cron schedule
//   - Manager: Central API for task operations (enqueue, reserve, complete, retry)
//   - Scheduler: Cron-based execution of recurring task templates
//
// # Quick Start
//
// Create a manager and enqueue tasks:
//
//	import (
//	    "github.com/agincgit/taskforge/pkg/taskforge"
//	    "github.com/agincgit/taskforge/pkg/model"
//	    "github.com/agincgit/taskforge/internal/persistence"
//	)
//
//	// Setup
//	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
//	persistence.Migrate(db)
//
//	mgr, _ := taskforge.NewManager(taskforge.Config{
//	    DB:      db,
//	    Context: context.Background(),
//	})
//
//	// Enqueue
//	task := &model.Task{Type: "email", Payload: `{"to":"user@example.com"}`}
//	mgr.Enqueue(ctx, task)
//
//	// Process
//	reserved, _ := mgr.Reserve(ctx)
//	mgr.Complete(ctx, reserved.ID, true)
//
// # Task Lifecycle
//
// Tasks transition through the following states:
//
//	Pending → InProgress → Succeeded
//	       ↘            ↘ Failed
//	         → PendingCancellation → Cancelled
//
// Failed tasks can be retried, which creates a new task linked to the original.
//
// # Architecture
//
// The package exposes a clean public API while keeping implementation details internal:
//
//   - pkg/taskforge: Manager, Status, Config (this package)
//   - pkg/model: Domain models (Task, TaskTemplate, Worker)
//   - pkg/scheduler: Cron-based recurring task scheduler
//   - internal/: HTTP handlers, persistence, configuration loading
//
// For the default server implementation, see cmd/taskforge.
package taskforge
