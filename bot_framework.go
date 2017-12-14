package bot_framework

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"errors"
	"log"
)

type Sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type commonHandler func(bot *BotFramework, update *tgbotapi.Update) error

type BotFramework struct {
	Sender
	commands map[string]commonHandler
	handlers map[string]commonHandler
}

func NewBotFramework(api Sender) *BotFramework {
	bot := BotFramework{
		api,
		make(map[string]commonHandler),
		make(map[string]commonHandler, 5),
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
	if update.InlineQuery != nil {
		return bot.handle(update, "inline")
	}
	if update.CallbackQuery != nil {
		return bot.handle(update, "callback")
	}
	if update.Message == nil {
		return errors.New("no message")
	}
	if update.Message.Photo != nil {
		return bot.handle(update, "photo")
	}
	if update.Message.Document != nil {
		return bot.handle(update, "file")
	}
	if update.Message.Text != "" {
		err := bot.handleCommand(update)
		if err == nil {
			return nil
		}
		if err.Error() != "command not found" {
			return err
		}
		err = bot.handle(update, "plain")
		if err == nil {
			return nil
		}
	}
	return errors.New("unknown command")
}

func (bot *BotFramework) RegisterCommand(name string, f commonHandler) error {
	if f == nil {
		return errors.New("handler must not be nil")
	}
	bot.commands[name] = f
	return nil
}

func (bot *BotFramework) UnregisterCommand(name string) error {
	delete(bot.commands, name)
	return nil
}

func (bot *BotFramework) handleCommand(update *tgbotapi.Update) error {
	if update.Message.IsCommand() {
		if command, ok := bot.commands[update.Message.Command()]; ok {
			return command(bot, update)
		}
	}
	if command, ok := bot.commands[update.Message.Text]; ok {
		return command(bot, update)
	}
	return errors.New("command not found")
}

func (bot *BotFramework) RegisterPlainTextHandler(f commonHandler) error {
	bot.handlers["plain"] = f
	return nil
}

func (bot *BotFramework) UnregisterPlainTextHandler() error {
	delete(bot.handlers, "plain")
	return nil
}

func (bot *BotFramework) RegisterPhotoHandler(f commonHandler) error {
	bot.handlers["photo"] = f
	return nil
}

func (bot *BotFramework) UnregisterPhotoHandler() error {
	delete(bot.handlers, "photo")
	return nil
}

func (bot *BotFramework) RegisterFileHandler(f commonHandler) error {
	bot.handlers["file"] = f
	return nil
}

func (bot *BotFramework) UnregisterFileHandler() error {
	delete(bot.handlers, "file")
	return nil
}

func (bot *BotFramework) RegisterInlineQueryHandler(f commonHandler) error {
	bot.handlers["inline"] = f
	return nil
}

func (bot *BotFramework) UnregisterInlineQueryHandler() error {
	delete(bot.handlers, "inline")
	return nil
}

func (bot *BotFramework) RegisterCallbackQueryHandler(f commonHandler) error {
	bot.handlers["callback"] = f
	return nil
}

func (bot *BotFramework) UnregisterCallbackQueryHandler() error {
	delete(bot.handlers, "callback")
	return nil
}

func (bot *BotFramework) handle(update *tgbotapi.Update, event string) error {
	if f, ok := bot.handlers[event]; ok && f != nil {
		return f(bot, update)
	}
	return errors.New("unknown handler")
}
