package bot

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Quiexx/narrator-bot/internal/model"
	"github.com/Quiexx/narrator-bot/internal/repository"
	"github.com/Quiexx/narrator-bot/internal/steosvoice"
	"github.com/Quiexx/narrator-bot/internal/templates"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	START_COMMAND       = "/start"
	HELP_COMMAND        = "/help"
	API_KEY_COMMAND     = "/apikey"
	VOICE_COMMAND       = "/voice"
	GET_SYMBOLS_COMMAND = "/symbols"

	DEFAULT_STATE     = "DEFAULT"
	SET_API_KEY_STATE = "SET_API_KEY"

	VOICE_PAGE_SIZE = 5

	NOT_ENOUGH_SYMBOLS_ERROR = "Not enough symbols"
	CONNECTION_TIMEOUT_ERROR = "Connection timeout"
)

var userVoicesCash = map[uint][]*steosvoice.Voice{}
var voicesMap = map[int64]*steosvoice.Voice{}

type TgBot struct {
	token          string
	setWebhookUrl  string
	serverUrl      string
	webhookPattern string
	bot            *tgbotapi.BotAPI
	svApi          *steosvoice.SteosVoiceAPI
	tgUserRep      *repository.TgUserRepository
}

func NewTgBot(
	token string,
	setWebhookUrl string,
	serverUrl string,
	webhookPattern string,
	svApi *steosvoice.SteosVoiceAPI,
	tgUserRep *repository.TgUserRepository,
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
		svApi:          svApi,
		tgUserRep:      tgUserRep,
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
	// updateCh := b.bot.ListenForWebhook(b.webhookPattern)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)
	for update := range updates {
		go b.handleUpdate(&update)
	}

	return nil
}

func (b *TgBot) handleUpdate(update *tgbotapi.Update) {
	tgUser, err := b.getOrCreateUser(update)

	if err != nil {
		go b.sendMessage(update, templates.FAIL_MESSAGE)
		log.Printf("failed to get or create user: %v\n", err)
		return
	}

	if tgUser.State == SET_API_KEY_STATE {
		go b.setUserApiKey(update, tgUser)
		return
	}

	switch {
	case b.messageIsCommand(update):
		go b.handleCommand(update, tgUser)
	case update.CallbackData() != "":
		go b.handleCallback(update, tgUser)
	case update.Message != nil && update.Message.Text != "":
		go b.synthesize(update, tgUser)
	case update.Message != nil && update.Message.Caption != "":
		update.Message.Text = update.Message.Caption
		go b.synthesize(update, tgUser)
	default:
		go b.sendMessage(update, templates.CAN_NOT_HANDLE)
	}

}

func (b *TgBot) getOrCreateUser(update *tgbotapi.Update) (*model.TgUser, error) {
	tgUser, err := b.tgUserRep.GetOrCreate(update.FromChat().ID, DEFAULT_STATE)
	if err != nil {
		return nil, err
	}

	return tgUser, nil
}

func (b *TgBot) messageIsCommand(update *tgbotapi.Update) bool {
	if update.Message != nil {
		if strings.HasPrefix(update.Message.Text, "/") && update.Message.Entities != nil {
			for _, ent := range update.Message.Entities {
				if ent.Type == "bot_command" {
					return true
				}
			}
		}
	}
	return false
}

func (b *TgBot) synthesize(update *tgbotapi.Update, tgUser *model.TgUser) {

	if tgUser.SteosvoiceApiKey == "" {
		go b.sendMessage(update, templates.NO_API_KEY_MESSAGE)
		return
	}

	if tgUser.VoiceId == -1 {
		go b.sendMessage(update, templates.NO_VOICE_MESSAGE)
		return
	}

	result, err := b.svApi.GetSynthesizedSpeech(tgUser.SteosvoiceApiKey, update.Message.Text, tgUser.VoiceId)

	switch {
	case err != nil:
		log.Printf("failed to synthesize text: %v\n", err)
		go b.sendMessage(update, templates.SOMETHING_GONE_WRONG_MESSAGE)
	case !result.Status && result.Message == NOT_ENOUGH_SYMBOLS_ERROR:
		go b.sendMessage(update, templates.NOT_ENOUGH_SYMBOLS)
	case !result.Status && result.Message == CONNECTION_TIMEOUT_ERROR:
		go b.sendMessage(update, templates.SERVICE_NOT_AVAILABLE)
	case !result.Status:
		log.Printf("failed to synthesize text: %v\n", result.Message)
		go b.sendMessage(update, templates.SOMETHING_GONE_WRONG_MESSAGE)
	default:
		go b.sendVoice(update, result.AudioUrl)
	}

}

