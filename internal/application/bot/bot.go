package bot

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aleksander-git/telegram-torrent/internal/bot"
	"github.com/aleksander-git/telegram-torrent/internal/database"
)

func Run() {
	db := database.New()

	tgbot, err := bot.New(os.Getenv("BOT_TOKEN"), slog.Default(), db)
	if err != nil {
		slog.Error(fmt.Sprintf("cannot run bot: %s", err.Error()))
		os.Exit(1)
	}

	tgbot.Start()
}
