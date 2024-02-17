package repository

import (
	"github.com/Quiexx/narrator-bot/internal/model"
	"gorm.io/gorm"
)

func MigrateModels(db *gorm.DB) error {
	return db.AutoMigrate(&model.TgUser{})
}
