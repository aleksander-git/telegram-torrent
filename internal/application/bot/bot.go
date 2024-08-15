package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/aleksander-git/telegram-torrent/internal/bot"
	"github.com/aleksander-git/telegram-torrent/internal/database/backend"
)

func Run() {
	logger := slog.New(slog.NewTextHandler(log.Writer(), nil))

	// Создаем объект базы данных
	db, err := backend.NewDatabase(os.Getenv("DATABASE_CONNECTION_STRING"))
	if err != nil {
		logger.Error("unable to create database", "error", err)
		return
	}

	tgbot, err := bot.New(os.Getenv("BOT_TOKEN"), logger, db)
	if err != nil {
		logger.Error("unable to create bot", "error", err)
		return
	}

	tgbot.Start()
}
