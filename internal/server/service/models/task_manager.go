package models

import (
	"fmt"
	"jamel/gen/go/jamel"

	"gorm.io/gorm"
)

type TaskManager struct {
	db *gorm.DB
}

func (tm *TaskManager) NewTask(task *jamel.TaskResponse) *Task {
	var (
		_task = &Task{}
		mapp  = func(_task *Task) {
			_task.ID = task.TaskId
			_task.Name = task.Name
			_task.TaskType = task.TaskType
			_task.Report = task.Report
			_task.Json = task.Json
			_task.Sbom = task.Sbom
		}
	)
	mapp(_task)
	return _task
}

func (tm *TaskManager) Create(task *Task) error {
	if err := tm.db.FirstOrCreate(&task).Error; err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}
	return nil
}
