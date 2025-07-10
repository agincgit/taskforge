package taskforge

import (
	"fmt"

	"github.com/agincgit/taskforge/model"
	"gorm.io/gorm"
)

// Status is the lifecycle state of a Task in TaskForge.
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusFailed     Status = "failed"
	StatusComplete   Status = "complete"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusFailed, StatusComplete:
		return true
	}
	return false
}

// DBMigrate will create and migrate the tables, and then make the some relationships if necessary
func DBMigrate(db *gorm.DB) (*gorm.DB, error) {
	if err := db.AutoMigrate(
		&model.Task{},
		&model.TaskInput{},
		&model.TaskOutput{},
		&model.TaskHistory{},
		&model.TaskTemplate{},
		&model.WorkerType{},
		&model.WorkerRegistration{},
		&model.WorkerHeartbeat{},
		&model.DeadLetterQueue{},
		&model.TaskCleanup{},
		&model.JobQueue{},
	); err != nil {
		return nil, fmt.Errorf("auto-migrate failed: %v", err)
	}
	return db, nil
}
