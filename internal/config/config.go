package config

type Config struct {
	PostgresHost     string `env:"BOT_POSTGRES_HOST" envDefault:"localhost"`
	PostgresPort     string `env:"BOT_POSTGRES_PORT" envDefault:"8432"`
	PostgresDBName   string `env:"POSTGRES_DB" envDefault:"postgres_db"`
	PostgresUser     string `env:"POSTGRES_USER" envDefault:"postgres"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	PostgresSSMode   string `env:"BOT_POSTGRES_SSL_MODE" envDefault:"disable"`

	BotToken       string `env:"BOT_TOKEN" envDefault:"7067695942:AAEdD8gTWgjSPrFthaS_flzhtRapTt6tfWw"`
	SetWebhookUrl  string `env:"SET_WEBHOOK_URL" envDefault:"https://api.telegram.org/bot%v/setWebhook?url=%v%v"`
	ServerUrl      string `env:"SERVER_URL" envDefault:""`
	WebhookPattern string `env:"WEBHOOK_PATTERN" envDefault:"/update"`
}
