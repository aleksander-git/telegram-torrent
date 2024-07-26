package main

import (
	"log/slog"
	"os"

	"github.com/aleksander-git/telegram-torrent/internal/bot"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("No .env file found")
		os.Exit(1)
	}

	tgbot, err := bot.New(os.Getenv("BOT_TOKEN"), slog.Default())
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	tgbot.Start()
}
