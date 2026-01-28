// Package persistence provides database migration and repository implementations.
package persistence

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/agincgit/taskforge/pkg/model"
)

// Migrate runs auto-migrations for all TaskForge models.
func Migrate(db *gorm.DB) error {
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
		return fmt.Errorf("auto-migrate failed: %w", err)
	}
	return nil
}
