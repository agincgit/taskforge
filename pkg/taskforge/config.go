package taskforge

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// RetryPolicy controls how and when retries happen.
type RetryPolicy struct {
	Attempts int           // total number of tries
	Backoff  time.Duration // delay between retries
}

// Config configures the Manager programmatically.
type Config struct {
	DB              *gorm.DB        // your GORM DB handle
	TableName       string          // e.g. "tasks"
	Retry           RetryPolicy     // retry/backoff settings
	CleanupInterval time.Duration   // how often to purge old tasks
	Logger          Logger          // optional logger (may be nil)
	Context         context.Context // root context for operations
}
