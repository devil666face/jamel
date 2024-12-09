package models

import "gorm.io/gorm"

type TaskManager struct {
	db *gorm.DB
}
