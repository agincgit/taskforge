// manager.go
package taskforge

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/agincgit/taskforge/model"
)

// Logger is the minimal interface for logging inside the Manager.
// You can pass nil if you don’t need logs.
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// RetryPolicy controls how and when retries happen.
type RetryPolicy struct {
	Attempts int           // total number of tries
	Backoff  time.Duration // delay between retries
}

// TaskForgeConfig configures the Manager programmatically.
type TaskForgeConfig struct {
	DB              *gorm.DB        // your GORM DB handle
	TableName       string          // e.g. "tasks"
	Retry           RetryPolicy     // retry/backoff settings
	CleanupInterval time.Duration   // how often to purge old tasks
	Logger          Logger          // optional logger (may be nil)
	Context         context.Context // root context for operations
}

// Manager is a Go-native façade on top of the Task model.
type Manager struct {
	db      *gorm.DB
	table   string
	retry   RetryPolicy
	cleanup time.Duration
	logger  Logger
	ctx     context.Context
}

// NewManager applies migrations and returns a new Manager.
func NewManager(cfg TaskForgeConfig) (*Manager, error) {
	// Auto-migrate the Task table under the configured name.
	migratedDB, err := DBMigrate(cfg.DB.WithContext(cfg.Context))
	if err != nil {
		return nil, err
	}

	return &Manager{
		db:      migratedDB,
		table:   cfg.TableName,
		retry:   cfg.Retry,
		cleanup: cfg.CleanupInterval,
		logger:  cfg.Logger,
		ctx:     cfg.Context,
	}, nil
}

// Enqueue inserts a new Task with StatusPending.
func (m *Manager) Enqueue(ctx context.Context, t *model.Task) error {
	// cast the typed Status into the string field
	t.Status = string(StatusPending)
	if m.logger != nil {
		m.logger.Infof("Enqueue task with ID=%d", t.ID)
	}
	return m.db.WithContext(ctx).Create(t).Error
}

// Reserve locks & returns the next pending task, marking it in-progress.
func (m *Manager) Reserve(ctx context.Context) (*model.Task, error) {
	var t model.Task

	err := m.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("status = ?", string(StatusPending)).
		Order("id").
		First(&t).
		Error
	if err != nil {
		return nil, err
	}

	t.Status = string(StatusInProgress)
	if m.logger != nil {
		m.logger.Infof("Reserving task ID=%d", t.ID)
	}
	if err := m.db.WithContext(ctx).Save(&t).Error; err != nil {
		return nil, fmt.Errorf("taskforge: reserve failed: %w", err)
	}
	return &t, nil
}

// UpdateStatus sets a Task’s status to any valid value.
func (m *Manager) UpdateStatus(ctx context.Context, id uint, s Status) error {
	if m.logger != nil {
		m.logger.Infof("Updating task ID=%d to status=%s", id, s)
	}
	return m.db.WithContext(ctx).
		Where("id = ?", id).
		Update("status", string(s)).
		Error
}

// Complete marks a Task as complete or failed.
func (m *Manager) Complete(ctx context.Context, id uint, success bool) error {
	newStatus := StatusFailed
	if success {
		newStatus = StatusComplete
	}
	return m.UpdateStatus(ctx, id, newStatus)
}
