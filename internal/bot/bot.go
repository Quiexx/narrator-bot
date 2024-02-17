package bot

import (
	"fmt"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgBot struct {
	token          string
	setWebhookUrl  string
	serverUrl      string
	webhookPattern string
	bot            *tgbotapi.BotAPI
}

func NewTgBot(
	token string,
	setWebhookUrl string,
	serverUrl string,
	webhookPattern string,
) (*TgBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &TgBot{
		token:          token,
		setWebhookUrl:  setWebhookUrl,
		serverUrl:      serverUrl,
		webhookPattern: webhookPattern,
		bot:            bot,
	}, nil
}

func (b *TgBot) SetWebhooks() error {

	_, err := http.Post(
		fmt.Sprintf(b.setWebhookUrl, b.token, b.serverUrl, b.webhookPattern),
		"plain/text",
		nil,
	)

	return err
}

func (b *TgBot) Start() error {
	updateCh := b.bot.ListenForWebhook(b.webhookPattern)

	for update := range updateCh {
		log.Printf("%v: %v\n", update.FromChat().ID, update.Message.Text)
	}

	return nil
}
