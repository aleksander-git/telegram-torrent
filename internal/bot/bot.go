package bot

import (
	"fmt"
	"log/slog"

	"github.com/aleksander-git/telegram-torrent/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	botAPI           *tgbotapi.BotAPI
	logger           *slog.Logger
	usersLastCommand map[string]string
	db               database.DBInterface
}

func New(token string, logger *slog.Logger, db database.DBInterface) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("unable to get bot API: %w", err)
	}

	return &Bot{
		botAPI:           bot,
		logger:           logger,
		usersLastCommand: make(map[string]string),
		db:               db,
	}, nil
}

func (b *Bot) Start() {
	b.botAPI.Debug = true

	b.logger.Info(fmt.Sprintf("started on account %s", b.botAPI.Self.UserName))

	offset := 0
	u := tgbotapi.NewUpdate(offset)

	timeoutSeconds := 60
	u.Timeout = timeoutSeconds

	updates := b.botAPI.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			b.logger.Info("received message %q from %q", update.Message.Text, update.Message.From.UserName)

			err := b.handleMessage(update.Message)
			if err != nil {
				b.logger.Error(fmt.Sprintf("cannot handle message: %s", err))
			}
		}
	}
}
