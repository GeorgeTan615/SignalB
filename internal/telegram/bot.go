package telegram

import (
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var Bot *BotClient

type BotClient struct {
	DefaultChatID int64
	BotAPI        *tgbotapi.BotAPI
}

func InitBot() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Failed to load .env file", err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	chatID, err := strconv.ParseInt(os.Getenv("TELEGRAM_CHAT_ID"), 10, 64)
	if err != nil {
		log.Println("error parsing chat id", err)
	}

	Bot = &BotClient{
		BotAPI:        bot,
		DefaultChatID: chatID,
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	go Bot.listenForUpdates()
}

func (b *BotClient) listenForUpdates() {
	u := tgbotapi.NewUpdate(0) // Set offset to 0
	u.Timeout = 60
	updates := b.BotAPI.GetUpdatesChan(u)

	for update := range updates {
		message := update.Message
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		log.Printf("%s wrote %s", message.From.FirstName, message.Text)

		var text string

		switch update.Message.Command() {
		case "register":
			text = "Multi-user registration not allowed at the moment."
		default:
			text = "Invalid command. /register command would be supported in the future."
		}

		if err := b.SendMessageByHTML(update.Message.Chat.ID, text); err != nil {
			log.Println("error sending message to user", err)
		}
	}
}

func (b *BotClient) SendMessageByHTML(chatID int64, message string) error {
	if message == "" {
		return nil
	}

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:              chatID,
			ReplyToMessageID:    0,
			DisableNotification: false,
		},
		ParseMode:             "HTML",
		Text:                  message,
		DisableWebPagePreview: false,
	}

	_, err := b.BotAPI.Send(msg)
	return err
}
