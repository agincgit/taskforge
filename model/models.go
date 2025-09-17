package model

import (
	"time"

	"github.com/google/uuid"
)

// ============================
// Task Models
// ============================
type Task struct {
	BaseModel
	FriendlyID    uint       `gorm:"autoIncrement;not null"`
	Type          string     `gorm:"index;not null"`
	ReferenceID   string     `gorm:"index"`
	Status        string     `gorm:"index;default:'pending'"`
	Payload       string     `gorm:"type:text"`
	Result        string     `gorm:"type:text"`
	TemplateID    *uuid.UUID `gorm:"type:uuid;index"`
	ParentTaskID  *uuid.UUID `gorm:"type:uuid"`
	Attempt       int
	ScheduledFor  *time.Time `gorm:"index"`
	StartedAt     *time.Time
	ItemsTotal    int
	ItemsImpacted int
	ItemsFailed   int
}

type TaskInput struct {
	BaseModel
	TaskID     uint   `gorm:"index;not null"`
	InputKey   string `gorm:"size:255;not null"`
	InputValue string `gorm:"type:text;not null"`
}

type TaskOutput struct {
	BaseModel
	TaskID    uint   `gorm:"index;not null"`
	OutputKey string `gorm:"size:255;not null"`
	Value     string `gorm:"type:text"`
}

type TaskHistory struct {
	BaseModel
	TaskID  uint   `gorm:"index;not null"`
	Status  string `gorm:"not null"`
	Message string `gorm:"type:text"`
}

// ============================
// TaskTemplate Models
// ============================
type TaskTemplate struct {
	BaseModel
	Name           string        `gorm:"size:255;not null"`
	Description    string        `gorm:"type:text"`
	WorkerTypeID   uuid.UUID     `gorm:"type:uuid;not null"`
	IsRecurring    bool          `gorm:"not null"`
	CronSchedule   string        `gorm:"size:255"`
	ExpirationTime time.Duration `gorm:"not null"`
	DefaultInputs  string        `gorm:"type:jsonb"`
}

// ============================
// Worker Models
// ============================
type WorkerType struct {
	BaseModel
	Name        string `gorm:"size:255;not null;unique"`
	Description string `gorm:"type:text"`
}

type WorkerRegistration struct {
	BaseModel
	WorkerTypeID uuid.UUID `gorm:"type:uuid;not null"`
	HostName     string    `gorm:"size:255;not null"`
	StartTime    time.Time `gorm:"not null"`
	ShutdownTime *time.Time
}

type WorkerHeartbeat struct {
	BaseModel
	WorkerID uuid.UUID `gorm:"type:uuid;not null"`
	LastPing time.Time `gorm:"not null"`
}

// ============================
// Queue & Cleanup Models
// ============================
type DeadLetterQueue struct {
	BaseModel
	WorkerID     uuid.UUID `gorm:"type:uuid;not null;index"`
	TaskID       uint      `gorm:"index;not null"`
	FailedAt     time.Time `gorm:"not null"`
	ErrorMessage string    `gorm:"type:text"`
	RetryCount   int       `gorm:"not null;default:0"`
	Handled      bool      `gorm:"default:false"`
}

type TaskCleanup struct {
	BaseModel
	WorkerID       uuid.UUID `gorm:"type:uuid;not null;index"`
	TaskID         uint      `gorm:"index;not null"`
	ExpirationTime time.Time `gorm:"not null"`
	DeletedAt      *time.Time
}

type JobQueue struct {
	BaseModel
	WorkerID       uuid.UUID `gorm:"type:uuid;not null;index"`
	TaskID         uint      `gorm:"index;not null"`
	QueueStatus    string    `gorm:"size:50;not null"`
	WorkerAssigned uuid.UUID `gorm:"type:uuid"`
	EnqueuedAt     time.Time `gorm:"not null"`
	DequeuedAt     *time.Time
}
