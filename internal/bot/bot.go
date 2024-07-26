package bot

import (
	"fmt"
	"log/slog"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	botAPI           *tgbotapi.BotAPI
	logger           *slog.Logger
	usersLastCommand map[string]string
}

func New(token string, logger *slog.Logger) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		return nil, fmt.Errorf("unable to get bot API: %w", err)
	}

	return &Bot{
		botAPI:           bot,
		logger:           logger,
		usersLastCommand: make(map[string]string),
	}, nil
}

func (b *Bot) Start() {
	b.botAPI.Debug = true

	b.logger.Info(fmt.Sprintf("started on account %s", b.botAPI.Self.UserName))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.botAPI.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			b.logger.Info("[%s] %s", update.Message.From.UserName, update.Message.Text)

			err := b.handleUpdate(&update)
			if err != nil {
				b.logger.Error(err.Error())
			}
		}
	}
}
