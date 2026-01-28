package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	CreatedBy *string `gorm:"type:varchar(254);index"`
	UpdatedBy *string `gorm:"type:varchar(254);index"`
	DeletedBy *string `gorm:"type:varchar(254);index"`
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	m.ID = id
	return nil
}
