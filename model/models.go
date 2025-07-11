package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedBy *string        `gorm:"type:varchar(254);index"`
	UpdatedBy *string        `gorm:"type:varchar(254);index"`
	DeletedBy *string        `gorm:"type:varchar(254);index"`
}

// ============================
// Task Models
// ============================
type Task struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	FriendlyID    uint       `gorm:"autoIncrement;not null"`
	Type          string     `gorm:"index;not null"`
	ReferenceID   string     `gorm:"index"`
	Status        string     `gorm:"index;default:'pending'"`
	Payload       string     `gorm:"type:text"`
	Result        string     `gorm:"type:text"`
	ParentTaskID  *uuid.UUID `gorm:"type:uuid"`
	Attempt       int
	StartedAt     *time.Time
	ItemsTotal    int
	ItemsImpacted int
	ItemsFailed   int
	AuditModel
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		t.ID = id
	}
	return nil
}

type TaskInput struct {
	gorm.Model
	TaskID     uint   `gorm:"index;not null"`
	InputKey   string `gorm:"size:255;not null"`
	InputValue string `gorm:"type:text;not null"`
}

type TaskOutput struct {
	gorm.Model
	TaskID    uint   `gorm:"index;not null"`
	OutputKey string `gorm:"size:255;not null"`
	Value     string `gorm:"type:text"`
}

type TaskHistory struct {
	gorm.Model
	TaskID  uint   `gorm:"index;not null"`
	Status  string `gorm:"not null"`
	Message string `gorm:"type:text"`
}

// ============================
// TaskTemplate Models
// ============================
type TaskTemplate struct {
	gorm.Model
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
	gorm.Model
	Name        string `gorm:"size:255;not null;unique"`
	Description string `gorm:"type:text"`
}

type WorkerRegistration struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	gorm.Model
	WorkerTypeID uuid.UUID `gorm:"type:uuid;not null"`
	HostName     string    `gorm:"size:255;not null"`
	StartTime    time.Time `gorm:"not null"`
	ShutdownTime *time.Time
}

type WorkerHeartbeat struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	gorm.Model
	WorkerID uuid.UUID `gorm:"type:uuid;not null"`
	LastPing time.Time `gorm:"not null"`
}

// ============================
// Queue & Cleanup Models
// ============================
type DeadLetterQueue struct {
	gorm.Model
	WorkerID     uuid.UUID `gorm:"type:uuid;not null;index"`
	TaskID       uint      `gorm:"index;not null"`
	FailedAt     time.Time `gorm:"not null"`
	ErrorMessage string    `gorm:"type:text"`
	RetryCount   int       `gorm:"not null;default:0"`
	Handled      bool      `gorm:"default:false"`
}

type TaskCleanup struct {
	gorm.Model
	WorkerID       uuid.UUID `gorm:"type:uuid;not null;index"`
	TaskID         uint      `gorm:"index;not null"`
	ExpirationTime time.Time `gorm:"not null"`
	DeletedAt      *time.Time
}

type JobQueue struct {
	gorm.Model
	WorkerID       uuid.UUID `gorm:"type:uuid;not null;index"`
	TaskID         uint      `gorm:"index;not null"`
	QueueStatus    string    `gorm:"size:50;not null"`
	WorkerAssigned uuid.UUID `gorm:"type:uuid"`
	EnqueuedAt     time.Time `gorm:"not null"`
	DequeuedAt     *time.Time
}
