package app

import (
	"net/http"

	"github.com/Quiexx/narrator-bot/internal/bot"
	"github.com/Quiexx/narrator-bot/internal/config"
)

func Run(cfg *config.Config) error {

	// dsn := fmt.Sprintf(
	// 	"host=%v user=%v password=%v dbname=%v port=%v sslmode=%v",
	// 	cfg.PostgresHost,
	// 	cfg.PostgresUser,
	// 	cfg.PostgresPassword,
	// 	cfg.PostgresDBName,
	// 	cfg.PostgresPort,
	// 	cfg.PostgresSSMode,
	// )

	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// if err != nil {
	// 	return err
	// }

	// tgurep := repository.NewTgUserRepository(db)

	bot, err := bot.NewTgBot(
		cfg.BotToken,
		cfg.SetWebhookUrl,
		cfg.ServerUrl,
		cfg.WebhookPattern,
	)

	if err != nil {
		return err
	}

	err = bot.SetWebhooks()

	if err != nil {
		return err
	}

	go bot.Start()

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return err
	}

	return nil
}
