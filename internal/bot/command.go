package bot

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aleksander-git/telegram-torrent/internal/database/backend"
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

	notSubscribeAnswerTemplate = "Для доступа к функциям необходимо подписаться на канал %s. Подпишитесь и повторите запрос снова"
)

func validateTorrentLink(link string) error {
	_, err := magneturi.Parse(link)
	if err != nil {
		return fmt.Errorf("torrent link %q is invalid: %w", link, err)
	}
	return nil
}

func (b *Bot) isUserSubscribed(userID int64) (bool, error) {
	ctx := context.Background()
	channelID, err := b.db.GetSetting(ctx, userID, "channel_id")
	if err != nil {
		return false, fmt.Errorf("b.db.GetSetting(%q): %w", "channel_id", err)
	}

	id, err := strconv.ParseInt(channelID, 10, 64)
	if err != nil {
		return false, fmt.Errorf("cannot parse channel_id to int64: %w", err)
	}

	config := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: id,
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

func (b *Bot) handleListTorrentCommand(userName string, chatID int64, userID int64) error {
	delete(b.usersLastCommand, userName)

	msg := tgbotapi.NewMessage(chatID, "")

	ctx := context.Background()
	torrents, err := b.db.GetTorrents(ctx, userID)
	if err != nil {
		b.logger.Error(fmt.Sprintf("b.handleListTorrentCommand(%q, %d): %s", userName, chatID, err))
		msg.Text = unavailableAnswer
	} else {
		msg.Text = torrentsToString(torrents)
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
	} else {
		ctx := context.Background()
		addTorrentParams := backend.AddTorrentParams{
			TorrentLink: link,
			TimeAdded:   time.Now(),
		}
		if err := b.db.AddTorrent(ctx, addTorrentParams); err != nil {
			wrappedErr := fmt.Errorf("adding torrent failed for user %q in chat %d, link %q: %w", userName, chatID, link, err)
			b.logger.Error(wrappedErr.Error())
			msg.Text = unavailableAnswer
		} else {
			delete(b.usersLastCommand, userName)
			msg.Text = addedTorrentAnswer
		}
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
		msg := tgbotapi.NewMessage(chatID, "")

		ctx := context.Background()
		chatLink, err := b.db.GetSetting(ctx, userID, "channel_link")
		if err != nil {
			b.logger.Error(fmt.Sprintf("b.db.GetSetting(%d, %q): %s", userID, "channel_link", err))
			msg.Text = unavailableAnswer
		} else {
			msg.Text = fmt.Sprintf(notSubscribeAnswerTemplate, chatLink)
		}

		_, err = b.botAPI.Send(msg)
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
		err := b.handleListTorrentCommand(userName, chatID, userID)
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

func torrentsToString(torrents []backend.Torrent) string {
	result := ""
	for _, torrent := range torrents {
		result += fmt.Sprintf("Name: %s\n", torrent.Name.String)
	}
	return result
}