func (b *TgBot) sendMessage(update *tgbotapi.Update, text string) {
	msg := tgbotapi.NewMessage(update.FromChat().ID, text)
	msg.ParseMode = "Markdown"
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("failed to send message: %v\n", err)
	}
}

func (b *TgBot) sendMessageWithKeyboard(update *tgbotapi.Update, text string, keyboardMarkup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(update.FromChat().ID, text)
	msg.ReplyMarkup = keyboardMarkup
	msg.ParseMode = "Markdown"

	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("failed to send message: %v\n", err)
	}
}

func (b *TgBot) editMessageWithKeyboard(update *tgbotapi.Update, keyboardMarkup tgbotapi.InlineKeyboardMarkup, messageId int) {
	msg := tgbotapi.NewEditMessageReplyMarkup(
		update.FromChat().ID,
		messageId,
		keyboardMarkup,
	)

	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("failed to send message: %v\n", err)
	}
}

func (b *TgBot) editMessage(update *tgbotapi.Update, text string, messageId int) {
	msg := tgbotapi.NewEditMessageText(
		update.FromChat().ID,
		messageId,
		text,
	)
	msg.ParseMode = "Markdown"

	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("failed to send message: %v\n", err)
	}
}

func (b *TgBot) sendVoice(update *tgbotapi.Update, url string) {
	fu := tgbotapi.FileURL(url)
	msg := tgbotapi.NewVoice(update.FromChat().ID, fu)
	msg.ReplyToMessageID = update.Message.MessageID
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Printf("failed to send message: %v\n", err)
	}
}

func (b *TgBot) handleCommand(update *tgbotapi.Update, tgUser *model.TgUser) {

	switch {
	case strings.HasPrefix(update.Message.Text, START_COMMAND):
		go b.handleStart(update, tgUser)
	case strings.HasPrefix(update.Message.Text, API_KEY_COMMAND):
		go b.handleApiKey(update, tgUser)
	case strings.HasPrefix(update.Message.Text, VOICE_COMMAND):
		go b.handleVoice(update, tgUser)
	case strings.HasPrefix(update.Message.Text, GET_SYMBOLS_COMMAND):
		go b.handleGetSymbols(update, tgUser)
	default:
		go b.sendMessage(update, templates.UNKNOWN_COMMAND_MESSAGE)
	}

}

func (b *TgBot) handleGetSymbols(update *tgbotapi.Update, tgUser *model.TgUser) {
	if tgUser.SteosvoiceApiKey == "" {
		b.sendMessage(update, templates.NO_API_KEY_MESSAGE)
		return
	}

	result, err := b.svApi.GetSymbols(tgUser.SteosvoiceApiKey)
	if err != nil || !result.Status {
		log.Printf("failed to get symbols count: %v\n", err)
		b.sendMessage(update, templates.FAIL_COMMAND_MESSAGE)
	} else {
		b.sendMessage(update, fmt.Sprintf(templates.SYMBOL_COUNT_MESSAGE, result.Symbols))
	}
}

func (b *TgBot) handleStart(update *tgbotapi.Update, tgUser *model.TgUser) {
	ticker := time.NewTicker(time.Second)
	for _, text := range templates.StartMessages {
		<-ticker.C
		b.sendMessage(update, text)
	}
}

func (b *TgBot) handleApiKey(update *tgbotapi.Update, tgUser *model.TgUser) {
	tgUser.State = SET_API_KEY_STATE
	err := b.tgUserRep.UpdateUser(tgUser)
	if err != nil {
		go b.sendMessage(update, templates.SOMETHING_GONE_WRONG_MESSAGE)
		return
	}
	b.sendMessage(update, templates.SET_API_KEY_MESSAGE)
}

