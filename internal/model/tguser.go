package model

import "gorm.io/gorm"

type TgUser struct {
	gorm.Model
	TgId             int64
	SteosvoiceApiKey string
	VoiceId          int64
}
