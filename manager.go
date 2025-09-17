// manager.go
package taskforge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	cfg     TaskForgeConfig
	db      *gorm.DB
	table   string
	retry   RetryPolicy
	cleanup time.Duration
	logger  Logger
	ctx     context.Context
}

// NewManager applies migrations and returns a new Manager.
func NewManager(cfg TaskForgeConfig) (*Manager, error) {
	if cfg.DB == nil {
		return nil, errors.New("taskforge: DB is required")
	}

	// Auto-migrate the Task table under the configured name.
	migratedDB, err := DBMigrate(cfg.DB.WithContext(cfg.Context))
	if err != nil {
		return nil, err
	}
	cfg.DB = migratedDB

	return &Manager{
		cfg:     cfg,
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
	t.Status = string(StatusPending)
	if m.logger != nil {
		m.logger.Infof("Enqueue task with ID=%s", t.ID)
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
		m.logger.Infof("Reserving task ID=%s", t.ID)
	}
	if err := m.db.WithContext(ctx).Save(&t).Error; err != nil {
		return nil, fmt.Errorf("taskforge: reserve failed: %w", err)
	}
	return &t, nil
}

// UpdateStatus sets a Task’s status to any valid value.
func (m *Manager) UpdateStatus(ctx context.Context, id uuid.UUID, s Status) error {
	if m.logger != nil {
		m.logger.Infof("Updating task ID=%s to status=%s", id, s)
	}
	return m.db.WithContext(ctx).
		Where("id = ?", id).
		Update("status", string(s)).
		Error
}

// Complete marks a Task as complete or failed.
func (m *Manager) Complete(ctx context.Context, id uuid.UUID, success bool) error {
	newStatus := StatusFailed
	if success {
		newStatus = StatusSucceeded
	}
	return m.UpdateStatus(ctx, id, newStatus)
}

// CancelTask attempts to cancel a task that is pending or in progress.
func (m *Manager) CancelTask(ctx context.Context, id uuid.UUID) error {
	return m.db.WithContext(ctx).
		Where("id = ? AND status IN ?", id, []string{string(StatusPending), string(StatusInProgress)}).
		Updates(map[string]interface{}{"status": string(StatusPendingCancel)}).
		Error
}

// RetryTask clones a failed task into a new retry task.
func (m *Manager) RetryTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var t model.Task
	if err := m.db.WithContext(ctx).First(&t, "id = ? AND status = ?", id, string(StatusFailed)).Error; err != nil {
		return nil, err
	}

	newTask := t
	newTask.ID = uuid.Nil
	newTask.FriendlyID = 0
	newTask.CreatedAt = time.Time{}
	newTask.UpdatedAt = time.Time{}
	newTask.DeletedAt = gorm.DeletedAt{}
	newTask.Status = string(StatusPending)
	newTask.ParentTaskID = &t.ID
	newTask.Attempt = t.Attempt + 1

	if err := m.db.WithContext(ctx).Create(&newTask).Error; err != nil {
		return nil, err
	}
	return &newTask, nil
}

// List returns tasks filtered by optional fields with pagination.
func (m *Manager) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]model.Task, error) {
	var tasks []model.Task
	db := m.db.WithContext(ctx)
	if v, ok := filter["type"]; ok {
		db = db.Where("type = ?", v)
	}
	if v, ok := filter["status"]; ok {
		db = db.Where("status = ?", v)
	}
	if v, ok := filter["reference_id"]; ok {
		db = db.Where("reference_id = ?", v)
	}
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}
	if err := db.Order("friendly_id").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// --- Basic CRUD operations ---

// CreateTask stores a new task record.
func (m *Manager) CreateTask(ctx context.Context, t *model.Task) error {
	if t.Type == "" {
		return fmt.Errorf("taskforge: task type required")
	}
	return m.cfg.DB.WithContext(ctx).Create(t).Error
}