func (b *TgBot) setUserApiKey(update *tgbotapi.Update, tgUser *model.TgUser) {
	if update.Message == nil {
		go b.sendMessage(update, templates.NOT_API_KEY_MESSAGE)
		return
	}

	if update.Message.Text == "" {
		go b.sendMessage(update, templates.NOT_API_KEY_MESSAGE)
		return
	}

	apiKey := update.Message.Text
	result, err := b.svApi.GetVoices(apiKey)
	if err != nil || !result.Status {
		go b.sendMessage(update, templates.FAILD_TO_CONNECT_API_KEY_MESSAGE)
		return
	}

	tgUser.SteosvoiceApiKey = apiKey

	if len(result.Voices) != 0 {
		tgUser.VoiceId = result.Voices[0].Id
	}

	tgUser.State = DEFAULT_STATE
	err = b.tgUserRep.UpdateUser(tgUser)
	if err != nil {
		go b.sendMessage(update, templates.SOMETHING_GONE_WRONG_MESSAGE)
		return
	}

	go b.sendMessage(update, templates.API_KEY_IS_SET_MESSAGE)
	go b.updateUserVoices(tgUser)
}

func (b *TgBot) handleVoice(update *tgbotapi.Update, tgUser *model.TgUser) {
	if tgUser.SteosvoiceApiKey == "" {
		go b.sendMessage(update, templates.NO_API_KEY_MESSAGE)
		return
	}

	go b.updateUserVoices(tgUser)
	go b.sendVoicesMarkup(update, tgUser, 1, VOICE_PAGE_SIZE, false)
}

func (b *TgBot) handleCallback(update *tgbotapi.Update, tgUser *model.TgUser) {

	data := update.CallbackData()
	switch {
	case strings.Split(data, "_")[0] == "voice":
		go b.handleVoiceCallback(update, tgUser)
	}
}

func (b *TgBot) handleVoiceCallback(update *tgbotapi.Update, tgUser *model.TgUser) {

	data := update.CallbackData()
	tokens := strings.Split(data, "_")

	if len(tokens) > 2 && tokens[1] == "page" {
		page, err := strconv.Atoi(tokens[2])
		if err == nil {
			go b.sendVoicesMarkup(update, tgUser, page, VOICE_PAGE_SIZE, true)
		}
		return
	}

	if len(tokens) > 3 && tokens[1] == "info" {
		voiceId, err := strconv.Atoi(tokens[2])
		if err != nil {
			return
		}
		page, err := strconv.Atoi(tokens[3])

		if err == nil {
			go b.sendVoiceDescription(update, tgUser, int64(voiceId), page)
		}
		return
	}

	voiceId, err := strconv.Atoi(tokens[1])
	if err != nil {
		go b.sendMessage(update, templates.SOMETHING_GONE_WRONG_MESSAGE)
		return
	}

	tgUser.VoiceId = int64(voiceId)

	err = b.tgUserRep.UpdateUser(tgUser)
	if err != nil {
		go b.sendMessage(update, templates.SOMETHING_GONE_WRONG_MESSAGE)
		return
	}

	go b.sendMessage(update, templates.VOICE_IS_SET_MESSAGE)

}

