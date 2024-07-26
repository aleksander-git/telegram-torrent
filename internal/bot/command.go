package bot

import (
	"errors"
	"fmt"

	"github.com/aleksander-git/telegram-torrent/internal/database"
	"github.com/aleksander-git/telegram-torrent/internal/torrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	startCommand       = "/start"
	helpCommand        = "/help"
	newTorrentCommand  = "/newtorrent"
	listTorrentCommand = "/listtorrents"

	helpAnswer = `Я могу помочь со скачиванием торрентов!
	
Вы можете обращаться ко мне, используя эти команды:

/newtorrent - добавить новый торрент
/listtorrents - посмотреть список торрентов`

	newTorrentAnswer = "Хорошо, теперь введите Magnet-ссылку на торрент"

	addedTorrentAnswer = "Торрент успешно добавлен"

	unknownCommandAnswer = "Неопознанная команда. Что вы хотели сказать?"

	unavailableAnswer = "Сервер в данный момент не доступен. Повторите запрос позже"
)

func (b *Bot) handleStartCommand(userName string, chatID int64) error {
	delete(b.usersLastCommand, userName)

	msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! "+helpAnswer)
	_, err := b.botAPI.Send(msg)
	if err != nil {
		return fmt.Errorf("cannot send start answer: %w", err)
	}

	return nil
}

func (b *Bot) handleHelpCommand(userName string, chatID int64) error {
	delete(b.usersLastCommand, userName)

	msg := tgbotapi.NewMessage(chatID, helpAnswer)
	_, err := b.botAPI.Send(msg)
	if err != nil {
		return fmt.Errorf("cannot send help answer: %w", err)
	}

	return nil
}

func (b *Bot) handleNewTorrentCommand(userName string, chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, newTorrentAnswer)
	if _, err := b.botAPI.Send(msg); err != nil {
		return fmt.Errorf("cannot send new torrent answer: %w", err)
	}

	b.usersLastCommand[userName] = newTorrentCommand
	return nil
}

func (b *Bot) handleListTorrentCommand(userName string, chatID int64) error {
	delete(b.usersLastCommand, userName)

	msg := tgbotapi.NewMessage(chatID, "")

	torrents, err := database.GetTorrents(userName)
	if err != nil {
		b.logger.Error(err.Error())
		msg.Text = unavailableAnswer
	} else {
		msg.Text = torrents.String()
		msg.ParseMode = "Markdown"
	}

	_, err = b.botAPI.Send(msg)
	if err != nil {
		return fmt.Errorf("cannot send list torrent answer: %w", err)
	}

	return nil
}

func (b *Bot) handleAddingNewTorrent(userName string, chatID int64, link string) error {
	msg := tgbotapi.NewMessage(chatID, "")

	err := database.AddTorrent(userName, link)
	if err != nil {
		if errors.As(err, &torrent.TorrentParseError{}) {
			b.logger.Info(err.Error())
			msg.Text = "Неверная Magnet-ссылка. Проверьте и отправьте снова"
		} else {
			b.logger.Error(err.Error())
			msg.Text = unavailableAnswer
		}
	} else {
		delete(b.usersLastCommand, userName)
		msg.Text = addedTorrentAnswer
	}

	_, err = b.botAPI.Send(msg)
	if err != nil {
		return fmt.Errorf("cannot send answer after adding torrent: %w", err)
	}

	return nil
}

func (b *Bot) handleUpdate(update *tgbotapi.Update) error {
	userName := update.Message.From.UserName
	chatID := update.Message.Chat.ID
	message := update.Message.Text

	switch message {
	case startCommand:
		return b.handleStartCommand(userName, chatID)

	case helpCommand:
		return b.handleHelpCommand(userName, chatID)

	case newTorrentCommand:
		return b.handleNewTorrentCommand(userName, chatID)

	case listTorrentCommand:
		return b.handleListTorrentCommand(userName, chatID)

	default:
		if lastCommand := b.usersLastCommand[userName]; lastCommand == newTorrentCommand {
			return b.handleAddingNewTorrent(userName, chatID, message)
		}

		msg := tgbotapi.NewMessage(chatID, unknownCommandAnswer)
		_, err := b.botAPI.Send(msg)
		if err != nil {
			return fmt.Errorf("cannot send unknown command answer: %w", err)
		}

		return nil
	}
}
