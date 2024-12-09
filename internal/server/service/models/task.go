package models

import (
	"jamel/gen/go/jamel"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Task struct {
	Base
	Filename string
	TaskType jamel.TaskType
}

func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.NewString()
	return
	// return tx.SetColumn("ID", uuid.String())
}
