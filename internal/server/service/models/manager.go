package models

import (
	"gorm.io/gorm"
)

type Manager struct {
	Task *TaskManager
}

func New(
	_db *gorm.DB,
) *Manager {
	return &Manager{
		Task: &TaskManager{
			db: _db,
		},
	}
}
