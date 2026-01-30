package taskforge

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/agincgit/taskforge/pkg/model"
)

// Logger is the minimal interface for logging inside the Manager.
// You can pass nil if you don't need logs.
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// TaskManager defines the core task operations.
type TaskManager interface {
	// Task lifecycle
	Enqueue(ctx context.Context, t *model.Task) error
	Reserve(ctx context.Context) (*model.Task, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, s Status) error
	Complete(ctx context.Context, id uuid.UUID, success bool) error
	CancelTask(ctx context.Context, id uuid.UUID) error
	RetryTask(ctx context.Context, id uuid.UUID) (*model.Task, error)
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]model.Task, error)

	// Task CRUD
	CreateTask(ctx context.Context, t *model.Task) error
	GetTasks(ctx context.Context) ([]model.Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error)
	UpdateTask(ctx context.Context, t *model.Task) error
	DeleteTask(ctx context.Context, id uuid.UUID) error

	// Template operations
	CreateTaskTemplate(ctx context.Context, t *model.TaskTemplate) error
	GetTaskTemplates(ctx context.Context) ([]model.TaskTemplate, error)
	GetTaskTemplate(ctx context.Context, id uuid.UUID) (*model.TaskTemplate, error)
	UpdateTaskTemplate(ctx context.Context, t *model.TaskTemplate) error
	DeleteTaskTemplate(ctx context.Context, id uuid.UUID) error
	CreateTaskFromTemplate(ctx context.Context, templateID uuid.UUID, overrides map[string]interface{}, scheduledFor *time.Time) (*model.Task, error)

	// Worker operations
	RegisterWorker(ctx context.Context, w *model.WorkerRegistration) error
	Heartbeat(ctx context.Context, workerID uuid.UUID) error

	// Queue operations
	EnqueueJob(ctx context.Context, j *model.JobQueue) error
	GetQueue(ctx context.Context) ([]model.JobQueue, error)
	DequeueJob(ctx context.Context, id uuid.UUID) error

	// Child task operations
	GetChildTasks(ctx context.Context, parentID uuid.UUID) ([]model.Task, error)
	HasChildren(ctx context.Context, taskID uuid.UUID) (bool, error)
	GetTaskTree(ctx context.Context, rootID uuid.UUID) (*TaskNode, error)
}

// TemplateScheduler defines the scheduler lifecycle hooks.
type TemplateScheduler interface {
	OnTemplateChanged(tpl model.TaskTemplate) error
	OnTemplateDeleted(templateID uuid.UUID)
}
