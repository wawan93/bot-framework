package bot_framework

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"errors"
	"log"
)

type Sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type commonHandler func(bot Sender, update *tgbotapi.Update) error

type BotFramework struct {
	Sender
	commands         map[string]commonHandler
	messages         chan tgbotapi.Chattable
	plainTextHandler commonHandler
	photoHandler     commonHandler
	fileHandler      commonHandler
}

func NewBotFramework(api Sender) *BotFramework {
	bot := BotFramework{
		api,
		make(map[string]commonHandler),
		make(chan tgbotapi.Chattable),
		nil,
		nil,
		nil,
	}
	return &bot
}

func (bot *BotFramework) HandleUpdates(ch tgbotapi.UpdatesChannel) {
	for update := range ch {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
				}
			}()
			err := bot.handleUpdate(&update)
			if err == nil {
				return
			}
			if update.Message != nil {
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					err.Error(),
				))
			}
		}()
	}
}

func (bot *BotFramework) handleUpdate(update *tgbotapi.Update) error {
	if update.Message == nil {
		return errors.New("no message")
	}
	if update.Message.Photo != nil {
		return bot.handlePhoto(update)
	}
	if update.Message.Document != nil {
		return bot.handleFile(update)
	}
	if update.Message.IsCommand() {
		return bot.handleCommand(update)
	}
	if update.Message.Text != "" {
		err := bot.handleKeyboardCommand(update)
		if err == nil {
			return nil
		}
		if err.Error() != "command not found" {
			return err
		}
		err = bot.handlePlainText(update)
		if err == nil {
			return nil
		}
	}
	return errors.New("unknown command")
}

func (bot *BotFramework) RegisterCommand(name string, f commonHandler) error {
	if name[0] != '/' {
		return errors.New("command must start with slash")
	}
	if f == nil {
		return errors.New("handler must not be nil")
	}
	bot.commands[name] = f
	return nil
}

func (bot *BotFramework) handleCommand(update *tgbotapi.Update) error {
	if command, ok := bot.commands["/"+update.Message.Command()]; ok {
		return command(bot, update)
	}
	return errors.New("command not found")
}

func (bot *BotFramework) RegisterKeyboardCommand(name string, f commonHandler) error {
	if name[0] == '/' {
		return errors.New("keyboard command must not start with slash")
	}
	if f == nil {
		return errors.New("handler must not be nil")
	}
	bot.commands[name] = f
	return nil
}

func (bot *BotFramework) handleKeyboardCommand(update *tgbotapi.Update) error {
	if update.Message == nil {
		return errors.New("no message")
	}
	if command, ok := bot.commands[update.Message.Text]; ok {
		return command(bot, update)
	}
	return errors.New("command not found")
}

func (bot *BotFramework) RegisterPlainTextHandler(f commonHandler) error {
	bot.plainTextHandler = f
	return nil
}

func (bot *BotFramework) handlePlainText(update *tgbotapi.Update) error {
	if bot.plainTextHandler != nil {
		return bot.plainTextHandler(bot, update)
	}
	return errors.New("handler not set")
}

func (bot *BotFramework) RegisterPhotoHandler(f commonHandler) error {
	bot.photoHandler = f
	return nil
}

func (bot *BotFramework) handlePhoto(update *tgbotapi.Update) error {
	if bot.photoHandler != nil {
		return bot.photoHandler(bot, update)
	}
	return errors.New("handler not set")
}

func (bot *BotFramework) RegisterFileHandler(f commonHandler) error {
	bot.fileHandler = f
	return nil
}

func (bot *BotFramework) handleFile(update *tgbotapi.Update) error {
	if bot.fileHandler != nil {
		return bot.fileHandler(bot, update)
	}
	return errors.New("handler not set")
}
