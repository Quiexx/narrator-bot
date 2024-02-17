package repository

import "gorm.io/gorm"

type TgUserRepository struct {
	db *gorm.DB
}

func NewTgUserRepository(db *gorm.DB) *TgUserRepository {
	return &TgUserRepository{db: db}
}