// CreateTaskFromTemplate creates a new task instance from a stored template.
func (m *Manager) CreateTaskFromTemplate(ctx context.Context, templateID uuid.UUID, overrides map[string]interface{}, scheduledFor *time.Time) (*model.Task, error) {
	var tpl model.TaskTemplate
	if err := m.cfg.DB.WithContext(ctx).First(&tpl, "id = ?", templateID).Error; err != nil {
		return nil, fmt.Errorf("taskforge: failed to load template %s: %w", templateID, err)
	}

	var worker model.WorkerType
	if err := m.cfg.DB.WithContext(ctx).First(&worker, "id = ?", tpl.WorkerTypeID).Error; err != nil {
		return nil, fmt.Errorf("taskforge: failed to load worker type %s: %w", tpl.WorkerTypeID, err)
	}

	merged := make(map[string]interface{})
	if tpl.DefaultInputs != "" {
		if err := json.Unmarshal([]byte(tpl.DefaultInputs), &merged); err != nil {
			return nil, fmt.Errorf("taskforge: invalid template inputs: %w", err)
		}
	}

	for key, value := range overrides {
		merged[key] = value
	}

	payloadBytes, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("taskforge: unable to encode task payload: %w", err)
	}

	tplID := tpl.ID
	task := &model.Task{
		Type:         worker.Name,
		Status:       string(StatusPending),
		Payload:      string(payloadBytes),
		TemplateID:   &tplID,
		ScheduledFor: scheduledFor,
	}

	if err := m.cfg.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}

		if len(merged) == 0 {
			return nil
		}

		inputs := make([]model.TaskInput, 0, len(merged))
		for key, value := range merged {
			valueBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("taskforge: unable to encode input %s: %w", key, err)
			}
			inputs = append(inputs, model.TaskInput{
				TaskID:     task.FriendlyID,
				InputKey:   key,
				InputValue: string(valueBytes),
			})
		}
		if len(inputs) > 0 {
			if err := tx.Create(&inputs).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return task, nil
}

// GetTasks retrieves all tasks.
func (m *Manager) GetTasks(ctx context.Context) ([]model.Task, error) {
	var tasks []model.Task
	if err := m.cfg.DB.WithContext(ctx).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTask fetches a task by its ID.
func (m *Manager) GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var t model.Task
	if err := m.cfg.DB.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateTask persists changes to an existing task.
func (m *Manager) UpdateTask(ctx context.Context, t *model.Task) error {
	if t.ID == uuid.Nil {
		return fmt.Errorf("taskforge: missing task ID")
	}
	return m.cfg.DB.WithContext(ctx).Save(t).Error
}

// DeleteTask removes a task by ID.
func (m *Manager) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return m.cfg.DB.WithContext(ctx).Delete(&model.Task{}, "id = ?", id).Error
}

// CreateTaskTemplate stores a new task template.
func (m *Manager) CreateTaskTemplate(ctx context.Context, t *model.TaskTemplate) error {
	return m.cfg.DB.WithContext(ctx).Create(t).Error
}

// GetTaskTemplates retrieves all task templates.
func (m *Manager) GetTaskTemplates(ctx context.Context) ([]model.TaskTemplate, error) {
	var tpls []model.TaskTemplate
	if err := m.cfg.DB.WithContext(ctx).Find(&tpls).Error; err != nil {
		return nil, err
	}
	return tpls, nil
}

// GetTaskTemplate fetches a task template by ID.
func (m *Manager) GetTaskTemplate(ctx context.Context, id uuid.UUID) (*model.TaskTemplate, error) {
	var tpl model.TaskTemplate
	if err := m.cfg.DB.WithContext(ctx).First(&tpl, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tpl, nil
}

// UpdateTaskTemplate saves changes to a task template.
func (m *Manager) UpdateTaskTemplate(ctx context.Context, t *model.TaskTemplate) error {
	if t.ID == uuid.Nil {
		return fmt.Errorf("taskforge: missing template ID")
	}
	return m.cfg.DB.WithContext(ctx).Save(t).Error
}

// DeleteTaskTemplate removes a template by ID.
func (m *Manager) DeleteTaskTemplate(ctx context.Context, id uuid.UUID) error {
	return m.cfg.DB.WithContext(ctx).Delete(&model.TaskTemplate{}, "id = ?", id).Error
}

// RegisterWorker persists a worker registration.
func (m *Manager) RegisterWorker(ctx context.Context, w *model.WorkerRegistration) error {
	return m.cfg.DB.WithContext(ctx).Create(w).Error
}

// Heartbeat updates a worker heartbeat record.
func (m *Manager) Heartbeat(ctx context.Context, workerID uuid.UUID) error {
	var beat model.WorkerHeartbeat
	if err := m.cfg.DB.WithContext(ctx).First(&beat, "worker_id = ?", workerID).Error; err != nil {
		return err
	}
	beat.LastPing = time.Now()
	return m.cfg.DB.WithContext(ctx).Save(&beat).Error
}

// EnqueueJob adds a job to the worker queue.
func (m *Manager) EnqueueJob(ctx context.Context, j *model.JobQueue) error {
	return m.cfg.DB.WithContext(ctx).Create(j).Error
}

// GetQueue returns all queued jobs.
func (m *Manager) GetQueue(ctx context.Context) ([]model.JobQueue, error) {
	var q []model.JobQueue
	if err := m.cfg.DB.WithContext(ctx).Find(&q).Error; err != nil {
		return nil, err
	}
	return q, nil
}

// DequeueJob removes a job from the queue by ID.
func (m *Manager) DequeueJob(ctx context.Context, id uuid.UUID) error {
	return m.cfg.DB.WithContext(ctx).Delete(&model.JobQueue{}, "id = ?", id).Error
}