func (b *TgBot) sendVoicesMarkup(update *tgbotapi.Update, tgUser *model.TgUser, startPage int, pageSize int, edit bool) {

	voices, err := b.getUserVoices(tgUser)

	if err != nil {
		go b.sendMessage(update, templates.SOMETHING_GONE_WRONG_MESSAGE)
		return
	}

	keyboard := [][]tgbotapi.InlineKeyboardButton{}
	start := (startPage - 1) * pageSize

	for i, voice := range voices {
		if i == start+pageSize {
			break
		}

		if i >= start {
			name, ok := voice.Name["RU"]
			if !ok {
				name, ok = voice.Name["EN"]
			}
			if !ok {
				name = fmt.Sprint(voice.Id)
			}

			callbackData := fmt.Sprintf("voice_info_%v_%v", voice.Id, startPage)

			keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
				{
					Text:         name,
					CallbackData: &callbackData,
				},
			})
		}
	}

	nextPageCallback := fmt.Sprintf("voice_page_%v", startPage+1)
	prevPageCallback := fmt.Sprintf("voice_page_%v", startPage-1)

	navigation := []tgbotapi.InlineKeyboardButton{}

	maxPage := int(math.Ceil(float64(len(voices)) / float64(pageSize)))

	if startPage != 1 {
		navigation = append(navigation, tgbotapi.InlineKeyboardButton{
			Text:         "<",
			CallbackData: &prevPageCallback,
		})
	}

	if startPage != maxPage {
		navigation = append(navigation, tgbotapi.InlineKeyboardButton{
			Text:         ">",
			CallbackData: &nextPageCallback,
		})
	}

	keyboard = append(keyboard, navigation)

	keyboardMarkup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	text := fmt.Sprintf(
		"%v\n\nСтраница %v/%v",
		templates.VOICE_LIST_MESSAGE,
		startPage,
		maxPage,
	)

	if !edit {
		b.sendMessageWithKeyboard(update, text, keyboardMarkup)
		return
	}
	b.editMessage(update, text, update.CallbackQuery.Message.MessageID)
	b.editMessageWithKeyboard(update, keyboardMarkup, update.CallbackQuery.Message.MessageID)
}

func (b *TgBot) sendVoiceDescription(update *tgbotapi.Update, tgUser *model.TgUser, voiceId int64, page int) {

	voice, ok := voicesMap[voiceId]

	if !ok {
		go b.sendMessage(update, templates.SOMETHING_GONE_WRONG_MESSAGE)
		return
	}

	name, ok := voice.Name["RU"]
	if !ok {
		name, ok = voice.Name["EN"]
	}
	if !ok {
		name = fmt.Sprint(voice.Id)
	}

	description, ok := voice.Description["RU"]
	if !ok {
		description, ok = voice.Description["EN"]
	}
	if !ok {
		description = ""
	}

	sex := voice.Sex
	if sex == "" {
		sex = "Неизвестен"
	}

	text := fmt.Sprintf(
		"**%v**\n%v\n\nПол: %v",
		name,
		description,
		sex,
	)

	backCallback := fmt.Sprintf("voice_page_%v", page)
	chooseCallback := fmt.Sprintf("voice_%v", voiceId)

	keyboard := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.InlineKeyboardButton{
				Text:         "Выбрать",
				CallbackData: &chooseCallback,
			},
		},
		{
			tgbotapi.InlineKeyboardButton{
				Text:         "Назад",
				CallbackData: &backCallback,
			},
		},
	}

	keyboardMarkup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	b.editMessage(update, text, update.CallbackQuery.Message.MessageID)
	b.editMessageWithKeyboard(update, keyboardMarkup, update.CallbackQuery.Message.MessageID)
}

func (b *TgBot) getUserVoices(tgUser *model.TgUser) ([]*steosvoice.Voice, error) {
	voices, ok := userVoicesCash[tgUser.ID]
	if ok {
		return voices, nil
	}

	result, err := b.svApi.GetVoices(tgUser.SteosvoiceApiKey)
	if err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, errors.New("failed to get user voices")
	}

	voices = result.Voices
	sort.Slice(voices, func(i, j int) bool { return voices[i].Id < voices[j].Id })
	userVoicesCash[tgUser.ID] = voices

	go b.updateVoicesMap(voices)

	return voices, nil
}

func (b *TgBot) updateUserVoices(tgUser *model.TgUser) ([]*steosvoice.Voice, error) {

	result, err := b.svApi.GetVoices(tgUser.SteosvoiceApiKey)
	if err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, errors.New("failed to get user voices")
	}

	voices := result.Voices
	sort.Slice(voices, func(i, j int) bool { return voices[i].Id < voices[j].Id })
	userVoicesCash[tgUser.ID] = voices

	go b.updateVoicesMap(voices)

	return voices, nil
}

func (b *TgBot) updateVoicesMap(voices []*steosvoice.Voice) {
	for _, voice := range voices {
		voicesMap[voice.Id] = voice
	}
}
