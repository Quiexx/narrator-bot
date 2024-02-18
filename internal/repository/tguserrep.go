package repository

import (
	"github.com/Quiexx/narrator-bot/internal/model"
	"gorm.io/gorm"
)

type TgUserRepository struct {
	db *gorm.DB
}

func NewTgUserRepository(db *gorm.DB) *TgUserRepository {
	return &TgUserRepository{db: db}
}

func (r *TgUserRepository) GetOrCreate(tgId int64, defaultState string) (*model.TgUser, error) {
	tgUser := &model.TgUser{}
	result := r.db.First(tgUser, "tg_id = ?", tgId)

	if result.Error == nil {
		return tgUser, nil
	}

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	tgUser.TgId = tgId
	tgUser.VoiceId = -1
	tgUser.State = defaultState
	result = r.db.Create(tgUser)

	if result.Error != nil {
		return nil, result.Error
	}

	return tgUser, nil
}

func (r *TgUserRepository) UpdateUser(tgUser *model.TgUser) error {
	result := r.db.Save(tgUser)
	return result.Error
}
