package bot

import (
	"database/sql"
	"fmt"

	"github.com/go-bittorrent/magneturi"
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

	notSubscribeAnswer = "Для доступа к функциям необходимо подписаться на канал https://t.me/torrent_tbot. Подпишитесь и повторите запрос снова"
)

func validateTorrentLink(link string) error {
	_, err := magneturi.Parse(link)
	if err != nil {
		return fmt.Errorf("torrent link %q is invalid: %w", link, err)
	}

	return nil
}

func (b *Bot) isUserSubscribed(userID int64) (bool, error) {
	channelID, err := b.db.GetSetting("channel", sql.NullInt64{Valid: false})
	if err != nil {
		return false, fmt.Errorf("b.db.GetSetting(%q, %#v): %w", "channel", sql.NullBool{Valid: false}, err)
	}

	config := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: channelID,
			UserID: userID,
		}}
	chatMember, err := b.botAPI.GetChatMember(config)
	if err != nil {
		return false, fmt.Errorf("b.botAPI.GetChatMember(%#v): %w", config, err)
	}

	return chatMember.Status != "left" && chatMember.Status != "kicked", nil
}

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

	torrents, err := b.db.GetTorrents(userName)
	if err != nil {
		b.logger.Error(fmt.Sprintf("b.handleListTorrentCommand(%q, %d): %s", userName, chatID, err))
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

	if err := validateTorrentLink(link); err != nil {
		b.logger.Error(fmt.Sprintf("b.handleAddingNewTorrent(%q, %d, %q): %s", userName, chatID, link, err))
		msg.Text = "Неверная Magnet-ссылка. Проверьте и отправьте снова"
	} else if err := b.db.AddTorrent(userName, link); err != nil {
		b.logger.Error(fmt.Sprintf("b.handleAddingNewTorrent(%q, %d, %q): %s", userName, chatID, link, err))
		msg.Text = unavailableAnswer
	} else {
		delete(b.usersLastCommand, userName)
		msg.Text = addedTorrentAnswer
	}

	_, err := b.botAPI.Send(msg)
	if err != nil {
		return fmt.Errorf("cannot send answer after adding torrent: %w", err)
	}

	return nil
}

func (b *Bot) handleMessage(receivedMessage *tgbotapi.Message) error {
	userName := receivedMessage.From.UserName
	chatID := receivedMessage.Chat.ID
	userID := receivedMessage.From.ID

	subscribed, err := b.isUserSubscribed(userID)
	if err != nil {
		b.logger.Error(fmt.Sprintf("b.isUserSubscribed(%d): %s", userID, err))
		msg := tgbotapi.NewMessage(chatID, unavailableAnswer)
		_, err := b.botAPI.Send(msg)
		if err != nil {
			return fmt.Errorf("cannot send bot unavailable message: %w", err)
		}
	}

	if !subscribed {
		msg := tgbotapi.NewMessage(chatID, notSubscribeAnswer)
		_, err := b.botAPI.Send(msg)
		if err != nil {
			return fmt.Errorf("cannot send not subscribe answer: %w", err)
		}

		return nil
	}

	switch receivedMessage.Text {
	case startCommand:
		err := b.handleStartCommand(userName, chatID)
		if err != nil {
			return fmt.Errorf("b.handleStartCommand(%q, %d): %w", userName, chatID, err)
		}

		return nil

	case helpCommand:
		err := b.handleHelpCommand(userName, chatID)
		if err != nil {
			return fmt.Errorf("b.handleHelpCommand(%q, %d): %w", userName, chatID, err)
		}

		return nil

	case newTorrentCommand:
		err := b.handleNewTorrentCommand(userName, chatID)
		if err != nil {
			return fmt.Errorf("b.handleNewTorrentCommand(%q, %d): %w", userName, chatID, err)
		}

		return nil

	case listTorrentCommand:
		err := b.handleListTorrentCommand(userName, chatID)
		if err != nil {
			return fmt.Errorf("b.handleListTorrentCommand(%q, %d): %w", userName, chatID, err)
		}

		return nil

	default:
		if lastCommand := b.usersLastCommand[userName]; lastCommand == newTorrentCommand {
			err := b.handleAddingNewTorrent(userName, chatID, receivedMessage.Text)
			if err != nil {
				return fmt.Errorf("b.handleAddingNewTorrent(%q, %d, %q): %w", userName, chatID, receivedMessage.Text, err)
			}

			return nil
		}

		msg := tgbotapi.NewMessage(chatID, unknownCommandAnswer)
		_, err := b.botAPI.Send(msg)
		if err != nil {
			return fmt.Errorf("cannot send unknown command answer: %w", err)
		}

		return nil
	}
}
