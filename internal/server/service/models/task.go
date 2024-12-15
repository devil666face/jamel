package models

import (
	"time"

	"jamel/gen/go/jamel"

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
	Name     string
	Report   string
	Json     string
	Sbom     string
	TaskType jamel.TaskType
}

func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	return
}

func (t *Task) TaskToResp(opts ...func(t *Task)) *jamel.TaskResponse {
	if len(opts) > 0 {
		opts[0](t)
	}
	return &jamel.TaskResponse{
		TaskId:    t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt.Unix(),
		TaskType:  t.TaskType,
		Sbom:      t.Sbom,
		Json:      t.Json,
		Report:    t.Report,
	}
}
