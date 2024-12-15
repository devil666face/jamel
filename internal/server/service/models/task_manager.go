package models

import (
	"fmt"
	"jamel/gen/go/jamel"
	"time"

	"gorm.io/gorm"
)

type TaskManager struct {
	db *gorm.DB
}

func (tm *TaskManager) NewTask(task *jamel.TaskResponse, opts ...func(t *Task)) *Task {
	task.CreatedAt = time.Now().Unix()
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
	if len(opts) > 0 {
		opts[0](_task)
	}
	return _task
}

func (tm *TaskManager) Create(task *Task) error {
	if err := tm.db.FirstOrCreate(&task).Error; err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}
	return nil
}

func (tm *TaskManager) All() ([]Task, error) {
	var tasks = []Task{}
	if err := tm.db.Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get all tasks: %w", err)
	}
	return tasks, nil
}

func (tm *TaskManager) Get(id string) (*Task, error) {
	var task = &Task{}
	if err := tm.db.Where("id = ?", id).First(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}
